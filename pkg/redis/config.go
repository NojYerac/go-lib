package redis

import (
	"github.com/go-redis/redis"
	"source.rad.af/libs/go-lib/pkg/health"
)

type Configuration struct {
	Addr string `config:"redis_address" validate:"required,uri"`
}

func NewConfiguration() *Configuration {
	return &Configuration{
		Addr: "redis:6379",
	}
}

type Option func(*client)

func WithHealthChecker(h health.Checker) Option {
	return func(c *client) {
		c.h = h
	}
}

func WithRedisClient(r redis.Cmdable) Option {
	return func(c *client) {
		c.r = r
	}
}
