package db

import (
	"context"
	"regexp"

	"github.com/nojyerac/go-lib/metrics"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
)

var (
	initial             bool
	durationHistogram   metric.Int64Histogram
	openConnectionGuage metric.Int64ObservableGauge
	idleConnectionGuage metric.Int64ObservableGauge
	meter               = metrics.MeterForPackage()
	reConnStr           = regexp.MustCompile(`password=\S+`)
)

func initMetrics(db *database) error {
	if !initial {
		initial = true
		var err error
		durationHistogram, err = meter.Int64Histogram(
			"database_query_duration",
			metric.WithDescription("duration of database queries"),
			metric.WithUnit("ms"),
		)
		if err != nil {
			return err
		}
		openConnectionGuage, err = meter.Int64ObservableGauge(
			"database_open_conns",
			metric.WithDescription("count of open database connections"),
		)
		if err != nil {
			return err
		}
		idleConnectionGuage, err = meter.Int64ObservableGauge(
			"database_idle_conns",
			metric.WithDescription("count of idle database connections"),
		)
		if err != nil {
			return err
		}
	}
	inst := []metric.Observable{openConnectionGuage, idleConnectionGuage}
	connStr := reConnStr.ReplaceAllLiteralString(db.connStr, "password=***")
	attrs := []attribute.KeyValue{
		semconv.DBConnectionStringKey.String(connStr),
	}
	callBack := func(ctx context.Context, o metric.Observer) error {
		stats := db.conn.Stats()
		o.ObserveInt64(openConnectionGuage, int64(stats.OpenConnections), metric.WithAttributes(attrs...))
		o.ObserveInt64(idleConnectionGuage, int64(stats.Idle), metric.WithAttributes(attrs...))
		return nil
	}
	reg, err := meter.RegisterCallback(callBack, inst...)
	if err != nil {
		return err
	}
	db.o.reg = reg
	return nil
}
