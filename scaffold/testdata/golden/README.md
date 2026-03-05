# Example

> Scaffolded with [go-lib](github.com/nojyerac/go-lib) — scaffold v1.25.1.

## Overview

`example` is a Go microservice built on go-lib conventions.  It provides:

- **HTTP** and optional **gRPC** serving on a single multiplexed port (`cmux`)
- Structured JSON logging via `logrus`
- Distributed tracing via OpenTelemetry
- Prometheus metrics (exported at `/metrics`)
- Liveness (`/livez`) and readiness (`/healthz`) probes
- Build-time version metadata (`/version`)

## Quick start

```bash
# Install dependencies
go mod tidy

# Run locally (no TLS, port 8080)
make run

# Run tests
make test

# Lint
make lint
```

## Configuration

All configuration is driven by environment variables prefixed with `EXAMPLE_`.
The table below lists the most important knobs; refer to the go-lib sub-package
READMEs for full documentation.

| Variable | Default | Description |
| -------- | ------- | ----------- |
| `EXAMPLE_PORT` | `80` | Listening port |
| `EXAMPLE_NO_TLS` | `false` | Disable TLS (useful for local dev) |
| `EXAMPLE_LOG_LEVEL` | `info` | Log level (`trace`…`panic`) |
| `EXAMPLE_SERVICE_NAME` | `example` | Value embedded in log fields |
| `EXAMPLE_EXPORTER_TYPE` | `noop` | Trace exporter (`stdout`, `otlp`, `noop`) |
| `EXAMPLE_HEALTHCHECK_CHECK_INTERVAL` | `30s` | Health-check tick interval |

## Building a Docker image

```bash
make docker          # builds example:dev
docker run --rm -p 8080:8080 example:dev
```

## Project layout

```text
cmd/example/     — main package / entry-point
config/            — root Config struct + defaults
internal/app/      — subsystem wiring (Run)
transport/         — HTTP route registration
scripts/           — lint + test helpers
.github/workflows/ — CI pipeline
Dockerfile
Makefile
```

## Adding routes

Edit `transport/server.go` and call `s.HandleFunc` or `s.Handle` with your
handler.  The API prefix `/api` is applied automatically by go-lib.

## gRPC

Instantiate a `*grpc.Server`, register your service implementations, then pass
it to `transport.NewServer` via `transport.WithGRPC(grpcServer)` in
`internal/app/app.go`.
