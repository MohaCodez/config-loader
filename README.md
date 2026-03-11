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

## Project Structure

\`\`\`
config-loader/
├── loader/
│   ├── loader.go        # Core loader logic
│   ├── yaml.go          # YAML parsing
│   ├── json.go          # JSON parsing
│   ├── env.go           # Environment variable handling
│   ├── merge.go         # Precedence and merging logic
│   └── watcher.go       # Hot-reload file watcher
├── validator/
│   └── validator.go     # Config validation
├── examples/
│   └── basic/
│       ├── main.go
│       └── config.yaml
├── config.go            # Public API surface
├── go.mod
└── README.md
\`\`\`

## Usage

\`\`\`go
type AppConfig struct {
    Port    int           \`yaml:"port"    env:"APP_PORT"    default:"8080"\`
    Debug   bool          \`yaml:"debug"   env:"APP_DEBUG"   default:"false"\`
    Timeout time.Duration \`yaml:"timeout" env:"APP_TIMEOUT" default:"30s"\`

    Database struct {
        Host     string \`yaml:"host"     env:"DB_HOST"     required:"true"\`
        Port     int    \`yaml:"port"     env:"DB_PORT"     default:"5432"\`
        Password string \`yaml:"password" env:"DB_PASSWORD" required:"true"\`
    } \`yaml:"database"\`
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
\`\`\`

## Precedence Rules

| Source | Priority |
|---|---|
| Environment Variables | Highest |
| JSON file | Medium |
| YAML file | Low |
| Struct tag defaults | Lowest |

## Learning Goals

- **Reflection** — dynamically mapping values onto structs at runtime
- **Struct tags** — reading metadata to drive behavior
- **Type coercion** — converting strings to typed values safely
- **Concurrency** — handling hot-reload without race conditions
- **Error handling** — graceful failures with useful messages

## Dependencies

- [`gopkg.in/yaml.v3`](https://pkg.go.dev/gopkg.in/yaml.v3) — YAML parsing
- [`github.com/fsnotify/fsnotify`](https://pkg.go.dev/github.com/fsnotify/fsnotify) — File watching

## Status

🚧 Work in progress — built for learning purposes.
