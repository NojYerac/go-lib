package http_test

import (
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/nojyerac/go-lib/log"
	"github.com/nojyerac/go-lib/metrics"
	mockhealth "github.com/nojyerac/go-lib/mocks/github.com/nojyerac/go-lib/health"
	"github.com/nojyerac/go-lib/tracing"
	. "github.com/nojyerac/go-lib/transport/http"
	"github.com/nojyerac/go-lib/version"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.opentelemetry.io/otel/sdk/metric"
)

var metricHandler http.Handler
var _ = BeforeSuite(func() {
	version.SetServiceName("test")
	var mp *metric.MeterProvider
	mp, metricHandler, _ = metrics.NewMetricProvider()
	metrics.SetGlobal(mp)
	tp, _ := tracing.NewTracerProvider(&tracing.Configuration{
		ExporterType: "noop",
	})
	tracing.SetGlobal(tp)
})

var _ = Describe("server", func() {
	var (
		s           Server
		w           *httptest.ResponseRecorder
		req         *http.Request
		body        string
		mockChecker *mockhealth.MockChecker
	)
	BeforeEach(func() {
		mockChecker = &mockhealth.MockChecker{}
		w = httptest.NewRecorder()
		l := log.NewLogger(log.TestConfig)
		s = NewServer(
			&Configuration{},
			WithLogger(l),
			WithHealthChecker(mockChecker),
			WithMetricsHandler(metricHandler),
		)
		s.Handle("GET /ok", serviceRoute())
		s.Handle("GET /err", serviceRoute())
		s.Handle("GET /panic", serviceRoute())
	})
	JustBeforeEach(func() {
		s.ServeHTTP(w, req)
		body = w.Body.String()
	})
	AfterEach(func() {
		mockChecker.AssertExpectations(GinkgoT())
	})
	Context("util routes", func() {
		Describe("GET /livez", func() {
			BeforeEach(func() {
				req = httptest.NewRequest(http.MethodGet, "/livez", http.NoBody)
			})
			It("returns ok", func() {
				Expect(w.Code).To(Equal(http.StatusOK))
				Expect(body).To(Equal("ok"))
			})
		})
		Describe("GET /healthz", func() {
			Context("healthy", func() {
				BeforeEach(func() {
					mockChecker.On("Passed").Return(true)
					req = httptest.NewRequest(http.MethodGet, "/healthz", http.NoBody)
				})
				It("returns health", func() {
					Expect(w.Code).To(Equal(http.StatusOK))
					Expect(body).To(Equal(""))
				})
			})
			Context("sick", func() {
				BeforeEach(func() {
					mockChecker.On("Passed").Return(false)
					req = httptest.NewRequest(http.MethodGet, "/healthz", http.NoBody)
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
					req = httptest.NewRequest(http.MethodGet, "/healthz?v", http.NoBody)
				})
				It("returns health", func() {
					Expect(w.Code).To(Equal(http.StatusOK))
					Expect(body).To(Equal("verbose"))
				})
			})

			BeforeEach(func() {
				mockChecker.On("Passed").Return(true)
				// mockChecker.On("String").Return("passed")
				req = httptest.NewRequest(http.MethodGet, "/healthz", http.NoBody)
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

func serviceRoute() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/err":
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("error"))
		case "/api/panic":
			panic("mock panic")
		default:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK"))
		}
	}
}
