package worker

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

// Serve runs the long-lived RLM worker daemon: it hosts the workflow + podman
// activity on the shared task queue until ctx is cancelled or the process is
// interrupted. This is the primary worker model for `--mode agent`.
func Serve(ctx context.Context, addr string) error {
	c, err := client.Dial(client.Options{HostPort: addr, Logger: noopLogger{}})
	if err != nil {
		return fmt.Errorf("worker serve: dial temporal: %w", err)
	}
	defer c.Close()

	w := worker.New(c, TaskQueue, worker.Options{})
	registerWorker(w, defaultActivities())
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
