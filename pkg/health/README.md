# Health Package

The `health` package provides a simple framework for performing health checks within a Go application. It exposes a `Checker` interface that can register custom checks and periodically run them, reporting a combined health status.

## Configuration

Health checks are configured via the `Configuration` struct in `pkg/health/config.go`. The struct contains:

- `CheckInterval`: how often the checker runs. This is a `time.Duration` and must be at least one second. The default value is `30 * time.Second`.

Example:

```go
cfg := health.NewConfiguration()
cfg.CheckInterval = 10 * time.Second
```

The struct uses the `config` struct tag to bind to configuration files or environment variables.

## Usage

```go
import (
    "context"
    "time"
    "github.com/nojyerac/go-lib/pkg/health"
    "github.com/rs/zerolog"
)

func main() {
    // Create logger for the checker
    logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
    ctx := logger.WithContext(context.Background())

    // Build the checker
    cfg := health.NewConfiguration()
    checker := health.NewChecker(cfg)

    // Register a custom health check
    checker.Register("database", func(ctx context.Context) error {
        // Perform health test – return nil on success, error on failure
        return nil
    })

    // Start the checker in a goroutine
    go func() {
        if err := checker.Start(ctx); err != nil && err != context.Canceled {
            logger.Error().Err(err).Msg("health check failed")
        }
    }()

    // Use checker.Passed() or checker.String() to examine health state
    if checker.Passed() {
        logger.Info("system is healthy")
    }

    // Run for a short period then cancel
    time.Sleep(5 * time.
}
```

## Examples

