# Metrics Package

The **metrics** package exposes OpenTelemetry metrics instrumentation and a Prometheus exporter.

## Configuration

```go
// pkg/metrics/config.go
package metrics

type Configuration struct {
    // No runtime configuration is required for the current implementation.
}
```

### Meter Provider

```go
func NewMeterProvider(cfg *Configuration) (metric.MeterProvider, http.HandlerFunc, error) {
    // Returns a MeterProvider and a handler to expose Prometheus metrics.
}
```

## Usage

```go
import (
    "github.com/nojyerac/go-lib/pkg/metrics"
    "net/http"
)

func main() {
    mp, promHandler, err := metrics.NewMeterProvider(nil)
    if err != nil {
        log.Fatal().Err(err).Msg("metrics init")
    }
    metrics.SetGlobal(mp)
    http.Handle("/metrics", promHandler)
    http.ListenAndServe(":9090", nil)
}
```

## Examples

- Expose metrics at `/metrics`.
- Create custom meters for a package using `metrics.MeterForPackage()`.
