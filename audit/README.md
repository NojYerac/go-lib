# Audit Package

`audit` is a resilient, asynchronous audit logging library for Go. It provides a clean interface for capturing system events and state changes, with support for worker pooling, batching, and configurable retry logic.

## Overview

The package is built around four core primitives:
1. **AuditLogger**: The primary entry point for recording events.
2. **Publisher**: An interface for receiving and buffering events from the logger.
3. **Dispatcher**: A background orchestrator that manages event delivery.
4. **Sink**: The final destination for audit events (e.g., a database, API, or message queue).

## Installation

```bash
go get github.com/NojYerac/go-lib/audit
```

## Quick Start

```go
// 1. Initialize the Publisher (buffers events)
pub := audit.NewMemoryPublisher(1000)

// 2. Configure Sinks (where events go)
sinks := []audit.Sink{
    &MyCustomSink{},
}

// 3. Setup the Dispatcher (delivers events from Publisher to Sinks)
dispatcher := audit.NewDefaultDispatcher(pub, sinks)
go dispatcher.Start(ctx)

// 4. Create the Logger
logger, _ := audit.NewAuditLogger(
    audit.NewConfiguration(audit.WithLoggerType("stdout")),
    audit.WithPublisher(pub),
)

// 5. Log an event
logger.Log(ctx, "user-123", "document.deleted", map[string]any{
    "doc_id": "999",
})
```

## Core Primitives

### 1. AuditLogger
The `AuditLogger` interface provides two methods for capturing events:
- `Log(ctx, actorID, action, details)`: Logs a generic event with a key-value payload.
- `LogChange(ctx, actorID, action, before, after)`: Automatically computes the diff between two maps and logs the resulting change set.

### 2. Publisher
The `Publisher` acts as a buffer between the synchronous logging call and the asynchronous delivery.
- **MemoryPublisher**: The default implementation. It uses a bounded channel to prevent memory exhaustion and supports a circular buffer to manage overflow gracefully.

### 3. Dispatcher
The `Dispatcher` is responsible for reliable delivery. The `DefaultDispatcher` includes:
- **Worker Pool**: Parallelizes delivery across multiple goroutines.
- **Batching**: Aggregates events before sending to Sinks to improve throughput.
- **Retry Logic**: Implements exponential backoff for transient errors.
- **Graceful Shutdown**: Ensures in-flight batches are flushed before the service stops.

### 4. Sink
A `Sink` is an interface for external destinations.
- **Sink**: Basic interface with a `Send` method.
- **BatchingSink**: An optional interface. If a Sink implements `SendBatch`, the Dispatcher will use it to deliver multiple events in a single call.

## Configuration

`Configuration` fields:
- `AuditLoggerType`: `noop`, `stdout`, or `http`.
- `MaxPayloadBytes`: Maximum size for the audit event payload (default `4096`).
- `AuditLoggerURL`: Target URL for the `http` logger.

## Resilience & Performance

- **Bounded Buffers**: Prevents out-of-memory issues if Sinks are slow or down.
- **Asynchronous**: `Log` calls return immediately after handoff to the Publisher.
- **Retries**: Distinguishes between transient and permanent errors to avoid unnecessary retries.
- **Worker Management**: Configurable worker counts and batch sizes allow tuning for different workloads.
