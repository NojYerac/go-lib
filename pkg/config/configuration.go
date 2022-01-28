package config

import (
	"os"

	"github.com/rs/zerolog"
)

type Configuration struct {
	LogConfigOnInit bool   `config:"log_config_on_init"`
	ConfigPath      string `config:"config_dir" flag:"configs,c" validate:"dir"`
}

type Option func(*configLoader)

func WithArgs(args ...string) Option {
	return func(cl *configLoader) {
		os.Args = append([]string{os.Args[0]}, args...)
	}
}

func WithLogger(l *zerolog.Logger) Option {
	return func(cl *configLoader) {
		cl.logger = l
	}
}
