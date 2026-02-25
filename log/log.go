package log

import (
	"context"
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

func NewLogger(c *Configuration, opts ...Option) *logrus.Logger {
	o := &options{
		output: os.Stdout,
	}
	for _, opt := range opts {
		opt(o)
	}
	logger := logrus.New()
	logger.SetLevel(levelFromString(c.LogLevel))
	logger.SetOutput(o.output)
	if c.HumanFrendly {
		logger.SetFormatter(&logrus.TextFormatter{})
	} else {
		logger.SetFormatter(&logrus.JSONFormatter{})
	}
	return logger.WithField("service", c.ServiceName).Logger
}

func Nop() *logrus.Logger {
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	return logger
}

var defaultCtxLogger *logrus.Logger

func SetDefaultCtxLogger(l *logrus.Logger) {
	defaultCtxLogger = l
}

type ctxLoggerKeyType struct{}

var ctxLoggerKey = ctxLoggerKeyType{}

func WithLogger(ctx context.Context, l *logrus.Logger) context.Context {
	return context.WithValue(ctx, ctxLoggerKey, l)
}

func FromContext(ctx context.Context) *logrus.Logger {
	if logger, ok := ctx.Value(ctxLoggerKey).(*logrus.Logger); ok {
		return logger
	}
	if defaultCtxLogger != nil {
		return defaultCtxLogger
	}
	return Nop()
}

var TestConfig = &Configuration{
	ServiceName:  "testing",
	HumanFrendly: true,
	LogLevel:     "fatal",
}
