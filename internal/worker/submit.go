package worker

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
	tlog "go.temporal.io/sdk/log"
	"go.temporal.io/sdk/worker"

	"polymetrics.ai/internal/rlm"
)

// TaskQueue is the shared queue served by the `pm worker serve` daemon.
const TaskQueue = "polymetrics-rlm"

// DefaultEnvPass is the set of LLM env vars forwarded into the agent container.
var DefaultEnvPass = []string{"PM_LLM_BASE_URL", "PM_LLM_API_KEY", "PM_LLM_MODEL", "OPENROUTER_API_KEY", "PM_LLM_PROVIDER"}

// SubmitterFor returns an rlm.SubmitFunc that runs the RLM workflow over Temporal
// and the closer to release resources. When embedded is true it starts a worker
// in this process on a unique per-process queue (dev fallback). When false it is
// a thin client that targets the shared queue served by `pm worker serve`.
func SubmitterFor(addr string, embedded bool) (rlm.SubmitFunc, func() error, error) {
	return SubmitterForActivities(addr, embedded, defaultActivities())
}

func SubmitterForActivities(addr string, embedded bool, acts *PodmanActivities) (rlm.SubmitFunc, func() error, error) {
	if acts == nil {
		acts = defaultActivities()
	}
	c, err := client.Dial(client.Options{HostPort: addr, Logger: noopLogger{}})
	if err != nil {
		return nil, nil, fmt.Errorf("worker: dial temporal: %w", err)
	}

	taskQueue := TaskQueue
	var w worker.Worker
	if embedded {
		taskQueue = TaskQueue + "-embedded-" + randSuffix()
		w = worker.New(c, taskQueue, worker.Options{})
		registerWorker(w, acts)
		if err := w.Start(); err != nil {
			c.Close()
			return nil, nil, fmt.Errorf("worker: start embedded worker: %w", err)
		}
	}

	submit := func(ctx context.Context, req rlm.AgentRequest) (rlm.AgentResult, error) {
		opts := client.StartWorkflowOptions{
			ID:                       "rlm-" + req.Fingerprint,
			TaskQueue:                taskQueue,
			WorkflowIDConflictPolicy: enumspb.WORKFLOW_ID_CONFLICT_POLICY_USE_EXISTING,
			WorkflowIDReusePolicy:    enumspb.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE_FAILED_ONLY,
		}
		run, err := c.ExecuteWorkflow(ctx, opts, RemoteRLMWorkflow, req)
		if err != nil {
			return rlm.AgentResult{}, fmt.Errorf("worker: start workflow: %w", err)
		}
		var res rlm.AgentResult
		if err := run.Get(ctx, &res); err != nil {
			return rlm.AgentResult{}, err
		}
		return res, nil
	}

	closer := func() error {
		if w != nil {
			w.Stop()
		}
		c.Close()
		return nil
	}
	return submit, closer, nil
}

// registerWorker registers the workflow and podman activities on a worker.
func registerWorker(w worker.Worker, acts *PodmanActivities) {
	w.RegisterWorkflow(RemoteRLMWorkflow)
	w.RegisterActivity(acts)
}

// NewPodmanActivities builds production PodmanActivities from explicit typed settings.
func NewPodmanActivities(podmanBin, image string) *PodmanActivities {
	if podmanBin == "" {
		podmanBin = "podman"
	}
	if image == "" {
		image = "ghcr.io/polymetrics/rlm-agent:latest"
	}
	return &PodmanActivities{PodmanBin: podmanBin, Image: image, EnvPass: DefaultEnvPass}
}

func defaultActivities() *PodmanActivities {
	return NewPodmanActivities("", "")
}

func randSuffix() string {
	var b [6]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

type noopLogger struct{}

func (noopLogger) Debug(string, ...interface{}) {}
func (noopLogger) Info(string, ...interface{})  {}
func (noopLogger) Warn(string, ...interface{})  {}
func (noopLogger) Error(string, ...interface{}) {}

var _ tlog.Logger = noopLogger{}
