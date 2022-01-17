package health

import "time"

type Configuration struct {
	CheckInterval time.Duration `config:"healthcheck_check_interval" validate:"reqired,min=1s"`
}

func NewConfiguration() *Configuration {
	return &Configuration{
		CheckInterval: 30 * time.Second,
	}
}
