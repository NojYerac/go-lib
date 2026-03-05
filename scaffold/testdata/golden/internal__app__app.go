// Package app wires all go-lib subsystems and starts the example server.
package app

import (
	"context"
	"fmt"

	libconfig "github.com/nojyerac/go-lib/config"
	"github.com/nojyerac/go-lib/health"
	"github.com/nojyerac/go-lib/log"
	"github.com/nojyerac/go-lib/metrics"
	"github.com/nojyerac/go-lib/tracing"
	"github.com/nojyerac/go-lib/transport"
	libhttp "github.com/nojyerac/go-lib/transport/http"
	"github.com/nojyerac/go-lib/version"

	svcconfig "github.com/example/example/config"
	svctransport "github.com/example/example/transport"
)

// Run initialises all subsystems and blocks until ctx is cancelled
// (SIGINT / SIGTERM).  It returns any non-context-cancellation error.
func Run(ctx context.Context) error {
	// ── config ────────────────────────────────────────────────────────────────
	cfg := svcconfig.NewConfig()
	loader := libconfig.NewConfigLoader("EXAMPLE")
	if err := loader.RegisterConfig(cfg); err != nil {
		return fmt.Errorf("register config: %w", err)
	}
	if err := loader.InitAndValidate(); err != nil {
		return fmt.Errorf("init config: %w", err)
	}

	// ── logger ────────────────────────────────────────────────────────────────
	logger := log.NewLogger(&cfg.Log)
	log.SetDefaultCtxLogger(logger)
	ctx = log.WithLogger(ctx, logger)
	l := log.FromContext(ctx)

	// ── version ───────────────────────────────────────────────────────────────
	version.SetServiceName("example")

	// ── tracing ───────────────────────────────────────────────────────────────
	tp := tracing.NewTracerProvider(&cfg.Tracing)
	tracing.SetGlobal(tp)

	// ── metrics ───────────────────────────────────────────────────────────────
	mp, metricsHandler, err := metrics.NewMetricProvider()
	if err != nil {
		return fmt.Errorf("init metrics: %w", err)
	}
	metrics.SetGlobal(mp)

	// ── health ────────────────────────────────────────────────────────────────
	checker := health.NewChecker(&cfg.Health)
	go func() { _ = checker.Start(ctx) }()

	// ── HTTP server ───────────────────────────────────────────────────────────
	httpServer := libhttp.NewServer(&cfg.HTTP,
		libhttp.WithHealthChecker(checker),
		libhttp.WithMetricsHandler(metricsHandler),
	)
	svctransport.RegisterRoutes(httpServer)

	// ── transport (combined HTTP + gRPC) ──────────────────────────────────────
	srv, err := transport.NewServer(&cfg.Transport,
		transport.WithHTTP(httpServer),
	)
	if err != nil {
		return fmt.Errorf("create transport: %w", err)
	}

	l.WithField("module", "github.com/example/example").Info("example starting")
	return srv.Start(ctx)
}
