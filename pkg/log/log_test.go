package log_test

import (
	"bytes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rs/zerolog"
	. "source.rad.af/libs/go-lib/pkg/log"
)

var _ = Describe("log", func() {
	var (
		l zerolog.Logger
		b *bytes.Buffer
	)
	BeforeEach(func() {
		b = &bytes.Buffer{}
		c := NewConfiguration()
		c.HumanFrendly = true
		c.ServiceName = "test-service"
		l = NewLogger(c, WithOutput(b))
	})
	AfterEach(func() {
		b.Reset()
	})
	It("", func() {
		l.Debug().Msg("log")
		Expect(b.Len()).To(BeZero())
		l.Info().Msg("log")
		Expect(b.String()).To(And(
			ContainSubstring("INF"),
			ContainSubstring("log"),
			ContainSubstring("service="),
			ContainSubstring("test-service")))
	})
})
