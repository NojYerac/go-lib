# Go Library

> Common libraries for microservices

`go-lib` is an opinionated toolkit for Go microservices. It provides reusable
building blocks for configuration loading, structured logging, health checks,
metrics, tracing, transport, and version metadata.

## Quick Start

```go
package main

import (
  "context"
  "net/http"

  "github.com/nojyerac/go-lib/config"
  "github.com/nojyerac/go-lib/health"
  "github.com/nojyerac/go-lib/log"
  "github.com/nojyerac/go-lib/metrics"
  "github.com/nojyerac/go-lib/tracing"
  "github.com/nojyerac/go-lib/transport"
  transporthttp "github.com/nojyerac/go-lib/transport/http"
  "github.com/nojyerac/go-lib/version"
)

type AppConfig struct {
  Base      *config.Configuration
  Log       *log.Configuration
  Health    *health.Configuration
  Transport *transport.Configuration
  Trace     *tracing.Configuration
}

func main() {
  version.SetServiceName("example-service")

  cfg := &AppConfig{
    Base:      &config.Configuration{},
    Log:       log.NewConfiguration(),
    Health:    health.NewConfiguration(),
    Transport: transport.NewConfiguration(),
    Trace:     &tracing.Configuration{ExporterType: "noop"},
  }

  loader := config.NewConfigLoader("example")
  _ = loader.RegisterConfig(cfg.Base)
  _ = loader.RegisterConfig(cfg.Log)
  _ = loader.RegisterConfig(cfg.Health)
  _ = loader.RegisterConfig(cfg.Transport)
  _ = loader.RegisterConfig(cfg.Trace)
  if err := loader.InitAndValidate(); err != nil {
    panic(err)
  }

  l := log.NewLogger(cfg.Log)
  log.SetDefaultCtxLogger(l)

  h := health.NewChecker(cfg.Health)
  go h.Start(context.Background())

  mp, metricsHandler, _ := metrics.NewMetricProvider()
  metrics.SetGlobal(mp)

  tp := tracing.NewTracerProvider(cfg.Trace)
  tracing.SetGlobal(tp)

  httpSrv := transporthttp.NewServer(
    transporthttp.NewConfiguration(),
    transporthttp.WithLogger(l),
    transporthttp.WithHealthChecker(h),
    transporthttp.WithMetricsHandler(metricsHandler),
  )
  httpSrv.HandleFunc("GET /hello", func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    _, _ = w.Write([]byte("hello"))
  })

  s, err := transport.NewServer(cfg.Transport, transport.WithHTTP(httpSrv))
  if err != nil {
    panic(err)
  }
  if err := s.Start(context.Background()); err != nil {
    panic(err)
  }
}
```

## Available Packages

- [auth](./auth/README.md)
- [config](./config/README.md)
- [db](./db/README.md)
- [health](./health/README.md)
- [log](./log/README.md)
- [metrics](./metrics/README.md)
- [tracing](./tracing/README.md)
- [transport](./transport/README.md)
  - [http](./transport/http/README.md)
  - [grpc](./transport/grpc/README.md)
- [version](./version/README.md)

## Roadmap

- [go-lib roadmap](./ROADMAP.md)

## Developer Guide

### Requirements

- Go `1.25+`
- Tooling used by scripts: `ginkgo`, `golangci-lint`, `mockery`

### Common Commands

From `go-lib/`:

```bash
go test ./...
ginkgo -r
golangci-lint run
./scripts/generate.sh
```

### Adding a New Package

1. Keep package APIs small and interface-first (similar to existing packages).
2. Add/update tests for behavior and integration points.
3. Document defaults, options, and usage in a package README.
4. Add the package link in this README when it is ready for use.
