# config-loader

A Go library for loading application configuration from multiple sources — YAML files, JSON files, and environment variables — with defined precedence rules and hot-reload support.

Built as a learning project to understand reflection, struct tags, and environment handling in Go.

## Features

- Load config from YAML and JSON files
- Override any value via environment variables
- Precedence: env vars > JSON > YAML > defaults
- Hot-reload: automatically detects file changes and reloads config
- Struct tags for mapping and default values
- Type coercion from string env vars to int, bool, duration, etc.
- Validation for required fields and value ranges

## Installation

```bash
go get github.com/amit/config-loader
```

## Usage

```go
package main

import (
    "fmt"
    "log"
    "time"
    
    config "github.com/amit/config-loader"
)

type AppConfig struct {
    Port    int           `yaml:"port"    env:"APP_PORT"    default:"8080"`
    Debug   bool          `yaml:"debug"   env:"APP_DEBUG"   default:"false"`
    Timeout time.Duration `yaml:"timeout" env:"APP_TIMEOUT" default:"30s"`

    Database struct {
        Host     string `yaml:"host"     env:"DB_HOST"     required:"true"`
        Port     int    `yaml:"port"     env:"DB_PORT"     default:"5432"`
        Password string `yaml:"password" env:"DB_PASSWORD" required:"true"`
    } `yaml:"database"`
}

func main() {
    var cfg AppConfig

    loader := config.New().
        WithYAML("config.yaml").
        WithJSON("config.json").
        WithEnv().
        WithHotReload()

    if err := loader.Load(&cfg); err != nil {
        log.Fatal(err)
    }

    fmt.Println("Server running on port:", cfg.Port)
}
```

## Struct Tags

- `yaml:"key"` - YAML key name
- `env:"ENV_VAR"` - Environment variable name
- `default:"value"` - Default value if not set
- `required:"true"` - Field must be set

## Precedence Rules

Configuration sources are merged in this order (highest to lowest priority):

| Source | Priority |
|---|---|
| Environment Variables | Highest |
| JSON file | Medium |
| YAML file | Low |
| Struct tag defaults | Lowest |

## Hot-Reload

Enable hot-reload to automatically reload configuration when files change:

```go
loader := config.New().
    WithYAML("config.yaml").
    WithHotReload().
    OnReload(func(cfg interface{}) {
        fmt.Println("Config reloaded!")
    })
```

Changes are debounced (300ms) and reloading is goroutine-safe. If the new config is invalid, the old config is kept.

## Validation

The library validates:
- Required fields are set
- Port numbers are between 1-65535
- Duration values are non-negative

## Type Coercion

Environment variables (strings) are automatically converted to:
- `int`, `int64`
- `float64`
- `bool` (accepts "true", "false", "1", "0")
- `time.Duration` (e.g., "30s", "5m", "1h")

## Example

See `examples/basic/` for a complete working example.

```bash
cd examples/basic
go run main.go
```

## Running Tests

```bash
go test ./...
```

## Dependencies

- [`gopkg.in/yaml.v3`](https://pkg.go.dev/gopkg.in/yaml.v3) — YAML parsing
- [`github.com/fsnotify/fsnotify`](https://pkg.go.dev/github.com/fsnotify/fsnotify) — File watching

## License

MIT
