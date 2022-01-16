package log

import (
	"io"
	"os"

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

func NewLogger(config *Configuration, opts ...Option) zerolog.Logger {
	o := &options{
		output: os.Stdout,
	}
	for _, applyOpt := range opts {
		applyOpt(o)
	}
	if config.HumanFrendly {
		o.output = zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
			w.Out = o.output
		})
	}
	level, ok := levelMap[config.LogLevel]
	if !ok {
		level = zerolog.InfoLevel
	}
	l := zerolog.New(o.output).Level(level).With().Timestamp().Logger()
	if len(config.ServiceName) > 0 {
		l = l.With().Str("service", config.ServiceName).Logger()
	}
	return l
}
