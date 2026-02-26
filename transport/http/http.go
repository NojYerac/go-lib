package http

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/nojyerac/go-lib/health"
	"github.com/nojyerac/go-lib/log"
	"github.com/nojyerac/go-lib/metrics"
	"github.com/nojyerac/go-lib/tracing"
	"github.com/nojyerac/go-lib/version"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"
)

type Server interface {
	ListenAndServe(addr string) error
	Listen(net.Listener) error
	ServeHTTP(http.ResponseWriter, *http.Request)
	Handle(pattern string, handler http.Handler)
	HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request))
}

type server struct {
	logger         *logrus.Logger
	tracer         trace.Tracer
	meter          metric.Meter
	metricsHandler http.Handler
	middleware     []func(http.Handler) http.Handler
	apiPrefix      string
	mux            *http.ServeMux
	h              health.Checker
}

func (s *server) Handle(pattern string, handler http.Handler) {
	method, pattern, ok := strings.Cut(pattern, " ")
	if ok {
		pattern = s.apiPrefix + pattern
		pattern = method + " " + pattern
	} else {
		// no method specified, just prepend /api
		pattern = s.apiPrefix + method
	}
	s.mux.Handle(pattern, s.applyMiddleware(handler))
}

func (s *server) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	s.Handle(pattern, http.HandlerFunc(handler))
}

func (s *server) ListenAndServe(addr string) error {
	srv := &http.Server{
		Addr:              addr,
		Handler:           s.mux,
		ReadHeaderTimeout: time.Second * 10,
	}
	return srv.ListenAndServe()
}

func (s *server) Listen(l net.Listener) error {
	srv := &http.Server{
		Handler:           s.mux,
		ReadHeaderTimeout: time.Second * 10,
	}
	return srv.Serve(l)
}

func (s *server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.mux.ServeHTTP(w, req)
}

func (s *server) applyMiddleware(handler http.Handler) http.Handler {
	for _, mw := range s.middleware {
		handler = mw(handler)
	}
	return handler
}

func NewServer(c *Configuration, opts ...Option) Server {
	mux := http.NewServeMux()
	s := &server{
		logger:     logrus.StandardLogger(),
		tracer:     tracing.TracerForPackage(),
		meter:      metrics.MeterForPackage(),
		mux:        mux,
		middleware: make([]func(http.Handler) http.Handler, 0),
		apiPrefix:  "/api",
	}
	for _, applyOpt := range opts {
		applyOpt(s)
	}
	s.middleware = append(s.middleware, s.telemetryMiddleware(), panicHandler)
	s.mux.HandleFunc("/livez", pingHandler)
	s.mux.HandleFunc("/version", versionHandler)
	if s.metricsHandler != nil {
		s.mux.Handle("/metrics", s.metricsHandler)
	}
	if s.h != nil {
		s.mux.HandleFunc("/healthz", healthCheck(s.h))
	}
	return s
}

func panicHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte("[PANIC RECOVER] " + fmt.Sprint(err)))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

type loggingResponseWriter struct {
	http.ResponseWriter
	StatusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.StatusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (s *server) telemetryMiddleware() func(http.Handler) http.Handler {
	// run once to initialize telemetry components
	logger := s.logger
	tracer := s.tracer
	meter := s.meter
	if logger == nil {
		logger = log.Nop()
	}
	if meter == nil {
		meter = metrics.MeterForPackage()
	}
	if tracer == nil {
		tracer = tracing.TracerForPackage()
	}
	durationHistogram, err := meter.Float64Histogram("http_request_duration_histogram")
	if err != nil {
		logger.WithError(err).Fatal("failed to create duration histogram")
	}
	requestCounter, err := meter.Int64Counter("http_request_counter")
	if err != nil {
		logger.WithError(err).Fatal("failed to create request counter")
	}
	// middleware function
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// per-request telemetry setup
			now := time.Now()
			ctx := r.Context()
			ctx = otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier(r.Header))
			ctx, span := tracer.Start(ctx, "server.http_request")
			defer span.End()
			l := logger.WithFields(logrus.Fields{
				"spanID":     span.SpanContext().SpanID(),
				"traceID":    span.SpanContext().TraceID(),
				"method":     r.Method,
				"path":       r.URL.Path,
				"remoteAddr": r.RemoteAddr,
				"userAgent":  r.UserAgent(),
			}).Logger
			r = r.WithContext(log.WithLogger(ctx, l))
			lrw := &loggingResponseWriter{ResponseWriter: w, StatusCode: http.StatusOK}
			// handle the request
			next.ServeHTTP(lrw, r)
			// record telemetry
			duration := time.Since(now)
			logEntry := l.WithFields(logrus.Fields{
				"contentLength": r.ContentLength,
				"statusCode":    lrw.StatusCode,
				"duration":      duration.Milliseconds(),
			})
			switch {
			case lrw.StatusCode >= 500:
				logEntry.Error("server error")
			case lrw.StatusCode >= 400:
				logEntry.Warn("client error")
			default:
				logEntry.Info("success")
			}

			attrs := []attribute.KeyValue{}
			route := r.Method + " " + r.URL.Path
			attrs = append(attrs, semconv.HTTPServerAttributesFromHTTPRequest(version.GetVersion().Name, route, r)...)
			attrs = append(attrs, semconv.HTTPAttributesFromHTTPStatusCode(lrw.StatusCode)...)
			span.SetAttributes(attrs...)

			opt := metric.WithAttributeSet(attribute.NewSet(attrs...))
			durationHistogram.Record(ctx, float64(duration.Seconds()), opt)
			requestCounter.Add(ctx, 1, opt)
		})
	}
}

func pingHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func healthCheck(h health.Checker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := ""
		status := http.StatusOK
		if r.URL.Query().Has("v") {
			body = h.String()
		}
		if !h.Passed() {
			status = http.StatusServiceUnavailable
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}
}

func versionHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json charset=utf-8")
	_ = json.NewEncoder(w).Encode(version.GetVersion())
}
