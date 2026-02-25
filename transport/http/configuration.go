package http

import (
	"net/http"

	"github.com/nojyerac/go-lib/health"
	"github.com/sirupsen/logrus"
)

type Configuration struct{}

func NewConfiguration() *Configuration {
	return &Configuration{}
}

type Option func(*server)

func WithHealthChecker(h health.Checker) Option {
	return func(s *server) {
		s.h = h
	}
}

func WithMetricsHandler(mh http.Handler) Option {
	return func(s *server) {
		s.metricsHandler = mh
	}
}

func WithLogger(l *logrus.Logger) Option {
	return func(s *server) {
		s.logger = l
	}
}

func WithMiddleware(mw func(http.Handler) http.Handler) Option {
	return func(s *server) {
		s.middleware = append(s.middleware, mw)
	}
}
