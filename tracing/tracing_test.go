package tracing_test

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/nojyerac/go-lib/tracing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const port = "9999"

var _ = Describe("", func() {
	It("", func() {
		Expect(true).To(BeTrue())
	})
	Describe("otlp", func() {
		var (
			srv *http.Server
		)
		BeforeEach(func() {
			srv = mockOtlpSrv(port)
			go func() {
				defer GinkgoRecover()
				if err := srv.ListenAndServe(); err != nil {
					Expect(err).To(MatchError(http.ErrServerClosed))
				}
			}()
		})
		AfterEach(func() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			srv.Shutdown(ctx)
		})
		It("is testable", func() {
			req, err := http.NewRequest("GET", "http://localhost:9999/v1", http.NoBody)
			Expect(err).NotTo(HaveOccurred())
			res, err := http.DefaultClient.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer res.Body.Close()
			Expect(res.StatusCode).To(Equal(http.StatusOK))
		})
		Describe("tracing", func() {
			It("reports spans", func() {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				tp, starter := tracing.NewTracerProvider(&tracing.Configuration{
					ExporterType: "otlp",
					OtlpEndpoint: "localhost:9999",
					SampleRatio:  1,
				})
				starter.Start(ctx)
				tracer := tp.Tracer("tracing-test")
				ctx1, span1 := tracer.Start(ctx, "span1")
				_, span2 := tracer.Start(ctx1, "span2")
				span2.End()
				span1.End()
				time.Sleep(time.Second)
				err := starter.Shutdown(ctx)
				Expect(err).NotTo(HaveOccurred())
				time.Sleep(time.Second)
			})
		})
	})
})

func mockOtlpSrv(port string) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("received request: %+v\n", r)
		bodyBytes, _ := io.ReadAll(r.Body)
		defer r.Body.Close()
		fmt.Printf("body: %s", string(bodyBytes))
		_, _ = w.Write([]byte("ok"))
	})
	return &http.Server{
		Addr:                         net.JoinHostPort("localhost", port),
		Handler:                      mux,
		DisableGeneralOptionsHandler: true,
	}
}
