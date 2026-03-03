# Config Package

The `config` package loads and validates configuration structs from:

- environment variables
- command-line flags
- config files in a directory (any file name containing `config.`)

It uses `config` tags for field mapping and `validate` tags for validation.

## API

### `type Loader interface`

- `RegisterConfig(interface{}) error`
- `InitAndValidate() error`

### `NewConfigLoader(prefix string, opts ...Option) Loader`

Creates a loader with:

- environment prefix (`prefix`)
- built-in base config (`config.Configuration`)
- default `config_dir` of `./config`

### Base Config

```go
type Configuration struct {
    LogConfigOnInit bool   `config:"log_config_on_init"`
    ConfigPath      string `config:"config_dir" flag:"configs,c" validate:"dir"`
}
```

`log_config_on_init=true` logs all loaded settings and can expose sensitive values.

### Options

- `WithArgs(args ...string)`: replace `os.Args` (mainly useful in tests).
- `WithLogger(l *logrus.Logger)`: set loader logger.

## Tag Behavior

- `config:"name"`: binds env + file key + (optional) flag to a field.
- `flag:"name,short,usage"`: registers a pflag and binds it.
- `validate:"..."`: validated using `go-playground/validator` after load.

Built-in custom validation tags included in this package:

- `pub_key`
- `priv_ec_key`

## Example

```go
package main

import "github.com/nojyerac/go-lib/config"

type DBConfig struct {
    Driver string `config:"database_driver" validate:"required"`
    DSN    string `config:"database_connection_string" validate:"required"`
}

func main() {
    cfg := &DBConfig{Driver: "postgres"}

    loader := config.NewConfigLoader("example") // EXAMPLE_* env vars
    if err := loader.RegisterConfig(cfg); err != nil {
        panic(err)
    }
    if err := loader.InitAndValidate(); err != nil {
        panic(err)
    }
}
```

## Notes

- `RegisterConfig` expects `pointer to struct`; any other type returns an error.
- Load + validate runs for every registered struct.
- The loader parses flags during `InitAndValidate()`.
