# Audit Package

The `audit` package now exposes a minimal API:

```go
logger, err := audit.NewAuditLogger(audit.NewConfiguration())
if err != nil {
    return err
}

err = logger.Log(ctx, user.UUID, action, details)
```

## API

```go
type AuditLogger interface {
    Log(ctx context.Context, actorID, action string, details map[string]any) error
    LogChange(ctx context.Context, actorID, action string, before, after map[string]any) error
}

func NewAuditLogger(cfg *Configuration, options ...Option) (AuditLogger, error)

func WithOutput(output io.Writer) Option
func WithTimeNow(timeNowFunc func() time.Time) Option
func WithHTTPBaseURL(baseURL string) Option
func WithHTTPClient(client *http.Client) Option
```

`Configuration` fields:

- `AuditLoggerType`: `noop`, `stdout`, or `http`
- `AuditLoggerURL`: required for `http` logger
- `MaxPayloadBytes`: max JSON byte size for `details` payload (default `4096`)

## Current implementation

Supported logger types (`Configuration.AuditLoggerType`):

- `noop`
- `stdout`
- `http`

`stdout` pretty-prints each audit event as indented JSON to stdout by default.
Use `WithOutput(...)` to redirect output to a custom `io.Writer`.
Use `WithTimeNow(...)` to control generated timestamps, mostly for testing.
`Log(...)` enforces validation and payload size limits before writing.

`http` sends each audit event as JSON with `POST` to:

- `<AuditLoggerURL>/api/auditlog`

Use `WithHTTPBaseURL(...)` to override `AuditLoggerURL`, and
`WithHTTPClient(...)` to provide a custom client.
`Log(...)` enforces validation and payload size limits before posting.

No-op behavior:

- `NewAuditLogger(...)` returns a logger that accepts all calls.
- `Log(...)` always returns `nil`.

This keeps integration friction low while allowing future implementations to be
added behind the same constructor and interface.
