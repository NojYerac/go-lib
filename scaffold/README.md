# scaffold

One-command generator that creates a production-ready Go microservice skeleton
already wired to `go-lib` conventions.

## Usage

```bash
go run github.com/nojyerac/go-lib/scaffold \
  --name orders \
  --module github.com/acme/orders
```

| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| `--name` | yes | — | Short service name used as the directory name, binary name, and env-var prefix |
| `--module` | yes | — | Fully-qualified Go module path written into `go.mod` |
| `--out` | no | `.` | Parent directory to write the generated service into |
| `--dry-run` | no | `false` | Print the file list without writing anything to disk |
| `--golib-module` | no | `github.com/nojyerac/go-lib` | Override the go-lib import path (useful when forking) |
| `--go-version` | no | `1.25.1` | Minimum Go version written into the generated `go.mod` |

## Generated layout

Running the command above produces an `orders/` directory:

```tree
orders/
├── cmd/orders/
│   └── main.go                  # signal-aware entry-point
├── config/
│   └── config.go                # root Config struct + defaults
├── internal/app/
│   ├── app.go                   # subsystem wiring (Run)
│   └── app_test.go              # smoke test with env overrides
├── transport/
│   └── server.go                # HTTP route registration
├── .github/workflows/
│   └── ci.yml                   # test → lint → docker build pipeline
├── scripts/
│   ├── lint.sh
│   └── test.sh
├── Dockerfile                   # multi-stage, scratch runtime image
├── Makefile                     # run / test / lint / build / docker targets
├── go.mod
└── README.md
```

### What is wired automatically

The generated `internal/app/app.go` initialises all go-lib subsystems in the
correct order:

1. **Config** — `config.NewConfigLoader` with your service's env-var prefix
2. **Logger** — `log.NewLogger`, set as the global context logger
3. **Version** — `version.SetServiceName`
4. **Tracing** — `tracing.NewTracerProvider` + `tracing.SetGlobal`
5. **Metrics** — `metrics.NewMetricProvider` + `metrics.SetGlobal`
6. **Health** — `health.NewChecker` started in a goroutine
7. **HTTP server** — `transport/http.NewServer` with health + metrics wired
8. **Transport** — `transport.NewServer` (cmux; HTTP and optionally gRPC)

Utility endpoints wired by go-lib's HTTP server out of the box:

| Path | Description |
|------|-------------|
| `/livez` | Liveness probe |
| `/healthz` | Readiness probe (aggregated health checks) |
| `/metrics` | Prometheus metrics |
| `/version` | Build metadata JSON |

### Adding routes

Edit `transport/server.go`:

```go
func RegisterRoutes(s libhttp.Server) {
    s.HandleFunc("GET /orders", listOrdersHandler)
    s.HandleFunc("POST /orders", createOrderHandler)
}
```

The HTTP server automatically prepends the `/api` prefix, so the above becomes
`GET /api/orders`.

### Adding gRPC

In `internal/app/app.go`, create a `*grpc.Server`, register your service
implementations, and pass it to the transport:

```go
grpcSrv := grpc.NewServer(/* interceptors */)
orderspb.RegisterOrdersServer(grpcSrv, &ordersHandler{})

srv, err := transport.NewServer(&cfg.Transport,
    transport.WithHTTP(httpServer),
    transport.WithGRPC(grpcSrv),
)
```

## Running the generated service locally

```bash
cd orders
go mod tidy
make run      # no TLS, port 8080, stdout tracing, debug logging
```

### Environment variables

All env vars are prefixed with the uppercased service name (e.g. `ORDERS_`).
The most common knobs:

| Variable | Default | Description |
|----------|---------|-------------|
| `ORDERS_PORT` | `80` | Listening port |
| `ORDERS_NO_TLS` | `false` | Disable TLS |
| `ORDERS_LOG_LEVEL` | `info` | `trace` / `debug` / `info` / `warn` / `error` |
| `ORDERS_SERVICE_NAME` | `orders` | Embedded in every log line |
| `ORDERS_EXPORTER_TYPE` | `noop` | Trace exporter: `stdout`, `otlp`, or `noop` |
| `ORDERS_HEALTHCHECK_CHECK_INTERVAL` | `30s` | Health-check tick interval |

## Building a Docker image

```bash
make docker            # builds orders:dev
docker run --rm -p 8080:8080 \
  -e ORDERS_NO_TLS=true \
  orders:dev
```

## Updating templates

Templates live in `scaffold/templates/*.tmpl` and are embedded at compile time
via `//go:embed`.  After editing a template, regenerate the golden snapshots:

```bash
UPDATE_GOLDEN=1 go test ./scaffold/... -run TestGenerate_Golden
```

## Testing the scaffold itself

```bash
# Unit + golden tests (fast)
go test ./scaffold/... -short

# Full suite including compilation of a generated service
go test ./scaffold/... -v
```

Test descriptions:

| Test | What it covers |
|------|----------------|
| `TestGenerate_Golden` | Output matches checked-in snapshots under `testdata/golden/` |
| `TestGenerate_AllFilesPresent` | Every expected file is produced |
| `TestGenerate_NameIsSubstituted` | Name and module are injected correctly |
| `TestGenerate_Deterministic` | Repeated runs produce identical output |
| `TestE2E_WritesExpectedFileTree` | Files are written to disk at the correct paths |
| `TestE2E_FileContentsAreCoherent` | Key strings (name, module, prefix) appear in the right files |
| `TestE2E_NoStrayTemplateSyntax` | No `{{.` markers remain in any generated file |
| `TestE2E_DeterminsticLayout` | Two on-disk generations are byte-for-byte identical |
| `TestE2E_DryRun` | `--dry-run` prints paths and writes nothing |
| `TestE2E_InvalidInput` | Missing / invalid flags cause non-zero exit |
| `TestE2E_Compiles` | Generated service compiles against local go-lib source |
