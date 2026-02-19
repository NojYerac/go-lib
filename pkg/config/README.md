# Config Package

The **config** package provides a flexible configuration loader that supports environment variables, flags, and file‑based configuration (YAML, JSON, TOML, etc.). It is used by many other packages in the library.

## Configuration

The package defines a `Configuration` struct used to hold common config values:

- `config_dir`: Use the `-c` flag to set the directory to search for configuration files.
- `log_config_on_init`: **Warning this may leak secrets** Set to `True` to log the full configuration during initialization

The `NewConfigLoader` function creates a `Loader` that can be used to register structs and initialise configuration with options:

- `WithArgs`: allows for testing by overriding the actual os.Args array with mock values
- `WithLogger`: override the default logger.

### Custom Validators

Custom validation functions can be added via `RegisterCustomValidator(tag string, fn validator.Func)`.

## Usage

1. **Create a configuration struct** (can be a custom struct with `config` tags).
2. **Register the struct** with the loader.
3. **Call `InitAndValidate()`** to load from env/flags/files.
4. The struct is populated and ready to use.

## Examples

```go
package main

import (
  "github.com/nojyerac/go-lib/pkg/config"
)

// create configuration with types, variable names, and validation.
type Configuration struct {
  ExporterType string  `config:"trace_exporter_type" validate:"oneof=jaeger stdout file noop"`
  SampleRatio  float64 `config:"trace_sample_ratio" validate:"required_if=ExporterType jaeger,max=1,min=0"`
  FilePath     string  `config:"trace_file_path" validate:"omitempty,required_if=ExporterType file,file"`
}

// create a configuration (pointer to struct) with defaults.
func NewConfiguration() *Configuration {
  return &Configuration{
    ExporterType: "stdout",
    SampleRatio:  0.5,
  }
}

func main() {
  cfg := NewConfiguration()
  loader := config.NewConfigLoader("example") // env var names are prefixed with EXAMPLE_
  loader.RegisterConfig(cfg)
  if err := loader.InitAndValidate(); err != nil {
      panic(err)
  }
}
```
