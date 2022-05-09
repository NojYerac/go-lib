package db

import (
	"context"
	"regexp"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/metric/instrument/asyncint64"
	"go.opentelemetry.io/otel/metric/instrument/syncint64"
	"go.opentelemetry.io/otel/metric/unit"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"source.rad.af/libs/go-lib/pkg/metrics"
)

var (
	initial             bool
	durationHistogram   syncint64.Histogram
	openConnectionGuage asyncint64.Gauge
	idleConnectionGuage asyncint64.Gauge
	meter               = metrics.MeterForPackage()
	reConnStr           = regexp.MustCompile(`password=\S+`)
)

func initMetrics(db *database) error {
	if !initial {
		initial = true
		var err error
		durationHistogram, err = meter.SyncInt64().Histogram(
			"database_query_duration",
			instrument.WithDescription("duration (ms) of database queries"),
			instrument.WithUnit(unit.Milliseconds),
		)
		if err != nil {
			return err
		}
		openConnectionGuage, err = meter.AsyncInt64().Gauge(
			"database_open_conns",
			instrument.WithDescription("count of open database connections"),
		)
		if err != nil {
			return err
		}
		idleConnectionGuage, err = meter.AsyncInt64().Gauge(
			"database_idle_conns",
			instrument.WithDescription("count of idle database connections"),
		)
		if err != nil {
			return err
		}
	}
	inst := []instrument.Asynchronous{openConnectionGuage, idleConnectionGuage}
	connStr := reConnStr.ReplaceAllLiteralString(db.connStr, "password=***")
	attrs := []attribute.KeyValue{
		semconv.DBConnectionStringKey.String(connStr),
	}
	callBack := func(ctx context.Context) {
		stats := db.conn.Stats()
		openConnectionGuage.Observe(
			ctx,
			int64(stats.OpenConnections),
			attrs...,
		)
		idleConnectionGuage.Observe(
			ctx,
			int64(stats.Idle),
			attrs...,
		)
	}
	return meter.RegisterCallback(inst, callBack)
}
