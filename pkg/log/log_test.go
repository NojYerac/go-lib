package log_test

import (
	"bytes"
	"context"

	. "github.com/nojyerac/go-lib/pkg/log"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rs/zerolog"
)

var _ = Describe("log", func() {
	var (
		l *zerolog.Logger
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
	It("sets a default ctx logger", func() {
		SetDefaultCtxLogger(l)
		Expect(zerolog.Ctx(context.Background())).To(Equal(l))
	})
})
