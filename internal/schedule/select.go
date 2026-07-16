package schedule

import (
	"context"
	"runtime"

	pmconfig "polymetrics.ai/internal/config"
)

// ProbeFunc checks whether a Temporal server at addr is reachable.
// Return true if reachable.
type ProbeFunc func(ctx context.Context, addr string) bool

// BackendConfig is the schedule backend subset resolved from typed config.
type BackendConfig struct {
	TemporalAddr string
	CrontabFile  string
}

// SelectBackend chooses the appropriate Backend for the current environment.
//
//   - forceCrontab: --crontab flag; always returns CrontabBackend.
//   - probe: injected probe for Temporal reachability; pass nil to use default (always false stub).
func SelectBackend(ctx context.Context, forceCrontab bool, probe ProbeFunc) Backend {
	cfg, err := pmconfig.Load(pmconfig.Options{})
	if err != nil {
		return SelectBackendFromConfig(ctx, forceCrontab, probe, BackendConfig{})
	}
	backendCfg := BackendConfig{CrontabFile: cfg.Schedule.CrontabFile}
	if cfg.IsExplicit("runtime.temporal_addr") {
		backendCfg.TemporalAddr = cfg.Runtime.TemporalAddr
	}
	return SelectBackendFromConfig(ctx, forceCrontab, probe, backendCfg)
}

func SelectBackendFromConfig(ctx context.Context, forceCrontab bool, probe ProbeFunc, cfg BackendConfig) Backend {
	if forceCrontab {
		return CrontabBackend{File: cfg.CrontabFile}
	}

	addr := cfg.TemporalAddr
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
		return CrontabBackend{File: cfg.CrontabFile}
	}
}

// goOS returns runtime.GOOS; exposed so tests can call it without importing "runtime".
func goOS() string { return runtime.GOOS }
