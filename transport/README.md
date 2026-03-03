# Transport Package

The `transport` package runs HTTP and gRPC servers on one listener using `cmux`.

## Configuration

```go
type Configuration struct {
    NoTLS    bool   `config:"no_tls"`
    PubCert  string `config:"tls_public_cert" validate:"required_unless=NoTLS true"`
    PrivKey  string `config:"tls_private_key" validate:"required_unless=NoTLS true"`
    RootCA   string `config:"tls_root_ca"`
    Hostname string `config:"hostname" validate:"required,hostname_rfc1123"`
    Port     string `config:"port" validate:"required,numeric,min=1,max=65535"`
}
```

`NewConfiguration()` defaults:

- `Hostname: "0.0.0.0"`
- `Port: "80"`

## API

- `NewServer(config, opts...) (Server, error)`
- `WithHTTP(http.Server)`
- `WithGRPC(*grpc.Server)`
- `WithListener(net.Listener)`
- `Server.Start(context.Context) error`

`Start` blocks until context cancellation, then gracefully stops the gRPC server
and closes listeners.

## Example (HTTP + gRPC)

```go
h := transporthttp.NewServer(transporthttp.NewConfiguration())
g := transportgrpc.NewServer(func(s *grpc.Server) {
    pb.RegisterMyServiceServer(s, myImpl)
})

cfg := transport.NewConfiguration()
cfg.NoTLS = true
cfg.Port = "8080"

srv, err := transport.NewServer(cfg, transport.WithHTTP(h), transport.WithGRPC(g))
if err != nil {
    panic(err)
}
if err := srv.Start(ctx); err != nil {
    panic(err)
}
```

## Example (HTTP + gRPC with AuthN/AuthZ)

```go
validator := auth.NewValidator(authConfig)

httpPolicies := authz.NewPolicyMap()
httpPolicies.Set(authz.HTTPOperation("GET", "/api/orders"), authz.RequireAny("reader", "admin"))

grpcPolicies := authz.NewPolicyMap()
grpcPolicies.Set(authz.GRPCOperation("/orders.v1.Orders/GetOrder"), authz.RequireAny("reader", "admin"))

h := transporthttp.NewServer(
    transporthttp.NewConfiguration(),
    transporthttp.WithAuthMiddleware(validator, httpPolicies),
)

g := transportgrpc.NewServer(func(s *grpc.Server) {
    pb.RegisterOrdersServer(s, impl)
}, transportgrpc.AuthServerOptions(validator, grpcPolicies)...)

srv, err := transport.NewServer(cfg, transport.WithHTTP(h), transport.WithGRPC(g))
if err != nil {
    panic(err)
}
if err := srv.Start(ctx); err != nil {
    panic(err)
}
```
