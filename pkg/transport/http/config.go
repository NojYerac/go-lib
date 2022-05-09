package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"source.rad.af/libs/go-lib/pkg/health"
)

type Configuration struct {
	ServerDebug bool `config:"server_debug"`
}

func NewConfiguration() *Configuration {
	return &Configuration{}
}

type Option func(*options)

type options struct {
	logger        *zerolog.Logger
	meter         metric.Meter
	tracer        trace.Tracer
	middleware    []gin.HandlerFunc
	healthChecker health.Checker
	metricHandler http.Handler
}

func WithLogger(l *zerolog.Logger) Option {
	return func(o *options) {
		o.logger = l
	}
}

func WithHealthCheck(h health.Checker) Option {
	return func(o *options) {
		o.healthChecker = h
	}
}

func WithMetricHandler(h http.Handler) Option {
	return func(o *options) {
		o.metricHandler = h
	}
}

func WithMiddleware(h gin.HandlerFunc) Option {
	return func(o *options) {
		o.middleware = append(o.middleware, h)
	}
}
