package auth

import "context"

type ctxClaimsKeyType struct{}

var ctxClaimsKey = ctxClaimsKeyType{}

func WithClaims(ctx context.Context, claims *Claims) context.Context {
	return context.WithValue(ctx, ctxClaimsKey, claims)
}

func FromContext(ctx context.Context) (*Claims, bool) {
	if ctx == nil {
		return nil, false
	}
	claims, ok := ctx.Value(ctxClaimsKey).(*Claims)
	if !ok || claims == nil {
		return nil, false
	}
	return claims, true
}
