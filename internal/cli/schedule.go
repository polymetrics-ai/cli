package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/spf13/cobra"

	"polymetrics.ai/internal/config"
	"polymetrics.ai/internal/schedule"
)

type scheduleCreateFlags struct {
	Names []string
	Crons []string
	Flows []string
}

type scheduleBackendFlags struct {
	Crontab bool
}

type scheduleCommandRuntime struct {
	now            func() time.Time
	executable     func() (string, error)
	selectBackend  func(context.Context, bool, schedule.BackendConfig) schedule.Backend
	crontabBackend func(string) schedule.Backend
}

func defaultScheduleCommandRuntime() scheduleCommandRuntime {
	return scheduleCommandRuntime{
		now:        func() time.Time { return time.Now().UTC() },
		executable: os.Executable,
		selectBackend: func(ctx context.Context, forceCrontab bool, cfg schedule.BackendConfig) schedule.Backend {
			return schedule.SelectBackendFromConfig(ctx, forceCrontab, nil, cfg)
		},
		crontabBackend: func(file string) schedule.Backend {
			return schedule.CrontabBackend{File: file}
		},
	}
}

func newScheduleCobraCommand(ctx context.Context, cfg config.Config, root string, stdout io.Writer, jsonOut bool) *cobra.Command {
	return newScheduleCobraCommandWithRuntime(ctx, cfg, root, stdout, jsonOut, defaultScheduleCommandRuntime())
}

func newScheduleCobraCommandWithRuntime(ctx context.Context, cfg config.Config, root string, stdout io.Writer, jsonOut bool, runtime scheduleCommandRuntime) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "schedule",
		Args:              cobra.ArbitraryArgs,
		SilenceErrors:     true,
		SilenceUsage:      true,
		ValidArgsFunction: completeNoFile,
		RunE: func(_ *cobra.Command, args []string) error {
			if len(args) > 0 {
				return markCobraLegacyError(usageErrorf("unknown schedule subcommand %q", args[0]))
			}
			return markCobraLegacyError(writeManual("schedule", stdout, jsonOut))
		},
	}
	setManualHelp(cmd, "schedule", stdout, jsonOut)
	cmd.AddCommand(newScheduleCreateCobraCommand(root, stdout, jsonOut, runtime))
	cmd.AddCommand(newScheduleListCobraCommand(root, stdout, jsonOut))
	cmd.AddCommand(newScheduleInstallCobraCommand(ctx, cfg, root, stdout, jsonOut, runtime))
	cmd.AddCommand(newScheduleRemoveCobraCommand(ctx, cfg, root, stdout, jsonOut, runtime))
	cmd.AddCommand(newScheduleHelpCobraCommand(stdout, jsonOut))
	return cmd
}

func newScheduleCreateCobraCommand(root string, stdout io.Writer, jsonOut bool, runtime scheduleCommandRuntime) *cobra.Command {
	var flags scheduleCreateFlags
	cmd := newScheduleActionCobraCommand("create", func(_ *cobra.Command, _ []string) error {
		return markCobraLegacyError(runScheduleCreate(root, flags, stdout, jsonOut, runtime.now))
	})
	setManualHelp(cmd, "schedule", stdout, jsonOut)
	addScheduleStringArrayFlag(cmd, &flags.Names, "name", "schedule name")
	addScheduleStringArrayFlag(cmd, &flags.Crons, "cron", "five-field cron expression")
	addScheduleStringArrayFlag(cmd, &flags.Flows, "flow", "named flow to run")
	return cmd
}

func newScheduleListCobraCommand(root string, stdout io.Writer, jsonOut bool) *cobra.Command {
	cmd := newScheduleActionCobraCommand("list", func(_ *cobra.Command, _ []string) error {
		return markCobraLegacyError(runScheduleList(root, stdout, jsonOut))
	})
	setManualHelp(cmd, "schedule", stdout, jsonOut)
	return cmd
}

func newScheduleInstallCobraCommand(ctx context.Context, cfg config.Config, root string, stdout io.Writer, jsonOut bool, runtime scheduleCommandRuntime) *cobra.Command {
	var flags scheduleBackendFlags
	cmd := newScheduleActionCobraCommand("install", func(cmd *cobra.Command, _ []string) error {
		name, ok := scheduleOperand(cmd)
		if !ok {
			return markCobraLegacyError(usageErrorf("pm schedule install <name> [--crontab]"))
		}
		return markCobraLegacyError(runScheduleInstall(ctx, cfg, root, name, flags, stdout, jsonOut, runtime))
	})
	setManualHelp(cmd, "schedule", stdout, jsonOut)
	cmd.Flags().BoolVar(&flags.Crontab, "crontab", false, "force the crontab backend")
	return cmd
}

