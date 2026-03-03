package health

import (
	"context"
	"sync"
	"time"

	"github.com/nojyerac/go-lib/log"
)

type CheckFn func(context.Context) error

func pingCheck(_ context.Context) error { return nil }

type Checker interface {
	Start(context.Context) error
	Register(string, CheckFn)
	Reporter
}

func NewChecker(config *Configuration) Checker {
	return &checker{
		r:     &reports{m: make(map[string]Reporter)},
		c:     map[string]CheckFn{"ping": pingCheck},
		clock: time.NewTicker(config.CheckInterval),
		ready: make(chan struct{}),
	}
}

type checker struct {
	sync.RWMutex
	ready chan struct{}
	r     Reporter
	c     map[string]CheckFn
	clock *time.Ticker
}

func (c *checker) Register(name string, checkFn CheckFn) {
	c.Lock()
	defer c.Unlock()
	c.c[name] = checkFn
}

func (c *checker) Start(ctx context.Context) error {
	l := log.FromContext(ctx)
	c.checkNow(ctx)
	close(c.ready)
	for {
		ll := l.WithField("report", c.String())
		if c.Passed() {
			ll.Debug("healthy")
		} else {
			ll.Warn("sick")
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
	m := make(map[string]Reporter, len(c.c))
	for name, checkFn := range c.c {
		m[name] = newReport(checkFn(ctx))
	}
	c.r = &reports{m: m}
}
