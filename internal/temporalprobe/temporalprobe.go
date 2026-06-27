// Package temporalprobe reports whether a Temporal frontend is reachable. It is
// kept separate from internal/rlm so the rlm package never imports
// go.temporal.io; the CLI injects Probe into the RLM agent backend.
package temporalprobe

import (
	"context"
	"time"

	"go.temporal.io/sdk/client"
	tlog "go.temporal.io/sdk/log"
)

const defaultTimeout = 3 * time.Second

// Probe reports whether a Temporal frontend at addr is reachable and healthy.
// It bounds dial+health-check time so an unreachable address fails fast instead
// of hanging. An empty addr is always false. The bound honors a shorter context
// deadline when present.
func Probe(ctx context.Context, addr string) bool {
	if addr == "" {
		return false
	}
	timeout := defaultTimeout
	if dl, ok := ctx.Deadline(); ok {
		if d := time.Until(dl); d > 0 && d < timeout {
			timeout = d
		}
	}

	done := make(chan bool, 1)
	go func() {
		c, err := client.Dial(client.Options{HostPort: addr, Logger: noopLogger{}})
		if err != nil {
			done <- false
			return
		}
		defer c.Close()
		checkCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		_, err = c.CheckHealth(checkCtx, &client.CheckHealthRequest{})
		done <- err == nil
	}()

	select {
	case ok := <-done:
		return ok
	case <-time.After(timeout):
		return false
	}
}

type noopLogger struct{}

func (noopLogger) Debug(string, ...interface{}) {}
func (noopLogger) Info(string, ...interface{})  {}
func (noopLogger) Warn(string, ...interface{})  {}
func (noopLogger) Error(string, ...interface{}) {}

var _ tlog.Logger = noopLogger{}
