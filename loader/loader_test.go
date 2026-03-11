package loader

import (
	"os"
	"testing"
	"time"
)

type TestConfig struct {
	Port    int           `yaml:"port"    env:"TEST_PORT"    default:"8080"`
	Debug   bool          `yaml:"debug"   env:"TEST_DEBUG"   default:"false"`
	Timeout time.Duration `yaml:"timeout" env:"TEST_TIMEOUT" default:"30s"`
	Name    string        `yaml:"name"    env:"TEST_NAME"    required:"true"`
}

func TestApplyDefaults(t *testing.T) {
	loader := New()
	var cfg TestConfig
	cfg.Name = "test" // satisfy required

	if err := loader.Load(&cfg); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Port != 8080 {
		t.Errorf("expected Port=8080, got %d", cfg.Port)
	}
	if cfg.Debug != false {
		t.Errorf("expected Debug=false, got %v", cfg.Debug)
	}
	if cfg.Timeout != 30*time.Second {
		t.Errorf("expected Timeout=30s, got %v", cfg.Timeout)
	}
}

func TestEnvOverride(t *testing.T) {
	os.Setenv("TEST_PORT", "9000")
	os.Setenv("TEST_DEBUG", "true")
	os.Setenv("TEST_NAME", "myapp")
	defer func() {
		os.Unsetenv("TEST_PORT")
		os.Unsetenv("TEST_DEBUG")
		os.Unsetenv("TEST_NAME")
	}()

	loader := New().WithEnv()
	var cfg TestConfig

	if err := loader.Load(&cfg); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Port != 9000 {
		t.Errorf("expected Port=9000, got %d", cfg.Port)
	}
	if cfg.Debug != true {
		t.Errorf("expected Debug=true, got %v", cfg.Debug)
	}
	if cfg.Name != "myapp" {
		t.Errorf("expected Name=myapp, got %s", cfg.Name)
	}
}

func TestValidationRequired(t *testing.T) {
	loader := New()
	var cfg TestConfig

	err := loader.Load(&cfg)
	if err == nil {
		t.Fatal("expected validation error for missing required field")
	}
}

func TestValidationPortRange(t *testing.T) {
	type BadConfig struct {
		Port int    `yaml:"port" default:"99999"`
		Name string `yaml:"name" required:"true"`
	}

	loader := New()
	var cfg BadConfig
	cfg.Name = "test"

	err := loader.Load(&cfg)
	if err == nil {
		t.Fatal("expected validation error for invalid port range")
	}
}

func TestYAMLLoading(t *testing.T) {
	// Create temp YAML file
	content := []byte("port: 3000\nname: testapp\n")
	tmpfile, err := os.CreateTemp("", "config*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write(content); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	loader := New().WithYAML(tmpfile.Name())
	var cfg TestConfig

	if err := loader.Load(&cfg); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Port != 3000 {
		t.Errorf("expected Port=3000, got %d", cfg.Port)
	}
	if cfg.Name != "testapp" {
		t.Errorf("expected Name=testapp, got %s", cfg.Name)
	}
}

func TestPrecedence(t *testing.T) {
	// Create temp YAML file
	content := []byte("port: 3000\nname: yaml-name\n")
	tmpfile, err := os.CreateTemp("", "config*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write(content); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	// Set env var (should override YAML)
	os.Setenv("TEST_PORT", "4000")
	defer os.Unsetenv("TEST_PORT")

	loader := New().WithYAML(tmpfile.Name()).WithEnv()
	var cfg TestConfig

	if err := loader.Load(&cfg); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Env should override YAML
	if cfg.Port != 4000 {
		t.Errorf("expected Port=4000 (from env), got %d", cfg.Port)
	}
	// YAML value should be used for name
	if cfg.Name != "yaml-name" {
		t.Errorf("expected Name=yaml-name, got %s", cfg.Name)
	}
}

func TestHotReloadErrorPreservesOldConfig(t *testing.T) {
	content := []byte("port: 3000\nname: testapp\n")
	tmpfile, err := os.CreateTemp("", "config*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write(content); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	var reloadErrCalled bool
	loader := New().
		WithYAML(tmpfile.Name()).
		WithHotReload().
		OnReloadError(func(err error) {
			reloadErrCalled = true
		})

	var cfg TestConfig
	if err := loader.Load(&cfg); err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	defer loader.Close()

	if cfg.Port != 3000 || cfg.Name != "testapp" {
		t.Fatalf("initial config incorrect: Port=%d, Name=%s", cfg.Port, cfg.Name)
	}

	// Write invalid YAML to trigger reload error
	if err := os.WriteFile(tmpfile.Name(), []byte("port: invalid\nname: newname\n"), 0644); err != nil {
		t.Fatal(err)
	}

	time.Sleep(500 * time.Millisecond)

	// Old config should be preserved
	if cfg.Port != 3000 {
		t.Errorf("expected Port=3000 (old value), got %d", cfg.Port)
	}
	if cfg.Name != "testapp" {
		t.Errorf("expected Name=testapp (old value), got %s", cfg.Name)
	}
	if !reloadErrCalled {
		t.Error("expected OnReloadError callback to be called")
	}
}
