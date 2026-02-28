# Plan 03: Operations Hardening

## Problem

`go-lib` provides core transport wrappers but lacks several production safeguards
such as explicit limits, rate limiting, and readiness composition helpers.

## Goal

Make runtime defaults safer and reduce service-specific reliability boilerplate.

## Scope

- Add HTTP request size and timeout controls.
- Add gRPC request timeout and rate-limiting interceptors.
- Provide shared rate-limiter package.
- Provide startup/readiness dependency composition helpers.

## Out of Scope

- Multi-region deployment architecture.
- Autoscaling policy management.

## Deliverables

- `transport/http` options for:
  - max request body bytes
  - read/write/idle timeouts
  - optional per-route timeout middleware
- `transport/grpc` interceptors for:
  - max handler duration
  - rate limiting
- New `ratelimit/` package (token bucket + identity key function).
- New `lifecycle/` package for readiness gates and startup dependency checks.
- Tests and operational docs/examples.

## Suggested Milestones

1. **C1** HTTP limits + tests.
2. **C2** gRPC limits + tests.
3. **C3** shared `ratelimit` package + tests/benchmarks.
4. **C4** `lifecycle` package + docs.

## Acceptance Criteria

- Limits and timeouts are configurable with safe defaults.
- Rate limiting works in both transports with test coverage.
- Services can compose readiness checks without custom framework code.

## Risks / Notes

- Roll out limits conservatively to avoid accidental client breakage.
- Document default values and override guidance clearly.
