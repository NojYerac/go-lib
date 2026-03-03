package audit_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/nojyerac/go-lib/audit"
	"github.com/nojyerac/go-lib/auth"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type mockAuditLogger struct {
	loggedActorID string
	loggedAction  string
	loggedDetails map[string]any
	callCount     int
}

func (m *mockAuditLogger) LogChange(ctx context.Context, actorID, action string, before, after map[string]any) error {
	return nil
}

func (m *mockAuditLogger) Log(ctx context.Context, actorID, action string, details map[string]any) error {
	m.loggedActorID = actorID
	m.loggedAction = action
	m.loggedDetails = details
	m.callCount++
	return nil
}

var _ = Describe("Audit Middleware", func() {
	var (
		logger  *mockAuditLogger
		handler http.Handler
	)

	BeforeEach(func() {
		logger = &mockAuditLogger{}
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})
		handler = audit.Middleware(logger)(nextHandler)
	})

	It("should log successful mutation requests with actor from context", func() {
		claims := &auth.Claims{Subject: "user-123"}
		req := httptest.NewRequest(http.MethodPost, "/api/resource", nil)
		ctx := auth.WithClaims(req.Context(), claims)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		Expect(rr.Code).To(Equal(http.StatusNoContent))
		Expect(logger.callCount).To(Equal(1))
		Expect(logger.loggedActorID).To(Equal("user-123"))
		Expect(logger.loggedAction).To(Equal("POST:/api/resource"))
		Expect(logger.loggedDetails["method"]).To(Equal("POST"))
		Expect(logger.loggedDetails["path"]).To(Equal("/api/resource"))
	})

	It("should use 'system' as actor when no claims in context", func() {
		req := httptest.NewRequest(http.MethodPut, "/api/resource/1", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		Expect(rr.Code).To(Equal(http.StatusNoContent))
		Expect(logger.callCount).To(Equal(1))
		Expect(logger.loggedActorID).To(Equal("system"))
	})

	It("should not log GET requests", func() {
		req := httptest.NewRequest(http.MethodGet, "/api/resource", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		Expect(rr.Code).To(Equal(http.StatusNoContent))
		Expect(logger.callCount).To(Equal(0))
	})

	It("should not log failed requests", func() {
		failHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
		})
		handler = audit.Middleware(logger)(failHandler)

		req := httptest.NewRequest(http.MethodPost, "/api/resource", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		Expect(rr.Code).To(Equal(http.StatusBadRequest))
		Expect(logger.callCount).To(Equal(0))
	})
})
