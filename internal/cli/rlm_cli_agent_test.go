package cli

import (
	"context"
	"errors"
	"testing"

	"polymetrics.ai/internal/config"
	"polymetrics.ai/internal/rlm"
	"polymetrics.ai/internal/worker"
)

func TestBuildAgentAnalyzerProbesBeforeConstructingSubmitter(t *testing.T) {
	oldProbe := temporalProbe
	oldSubmitter := workerSubmitterForActivities
	t.Cleanup(func() {
		temporalProbe = oldProbe
		workerSubmitterForActivities = oldSubmitter
	})

	probeCalled := false
	temporalProbe = func(ctx context.Context, _ string) bool {
		probeCalled = true
		if _, ok := ctx.Deadline(); !ok {
			t.Fatal("temporal probe context missing deadline")
		}
		return false
	}
	workerSubmitterForActivities = func(context.Context, string, bool, *worker.PodmanActivities) (rlm.SubmitFunc, func() error, error) {
		t.Fatal("submitter was constructed before finite temporal probe succeeded")
		return nil, nil, nil
	}

	cfg := config.Config{
		Runtime:      config.RuntimeConfig{TemporalAddr: "127.0.0.1:7233"},
		ExplicitKeys: map[string]bool{"runtime.temporal_addr": true},
	}
	_, _, err := buildAgentAnalyzer(context.Background(), cfg, "")
	if !probeCalled {
		t.Fatal("temporal probe was not called")
	}
	if !errors.Is(err, rlm.ErrRemoteUnavailable) {
		t.Fatalf("buildAgentAnalyzer error = %v, want remote unavailable", err)
	}
}
