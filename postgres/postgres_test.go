package postgres

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestGormLogsNeverExpandSensitiveParameters(t *testing.T) {
	dsn := strings.TrimSpace(os.Getenv("TEST_POSTGRES_DSN"))
	if dsn == "" {
		t.Skip("TEST_POSTGRES_DSN is not configured")
	}

	var logs bytes.Buffer
	database, err := newGormDB(&Config{
		DSN: dsn, IdleConns: 1, MaxConns: 1, MaxLifetime: 60, CreateBatchSize: 10,
	}, &logs)
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := database.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	sensitive := []string{
		"SIGNED_BEARER_SENTINEL_7b3c",
		"CIPHERTEXT_SENTINEL_v1_abcd",
		"KEK_SENTINEL_32_BYTES_MATERIAL",
	}
	var result struct {
		Combined string
	}
	if err := database.Raw(
		`SELECT (?::text || ?::text || ?::text) AS combined FROM (SELECT pg_sleep(0.25)) AS delayed`,
		sensitive[0], sensitive[1], sensitive[2],
	).Scan(&result).Error; err != nil {
		t.Fatal(err)
	}
	if result.Combined != strings.Join(sensitive, "") {
		t.Fatalf("parameterized query returned an unexpected value")
	}

	visible := logs.String()
	if !strings.Contains(visible, "SELECT") {
		t.Fatalf("slow query did not produce the expected warning log")
	}
	redacted := visible
	for _, secret := range sensitive {
		redacted = strings.ReplaceAll(redacted, secret, "[REDACTED]")
	}
	for _, secret := range sensitive {
		if strings.Contains(visible, secret) {
			t.Fatalf("SQL log disclosed sensitive parameter %q: %s", secret, redacted)
		}
	}
}
