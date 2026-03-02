package audit_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/google/uuid"
	. "github.com/nojyerac/go-lib/audit"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func timeNow() time.Time {
	return time.Date(2026, 3, 1, 12, 0, 0, 123456789, time.UTC)
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

const (
	auditLoggerExampleURL = "https://audit.example.test"
)

var _ = Describe("NewAuditLogger", func() {
	var (
		cfg     *Configuration
		out     bytes.Buffer
		logger  AuditLogger
		err     error
		actorID = uuid.New().String()
	)
	JustBeforeEach(func() {
		logger, err = NewAuditLogger(cfg, WithOutput(&out), WithTimeNow(timeNow))
	})
	Describe("Noop Logger", func() {
		BeforeEach(func() {
			cfg = NewConfiguration()
		})
		It("returns a no-op logger", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(logger).ToNot(BeNil())
			err = logger.LogChange(context.Background(), actorID, "test_action", map[string]any{"key": "value"}, nil)
			Expect(err).ToNot(HaveOccurred())
			err = logger.Log(context.Background(), actorID, "test_action", map[string]any{"key": "value"})
			Expect(err).ToNot(HaveOccurred())
			Expect(out.Len()).To(Equal(0))
		})
	})

	Describe("Stdout Logger", func() {
		BeforeEach(func() {
			out.Reset()
			cfg = NewConfiguration()
			cfg.AuditLoggerType = "stdout"
		})

		It("logs pretty json to configured output for stdout logger", func() {
			err = logger.LogChange(context.Background(), actorID, "user.update", map[string]any{"user_id": "u-1"}, nil)
			Expect(err).ToNot(HaveOccurred())

			payload := out.String()
			Expect(payload).To(MatchJSON(`{
				"actorID": "` + actorID + `",
				"action": "user.update",
				"details": {
					"user_id": {
						"old_value": "u-1"
					}
				},
				"timestamp": "2026-03-01T12:00:00.123456789Z"
			}`))
			Expect(payload).To(HaveSuffix("\n"))
		})

		It("returns validation error for invalid event", func() {
			err = logger.Log(context.Background(), "bad-actor-id", "user.login", map[string]any{"user_id": "u-1"})
			Expect(err).To(MatchError(ErrInvalidEventActorID))
			err = logger.Log(context.Background(), actorID, "", map[string]any{"user_id": "u-1"})
			Expect(err).To(MatchError(ErrInvalidEventAction))
			err = logger.Log(context.Background(), actorID, "user.login", nil)
			Expect(err).To(MatchError(ErrInvalidEventDetails))
		})

		It("enforces configured bounded payload limit", func() {
			cfg.MaxPayloadBytes = 10
			logger, err = NewAuditLogger(cfg, WithOutput(&out), WithTimeNow(timeNow))
			Expect(err).NotTo(HaveOccurred())

			err = logger.Log(context.Background(), actorID, "user.login", map[string]any{"blob": "123456789012345"})
			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, ErrPayloadTooLarge)).To(BeTrue())
		})

		It("treats MaxPayloadBytes=0 as unlimited", func() {
			cfg.MaxPayloadBytes = 0
			logger, err = NewAuditLogger(cfg, WithOutput(&out), WithTimeNow(timeNow))
			Expect(err).NotTo(HaveOccurred())

			err = logger.Log(
				context.Background(),
				actorID,
				"user.login",
				map[string]any{"blob": "123456789012345678901234567890"},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(out.String()).To(ContainSubstring("\"action\": \"user.login\""))
		})
	})

	Describe("HTTP Logger", func() {
		BeforeEach(func() {
			out.Reset()
			cfg = NewConfiguration()
			cfg.AuditLoggerType = "http"
		})

		It("posts audit logs to /api/auditlog", func() {
			var received map[string]any
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				Expect(r.Method).To(Equal(http.MethodPost))
				Expect(r.URL.Path).To(Equal("/api/auditlog"))
				Expect(r.Header.Get("Content-Type")).To(Equal("application/json"))

				body, readErr := io.ReadAll(r.Body)
				Expect(readErr).NotTo(HaveOccurred())
				Expect(r.Body.Close()).To(Succeed())
				Expect(body).NotTo(BeEmpty())
				Expect(json.Unmarshal(body, &received)).To(Succeed())

				w.WriteHeader(http.StatusAccepted)
			}))
			defer server.Close()

			cfg.AuditLoggerURL = server.URL
			logger, err = NewAuditLogger(cfg, WithTimeNow(timeNow))
			Expect(err).NotTo(HaveOccurred())

			err = logger.Log(context.Background(), actorID, "user.login", map[string]any{"user_id": "u-1"})
			Expect(err).NotTo(HaveOccurred())
			Expect(received["actorID"]).To(Equal(actorID))
			Expect(received["action"]).To(Equal("user.login"))
			details, ok := received["details"].(map[string]any)
			Expect(ok).To(BeTrue())
			Expect(details).To(HaveKeyWithValue("user_id", "u-1"))
			Expect(received).To(HaveKey("timestamp"))
		})

		It("returns error when url is missing", func() {
			cfg.AuditLoggerURL = ""
			logger, err = NewAuditLogger(cfg)
			Expect(err).To(HaveOccurred())
			Expect(logger).To(BeNil())
		})

		It("returns error for non-2xx response", func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			}))
			defer server.Close()

			cfg.AuditLoggerURL = server.URL
			logger, err = NewAuditLogger(cfg, WithTimeNow(timeNow))
			Expect(err).NotTo(HaveOccurred())

			err = logger.Log(context.Background(), actorID, "user.login", map[string]any{"user_id": "u-1"})
			Expect(err).To(HaveOccurred())
		})

		It("uses WithHTTPClient for request execution", func() {
			cfg.AuditLoggerURL = auditLoggerExampleURL
			used := false
			client := &http.Client{
				Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
					used = true
					Expect(req.URL.String()).To(Equal("https://audit.example.test/api/auditlog"))
					Expect(req.Method).To(Equal(http.MethodPost))
					return &http.Response{
						StatusCode: http.StatusAccepted,
						Header:     make(http.Header),
						Body:       io.NopCloser(strings.NewReader("")),
					}, nil
				}),
			}

			logger, err = NewAuditLogger(cfg, WithHTTPClient(client), WithTimeNow(timeNow))
			Expect(err).NotTo(HaveOccurred())

			err = logger.Log(context.Background(), actorID, "user.login", map[string]any{"user_id": "u-1"})
			Expect(err).NotTo(HaveOccurred())
			Expect(used).To(BeTrue())
		})

		It("enforces bounded payload limit before HTTP post", func() {
			cfg.AuditLoggerURL = auditLoggerExampleURL
			cfg.MaxPayloadBytes = 10

			client := &http.Client{
				Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
					Fail("request should not be executed when payload is too large")
					return nil, nil
				}),
			}

			logger, err = NewAuditLogger(cfg, WithHTTPClient(client), WithTimeNow(timeNow))
			Expect(err).NotTo(HaveOccurred())

			err = logger.Log(context.Background(), actorID, "user.login", map[string]any{"blob": "123456789012345"})
			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, ErrPayloadTooLarge)).To(BeTrue())
		})
	})
})
