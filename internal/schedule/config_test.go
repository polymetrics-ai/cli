package schedule

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestSelectBackendFromConfigUsesExplicitTemporalAddr(t *testing.T) {
	ctx := context.Background()
	cfg := BackendConfig{TemporalAddr: "configured-temporal:7233"}
	alwaysOK := func(_ context.Context, got string) bool {
		return got == cfg.TemporalAddr
	}

	backend := SelectBackendFromConfig(ctx, false, alwaysOK, cfg)
	if backend.Kind() != KindTemporal {
		t.Fatalf("configured Temporal reachable: got kind %q, want %q", backend.Kind(), KindTemporal)
	}
}

func TestCrontabBackendUsesConfiguredFile(t *testing.T) {
	ctx := context.Background()
	crontabFile := t.TempDir() + "/crontab"
	backend := CrontabBackend{File: crontabFile}

	if err := backend.Install(ctx, nightlyManifest, testPmBin); err != nil {
		t.Fatalf("Install: %v", err)
	}
	data, err := os.ReadFile(crontabFile)
	if err != nil {
		t.Fatalf("read configured crontab: %v", err)
	}
	if !strings.Contains(string(data), "pm-schedule-nightly-leads") {
		t.Fatalf("configured crontab missing sentinel: %q", string(data))
	}

	if err := backend.Remove(ctx, nightlyManifest.Name); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	data, err = os.ReadFile(crontabFile)
	if err != nil {
		t.Fatalf("read configured crontab after remove: %v", err)
	}
	if strings.Contains(string(data), "pm-schedule-nightly-leads") {
		t.Fatalf("configured crontab sentinel remained: %q", string(data))
	}
}
