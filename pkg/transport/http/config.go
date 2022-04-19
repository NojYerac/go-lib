package http

import (
	"io"
	"time"

	"github.com/gin-contrib/logger"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"source.rad.af/libs/go-lib/pkg/health"
	"source.rad.af/libs/go-lib/pkg/version"
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
		o.middleware = append(
			o.middleware,
			otelgin.Middleware(version.GetVersion().Name),
			logger.SetLogger(
				logger.WithLogger(func(c *gin.Context, w io.Writer, d time.Duration) zerolog.Logger {
					gLogger := l.
						Output(w).With().
						Timestamp().
						Int("status", c.Writer.Status()).
						Str("method", c.Request.Method).
						Str("path", c.Request.URL.Path).
						Str("ip", c.ClientIP()).
						Dur("latency", d).
						Stringer("latentcy_human", d).
						Str("user_agent", c.Request.UserAgent()).
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
