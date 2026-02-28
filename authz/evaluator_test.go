package authz_test

import (
	. "github.com/nojyerac/go-lib/auth"
	. "github.com/nojyerac/go-lib/authz"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Evaluator", func() {
	Describe("RequireAny", func() {
		It("normalizes empty and duplicate roles", func() {
			req := RequireAny("", "reader", "reader", "admin")

			Expect(req.AnyOf).To(Equal([]string{"reader", "admin"}))
			Expect(req.AllOf).To(BeNil())
		})
	})

	Describe("RequireAll", func() {
		It("normalizes empty and duplicate roles", func() {
			req := RequireAll("", "reader", "reader", "admin")

			Expect(req.AllOf).To(Equal([]string{"reader", "admin"}))
			Expect(req.AnyOf).To(BeNil())
		})
	})

	Describe("IsEmpty", func() {
		It("returns true when no requirements are provided", func() {
			Expect(Requirement{}.IsEmpty()).To(BeTrue())
		})

		It("returns false when any requirement exists", func() {
			Expect(Requirement{AnyOf: []string{"reader"}}.IsEmpty()).To(BeFalse())
			Expect(Requirement{AllOf: []string{"admin"}}.IsEmpty()).To(BeFalse())
		})
	})

	Describe("SatisfiedBy", func() {
		It("allows empty requirement", func() {
			Expect(Requirement{}.SatisfiedBy(nil)).To(BeTrue())
		})

		It("denies non-empty requirement for nil claims", func() {
			Expect(RequireAny("reader").SatisfiedBy(nil)).To(BeFalse())
		})

		It("requires all roles in AllOf", func() {
			claims := &Claims{Roles: []string{"reader", "admin"}}

			Expect(RequireAll("reader", "admin").SatisfiedBy(claims)).To(BeTrue())
			Expect(RequireAll("reader", "owner").SatisfiedBy(claims)).To(BeFalse())
		})

		It("requires at least one role in AnyOf", func() {
			claims := &Claims{Roles: []string{"reader", "admin"}}

			Expect(RequireAny("owner", "admin").SatisfiedBy(claims)).To(BeTrue())
			Expect(RequireAny("owner", "writer").SatisfiedBy(claims)).To(BeFalse())
		})

		It("supports combined AnyOf and AllOf requirements", func() {
			claims := &Claims{Roles: []string{"reader", "admin"}}

			req := Requirement{AllOf: []string{"reader"}, AnyOf: []string{"writer", "admin"}}
			Expect(req.SatisfiedBy(claims)).To(BeTrue())

			req = Requirement{AllOf: []string{"reader"}, AnyOf: []string{"writer", "owner"}}
			Expect(req.SatisfiedBy(claims)).To(BeFalse())
		})
	})

	Describe("Authorize", func() {
		It("returns nil when requirement is satisfied", func() {
			claims := &Claims{Roles: []string{"reader"}}
			Expect(Authorize(claims, RequireAny("reader"))).To(Succeed())
		})

		It("returns permission denied when requirement is not satisfied", func() {
			claims := &Claims{Roles: []string{"reader"}}
			Expect(Authorize(claims, RequireAll("admin"))).To(MatchError(ErrPermissionDenied))
		})
	})
})
