package audit_test

import (
	"bytes"
	"context"
	"encoding/json"

	. "github.com/nojyerac/go-lib/audit"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("NewAuditLogger", func() {
	It("returns a no-op logger", func() {
		cfg := NewConfiguration()
		logger, err := NewAuditLogger(cfg)
		Expect(err).ToNot(HaveOccurred())
		Expect(logger).ToNot(BeNil())
		err = logger.Log(context.Background(), "test_action", map[string]any{"key": "value"})
		Expect(err).ToNot(HaveOccurred())
	})

	It("logs pretty json to configured output for stdout logger", func() {
		cfg := NewConfiguration()
		cfg.AuditLoggerType = "stdout"

		var out bytes.Buffer
		logger, err := NewAuditLogger(cfg, WithOutput(&out))
		Expect(err).ToNot(HaveOccurred())

		err = logger.Log(context.Background(), "user.login", map[string]any{"user_id": "u-1"})
		Expect(err).ToNot(HaveOccurred())

		payload := out.String()
		Expect(payload).To(ContainSubstring("\n  \"action\": \"user.login\","))
		Expect(payload).To(ContainSubstring("\n  \"details\": {"))
		Expect(payload).To(HaveSuffix("}\n"))

		var decoded map[string]any
		Expect(json.Unmarshal([]byte(payload), &decoded)).To(Succeed())
		Expect(decoded["action"]).To(Equal("user.login"))
		details, ok := decoded["details"].(map[string]any)
		Expect(ok).To(BeTrue())
		Expect(details["user_id"]).To(Equal("u-1"))
		Expect(decoded).To(HaveKey("timestamp"))
	})
})
