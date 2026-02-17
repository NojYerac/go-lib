package grpc_test

import (
	"bytes"
	"io"

	"github.com/nojyerac/go-lib/pkg/log"
	. "github.com/nojyerac/go-lib/pkg/transport/grpc"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rs/zerolog"
)

var _ = Describe("GrpcV2Logger", func() {
	var (
		b  *bytes.Buffer
		l  *zerolog.Logger
		gl *GrpcV2Logger
	)
	BeforeEach(func() {
		b = bytes.NewBuffer(make([]byte, 0, 1024))
		config := &log.Configuration{
			ServiceName: "grpclogger",
			LogLevel:    "debug",
		}
		l = log.NewLogger(config, log.WithOutput(b))
		gl = &GrpcV2Logger{L: l}
	})
	AfterEach(func() {
		b.Reset()
	})
	Describe("log methods", func() {
		It("logs Info level", func() {
			Expect(gl).NotTo(BeNil())
			gl.Info("info")
			gl.Infof("%s", "info")
			gl.Infoln("info")
			logBytes, err := io.ReadAll(b)
			Expect(err).NotTo(HaveOccurred())
			logs := string(logBytes)
			Expect(logs).To(ContainSubstring("info"))
		})
		It("logs Warn level", func() {
			Expect(gl).NotTo(BeNil())
			gl.Warning("warn")
			gl.Warningf("%s", "warn")
			gl.Warningln("warn")
			logBytes, err := io.ReadAll(b)
			Expect(err).NotTo(HaveOccurred())
			logs := string(logBytes)
			Expect(logs).To(ContainSubstring("warn"))
		})
		It("logs error level", func() {
			Expect(gl).NotTo(BeNil())
			gl.Error("error")
			gl.Errorf("%s", "error")
			gl.Errorln("error")
			logBytes, err := io.ReadAll(b)
			Expect(err).NotTo(HaveOccurred())
			logs := string(logBytes)
			Expect(logs).To(ContainSubstring("error"))
		})
	})

})
