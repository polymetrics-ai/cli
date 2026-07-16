package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"polymetrics.ai/internal/config"
	"polymetrics.ai/internal/schedule"
)

func runSchedule(ctx context.Context, cfg config.Config, root string, args []string, stdout io.Writer, jsonOut bool) error {
	if len(args) == 0 {
		return usageErrorf("usage: pm schedule <create|list|install|remove>")
	}
	switch args[0] {
	case "create":
		return runScheduleCreate(root, args[1:], stdout, jsonOut)
	case "list":
		return runScheduleList(root, args[1:], stdout, jsonOut)
	case "install":
		return runScheduleInstall(ctx, cfg, root, args[1:], stdout, jsonOut)
	case "remove":
		return runScheduleRemove(ctx, cfg, root, args[1:], stdout, jsonOut)
	default:
		return usageErrorf("unknown schedule subcommand %q", args[0])
	}
}

func runScheduleCreate(root string, args []string, stdout io.Writer, jsonOut bool) error {
	flags := parseFlags(args)
	name := flags.first("name")
	cron := flags.first("cron")
	flow := flags.first("flow")

	if name == "" || cron == "" || flow == "" {
		return usageErrorf("pm schedule create requires --name, --cron, --flow")
	}

	if _, err := schedule.ParseCron(cron); err != nil {
		return validationErrorf("invalid --cron: %v", err)
	}

	now := time.Now().UTC()
	m := schedule.Manifest{
		Name:      name,
		Cron:      cron,
		Flow:      flow,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := schedule.Save(root, m, false); err != nil {
		if isAlreadyExists(err) {
			return validationErrorf("%v", err)
		}
		return err
	}

	if jsonOut {
		return writeJSON(stdout, envelope{"kind": "Schedule", "ok": true, "schedule": m})
	}
	fmt.Fprintf(stdout, "Created schedule %s (cron: %s, flow: %s)\n", m.Name, m.Cron, m.Flow)
	return nil
}

func runScheduleList(root string, args []string, stdout io.Writer, jsonOut bool) error {
	manifests, err := schedule.List(root)
	if err != nil {
		return err
	}
	if manifests == nil {
		manifests = []schedule.Manifest{}
	}

	if jsonOut {
		return writeJSON(stdout, envelope{"kind": "ScheduleList", "schedules": manifests})
	}
	for _, m := range manifests {
		fmt.Fprintf(stdout, "%s\t%s\t%s\n", m.Name, m.Cron, m.Flow)
	}
	return nil
}

func runScheduleInstall(ctx context.Context, cfg config.Config, root string, args []string, stdout io.Writer, jsonOut bool) error {
	flags := parseFlags(args)
	positionals := flags.values["_"]
	if len(positionals) == 0 {
		return usageErrorf("pm schedule install <name> [--crontab]")
	}
	name := positionals[0]
	forceCrontab := flags.first("crontab") == "true"

	m, err := schedule.Load(root, name)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return validationErrorf("schedule %q not found", name)
		}
		return err
	}
	m.Root = root

	pmBin, _ := os.Executable()
	backend := schedule.SelectBackendFromConfig(ctx, forceCrontab, nil, scheduleConfig(cfg))

	if err := backend.Install(ctx, m, pmBin); err != nil {
		return fmt.Errorf("install failed: %w", err)
	}

	if jsonOut {
		return writeJSON(stdout, envelope{"kind": "ScheduleInstall", "ok": true, "schedule": m, "backend": string(backend.Kind())})
	}
	fmt.Fprintf(stdout, "Installed schedule %s via %s\n", m.Name, backend.Kind())
	return nil
}

func runScheduleRemove(ctx context.Context, cfg config.Config, root string, args []string, stdout io.Writer, jsonOut bool) error {
	flags := parseFlags(args)
	positionals := flags.values["_"]
	if len(positionals) == 0 {
		return usageErrorf("pm schedule remove <name> [--crontab]")
	}
	name := positionals[0]
	forceCrontab := flags.first("crontab") == "true"

	if _, err := schedule.Load(root, name); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return validationErrorf("schedule %q not found", name)
		}
		return err
	}

	// Best-effort backend removal (ignores errors if backend binary absent).
	backendCfg := scheduleConfig(cfg)
	backend := schedule.SelectBackendFromConfig(ctx, forceCrontab, nil, backendCfg)
	_ = backend.Remove(ctx, name)
	if backend.Kind() != schedule.KindCrontab {
		_ = (schedule.CrontabBackend{File: backendCfg.CrontabFile}).Remove(ctx, name)
	}

	if err := schedule.Delete(root, name); err != nil {
		return err
	}

	if jsonOut {
		return writeJSON(stdout, envelope{"kind": "ScheduleRemove", "ok": true, "name": name})
	}
	fmt.Fprintf(stdout, "Removed schedule %s\n", name)
	return nil
}

func scheduleConfig(cfg config.Config) schedule.BackendConfig {
	backendCfg := schedule.BackendConfig{CrontabFile: cfg.Schedule.CrontabFile}
	if cfg.IsExplicit("runtime.temporal_addr") {
		backendCfg.TemporalAddr = cfg.Runtime.TemporalAddr
	}
	return backendCfg
}

func isAlreadyExists(err error) bool {
	if err == nil {
		return false
	}
	// schedule.Save returns a plain error with "already exists" in the message
	return containsStr(err.Error(), "already exists")
}

func containsStr(s, sub string) bool {
	if len(sub) == 0 {
		return true
	}
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
