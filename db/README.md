# DB Package

The `db` package wraps `sqlx` with:

- tracing spans per database operation
- structured operation logs
- OpenTelemetry metrics for query duration and pool state
- optional health check registration

## Configuration

```go
type Configuration struct {
    Driver          string        `config:"database_driver" validate:"required"`
    DBConnStr       string        `config:"database_connection_string" validate:"required"`
    MaxIdleConns    uint          `config:"database_max_idle_connections"`
    MaxOpenConns    uint          `config:"database_max_open_connections"`
    ConnMaxIdleTime time.Duration `config:"database_connection_max_idle_time"`
    ConnMaxLifetime time.Duration `config:"database_connection_max_life_time"`
}
```

`NewConfiguration()` defaults `Driver` to `postgres`.

## API

### `NewDatabase(config *Configuration, opts ...Option) Database`

Returns a `Database` implementation.

### `type Database interface`

- `Open(context.Context) error`
- `Close() error`
- `Begin(context.Context) (Tx, error)`
- `Select/Get/Exec/Query` (from `DataInterface`)

### `type Tx interface`

- `Commit(context.Context) error`
- `Rollback(context.Context) error`
- `Select/Get/Exec/Query`

### Options

- `WithLogger(logrus.FieldLogger)`
- `WithHealthChecker(health.Checker)`

## Metrics

When opened, the package registers:

- `database_query_duration` (ms histogram)
- `database_open_conns` (observable gauge)
- `database_idle_conns` (observable gauge)

## Example

```go
cfg := db.NewConfiguration()
cfg.DBConnStr = "postgres://user:pass@localhost:5432/app?sslmode=disable"

conn := db.NewDatabase(cfg, db.WithLogger(logger), db.WithHealthChecker(checker))
if err := conn.Open(ctx); err != nil {
    panic(err)
}
defer conn.Close()

var rows []MyRow
if err := conn.Select(ctx, &rows, "SELECT id, name FROM items"); err != nil {
    panic(err)
}

tx, err := conn.Begin(ctx)
if err != nil {
    panic(err)
}
if _, err := tx.Exec(ctx, "UPDATE items SET seen = true"); err != nil {
    _ = tx.Rollback(ctx)
    panic(err)
}
if err := tx.Commit(ctx); err != nil {
    panic(err)
}
```
