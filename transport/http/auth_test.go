package http_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"

	"github.com/nojyerac/go-lib/auth"
	"github.com/nojyerac/go-lib/authz"
	. "github.com/nojyerac/go-lib/transport/http"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type validatorStub struct {
	claims *auth.Claims
	err    error
}

func (v *validatorStub) Validate(context.Context, string) (*auth.Claims, error) {
	if v.err != nil {
		return nil, v.err
	}
	return v.claims, nil
}

var _ = Describe("Auth middleware", func() {
	var (
		s       Server
		stubVal *validatorStub
	)

	doRequest := func(req *http.Request) (int, string) {
		w := httptest.NewRecorder()
		s.ServeHTTP(w, req)
		return w.Code, w.Body.String()
	}

	BeforeEach(func() {
		stubVal = &validatorStub{claims: &auth.Claims{Subject: "user-1", Roles: []string{"reader"}}}
		policies := authz.NewPolicyMap()
		policies.Set(authz.HTTPOperation(http.MethodGet, "/api/protected"), authz.RequireAny("reader"))

		s = NewServer(
			&Configuration{},
			WithAuthMiddleware(stubVal, policies),
		)
		s.HandleFunc("GET /public", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("public"))
		})
		s.HandleFunc("GET /protected", func(w http.ResponseWriter, r *http.Request) {
			claims, ok := auth.FromContext(r.Context())
			if !ok {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(claims.Subject))
		})
	})

	It("allows unprotected routes without auth header", func() {
		code, body := doRequest(httptest.NewRequest(http.MethodGet, "/api/public", http.NoBody))
		Expect(code).To(Equal(http.StatusOK))
		Expect(body).To(Equal("public"))
	})

	It("returns 401 when protected route is missing token", func() {
		code, body := doRequest(httptest.NewRequest(http.MethodGet, "/api/protected", http.NoBody))
		Expect(code).To(Equal(http.StatusUnauthorized))
		Expect(body).To(Equal("Unauthorized"))
	})

	It("returns 401 when validator fails token", func() {
		stubVal.err = auth.ErrInvalidToken
		req := httptest.NewRequest(http.MethodGet, "/api/protected", http.NoBody)
		req.Header.Set("Authorization", "Bearer mock-token")
		code, body := doRequest(req)

		Expect(code).To(Equal(http.StatusUnauthorized))
		Expect(body).To(Equal("Unauthorized"))
	})

	It("returns 403 when role requirement fails", func() {
		stubVal.claims = &auth.Claims{Subject: "user-1", Roles: []string{"viewer"}}
		req := httptest.NewRequest(http.MethodGet, "/api/protected", http.NoBody)
		req.Header.Set("Authorization", "Bearer mock-token")
		code, body := doRequest(req)

		Expect(code).To(Equal(http.StatusForbidden))
		Expect(body).To(Equal("Forbidden"))
	})

	It("allows protected route and injects claims when authorized", func() {
		req := httptest.NewRequest(http.MethodGet, "/api/protected", http.NoBody)
		req.Header.Set("Authorization", "Bearer mock-token")
		code, body := doRequest(req)

		Expect(code).To(Equal(http.StatusOK))
		Expect(body).To(Equal("user-1"))
	})

	It("maps unknown validator errors to 401", func() {
		stubVal.err = errors.New("validator unavailable")
		req := httptest.NewRequest(http.MethodGet, "/api/protected", http.NoBody)
		req.Header.Set("Authorization", "Bearer mock-token")
		code, body := doRequest(req)

		Expect(code).To(Equal(http.StatusUnauthorized))
		Expect(body).To(Equal("Unauthorized"))
	})
})
