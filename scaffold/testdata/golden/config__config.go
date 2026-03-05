// Package config holds the service-wide configuration structs and their
// defaults.  Each embedded struct maps to a distinct go-lib subsystem.
package config

import (
	"github.com/nojyerac/go-lib/health"
	"github.com/nojyerac/go-lib/log"
	"github.com/nojyerac/go-lib/tracing"
	libhttp "github.com/nojyerac/go-lib/transport/http"
	"github.com/nojyerac/go-lib/transport"
)

// Config is the root configuration for example.
// Fields are loaded by go-lib's config.Loader using the EXAMPLE_ env-var
// prefix and optional config files in ./config/.
type Config struct {
	Log       log.Configuration       `mapstructure:",squash"`
	Health    health.Configuration    `mapstructure:",squash"`
	Transport transport.Configuration `mapstructure:",squash"`
	HTTP      libhttp.Configuration   `mapstructure:",squash"`
	Tracing   tracing.Configuration   `mapstructure:",squash"`
}

// NewConfig returns a Config initialised with safe, local-development defaults.
func NewConfig() *Config {
	return &Config{
		Log:       *log.NewConfiguration(),
		Health:    *health.NewConfiguration(),
		Transport: *transport.NewConfiguration(),
		HTTP:      *libhttp.NewConfiguration(),
		Tracing: tracing.Configuration{
			ExporterType: "noop",
			SampleRatio:  1.0,
		},
	}
}
