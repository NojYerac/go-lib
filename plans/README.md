# go-lib Plans

This directory tracks implementation plans for missing library functionality.

## Active Plans

1. [01-AUTHN-AUTHZ-PRIMITIVES.md](./01-AUTHN-AUTHZ-PRIMITIVES.md)
2. [02-AUDIT-EVENT-PRIMITIVES.md](./02-AUDIT-EVENT-PRIMITIVES.md)
3. [02A-EXTERNAL-AUDIT-LOGGING-PRIMITIVES.md](./02A-EXTERNAL-AUDIT-LOGGING-PRIMITIVES.md)
4. [03-OPERATIONS-HARDENING.md](./03-OPERATIONS-HARDENING.md)
5. [04-SCAFFOLD-NEW-SERVICE.md](./04-SCAFFOLD-NEW-SERVICE.md)
6. [05-CI-RELEASE-TEMPLATES-AND-SDK-UTILS.md](./05-CI-RELEASE-TEMPLATES-AND-SDK-UTILS.md)
7. [06-HTTP-CLIENT-PRIMITIVES.md](./06-HTTP-CLIENT-PRIMITIVES.md)

## Prioritization

Recommended implementation order:

1. AuthN/AuthZ primitives
2. Audit event primitives
3. External audit logging primitives
4. Operations hardening
5. One-click service scaffolding
6. CI/release templates and SDK support utilities
7. HTTP client primitives

## Notes

- Plans are designed to be transport-agnostic where possible.
- Every plan should include tests, examples, and documentation updates.
- Keep package APIs small, stable, and reusable across services.
