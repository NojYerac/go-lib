# go-lib Roadmap

This file is the high-level index for roadmap work. Detailed execution plans
are now split into individual files under [plans](./plans/README.md).

## Active Plan Set

1. [x] [Plan 01: AuthN/AuthZ Primitives](./plans/01-AUTHN-AUTHZ-PRIMITIVES.md)
2. [x] [Plan 02: Audit Event Primitives](./plans/02-AUDIT-EVENT-PRIMITIVES.md)
3. [ ] [Plan 02A: External Audit Logging Primitives](./plans/02A-EXTERNAL-AUDIT-LOGGING-PRIMITIVES.md)
4. [ ] [Plan 03: Operations Hardening](./plans/03-OPERATIONS-HARDENING.md)
5. [ ] [Plan 04: One-Click Scaffold a New Service](./plans/04-SCAFFOLD-NEW-SERVICE.md)
6. [ ] [Plan 05: CI/Release Templates and SDK Utilities](./plans/05-CI-RELEASE-TEMPLATES-AND-SDK-UTILS.md)
7. [ ] [Plan 06: HTTP Client Primitives](./plans/06-HTTP-CLIENT-PRIMITIVES.md)

## Definition of Done (Roadmap)

The roadmap is complete when:

- Auth, audit, hardening, and client utility primitives are reusable in at
  least one consuming service without copy-paste logic.
- External audit delivery primitives (outbox + dispatcher + idempotency
  helpers) are reusable for at-least-once delivery workflows.
- Scaffold command generates a runnable service in one command.
- CI/release templates are used by newly scaffolded services.
