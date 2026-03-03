# Plan 02: Audit Event Primitives (DONE)

## Status
**Completed on 2026-03-03**

## Problem

Services need audit logs but currently lack standardized event schemas and reusable write interfaces.

## Goal

Provide reusable audit building blocks so services can emit consistent audit records.

## Scope

- Define audit event schema.
- Provide a simple logger interface and implementations.
- Support bounded details payloads.

## Out of Scope

- SIEM streaming integrations.
- Long-term archival systems.

## Deliverables

- [x] New `audit/` package:
  - [x] event model (`actor`, `action`, `timestamp`, `details`)
  - [x] logger interface (`Log`, `LogChange`)
  - [x] bounded payload enforcement in logger path
- [x] Concrete implementations:
  - [x] no-op logger
  - [x] stdout logger
  - [x] HTTP logger (`POST /api/auditlog`)
- [x] Documentation and examples for mutation + audit write patterns.

## Suggested Milestones

1. **B1** event schema + validation. (DONE)
2. **B2** logger interface + no-op/stdout/http implementations. (DONE)
3. **B3** transport cookbook examples (HTTP/gRPC handler usage). (DONE)

## Acceptance Criteria

- Audit event shape is consistent across services.
- Details payload size limits are configurable and enforced.
- Examples demonstrate audit logging flow with minimal integration overhead.

## Risks / Notes

- Keep payload growth bounded to avoid oversized records.
- Keep logger API simple; sidecar handles persistence/retry complexity.
