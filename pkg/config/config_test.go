package config_test

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rs/zerolog"
	. "source.rad.af/libs/go-lib/pkg/config"
)

var _ = Describe("ConfigLoader", func() {
	var (
		c   ConfigLoader
		err error
	)
	BeforeEach(func() {
		l := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).Level(zerolog.FatalLevel)
		c = NewConfigLoader("test", WithLogger(l), WithArgs("-c", "./testdata"))
	})
	It("is testable", func() {
		Expect(err).NotTo(HaveOccurred())
		Expect(c).NotTo(BeNil())
		Expect(c.InitAndValidate()).To(Succeed())
	})
})
