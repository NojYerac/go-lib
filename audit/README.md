# Audit Package

The `audit` package provides reusable audit event primitives and contracts for
writing and querying audit records consistently across services.

## API

### Event model

```go
type Event struct {
    ID        string
    Action    string
    Actor     Actor
    Resource  Resource
    Timestamp time.Time
    Details   map[string]any
}
```

Validation helper:

- `ValidateEvent(event, cfg)`

### Bounded payload helper

- `MarshalBoundedJSON(payload, maxBytes)` marshals payload and enforces a max
  serialized size.

### Compact diff helper

- `CompactDiff(before, after)` returns only changed/added/removed keys as
  `map[string]Change`.

### Interfaces

```go
type Writer interface {
    Append(ctx context.Context, event *Event, opts AppendOptions) error
}

type Reader interface {
    List(ctx context.Context, opts *ListOptions) (ListResult, error)
}
```

The writer includes `AppendOptions.TransactionID` so callers can carry
transaction context when persisting audit events atomically with business
mutations.

### In-memory implementation

- `NewMemoryStore(cfg)` provides an in-memory `Writer` + `Reader`
  implementation for tests and package-level validation.
- `List` supports deterministic ordering (`asc` / `desc`) and cursor
  pagination.

## Example: mutation + audit write

```go
cfg := audit.NewConfiguration()
store := audit.NewMemoryStore(cfg)

event := &audit.Event{
    ID:     "evt-1",
    Action: "order.update",
    Actor: audit.Actor{Type: "user", ID: "u-1"},
    Resource: audit.Resource{Type: "order", ID: "o-1"},
    Timestamp: time.Now().UTC(),
    Details: map[string]any{"status": "approved"},
}

// inside the same application transaction boundary:
err := store.Append(ctx, event, audit.AppendOptions{TransactionID: txID})
if err != nil {
    return err
}
```

## Example: query with pagination

```go
res, err := store.List(ctx, &audit.ListOptions{
    Filter: audit.Query{Action: "order.update"},
    Page:   audit.Page{Limit: 50, Order: audit.OrderDesc},
})
if err != nil {
    return err
}

nextCursor := res.PageInfo.NextCursor
_ = nextCursor
```
