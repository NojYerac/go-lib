package tracing_test

import (
	"context"
	"os"
	"strings"

	"github.com/go-playground/validator/v10"
	. "github.com/nojyerac/go-lib/pkg/tracing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var _ = Describe("tracing", func() {
	var (
		tp trace.TracerProvider
		c  *Configuration
	)
	JustBeforeEach(func() {
		Expect(validator.New().Struct(c)).To(Succeed())
		tp = NewTracerProvider(c)
	})
	BeforeEach(func() {
		c = NewConfiguration()
	})
	Context("default config", func() {
		It("returns a stdout trace provider", func() {
			Expect(tp).NotTo(BeNil())
		})
		Context("set global", func() {
			It("sets the global", func() {
				SetGlobal(tp)
				Expect(otel.GetTracerProvider()).To(Equal(tp))
			})
		})
	})
	Context("file exporter", func() {
		var (
			f   *os.File
			err error
		)
		BeforeEach(func() {
			f, err = os.CreateTemp("", "tracing.txt")
			Expect(err).NotTo(HaveOccurred())
			c.ExporterType = "file"
			c.FilePath = f.Name()
		})
		AfterEach(func() {
			Expect(os.Remove(f.Name())).To(Succeed())
		})
		It("returns a stdout trace provider", func() {
			Expect(tp).NotTo(BeNil())
			SetGlobal(tp)
			tracer := TracerForPackage(0)
			c1, serverSpan := tracer.Start(context.Background(), "test_span", trace.WithSpanKind(trace.SpanKindServer))
			serverSpan.AddEvent("test event", trace.WithAttributes(attribute.String("testEventAttrKey", "testEventAttrVal")))
			serverSpan.SetStatus(codes.Ok, "test status")
			serverSpan.SetAttributes(attribute.String("testAttrKey", "testAttrValue"))
			c2, internalSpan := tracer.Start(c1, "test_child_span")
			_, clientSpan := tracer.Start(c2, "test_client_span", trace.WithSpanKind(trace.SpanKindClient))
			clientSpan.End()
			internalSpan.End()
			serverSpan.End(trace.WithStackTrace(true))
			content, err := os.ReadFile(f.Name())
			Expect(err).NotTo(HaveOccurred())
			// _, _ = fmt.Println(string(content))
			traces := strings.Split(strings.Trim(string(content), "\n"), "\n")
			Expect(traces).To(HaveLen(3))
			for _, t := range traces {
				Expect(t).To(And(
					ContainSubstring(`"Resource":[{"Key":"service.name","Value":{"Type":"STRING","Value":""}},`),
					ContainSubstring(`"InstrumentationLibrary":{"Name":"github.com/nojyerac/go-lib/pkg/tracing_test"`),
				))
			}
		})
	})
	Context("jaeger exporter", func() {
		BeforeEach(func() {
			c.ExporterType = "jaeger"
		})
		It("returns a jaeger trace provider", func() {
			Expect(tp).NotTo(BeNil())
		})
	})
	Context("noop exporter", func() {
		BeforeEach(func() {
			c.ExporterType = "noop"
		})
		It("returns a jaeger trace provider", func() {
			Expect(tp).NotTo(BeNil())
		})
	})
})
