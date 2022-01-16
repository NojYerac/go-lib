package health

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

type Configuration struct {
	CheckInterval time.Duration `config:"healthcheck_check_interval" validate:"reqired,min=1s"`
}

func NewConfiguration() *Configuration {
	return &Configuration{
		CheckInterval: 30 * time.Second,
	}
}

type Checker interface {
	Start(context.Context) error
	Register(string, CheckFn)
	Report
}

type Option func(*checker)

func WithReadyChan(ready chan<- struct{}) Option {
	return func(c *checker) {
		c.ready = ready
	}
}

func pingCheck(_ context.Context) error { return nil }

func NewChecker(config *Configuration, opts ...Option) Checker {
	c := &checker{
		r:     &reports{m: make(map[string]Report)},
		c:     map[string]CheckFn{"ping": pingCheck},
		clock: time.NewTicker(config.CheckInterval),
	}
	for _, applyOpt := range opts {
		applyOpt(c)
	}
	return c
}

type checker struct {
	sync.RWMutex
	ready chan<- struct{}
	r     Report
	c     map[string]CheckFn
	clock *time.Ticker
}

func (c *checker) Register(name string, checkFn CheckFn) {
	c.Lock()
	defer c.Unlock()
	c.c[name] = checkFn
}

func (c *checker) Passed() bool {
	if c == nil {
		return true
	}
	c.RLock()
	defer c.RUnlock()
	return c.r.Passed()
}

func (c *checker) String() string {
	if c == nil {
		return ""
	}
	c.RLock()
	defer c.RUnlock()
	return c.r.String()
}

func (c *checker) Start(ctx context.Context) error {
	l := zerolog.Ctx(ctx)
	c.checkNow(ctx)
	close(c.ready)
	for {
		if c.Passed() {
			l.Trace().Stringer("report", c).Msg("healthy")
		} else {
			l.Warn().Stringer("report", c).Msg("sick")
		}
		select {
		case <-c.clock.C:
			c.checkNow(ctx)
			continue
		case <-ctx.Done():
			c.clock.Stop()
		}
		break
	}
	return ctx.Err()
}

func (c *checker) checkNow(ctx context.Context) {
	c.Lock()
	defer c.Unlock()
	m := make(map[string]Report)
	for name, checkFn := range c.c {
		m[name] = &report{err: checkFn(ctx)}
	}
	c.r = &reports{m: m}
}

type CheckFn func(context.Context) error

type Report interface {
	Passed() bool
	String() string
}

type reports struct {
	sync.RWMutex
	m map[string]Report
}

func (r *reports) Passed() bool {
	r.RLock()
	defer r.RUnlock()
	for _, sub := range r.m {
		if !sub.Passed() {
			return false
		}
	}
	return true
}

func (r *reports) String() string {
	r.RLock()
	defer r.RUnlock()
	strs := make([]string, 0, len(r.m))
	for name, sub := range r.m {
		strs = append(strs, fmt.Sprint("[", name, "] ", sub.String()))
	}
	return strings.Join(strs, "\n")
}

type report struct {
	err error
}

func (r *report) Passed() bool {
	return r.err == nil
}

func (r *report) String() string {
	if r.err != nil {
		return r.err.Error()
	}
	return "ok"
}
