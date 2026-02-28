package auth

import "time"

type Configuration struct {
	Issuer     string        `config:"auth_issuer" validate:"required"`
	Audience   string        `config:"auth_audience" validate:"required"`
	HMACSecret string        `config:"auth_hmac_secret" validate:"required"`
	ClockSkew  time.Duration `config:"auth_clock_skew" validate:"min=0s"`
}

func NewConfiguration() *Configuration {
	return &Configuration{
		ClockSkew: 30 * time.Second,
	}
}
