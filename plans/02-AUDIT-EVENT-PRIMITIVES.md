# Plan 02: Audit Event Primitives

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

- New `audit/` package:
  - event model (`actor`, `action`, `timestamp`, `details`)
  - logger interface (`Log`, `LogChange`)
  - bounded payload enforcement in logger path
- Concrete implementations:
  - no-op logger
  - stdout logger
  - HTTP logger (`POST /api/auditlog`)
- Documentation and examples for mutation + audit write patterns.

## Suggested Milestones

1. **B1** event schema + validation.
2. **B2** logger interface + no-op/stdout/http implementations.
3. **B3** transport cookbook examples (HTTP/gRPC handler usage).

## Acceptance Criteria

- Audit event shape is consistent across services.
- Details payload size limits are configurable and enforced.
- Examples demonstrate audit logging flow with minimal integration overhead.

## Risks / Notes

- Keep payload growth bounded to avoid oversized records.
- Keep logger API simple; sidecar handles persistence/retry complexity.
