package runtimecheck

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"syscall"
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

func TestRedactedConfigSanitizesAllRuntimeEndpoints(t *testing.T) {
	marker := "pm-review-fix-endpoint-secret"
	cfg := Config{
		PostgresURL:   "postgres://user:" + marker + "@postgres.local:5432/db?password=" + marker + "#frag" + marker,
		DragonflyAddr: "redis://user:" + marker + "@dragonfly.local:6379/0?token=" + marker + "#frag" + marker,
		TemporalAddr:  "temporal://user:" + marker + "@temporal.local:7233/ns?token=" + marker + "#frag" + marker + "\nINJECT",
	}

	got := RedactedConfig(cfg)
	assertRuntimeEndpointSafe(t, "postgres_url", got.PostgresURL, marker)
	assertRuntimeEndpointSafe(t, "dragonfly_addr", got.DragonflyAddr, marker)
	assertRuntimeEndpointSafe(t, "temporal_addr", got.TemporalAddr, marker)
	if got.PostgresURL != "postgres://***@postgres.local:5432/db" {
		t.Fatalf("PostgresURL = %q, want redacted URL without query/fragment", got.PostgresURL)
	}
	if got.DragonflyAddr != "dragonfly.local:6379" {
		t.Fatalf("DragonflyAddr = %q, want sanitized topology endpoint", got.DragonflyAddr)
	}
	if got.TemporalAddr != "temporal.local:7233" {
		t.Fatalf("TemporalAddr = %q, want sanitized topology endpoint", got.TemporalAddr)
	}
}

func TestRuntimeCheckResultsSanitizeServiceEndpointsAndErrors(t *testing.T) {
	marker := "pm-review-fix-check-secret"
	cfg := Config{
		DragonflyAddr: "redis://user:" + marker + "@127.0.0.1:2/0?token=" + marker + "#frag" + marker,
		TemporalAddr:  "temporal://user:" + marker + "@127.0.0.1:3/ns?token=" + marker + "#frag" + marker + "\nINJECT",
		Timeout:       20 * time.Millisecond,
	}

	dragonfly := checkDragonfly(context.Background(), cfg)
	temporal := checkTemporal(context.Background(), cfg)
	if dragonfly.Endpoint != "127.0.0.1:2" {
		t.Fatalf("dragonfly endpoint = %q, want sanitized topology endpoint", dragonfly.Endpoint)
	}
	if temporal.Endpoint != "127.0.0.1:3" {
		t.Fatalf("temporal endpoint = %q, want sanitized topology endpoint", temporal.Endpoint)
	}
	assertRuntimeEndpointSafe(t, "dragonfly endpoint", dragonfly.Endpoint, marker)
	assertRuntimeEndpointSafe(t, "dragonfly error", dragonfly.Error, marker)
	assertRuntimeEndpointSafe(t, "temporal endpoint", temporal.Endpoint, marker)
	assertRuntimeEndpointSafe(t, "temporal error", temporal.Error, marker)
}

func assertRuntimeEndpointSafe(t *testing.T, label, value, marker string) {
	t.Helper()
	for _, forbidden := range []string{marker, "user:", "token=", "password=", "#frag", "\n", "\r", "\t", "\x1b"} {
		if strings.Contains(value, forbidden) {
			t.Fatalf("%s leaked %q in %q", label, forbidden, value)
		}
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

func TestDragonflyCheckDoesNotWriteRedisDiagnosticsToProcessStderr(t *testing.T) {
	stderr := captureProcessStderr(t, func() {
		_ = checkDragonfly(context.Background(), Config{DragonflyAddr: "127.0.0.1:2", Timeout: 50 * time.Millisecond})
	})
	if strings.Contains(stderr, "redis:") {
		t.Fatalf("Dragonfly check wrote raw Redis diagnostics to process stderr: %q", stderr)
	}
}

func captureProcessStderr(t *testing.T, fn func()) string {
	t.Helper()

	oldFD, err := syscall.Dup(int(os.Stderr.Fd()))
	if err != nil {
		t.Fatalf("dup stderr: %v", err)
	}
	restored := false
	restore := func() {
		if restored {
			return
		}
		restored = true
		_ = syscall.Dup2(oldFD, int(os.Stderr.Fd()))
		_ = syscall.Close(oldFD)
	}
	defer restore()

	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe stderr: %v", err)
	}
	defer reader.Close()

	done := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, reader)
		done <- buf.String()
	}()

	if err := syscall.Dup2(int(writer.Fd()), int(os.Stderr.Fd())); err != nil {
		_ = writer.Close()
		t.Fatalf("redirect stderr: %v", err)
	}
	fn()
	restore()
	_ = writer.Close()
	return <-done
}
