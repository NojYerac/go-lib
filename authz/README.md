# Authz Package

The `authz` package provides reusable role-based authorization primitives that
work with `auth.Claims` from the `auth` package.

## API

### Requirement constructors

- `RequireAny(roles ...string) Requirement`
- `RequireAll(roles ...string) Requirement`

Both constructors remove empty and duplicate role values.

### `type Requirement struct`

- `AnyOf []string`: user must have at least one role from this list.
- `AllOf []string`: user must have every role from this list.
- `IsEmpty() bool`
- `SatisfiedBy(claims *auth.Claims) bool`

If both `AnyOf` and `AllOf` are populated, both conditions must pass.

### Authorization helper

- `Authorize(claims *auth.Claims, requirement Requirement) error`

Returns `auth.ErrPermissionDenied` when the requirement is not satisfied.

### Policy mapping helpers

- `type PolicyMap map[string]Requirement`
- `NewPolicyMap() PolicyMap`
- `(PolicyMap).Set(operation string, requirement Requirement)`
- `(PolicyMap).Requirement(operation string) (Requirement, bool)`
- `HTTPOperation(method, path string) string`
- `GRPCOperation(fullMethod string) string`

Use operation keys such as `GET /v1/flags` (HTTP) or
`/flag.v1.FlagService/GetFlag` (gRPC).

For HTTP policies, path-parameter templates are supported in keys, including
`{id}` and `:id` segment formats (for example, `GET /v1/flags/{id}`).

## Example

```go
claims := &auth.Claims{Subject: "user-1", Roles: []string{"reader"}}

policies := authz.NewPolicyMap()
policies.Set(authz.HTTPOperation("GET", "/v1/flags"), authz.RequireAny("reader", "admin"))

if requirement, ok := policies.Requirement(authz.HTTPOperation("GET", "/v1/flags")); ok {
    if err := authz.Authorize(claims, requirement); err != nil {
        // map with auth.HTTPStatus(err) or auth.GRPCCode(err)
    }
}
```
