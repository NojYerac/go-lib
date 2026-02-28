package auth

import (
	"errors"
	"net/http"

	"google.golang.org/grpc/codes"
)

var (
	ErrMissingToken     = errors.New("missing auth token")
	ErrInvalidToken     = errors.New("invalid auth token")
	ErrTokenExpired     = errors.New("expired auth token")
	ErrPermissionDenied = errors.New("permission denied")
)

func HTTPStatus(err error) int {
	if errors.Is(err, ErrPermissionDenied) {
		return http.StatusForbidden
	}
	return http.StatusUnauthorized
}

func GRPCCode(err error) codes.Code {
	if errors.Is(err, ErrPermissionDenied) {
		return codes.PermissionDenied
	}
	return codes.Unauthenticated
}
