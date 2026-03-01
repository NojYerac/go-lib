# Audit Package

The `audit` package now exposes a minimal API:

```go
logger, err := audit.NewAuditLogger(audit.NewConfiguration())
if err != nil {
    return err
}

err = logger.Log(ctx, "user.login", map[string]any{"user_id": "u-1"})
```

## API

```go
type AuditLogger interface {
    Log(ctx context.Context, action string, details map[string]any) error
}

func NewAuditLogger(cfg *Configuration, options ...Option) (AuditLogger, error)

func WithOutput(output io.Writer) Option
```

## Current implementation

Supported logger types (`Configuration.AuditLoggerType`):

- `noop`
- `stdout`

`stdout` pretty-prints each audit event as indented JSON to stdout by default.
Use `WithOutput(...)` to redirect output to a custom `io.Writer`.

No-op behavior:

- `NewAuditLogger(...)` returns a logger that accepts all calls.
- `Log(...)` always returns `nil`.

This keeps integration friction low while allowing future implementations to be
added behind the same constructor and interface.
