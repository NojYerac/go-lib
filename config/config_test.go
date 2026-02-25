package config_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ConfigLoader", func() {
	It("is testable", func() {
		Expect(c).NotTo(BeNil())
		Expect(c.InitAndValidate()).To(Succeed())
	})
})
