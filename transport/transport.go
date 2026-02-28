package transport

import (
	"context"
	"crypto/tls"
	"errors"
	"net"

	"github.com/nojyerac/go-lib/log"
	libhttp "github.com/nojyerac/go-lib/transport/http"
	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"
)

type Configuration struct {
	NoTLS    bool   `config:"no_tls"`
	PubCert  string `config:"tls_public_cert" validate:"required_unless=NoTLS true"`
	PrivKey  string `config:"tls_private_key" validate:"required_unless=NoTLS true"`
	RootCA   string `config:"tls_root_ca"`
	Hostname string `config:"hostname" validate:"required,hostname_rfc1123"`
	Port     string `config:"port" validate:"required,numeric,min=1,max=65535"`
}

func NewConfiguration() *Configuration {
	return &Configuration{
		Hostname: "0.0.0.0",
		Port:     "80",
	}
}

type Server interface {
	Start(context.Context) error
}

type server struct {
	grpcServer *grpc.Server
	httpServer libhttp.Server
	listener   net.Listener
}

func getListener(target string, config *Configuration) (net.Listener, error) {
	if config.NoTLS {
		return net.Listen("tcp", target)
	}
	cert, err := tls.LoadX509KeyPair(config.PubCert, config.PrivKey)
	if err != nil {
		return nil, err
	}
	return tls.Listen("tcp", target, &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	})
}

func NewServer(config *Configuration, opts ...Option) (Server, error) {
	target := net.JoinHostPort(config.Hostname, config.Port)
	listener, err := getListener(target, config)
	if err != nil {
		return nil, err
	}
	s := &server{
		listener: listener,
	}
	for _, applyOpt := range opts {
		applyOpt(s)
	}
	return s, nil
}

type Option func(s *server)

func WithHTTP(h libhttp.Server) Option {
	return func(s *server) {
		s.httpServer = h
	}
}

func WithGRPC(g *grpc.Server) Option {
	return func(s *server) {
		s.grpcServer = g
	}
}

func WithListener(l net.Listener) Option {
	return func(s *server) {
		s.listener = l
	}
}

// Start starts the server and blocks until the context is canceled
func (s *server) Start(ctx context.Context) error {
	logger := log.FromContext(ctx)

	m := cmux.New(s.listener)

	var grpcListener net.Listener
	var httpListener net.Listener

	// Register gRPC matcher first (more specific)
	if s.grpcServer != nil {
		grpcListener = m.Match(cmux.HTTP2())
	}

	// Register HTTP matcher second (less specific, acts as fallback)
	if s.httpServer != nil {
		httpListener = m.Match(cmux.HTTP1Fast())
	}

	// Start gRPC server
	if s.grpcServer != nil && grpcListener != nil {
		go func() {
			if err := s.grpcServer.Serve(grpcListener); err != nil && err != grpc.ErrServerStopped {
				logger.WithError(err).Panic("gRPC server failed")
			}
		}()
	}

	// Start HTTP server
	if s.httpServer != nil && httpListener != nil {
		go func() {
			if err := s.httpServer.Listen(httpListener); err != nil && err != cmux.ErrServerClosed {
				logger.WithError(err).Panic("HTTP server failed")
			}
		}()
	}

	go func() {
		<-ctx.Done()
		if s.grpcServer != nil {
			s.grpcServer.GracefulStop()
		}
		m.Close()
		if err := s.listener.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
			logger.WithError(err).Panic("failed to stop server")
		}
	}()

	err := m.Serve()
	if errors.Is(err, net.ErrClosed) {
		return nil
	}
	return err
}
