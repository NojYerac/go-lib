package audit

import (
	"context"
)

// Publisher defines the interface for publishing audit events to a persistent store or queue.
type Publisher interface {
	Publish(ctx context.Context, evt event) error
}

// Dispatcher defines the interface for dispatching audit events to external sinks.
type Dispatcher interface {
	// Start begins the background dispatching process.
	Start(ctx context.Context) error
	// Stop gracefully shuts down the dispatcher.
	Stop(ctx context.Context) error
}

// Sink defines the interface for an external destination for audit events.
type Sink interface {
	Send(ctx context.Context, evt event) error
	Name() string
}

// BatchingSink is an optional interface for sinks that support batch delivery.
type BatchingSink interface {
	Sink
	SendBatch(ctx context.Context, events []event) error
}

// IsTransientError returns true if the error is considered transient (e.g., 5xx).
// TODO: Implement proper HTTP status code checking for production use.
func IsTransientError(err error) bool {
	if err == nil {
		return false
	}
	return true
}
