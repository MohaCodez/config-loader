package loader

import (
	"reflect"
	"testing"
	"time"
)

func TestMergeSimple(t *testing.T) {
	type Config struct {
		Port int    `yaml:"port"`
		Name string `yaml:"name"`
	}

	var cfg Config
	v := reflect.ValueOf(&cfg).Elem()

	data := map[string]interface{}{
		"port": 8080,
		"name": "myapp",
	}

	if err := Merge(v, data); err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	if cfg.Port != 8080 {
		t.Errorf("expected Port=8080, got %d", cfg.Port)
	}
	if cfg.Name != "myapp" {
		t.Errorf("expected Name=myapp, got %s", cfg.Name)
	}
}

func TestMergeNested(t *testing.T) {
	type Config struct {
		Database struct {
			Host string `yaml:"host"`
			Port int    `yaml:"port"`
		} `yaml:"database"`
	}

	var cfg Config
	v := reflect.ValueOf(&cfg).Elem()

	data := map[string]interface{}{
		"database": map[string]interface{}{
			"host": "localhost",
			"port": 5432,
		},
	}

	if err := Merge(v, data); err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	if cfg.Database.Host != "localhost" {
		t.Errorf("expected Database.Host=localhost, got %s", cfg.Database.Host)
	}
	if cfg.Database.Port != 5432 {
		t.Errorf("expected Database.Port=5432, got %d", cfg.Database.Port)
	}
}

func TestMergeDuration(t *testing.T) {
	type Config struct {
		Timeout time.Duration `yaml:"timeout"`
	}

	var cfg Config
	v := reflect.ValueOf(&cfg).Elem()

	data := map[string]interface{}{
		"timeout": "5m",
	}

	if err := Merge(v, data); err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	if cfg.Timeout != 5*time.Minute {
		t.Errorf("expected Timeout=5m, got %v", cfg.Timeout)
	}
}

func TestMergeTypeCoercion(t *testing.T) {
	type Config struct {
		Port  int     `yaml:"port"`
		Rate  float64 `yaml:"rate"`
		Debug bool    `yaml:"debug"`
	}

	tests := []struct {
		name string
		data map[string]interface{}
		want Config
	}{
		{
			name: "int from float",
			data: map[string]interface{}{"port": 8080.0},
			want: Config{Port: 8080},
		},
		{
			name: "float from int",
			data: map[string]interface{}{"rate": 100},
			want: Config{Rate: 100.0},
		},
		{
			name: "bool from string",
			data: map[string]interface{}{"debug": "true"},
			want: Config{Debug: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cfg Config
			v := reflect.ValueOf(&cfg).Elem()

			if err := Merge(v, tt.data); err != nil {
				t.Fatalf("Merge failed: %v", err)
			}

			if cfg.Port != tt.want.Port {
				t.Errorf("Port: expected %d, got %d", tt.want.Port, cfg.Port)
			}
			if cfg.Rate != tt.want.Rate {
				t.Errorf("Rate: expected %f, got %f", tt.want.Rate, cfg.Rate)
			}
			if cfg.Debug != tt.want.Debug {
				t.Errorf("Debug: expected %v, got %v", tt.want.Debug, cfg.Debug)
			}
		})
	}
}
