package worker

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/client"
	tlog "go.temporal.io/sdk/log"
	"go.temporal.io/sdk/worker"

	pmlogging "polymetrics.ai/internal/logging"
)

// Serve runs the long-lived RLM worker daemon: it hosts the workflow + podman
// activity on the shared task queue until ctx is cancelled or the process is
// interrupted. This is the primary worker model for `--mode agent`.
func Serve(ctx context.Context, addr string) error {
	return ServeWithActivities(ctx, addr, defaultActivities())
}

// ServeWithActivities runs the worker with explicitly configured activities.
func ServeWithActivities(ctx context.Context, addr string, acts *PodmanActivities) error {
	if acts == nil {
		acts = defaultActivities()
	}
	c, err := client.Dial(client.Options{HostPort: addr, Logger: tlog.NewStructuredLogger(pmlogging.FromContext(ctx))})
	if err != nil {
		return fmt.Errorf("worker serve: dial temporal: %w", err)
	}
	defer c.Close()

	w := worker.New(c, TaskQueue, worker.Options{})
	registerWorker(w, acts)
	if err := w.Start(); err != nil {
		return fmt.Errorf("worker serve: start: %w", err)
	}
	defer w.Stop()

	select {
	case <-ctx.Done():
		return nil
	case <-worker.InterruptCh():
		return nil
	}
}
