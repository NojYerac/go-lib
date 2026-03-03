package auth_test

import (
	. "github.com/nojyerac/go-lib/auth"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("BearerToken", func() {
	It("extracts bearer token", func() {
		token, err := BearerToken("Bearer abc123")
		Expect(err).NotTo(HaveOccurred())
		Expect(token).To(Equal("abc123"))
	})

	It("accepts case-insensitive bearer scheme", func() {
		token, err := BearerToken("bearer abc123")
		Expect(err).NotTo(HaveOccurred())
		Expect(token).To(Equal("abc123"))
	})

	It("returns missing token for empty value", func() {
		_, err := BearerToken("   ")
		Expect(err).To(MatchError(ErrMissingToken))
	})

	It("returns invalid token for malformed value", func() {
		_, err := BearerToken("abc123")
		Expect(err).To(MatchError(ErrInvalidToken))
	})
})
