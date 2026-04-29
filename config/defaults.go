package config

import (
	"fmt"
	"reflect"
	"strconv"
	"time"
)

// ApplyDefaults 根据字段上的 default tag 为零值字段填充默认值。
// ptr 需要是非 nil 的指针，且通常指向结构体。
func ApplyDefaults(ptr interface{}) error {
	if ptr == nil {
		return fmt.Errorf("target is nil")
	}

	v := reflect.ValueOf(ptr)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return fmt.Errorf("target must be a non-nil pointer")
	}

	elem := v.Elem()
	if elem.Kind() != reflect.Struct {
		return nil
	}

	return applyStructDefaults(elem, elem.Type().Name())
}

func applyStructDefaults(v reflect.Value, parentPath string) error {
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		// 跳过非导出字段。
		if field.PkgPath != "" && !field.Anonymous {
			continue
		}

		fieldPath := field.Name
		if parentPath != "" {
			fieldPath = parentPath + "." + field.Name
		}

		if tag, ok := field.Tag.Lookup("default"); ok && tag != "-" {
			if err := setFieldDefault(fieldValue, field.Type, tag, fieldPath); err != nil {
				return err
			}
		}

		if err := applyNestedDefaults(fieldValue, fieldPath); err != nil {
			return err
		}
	}

	return nil
}

func applyNestedDefaults(v reflect.Value, fieldPath string) error {
	switch v.Kind() {
	case reflect.Struct:
		return applyStructDefaults(v, fieldPath)
	case reflect.Ptr:
		if v.IsNil() || v.Elem().Kind() != reflect.Struct {
			return nil
		}
		return applyStructDefaults(v.Elem(), fieldPath)
	default:
		return nil
	}
}

func setFieldDefault(v reflect.Value, fieldType reflect.Type, raw, fieldPath string) error {
	if !v.CanSet() {
		return nil
	}

	if v.Kind() == reflect.Ptr {
		if !v.IsNil() {
			return nil
		}

		elemType := v.Type().Elem()
		elemValue := reflect.New(elemType).Elem()
		if err := setScalarValue(elemValue, elemType, raw); err != nil {
			return fmt.Errorf("field %s: %w", fieldPath, err)
		}

		ptrValue := reflect.New(elemType)
		ptrValue.Elem().Set(elemValue)
		v.Set(ptrValue)
		return nil
	}

	if !v.IsZero() {
		return nil
	}

	if err := setScalarValue(v, fieldType, raw); err != nil {
		return fmt.Errorf("field %s: %w", fieldPath, err)
	}

	return nil
}

func setScalarValue(v reflect.Value, valueType reflect.Type, raw string) error {
	switch v.Kind() {
	case reflect.String:
		v.SetString(raw)
		return nil
	case reflect.Bool:
		parsed, err := strconv.ParseBool(raw)
		if err != nil {
			return fmt.Errorf("invalid bool value %q: %w", raw, err)
		}
		v.SetBool(parsed)
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if valueType.PkgPath() == "time" && valueType.Name() == "Duration" {
			d, err := time.ParseDuration(raw)
			if err != nil {
				return fmt.Errorf("invalid duration value %q: %w", raw, err)
			}
			v.SetInt(int64(d))
			return nil
		}

		parsed, err := strconv.ParseInt(raw, 10, v.Type().Bits())
		if err != nil {
			return fmt.Errorf("invalid integer value %q: %w", raw, err)
		}
		v.SetInt(parsed)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		parsed, err := strconv.ParseUint(raw, 10, v.Type().Bits())
		if err != nil {
			return fmt.Errorf("invalid unsigned integer value %q: %w", raw, err)
		}
		v.SetUint(parsed)
		return nil
	case reflect.Float32, reflect.Float64:
		parsed, err := strconv.ParseFloat(raw, v.Type().Bits())
		if err != nil {
			return fmt.Errorf("invalid float value %q: %w", raw, err)
		}
		v.SetFloat(parsed)
		return nil
	default:
		return fmt.Errorf("unsupported kind %s", v.Kind())
	}
}
