package log

import (
	"os"

	"github.com/rs/zerolog"
)

func NewLogger(config *Configuration, opts ...Option) *zerolog.Logger {
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
	return &l
}

func Nop() *zerolog.Logger {
	l := zerolog.Nop()
	return &l
}

func SetDefaultCtxLogger(l *zerolog.Logger) {
	zerolog.DefaultContextLogger = l
}
