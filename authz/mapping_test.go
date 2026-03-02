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

		It("matches policies with brace-style path params", func() {
			policies := NewPolicyMap()
			policies.Set(HTTPOperation("GET", "/v1/flags/{id}"), RequireAny("reader"))

			req, ok := policies.Requirement(HTTPOperation("GET", "/v1/flags/123"))
			Expect(ok).To(BeTrue())
			Expect(req.AnyOf).To(Equal([]string{"reader"}))
		})

		It("matches policies with colon-style path params", func() {
			policies := NewPolicyMap()
			policies.Set(HTTPOperation("GET", "/v1/flags/:id"), RequireAny("reader"))

			req, ok := policies.Requirement(HTTPOperation("GET", "/v1/flags/abc"))
			Expect(ok).To(BeTrue())
			Expect(req.AnyOf).To(Equal([]string{"reader"}))
		})

		It("prefers exact path match over param match", func() {
			policies := NewPolicyMap()
			policies.Set(HTTPOperation("GET", "/v1/flags/{id}"), RequireAny("reader"))
			policies.Set(HTTPOperation("GET", "/v1/flags/me"), RequireAny("admin"))

			req, ok := policies.Requirement(HTTPOperation("GET", "/v1/flags/me"))
			Expect(ok).To(BeTrue())
			Expect(req.AnyOf).To(Equal([]string{"admin"}))
		})

		It("does not match paths with different segment counts", func() {
			policies := NewPolicyMap()
			policies.Set(HTTPOperation("GET", "/v1/flags/{id}"), RequireAny("reader"))

			_, ok := policies.Requirement(HTTPOperation("GET", "/v1/flags/123/meta"))
			Expect(ok).To(BeFalse())
		})
	})
})

var _ = Describe("Operation Helpers", func() {
	Describe("HTTPOperation", func() {
		It("normalizes method and preserves path", func() {
			Expect(HTTPOperation(" get ", " /v1/flags ")).To(Equal("GET /v1/flags"))
		})

		It("normalizes query strings and trailing slashes", func() {
			Expect(HTTPOperation("GET", "/v1/flags/123/?expand=all")).To(Equal("GET /v1/flags/123"))
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
