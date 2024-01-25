package transport_test

import (
	"context"
	"crypto/tls"
	"io"
	nethttp "net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	pb "google.golang.org/grpc/examples/features/proto/echo"
	"source.rad.af/libs/go-lib/pkg/log"
	. "source.rad.af/libs/go-lib/pkg/transport"
	"source.rad.af/libs/go-lib/pkg/transport/http"
)

var _ = Describe("transport", func() {
	var (
		config *Configuration
		s      Server
		g      *grpc.Server
		h      http.Server
		ctx    context.Context
		cancel context.CancelFunc
		err    error
	)
	BeforeEach(func() {
		config = &Configuration{
			Hostname: "0.0.0.0",
			Port:     "8443",
			PubCert:  "testdata/pub.crt",
			PrivKey:  "testdata/priv.key",
			RootCA:   "testdata/ca.crt",
		}
		ctx, cancel = context.WithCancel(context.Background())
		ctx = log.NewLogger(log.TestConfig).WithContext(ctx)
	})
	JustBeforeEach(func() {
		s, err = NewTLSServer(config, WithHTTP(h), WithGRPC(g))
		Expect(err).NotTo(HaveOccurred())

		go func() {
			defer GinkgoRecover()
			Expect(s.Start(ctx)).To(Succeed())
		}()
	})
	AfterEach(func() {
		cancel()
		time.Sleep(100 * time.Millisecond)
	})
	It("is testable", func() {
		Expect(true).To(BeTrue())
	})
	Context("httpServer", func() {
		BeforeEach(func() {
			h = http.NewServer(&http.Configuration{})
		})
		It("routes to httpServer", func() {
			req, err := nethttp.NewRequest("GET", "https://localhost:8443/ping", nethttp.NoBody)
			Expect(err).NotTo(HaveOccurred())
			httpClient := nethttp.Client{
				Transport: &nethttp.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: true, //nolint:gosec //testing
					},
				},
			}
			res, err := httpClient.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer res.Body.Close()
			body, err := io.ReadAll(res.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(body).To(Equal([]byte("pong")))
		})
	})
	Context("grpcServer", func() {
		BeforeEach(func() {
			g = grpc.NewServer()
			pb.RegisterEchoServer(g, &echoSrv{})
		})
		It("routes grpc requests", func() {
			cc, err := grpc.Dial("localhost:8443", grpc.WithTransportCredentials(
				credentials.NewTLS(
					&tls.Config{
						InsecureSkipVerify: true, //nolint:gosec //testing
					},
				),
			))
			Expect(err).NotTo(HaveOccurred())
			client := pb.NewEchoClient(cc)
			res, err := client.UnaryEcho(context.Background(), &pb.EchoRequest{Message: "echo"})
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
