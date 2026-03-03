# Go Library (go-lib)

**Production-Grade Go Foundation Library.**

`go-lib` is an opinionated toolkit for building resilient, observable, and maintainable Go microservices. It codifies production patterns—not just utilities—to ensure consistency across the service mesh.

## Core Pillars

### 1. Observability by Default
Every component in `go-lib` is designed with observability in mind. From structured logging with context propagation to automatic metric collection and distributed tracing (OpenTelemetry), the library ensures your services are never a black box.

### 2. Standardized Transport Patterns
`go-lib` provides thin but powerful abstractions over HTTP and gRPC. 
- **Middleware/Interceptors**: Standard implementations for Authn/Authz, Panic Recovery, Request ID propagation, and Telemetry.
- **Uniform Error Handling**: Consistent mapping between domain errors, HTTP status codes, and gRPC status codes.

### 3. Type-Safe Configuration
A robust configuration loader that supports environment variables, files, and defaults with built-in validation. It prevents services from starting in an invalid state by enforcing schema constraints at boot time.

## Quick Start: Bootstrapping a Service

Bootstrap a production-ready HTTP server with logging, health checks, and metrics in just a few lines.

```go
func main() {
    // 1. Initialize Configuration
    loader := config.NewConfigLoader("my-service")
    cfg := &MyConfig{Log: log.NewConfiguration(), Transport: transport.NewConfiguration()}
    loader.RegisterConfig(cfg.Log)
    loader.RegisterConfig(cfg.Transport)
    if err := loader.InitAndValidate(); err != nil {
        panic(err)
    }

    // 2. Setup Logger with Context Propagation
    l := log.NewLogger(cfg.Log)
    log.SetDefaultCtxLogger(l)

    // 3. Create HTTP Server with Production Middleware
    httpSrv := transporthttp.NewServer(
        transporthttp.NewConfiguration(),
        transporthttp.WithLogger(l),
        transporthttp.WithMetrics(),
    )
    
    // 4. Register Routes and Start
    httpSrv.HandleFunc("GET /api/v1/resource", HandleResource)
    srv, _ := transport.NewServer(cfg.Transport, transport.WithHTTP(httpSrv))
    srv.Start(context.Background())
}
```

## Pattern Examples

### Structured Logging with Context
Automatically pull Request IDs and Trace IDs from context into every log line.
```go
func MyHandler(w http.ResponseWriter, r *http.Request) {
    logger := log.FromContext(r.Context())
    logger.Info("Processing request", "resource_id", "123")
}
```

### Policy-Based Authorization
Define RBAC or ABAC requirements outside of your business logic.
```go
policies := authz.NewPolicyMap()
policies.Set(authz.HTTPOperation("POST", "/api/flags"), authz.RequireAny("admin"))

// Applied automatically in transport middleware
httpSrv := transporthttp.NewServer(cfg, transporthttp.WithAuthz(policies))
```

## Available Packages

| Package | Description |
|---------|-------------|
| [**audit**](./audit) | Standardized audit event primitives and logging. |
| [**auth**](./auth) | JWT validation and identity extraction. |
| [**authz**](./authz) | Policy-based authorization mapping. |
| [**config**](./config) | Multi-source configuration loader with validation. |
| [**db**](./db) | PostgreSQL connection management and metrics. |
| [**health**](./health) | Liveness/Readiness probe management. |
| [**log**](./log) | Context-aware structured logging (Logrus-based). |
| [**metrics**](./metrics) | Prometheus metric providers and helpers. |
| [**transport**](./transport) | Unified server management for gRPC and HTTP. |

## Developer Guide

### Requirements
- Go `1.25+`
- `golangci-lint`, `mockery`

### Common Commands
```bash
go test ./...         # Run all tests
./scripts/generate.sh # Generate mocks and proto
```
