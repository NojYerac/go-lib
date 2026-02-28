````markdown
# Transport gRPC Package

The `transport/grpc` package provides opinionated gRPC client/server setup with
telemetry, logging, metrics, and panic recovery interceptors.

## API

### Client

- `ClientConn(target string, opts ...ClientOpt) (*grpc.ClientConn, error)`
- `WithDialOptions(...grpc.DialOption)`
- `WithHealthChecker(health.Checker)`

When no dial options are provided, insecure transport credentials are used.

If a health checker is provided, a `grpc_client` check is registered and passes
when connection state is `Ready` or `Idle`.

### Server

- `NewServer(registerServices func(*grpc.Server), opts ...grpc.ServerOption) *grpc.Server`
- `SetLogger(logrus.FieldLogger)`
- `AuthServerOptions(auth.Validator, authz.PolicyMap) []grpc.ServerOption`
- `AuthUnaryServerInterceptor(auth.Validator, authz.PolicyMap) grpc.UnaryServerInterceptor`
- `AuthStreamServerInterceptor(auth.Validator, authz.PolicyMap) grpc.StreamServerInterceptor`

`NewServer` applies:

- OpenTelemetry gRPC handlers
- Prometheus interceptors
- logging interceptors
- panic recovery interceptors

`AuthServerOptions` provides auth interceptors for unary and stream RPCs.
Only RPCs present in the policy map are enforced. Missing/invalid tokens map to
`Unauthenticated`; failed role checks map to `PermissionDenied`.

## Example

```go
grpcSrv := transportgrpc.NewServer(func(s *grpc.Server) {
    pb.RegisterMyServiceServer(s, impl)
}, transportgrpc.AuthServerOptions(validator, policies)...)

conn, err := transportgrpc.ClientConn(
    "localhost:8080",
    transportgrpc.WithDialOptions(grpc.WithTransportCredentials(insecure.NewCredentials())),
)
if err != nil {
    panic(err)
}
defer conn.Close()
```

````
