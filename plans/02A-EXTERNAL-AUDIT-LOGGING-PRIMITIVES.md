# Plan 02A: External Audit Logging Primitives

## Problem

Current audit primitives focus on in-service persistence patterns. This is useful for local visibility but does not provide strong tamper resistance, cross-service centralization, or compliance-grade retention controls.

## Goal

Provide reusable primitives that let services publish audit events to external systems reliably, with deterministic delivery semantics, bounded payloads, and operational guardrails.

## Scope

- Define a transport-agnostic `audit.Publisher` abstraction for external sinks.
- Add outbox primitives for transaction-safe event staging.
- Add idempotency and ordering helpers for at-least-once delivery.
- Define standard metadata fields required for compliance usage.
- Provide retry, backoff, and dead-letter handling contracts.
- Provide guidance for PII minimization and payload redaction hooks.

## Out of Scope

- Vendor-specific sink adapters (e.g., Splunk, Datadog, CloudTrail).
- End-user retention policy management in external platforms.
- Service-specific HTTP/gRPC API behavior.

## Deliverables

- New package contracts (or extensions under `audit/`):
  - `Publisher` interface for external dispatch.
  - `OutboxStore` interface for append + lease/claim + ack/fail.
  - `Dispatcher` worker abstraction with retries and DLQ routing.
- Standard event envelope:
  - `event_id` (globally unique)
  - `schema_version`
  - `occurred_at` (UTC)
  - `actor`, `action`, `resource`, `tenant` (where applicable)
  - `trace_id` / `request_id`
  - bounded `details` payload
- Utilities:
  - deterministic event hashing helper
  - payload size validator + redaction callback
  - idempotency key helper (`event_id`-based)
- In-memory and mock implementations for package tests.
- Documentation: outbox pattern, failure modes, replay, and operational tuning.

## Suggested Milestones

1. **B1 – Envelope and validation**
   - Add envelope fields and schema versioning rules.
   - Enforce payload bounds and UTC timestamp normalization.
2. **B2 – Outbox contract and reference implementation**
   - Define durable staging interface and claim/ack lifecycle.
   - Provide transaction-friendly append pattern.
3. **B3 – Dispatcher and resilience**
   - Add retry policy, backoff, max-attempt handling, and DLQ interface.
   - Ensure safe restart/replay behavior.
4. **B4 – Idempotency and observability hooks**
   - Add duplicate suppression contract and delivery metrics/tracing hooks.
5. **B5 – Examples and docs**
   - Provide cookbook for mutation + outbox append + async publish.

## Acceptance Criteria

- Library supports atomic mutation + outbox append pattern.
- External delivery is at-least-once with idempotency support.
- Event schema is versioned and backward-compatible by contract.
- Payload bounds and redaction are enforceable before dispatch.
- Retry exhaustion routes to DLQ without dropping events silently.
- Examples demonstrate deterministic ordering strategy per resource.

## Industry Best Practices Captured

- Prefer append-only outbox + asynchronous publish over direct synchronous sink writes.
- Treat delivery as at-least-once and require idempotent consumers.
- Use immutable event identifiers and schema versioning from day one.
- Minimize sensitive data in `details`; redact/tokenize before externalization.
- Encrypt in transit and assume sink-side encryption at rest.
- Emit audit pipeline metrics (`queued`, `published`, `retried`, `failed`, `dlq`).
- Document replay and backfill runbooks for incident recovery.

## Risks / Notes

- Misconfigured retries can create sink amplification; cap attempts and use jitter.
- Ordering is best-effort globally; guarantee determinism per resource key where needed.
- Payload growth pressure increases cost; keep compact diff-first details.
