package log

import (
	"io"

	"github.com/rs/zerolog"
)

type Configuration struct {
	ServiceName  string `config:"service_name"`
	LogLevel     string `config:"log_level" flag:"logLevel,l" validate:"oneof=trace debug info warn error fatal panic"`
	HumanFrendly bool   `config:"human_friendly"`
}

var levelMap = map[string]zerolog.Level{
	"trace": zerolog.TraceLevel,
	"debug": zerolog.DebugLevel,
	"info":  zerolog.InfoLevel,
	"warn":  zerolog.WarnLevel,
	"error": zerolog.ErrorLevel,
	"fatal": zerolog.FatalLevel,
	"panic": zerolog.PanicLevel,
}

func NewConfiguration() *Configuration {
	return &Configuration{
		LogLevel: "info",
	}
}

type options struct {
	output io.Writer
}

type Option func(*options)

func WithOutput(w io.Writer) Option {
	return func(o *options) {
		o.output = w
	}
}

var (
	DebugConfig = &Configuration{HumanFrendly: true, LogLevel: "debug"}
	TestConfig  = &Configuration{LogLevel: "fatal"}
)
