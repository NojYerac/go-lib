# Health Package

The `health` package provides a periodic health checker that aggregates named
checks and exposes status through a simple reporting interface.

## Configuration

```go
type Configuration struct {
    CheckInterval time.Duration `config:"healthcheck_check_interval" validate:"required,min=1s"`
}
```

`NewConfiguration()` defaults `CheckInterval` to `30s`.

## API

### `type Checker interface`

- `Start(context.Context) error`
- `Register(name string, fn CheckFn)`
- `Passed() bool`
- `String() string`

`NewChecker` always includes a built-in `ping` check.

## Behavior

- `Start` performs one immediate check, then runs on each ticker interval.
- `Passed` returns `true` only if every registered check passes.
- `String` returns one line per check, e.g. `[db_check] ok`.
- If a checker is `nil`, `Passed()` is `true` and `String()` is empty.

## Example

```go
h := health.NewChecker(health.NewConfiguration())
h.Register("db", func(ctx context.Context) error {
    return db.PingContext(ctx)
})

go func() {
    _ = h.Start(ctx)
}()

if h.Passed() {
    // ready
}
```