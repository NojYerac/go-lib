package app_test

import (
	"context"
	"testing"
	"time"
)

// TestRunSucceeds is a basic smoke test.  Wire up a real in-process invocation
// of app.Run using test config overrides and cancel after a short deadline to
// verify clean shutdown.
func TestRunSucceeds(t *testing.T) {
	t.Setenv("EXAMPLE_NO_TLS", "true")
	t.Setenv("EXAMPLE_PORT", "0")
	t.Setenv("EXAMPLE_SERVICE_NAME", "example-test")
	t.Setenv("EXAMPLE_LOG_LEVEL", "error")
	t.Setenv("EXAMPLE_EXPORTER_TYPE", "noop")
	t.Setenv("EXAMPLE_HEALTHCHECK_CHECK_INTERVAL", "1s")

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	err := Run(ctx)
	if err != nil && err != context.DeadlineExceeded {
		t.Fatalf("Run returned unexpected error: %v", err)
	}
}
