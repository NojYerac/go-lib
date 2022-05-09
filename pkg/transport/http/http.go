package http

import (
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/metric/instrument/syncint64"
	"go.opentelemetry.io/otel/metric/unit"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"source.rad.af/libs/go-lib/pkg/health"
	"source.rad.af/libs/go-lib/pkg/log"
	"source.rad.af/libs/go-lib/pkg/metrics"
	"source.rad.af/libs/go-lib/pkg/tracing"
	"source.rad.af/libs/go-lib/pkg/version"
)

type Server interface {
	Listen(net.Listener) error
	APIRoutes() gin.IRouter
	ServeHTTP(http.ResponseWriter, *http.Request)
}

type server struct {
	engine    *gin.Engine
	apiRoutes *gin.RouterGroup
	options   *options
}

func (s *server) Listen(l net.Listener) error {
	return s.engine.RunListener(l)
}

func (s *server) APIRoutes() gin.IRouter {
	return s.apiRoutes
}

func (s *server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.engine.ServeHTTP(w, req)
}

func NewServer(config *Configuration, opts ...Option) Server {
	o := &options{
		logger: log.Nop(),
		tracer: tracing.TracerForPackage(),
		meter:  metrics.MeterForPackage(),
	}
	for _, applyOpt := range opts {
		applyOpt(o)
	}
	if !config.ServerDebug {
		gin.SetMode(gin.ReleaseMode)
	}
	engine := gin.New()
	engine.GET("/ping", pingHandler)
	engine.GET("/version", versionHandler)
	if o.healthChecker != nil {
		engine.GET("/health", healthHandler(o.healthChecker))
	}
	if o.metricHandler != nil {
		engine.GET("/metrics", gin.WrapH(o.metricHandler))
	}
	s := &server{
		apiRoutes: engine.Group("/api"),
		engine:    engine,
		options:   o,
	}
	s.apiRoutes.Use(s.telemetryMW(), recoverMW())
	return s
}

func healthHandler(checker health.Checker) gin.HandlerFunc {
	return func(c *gin.Context) {
		body := ""
		code := http.StatusOK
		if !checker.Passed() {
			code = http.StatusServiceUnavailable
		}
		if _, ok := c.GetQuery("verbose"); ok {
			body = checker.String()
		}
		c.String(code, body)
	}
}

func versionHandler(c *gin.Context) {
	c.JSON(http.StatusOK, version.GetVersion())
}

func pingHandler(c *gin.Context) {
	c.String(http.StatusOK, "pong")
}

var durationHistogram syncint64.Histogram

func (s *server) telemetryMW() gin.HandlerFunc {
	gin.DefaultWriter = s.options.logger
	serverName := version.GetVersion().Name
	if durationHistogram == nil {
		var err error
		durationHistogram, err = s.options.meter.SyncInt64().Histogram(
			"http_request_duration_histogram",
			instrument.WithDescription("duration (ms) of http requests"),
			instrument.WithUnit(unit.Milliseconds),
		)
		if err != nil {
			panic(err)
		}
	}
	return func(c *gin.Context) {
		ctx := otel.GetTextMapPropagator().Extract(
			c.Request.Context(),
			propagation.HeaderCarrier(c.Request.Header),
		)
		ctx, span := s.options.tracer.Start(ctx, "server.http_request")
		defer span.End()
		gLogger := s.options.logger.With().Stringer("traceID", span.SpanContext().TraceID()).Logger()
		c.Request = c.Request.WithContext(gLogger.WithContext(ctx))
		startTime := time.Now()
		c.Next()
		d := time.Since(startTime)
		span.SetAttributes(semconv.HTTPServerAttributesFromHTTPRequest(
			serverName,
			c.FullPath(),
			c.Request,
		)...)
		gLogger.Info().
			Int("status", c.Writer.Status()).
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Str("ip", c.ClientIP()).
			Dur("latency", d).
			Stringer("latentcy_human", d).
			Str("user_agent", c.Request.UserAgent()).
			Send()
		durationHistogram.Record(
			c.Request.Context(),
			d.Milliseconds(),
			semconv.HTTPServerMetricAttributesFromHTTPRequest(
				serverName,
				c.Request,
			)...)
	}
}

func recoverMW() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				c.String(http.StatusInternalServerError, "[PANIC RECOVER] %v", r)
			}
		}()
		c.Next()
	}
}
