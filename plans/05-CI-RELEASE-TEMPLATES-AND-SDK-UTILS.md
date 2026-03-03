# Plan 05: CI/Release Templates and SDK Support Utilities (PLANNED / STALLED)

## Status
**Pending Implementation** (Identified as missing in Minerva Audit 2026-03-03; `templates/` directory and SDK helper packages not found)

## Problem

Service repos repeatedly reinvent CI workflows and client behavior patterns
(retries, auth attachment, error normalization).

## Goal

Standardize service delivery automation and client-side integration primitives.

## Scope

- Provide reusable GitHub Actions templates.
- Provide shared utilities for SDK/client implementations.
- Include guidance for adopting templates/utilities in new and existing services.

## Out of Scope

- Cloud-specific deployment workflows.
- Non-Go SDK implementations.

## Deliverables

- Template files under `templates/github/`:
  - `ci.yml` (lint/test/build)
  - `release.yml` (tag-based artifacts)
- New helper package for SDK/client behavior:
  - retry/backoff helpers
  - auth token/header provider interfaces
  - HTTP/gRPC error normalization helpers
- Examples showing usage from scaffolded services.
- Documentation on expected branch protection and release tagging flow.

## Suggested Milestones

1. **E1** baseline CI/release templates.
2. **E2** reusable client utility package.
3. **E3** integration examples in scaffold output.

## Acceptance Criteria

- New services can adopt CI/release workflows without bespoke setup.
- Shared client utilities reduce duplicated integration code.
- Example service wiring demonstrates default usage patterns.

## Risks / Notes

- Keep initial templates minimal and fast to keep CI stable.
- Document compatibility policy for helper package APIs.
