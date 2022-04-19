package config_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "source.rad.af/libs/go-lib/pkg/config"
	"source.rad.af/libs/go-lib/pkg/log"
)

var _ = Describe("ConfigLoader", func() {
	var (
		c   Loader
		err error
	)
	BeforeEach(func() {
		l := log.NewLogger(&log.Configuration{HumanFrendly: true, LogLevel: "fatal"})
		c = NewConfigLoader("test", WithLogger(l), WithArgs("-c", "./testdata"))
	})
	It("is testable", func() {
		Expect(err).NotTo(HaveOccurred())
		Expect(c).NotTo(BeNil())
		Expect(c.InitAndValidate()).To(Succeed())
	})
})
