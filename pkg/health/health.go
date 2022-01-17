package health

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

type CheckFn func(context.Context) error

func pingCheck(_ context.Context) error { return nil }

type Checker interface {
	Start(context.Context) error
	Register(string, CheckFn)
	Report
}

func NewChecker(config *Configuration) Checker {
	return &checker{
		r:     &reports{m: make(map[string]Report)},
		c:     map[string]CheckFn{"ping": pingCheck},
		clock: time.NewTicker(config.CheckInterval),
		ready: make(chan struct{}),
	}
}

type checker struct {
	sync.RWMutex
	ready chan struct{}
	r     Report
	c     map[string]CheckFn
	clock *time.Ticker
}

func (c *checker) Register(name string, checkFn CheckFn) {
	c.Lock()
	defer c.Unlock()
	c.c[name] = checkFn
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

func (c *checker) Passed() bool {
	if c == nil {
		return true
	}
	<-c.ready
	c.RLock()
	defer c.RUnlock()
	return c.r.Passed()
}

func (c *checker) String() string {
	if c == nil {
		return ""
	}
	<-c.ready
	c.RLock()
	defer c.RUnlock()
	return c.r.String()
}

func (c *checker) checkNow(ctx context.Context) {
	c.Lock()
	defer c.Unlock()
	m := make(map[string]Report, len(c.c))
	for name, checkFn := range c.c {
		m[name] = newReport(checkFn(ctx))
	}
	c.r = &reports{m: m}
}
