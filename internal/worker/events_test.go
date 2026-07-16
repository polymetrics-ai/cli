package worker

import (
	"context"
	"errors"
	"testing"
	"time"

	workflowservice "go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"

	"polymetrics.ai/internal/events"
	"polymetrics.ai/internal/rlm"
)

func TestSubmitterEmitsPollingEventsAndPreservesCancellation(t *testing.T) {
	oldInterval := workflowPollInterval
	workflowPollInterval = time.Millisecond
	t.Cleanup(func() { workflowPollInterval = oldInterval })

	run := &fakeWorkflowRun{id: "rlm-fp", runID: "run-1"}
	fake := &fakeWorkflowClient{run: run, described: make(chan struct{}, 1)}
	submit := submitterForWorkflowClient(fake, TaskQueue)
	collector := events.NewCollector()
	ctx, cancel := context.WithCancel(events.WithEmitter(context.Background(), collector))
	defer cancel()

	done := make(chan error, 1)
	go func() {
		_, err := submit(ctx, rlm.AgentRequest{Fingerprint: "fp"})
		done <- err
	}()

	select {
	case <-fake.described:
		cancel()
	case <-time.After(time.Second):
		t.Fatal("DescribeWorkflowExecution was not called")
	}

	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("submit error = %v, want context.Canceled", err)
		}
	case <-time.After(time.Second):
		t.Fatal("submit did not honor context cancellation")
	}

	if fake.workflowID != "rlm-fp" {
		t.Fatalf("workflow ID = %q, want rlm-fp", fake.workflowID)
	}
	got := workerEventSequence(collector.Events())
	want := []string{
		"worker::started:submitted",
		"worker::progress:polling",
		"worker::failed:canceled",
	}
	if len(got) < len(want) {
		t.Fatalf("sequence too short: got %#v want prefix %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("sequence[%d] = %q, want %q\ngot %#v", i, got[i], want[i], got)
		}
	}
}

type fakeWorkflowClient struct {
	run        *fakeWorkflowRun
	workflowID string
	described  chan struct{}
}

func (f *fakeWorkflowClient) ExecuteWorkflow(_ context.Context, opts client.StartWorkflowOptions, _ any, _ ...any) (workflowRun, error) {
	f.workflowID = opts.ID
	return f.run, nil
}

func (f *fakeWorkflowClient) DescribeWorkflowExecution(ctx context.Context, workflowID, runID string) (*workflowservice.DescribeWorkflowExecutionResponse, error) {
	select {
	case f.described <- struct{}{}:
	default:
	}
	return &workflowservice.DescribeWorkflowExecutionResponse{}, ctx.Err()
}

type fakeWorkflowRun struct {
	id    string
	runID string
}

func (f *fakeWorkflowRun) Get(ctx context.Context, _ any) error {
	<-ctx.Done()
	return ctx.Err()
}

func (f *fakeWorkflowRun) GetID() string { return f.id }

func (f *fakeWorkflowRun) GetRunID() string { return f.runID }

func workerEventSequence(in []events.Event) []string {
	out := make([]string, 0, len(in))
	for _, ev := range in {
		out = append(out, string(ev.Scope)+":"+ev.StepID+":"+string(ev.Kind)+":"+ev.Status)
	}
	return out
}
