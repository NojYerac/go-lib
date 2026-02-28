package grpc

import (
	"context"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/nojyerac/go-lib/auth"
	"github.com/nojyerac/go-lib/authz"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func AuthServerOptions(validator auth.Validator, policies authz.PolicyMap) []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(AuthUnaryServerInterceptor(validator, policies)),
		grpc.ChainStreamInterceptor(AuthStreamServerInterceptor(validator, policies)),
	}
}

func AuthUnaryServerInterceptor(validator auth.Validator, policies authz.PolicyMap) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		requirement, ok := policies.Requirement(authz.GRPCOperation(info.FullMethod))
		if !ok {
			return handler(ctx, req)
		}

		claims, err := authenticateClaims(ctx, validator)
		if err != nil {
			return nil, grpcAuthError(err)
		}
		if err := authz.Authorize(claims, requirement); err != nil {
			return nil, grpcAuthError(err)
		}

		return handler(auth.WithClaims(ctx, claims), req)
	}
}

func AuthStreamServerInterceptor(validator auth.Validator, policies authz.PolicyMap) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		requirement, ok := policies.Requirement(authz.GRPCOperation(info.FullMethod))
		if !ok {
			return handler(srv, ss)
		}

		claims, err := authenticateClaims(ss.Context(), validator)
		if err != nil {
			return grpcAuthError(err)
		}
		if err := authz.Authorize(claims, requirement); err != nil {
			return grpcAuthError(err)
		}

		wrapped := grpc_middleware.WrapServerStream(ss)
		wrapped.WrappedContext = auth.WithClaims(ss.Context(), claims)
		return handler(srv, wrapped)
	}
}

func authenticateClaims(ctx context.Context, validator auth.Validator) (*auth.Claims, error) {
	token, err := bearerTokenFromIncomingMetadata(ctx)
	if err != nil {
		return nil, err
	}
	return validator.Validate(ctx, token)
}

func bearerTokenFromIncomingMetadata(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", auth.ErrMissingToken
	}
	header := md.Get("authorization")
	if len(header) < 1 {
		return "", auth.ErrMissingToken
	}
	return auth.BearerToken(header[0])
}

func grpcAuthError(err error) error {
	return status.Error(auth.GRPCCode(err), err.Error())
}
