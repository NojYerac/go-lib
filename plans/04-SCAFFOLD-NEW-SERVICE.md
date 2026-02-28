# Plan 04: One-Click Scaffold a New Service

## Problem

New services still require repetitive manual setup (config/log/transport wiring,
health endpoints, CI files, test scaffolding), slowing teams and causing drift.

## Goal

Provide a one-command generator that creates a production-ready service skeleton
already wired to `go-lib` conventions.

## Scope

- Add a scaffold CLI in `go-lib`.
- Generate runnable service structure with HTTP and gRPC bootstrap.
- Generate baseline tests, lint scripts, CI workflows, and container files.
- Support deterministic output and dry-run preview.

## Out of Scope

- Full business-domain code generation.
- Cross-language scaffolding.

## Deliverables

- New CLI package/tool: `scaffold/`.
- Command shape:
  - `go run ./scaffold --name orders --module github.com/acme/orders`
- Generated files include:
  - app entrypoint + lifecycle wiring
  - config structs + loader wiring
  - log/metrics/tracing initialization
  - HTTP and gRPC server setup
  - utility endpoints (`/livez`, `/healthz`, `/metrics`, `/version`)
  - basic test skeletons
  - lint/test scripts
  - GitHub Actions workflow templates
  - Dockerfile + run instructions
- Template system:
  - embedded templates under `scaffold/templates`
  - version metadata for scaffold output
- Golden-file tests for generator output.

## Suggested Milestones

1. **D1** template specification and output file map.
2. **D2** CLI implementation + dry-run mode.
3. **D3** golden tests for deterministic generation.
4. **D4** docs and walkthrough in top-level README.

## Acceptance Criteria

- One command produces a compilable, testable service.
- Generated service can run locally with minimal edits.
- Repeated generation with same inputs produces deterministic output.

## Risks / Notes

- Keep template surface minimal and opinionated at first.
- Version templates to support future non-breaking upgrades.
