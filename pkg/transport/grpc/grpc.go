package grpc

import (
	"context"
	"errors"

	grpc_zerolog "github.com/grpc-ecosystem/go-grpc-middleware/providers/zerolog/v2"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware/v2"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/tags"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	ini        bool
	grpcLogger *zerolog.Logger
)

// ClientConn returns a pointer to a new client connection
func ClientConn(
	target string,
	opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	initialize()
	if len(opts) < 1 {
		opts = []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithBlock(),
		}
	}
	opts = append(opts,
		grpc.WithChainStreamInterceptor(streamClientInterceptors()...),
		grpc.WithChainUnaryInterceptor(unaryClientInterceptor()...),
	)
	return grpc.Dial(target, opts...)
}

// NewServer returns a pointer to a new gRPC server. It takes a function to register
// services, which allows metrics to be registered on each service.
func NewServer(registerServices func(*grpc.Server), opts ...grpc.ServerOption) *grpc.Server {
	initialize()
	opts = append(opts,
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(unaryServerInterceptor()...),
		),
		grpc.StreamInterceptor(
			grpc_middleware.ChainStreamServer(streamServerInterceptor()...),
		),
	)
	server := grpc.NewServer(opts...)

	registerServices(server)
	grpc_prometheus.Register(server)

	return server
}

func SetGrpcLogger(l *zerolog.Logger) error {
	if ini {
		return errors.New("must call SetGrpcLogger before initializing")
	}
	grpcLogger = l
	return nil
}

func initialize() {
	if ini {
		return
	}
	if grpcLogger == nil {
		grpcLogger = zerolog.Ctx(context.Background())
	}
	replaceGRPCLogger(grpcLogger)
	ini = true
}

func streamClientInterceptors() []grpc.StreamClientInterceptor {
	return []grpc.StreamClientInterceptor{
		grpc_prometheus.StreamClientInterceptor,
		otelgrpc.StreamClientInterceptor(),
		logging.StreamClientInterceptor(grpc_zerolog.InterceptorLogger(*grpcLogger)),
	}
}

func unaryClientInterceptor() []grpc.UnaryClientInterceptor {
	return []grpc.UnaryClientInterceptor{
		grpc_prometheus.UnaryClientInterceptor,
		otelgrpc.UnaryClientInterceptor(),
		logging.UnaryClientInterceptor(grpc_zerolog.InterceptorLogger(*grpcLogger)),
	}
}

func streamServerInterceptor() []grpc.StreamServerInterceptor {
	return []grpc.StreamServerInterceptor{
		tags.StreamServerInterceptor(),
		grpc_prometheus.StreamServerInterceptor,
		otelgrpc.StreamServerInterceptor(),
		logging.StreamServerInterceptor(grpc_zerolog.InterceptorLogger(*grpcLogger)),
		grpc_recovery.StreamServerInterceptor(),
		streamServerCtxWithLoggerInterceptor(),
	}
}

func streamServerCtxWithLoggerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := ss.Context()
		wrapped := grpc_middleware.WrapServerStream(ss)
		wrapped.WrappedContext = grpcLogger.WithContext(ctx)
		return handler(srv, wrapped)
	}
}

func unaryServerInterceptor() []grpc.UnaryServerInterceptor {
	return []grpc.UnaryServerInterceptor{
		tags.UnaryServerInterceptor(),
		grpc_prometheus.UnaryServerInterceptor,
		otelgrpc.UnaryServerInterceptor(),
		logging.UnaryServerInterceptor(grpc_zerolog.InterceptorLogger(*grpcLogger)),
		grpc_recovery.UnaryServerInterceptor(),
		unaryServerCtxWithLoggerInterceptor(),
	}
}

func unaryServerCtxWithLoggerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		gCtx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		ctx := grpcLogger.WithContext(gCtx)
		return handler(ctx, req)
	}
}
