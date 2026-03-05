// Package transport registers example-specific HTTP routes.
// Add gRPC service registrations here as the service grows.
package transport

import (
	libhttp "github.com/nojyerac/go-lib/transport/http"
	"net/http"
)

// RegisterRoutes mounts example's application-level routes onto the shared
// HTTP server.  The go-lib HTTP server automatically wires /livez, /healthz,
// /metrics, and /version — add only business routes here.
func RegisterRoutes(s libhttp.Server) {
	s.HandleFunc("GET /hello", helloHandler)
}

func helloHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"service":"example","status":"ok"}`))
}
