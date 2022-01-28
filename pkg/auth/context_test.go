package auth_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "source.rad.af/libs/go-lib/pkg/auth"
)

var _ = Describe("context", func() {
	var (
		ctx  context.Context
		user *User
		err  error
	)
	BeforeEach(func() {
		ctx = context.Background()
		user = nil
		err = nil
	})
	Describe("WithCtxt", func() {
		JustBeforeEach(func() {
			ctx = WithCtx(ctx, user)
		})
		When("user is nil", func() {
			It("returns the ctx", func() {
				Expect(ctx).NotTo(BeNil())
			})
		})
		When("user is present", func() {
			BeforeEach(func() {
				user = &User{
					UserID:   99,
					Username: "testuser",
				}
			})
			It("returns the ctx", func() {
				Expect(ctx).NotTo(BeNil())
			})
		})
	})

	Describe("FromCtxt", func() {
		JustBeforeEach(func() {
			user, err = FromCtx(ctx)
		})
		When("user is absent", func() {
			It("returns an unauthenticated error", func() {
				Expect(err).To(MatchError("unauthenticated"))
				Expect(user).To(BeNil())
			})
		})
		When("user is present", func() {
			BeforeEach(func() {
				ctx = WithCtx(ctx, &User{
					UserID:   99,
					Username: "testuser",
				})
			})
			It("returns the user", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(user).To(Equal(&User{
					UserID:   99,
					Username: "testuser",
				}))
			})
		})
	})
})
