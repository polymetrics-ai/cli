package runtimecheck

import (
	"os"
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
	if strings.Contains(got.PostgresURL, "sslmode") {
		t.Fatalf("PostgresURL leaked query string: %s", got.PostgresURL)
	}
	if !strings.Contains(got.PostgresURL, "***@localhost") {
		t.Fatalf("PostgresURL not redacted as expected: %s", got.PostgresURL)
	}
}

func TestRedactedConfigScrubsPostgresQuerySecretsAndMalformedDSNs(t *testing.T) {
	cases := []struct {
		name string
		raw  string
	}{
		{name: "query token", raw: "postgres://localhost:5432/db?password=querysecret&sslmode=disable"},
		{name: "fragment token", raw: "postgres://localhost:5432/db#fragsecret"},
		{name: "malformed with query", raw: "postgres://user:badpass@%zz/db?token=querysecret"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := RedactedConfig(Config{PostgresURL: tc.raw}).PostgresURL
			for _, forbidden := range []string{"querysecret", "fragsecret", "badpass", "password=", "token=", "%zz"} {
				if strings.Contains(got, forbidden) {
					t.Fatalf("PostgresURL leaked %q in %q", forbidden, got)
				}
			}
		})
	}
}

func TestRuntimeCheckTemporalDialUsesDialContext(t *testing.T) {
	src, err := os.ReadFile("runtimecheck.go")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(src), "client.DialContext") {
		t.Fatal("runtimecheck temporal dial must use client.DialContext so caller cancellation bounds dial setup")
	}
}
