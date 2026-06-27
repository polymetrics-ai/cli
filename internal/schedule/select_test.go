package schedule

import (
	"context"
	"os"
	"testing"
)

// Group D — SelectBackend env-driven selection.

func TestSelectBackend_ForceCrontab(t *testing.T) {
	ctx := context.Background()
	b := SelectBackend(ctx, true, nil)
	if b.Kind() != KindCrontab {
		t.Fatalf("--crontab flag: got kind %q, want %q", b.Kind(), KindCrontab)
	}
}

func TestSelectBackend_TemporalReachable(t *testing.T) {
	ctx := context.Background()
	os.Setenv("POLYMETRICS_TEMPORAL_ADDR", "localhost:7233")
	defer os.Unsetenv("POLYMETRICS_TEMPORAL_ADDR")
	alwaysOK := func(_ context.Context, _ string) bool { return true }
	b := SelectBackend(ctx, false, alwaysOK)
	if b.Kind() != KindTemporal {
		t.Fatalf("Temporal reachable: got kind %q, want %q", b.Kind(), KindTemporal)
	}
}

func TestSelectBackend_TemporalUnreachableFallsBack(t *testing.T) {
	ctx := context.Background()
	os.Setenv("POLYMETRICS_TEMPORAL_ADDR", "localhost:7233")
	defer os.Unsetenv("POLYMETRICS_TEMPORAL_ADDR")
	alwaysFail := func(_ context.Context, _ string) bool { return false }
	b := SelectBackend(ctx, false, alwaysFail)
	if b.Kind() == KindTemporal {
		t.Fatal("Temporal unreachable: expected fallback, got temporal")
	}
}

func TestSelectBackend_Darwin(t *testing.T) {
	ctx := context.Background()
	os.Unsetenv("POLYMETRICS_TEMPORAL_ADDR")
	b := SelectBackend(ctx, false, nil)
	// On darwin expect launchd; the NOT-YET-IMPLEMENTED stub returns crontab, so this is RED on darwin.
	if goOS() == "darwin" {
		if b.Kind() != KindLaunchd {
			t.Fatalf("darwin: got kind %q, want %q", b.Kind(), KindLaunchd)
		}
	}
}

func TestSelectBackend_Linux(t *testing.T) {
	ctx := context.Background()
	os.Unsetenv("POLYMETRICS_TEMPORAL_ADDR")
	b := SelectBackend(ctx, false, nil)
	if goOS() == "linux" {
		if b.Kind() == KindLaunchd {
			t.Fatalf("linux: should not return launchd backend")
		}
	}
}
