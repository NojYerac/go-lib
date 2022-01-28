package auth

import "context"

type key struct{}

var userCtxKey = new(key)

func WithCtx(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, userCtxKey, user)
}

func FromCtx(ctx context.Context) (*User, error) {
	user, ok := ctx.Value(userCtxKey).(*User)
	if ok {
		return user, nil
	}
	return nil, new(ErrUnauthenticated)
}
