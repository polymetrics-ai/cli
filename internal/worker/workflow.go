package worker

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"polymetrics.ai/internal/rlm"
)

// RemoteRLMWorkflow runs one RLM agent job as a single durable activity. On
// cancellation it reaps the container via a disconnected-context cleanup activity
// so a cancelled job never leaks a running container.
func RemoteRLMWorkflow(ctx workflow.Context, req rlm.AgentRequest) (rlm.AgentResult, error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout:    30 * time.Minute,
		ScheduleToCloseTimeout: 45 * time.Minute,
		HeartbeatTimeout:       60 * time.Second,
		WaitForCancellation:    true,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts:        3,
			BackoffCoefficient:     2.0,
			MaximumInterval:        time.Minute,
			NonRetryableErrorTypes: nonRetryableErrorTypes(),
		},
	}
	actCtx := workflow.WithActivityOptions(ctx, ao)

	var acts *PodmanActivities // method expression → resolves the registered activity name
	var res rlm.AgentResult
	err := workflow.ExecuteActivity(actCtx, acts.RunPodman, req).Get(actCtx, &res)
	if err != nil {
		if temporal.IsCanceledError(err) {
			cleanup(ctx, req.Fingerprint)
		}
		return rlm.AgentResult{}, err
	}
	return res, nil
}

func cleanup(ctx workflow.Context, fingerprint string) {
	dctx, _ := workflow.NewDisconnectedContext(ctx)
	dctx = workflow.WithActivityOptions(dctx, workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
	})
	var acts *PodmanActivities
	_ = workflow.ExecuteActivity(dctx, acts.Cleanup, fingerprint).Get(dctx, nil)
}
