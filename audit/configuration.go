package audit

import (
	"io"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
)

type Configuration struct {
	AuditLoggerType string `config:"audit_logger_type" validate:"required,oneof=noop stdout http"`
	AuditLoggerURL  string `config:"audit_logger_url" validate:"required_if=AuditLoggerType http,omitempty,url"`
	MaxPayloadBytes int    `config:"audit_max_payload_bytes" validate:"gte=0"`
}

func NewConfiguration() *Configuration {
	return &Configuration{
		AuditLoggerType: "noop",
		MaxPayloadBytes: 4096, // 4 KB default max payload size
	}
}

type Option func(*options)

type options struct {
	validator  *validator.Validate
	output     io.Writer
	now        func() time.Time
	httpBase   string
	httpClient *http.Client
}

// WithOutput sets the destination writer for logger output.
func WithOutput(output io.Writer) Option {
	return func(options *options) {
		if options == nil {
			return
		}
		options.output = output
	}
}

func WithTimeNow(now func() time.Time) Option {
	return func(options *options) {
		if options == nil {
			return
		}
		options.now = now
	}
}

// WithHTTPBaseURL sets the HTTP audit server base URL for the http logger.
func WithHTTPBaseURL(baseURL string) Option {
	return func(options *options) {
		if options == nil {
			return
		}
		options.httpBase = baseURL
	}
}

// WithHTTPClient sets the HTTP client used by the http logger.
func WithHTTPClient(client *http.Client) Option {
	return func(options *options) {
		if options == nil {
			return
		}
		options.httpClient = client
	}
}
