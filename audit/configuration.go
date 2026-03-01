package audit

import "io"

type Configuration struct {
	AuditLoggerType string `config:"audit_logger_type" validate:"required,oneof=noop stdout"`
}

func NewConfiguration() *Configuration {
	return &Configuration{
		AuditLoggerType: "noop",
	}
}

type Option func(*options)

type options struct {
	output io.Writer
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
