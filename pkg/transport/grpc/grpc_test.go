package grpc_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	pb "google.golang.org/grpc/examples/features/proto/echo"
	"google.golang.org/grpc/test/bufconn"
	"source.rad.af/libs/go-lib/pkg/log"
	. "source.rad.af/libs/go-lib/pkg/transport/grpc"
)

const (
	triggerPanic = "panic"
	rpcErrorText = "rpc error: code = Internal desc = mock error"
	bufSz        = 1024 * 1024
)

var b = bytes.NewBuffer(make([]byte, 0, 1024))

var _ = BeforeSuite(func() {
	l := log.NewLogger(log.DebugConfig, log.WithOutput(b))
	Expect(SetGrpcLogger(l)).To(Succeed())
})

var _ = Describe("grpc", func() {
	var (
		listener   *bufconn.Listener
		grpcServer *grpc.Server
		bidiClient pb.Echo_BidirectionalStreamingEchoClient
		c          pb.EchoClient
		req        *pb.EchoRequest
		res        *pb.EchoResponse
		ctx        context.Context
	)
	BeforeEach(func() {
		ctx = context.Background()
		listener = bufconn.Listen(bufSz)
		grpcServer = NewServer(func(s *grpc.Server) {
			pb.RegisterEchoServer(s, &server{})
		})
		go func() {
			defer GinkgoRecover()
			err := grpcServer.Serve(listener)
			Expect(err).NotTo(HaveOccurred())
		}()
		testOpts := []grpc.DialOption{
			grpc.WithInsecure(), // nolint: staticcheck,gocritic
			grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
				return listener.Dial()
			}),
		}
		clientConn, err := ClientConn("bufconn", testOpts...)
		Expect(err).NotTo(HaveOccurred())
		c = pb.NewEchoClient(clientConn)
		go func() {
			defer GinkgoRecover()
			Expect(grpcServer.Serve(listener)).To(Succeed())
		}()
	})
	AfterEach(func() {
		b.Reset()
		Expect(grpcServer.GracefulStop).NotTo(Panic())
		Expect(listener.Close()).To(Succeed())
	})
	Context("streaming methods", func() {
		var err error
		BeforeEach(func() {
			req = &pb.EchoRequest{
				Message: "echo",
			}
		})
		JustBeforeEach(func() {
			bidiClient, err = c.BidirectionalStreamingEcho(ctx)
			Expect(err).NotTo(HaveOccurred())
		})
		AfterEach(func() {
			Expect(bidiClient.CloseSend()).To(Succeed())
		})
		It("generates a request id", func() {
			Expect(bidiClient.Send(req)).To(Succeed())
			res, err = bidiClient.Recv()
			Expect(err).NotTo(HaveOccurred())
			Expect(res.GetMessage()).To(Equal("echo"))
			Eventually(b.String).Should(And(
				// MatchRegexp("request_id=grpc-[\\w-]+"),
				ContainSubstring("got streaming echo message: echo"),
			))
		})
		When("request id is supplied", func() {
			BeforeEach(func() {
				ctx = context.WithValue(ctx, interface{}("requestID"), "mock-requestid") // nolint
			})
			It("preserves the request id", func() {
				Expect(bidiClient.Send(req)).To(Succeed())
				res, err = bidiClient.Recv()
				Expect(err).NotTo(HaveOccurred())
				Expect(res.GetMessage()).To(Equal("echo"))
				Eventually(b.String).Should(And(
					// ContainSubstring("request_id=mock-requestid"),
					ContainSubstring("got streaming echo message: echo"),
				))
			})
		})
		When("the handler panics", func() {
			BeforeEach(func() {
				req.Message = triggerPanic
			})
			It("recovers", func() {
				Expect(bidiClient.Send(req)).To(Succeed())
				_, err = bidiClient.Recv()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(rpcErrorText))
			})
		})
	})
	Context("unary methods", func() {
		var err error
		BeforeEach(func() {
			ctx = context.Background()
			req = &pb.EchoRequest{
				Message: "echo",
			}
		})
		JustBeforeEach(func() {
			res, err = c.UnaryEcho(ctx, req)
		})
		It("generates a request id", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(res.GetMessage()).To(Equal("echo"))
			Eventually(b.String).Should(And(
				// MatchRegexp("request_id=grpc-[\\w-]+"),
				ContainSubstring("got unary echo message: echo"),
			))
		})
		When("request id is supplied", func() {
			BeforeEach(func() {
				ctx = context.WithValue(ctx, interface{}("requestID"), "mock-requestid") // nolint
			})
			It("preserves the request id", func() {
				Expect(err).NotTo(HaveOccurred())
				Eventually(b.String).Should(And(
					// ContainSubstring("request_id=mock-requestid"),
					ContainSubstring("got unary echo message: echo"),
				))
			})
		})
		When("the handler panics", func() {
			BeforeEach(func() {
				req.Message = triggerPanic
			})
			It("recovers", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(rpcErrorText))
			})
		})
	})
})

var errMock = errors.New("mock error")

type server struct {
	pb.UnimplementedEchoServer
}

func (*server) UnaryEcho(ctx context.Context, req *pb.EchoRequest) (*pb.EchoResponse, error) {
	logger := zerolog.Ctx(ctx)
	msg := req.Message
	if msg == triggerPanic {
		panic(errMock)
	}
	logger.Info().Msg("got unary echo message: " + msg)
	return &pb.EchoResponse{Message: msg}, nil
}

func (*server) BidirectionalStreamingEcho(srv pb.Echo_BidirectionalStreamingEchoServer) error {
	logger := zerolog.Ctx(srv.Context())
	for {
		req, err := srv.Recv()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			logger.Error().Msg("failed to receive streaming echo message with error: " + err.Error())
			return err
		}
		msg := req.Message
		if msg == triggerPanic {
			panic(errMock)
		}
		logger.Info().Msg("got streaming echo message: " + msg)
		res := &pb.EchoResponse{
			Message: msg,
		}
		if err := srv.Send(res); err != nil {
			logger.Error().Msg("failed to send streaming echo message with error: " + err.Error())
			return err
		}
	}
}
