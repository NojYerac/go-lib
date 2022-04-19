package version_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "source.rad.af/libs/go-lib/pkg/version"
)

var _ = Describe("Version", func() {
	var v Version
	BeforeEach(func() {
		SetSemVer("0.0.1-test")
		SetServiceName("test-service")
		v = GetVersion()
	})
	It("is testable", func() {
		Expect(v.SemVer).To(Equal("0.0.1-test"))
	})
})
