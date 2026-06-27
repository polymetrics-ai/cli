package schedule

import (
	"context"
	"os"
	"runtime"
)

// ProbeFunc checks whether a Temporal server at addr is reachable.
// Return true if reachable.
type ProbeFunc func(ctx context.Context, addr string) bool

// SelectBackend chooses the appropriate Backend for the current environment.
//
//   - forceCrontab: --crontab flag; always returns CrontabBackend.
//   - probe: injected probe for Temporal reachability; pass nil to use default (always false stub).
func SelectBackend(ctx context.Context, forceCrontab bool, probe ProbeFunc) Backend {
	if forceCrontab {
		return CrontabBackend{}
	}

	addr := os.Getenv("POLYMETRICS_TEMPORAL_ADDR")
	if addr != "" {
		p := probe
		if p == nil {
			p = func(_ context.Context, _ string) bool { return false }
		}
		if p(ctx, addr) {
			return TemporalBackend{Addr: addr}
		}
	}

	switch runtime.GOOS {
	case "darwin":
		return LaunchdBackend{}
	case "linux":
		return SystemdBackend{}
	default:
		return CrontabBackend{}
	}
}

// goOS returns runtime.GOOS; exposed so tests can call it without importing "runtime".
func goOS() string { return runtime.GOOS }
