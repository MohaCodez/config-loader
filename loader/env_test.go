package loader

import (
	"os"
	"reflect"
	"testing"
	"time"
)

func TestSetFieldValue(t *testing.T) {
	tests := []struct {
		name      string
		fieldType reflect.Type
		value     string
		expected  interface{}
		wantErr   bool
	}{
		{"string", reflect.TypeOf(""), "hello", "hello", false},
		{"int", reflect.TypeOf(0), "42", int64(42), false},
		{"bool true", reflect.TypeOf(false), "true", true, false},
		{"bool 1", reflect.TypeOf(false), "1", true, false},
		{"duration", reflect.TypeOf(time.Duration(0)), "5m", int64(5 * time.Minute), false},
		{"invalid int", reflect.TypeOf(0), "abc", nil, true},
		{"invalid bool", reflect.TypeOf(false), "maybe", nil, true},
		{"invalid duration", reflect.TypeOf(time.Duration(0)), "5x", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := reflect.New(tt.fieldType).Elem()
			err := setFieldValue(field, tt.value)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			var got interface{}
			switch field.Kind() {
			case reflect.String:
				got = field.String()
			case reflect.Int, reflect.Int64:
				got = field.Int()
			case reflect.Bool:
				got = field.Bool()
			}

			if got != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, got)
			}
		})
	}
}

func TestLoadEnv(t *testing.T) {
	type Config struct {
		Host string `env:"TEST_HOST"`
		Port int    `env:"TEST_PORT"`
	}

	os.Setenv("TEST_HOST", "example.com")
	os.Setenv("TEST_PORT", "3000")
	defer func() {
		os.Unsetenv("TEST_HOST")
		os.Unsetenv("TEST_PORT")
	}()

	var cfg Config
	v := reflect.ValueOf(&cfg).Elem()

	if err := LoadEnv(v); err != nil {
		t.Fatalf("LoadEnv failed: %v", err)
	}

	if cfg.Host != "example.com" {
		t.Errorf("expected Host=example.com, got %s", cfg.Host)
	}
	if cfg.Port != 3000 {
		t.Errorf("expected Port=3000, got %d", cfg.Port)
	}
}

func TestLoadEnvNested(t *testing.T) {
	type Config struct {
		Database struct {
			Host string `env:"DB_HOST"`
			Port int    `env:"DB_PORT"`
		}
	}

	os.Setenv("DB_HOST", "db.example.com")
	os.Setenv("DB_PORT", "5432")
	defer func() {
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
	}()

	var cfg Config
	v := reflect.ValueOf(&cfg).Elem()

	if err := LoadEnv(v); err != nil {
		t.Fatalf("LoadEnv failed: %v", err)
	}

	if cfg.Database.Host != "db.example.com" {
		t.Errorf("expected Database.Host=db.example.com, got %s", cfg.Database.Host)
	}
	if cfg.Database.Port != 5432 {
		t.Errorf("expected Database.Port=5432, got %d", cfg.Database.Port)
	}
}
