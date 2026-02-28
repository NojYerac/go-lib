# Plan 01: AuthN/AuthZ Primitives

## Problem

`go-lib` currently has no reusable authentication/authorization layer for HTTP and gRPC.
Services must implement auth middleware/interceptors and role checks independently.

## Goal

Provide shared auth and RBAC primitives that can be dropped into any service using `go-lib` transports.

## Scope

- Add JWT validation helpers (issuer, audience, expiry).
- Add claims extraction + context propagation helpers.
- Add role-policy evaluation helpers.
- Add transport integration points for HTTP middleware and gRPC interceptors.
- Add shared auth error mapping helpers.

## Out of Scope

- External identity provider provisioning flows.
- Multi-tenant policy engines.

## Deliverables

- New `auth/` package:
  - token validator
  - claims model
  - context helpers (`WithClaims`, `FromContext`)
- New `authz/` package:
  - role policy evaluator
  - endpoint/method to required-role mapping helpers
- `transport/http` option for auth middleware registration.
- `transport/grpc` option for unary + stream auth interceptors.
- Unified error mapping:
  - HTTP: `401`, `403`
  - gRPC: `Unauthenticated`, `PermissionDenied`
- Unit and integration test coverage.

## Suggested Milestones

1. **A1** token validator + claims model + tests.
2. **A2** HTTP middleware + integration tests.
3. **A3** gRPC interceptors + integration tests.
4. **A4** shared endpoint-role mapping + docs examples.

## Acceptance Criteria

- Requests with missing/invalid tokens are rejected consistently.
- Role checks are centralized and shared across transports.
- Logs/metrics expose auth failures without leaking sensitive token contents.

## Risks / Notes

- Keep transport logic thin; centralize policy decisions in library packages.
- Avoid embedding PII in span/log attributes beyond non-sensitive identity keys.
