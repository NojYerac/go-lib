package http

import (
	"io"
	"time"

	"github.com/gin-contrib/logger"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
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
	middleware    []gin.HandlerFunc
	healthChecker health.Checker
}

func WithLogger(l *zerolog.Logger) Option {
	return func(o *options) {
		o.logger = l
		o.middleware = append(o.middleware, logger.SetLogger(
			logger.WithLogger(func(c *gin.Context, w io.Writer, d time.Duration) zerolog.Logger {
				gLogger := l.
					Output(w).With().
					Int("status", c.Writer.Status()).
					Str("method", c.Request.Method).
					Str("path", c.Request.URL.String()).
					Dur("latentcy", d).
					Stringer("latentcy_human", d).
					Logger()
				if gin.IsDebugging() {
					gLogger = gLogger.Output(zerolog.ConsoleWriter{Out: w})
					if gLogger.GetLevel() > zerolog.DebugLevel {
						gLogger = gLogger.Level(zerolog.DebugLevel)
					}
				}
				return gLogger
			}),
		), func(c *gin.Context) {
			c.Request = c.Request.WithContext(l.WithContext(c.Request.Context()))
			c.Next()
		})
	}
}

func WithHealthCheck(h health.Checker) Option {
	return func(o *options) {
		o.healthChecker = h
	}
}
