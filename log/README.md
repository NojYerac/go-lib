# Log Package

The `log` package builds structured service loggers and provides context-based
logger propagation.

## Configuration

```go
type Configuration struct {
    ServiceName  string `config:"service_name" validate:"required"`
    HumanFrendly bool   `config:"human_friendly"`
    LogLevel     string `config:"log_level" flag:"logLevel,l" validate:"oneof=trace debug info warn error fatal panic"`
}
```

`NewConfiguration()` defaults:

- `ServiceName`: `version.GetVersion().Name`
- `LogLevel`: `info`

## API

- `NewLogger(config, opts...) logrus.FieldLogger`
- `WithOutput(io.Writer)`
- `Nop() *logrus.Logger`
- `WithLogger(ctx, logger) context.Context`
- `FromContext(ctx) logrus.FieldLogger`
- `SetDefaultCtxLogger(logger)`
- `TestConfig` (ready-made config for tests)

## Example

```go
cfg := log.NewConfiguration()
cfg.ServiceName = "orders"
cfg.LogLevel = "debug"

l := log.NewLogger(cfg)
log.SetDefaultCtxLogger(l)

ctx := log.WithLogger(context.Background(), l)
log.FromContext(ctx).WithField("request_id", "abc-123").Info("received request")
```