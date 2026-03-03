package auth_test

import (
	. "github.com/nojyerac/go-lib/auth"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Claims", func() {
	Describe("HasRole", func() {
		It("matches role on claims", func() {
			claims := &Claims{Roles: []string{"reader", "admin"}}
			Expect(claims.HasRole("admin")).To(BeTrue())
			Expect(claims.HasRole("writer")).To(BeFalse())
		})

		It("returns false for nil claims", func() {
			var claims *Claims
			Expect(claims.HasRole("admin")).To(BeFalse())
		})
	})

	Describe("HasAnyRole", func() {
		It("matches at least one role", func() {
			claims := &Claims{Roles: []string{"reader", "admin"}}
			Expect(claims.HasAnyRole("writer", "admin")).To(BeTrue())
			Expect(claims.HasAnyRole("writer", "owner")).To(BeFalse())
		})

		It("returns false for nil claims", func() {
			var claims *Claims
			Expect(claims.HasAnyRole("admin")).To(BeFalse())
		})
	})
})
