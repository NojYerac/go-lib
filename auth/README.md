# Auth Package

The `auth` package provides reusable authentication primitives for services using
`go-lib`.

## Configuration

```go
type Configuration struct {
    Issuer     string        `config:"auth_issuer" validate:"required"`
    Audience   string        `config:"auth_audience" validate:"required"`
    HMACSecret string        `config:"auth_hmac_secret" validate:"required"`
    ClockSkew  time.Duration `config:"auth_clock_skew" validate:"min=0s"`
}
```

`NewConfiguration()` defaults `ClockSkew` to `30s`.

## API

- `NewValidator(config, opts...) Validator`
- `WithNow(func() time.Time) Option` (useful for deterministic tests)
- `WithClaims(ctx, claims) context.Context`
- `FromContext(ctx) (*Claims, bool)`
- `(*Claims).HasRole(role)`
- `(*Claims).HasAnyRole(roles...)`
- `HTTPStatus(err) int`
- `GRPCCode(err) codes.Code`

## Error Mapping

The package exposes common auth errors and transport-neutral mapping helpers.

- `ErrMissingToken`, `ErrInvalidToken`, `ErrTokenExpired`
  - HTTP: `401 Unauthorized`
  - gRPC: `Unauthenticated`
- `ErrPermissionDenied`
  - HTTP: `403 Forbidden`
  - gRPC: `PermissionDenied`

## Example

```go
cfg := auth.NewConfiguration()
cfg.Issuer = "issuer.test"
cfg.Audience = "orders-service"
cfg.HMACSecret = os.Getenv("AUTH_HMAC_SECRET")

validator := auth.NewValidator(cfg)
claims, err := validator.Validate(ctx, bearerToken)
if err != nil {
    status := auth.HTTPStatus(err)
    _ = status
    return
}

ctx = auth.WithClaims(ctx, claims)
if claims.HasAnyRole("admin", "writer") {
    // allow write operation
}
```
