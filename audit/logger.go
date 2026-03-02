package audit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
)

// AuditLogger provides the simplest possible interface for audit logging.
type AuditLogger interface {
	LogChange(ctx context.Context, actorID, action string, before, after map[string]any) error
	Log(ctx context.Context, actorID, action string, details map[string]any) error
}

// NewAuditLogger creates an AuditLogger.
// The first and current implementation is a no-op logger.
func NewAuditLogger(cfg *Configuration, opts ...Option) (AuditLogger, error) {
	if cfg == nil {
		cfg = NewConfiguration()
	}

	o := &options{
		// Default validator with struct tag support
		validator: validator.New(),
	}
	for _, applyOpt := range opts {
		if applyOpt == nil {
			continue
		}
		applyOpt(o)
	}

	if o.output == nil {
		o.output = os.Stdout
	}

	if o.now == nil {
		o.now = time.Now
	}
	if o.httpClient == nil {
		o.httpClient = &http.Client{Timeout: 5 * time.Second}
	}

	switch cfg.AuditLoggerType {
	case "noop":
		return noopAuditLogger{}, nil
	case "stdout":
		return &stdoutAuditLogger{
			v:               o.validator,
			output:          o.output,
			now:             o.now,
			maxPayloadBytes: cfg.MaxPayloadBytes,
		}, nil
	case "http":
		baseURL := cfg.AuditLoggerURL
		if strings.TrimSpace(o.httpBase) != "" {
			baseURL = o.httpBase
		}
		baseURL = strings.TrimSpace(baseURL)
		if baseURL == "" {
			return nil, fmt.Errorf("audit logger url is required for http logger")
		}

		return &httpAuditLogger{
			v:               o.validator,
			client:          o.httpClient,
			endpoint:        strings.TrimRight(baseURL, "/") + "/api/auditlog",
			now:             o.now,
			maxPayloadBytes: cfg.MaxPayloadBytes,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported audit logger type: %s", cfg.AuditLoggerType)
	}
}

type noopAuditLogger struct{}

func (noopAuditLogger) LogChange(_ context.Context, _, _ string, _, _ map[string]any) error {
	return nil
}

func (noopAuditLogger) Log(_ context.Context, _, _ string, _ map[string]any) error {
	return nil
}

type stdoutAuditLogger struct {
	v               *validator.Validate
	mu              sync.Mutex
	output          io.Writer
	now             func() time.Time
	maxPayloadBytes int
}

func (s *stdoutAuditLogger) LogChange(ctx context.Context, actorID, action string, before, after map[string]any) error {
	details := processDetails(before, after)

	return s.Log(ctx, actorID, action, details)
}

func (s *stdoutAuditLogger) Log(_ context.Context, actorID, action string, details map[string]any) error {
	evt := event{
		ActorID:   actorID,
		Action:    action,
		Details:   details,
		Timestamp: s.now(),
	}

	if err := s.v.Struct(evt); err != nil {
		return validationErr(err)
	}

	detailsPayload, err := json.Marshal(evt.Details)
	if err != nil {
		return err
	}
	if s.maxPayloadBytes > 0 && len(detailsPayload) > s.maxPayloadBytes {
		return fmt.Errorf("%w: got=%d max=%d", ErrPayloadTooLarge, len(detailsPayload), s.maxPayloadBytes)
	}

	payload, err := json.MarshalIndent(evt, "", "  ")
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, err := s.output.Write(payload); err != nil {
		return err
	}
	if _, err := s.output.Write([]byte("\n")); err != nil {
		return err
	}

	return nil
}

type httpAuditLogger struct {
	v               *validator.Validate
	client          *http.Client
	endpoint        string
	now             func() time.Time
	maxPayloadBytes int
}

func (h *httpAuditLogger) LogChange(ctx context.Context, actorID, action string, before, after map[string]any) error {
	details := processDetails(before, after)
	return h.Log(ctx, actorID, action, details)
}

func (h *httpAuditLogger) Log(ctx context.Context, actorID, action string, details map[string]any) error {
	evt := event{
		ActorID:   actorID,
		Action:    action,
		Details:   details,
		Timestamp: h.now(),
	}

	if err := h.v.Struct(evt); err != nil {
		return validationErr(err)
	}

	payload, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	if h.maxPayloadBytes > 0 && len(payload) > h.maxPayloadBytes {
		return fmt.Errorf("%w: got=%d max=%d", ErrPayloadTooLarge, len(payload), h.maxPayloadBytes)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.endpoint, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("audit http logger request failed with status %d", resp.StatusCode)
	}

	return nil
}
