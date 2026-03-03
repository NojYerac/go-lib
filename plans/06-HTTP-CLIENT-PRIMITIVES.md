# Plan 06: HTTP Client Primitives

## Problem

Services repeatedly rebuild HTTP client concerns (retry behavior, TLS/CA wiring,
timeouts, middleware hooks), causing inconsistent reliability and security.

## Goal

Provide a reusable `httpclient` package in `go-lib` with safe defaults and
opt-in extensibility for service-specific policies.

## Scope

- Add a new package for HTTP client construction and request execution.
- Support pluggable/custom retry strategies:
  - retryable status-code policies
  - retryable transport error policies
  - configurable backoff and max-attempt limits
- Support transport hardening features:
  - custom CA certificates
  - mTLS hooks (client cert/key or tls.Config integration)
  - sane timeout defaults with per-call overrides
- Support request lifecycle hooks/middleware for:
  - logging
  - tracing propagation
  - metrics
  - auth header injection
- Keep API simple and transport-focused; avoid service/domain coupling.

## Out of Scope

- Full circuit-breaker implementation in v1.
- HTTP server middleware.
- Service-specific auth/token refresh workflows.

## Deliverables

- New package (proposed path: `httpclient/`) with:
  - `Configuration` for client defaults
  - constructor for `*http.Client` with secure defaults
  - retry policy interfaces and built-in strategies
  - middleware/hook interfaces for request/response/error lifecycle
  - CA/mTLS configuration helpers
- Unit tests for retry, TLS config, and middleware execution order.
- Example usage docs for common service integration patterns.

## Suggested Milestones

1. **E1** baseline client constructor + secure defaults.
2. **E2** retry strategy interfaces + built-ins.
3. **E3** custom CA/mTLS support.
4. **E4** middleware/hook chain + observability examples.

## Acceptance Criteria

- Consumers can construct clients with default-safe timeouts and TLS settings.
- Consumers can plug in custom retry policies without forking core logic.
- Consumers can configure custom CA trust and optional mTLS.
- Middleware/hook chain enables logging/tracing/metrics/auth composition.
- Docs include at least one end-to-end example with retries and custom CA.

## Risks / Notes

- Keep retry defaults conservative to avoid thundering-herd retries.
- Ensure idempotency guidance is explicit for retried requests.
- Prefer composable interfaces over large monolithic client wrappers.
