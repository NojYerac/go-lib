# Transport HTTP Package

The `transport/http` package wraps `net/http` with:

- utility endpoints (`/livez`, `/version`, optional `/healthz`, `/metrics`)
- API route registration with method patterns
- telemetry + structured request logging middleware
- panic recovery middleware

## API

### `NewServer(config *Configuration, opts ...Option) Server`

Returns a server implementing:

- `Handle(pattern string, handler http.Handler)`
- `HandleFunc(pattern string, fn func(http.ResponseWriter, *http.Request))`
- `ServeHTTP`, `Listen`, `ListenAndServe`

### Options

- `WithHealthChecker(health.Checker)`
- `WithMetricsHandler(http.Handler)`
- `WithLogger(logrus.FieldLogger)`
- `WithMiddleware(func(http.Handler) http.Handler)`
- `WithAuthMiddleware(auth.Validator, authz.PolicyMap)`

`WithAuthMiddleware` enforces auth only for operations present in the provided
policy map. Missing/invalid tokens map to `401`, and failed role checks map to
`403`.

## Routes

Always available:

- `GET /livez` → `200 ok`
- `GET /version` → JSON `version.GetVersion()`

Optional:

- `GET /healthz` if `WithHealthChecker` is set
  - returns `200` when healthy, `503` when unhealthy
  - `?v` adds verbose report body
- `GET /metrics` if `WithMetricsHandler` is set

Custom API routes are mounted under `/api`.

Example: `Handle("GET /orders", h)` serves `GET /api/orders`.

## Example

```go
h := transporthttp.NewServer(
    transporthttp.NewConfiguration(),
    transporthttp.WithLogger(logger),
    transporthttp.WithHealthChecker(checker),
    transporthttp.WithMetricsHandler(metricsHandler),
)

h.HandleFunc("GET /orders", func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    _, _ = w.Write([]byte("[]"))
})
```

## Auth Example

```go
policies := authz.NewPolicyMap()
policies.Set(authz.HTTPOperation("GET", "/api/orders"), authz.RequireAny("reader", "admin"))

h := transporthttp.NewServer(
    transporthttp.NewConfiguration(),
    transporthttp.WithAuthMiddleware(validator, policies),
)
```
