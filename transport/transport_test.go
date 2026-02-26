package transport_test

import (
	"context"
	"crypto/tls"
	"io"
	nethttp "net/http"
	"time"

	"github.com/nojyerac/go-lib/log"
	. "github.com/nojyerac/go-lib/transport"
	libgrpc "github.com/nojyerac/go-lib/transport/grpc"
	"github.com/nojyerac/go-lib/transport/http"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "google.golang.org/grpc/examples/features/proto/echo"
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
		ctx = log.WithLogger(ctx, log.NewLogger(log.NewConfiguration()))
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
				Port:     "8080",
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
			req, err := nethttp.NewRequest("GET", "http://localhost:8080/livez", nethttp.NoBody)
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
				"localhost:8080",
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

	Context("NewTLSServer", func() {
		BeforeEach(func() {
			config = &Configuration{
				Hostname: "0.0.0.0",
				Port:     "8443",
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
			s, err = NewTLSServer(config, WithHTTP(h), WithGRPC(g))
			Expect(err).NotTo(HaveOccurred())

			go func() {
				defer GinkgoRecover()
				Expect(s.Start(ctx)).To(Succeed())
			}()

			// Give cmux a moment to start accepting connections
			time.Sleep(200 * time.Millisecond)
		})
		It("routes to httpServer", func() {
			req, err := nethttp.NewRequest("GET", "https://localhost:8443/livez", nethttp.NoBody)
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
				"localhost:8443",
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

func (e *echoSrv) UnaryEcho(_ context.Context, req *pb.EchoRequest) (*pb.EchoResponse, error) {
	return &pb.EchoResponse{Message: req.Message}, nil
}
