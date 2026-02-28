# Tracing Package

The `tracing` package configures OpenTelemetry tracing providers and exposes a
convenient package-based tracer helper.

## Configuration

```go
type Configuration struct {
    ExporterType string  `config:"exporter_type" validate:"oneof=stdout file otlp noop"`
    FilePath     string  `config:"file_path" validate:"required_if=ExporterType file"`
    OTLPURL      string  `config:"otlp_url" validate:"required_if=ExporterType otlp"`
    SampleRatio  float64 `config:"sample_ratio" validate:"gte=0,lte=1"`
}
```

## Exporter Modes

- `stdout`: pretty-printed spans to stdout
- `file`: spans written to `FilePath`
- `otlp`: OTLP HTTP exporter (`OTLPURL`, insecure mode)
- `noop`: no-op tracer provider

## API

- `NewTracerProvider(config) trace.TracerProvider`
- `SetGlobal(tp trace.TracerProvider)`
- `TracerForPackage(skipMore ...int) trace.Tracer`

## Example

```go
tp := tracing.NewTracerProvider(&tracing.Configuration{
    ExporterType: "otlp",
    OTLPURL:      "localhost:4318",
    SampleRatio:  0.25,
})
tracing.SetGlobal(tp)

tracer := tracing.TracerForPackage()
ctx, span := tracer.Start(ctx, "work")
defer span.End()
```
