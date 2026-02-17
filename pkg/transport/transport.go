package transport

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net"
	"os"

	"github.com/rs/zerolog"
	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"

	libhttp "github.com/nojyerac/go-lib/pkg/transport/http"
)

type Configuration struct {
	PubCert  string `config:"tls_public_cert" validate:"required,file"`
	PrivKey  string `config:"tls_private_key" validate:"required,file"`
	RootCA   string `config:"tls_root_ca" validate:"required,file"`
	Hostname string `config:"hostname" validate:"required"`
	Port     string `config:"port" validate:"required"`
}

func NewConfiguration() *Configuration {
	return &Configuration{
		Hostname: "0.0.0.0",
		Port:     "443",
	}
}

func tlsConfig(config *Configuration) (*tls.Config, error) {
	suites := []uint16{
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
		tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
		tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
	}
	cert, err := tls.LoadX509KeyPair(config.PubCert, config.PrivKey)
	if err != nil {
		return nil, err
	}
	rootCAs := x509.NewCertPool()
	if ca, err := os.ReadFile(config.RootCA); err == nil {
		rootCAs.AppendCertsFromPEM(ca)
	}
	return &tls.Config{
		RootCAs:                  rootCAs,
		Certificates:             []tls.Certificate{cert},
		MinVersion:               tls.VersionTLS12,
		PreferServerCipherSuites: true,
		CipherSuites:             suites,
	}, nil
}

type Server interface {
	Start(context.Context) error
}

type server struct {
	grpcServer *grpc.Server
	httpServer libhttp.Server
	listener   net.Listener
}

func NewTLSServer(config *Configuration, opts ...Option) (Server, error) {
	conf, err := tlsConfig(config)
	if err != nil {
		return nil, err
	}
	target := net.JoinHostPort(config.Hostname, config.Port)
	listener, err := tls.Listen("tcp", target, conf)
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

// Start starts the server and blocks until the context is canceled
func (s *server) Start(ctx context.Context) error {
	logger := zerolog.Ctx(ctx)

	m := cmux.New(s.listener)

	if s.httpServer != nil {
		httpListener := m.Match(cmux.HTTP1())
		go func() {
			logger.Info().Msg("starting HTTP server")
			if err := s.httpServer.Listen(httpListener); err != nil && err != cmux.ErrServerClosed {
				logger.Panic().Err(err).Msg("HTTP server failed")
			}
		}()
	}

	if s.grpcServer != nil {
		grpcListener := m.MatchWithWriters(
			cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"),
		)
		go func() {
			logger.Info().Msg("starting gRPC server")
			if err := s.grpcServer.Serve(grpcListener); err != nil && err != cmux.ErrServerClosed {
				logger.Panic().Err(err).Msg("gRPC server failed")
			}
		}()
	}

	go func() {
		<-ctx.Done()
		logger.Info().Msg("stopping server")
		if s.grpcServer != nil {
			s.grpcServer.GracefulStop()
		}
		m.Close()
		if err := s.listener.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
			logger.Panic().Err(err).Msg("failed to stop server")
		}
	}()

	err := m.Serve()
	if errors.Is(err, net.ErrClosed) {
		return nil
	}
	return err
}
