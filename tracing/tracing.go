package tracing

import (
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	stdout "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"

	"github.com/nojyerac/go-lib/internal/runtime"
)

type Configuration struct {
	ExporterType string  `config:"exporter_type" validate:"oneof=stdout file otlp noop"`
	FilePath     string  `config:"file_path" validate:"required_if=ExporterType file"`
	OTLPURL      string  `config:"otlp_url" validate:"url,required_if=ExporterType otlp"`
	SampleRatio  float64 `config:"sample_ratio" validate:"gte=0,lte=1"`
}

func NewTracerProvider(config *Configuration) trace.TracerProvider {
	opts := []sdktrace.TracerProviderOption{}
	switch config.ExporterType {
	case "stdout":
		opts = append(opts, stdOutTracerProviderOpts()...)
	case "file":
		opts = append(opts, fileTracerExporterOpts(config)...)
	case "otlp":
		opts = append(opts, otlpTracerProviderOpts(config)...)
	default:
		return noop.NewTracerProvider()
	}
	tp := sdktrace.NewTracerProvider(opts...)
	return tp
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

func otlpTracerProviderOpts(config *Configuration) []sdktrace.TracerProviderOption {
	// TODO: support TLS, auth, & gRPC
	exporter := otlptracehttp.NewUnstarted(otlptracehttp.WithEndpoint(config.OTLPURL), otlptracehttp.WithInsecure())
	return []sdktrace.TracerProviderOption{
		sdktrace.WithSampler(sdktrace.ParentBased(
			sdktrace.TraceIDRatioBased(config.SampleRatio),
		)),
		sdktrace.WithBatcher(exporter),
	}
}

func fileTracerExporterOpts(config *Configuration) []sdktrace.TracerProviderOption {
	writer, err := os.OpenFile(config.FilePath, os.O_CREATE|os.O_RDWR|os.O_APPEND, os.ModeAppend)
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
