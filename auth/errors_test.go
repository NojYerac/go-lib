package auth_test

import (
	"errors"
	"net/http"

	. "github.com/nojyerac/go-lib/auth"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc/codes"
)

var _ = Describe("Errors", func() {
	Describe("HTTPStatus", func() {
		It("maps permission errors to forbidden", func() {
			err := errors.Join(ErrPermissionDenied, errors.New("wrapped"))
			Expect(HTTPStatus(err)).To(Equal(http.StatusForbidden))
		})

		It("maps non-permission errors to unauthorized", func() {
			Expect(HTTPStatus(ErrInvalidToken)).To(Equal(http.StatusUnauthorized))
		})
	})

	Describe("GRPCCode", func() {
		It("maps permission errors to permission denied", func() {
			err := errors.Join(ErrPermissionDenied, errors.New("wrapped"))
			Expect(GRPCCode(err)).To(Equal(codes.PermissionDenied))
		})

		It("maps non-permission errors to unauthenticated", func() {
			Expect(GRPCCode(ErrInvalidToken)).To(Equal(codes.Unauthenticated))
		})
	})
})
