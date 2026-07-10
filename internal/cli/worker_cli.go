package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"polymetrics.ai/internal/temporalprobe"
	"polymetrics.ai/internal/worker"
)

// runWorker dispatches `pm worker serve|status`. The worker daemon hosts the RLM
// Temporal workflow + podman activity for `pm rlm run --mode agent`.
func runWorker(ctx context.Context, args []string, stdout io.Writer, jsonOut bool) error {
	if len(args) == 0 {
		return usageErrorf("worker: missing subcommand (serve|status)")
	}
	switch args[0] {
	case "serve":
		return runWorkerServe(ctx, stdout, jsonOut)
	case "status":
		return runWorkerStatus(ctx, stdout, jsonOut)
	default:
		return usageErrorf("worker: unknown subcommand %q (want serve|status)", args[0])
	}
}

func temporalAddr() string { return os.Getenv("POLYMETRICS_TEMPORAL_ADDR") }

func runWorkerServe(ctx context.Context, stdout io.Writer, jsonOut bool) error {
	addr := temporalAddr()
	if addr == "" {
		return validationErrorf("worker serve: POLYMETRICS_TEMPORAL_ADDR is not set")
	}
	if jsonOut {
		_ = writeJSON(stdout, envelope{"kind": "WorkerServe", "status": "starting", "addr": addr, "task_queue": worker.TaskQueue})
	} else {
		if _, err := fmt.Fprintf(stdout, "pm worker serving on task queue %q (temporal=%s); Ctrl-C to stop\n", worker.TaskQueue, addr); err != nil {
			return err
		}
	}
	return worker.Serve(ctx, addr)
}

func runWorkerStatus(ctx context.Context, stdout io.Writer, jsonOut bool) error {
	addr := temporalAddr()
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
		_, err := fmt.Fprintln(stdout, "worker status: POLYMETRICS_TEMPORAL_ADDR not set (RLM agent backend disabled)")
		return err
	}
	_, err := fmt.Fprintf(stdout, "worker status: temporal=%s reachable=%v task_queue=%s\n", addr, reachable, worker.TaskQueue)
	return err
}
