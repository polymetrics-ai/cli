package worker

import (
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/testsuite"

	"polymetrics.ai/internal/rlm"
)

func TestWorkflow_Success(t *testing.T) {
	ts := &testsuite.WorkflowTestSuite{}
	env := ts.NewTestWorkflowEnvironment()

	var acts *PodmanActivities
	env.RegisterActivity(acts.RunPodman)
	env.RegisterActivity(acts.Cleanup)

	want := rlm.AgentResult{JobDir: "/jobs/fp", RecordsScored: 5, RecordsRead: 5}
	env.OnActivity(acts.RunPodman, mock.Anything, mock.Anything).Return(want, nil)

	env.ExecuteWorkflow(RemoteRLMWorkflow, rlm.AgentRequest{Fingerprint: "fp", JobDir: "/jobs/fp"})

	if !env.IsWorkflowCompleted() {
		t.Fatal("workflow did not complete")
	}
	if err := env.GetWorkflowError(); err != nil {
		t.Fatalf("workflow error: %v", err)
	}
	var got rlm.AgentResult
	if err := env.GetWorkflowResult(&got); err != nil {
		t.Fatalf("result: %v", err)
	}
	if got.RecordsScored != 5 {
		t.Fatalf("RecordsScored = %d, want 5", got.RecordsScored)
	}
}

func TestWorkflow_NonRetryableFails(t *testing.T) {
	ts := &testsuite.WorkflowTestSuite{}
	env := ts.NewTestWorkflowEnvironment()

	var acts *PodmanActivities
	env.RegisterActivity(acts.RunPodman)
	env.RegisterActivity(acts.Cleanup)

	env.OnActivity(acts.RunPodman, mock.Anything, mock.Anything).Return(
		rlm.AgentResult{},
		temporal.NewNonRetryableApplicationError("oom", errOOMKilled, nil),
	).Once()

	env.ExecuteWorkflow(RemoteRLMWorkflow, rlm.AgentRequest{Fingerprint: "fp"})

	if !env.IsWorkflowCompleted() {
		t.Fatal("workflow did not complete")
	}
	if env.GetWorkflowError() == nil {
		t.Fatal("want workflow error for a non-retryable activity failure")
	}
}

func TestWorkflow_CancelTriggersCleanup(t *testing.T) {
	ts := &testsuite.WorkflowTestSuite{}
	env := ts.NewTestWorkflowEnvironment()

	var acts *PodmanActivities
	env.RegisterActivity(acts.RunPodman)
	env.RegisterActivity(acts.Cleanup)

	// RunPodman blocks (honoring cancellation); Cleanup must then be invoked.
	env.OnActivity(acts.RunPodman, mock.Anything, mock.Anything).
		Return(rlm.AgentResult{}, temporal.NewCanceledError())
	cleanupCalled := false
	env.OnActivity(acts.Cleanup, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) { cleanupCalled = true }).
		Return(nil)

	env.RegisterDelayedCallback(func() { env.CancelWorkflow() }, time.Millisecond)
	env.ExecuteWorkflow(RemoteRLMWorkflow, rlm.AgentRequest{Fingerprint: "fp"})

	if !env.IsWorkflowCompleted() {
		t.Fatal("workflow did not complete")
	}
	if !cleanupCalled {
		t.Fatal("cleanup activity was not invoked on cancellation")
	}
}
