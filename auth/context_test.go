package auth_test

import (
	"context"

	. "github.com/nojyerac/go-lib/auth"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Context", func() {
	It("stores and loads claims", func() {
		base := context.Background()
		in := &Claims{Subject: "user-1", Roles: []string{"reader"}}

		ctx := WithClaims(base, in)
		out, ok := FromContext(ctx)

		Expect(ok).To(BeTrue())
		Expect(out).To(Equal(in))
	})

	It("returns false when context has no claims", func() {
		_, ok := FromContext(context.Background())
		Expect(ok).To(BeFalse())
	})

	It("returns false for nil context", func() {
		var ctx context.Context
		_, ok := FromContext(ctx)
		Expect(ok).To(BeFalse())
	})

	It("returns false for nil claims in context", func() {
		ctx := WithClaims(context.Background(), nil)
		_, ok := FromContext(ctx)
		Expect(ok).To(BeFalse())
	})
})
