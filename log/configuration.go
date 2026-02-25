package log

import (
	"io"

	"github.com/nojyerac/go-lib/version"
	"github.com/sirupsen/logrus"
)

func NewConfiguration() *Configuration {
	return &Configuration{
		ServiceName: version.GetVersion().Name,
		LogLevel:    "info",
	}
}

type Configuration struct {
	ServiceName  string `config:"service_name" validate:"required"`
	HumanFrendly bool   `config:"human_friendly"`
	LogLevel     string `config:"log_level" flag:"logLevel,l" validate:"oneof=trace debug info warn error fatal panic"`
}

func levelFromString(level string) logrus.Level {
	switch level {
	case "debug":
		return logrus.DebugLevel
	case "info":
		return logrus.InfoLevel
	case "warn":
		return logrus.WarnLevel
	case "error":
		return logrus.ErrorLevel
	case "fatal":
		return logrus.FatalLevel
	case "panic":
		return logrus.PanicLevel
	default:
		return logrus.InfoLevel
	}
}

type Option func(*options)

type options struct {
	output io.Writer
}

func WithOutput(output io.Writer) Option {
	return func(o *options) {
		o.output = output
	}
}
