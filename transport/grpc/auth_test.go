package grpc_test

import (
	"context"

	"github.com/nojyerac/go-lib/auth"
	"github.com/nojyerac/go-lib/authz"
	. "github.com/nojyerac/go-lib/transport/grpc"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type grpcValidatorStub struct {
	claims *auth.Claims
	err    error
}

func (v *grpcValidatorStub) Validate(context.Context, string) (*auth.Claims, error) {
	if v.err != nil {
		return nil, v.err
	}
	return v.claims, nil
}

type streamStub struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *streamStub) SetHeader(metadata.MD) error {
	return nil
}

func (s *streamStub) SendHeader(metadata.MD) error {
	return nil
}

func (s *streamStub) SetTrailer(metadata.MD) {}

func (s *streamStub) Context() context.Context {
	return s.ctx
}

func (s *streamStub) SendMsg(any) error {
	return nil
}

func (s *streamStub) RecvMsg(any) error {
	return nil
}

var _ = Describe("Auth interceptors", func() {
	var (
		validator *grpcValidatorStub
		policies  authz.PolicyMap
	)

	BeforeEach(func() {
		validator = &grpcValidatorStub{claims: &auth.Claims{Subject: "user-1", Roles: []string{"reader"}}}
		policies = authz.NewPolicyMap()
		policies.Set(authz.GRPCOperation("/svc.Example/Read"), authz.RequireAny("reader"))
	})

	Describe("AuthServerOptions", func() {
		It("returns unary and stream options", func() {
			options := AuthServerOptions(validator, policies)
			Expect(options).To(HaveLen(2))
		})
	})

	Describe("AuthUnaryServerInterceptor", func() {
		It("allows unprotected method", func() {
			interceptor := AuthUnaryServerInterceptor(validator, policies)
			ctx := context.Background()
			called := false

			_, err := interceptor(
				ctx,
				nil,
				&grpc.UnaryServerInfo{FullMethod: "/svc.Example/Public"},
				func(context.Context, any) (any, error) {
					called = true
					return nil, nil
				},
			)

			Expect(err).NotTo(HaveOccurred())
			Expect(called).To(BeTrue())
		})

		It("returns unauthenticated when token is missing", func() {
			interceptor := AuthUnaryServerInterceptor(validator, policies)

			_, err := interceptor(
				context.Background(),
				nil,
				&grpc.UnaryServerInfo{FullMethod: "/svc.Example/Read"},
				func(context.Context, any) (any, error) { return nil, nil },
			)

			Expect(status.Code(err)).To(Equal(codes.Unauthenticated))
		})

		It("returns permission denied for insufficient roles", func() {
			validator.claims = &auth.Claims{Subject: "user-1", Roles: []string{"viewer"}}
			interceptor := AuthUnaryServerInterceptor(validator, policies)
			ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer token"))

			_, err := interceptor(
				ctx,
				nil,
				&grpc.UnaryServerInfo{FullMethod: "/svc.Example/Read"},
				func(context.Context, any) (any, error) { return nil, nil },
			)

			Expect(status.Code(err)).To(Equal(codes.PermissionDenied))
		})

		It("injects claims into handler context when authorized", func() {
			interceptor := AuthUnaryServerInterceptor(validator, policies)
			ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer token"))

			_, err := interceptor(
				ctx,
				nil,
				&grpc.UnaryServerInfo{FullMethod: "/svc.Example/Read"},
				func(handlerCtx context.Context, _ any) (any, error) {
					claims, ok := auth.FromContext(handlerCtx)
					Expect(ok).To(BeTrue())
					Expect(claims.Subject).To(Equal("user-1"))
					return nil, nil
				},
			)

			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("AuthStreamServerInterceptor", func() {
		It("injects claims into stream context when authorized", func() {
			interceptor := AuthStreamServerInterceptor(validator, policies)
			ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer token"))
			stream := &streamStub{ctx: ctx}

			err := interceptor(
				nil,
				stream,
				&grpc.StreamServerInfo{FullMethod: "/svc.Example/Read"},
				func(_ any, wrapped grpc.ServerStream) error {
					claims, ok := auth.FromContext(wrapped.Context())
					Expect(ok).To(BeTrue())
					Expect(claims.Subject).To(Equal("user-1"))
					return nil
				},
			)

			Expect(err).NotTo(HaveOccurred())
		})

		It("returns unauthenticated when metadata is missing", func() {
			interceptor := AuthStreamServerInterceptor(validator, policies)
			stream := &streamStub{ctx: context.Background()}

			err := interceptor(
				nil,
				stream,
				&grpc.StreamServerInfo{FullMethod: "/svc.Example/Read"},
				func(_ any, _ grpc.ServerStream) error { return nil },
			)

			Expect(status.Code(err)).To(Equal(codes.Unauthenticated))
		})
	})
})
