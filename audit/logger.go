package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// AuditLogger provides the simplest possible interface for audit logging.
type AuditLogger interface {
	Log(ctx context.Context, action string, details map[string]any) error
}

// NewAuditLogger creates an AuditLogger.
// The first and current implementation is a no-op logger.
func NewAuditLogger(cfg *Configuration, opts ...Option) (AuditLogger, error) {
	if cfg == nil {
		cfg = NewConfiguration()
	}

	o := new(options)
	for _, applyOpt := range opts {
		if applyOpt == nil {
			continue
		}
		applyOpt(o)
	}

	if o.output == nil {
		o.output = os.Stdout
	}

	if cfg.AuditLoggerType == "noop" {
		return noopAuditLogger{}, nil
	}
	if cfg.AuditLoggerType == "stdout" {
		return &stdoutAuditLogger{output: o.output}, nil
	}
	return nil, fmt.Errorf("unsupported audit logger type: %s", cfg.AuditLoggerType)
}

type noopAuditLogger struct{}

func (noopAuditLogger) Log(_ context.Context, _ string, _ map[string]any) error {
	return nil
}

type stdoutAuditLogger struct {
	mu     sync.Mutex
	output io.Writer
}

type stdoutLogRecord struct {
	Timestamp string         `json:"timestamp"`
	Action    string         `json:"action"`
	Details   map[string]any `json:"details,omitempty"`
}

func (s *stdoutAuditLogger) Log(_ context.Context, action string, details map[string]any) error {
	record := stdoutLogRecord{
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Action:    action,
		Details:   details,
	}

	payload, err := json.MarshalIndent(record, "", "  ")
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
