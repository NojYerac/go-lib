package transport_test

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	nethttp "net/http"
	"time"

	"github.com/nojyerac/go-lib/auth"
	"github.com/nojyerac/go-lib/authz"
	"github.com/nojyerac/go-lib/log"
	. "github.com/nojyerac/go-lib/transport"
	libgrpc "github.com/nojyerac/go-lib/transport/grpc"
	"github.com/nojyerac/go-lib/transport/http"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	pb "google.golang.org/grpc/examples/features/proto/echo"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var _ = Describe("transport", func() {
	var (
		// clientCAs *x509.CertPool
		config     *Configuration
		s          Server
		g          *grpc.Server
		h          http.Server
		httpClient nethttp.Client
		ctx        context.Context
		cancel     context.CancelFunc
		err        error
	)
	BeforeEach(func() {
		httpClient = nethttp.Client{
			Transport: &nethttp.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true, //nolint:gosec //testing
				},
			},
		}
		ctx, cancel = context.WithCancel(context.Background())
		ctx = log.WithLogger(ctx, log.NewLogger(log.TestConfig))
	})

	Context("NewServer with auth-enabled transports", func() {
		BeforeEach(func() {
			config = &Configuration{
				NoTLS:    true,
				Hostname: "0.0.0.0",
				Port:     "9998",
			}

			validator := &authValidatorStub{}
			httpPolicies := authz.NewPolicyMap()
			httpPolicies.Set(authz.HTTPOperation(nethttp.MethodGet, "/api/private"), authz.RequireAny("reader"))

			h = http.NewServer(
				&http.Configuration{},
				http.WithAuthMiddleware(validator, httpPolicies),
				http.WithLogger(log.NewLogger(log.TestConfig)),
			)
			h.HandleFunc("GET /private", func(w nethttp.ResponseWriter, r *nethttp.Request) {
				claims, ok := auth.FromContext(r.Context())
				if !ok {
					w.WriteHeader(nethttp.StatusInternalServerError)
					return
				}
				w.WriteHeader(nethttp.StatusOK)
				_, _ = w.Write([]byte(claims.Subject))
			})

			grpcPolicies := authz.NewPolicyMap()
			grpcPolicies.Set(authz.GRPCOperation("/grpc.examples.echo.Echo/UnaryEcho"), authz.RequireAny("reader"))

			g = libgrpc.NewServer(
				func(s *grpc.Server) { pb.RegisterEchoServer(s, &echoSrv{}) },
				append(
					libgrpc.AuthServerOptions(validator, grpcPolicies),
					grpc.Creds(insecure.NewCredentials()), //nolint:gosec //testing
				)...,
			)
		})

		JustBeforeEach(func() {
			s, err = NewServer(config, WithHTTP(h), WithGRPC(g))
			Expect(err).NotTo(HaveOccurred())

			go func() {
				defer GinkgoRecover()
				Expect(s.Start(ctx)).To(Succeed())
			}()

			time.Sleep(200 * time.Millisecond)
		})

		It("returns 401 for protected HTTP route without token", func() {
			req, reqErr := nethttp.NewRequest("GET", "http://localhost:9998/api/private", nethttp.NoBody)
			Expect(reqErr).NotTo(HaveOccurred())

			res, doErr := httpClient.Do(req)
			Expect(doErr).NotTo(HaveOccurred())
			defer res.Body.Close()
			Expect(res.StatusCode).To(Equal(nethttp.StatusUnauthorized))
		})

		It("allows protected HTTP route with bearer token", func() {
			req, reqErr := nethttp.NewRequest("GET", "http://localhost:9998/api/private", nethttp.NoBody)
			Expect(reqErr).NotTo(HaveOccurred())
			req.Header.Set("Authorization", "Bearer good")

			res, doErr := httpClient.Do(req)
			Expect(doErr).NotTo(HaveOccurred())
			defer res.Body.Close()

			body, bodyErr := io.ReadAll(res.Body)
			Expect(bodyErr).NotTo(HaveOccurred())
			Expect(res.StatusCode).To(Equal(nethttp.StatusOK))
			Expect(body).To(Equal([]byte("auth-user")))
		})

		It("returns Unauthenticated for protected gRPC method without token", func() {
			cc, dialErr := grpc.NewClient(
				net.JoinHostPort(config.Hostname, config.Port),
				grpc.WithTransportCredentials(insecure.NewCredentials()),
			)
			Expect(dialErr).NotTo(HaveOccurred())
			defer cc.Close()

			client := pb.NewEchoClient(cc)
			_, callErr := client.UnaryEcho(context.Background(), &pb.EchoRequest{Message: "echo"})
			Expect(status.Code(callErr)).To(Equal(codes.Unauthenticated))
		})

		It("allows protected gRPC method with bearer token", func() {
			cc, dialErr := grpc.NewClient(
				net.JoinHostPort(config.Hostname, config.Port),
				grpc.WithTransportCredentials(insecure.NewCredentials()),
			)
			Expect(dialErr).NotTo(HaveOccurred())
			defer cc.Close()

			client := pb.NewEchoClient(cc)
			rpcCtx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", "Bearer good"))
			res, callErr := client.UnaryEcho(rpcCtx, &pb.EchoRequest{Message: "echo"})
			Expect(callErr).NotTo(HaveOccurred())
			Expect(res.Message).To(Equal("echo"))
		})
	})
	AfterEach(func() {
		cancel()
		time.Sleep(100 * time.Millisecond)
	})
	Context("NewServer", func() {
		BeforeEach(func() {
			config = &Configuration{
				NoTLS:    true,
				Hostname: "0.0.0.0",
				Port:     "9999",
			}
			h = http.NewServer(&http.Configuration{})
			g = libgrpc.NewServer(
				func(s *grpc.Server) { pb.RegisterEchoServer(s, &echoSrv{}) },
				grpc.Creds(insecure.NewCredentials()), //nolint:gosec //testing
			)
		})
		JustBeforeEach(func() {
			s, err = NewServer(config, WithHTTP(h), WithGRPC(g))
			Expect(err).NotTo(HaveOccurred())

			startCh := make(chan struct{})
			go func() {
				defer GinkgoRecover()
				defer close(startCh)
				Expect(s.Start(ctx)).To(Succeed())
			}()

			// Give cmux a moment to start accepting connections
			time.Sleep(200 * time.Millisecond)
		})
		It("routes to httpServer", func() {
			req, err := nethttp.NewRequest("GET", "http://localhost:9999/livez", nethttp.NoBody)
			Expect(err).NotTo(HaveOccurred())
			res, err := httpClient.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer res.Body.Close()
			body, err := io.ReadAll(res.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(body).To(Equal([]byte("ok")))
		})
		It("routes grpc requests", func() {
			cc, err := grpc.NewClient(
				net.JoinHostPort(config.Hostname, config.Port),
				grpc.WithTransportCredentials(insecure.NewCredentials()),
			)
			Expect(err).NotTo(HaveOccurred())
			defer cc.Close()

			client := pb.NewEchoClient(cc)
			res, err := client.UnaryEcho(ctx, &pb.EchoRequest{Message: "echo"})
			Expect(err).NotTo(HaveOccurred())
			Expect(res.Message).To(Equal("echo"))
		})
	})

	Context("NewServer with TLS", func() {
		BeforeEach(func() {
			config = &Configuration{
				Hostname: "0.0.0.0",
				Port:     "9999",
				PubCert:  "testdata/pub.crt",
				PrivKey:  "testdata/priv.key",
				RootCA:   "testdata/ca.crt",
			}
			h = http.NewServer(&http.Configuration{})
			g = libgrpc.NewServer(
				func(s *grpc.Server) { pb.RegisterEchoServer(s, &echoSrv{}) },
				grpc.Creds(insecure.NewCredentials()), //nolint:gosec //testing
			)
		})
		JustBeforeEach(func() {
			s, err = NewServer(config, WithHTTP(h), WithGRPC(g))
			Expect(err).NotTo(HaveOccurred())

			go func() {
				defer GinkgoRecover()
				Expect(s.Start(ctx)).To(Succeed())
			}()
		})
		It("routes to httpServer", func() {
			req, err := nethttp.NewRequest("GET", "https://localhost:9999/livez", nethttp.NoBody)
			Expect(err).NotTo(HaveOccurred())
			res, err := httpClient.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer res.Body.Close()
			body, err := io.ReadAll(res.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(body).To(Equal([]byte("ok")))
		})
		XIt("routes grpc requests", func() {
			cc, err := libgrpc.ClientConn(
				net.JoinHostPort(config.Hostname, config.Port),
				libgrpc.WithDialOptions(
					grpc.WithTransportCredentials(insecure.NewCredentials()), //nolint:gosec //testing
				),
			)
			Expect(err).NotTo(HaveOccurred())
			defer cc.Close()

			client := pb.NewEchoClient(cc)
			res, err := client.UnaryEcho(ctx, &pb.EchoRequest{Message: "echo"})
			Expect(err).NotTo(HaveOccurred())
			Expect(res.Message).To(Equal("echo"))
		})
	})
})

type echoSrv struct {
	pb.UnimplementedEchoServer
}

type authValidatorStub struct{}

func (s *authValidatorStub) Validate(_ context.Context, token string) (*auth.Claims, error) {
	if token != "good" {
		return nil, auth.ErrInvalidToken
	}
	return &auth.Claims{Subject: "auth-user", Roles: []string{"reader"}}, nil
}

func (e *echoSrv) UnaryEcho(_ context.Context, req *pb.EchoRequest) (*pb.EchoResponse, error) {
	return &pb.EchoResponse{Message: req.Message}, nil
}
