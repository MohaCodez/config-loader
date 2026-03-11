package loader

import (
	"fmt"
	"reflect"
	"sync"
)

type Loader struct {
	yamlPath   string
	jsonPath   string
	useEnv     bool
	hotReload  bool
	onReload   func(interface{})
	watcher    *Watcher
	mu         sync.RWMutex
	configPtr  interface{}
}

func New() *Loader {
	return &Loader{}
}

func (l *Loader) WithYAML(path string) *Loader {
	l.yamlPath = path
	return l
}

func (l *Loader) WithJSON(path string) *Loader {
	l.jsonPath = path
	return l
}

func (l *Loader) WithEnv() *Loader {
	l.useEnv = true
	return l
}

func (l *Loader) WithHotReload() *Loader {
	l.hotReload = true
	return l
}

func (l *Loader) OnReload(callback func(interface{})) *Loader {
	l.onReload = callback
	return l
}

func (l *Loader) Load(configPtr interface{}) error {
	if reflect.TypeOf(configPtr).Kind() != reflect.Ptr {
		return fmt.Errorf("config must be a pointer to struct")
	}

	l.mu.Lock()
	l.configPtr = configPtr
	l.mu.Unlock()

	if err := l.load(); err != nil {
		return err
	}

	if l.hotReload {
		var paths []string
		if l.yamlPath != "" {
			paths = append(paths, l.yamlPath)
		}
		if l.jsonPath != "" {
			paths = append(paths, l.jsonPath)
		}
		
		if len(paths) > 0 {
			watcher, err := NewWatcher(paths, func() {
				l.mu.Lock()
				defer l.mu.Unlock()
				
				if err := l.load(); err != nil {
					fmt.Printf("hot-reload failed: %v\n", err)
					return
				}
				
				if l.onReload != nil {
					l.onReload(l.configPtr)
				}
			})
			if err != nil {
				return fmt.Errorf("failed to start watcher: %w", err)
			}
			l.watcher = watcher
		}
	}

	return nil
}

func (l *Loader) load() error {
	v := reflect.ValueOf(l.configPtr).Elem()
	
	// Apply defaults
	if err := applyDefaults(v); err != nil {
		return fmt.Errorf("applying defaults: %w", err)
	}

	// Load YAML
	if l.yamlPath != "" {
		data, err := LoadYAML(l.yamlPath)
		if err != nil {
			return fmt.Errorf("loading YAML: %w", err)
		}
		if err := Merge(v, data); err != nil {
			return fmt.Errorf("merging YAML: %w", err)
		}
	}

	// Load JSON
	if l.jsonPath != "" {
		data, err := LoadJSON(l.jsonPath)
		if err != nil {
			return fmt.Errorf("loading JSON: %w", err)
		}
		if err := Merge(v, data); err != nil {
			return fmt.Errorf("merging JSON: %w", err)
		}
	}

	// Load env vars
	if l.useEnv {
		if err := LoadEnv(v); err != nil {
			return fmt.Errorf("loading env vars: %w", err)
		}
	}

	// Validate
	if err := Validate(v); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	return nil
}

func (l *Loader) Close() error {
	if l.watcher != nil {
		return l.watcher.Close()
	}
	return nil
}

func applyDefaults(v reflect.Value) error {
	t := v.Type()
	
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)
		
		if !field.CanSet() {
			continue
		}

		if field.Kind() == reflect.Struct {
			if err := applyDefaults(field); err != nil {
				return err
			}
			continue
		}

		defaultVal := fieldType.Tag.Get("default")
		if defaultVal != "" {
			if err := setFieldValue(field, defaultVal); err != nil {
				return fmt.Errorf("field %s: %w", fieldType.Name, err)
			}
		}
	}
	
	return nil
}

func Validate(v reflect.Value) error {
	return validateStruct(v, "")
}

func validateStruct(v reflect.Value, prefix string) error {
	t := v.Type()
	var errors []string
	
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)
		
		fieldName := fieldType.Name
		if prefix != "" {
			fieldName = prefix + "." + fieldName
		}

		if field.Kind() == reflect.Struct {
			if err := validateStruct(field, fieldName); err != nil {
				errors = append(errors, err.Error())
			}
			continue
		}

		// Check required
		if fieldType.Tag.Get("required") == "true" {
			if isZero(field) {
				errors = append(errors, fmt.Sprintf("field %q: required but not set", fieldName))
			}
		}

		// Validate port ranges
		if field.Kind() == reflect.Int && (fieldType.Name == "Port" || fieldType.Tag.Get("yaml") == "port") {
			port := field.Int()
			if port != 0 && (port < 1 || port > 65535) {
				errors = append(errors, fmt.Sprintf("field %q: port must be between 1 and 65535, got %d", fieldName, port))
			}
		}

		// Validate duration > 0
		if field.Type().String() == "time.Duration" {
			if field.Int() < 0 {
				errors = append(errors, fmt.Sprintf("field %q: duration must be greater than zero", fieldName))
			}
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("%v", errors)
	}
	
	return nil
}

func isZero(v reflect.Value) bool {
	return v.IsZero()
}
