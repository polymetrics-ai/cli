package cli

import (
	"context"
	"fmt"
	"io"
	"time"

	"polymetrics.ai/internal/config"
	"polymetrics.ai/internal/temporalprobe"
	"polymetrics.ai/internal/worker"
)

type workerServeFunc func(context.Context, string, *worker.PodmanActivities, func()) error

var workerServe workerServeFunc = worker.ServeWithActivitiesReady

// runWorker dispatches `pm worker serve|status`. The worker daemon hosts the RLM
// Temporal workflow + podman activity for `pm rlm run --mode agent`.
func runWorker(ctx context.Context, cfg config.Config, args []string, stdout io.Writer, jsonOut bool) error {
	if len(args) == 0 {
		return usageErrorf("worker: missing subcommand (serve|status)")
	}
	switch args[0] {
	case "serve":
		return runWorkerServe(ctx, cfg, stdout, jsonOut)
	case "status":
		return runWorkerStatus(ctx, cfg, stdout, jsonOut)
	default:
		return usageErrorf("worker: unknown subcommand %q (want serve|status)", args[0])
	}
}

func explicitTemporalAddr(cfg config.Config) string {
	if cfg.IsExplicit("runtime.temporal_addr") {
		return cfg.Runtime.TemporalAddr
	}
	return ""
}

func runWorkerServe(ctx context.Context, cfg config.Config, stdout io.Writer, jsonOut bool) error {
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
	return workerServe(ctx, addr, activities, ready)
}

func runWorkerStatus(ctx context.Context, cfg config.Config, stdout io.Writer, jsonOut bool) error {
	addr := explicitTemporalAddr(cfg)
	probeCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	reachable := addr != "" && temporalprobe.Probe(probeCtx, addr)

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
