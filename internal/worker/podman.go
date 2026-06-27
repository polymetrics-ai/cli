// Package worker hosts the Temporal workflow and podman activity that run the
// containerized PI mono RLM agent. It is the only package (besides
// temporalprobe) that imports go.temporal.io, keeping internal/rlm Temporal-free.
package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"

	"polymetrics.ai/internal/rlm"
)

// Non-retryable application error type names. They are listed in the workflow's
// RetryPolicy.NonRetryableErrorTypes AND produced via NewNonRetryableApplicationError
// so a deterministic failure is never retried into an LLM-cost storm.
const (
	errBadAnalysisRequest  = "BadAnalysisRequest"
	errReflectionExhausted = "ReflectionExhausted"
	errLLMUnreachable      = "LLMUnreachable"
	errOOMKilled           = "OOMKilled"
)

func nonRetryableErrorTypes() []string {
	return []string{errBadAnalysisRequest, errReflectionExhausted, errLLMUnreachable, errOOMKilled}
}

// PodmanActivities holds the podman activity and its injectable seams (NewCmd /
// Reap / Heartbeat) so it is fully testable without a real container runtime.
type PodmanActivities struct {
	PodmanBin string
	// EnvPass lists env var names forwarded into the container (e.g. PM_LLM_*).
	EnvPass []string
	// Image is the pinned container image reference.
	Image string

	// NewCmd builds the command that runs one job. Injected in tests with a fake.
	NewCmd func(ctx context.Context, req rlm.AgentRequest, name string) *exec.Cmd
	// Reap force-removes a container by name. Injected in tests with a no-op.
	Reap func(name string)
	// Heartbeat records activity liveness. Defaults to activity.RecordHeartbeat.
	Heartbeat func(ctx context.Context)

	HeartbeatInterval time.Duration
}

func (p *PodmanActivities) podmanBin() string {
	if p.PodmanBin != "" {
		return p.PodmanBin
	}
	return "podman"
}

func (p *PodmanActivities) newCmd() func(context.Context, rlm.AgentRequest, string) *exec.Cmd {
	if p.NewCmd != nil {
		return p.NewCmd
	}
	return p.defaultNewCmd
}

func (p *PodmanActivities) reap() func(string) {
	if p.Reap != nil {
		return p.Reap
	}
	bin := p.podmanBin()
	return func(name string) {
		// Best-effort: force-remove from a fresh, short context so a stale or
		// orphaned container is reliably reaped even after cancellation.
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		_ = exec.CommandContext(ctx, bin, "rm", "-f", name).Run()
	}
}

func (p *PodmanActivities) heartbeat() func(context.Context) {
	if p.Heartbeat != nil {
		return p.Heartbeat
	}
	return func(ctx context.Context) { activity.RecordHeartbeat(ctx, "running") }
}

func (p *PodmanActivities) heartbeatInterval() time.Duration {
	if p.HeartbeatInterval > 0 {
		return p.HeartbeatInterval
	}
	return 20 * time.Second
}

// RunPodman executes one RLM job in a one-shot container and returns counts. The
// scored rows are written to req.JobDir/out by the container and read by the
// analyzer; they never flow back through Temporal history.
func (p *PodmanActivities) RunPodman(ctx context.Context, req rlm.AgentRequest) (rlm.AgentResult, error) {
	name := containerName(req.Fingerprint)
	reap := p.reap()
	reap(name) // pre-reap any stale container with the same deterministic name

	cmd := p.newCmd()(ctx, req, name)
	var out bytes.Buffer
	if cmd.Stdout == nil {
		cmd.Stdout = &out
	}
	if cmd.Stderr == nil {
		cmd.Stderr = &out
	}

	stopHB := p.startHeartbeat(ctx)
	runErr := cmd.Run()
	stopHB()

	if ctx.Err() != nil {
		// Cancellation/timeout: ensure the container is gone, not just the client.
		reap(name)
		return rlm.AgentResult{}, ctx.Err()
	}
	if runErr != nil {
		reap(name)
		return rlm.AgentResult{}, classifyExit(runErr, out.String())
	}

	res := rlm.AgentResult{JobDir: req.JobDir}
	if m := readManifest(req.JobDir); m != nil {
		res.RecordsScored = m.ExpectedCount
		res.RecordsRead = m.RecordsRead
	}
	return res, nil
}

// Cleanup force-removes the container for a fingerprint. Run from a disconnected
// workflow context on cancellation.
func (p *PodmanActivities) Cleanup(ctx context.Context, fingerprint string) error {
	p.reap()(containerName(fingerprint))
	return nil
}

func (p *PodmanActivities) startHeartbeat(ctx context.Context) func() {
	hb := p.heartbeat()
	ticker := time.NewTicker(p.heartbeatInterval())
	done := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				hb(ctx)
			case <-done:
				return
			}
		}
	}()
	return func() {
		ticker.Stop()
		close(done)
	}
}

// classifyExit maps a failed container run to a retryable or non-retryable
// Temporal error. Deterministic failures (image missing, OOM, reflection
// exhausted, LLM unreachable) are non-retryable; transient podman/infra errors
// are retryable.
func classifyExit(runErr error, stderr string) error {
	var ee *exec.ExitError
	if errors.As(runErr, &ee) {
		code := ee.ExitCode()
		low := strings.ToLower(stderr)
		switch {
		case code == 125 && (strings.Contains(low, "no such image") || strings.Contains(low, "unable to find image") || strings.Contains(low, "manifest unknown")):
			return temporal.NewNonRetryableApplicationError(
				"container image missing: run `pm agent image pull|build`", errBadAnalysisRequest, runErr)
		case code == 137:
			return temporal.NewNonRetryableApplicationError(
				"container OOM-killed: raise --memory or reduce the batch", errOOMKilled, runErr)
		case code == 3:
			return temporal.NewNonRetryableApplicationError(
				"agent exhausted reflection iterations without passing validation", errReflectionExhausted, runErr)
		case code == 4:
			return temporal.NewNonRetryableApplicationError(
				"LLM endpoint unreachable from the container", errLLMUnreachable, runErr)
		default:
			return fmt.Errorf("podman run failed (exit %d): %s", code, truncate(stderr, 2048))
		}
	}
	return fmt.Errorf("podman run error: %w", runErr)
}

func containerName(fingerprint string) string { return "rlm-" + fingerprint }

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

type manifest struct {
	ExpectedCount int `json:"expected_count"`
	RecordsRead   int `json:"records_read"`
}

func readManifest(jobDir string) *manifest {
	b, err := os.ReadFile(filepath.Join(jobDir, "out", "manifest.json"))
	if err != nil {
		return nil
	}
	var m manifest
	if json.Unmarshal(b, &m) != nil {
		return nil
	}
	return &m
}
