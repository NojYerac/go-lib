package tracing

import (
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	stdout "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"
	"source.rad.af/libs/go-lib/internal/runtime"
	"source.rad.af/libs/go-lib/pkg/version"
)

func NewTracerProvider(config *Configuration) trace.TracerProvider {
	v := version.GetVersion()
	opts := []sdktrace.TracerProviderOption{
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(v.Name),
			semconv.ServiceVersionKey.String(v.SemVer),
		)),
	}
	switch config.ExporterType {
	case "stdout":
		opts = append(opts, stdOutTracerProviderOpts()...)
	case "file":
		opts = append(opts, fileTracerExporterOpts(config)...)
	case "jaeger":
		opts = append(opts, jaegerTracerProviderOpts(config)...)
	case "noop":
		return trace.NewNoopTracerProvider()
	default:
		panic("unrecognized exporter type")
	}
	return sdktrace.NewTracerProvider(opts...)
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

func jaegerTracerProviderOpts(config *Configuration) []sdktrace.TracerProviderOption {
	exporter, err := jaeger.New(jaeger.WithAgentEndpoint())
	if err != nil {
		panic(err)
	}
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
