package grpc

import (
	"crypto/tls"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"source.rad.af/libs/lib-grpc/pkg/interceptors/logging"
	"source.rad.af/libs/lib-grpc/pkg/interceptors/requestid"
)

var (
	ini bool
)

func initialize(logger *logrus.Entry) {
	if ini {
		return
	}
	grpc_logrus.ReplaceGrpcLogger(logger)
	grpc_prometheus.EnableHandlingTimeHistogram()
	ini = true
}

func streamClientInterceptors(logger *logrus.Entry) []grpc.StreamClientInterceptor {
	return []grpc.StreamClientInterceptor{
		requestid.StreamClientInterceptor,
		logging.StreamClientInterceptor,
		grpc_prometheus.StreamClientInterceptor,
		grpc_opentracing.StreamClientInterceptor(),
		grpc_logrus.StreamClientInterceptor(logger),
	}
}

func unaryClientInterceptor(logger *logrus.Entry) []grpc.UnaryClientInterceptor {
	return []grpc.UnaryClientInterceptor{
		requestid.UnaryClientInterceptor,
		logging.UnaryClientInterceptor,
		grpc_prometheus.UnaryClientInterceptor,
		grpc_opentracing.UnaryClientInterceptor(),
		grpc_logrus.UnaryClientInterceptor(logger),
	}
}

func streamServerInterceptor(logger *logrus.Entry) []grpc.StreamServerInterceptor {
	return []grpc.StreamServerInterceptor{
		requestid.StreamServerInterceptor,
		logging.StreamServerInterceptor,
		grpc_prometheus.StreamServerInterceptor,
		grpc_opentracing.StreamServerInterceptor(),
		grpc_logrus.StreamServerInterceptor(logger),
		grpc_recovery.StreamServerInterceptor(),
	}
}

func unaryServerInterceptor(logger *logrus.Entry) []grpc.UnaryServerInterceptor {
	return []grpc.UnaryServerInterceptor{
		requestid.UnaryServerInterceptor,
		logging.UnaryServerInterceptor,
		grpc_prometheus.UnaryServerInterceptor,
		grpc_opentracing.UnaryServerInterceptor(),
		grpc_logrus.UnaryServerInterceptor(logger),
		grpc_recovery.UnaryServerInterceptor(),
	}
}

// ClientConn returns a pointer to a new client connection
func ClientConn(
	target string,
	opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	logger := logrus.WithField("grpc.target", target)
	initialize(logger)
	if len(opts) < 1 {
		cred := credentials.NewTLS(&tls.Config{
			InsecureSkipVerify: true, // nolint:gosec // internal only
		})
		opts = []grpc.DialOption{
			grpc.WithTransportCredentials(cred),
			grpc.WithBlock(),
		}
	}
	opts = append(opts,
		grpc.WithChainStreamInterceptor(streamClientInterceptors(logger)...),
		grpc.WithChainUnaryInterceptor(unaryClientInterceptor(logger)...),
	)
	return grpc.Dial(target, opts...)
}

// NewServer returns a pointer to a new gRPC server. It takes a function to register
// services, which allows metrics to be registered on each service.
func NewServer(registerServices func(*grpc.Server), opts ...grpc.ServerOption) *grpc.Server {
	logger := logrus.NewEntry(logrus.StandardLogger())
	initialize(logger)
	opts = append(opts,
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(unaryServerInterceptor(logger)...),
		),
		grpc.StreamInterceptor(
			grpc_middleware.ChainStreamServer(streamServerInterceptor(logger)...),
		),
	)
	server := grpc.NewServer(opts...)

	registerServices(server)
	grpc_prometheus.Register(server)

	return server
}
