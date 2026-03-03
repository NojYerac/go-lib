# Technical Overview: audit package

## Architecture

The `audit` package implements a producer-consumer pattern designed for high-throughput, resilient event logging.

### Component Diagram

```text
[ Application ]
       |
       v
[ AuditLogger ] ----> [ Publisher (Buffer) ]
                             |
                             v
                      [ Dispatcher (Orchestrator) ]
                             |
           +-----------------+-----------------+
           |                 |                 |
      [ Sink A ]        [ Sink B ]        [ Sink C ]
```

## Primitives in Detail

### Publisher (The Buffer)
The `Publisher` interface defines how events are handed off. The `MemoryPublisher` uses a Go channel with a configurable buffer size. When the buffer is full, it implements a FIFO eviction policy for its internal slice storage to maintain a history of the most recent events while preventing backpressure from blocking the main application flow.

### Dispatcher (The Orchestrator)
The `DefaultDispatcher` is the engine of the package. It starts a configurable number of workers (goroutines) that listen for events from the `Publisher`.

- **Batching Strategy**: Workers collect events until a `batchSize` is reached or a `batchTimeout` expires. This reduces the number of network round-trips to external Sinks.
- **Error Handling**: The dispatcher distinguishes between transient errors (retriable) and permanent errors (non-retriable). It uses an exponential backoff strategy to avoid overwhelming a struggling Sink.
- **Concurrency**: Each Sink receives a copy of the batch. Deliveries to different Sinks happen in parallel.

### Sink (The Destination)
Sinks are decoupled from the logging logic. 
- **Graceful Degradation**: If a Sink doesn't support the `BatchingSink` interface, the Dispatcher automatically falls back to sending events individually.
- **Context Awareness**: All Sink calls are context-aware, allowing for timeouts and cancellations.

## Design Decisions (ADR Summary)

1. **Interface-First Design**: By defining `Publisher`, `Dispatcher`, and `Sink` as interfaces, the package allows users to swap out the in-memory buffer for a persistent queue (like NATS or Redis) without changing the `AuditLogger` implementation.
2. **Mandatory Validation**: The `AuditLogger` uses `go-playground/validator` to ensure all events have a valid UUID `ActorID` and required fields before they enter the pipeline.
3. **Async by Default**: To ensure that audit logging never becomes a bottleneck for the primary application logic, the handoff to the `Publisher` is the only synchronous part of the flow.
