package runtimecheck

import (
	"strings"
	"testing"
	"time"
)

func TestRedactedConfigHidesPostgresPassword(t *testing.T) {
	cfg := Config{
		PostgresURL:   "postgres://user:secret@localhost:5432/db?sslmode=disable",
		DragonflyAddr: "localhost:6379",
		TemporalAddr:  "localhost:7233",
		Timeout:       time.Second,
	}
	got := RedactedConfig(cfg)
	if strings.Contains(got.PostgresURL, "secret") {
		t.Fatalf("PostgresURL leaked password: %s", got.PostgresURL)
	}
	if !strings.Contains(got.PostgresURL, "***@localhost") {
		t.Fatalf("PostgresURL not redacted as expected: %s", got.PostgresURL)
	}
}
