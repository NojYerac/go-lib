# go-lib

[![Go Reference](https://pkg.go.dev/badge/github.com/nojyerac/go-lib.svg)](https://pkg.go.dev/github.com/nojyerac/go-lib)
[![Go Report Card](https://goreportcard.com/badge/github.com/nojyerac/go-lib)](https://goreportcard.com/report/github.com/nojyerac/go-lib)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

> **Personal toolkit for Go microservices**  
> Opinionated, reusable building blocks extracted from real projects.

## Why This Exists

This is a **personal library** built to capture patterns I've used across multiple Go microservices. Instead of copying boilerplate between projects, I've extracted common components into a shared toolkit.

**What it provides:**
- Configuration loading with validation (env vars, files, defaults)
- Structured logging (logrus-based with context support)
- Health checks and liveness probes
- Metrics (Prometheus-compatible) and distributed tracing (OTLP)
- HTTP and gRPC server scaffolding
- JWT authentication and role-based authorization
- Database connection management (PostgreSQL-focused)
- Audit logging for mutation tracking

**Not a public library.** This is a learning project and personal toolbox. APIs may change without notice. If you find it useful, feel free to fork or take inspiration—but expect rough edges.

## Quick Start

The fastest way to start a new service is with the **scaffold** generator.
It creates a fully wired, compilable skeleton in one command:

```bash
go run github.com/nojyerac/go-lib/scaffold \
  --name orders \
  --module github.com/acme/orders

cd orders
go mod tidy
make run   # no TLS, port 8080, stdout tracing
```

The generated service ships with:
- Signal-aware entry-point with clean shutdown
- Config loading from env vars (prefixed `ORDERS_`)
- Structured JSON logging, Prometheus metrics, and OpenTelemetry tracing
- `/livez`, `/healthz`, `/metrics`, and `/version` endpoints
- Multi-stage Dockerfile, Makefile, and a GitHub Actions CI workflow

See the **[scaffold README](./scaffold/README.md)** for the full flag reference,
generated layout, and how to add routes or gRPC services.

## Available Packages

Each package is documented in its own README with usage examples.

### Scaffold
**[scaffold](./scaffold/README.md)** - One-command generator that creates a production-ready service skeleton wired to go-lib.

```bash
go run github.com/nojyerac/go-lib/scaffold --name orders --module github.com/acme/orders
```

### Configuration
**[config](./config/README.md)** - Load and validate configuration from environment variables, files, and defaults.

```go
import "github.com/nojyerac/go-lib/config"

cfg := &MyConfig{}
loader := config.NewConfigLoader("myapp")
loader.RegisterConfig(cfg)
loader.InitAndValidate()
```

### Logging
**[log](./log/README.md)** - Structured logging with context support (logrus-based).

```go
import "github.com/nojyerac/go-lib/log"

logger := log.NewLogger(log.NewConfiguration())
logger.Info("service started")

ctx := log.WithFields(context.Background(), log.Fields{"user_id": "123"})
log.FromContext(ctx).Info("user action")
```

### Health Checks
**[health](./health/README.md)** - Readiness and liveness probes for Kubernetes deployments.

```go
import "github.com/nojyerac/go-lib/health"

checker := health.NewChecker(health.NewConfiguration())
checker.RegisterCheck("database", dbHealthCheck)
go checker.Start(context.Background())

// Serves /health endpoint automatically
```

### Metrics
**[metrics](./metrics/README.md)** - Prometheus-compatible metrics collection.

```go
import "github.com/nojyerac/go-lib/metrics"

provider, handler, _ := metrics.NewMetricProvider()
metrics.SetGlobal(provider)

counter := provider.Counter("requests_total", "Total requests")
counter.Inc()

// handler serves /metrics endpoint
```

### Tracing
**[tracing](./tracing/README.md)** - Distributed tracing with OpenTelemetry (OTLP exporter).

```go
import "github.com/nojyerac/go-lib/tracing"

cfg := &tracing.Configuration{
  ExporterType: "otlp",
  OtlpEndpoint: "localhost:4317",
}
provider := tracing.NewTracerProvider(cfg)
tracing.SetGlobal(provider)
```

### HTTP Transport
**[transport/http](./transport/http/README.md)** - HTTP server with middleware, auth, and observability built-in.

```go
import transporthttp "github.com/nojyerac/go-lib/transport/http"

server := transporthttp.NewServer(
  transporthttp.NewConfiguration(),
  transporthttp.WithLogger(logger),
  transporthttp.WithHealthChecker(checker),
)
server.HandleFunc("GET /api/users", handleUsers)
```

### gRPC Transport
**[transport/grpc](./transport/grpc/README.md)** - gRPC server with interceptors for auth, logging, and metrics.

```go
import transportgrpc "github.com/nojyerac/go-lib/transport/grpc"

server := transportgrpc.NewServer(
  transportgrpc.NewConfiguration(),
  transportgrpc.WithLogger(logger),
)
// Register your gRPC services...
```

### Authentication
**[auth](./auth/README.md)** - JWT validation with HMAC and RSA support.

```go
import "github.com/nojyerac/go-lib/auth"

validator := auth.NewJWTValidator(&auth.Config{
  Issuer:     "myapp",
  Audience:   "api",
  HMACSecret: "secret-key",
})

claims, err := validator.ValidateToken(tokenString)
```

### Authorization
**[authz](./authz/README.md)** - Role-based access control with policy enforcement.

```go
import "github.com/nojyerac/go-lib/authz"

policies := map[string]authz.Policy{
  "/api/admin": {RequiredRoles: []string{"admin"}},
}

enforcer := authz.NewEnforcer(policies)
allowed := enforcer.Enforce("/api/admin", userRoles)
```

### Database
**[db](./db/README.md)** - PostgreSQL connection management with health checks.

```go
import "github.com/nojyerac/go-lib/db"

dbConn := db.NewConnection(&db.Config{
  Host:     "localhost",
  Port:     5432,
  Database: "myapp",
})
conn, _ := dbConn.Connect(context.Background())
```

### Audit Logging
**[audit](./audit/README.md)** - Transaction-safe audit trail for mutations.

```go
import "github.com/nojyerac/go-lib/audit"

logger := audit.NewLogger(db, "users")
logger.Log(ctx, tx, "UPDATE", userID, oldValue, newValue)
```

### Versioning
**[version](./version/README.md)** - Build metadata injection for version tracking.

```go
import "github.com/nojyerac/go-lib/version"

version.SetServiceName("myapp")
version.SetVersion("1.2.3")
info := version.GetInfo()
```

## Roadmap

See [ROADMAP.md](./ROADMAP.md) for planned features and improvements.

## Developer Guide

### Requirements

- Go `1.25+`
- Tooling: `ginkgo`, `golangci-lint`, `mockery`

### Common Commands

From `go-lib/`:

```bash
go test ./...             # Run all tests
ginkgo -r                 # Run tests with Ginkgo runner
golangci-lint run         # Lint codebase
./scripts/generate.sh     # Generate mocks and codegen
```

### Adding a New Package

1. Keep package APIs small and interface-first (similar to existing packages).
2. Add/update tests for behavior and integration points.
3. Document defaults, options, and usage in a package README.
4. Add the package link in this README when it is ready for use.

## License

MIT License - This is personal code, use at your own risk.
