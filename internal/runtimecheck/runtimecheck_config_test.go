package runtimecheck

import (
	"testing"

	pmconfig "polymetrics.ai/internal/config"
)

func TestFromConfigUsesTypedRuntimeConfig(t *testing.T) {
	cfg := pmconfig.Config{
		Runtime: pmconfig.RuntimeConfig{
			PostgresURL:   "postgres://configured-host/polymetrics?sslmode=disable",
			DragonflyAddr: "configured-dragonfly:6379",
			TemporalAddr:  "configured-temporal:7233",
		},
	}

	got := FromConfig(cfg)
	if got.PostgresURL != cfg.Runtime.PostgresURL {
		t.Fatalf("PostgresURL = %q, want %q", got.PostgresURL, cfg.Runtime.PostgresURL)
	}
	if got.DragonflyAddr != cfg.Runtime.DragonflyAddr {
		t.Fatalf("DragonflyAddr = %q, want %q", got.DragonflyAddr, cfg.Runtime.DragonflyAddr)
	}
	if got.TemporalAddr != cfg.Runtime.TemporalAddr {
		t.Fatalf("TemporalAddr = %q, want %q", got.TemporalAddr, cfg.Runtime.TemporalAddr)
	}
	if got.Timeout <= 0 {
		t.Fatalf("Timeout = %s, want positive default", got.Timeout)
	}
}

func TestFromEnvDelegatesToTypedConfigAliases(t *testing.T) {
	t.Setenv("POLYMETRICS_POSTGRES_URL", "")
	t.Setenv("POLYMETRICS_DRAGONFLY_ADDR", "")
	t.Setenv("POLYMETRICS_TEMPORAL_ADDR", "")
	t.Setenv("PM_POSTGRES_URL", "postgres://alias-host/polymetrics?sslmode=disable")
	t.Setenv("PM_DRAGONFLY_ADDR", "alias-dragonfly:6379")
	t.Setenv("PM_TEMPORAL_ADDR", "alias-temporal:7233")

	got := FromEnv()
	if got.PostgresURL != "postgres://alias-host/polymetrics?sslmode=disable" {
		t.Fatalf("PostgresURL = %q, want PM_POSTGRES_URL alias", got.PostgresURL)
	}
	if got.DragonflyAddr != "alias-dragonfly:6379" {
		t.Fatalf("DragonflyAddr = %q, want PM_DRAGONFLY_ADDR alias", got.DragonflyAddr)
	}
	if got.TemporalAddr != "alias-temporal:7233" {
		t.Fatalf("TemporalAddr = %q, want PM_TEMPORAL_ADDR alias", got.TemporalAddr)
	}
}
