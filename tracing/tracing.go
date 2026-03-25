package tracing

import (
	"context"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	stdout "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"

	"github.com/nojyerac/go-lib/internal/runtime"
)

type Starter interface {
	Start(context.Context) error
	Shutdown(context.Context) error
}

var _ Starter = (*noopStarter)(nil)

type noopStarter struct{}

func (*noopStarter) Start(_ context.Context) error {
	return nil
}

func (*noopStarter) Shutdown(_ context.Context) error {
	return nil
}

type Configuration struct {
	ExporterType string  `config:"exporter_type" validate:"oneof=stdout file otlp_http otlp_grpc noop"`
	FilePath     string  `config:"file_path" validate:"required_if=ExporterType file"`
	OtlpEndpoint string  `config:"otlp_endpoint" validate:"required_if=ExporterType otlp_http otlp_grpc,omitempty,hostname_port"`
	SampleRatio  float64 `config:"sample_ratio" validate:"gte=0,lte=1"`
}

func NewTracerProvider(config *Configuration) (trace.TracerProvider, Starter) {
	opts := []sdktrace.TracerProviderOption{}
	var starter Starter = &noopStarter{}
	switch config.ExporterType {
	case "stdout":
		opts = stdOutTracerProviderOpts()
	case "file":
		opts = fileTracerExporterOpts(config)
	case "otlp_http":
		opts, starter = httpTracerProviderOpts(config)
	case "otlp_grpc":
		opts, starter = grpcTracerProviderOpts(config)
	default:
		return noop.NewTracerProvider(), starter
	}
	tp := sdktrace.NewTracerProvider(opts...)
	return tp, starter
}

func SetGlobal(tp trace.TracerProvider) {
	otel.SetTracerProvider(tp)
}

func TracerForPackage(skipMore ...int) trace.Tracer {
	skip := 2
	for _, s := range skipMore {
		skip += s
	}
	return otel.Tracer(runtime.GetPackageName(skip))
}

func stdOutTracerProviderOpts() []sdktrace.TracerProviderOption {
	exporter, err := stdout.New(stdout.WithPrettyPrint())
	if err != nil {
		panic(err)
	}
	return []sdktrace.TracerProviderOption{
		sdktrace.WithSyncer(exporter),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	}
}

func httpTracerProviderOpts(config *Configuration) ([]sdktrace.TracerProviderOption, Starter) {
	exporter := otlptracehttp.NewUnstarted(otlptracehttp.WithEndpoint(config.OtlpEndpoint), otlptracehttp.WithInsecure())
	return []sdktrace.TracerProviderOption{
		sdktrace.WithSampler(sdktrace.ParentBased(
			sdktrace.TraceIDRatioBased(config.SampleRatio),
		)),
		sdktrace.WithSyncer(exporter),
	}, exporter
}

func grpcTracerProviderOpts(config *Configuration) ([]sdktrace.TracerProviderOption, Starter) {
	exporter := otlptracegrpc.NewUnstarted(otlptracegrpc.WithEndpoint(config.OtlpEndpoint), otlptracegrpc.WithInsecure())
	return []sdktrace.TracerProviderOption{
		sdktrace.WithSampler(sdktrace.ParentBased(
			sdktrace.TraceIDRatioBased(config.SampleRatio),
		)),
		sdktrace.WithBatcher(exporter),
	}, exporter
}

func fileTracerExporterOpts(config *Configuration) []sdktrace.TracerProviderOption {
	writer, err := os.OpenFile(config.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		panic(err)
	}
	exporter, err := stdout.New(stdout.WithWriter(writer))
	if err != nil {
		panic(err)
	}
	return []sdktrace.TracerProviderOption{
		sdktrace.WithSyncer(exporter),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	}
}
