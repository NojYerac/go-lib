package log

import (
	"context"
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

func NewLogger(c *Configuration, opts ...Option) logrus.FieldLogger {
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
	return logger.WithField("service", c.ServiceName)
}

func Nop() *logrus.Logger {
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	return logger
}

var defaultCtxLogger logrus.FieldLogger

func SetDefaultCtxLogger(l logrus.FieldLogger) {
	defaultCtxLogger = l
}

type ctxLoggerKeyType struct{}

var ctxLoggerKey = ctxLoggerKeyType{}

func WithLogger(ctx context.Context, l logrus.FieldLogger) context.Context {
	return context.WithValue(ctx, ctxLoggerKey, l)
}

func FromContext(ctx context.Context) logrus.FieldLogger {
	if logger, ok := ctx.Value(ctxLoggerKey).(logrus.FieldLogger); ok {
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
