# Plan 02: Audit Event Primitives

## Problem

Services need audit logs but currently lack standardized event schemas and reusable write/query interfaces.

## Goal

Provide reusable audit building blocks so services can emit and query consistent audit records.

## Scope

- Define audit event schema.
- Provide write/read interfaces and helper implementations.
- Support bounded details payloads and optional before/after diff helpers.
- Provide pagination/order contracts for reads.

## Out of Scope

- SIEM streaming integrations.
- Long-term archival systems.

## Deliverables

- New `audit/` package:
  - event model (`actor`, `action`, `resource`, `timestamp`, `details`)
  - event validation helpers
  - bounded JSON payload helper
  - optional compact diff helper
- Interfaces:
  - `Writer` (append event; transaction-aware)
  - `Reader` (paginated retrieval)
- In-memory test implementation for fast package-level tests.
- Documentation and examples for mutation + audit write patterns.

## Suggested Milestones

1. **B1** event schema + validation.
2. **B2** writer/reader interfaces + in-memory implementation.
3. **B3** pagination and deterministic ordering helpers.
4. **B4** transport cookbook examples (HTTP/gRPC handler usage).

## Acceptance Criteria

- Audit event shape is consistent across services.
- Details payload size limits are configurable and enforced.
- Query contract supports deterministic ordering and pagination.
- Examples demonstrate atomic mutation + audit recording flow.

## Risks / Notes

- Keep payload growth bounded to avoid oversized records.
- Favor compact diffs over full snapshots where practical.
