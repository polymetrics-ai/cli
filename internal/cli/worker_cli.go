package cli

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/spf13/cobra"

	"polymetrics.ai/internal/config"
	"polymetrics.ai/internal/temporalprobe"
	"polymetrics.ai/internal/worker"
)

type workerStatusFunc func(context.Context, string) bool
type workerServeFunc func(context.Context, string, *worker.PodmanActivities, func()) error

type workerCommandRuntime struct {
	status workerStatusFunc
	serve  workerServeFunc
}

func defaultWorkerCommandRuntime() workerCommandRuntime {
	return workerCommandRuntime{
		status: temporalprobe.Probe,
		serve:  worker.ServeWithActivitiesReady,
	}
}

func newWorkerCobraCommand(ctx context.Context, cfg config.Config, stdout io.Writer, jsonOut bool) *cobra.Command {
	return newWorkerCobraCommandWithRuntime(ctx, cfg, stdout, jsonOut, defaultWorkerCommandRuntime())
}

func newWorkerCobraCommandWithRuntime(ctx context.Context, cfg config.Config, stdout io.Writer, jsonOut bool, runtime workerCommandRuntime) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "worker",
		Hidden:            true,
		Args:              cobra.ArbitraryArgs,
		SilenceErrors:     true,
		SilenceUsage:      true,
		ValidArgsFunction: completeNoFile,
		RunE: func(_ *cobra.Command, args []string) error {
			if len(args) > 0 {
				return markCobraLegacyError(usageErrorf("worker: unknown subcommand %q (want serve|status)", args[0]))
			}
			return markCobraLegacyError(writeManual("worker", stdout, jsonOut))
		},
	}
	setManualHelp(cmd, "worker", stdout, jsonOut)
	cmd.AddCommand(newWorkerStatusCobraCommand(ctx, cfg, stdout, jsonOut, runtime))
	cmd.AddCommand(newWorkerServeCobraCommand(ctx, cfg, stdout, jsonOut, runtime))
	cmd.AddCommand(newWorkerHelpCobraCommand(stdout, jsonOut))
	return cmd
}

func newWorkerStatusCobraCommand(ctx context.Context, cfg config.Config, stdout io.Writer, jsonOut bool, runtime workerCommandRuntime) *cobra.Command {
	cmd := newWorkerActionCobraCommand("status", func(_ *cobra.Command, args []string) error {
		if len(args) > 0 && isHelpArg(args[0]) {
			return markCobraLegacyError(writeManual("worker", stdout, jsonOut))
		}
		return markCobraLegacyError(runWorkerStatus(ctx, cfg, stdout, jsonOut, runtime))
	})
	setManualHelp(cmd, "worker", stdout, jsonOut)
	return cmd
}

func newWorkerServeCobraCommand(ctx context.Context, cfg config.Config, stdout io.Writer, jsonOut bool, runtime workerCommandRuntime) *cobra.Command {
	cmd := newWorkerActionCobraCommand("serve", func(_ *cobra.Command, args []string) error {
		if len(args) > 0 && isHelpArg(args[0]) {
			return markCobraLegacyError(writeManual("worker", stdout, jsonOut))
		}
		return markCobraLegacyError(runWorkerServe(ctx, cfg, stdout, jsonOut, runtime))
	})
	setManualHelp(cmd, "worker", stdout, jsonOut)
	return cmd
}

func newWorkerHelpCobraCommand(stdout io.Writer, jsonOut bool) *cobra.Command {
	cmd := newWorkerActionCobraCommand("help", func(_ *cobra.Command, _ []string) error {
		return markCobraLegacyError(writeManual("worker", stdout, jsonOut))
	})
	cmd.Hidden = true
	setManualHelp(cmd, "worker", stdout, jsonOut)
	return cmd
}

func newWorkerActionCobraCommand(use string, run func(*cobra.Command, []string) error) *cobra.Command {
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

func explicitTemporalAddr(cfg config.Config) string {
	if cfg.IsExplicit("runtime.temporal_addr") {
		return cfg.Runtime.TemporalAddr
	}
	return ""
}

func runWorkerServe(ctx context.Context, cfg config.Config, stdout io.Writer, jsonOut bool, runtime workerCommandRuntime) error {
	addr := explicitTemporalAddr(cfg)
	if addr == "" {
		return validationErrorf("worker serve: POLYMETRICS_TEMPORAL_ADDR is not set")
	}
	ready := func() {
		if jsonOut {
			_ = writeJSON(stdout, envelope{"kind": "WorkerServe", "status": "ready", "addr": addr, "task_queue": worker.TaskQueue})
		} else {
			fmt.Fprintf(stdout, "pm worker serving on task queue %q (temporal=%s); Ctrl-C to stop\n", worker.TaskQueue, addr)
		}
	}
	activities := worker.NewPodmanActivities(cfg.RLM.PodmanBin, cfg.RLM.Image)
	return runtime.serve(ctx, addr, activities, ready)
}

func runWorkerStatus(ctx context.Context, cfg config.Config, stdout io.Writer, jsonOut bool, runtime workerCommandRuntime) error {
	addr := explicitTemporalAddr(cfg)
	probeCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	reachable := addr != "" && runtime.status(probeCtx, addr)

	status := "unavailable"
	if reachable {
		status = "ok"
	}
	if jsonOut {
		return writeJSON(stdout, envelope{
			"kind":       "WorkerStatus",
			"status":     status,
			"addr":       addr,
			"task_queue": worker.TaskQueue,
			"reachable":  reachable,
		})
	}
	if addr == "" {
		fmt.Fprintln(stdout, "worker status: POLYMETRICS_TEMPORAL_ADDR not set (RLM agent backend disabled)")
		return nil
	}
	fmt.Fprintf(stdout, "worker status: temporal=%s reachable=%v task_queue=%s\n", addr, reachable, worker.TaskQueue)
	return nil
}
