# Tracing Package

The **tracing** package configures OpenTelemetry tracing and provides helpers to create a tracer provider and register a global tracer.

## Configuration

```go
// pkg/tracing/config.go
package tracing

type Configuration struct {
    ExporterType string  `config:"tracing_exporter"`
    SampleRatio  float64 `config:"tracing_sample_ratio"`
    FilePath     string  `config:"tracing_file_path"`
}
```

### New Tracer Provider

```go
func NewTracerProvider(cfg *Configuration) trace.TracerProvider {
    // Creates a tracer provider based on cfg.ExporterType.
}
```

## Usage

```go
import (
    "github.com/nojyerac/go-lib/pkg/tracing"
    "go.opentelemetry.io/otel"
)

func main() {
    cfg := tracing.NewConfiguration()
    tp := tracing.NewTracerProvider(cfg)
    tracing.SetGlobal(tp)
    // Tracing is now available via otel.Tracer.
}
```

## Examples

- Export traces to stdout.
- Export traces to a Jaeger collector.
- Persist traces to a file.
