# db Package

The `db` package provides utilities for configuring and initializing a database client. It supports PostgreSQL as the default driver and offers flexible options for logging and health checking.

---

## Configuration

Configuration is defined in `pkg/db/config.go`. The struct contains fields for driver, connection string, pool limits and idle settings.

[`db.Configuration`](pkg/db/config.go:11)

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

Create a default configuration using:

```go
cfg := db.NewConfiguration()
```

The returned configuration defaults to the PostgreSQL driver:

```go
Driver: "postgres"
```

---

## Options

The package exposes functional options to customize behaviour.

- [`WithLogger`](pkg/db/config.go:28): injects a `zerolog.Logger` instance.
- [`WithHealthChecker`](pkg/db/config.go:34): injects a health checker that implements the `health.Checker` interface.

## Usage

A typical usage pattern is:

```go
import (
    "os"

    "github.com/nojyerac/go-lib/pkg/db"
    "github.com/rs/zerolog"
    "github.com/nojyerac/go-lib/pkg/health"
)

func main() {
    cfg := db.NewConfiguration()
    
    // use config loader to read in vars
    // create logger & health checker

    database, err := db.NewDatabse(cfg, db.WithLogger(l), db.WithHealthChecker(h))
    if err != nil {
        // handle error
    }
    go func() {
        if err = database.Start(ctx); err != nil {
            // handle error
        }
    }

    // use database connection
}
```

## Examples

- **Connecting to a PostgreSQL database** – see `examples/postgres/main.go` (if present).
- **Using a mock client** – the `internal/mocks/db` package contains a mock implementation that can be used in tests.
