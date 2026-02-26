package grpc

import (
	"context"
	"fmt"
	"sync"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/nojyerac/go-lib/health"
	"github.com/nojyerac/go-lib/log"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	ini        bool
	lock       sync.Mutex
	grpcLogger *logrus.Logger
)

type clientOpts struct {
	dialOpts []grpc.DialOption
	hc       health.Checker
}

type ClientOpt func(o *clientOpts)

func WithDialOptions(dailOpts ...grpc.DialOption) ClientOpt {
	return func(o *clientOpts) {
		o.dialOpts = dailOpts
	}
}

func WithHealthChecker(hc health.Checker) ClientOpt {
	return func(o *clientOpts) {
		o.hc = hc
	}
}

// ClientConn returns a pointer to a new client connection
func ClientConn(target string, opts ...ClientOpt) (*grpc.ClientConn, error) {
	o := new(clientOpts)
	for _, apply := range opts {
		apply(o)
	}
	cc, err := clientConn(target, o.dialOpts...)
	if err != nil {
		return nil, err
	}
	if o.hc != nil {
		o.hc.Register("grpc_client", func(_ context.Context) error {
			state := cc.GetState()
			if state == connectivity.Ready || state == connectivity.Idle {
				return nil
			}
			return fmt.Errorf("grpc client not ready: %s", state)
		})
	}
	return cc, err
}

func clientConn(
	target string,
	opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	initialize()
	if len(opts) < 1 {
		opts = []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		}
	}
	opts = append(opts,
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		grpc.WithChainStreamInterceptor(streamClientInterceptors()...),
		grpc.WithChainUnaryInterceptor(unaryClientInterceptor()...),
	)
	return grpc.NewClient(target, opts...)
}

// NewServer returns a pointer to a new gRPC server. It takes a function to register
// services, which allows metrics to be registered on each service.
func NewServer(registerServices func(*grpc.Server), opts ...grpc.ServerOption) *grpc.Server {
	initialize()
	opts = append(opts,
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(unaryServerInterceptor()...),
		grpc.ChainStreamInterceptor(streamServerInterceptor()...),
	)
	server := grpc.NewServer(opts...)

	registerServices(server)
	grpc_prometheus.Register(server)

	return server
}

func SetLogger(l *logrus.Logger) {
	lock.Lock()
	defer lock.Unlock()
	if ini {
		panic("cannot set logger after initialization")
	}
	grpcLogger = l
}

func initialize() {
	lock.Lock()
	defer lock.Unlock()
	if ini {
		return
	}
	ini = true
	if grpcLogger == nil {
		grpcLogger = log.FromContext(context.Background())
	}
}

// InterceptorLogger
func InterceptorLogger() logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		f := make(map[string]any, len(fields)/2)
		i := logging.Fields(fields).Iterator()
		for i.Next() {
			k, v := i.At()
			f[k] = v
		}

		l := log.FromContext(ctx).WithFields(f)

		switch lvl {
		case logging.LevelDebug:
			l.Debug(msg)
		case logging.LevelInfo:
			l.Info(msg)
		case logging.LevelWarn:
			l.Warn(msg)
		case logging.LevelError:
			l.Error(msg)
		default:
			panic(fmt.Sprintf("unknown level %v", lvl))
		}
	})
}

func streamClientInterceptors() []grpc.StreamClientInterceptor {
	return []grpc.StreamClientInterceptor{
		grpc_prometheus.StreamClientInterceptor,
		logging.StreamClientInterceptor(InterceptorLogger()),
	}
}

func unaryClientInterceptor() []grpc.UnaryClientInterceptor {
	return []grpc.UnaryClientInterceptor{
		grpc_prometheus.UnaryClientInterceptor,
		logging.UnaryClientInterceptor(InterceptorLogger()),
	}
}

func streamServerInterceptor() []grpc.StreamServerInterceptor {
	return []grpc.StreamServerInterceptor{
		grpc_prometheus.StreamServerInterceptor,
		streamServerCtxWithLoggerInterceptor(),
		logging.StreamServerInterceptor(InterceptorLogger()),
		grpc_recovery.StreamServerInterceptor(),
	}
}

func streamServerCtxWithLoggerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := ss.Context()
		wrapped := grpc_middleware.WrapServerStream(ss)
		wrapped.WrappedContext = log.WithLogger(ctx, grpcLogger)
		return handler(srv, wrapped)
	}
}

func unaryServerInterceptor() []grpc.UnaryServerInterceptor {
	return []grpc.UnaryServerInterceptor{
		grpc_prometheus.UnaryServerInterceptor,
		unaryServerCtxWithLoggerInterceptor(),
		logging.UnaryServerInterceptor(InterceptorLogger()),
		grpc_recovery.UnaryServerInterceptor(),
	}
}

func unaryServerCtxWithLoggerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		gCtx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		ctx := log.WithLogger(gCtx, grpcLogger)
		return handler(ctx, req)
	}
}
