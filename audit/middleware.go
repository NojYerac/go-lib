package audit

import (
	"context"
	"net/http"

	"github.com/nojyerac/go-lib/auth"
)

// Middleware returns a middleware that logs successful mutations.
// It relies on auth.FromContext to identify the actor.
func Middleware(logger AuditLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// We only audit mutations (POST, PUT, PATCH, DELETE)
			if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			// Wrap the response writer to capture the status code
			ww := &statusResponseWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(ww, r)

			// Only log successful requests (2xx)
			if ww.status >= 200 && ww.status < 300 {
				claims, ok := auth.FromContext(r.Context())
				actorID := "system"
				if ok && claims.Subject != "" {
					actorID = claims.Subject
				}

				details := map[string]any{
					"method": r.Method,
					"path":   r.URL.Path,
					"status": ww.status,
				}

				// Best effort logging
				_ = logger.Log(r.Context(), actorID, r.Method+":"+r.URL.Path, details)
			}
		})
	}
}

type statusResponseWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusResponseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}
