# Log Package

The **log** package provides convenient wrappers around the zerolog logger and allows configuring log levels via environment variables or configuration files.

## Configuration

```go
// pkg/log/config.go
package log

import "github.com/rs/zerolog"

// Configuration holds log settings.
//
//   Level          string `config:"log_level"`
//   DisableColor   bool   `config:"disable_color"`
//   OutputFilePath string `config:"log_output_file"`
//
// The NewConfiguration helper returns defaults that write to stdout.
```

### Creating a Logger

```go
func NewLogger(cfg *Configuration) (*zerolog.Logger, error) {
    // Set level, color, and optionally write to a file.
}
```

## Usage

```go
import (
    "github.com/nojyerac/go-lib/pkg/log"
    "github.com/rs/zerolog"
)

func main() {
    cfg := log.NewConfiguration()
    l, err := log.NewLogger(cfg)
    if err != nil {
        zerolog.Ctx(context.Background()).Fatal().Err(err).Msg("log init")
    }
    l.Info().Msg("application started")
}
```

## Examples

- Log at different levels (debug, info, warn, error).
- Disable color output for CI environments.
- Redirect logs to a file for persistence.
