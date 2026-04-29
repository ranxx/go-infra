package config

import (
	"testing"

	"github.com/ranxx/go-infra/mysql"
)

type stubProvider map[string]string

func (s stubProvider) GetValue(key string) string {
	return s[key]
}

func TestApplyDefaults_MySQLConfig(t *testing.T) {
	cfg := mysql.Config{}

	if err := ApplyDefaults(&cfg); err != nil {
		t.Fatalf("ApplyDefaults() error = %v", err)
	}

	if cfg.IdleConns != 10 {
		t.Fatalf("IdleConns = %d, want 10", cfg.IdleConns)
	}
	if cfg.MaxConns != 100 {
		t.Fatalf("MaxConns = %d, want 100", cfg.MaxConns)
	}
	if cfg.MaxLifetime != 3600 {
		t.Fatalf("MaxLifetime = %d, want 3600", cfg.MaxLifetime)
	}
	if cfg.CreateBatchSize != 1000 {
		t.Fatalf("CreateBatchSize = %d, want 1000", cfg.CreateBatchSize)
	}
}

func TestApplyDefaults_DoesNotOverrideNonZero(t *testing.T) {
	cfg := mysql.Config{
		IdleConns: 3,
	}

	if err := ApplyDefaults(&cfg); err != nil {
		t.Fatalf("ApplyDefaults() error = %v", err)
	}

	if cfg.IdleConns != 3 {
		t.Fatalf("IdleConns = %d, want 3", cfg.IdleConns)
	}
	if cfg.MaxConns != 100 {
		t.Fatalf("MaxConns = %d, want 100", cfg.MaxConns)
	}
	if cfg.CreateBatchSize != 1000 {
		t.Fatalf("CreateBatchSize = %d, want 1000", cfg.CreateBatchSize)
	}
}

func TestLoadByKey_ApplyDefaultsAndOverrideByPayload(t *testing.T) {
	cfg := mysql.Config{}
	provider := stubProvider{
		"mysql": `{"DSN":"root:pwd@tcp(localhost:3306)/db","MaxConns":200}`,
	}

	if err := LoadByKey("mysql", &cfg, provider); err != nil {
		t.Fatalf("LoadByKey() error = %v", err)
	}

	if cfg.DSN != "root:pwd@tcp(localhost:3306)/db" {
		t.Fatalf("DSN = %q, want expected dsn", cfg.DSN)
	}
	if cfg.MaxConns != 200 {
		t.Fatalf("MaxConns = %d, want 200", cfg.MaxConns)
	}
	if cfg.IdleConns != 10 {
		t.Fatalf("IdleConns = %d, want 10", cfg.IdleConns)
	}
	if cfg.CreateBatchSize != 1000 {
		t.Fatalf("CreateBatchSize = %d, want 1000", cfg.CreateBatchSize)
	}
}

func TestLoadByKey_ApplyDefaultsWithoutProviders(t *testing.T) {
	cfg := mysql.Config{}

	if err := LoadByKey("mysql", &cfg); err != nil {
		t.Fatalf("LoadByKey() error = %v", err)
	}

	if cfg.MaxConns != 100 {
		t.Fatalf("MaxConns = %d, want 100", cfg.MaxConns)
	}
}
