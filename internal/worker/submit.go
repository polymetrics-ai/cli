package worker

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	enumspb "go.temporal.io/api/enums/v1"
	workflowservice "go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	tlog "go.temporal.io/sdk/log"
	"go.temporal.io/sdk/worker"

	"polymetrics.ai/internal/events"
	pmlogging "polymetrics.ai/internal/logging"
	"polymetrics.ai/internal/rlm"
)

// TaskQueue is the shared queue served by the `pm worker serve` daemon.
const TaskQueue = "polymetrics-rlm"

var workflowPollInterval = time.Second
var temporalDialTimeout = 3 * time.Second

var temporalClientDial = func(ctx context.Context, addr string, logger tlog.Logger) (client.Client, error) {
	return client.DialContext(ctx, client.Options{
		HostPort: addr,
		Logger:   logger,
		ConnectionOptions: client.ConnectionOptions{
			GetSystemInfoTimeout: temporalDialTimeout,
		},
	})
}

// DefaultEnvPass is the set of LLM env vars forwarded into the agent container.
var DefaultEnvPass = []string{"PM_LLM_BASE_URL", "PM_LLM_API_KEY", "PM_LLM_MODEL", "OPENROUTER_API_KEY", "PM_LLM_PROVIDER"}

// SubmitterFor returns an rlm.SubmitFunc that runs the RLM workflow over Temporal
// and the closer to release resources. When embedded is true it starts a worker
// in this process on a unique per-process queue (dev fallback). When false it is
// a thin client that targets the shared queue served by `pm worker serve`.
func SubmitterFor(addr string, embedded bool) (rlm.SubmitFunc, func() error, error) {
	return SubmitterForActivitiesContext(context.Background(), addr, embedded, defaultActivities())
}

func SubmitterForActivities(addr string, embedded bool, acts *PodmanActivities) (rlm.SubmitFunc, func() error, error) {
	return SubmitterForActivitiesContext(context.Background(), addr, embedded, acts)
}

func SubmitterForActivitiesContext(ctx context.Context, addr string, embedded bool, acts *PodmanActivities) (rlm.SubmitFunc, func() error, error) {
	if acts == nil {
		acts = defaultActivities()
	}
	c, err := dialTemporalClient(ctx, addr)
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

	submit := submitterForWorkflowClient(temporalWorkflowClient{Client: c}, taskQueue)

	closer := func() error {
		if w != nil {
			w.Stop()
		}
		c.Close()
		return nil
	}
	return submit, closer, nil
}

func dialTemporalClient(ctx context.Context, addr string) (client.Client, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	dialCtx, cancel := context.WithTimeout(ctx, temporalDialTimeout)
	defer cancel()
	return temporalClientDial(dialCtx, addr, tlog.NewStructuredLogger(pmlogging.FromContext(dialCtx)))
}

type workflowRun interface {
	Get(context.Context, any) error
	GetID() string
	GetRunID() string
}

type workflowClient interface {
	ExecuteWorkflow(context.Context, client.StartWorkflowOptions, any, ...any) (workflowRun, error)
	DescribeWorkflowExecution(context.Context, string, string) (*workflowservice.DescribeWorkflowExecutionResponse, error)
}

type temporalWorkflowClient struct {
	client.Client
}

func (c temporalWorkflowClient) ExecuteWorkflow(ctx context.Context, opts client.StartWorkflowOptions, workflow any, args ...any) (workflowRun, error) {
	return c.Client.ExecuteWorkflow(ctx, opts, workflow, args...)
}

func submitterForWorkflowClient(c workflowClient, taskQueue string) rlm.SubmitFunc {
	return func(ctx context.Context, req rlm.AgentRequest) (rlm.AgentResult, error) {
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
		workflowID := run.GetID()
		if workflowID == "" {
			workflowID = opts.ID
		}
		runID := run.GetRunID()
		events.Emit(ctx, workerEvent(events.KindStarted, workflowID, runID, "submitted", ""))

		getCtx, cancel := context.WithCancel(ctx)
		defer cancel()
		type workflowResult struct {
			result rlm.AgentResult
			err    error
		}
		resultCh := make(chan workflowResult, 1)
		go func() {
			var res rlm.AgentResult
			err := run.Get(getCtx, &res)
			resultCh <- workflowResult{result: res, err: err}
		}()

		ticker := time.NewTicker(workflowPollInterval)
		defer ticker.Stop()
		for {
			select {
			case out := <-resultCh:
				if out.err != nil {
					status := "failed"
					if errors.Is(out.err, context.Canceled) {
						status = "canceled"
					}
					events.Emit(ctx, workerEvent(events.KindFailed, workflowID, runID, status, pmlogging.RedactText(ctx, out.err.Error())))
					return rlm.AgentResult{}, out.err
				}
				events.Emit(ctx, workerEvent(events.KindCompleted, workflowID, runID, "success", ""))
				return out.result, nil
			case <-ticker.C:
				_, describeErr := c.DescribeWorkflowExecution(ctx, workflowID, runID)
				message := ""
				if describeErr != nil && ctx.Err() == nil {
					message = pmlogging.RedactText(ctx, describeErr.Error())
				}
				events.Emit(ctx, workerEvent(events.KindProgress, workflowID, runID, "polling", message))
			case <-ctx.Done():
				cancel()
				events.Emit(ctx, workerEvent(events.KindFailed, workflowID, runID, "canceled", pmlogging.RedactText(ctx, ctx.Err().Error())))
				return rlm.AgentResult{}, ctx.Err()
			}
		}
	}
}

func workerEvent(kind events.Kind, workflowID, runID, status, message string) events.Event {
	return events.Event{
		Kind:    kind,
		Scope:   events.ScopeWorker,
		RunID:   workflowID,
		Status:  status,
		Message: message,
		Attrs: map[string]string{
			"workflow_id": workflowID,
			"run_id":      runID,
		},
	}
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
