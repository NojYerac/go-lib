package db

import (
	"time"

	"github.com/nojyerac/go-lib/pkg/health"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/trace"
)

type Configuration struct {
	Driver          string        `config:"database_driver" validate:"required"`
	DBConnStr       string        `config:"database_connection_string" validate:"required"`
	MaxIdleConns    uint          `config:"database_max_idle_connections"`
	MaxOpenConns    uint          `config:"database_max_open_connections"`
	ConnMaxIdleTime time.Duration `config:"database_connection_max_idle_time"`
	ConnMaxLifetime time.Duration `config:"database_connection_max_life_time"`
}

func NewConfiguration() *Configuration {
	return &Configuration{
		Driver: "postgres",
	}
}

type Option func(*options)

func WithLogger(l *zerolog.Logger) Option {
	return func(o *options) {
		o.l = l
	}
}

func WithHealthChecker(h health.Checker) Option {
	return func(o *options) {
		o.h = h
	}
}

type options struct {
	l *zerolog.Logger
	t trace.Tracer
	h health.Checker
}
