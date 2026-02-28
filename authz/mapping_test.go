package authz_test

import (
	. "github.com/nojyerac/go-lib/authz"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("PolicyMap", func() {
	Describe("Set and Requirement", func() {
		It("stores and resolves requirements by normalized operation key", func() {
			policies := NewPolicyMap()
			policies.Set("  GET /v1/flags  ", RequireAny("reader"))

			req, ok := policies.Requirement("GET /v1/flags")
			Expect(ok).To(BeTrue())
			Expect(req.AnyOf).To(Equal([]string{"reader"}))
		})

		It("ignores empty operation keys", func() {
			policies := NewPolicyMap()
			policies.Set(" ", RequireAny("reader"))

			_, ok := policies.Requirement(" ")
			Expect(ok).To(BeFalse())
		})

		It("returns false when map is nil", func() {
			var policies PolicyMap

			policies.Set("GET /v1/flags", RequireAny("reader"))
			_, ok := policies.Requirement("GET /v1/flags")
			Expect(ok).To(BeFalse())
		})
	})
})

var _ = Describe("Operation Helpers", func() {
	Describe("HTTPOperation", func() {
		It("normalizes method and preserves path", func() {
			Expect(HTTPOperation(" get ", " /v1/flags ")).To(Equal("GET /v1/flags"))
		})

		It("returns empty string when method is empty", func() {
			Expect(HTTPOperation("", "/v1/flags")).To(BeEmpty())
		})

		It("returns empty string when path is empty", func() {
			Expect(HTTPOperation("GET", "")).To(BeEmpty())
		})
	})

	Describe("GRPCOperation", func() {
		It("trims full method", func() {
			Expect(GRPCOperation("  /flag.v1.FlagService/GetFlag  ")).To(Equal("/flag.v1.FlagService/GetFlag"))
		})

		It("returns empty string for blank method", func() {
			Expect(GRPCOperation(" ")).To(BeEmpty())
		})
	})
})
