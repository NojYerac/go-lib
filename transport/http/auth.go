package http

import (
	"net/http"

	"github.com/nojyerac/go-lib/auth"
	"github.com/nojyerac/go-lib/authz"
)

func WithAuthMiddleware(validator auth.Validator, policies authz.PolicyMap) Option {
	return WithMiddleware(authMiddleware(validator, policies))
}

func authMiddleware(validator auth.Validator, policies authz.PolicyMap) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requirement, ok := policies.Requirement(authz.HTTPOperation(r.Method, r.URL.Path))
			if !ok {
				next.ServeHTTP(w, r)
				return
			}

			token, err := auth.BearerToken(r.Header.Get("Authorization"))
			if err != nil {
				writeAuthError(w, err)
				return
			}

			claims, err := validator.Validate(r.Context(), token)
			if err != nil {
				writeAuthError(w, err)
				return
			}

			if err := authz.Authorize(claims, requirement); err != nil {
				writeAuthError(w, err)
				return
			}

			ctx := auth.WithClaims(r.Context(), claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func writeAuthError(w http.ResponseWriter, err error) {
	status := auth.HTTPStatus(err)
	w.WriteHeader(status)
	_, _ = w.Write([]byte(http.StatusText(status)))
}
