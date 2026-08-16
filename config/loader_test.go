package config

import (
	"os"
	"path/filepath"
	"testing"
)

type testConfig struct {
	Name   string  `yaml:"name" env:"TEST_NAME"`
	Port   int     `yaml:"port" env:"TEST_PORT"`
	Debug  bool    `yaml:"debug" env:"TEST_DEBUG"`
	Limit  int64   `yaml:"limit" env:"TEST_LIMIT"`
	Weight float64 `yaml:"weight" env:"TEST_WEIGHT"`
}

func TestLoadYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yaml")
	os.WriteFile(path, []byte("name: hello\nport: 8080\ndebug: true\nlimit: 100"), 0644)

	var cfg testConfig
	if err := loadYAML(path, &cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Name != "hello" || cfg.Port != 8080 || !cfg.Debug || cfg.Limit != 100 {
		t.Errorf("unexpected config: %+v", cfg)
	}
}

func TestLoadYAMLThenEnvOverrides(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yaml")
	os.WriteFile(path, []byte("name: yaml_name\nport: 3000\ndebug: false"), 0644)

	os.Setenv("TEST_NAME", "env_name")
	os.Setenv("TEST_PORT", "9999")
	defer os.Unsetenv("TEST_NAME")
	defer os.Unsetenv("TEST_PORT")

	var cfg testConfig
	if err := Load(path, &cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Name != "env_name" {
		t.Errorf("expected env override for Name, got %q", cfg.Name)
	}
	if cfg.Port != 9999 {
		t.Errorf("expected env override for Port, got %d", cfg.Port)
	}
	if cfg.Debug != false {
		t.Errorf("expected yaml value for Debug, got %v", cfg.Debug)
	}
}

func TestApplyEnvSlice(t *testing.T) {
	type sliceConfig struct {
		Origins []string `yaml:"origins" env:"TEST_ORIGINS"`
	}
	os.Setenv("TEST_ORIGINS", "a,b,c")
	defer os.Unsetenv("TEST_ORIGINS")

	var cfg sliceConfig
	if err := applyEnv(&cfg); err != nil {
		t.Fatal(err)
	}
	if len(cfg.Origins) != 3 || cfg.Origins[0] != "a" || cfg.Origins[2] != "c" {
		t.Errorf("unexpected origins: %v", cfg.Origins)
	}
}

func TestApplyEnvFloat(t *testing.T) {
	type fConfig struct {
		Weight float64 `yaml:"weight" env:"TEST_WEIGHT_NEW"`
	}
	os.Setenv("TEST_WEIGHT_NEW", "3.14")
	defer os.Unsetenv("TEST_WEIGHT_NEW")

	var cfg fConfig
	if err := applyEnv(&cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Weight != 3.14 {
		t.Errorf("expected 3.14, got %f", cfg.Weight)
	}
}

func TestLoadFileNotFound(t *testing.T) {
	var cfg testConfig
	err := Load("/nonexistent/path.yaml", &cfg)
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestToUpperSnake(t *testing.T) {
	cases := map[string]string{
		"":                "",
		"DSN":             "DSN",
		"IdleConns":       "IDLE_CONNS",
		"accessTTL":       "ACCESS_TTL",
		"TFGen":           "TF_GEN",
		"PerAdminRPM":     "PER_ADMIN_RPM",
		"GlobalRPM":       "GLOBAL_RPM",
		"ConnectionTTL":   "CONNECTION_TTL",
		"GRPC":            "GRPC",
		"HTTPAddr":        "HTTP_ADDR",
		"MaxLifetime":     "MAX_LIFETIME",
		"FinalizeDelay":   "FINALIZE_DELAY",
		"GapScanner":      "GAP_SCANNER",
		"ScanInterval":    "SCAN_INTERVAL",
		"startTime":       "START_TIME",
		"restartInterval": "RESTART_INTERVAL",
	}
	for in, want := range cases {
		if got := toUpperSnake(in); got != want {
			t.Errorf("toUpperSnake(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestBuildEnvKey(t *testing.T) {
	cases := map[string]string{
		"DSN":                    "DSN",
		"IdleConns":              "IDLE_CONNS",
		"POSTGRES|DSN":           "POSTGRES_DSN",
		"POSTGRES|IdleConns":     "POSTGRES_IDLE_CONNS",
		"RATE_LIMIT|PerAdminRPM": "RATE_LIMIT_PER_ADMIN_RPM",
		"GAP_SCANNER|Lookback":   "GAP_SCANNER_LOOKBACK",
	}
	for k, want := range cases {
		parts := splitV(k)
		got := buildEnvKey(parts[0], parts[1])
		if got != want {
			t.Errorf("buildEnvKey(%q,%q) = %q, want %q", parts[0], parts[1], got, want)
		}
	}
}

func splitV(k string) []string {
	idx := -1
	for i, r := range k {
		if r == '|' {
			idx = i
			break
		}
	}
	if idx < 0 {
		return []string{"", k}
	}
	return []string{k[:idx], k[idx+1:]}
}
