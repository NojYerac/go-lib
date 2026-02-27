# Transport Package

The **transport** package provides a simple HTTP server and gRPC server wrappers, along with helper functions for creating middleware and handlers.

## Configuration

- Hostname `hostname` default `0.0.0.0`
- Port `port` default `80`
- NoTLS `no_tls` default `false`
- PubCert `tls_public_cert`
- PrivKey `tls_private_key`
- RootCA `tls_root_ca`

## Usage

```go
// create http server `h`
cfg := transport.NewConfiguration()
srv, err := transport.NewServer(cfg, transport.WithHTTP(h))
if err != nil {
    log.Fatal("transport init")
}
log.Info("server listening")
ctx, cancel = context.WithCancel(context.Background())
srv.Start(ctx)
// cancel context to stop srv
```
