package config

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Load 从 YAML 文件加载配置，然后用环境变量覆盖。
// 优先级：env > yaml > default tag > 结构体零值
//
// 支持的 struct tag:
//   - yaml:"field_name"  — YAML 字段映射
//   - env:"ENV_VAR"      — 环境变量覆盖
//   - default:"value"    — 默认值（在 YAML 加载之前设置）
//
// cfg 必须是指向结构体的指针。如果 YAML 文件读取失败则返回错误。
func Load(path string, cfg any) error {
	if err := ApplyDefaults(cfg); err != nil {
		return fmt.Errorf("apply defaults: %w", err)
	}
	if err := loadYAML(path, cfg); err != nil {
		return fmt.Errorf("load yaml %s: %w", path, err)
	}
	if err := applyEnv(cfg); err != nil {
		return fmt.Errorf("apply env: %w", err)
	}
	return nil
}

// LoadOrDefault 从 YAML 文件加载配置，文件不存在时使用默认值+环境变量。
// 优先级：env > yaml > default tag > 结构体零值
func LoadOrDefault(path string, cfg any) error {
	if err := ApplyDefaults(cfg); err != nil {
		return fmt.Errorf("apply defaults: %w", err)
	}
	if err := loadYAML(path, cfg); err != nil {
		// YAML 文件不存在时使用默认值 + 环境变量
		if !os.IsNotExist(err) {
			return fmt.Errorf("load yaml %s: %w", path, err)
		}
	}
	if err := applyEnv(cfg); err != nil {
		return fmt.Errorf("apply env: %w", err)
	}
	return nil
}

// loadYAML 从 YAML 文件加载配置到结构体。cfg 必须已被 ApplyDefaults 初始化。
func loadYAML(path string, cfg any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, cfg)
}

// applyEnv 遍历结构体字段，将环境变量值覆盖到字段上。
// 支持的类型：string, int/uint 系列, float 系列, bool, []string（逗号分隔）
//
// 环境变量 key 的确定：
//  1. 字段有 env tag → 直接使用 tag 值
//  2. 字段无 env tag → 自动构造 "PARENT_FIELDNAME" 路径，递归向上拼接
//     例如 Config.Postgres.DSN → 读取 "POSTGRES_DSN" 环境变量
func applyEnv(cfg any) error {
	v := reflect.ValueOf(cfg)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("cfg must be a struct or pointer to struct")
	}
	return applyEnvToStruct(v, "")
}

// prefix 是上级字段名的大写拼接（如 "Postgres"），当前层字段追加在其后。
func applyEnvToStruct(v reflect.Value, prefix string) error {
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldVal := v.Field(i)

		if !fieldVal.CanSet() {
			continue
		}

		// 递归处理嵌套结构体
		if field.Type.Kind() == reflect.Struct && field.Anonymous {
			if err := applyEnvToStruct(fieldVal, prefix); err != nil {
				return err
			}
			continue
		}
		if field.Type.Kind() == reflect.Struct {
			nextPrefix := buildEnvKey(prefix, field.Name)
			if err := applyEnvToStruct(fieldVal, nextPrefix); err != nil {
				return err
			}
			continue
		}

		// 处理指针到结构体的字段
		if field.Type.Kind() == reflect.Ptr && field.Type.Elem().Kind() == reflect.Struct {
			if fieldVal.IsNil() {
				fieldVal.Set(reflect.New(field.Type.Elem()))
			}
			nextPrefix := buildEnvKey(prefix, field.Name)
			if err := applyEnvToStruct(fieldVal.Elem(), nextPrefix); err != nil {
				return err
			}
			continue
		}

		// 标量字段：确定环境变量 key
		envKey := field.Tag.Get("env")
		if envKey == "" {
			envKey = buildEnvKey(prefix, field.Name)
		}

		envVal := os.Getenv(envKey)
		if envVal == "" {
			continue
		}

		if err := setField(fieldVal, envVal); err != nil {
			return fmt.Errorf("set field %s from env %s: %w", field.Name, envKey, err)
		}
	}
	return nil
}

// buildEnvKey 拼接父前缀和当前字段名为大写蛇形环境变量 key。
// prefix="Postgres", name="DSN" → "POSTGRES_DSN"
func buildEnvKey(prefix, name string) string {
	part := toUpperSnake(name)
	if prefix == "" {
		return part
	}
	return prefix + "_" + part
}

// toUpperSnake 将 CamelCase / lowerCamel 转为大写蛇形。
// "DSN" → "DSN", "IdleConns" → "IDLE_CONNS", "accessTTL" → "ACCESS_TTL",
// "TFGen" → "TF_GEN", "PerAdminRPM" → "PER_ADMIN_RPM"
//
// 规则：连续大写字母视为一个词；仅当大写字母前是小写字母、
// 或前为大写字母且后跟小写字母时，才在该大写字母前插入下划线。
func toUpperSnake(s string) string {
	if s == "" {
		return ""
	}
	runes := []rune(s)
	var b strings.Builder
	for i, r := range runes {
		isUpper := r >= 'A' && r <= 'Z'
		if i > 0 && isUpper {
			prev := runes[i-1]
			prevLower := prev >= 'a' && prev <= 'z'
			nextLower := i+1 < len(runes) && runes[i+1] >= 'a' && runes[i+1] <= 'z'
			if prevLower || (prev >= 'A' && prev <= 'Z' && nextLower) {
				b.WriteRune('_')
			}
		}
		b.WriteRune(r)
	}
	return strings.ToUpper(b.String())
}

func setField(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(v)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetUint(v)
	case reflect.Float32, reflect.Float64:
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		field.SetFloat(v)
	case reflect.Bool:
		v, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(v)
	case reflect.Slice:
		if field.Type().Elem().Kind() == reflect.String {
			parts := strings.Split(value, ",")
			trimmed := make([]string, 0, len(parts))
			for _, p := range parts {
				p = strings.TrimSpace(p)
				if p != "" {
					trimmed = append(trimmed, p)
				}
			}
			field.Set(reflect.ValueOf(trimmed))
		}
	default:
		return fmt.Errorf("unsupported field type: %s", field.Kind())
	}
	return nil
}

// Flag 注册一个 --config 命令行 flag。
// 用法: cfgPath := config.Flag("config", "/app/config/config.yaml", "配置文件路径")
func Flag(name, defaultPath, usage string) *string {
	return flag.String(name, defaultPath, usage)
}
