package metrics

import (
	"net/http"

	"github.com/nojyerac/go-lib/internal/runtime"
	otelruntime "go.opentelemetry.io/contrib/instrumentation/runtime"
	prometheusexporter "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	"go.opentelemetry.io/otel/sdk/metric/export/aggregation"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	"go.opentelemetry.io/otel/sdk/metric/selector/simple"
)

func SetGlobal(mp metric.MeterProvider) {
	global.SetMeterProvider(mp)
}

func MeterForPackage(extraSkip ...int) metric.Meter {
	skip := 2
	for _, i := range extraSkip {
		skip += i
	}
	packageName := runtime.GetPackageName(skip)
	return global.Meter(packageName)
}

func NewMeterProvider(_ *Configuration) (metric.MeterProvider, http.HandlerFunc, error) {
	ctlr := controller.New(processor.NewFactory(
		simple.NewWithHistogramDistribution(),
		aggregation.StatelessTemporalitySelector(),
	))
	ex, err := prometheusexporter.New(prometheusexporter.Config{}, ctlr)
	if err != nil {
		return nil, nil, err
	}
	err = otelruntime.Start(otelruntime.WithMeterProvider(ex.MeterProvider()))
	return ex.MeterProvider(), ex.ServeHTTP, err
}