func newScheduleRemoveCobraCommand(ctx context.Context, cfg config.Config, root string, stdout io.Writer, jsonOut bool, runtime scheduleCommandRuntime) *cobra.Command {
	var flags scheduleBackendFlags
	cmd := newScheduleActionCobraCommand("remove", func(cmd *cobra.Command, _ []string) error {
		name, ok := scheduleOperand(cmd)
		if !ok {
			return markCobraLegacyError(usageErrorf("pm schedule remove <name> [--crontab]"))
		}
		return markCobraLegacyError(runScheduleRemove(ctx, cfg, root, name, flags, stdout, jsonOut, runtime))
	})
	setManualHelp(cmd, "schedule", stdout, jsonOut)
	cmd.Flags().BoolVar(&flags.Crontab, "crontab", false, "force the crontab backend")
	return cmd
}

func newScheduleHelpCobraCommand(stdout io.Writer, jsonOut bool) *cobra.Command {
	cmd := newScheduleActionCobraCommand("help", func(_ *cobra.Command, _ []string) error {
		return markCobraLegacyError(writeManual("schedule", stdout, jsonOut))
	})
	cmd.Hidden = true
	setManualHelp(cmd, "schedule", stdout, jsonOut)
	return cmd
}

func newScheduleActionCobraCommand(use string, run func(*cobra.Command, []string) error) *cobra.Command {
	return &cobra.Command{
		Use:           use,
		Args:          cobra.ArbitraryArgs,
		SilenceErrors: true,
		SilenceUsage:  true,
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
		ValidArgsFunction: completeNoFile,
		RunE:              run,
	}
}

func addScheduleStringArrayFlag(cmd *cobra.Command, target *[]string, name, usage string) {
	cmd.Flags().StringArrayVar(target, name, nil, usage)
	if flag := cmd.Flags().Lookup(name); flag != nil {
		flag.NoOptDefVal = "true"
	}
}

func scheduleOperand(cmd *cobra.Command) (string, bool) {
	state, ok := cmd.Context().Value(scheduleCommandStateKey{}).(scheduleCommandState)
	if !ok || !state.operandSet {
		return "", false
	}
	return state.operand, true
}

func runScheduleCreate(root string, flags scheduleCreateFlags, stdout io.Writer, jsonOut bool, now func() time.Time) error {
	name := lastString(flags.Names)
	cron := lastString(flags.Crons)
	flow := lastString(flags.Flows)

	if name == "" || cron == "" || flow == "" {
		return usageErrorf("pm schedule create requires --name, --cron, --flow")
	}

	if _, err := schedule.ParseCron(cron); err != nil {
		return validationErrorf("invalid --cron: %v", err)
	}

	createdAt := now().UTC()
	manifest := schedule.Manifest{
		Name:      name,
		Cron:      cron,
		Flow:      flow,
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
	}

	if err := schedule.Save(root, manifest, false); err != nil {
		if isAlreadyExists(err) {
			return validationErrorf("%v", err)
		}
		return err
	}

	if jsonOut {
		return writeJSON(stdout, envelope{"kind": "Schedule", "ok": true, "schedule": manifest})
	}
	fmt.Fprintf(stdout, "Created schedule %s (cron: %s, flow: %s)\n", manifest.Name, manifest.Cron, manifest.Flow)
	return nil
}

func runScheduleList(root string, stdout io.Writer, jsonOut bool) error {
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
	for _, manifest := range manifests {
		fmt.Fprintf(stdout, "%s\t%s\t%s\n", manifest.Name, manifest.Cron, manifest.Flow)
	}
	return nil
}

func runScheduleInstall(ctx context.Context, cfg config.Config, root, name string, flags scheduleBackendFlags, stdout io.Writer, jsonOut bool, runtime scheduleCommandRuntime) error {
	manifest, err := schedule.Load(root, name)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return validationErrorf("schedule %q not found", name)
		}
		return err
	}
	manifest.Root = root

	pmBin, _ := runtime.executable()
	backend := runtime.selectBackend(ctx, flags.Crontab, scheduleConfig(cfg))
	if err := backend.Install(ctx, manifest, pmBin); err != nil {
		return fmt.Errorf("install failed: %w", err)
	}

	if jsonOut {
		return writeJSON(stdout, envelope{"kind": "ScheduleInstall", "ok": true, "schedule": manifest, "backend": string(backend.Kind())})
	}
	fmt.Fprintf(stdout, "Installed schedule %s via %s\n", manifest.Name, backend.Kind())
	return nil
}

func runScheduleRemove(ctx context.Context, cfg config.Config, root, name string, flags scheduleBackendFlags, stdout io.Writer, jsonOut bool, runtime scheduleCommandRuntime) error {
	if _, err := schedule.Load(root, name); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return validationErrorf("schedule %q not found", name)
		}
		return err
	}

	backendCfg := scheduleConfig(cfg)
	backend := runtime.selectBackend(ctx, flags.Crontab, backendCfg)
	_ = backend.Remove(ctx, name)
	if backend.Kind() != schedule.KindCrontab {
		_ = runtime.crontabBackend(backendCfg.CrontabFile).Remove(ctx, name)
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
