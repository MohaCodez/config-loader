package loader

import (
	"fmt"
	"reflect"
	"strconv"
	"time"
)

func Merge(v reflect.Value, data map[string]interface{}) error {
	return mergeStruct(v, data)
}

func mergeStruct(v reflect.Value, data map[string]interface{}) error {
	t := v.Type()
	
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)
		
		if !field.CanSet() {
			continue
		}

		yamlKey := fieldType.Tag.Get("yaml")
		if yamlKey == "" {
			yamlKey = fieldType.Name
		}

		value, exists := data[yamlKey]
		if !exists {
			continue
		}

		if field.Kind() == reflect.Struct {
			if nestedMap, ok := value.(map[string]interface{}); ok {
				if err := mergeStruct(field, nestedMap); err != nil {
					return err
				}
			}
			continue
		}

		if err := setFieldFromInterface(field, value); err != nil {
			return fmt.Errorf("field %s: %w", fieldType.Name, err)
		}
	}
	
	return nil
}

func setFieldFromInterface(field reflect.Value, value interface{}) error {
	if value == nil {
		return nil
	}

	switch field.Kind() {
	case reflect.String:
		if s, ok := value.(string); ok {
			field.SetString(s)
		} else {
			field.SetString(fmt.Sprint(value))
		}
		
	case reflect.Int, reflect.Int64:
		if field.Type() == reflect.TypeOf(time.Duration(0)) {
			switch v := value.(type) {
			case string:
				d, err := time.ParseDuration(v)
				if err != nil {
					return fmt.Errorf("invalid duration %q: %w", v, err)
				}
				field.SetInt(int64(d))
			case int, int64, float64:
				field.SetInt(toInt64(v))
			default:
				return fmt.Errorf("cannot convert %T to duration", value)
			}
		} else {
			switch v := value.(type) {
			case int:
				field.SetInt(int64(v))
			case int64:
				field.SetInt(v)
			case float64:
				field.SetInt(int64(v))
			case string:
				i, err := strconv.ParseInt(v, 10, 64)
				if err != nil {
					return fmt.Errorf("invalid int %q: %w", v, err)
				}
				field.SetInt(i)
			default:
				return fmt.Errorf("cannot convert %T to int", value)
			}
		}
		
	case reflect.Float64:
		switch v := value.(type) {
		case float64:
			field.SetFloat(v)
		case int:
			field.SetFloat(float64(v))
		case int64:
			field.SetFloat(float64(v))
		case string:
			f, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return fmt.Errorf("invalid float %q: %w", v, err)
			}
			field.SetFloat(f)
		default:
			return fmt.Errorf("cannot convert %T to float", value)
		}
		
	case reflect.Bool:
		switch v := value.(type) {
		case bool:
			field.SetBool(v)
		case string:
			b, err := strconv.ParseBool(v)
			if err != nil {
				return fmt.Errorf("invalid bool %q: %w", v, err)
			}
			field.SetBool(b)
		default:
			return fmt.Errorf("cannot convert %T to bool", value)
		}
		
	default:
		return fmt.Errorf("unsupported type %s", field.Kind())
	}
	
	return nil
}

func toInt64(v interface{}) int64 {
	switch val := v.(type) {
	case int:
		return int64(val)
	case int64:
		return val
	case float64:
		return int64(val)
	default:
		return 0
	}
}
