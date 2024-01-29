package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis"
	"go.opentelemetry.io/otel/trace"
	"source.rad.af/libs/go-lib/pkg/health"
	"source.rad.af/libs/go-lib/pkg/tracing"
)

// NewClient returns a new redis client
func NewClient(config *Configuration, opts ...Option) (Client, error) {
	c := &client{
		t: tracing.TracerForPackage(),
	}
	for _, apply := range opts {
		apply(c)
	}
	if c.r == nil {
		c.r = redis.NewClient(&redis.Options{
			Addr: config.Addr,
		})
	}
	if err := c.r.Ping().Err(); err != nil {
		return nil, err
	}
	if c.h != nil {
		c.h.Register("redis_check", c.Ping)
	}
	return c, nil
}

// Client is a testable wrapper around redis.Cmdable
type Client interface {
	Ping(context.Context) error
	Del(parentCtx context.Context, keys ...string) (int64, error)
	Get(parentCtx context.Context, key string) (string, error)
	Set(parentCtx context.Context, key string, value interface{}, expiration time.Duration) error
}

type client struct {
	r redis.Cmdable
	t trace.Tracer
	h health.Checker
}

func (c client) Ping(context.Context) error {
	return c.r.Ping().Err()
}

func (c client) Del(parentCtx context.Context, keys ...string) (int64, error) {
	_, span := c.t.Start(parentCtx, "redis.del")
	defer span.End()
	return c.r.Del(keys...).Result()
}

func (c client) Get(parentCtx context.Context, key string) (string, error) {
	_, span := c.t.Start(parentCtx, "redis.get")
	defer span.End()
	return c.r.Get(key).Result()
}

func (c client) Set(parentCtx context.Context, key string, value interface{}, exp time.Duration) error {
	_, span := c.t.Start(parentCtx, "redis.set")
	defer span.End()
	return c.r.Set(key, value, exp).Err()
}
