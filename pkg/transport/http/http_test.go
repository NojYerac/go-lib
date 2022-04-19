package http_test

import (
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rs/zerolog"
	mockhealth "source.rad.af/libs/go-lib/internal/mocks/health"
	"source.rad.af/libs/go-lib/pkg/log"
	. "source.rad.af/libs/go-lib/pkg/transport/http"
	"source.rad.af/libs/go-lib/pkg/version"
)

var _ = Describe("server", func() {
	var (
		s           Server
		w           *httptest.ResponseRecorder
		req         *http.Request
		body        string
		mockChecker *mockhealth.Checker
	)
	BeforeEach(func() {
		mockChecker = &mockhealth.Checker{}
		w = httptest.NewRecorder()
		l := log.NewLogger(log.TestConfig)
		s = NewServer(&Configuration{ServerDebug: false}, WithLogger(l), WithHealthCheck(mockChecker))
	})
	JustBeforeEach(func() {
		s.ServeHTTP(w, req)
		body = w.Body.String()
	})
	AfterEach(func() {
		mockChecker.AssertExpectations(GinkgoT())
	})
	Context("util routes", func() {
		Describe("GET /ping", func() {
			BeforeEach(func() {
				req = httptest.NewRequest(http.MethodGet, "/ping", http.NoBody)
			})
			It("returns pong", func() {
				Expect(w.Code).To(Equal(http.StatusOK))
				Expect(body).To(Equal("pong"))
			})
		})
		Describe("GET /health", func() {
			Context("healthy", func() {
				BeforeEach(func() {
					mockChecker.On("Passed").Return(true)
					req = httptest.NewRequest(http.MethodGet, "/health", http.NoBody)
				})
				It("returns health", func() {
					Expect(w.Code).To(Equal(http.StatusOK))
					Expect(body).To(Equal(""))
				})
			})
			Context("sick", func() {
				BeforeEach(func() {
					mockChecker.On("Passed").Return(false)
					req = httptest.NewRequest(http.MethodGet, "/health", http.NoBody)
				})
				It("returns health", func() {
					Expect(w.Code).To(Equal(http.StatusOK))
					Expect(body).To(Equal(""))
				})
			})
			Context("verbose", func() {
				BeforeEach(func() {
					mockChecker.On("Passed").Return(true)
					mockChecker.On("String").Return("verbose")
					req = httptest.NewRequest(http.MethodGet, "/health?verbose", http.NoBody)
				})
				It("returns health", func() {
					Expect(w.Code).To(Equal(http.StatusOK))
					Expect(body).To(Equal("verbose"))
				})
			})

			BeforeEach(func() {
				mockChecker.On("Passed").Return(true)
				// mockChecker.On("String").Return("passed")
				req = httptest.NewRequest(http.MethodGet, "/health", http.NoBody)
			})
			It("returns health", func() {
				Expect(w.Code).To(Equal(http.StatusOK))
				Expect(body).To(Equal(""))
			})
		})
		Describe("GET /version", func() {
			BeforeEach(func() {
				version.SetServiceName("test")
				req = httptest.NewRequest(http.MethodGet, "/version", http.NoBody)
			})
			It("returns the version", func() {
				Expect(w.Code).To(Equal(http.StatusOK))
				Expect(body).To(And(
					ContainSubstring(`"serviceName":"test"`),
					ContainSubstring(`"semVer":"0.0.0"`),
				))
			})
		})
	})
	Context("api routes", func() {
		BeforeEach(func() {
			s.APIRoutes().GET("/test", serviceRoute())
			req = httptest.NewRequest(http.MethodGet, "/api/test", http.NoBody)
		})
		It("returns ok", func() {
			Expect(w.Code).To(Equal(http.StatusOK))
			Expect(body).To(Equal("OK"))
		})
	})
})

func serviceRoute() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		logger := zerolog.Ctx(ctx)
		logger.Debug().Msg("service route")

		c.String(http.StatusOK, "OK")
	}
}
