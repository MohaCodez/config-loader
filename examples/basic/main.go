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
		WithHotReload().
		OnReload(func(c interface{}) {
			fmt.Println("\n🔄 Config reloaded!")
			printConfig(c.(*AppConfig))
		})

	if err := loader.Load(&cfg); err != nil {
		log.Fatal(err)
	}

	fmt.Println("✅ Config loaded successfully!")
	printConfig(&cfg)

	fmt.Println("\n👀 Watching for config changes... (Press Ctrl+C to exit)")
	select {}
}

func printConfig(cfg *AppConfig) {
	fmt.Printf("Server:\n")
	fmt.Printf("  Port: %d\n", cfg.Port)
	fmt.Printf("  Debug: %v\n", cfg.Debug)
	fmt.Printf("  Timeout: %v\n", cfg.Timeout)
	fmt.Printf("Database:\n")
	fmt.Printf("  Host: %s\n", cfg.Database.Host)
	fmt.Printf("  Port: %d\n", cfg.Database.Port)
	fmt.Printf("  Password: %s\n", maskPassword(cfg.Database.Password))
}

func maskPassword(pwd string) string {
	if len(pwd) <= 4 {
		return "****"
	}
	return pwd[:2] + "****" + pwd[len(pwd)-2:]
}
