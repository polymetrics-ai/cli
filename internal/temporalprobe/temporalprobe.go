// Package temporalprobe reports whether a Temporal frontend is reachable. It is
// kept separate from internal/rlm so the rlm package never imports
// go.temporal.io; the CLI injects Probe into the RLM agent backend.
package temporalprobe

import (
	"context"
	"time"

	"go.temporal.io/sdk/client"
	tlog "go.temporal.io/sdk/log"

	pmlogging "polymetrics.ai/internal/logging"
)

const defaultTimeout = 3 * time.Second

type temporalHealthClient interface {
	CheckHealth(context.Context, *client.CheckHealthRequest) (*client.CheckHealthResponse, error)
	Close()
}

var dialTemporal = func(ctx context.Context, addr string, timeout time.Duration, logger tlog.Logger) (temporalHealthClient, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return client.DialContext(ctx, client.Options{
		HostPort: addr,
		Logger:   logger,
		ConnectionOptions: client.ConnectionOptions{
			GetSystemInfoTimeout: timeout,
		},
	})
}

// Probe reports whether a Temporal frontend at addr is reachable and healthy.
// It bounds dial+health-check time so an unreachable address fails fast instead
// of hanging. An empty addr is always false. The bound honors a shorter context
// deadline when present.
func Probe(ctx context.Context, addr string) bool {
	if addr == "" {
		return false
	}
	if ctx == nil {
		ctx = context.Background()
	}
	timeout := defaultTimeout
	if dl, ok := ctx.Deadline(); ok {
		if d := time.Until(dl); d <= 0 {
			return false
		} else if d < timeout {
			timeout = d
		}
	}
	probeCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	c, err := dialTemporal(probeCtx, addr, timeout, tlog.NewStructuredLogger(pmlogging.FromContext(probeCtx)))
	if err != nil {
		return false
	}
	defer c.Close()
	_, err = c.CheckHealth(probeCtx, &client.CheckHealthRequest{})
	return err == nil
}
