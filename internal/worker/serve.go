package worker

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/worker"
)

// Serve runs the long-lived RLM worker daemon: it hosts the workflow + podman
// activity on the shared task queue until ctx is cancelled or the process is
// interrupted. This is the primary worker model for `--mode agent`.
func Serve(ctx context.Context, addr string) error {
	return ServeWithActivities(ctx, addr, defaultActivities())
}

// ServeWithActivities runs the worker with explicitly configured activities.
func ServeWithActivities(ctx context.Context, addr string, acts *PodmanActivities) error {
	return ServeWithActivitiesReady(ctx, addr, acts, nil)
}

// ServeWithActivitiesReady runs the worker and calls ready after dial and worker start succeed.
func ServeWithActivitiesReady(ctx context.Context, addr string, acts *PodmanActivities, ready func()) error {
	if acts == nil {
		acts = defaultActivities()
	}
	c, err := dialTemporalClient(ctx, addr)
	if err != nil {
		return fmt.Errorf("worker serve: dial temporal: %w", err)
	}
	defer c.Close()
	logger := temporalLogger(ctx)

	w := worker.New(c, TaskQueue, temporalWorkerOptions(ctx, logger))
	registerWorker(w, acts)
	if err := w.Start(); err != nil {
		return fmt.Errorf("worker serve: start: %w", err)
	}
	defer w.Stop()
	if ready != nil {
		ready()
	}

	select {
	case <-ctx.Done():
		return nil
	case <-worker.InterruptCh():
		return nil
	}
}
