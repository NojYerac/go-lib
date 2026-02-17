package http_test

import (
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/gin-gonic/gin"
	mockhealth "github.com/nojyerac/go-lib/internal/mocks/health"
	"github.com/nojyerac/go-lib/pkg/log"
	"github.com/nojyerac/go-lib/pkg/metrics"
	"github.com/nojyerac/go-lib/pkg/tracing"
	. "github.com/nojyerac/go-lib/pkg/transport/http"
	"github.com/nojyerac/go-lib/pkg/version"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/metric"
)

var metricHandler http.HandlerFunc
var _ = BeforeSuite(func() {
	version.SetServiceName("test")
	var m metric.MeterProvider
	m, metricHandler, _ = metrics.NewMeterProvider(nil)
	metrics.SetGlobal(m)
	tracing.SetGlobal(tracing.NewTracerProvider(&tracing.Configuration{
		ExporterType: "noop",
	}))
})

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
		s = NewServer(
			&Configuration{ServerDebug: false},
			WithLogger(l),
			WithHealthCheck(mockChecker),
			WithMetricHandler(metricHandler),
		)
		s.APIRoutes().GET("/ok", serviceRoute())
		s.APIRoutes().GET("/err", serviceRoute())
		s.APIRoutes().GET("/panic", serviceRoute())
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
		Describe("GET /metrics", func() {
			BeforeEach(func() {
				s.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/api/ok", http.NoBody))
				time.Sleep(100 * time.Millisecond)
				req = httptest.NewRequest(http.MethodGet, "/metrics", http.NoBody)
			})
			It("returns the metrics", func() {
				Expect(w.Code).To(Equal(http.StatusOK))
				Expect(body).To(ContainSubstring("http_request_duration_histogram"))
			})
		})
	})
	Context("api routes", func() {
		When("handler succeeds", func() {
			BeforeEach(func() {
				req = httptest.NewRequest(http.MethodGet, "/api/ok", http.NoBody)
			})
			It("returns ok", func() {
				Expect(w.Code).To(Equal(http.StatusOK))
				Expect(body).To(Equal("OK"))
			})
		})
		When("handler errors", func() {
			BeforeEach(func() {
				req = httptest.NewRequest(http.MethodGet, "/api/err", http.NoBody)
			})
			It("returns error", func() {
				Expect(w.Code).To(Equal(http.StatusInternalServerError))
				Expect(body).To(Equal("error"))
			})
		})
		When("handler panics", func() {
			BeforeEach(func() {
				req = httptest.NewRequest(http.MethodGet, "/api/panic", http.NoBody)
			})
			It("returns error", func() {
				Expect(w.Code).To(Equal(http.StatusInternalServerError))
				Expect(body).To(Equal("[PANIC RECOVER] mock panic"))
			})
		})
	})
})

func serviceRoute() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		logger := zerolog.Ctx(ctx)
		logger.Debug().Msg("service route")
		switch c.FullPath() {
		case "/api/err":
			c.String(http.StatusInternalServerError, "error")
		case "/api/panic":
			panic("mock panic")
		default:
			c.String(http.StatusOK, "OK")
		}
	}
}
