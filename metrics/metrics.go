package metrics

import (
	"net/http"

	"github.com/nojyerac/go-lib/internal/runtime"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/prometheus"
	api "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric"
)

func NewMetricProvider() (*metric.MeterProvider, http.Handler, error) {
	exporter, err := prometheus.New()
	if err != nil {
		return nil, nil, err
	}
	httpHandler := promhttp.Handler()
	mp := metric.NewMeterProvider(metric.WithReader(exporter))
	return mp, httpHandler, nil
}

func SetGlobal(mp *metric.MeterProvider) {
	otel.SetMeterProvider(mp)
}

func MeterForPackage() api.Meter {
	mp := otel.GetMeterProvider()
	return mp.Meter(runtime.GetPackageName(2))
}
