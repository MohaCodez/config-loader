package loader

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"time"
)

func LoadEnv(v reflect.Value) error {
	return loadEnvStruct(v, "")
}

func loadEnvStruct(v reflect.Value, prefix string) error {
	t := v.Type()
	
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)
		
		if !field.CanSet() {
			continue
		}

		if field.Kind() == reflect.Struct {
			if err := loadEnvStruct(field, prefix); err != nil {
				return err
			}
			continue
		}

		envKey := fieldType.Tag.Get("env")
		if envKey == "" {
			continue
		}

		envVal := os.Getenv(envKey)
		if envVal == "" {
			continue
		}

		if err := setFieldValue(field, envVal); err != nil {
			return fmt.Errorf("field %s (env %s): %w", fieldType.Name, envKey, err)
		}
	}
	
	return nil
}

func setFieldValue(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
		
	case reflect.Int, reflect.Int64:
		if field.Type() == reflect.TypeOf(time.Duration(0)) {
			d, err := time.ParseDuration(value)
			if err != nil {
				return fmt.Errorf("invalid duration %q: %w", value, err)
			}
			field.SetInt(int64(d))
		} else {
			i, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid int %q: %w", value, err)
			}
			field.SetInt(i)
		}
		
	case reflect.Float64:
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid float %q: %w", value, err)
		}
		field.SetFloat(f)
		
	case reflect.Bool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid bool %q: %w", value, err)
		}
		field.SetBool(b)
		
	default:
		return fmt.Errorf("unsupported type %s", field.Kind())
	}
	
	return nil
}
