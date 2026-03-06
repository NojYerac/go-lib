package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"sync"

	"github.com/nojyerac/go-lib/auth"
	libdb "github.com/nojyerac/go-lib/db"
	"github.com/nojyerac/go-lib/health"
	"github.com/nojyerac/go-lib/log"
	"github.com/nojyerac/go-lib/metrics"
	"github.com/nojyerac/go-lib/tracing"
	"github.com/nojyerac/go-lib/transport"
	libgrpc "github.com/nojyerac/go-lib/transport/grpc"
	libhttp "github.com/nojyerac/go-lib/transport/http"
	"github.com/nojyerac/go-lib/version"
	"github.com/example/example/config"
	"github.com/example/example/data/db"
	"github.com/example/example/security"
	"github.com/example/example/transport/http"
	"github.com/example/example/transport/rpc"
)

func main() {
	// init & config
	version.SetSemVer("0.0.0")
	version.SetServiceName("example")
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	if err := config.InitAndValidate(); err != nil {
		panic(err)
	}

	// telemetry
	logger := log.NewLogger(config.LogConfig)
	ctx = log.WithLogger(ctx, logger)
	tracing.NewTracerProvider(config.TraceConfig)
	mp, metricHandler, err := metrics.NewMetricProvider()
	if err != nil {
		logger.WithError(err).Panic("failed to create metric provider")
	}
	metrics.SetGlobal(mp)
	hc := health.NewChecker(config.HealthConfig)

	// sources
	database := libdb.NewDatabase(
		config.DBConfig,
		libdb.WithHealthChecker(hc),
		libdb.WithLogger(logger),
	)
	dataSrc := db.NewDataSource(
		database,
		db.WithLogger(logger),
	)

	// transports
	av := auth.NewValidator(config.AuthConfig)
	hSrv := libhttp.NewServer(
		config.HTTPConfig,
		libhttp.WithMetricsHandler(metricHandler),
		libhttp.WithHealthChecker(hc),
		libhttp.WithAuthMiddleware(av, security.HTTPPolicyMap()),
	)
	http.RegisterRoutes(dataSrc, hSrv)

	libgrpc.SetLogger(logger)
	gSrv := libgrpc.NewServer(
		rpc.RegisterServices(dataSrc, rpc.WithLogger(logger)),
		libgrpc.AuthServerOptions(av, security.GRPCPolicyMap())...,
	)

	srv, err := transport.NewServer(
		config.TransConfig,
		transport.WithHTTP(hSrv),
		transport.WithGRPC(gSrv),
	)
	if err != nil {
		logger.WithError(err).Panic("failed to create server")
	}

	// start service
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := database.Open(ctx); err != nil {
			logger.WithError(err).Panic("database error")
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := srv.Start(ctx); err != nil && err != net.ErrClosed {
			logger.WithError(err).Panic("server error")
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := hc.Start(ctx); err != nil && err != context.Canceled {
			logger.WithError(err).Panic("health checker error")
		}
	}()

	logger.Info("service starting")
	<-ctx.Done()
	logger.Info("service stopping")
	wg.Wait()
	logger.Info("service stopped")
}
