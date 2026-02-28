# Metrics Package

The `metrics` package sets up OpenTelemetry metrics with a Prometheus exporter.

## API

- `NewMetricProvider() (*metric.MeterProvider, http.Handler, error)`
- `SetGlobal(mp *metric.MeterProvider)`
- `MeterForPackage() api.Meter`

`NewMetricProvider` returns:

1. meter provider for instrumentation
2. HTTP handler to expose `/metrics`

## Example

```go
mp, metricsHandler, err := metrics.NewMetricProvider()
if err != nil {
    panic(err)
}
metrics.SetGlobal(mp)

httpSrv := transporthttp.NewServer(
    transporthttp.NewConfiguration(),
    transporthttp.WithMetricsHandler(metricsHandler),
)
_ = httpSrv
```

For custom instrumentation in your package:

```go
meter := metrics.MeterForPackage()
counter, _ := meter.Int64Counter("my_requests_total")
counter.Add(ctx, 1)
```
