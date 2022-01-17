package db

import "github.com/rs/zerolog"

type Configuration struct {
	Driver    string `config:"database_driver" validate:"required"`
	DBConnStr string `config:"database_connection_string" validate:"required"`
}

func NewConfiguration() *Configuration {
	return &Configuration{
		Driver: "postgres",
	}
}

type Option func(*options)

func WithLogger(l *zerolog.Logger) Option {
	return func(o *options) {
		o.l = l
	}
}

type options struct {
	l *zerolog.Logger
}
