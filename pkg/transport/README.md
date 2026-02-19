# Transport Package

The **transport** package provides a simple HTTP server and gRPC server wrappers, along with helper functions for creating middleware and handlers.

## Configuration

```go
// pkg/transport/config.go
package transport

// Configuration holds server settings.
//
//   Host   string `config:"transport_host"`
//   Port   int    `config:"transport_port"`
//   TLS    bool   `config:"transport_tls"`
//   Cert   string `config:"transport_cert"`
//   Key    string `config:"transport_key"`
```

### HTTP Server

```go
func NewHTTPServer(cfg *Configuration) (*http.Server, error) {
    // Build and return an *http.Server based on cfg.
}
```

## Usage

```go
import (
    "github.com/nojyerac/go-lib/pkg/transport"
)

func main() {
    cfg := transport.NewConfiguration()
    srv, err := transport.NewHTTPServer(cfg)
    if err != nil {
        log.Fatal().Err(err).Msg("transport init")
    }
    log.Info().Msg("server listening")
    srv.ListenAndServe()
}
```

## Examples

- Set up a basic health‑check endpoint.
- Serve static assets.
- Add middleware for logging and metrics.
