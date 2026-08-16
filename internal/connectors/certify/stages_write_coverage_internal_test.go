package certify

import (
	"context"
	"strings"
	"testing"
)

func TestFullWriteSweepRecordsPreparedOnlyActionsAsNotLive(t *testing.T) {
	rc := &runContext{
		ctx:  context.Background(),
		opts: Options{Connector: "github", Full: true, Write: true},
	}
	report := Report{Passed: true}
	if err := stageWriteSweepAllPairings(rc, &report); err != nil {
		t.Fatalf("stageWriteSweepAllPairings() error = %v", err)
	}
	if got := len(report.Capabilities.WriteActions); got != 607 {
		t.Fatalf("reported write actions = %d, want 607", got)
	}
	for action, result := range report.Capabilities.WriteActions {
		if result.Result != "not_live" {
			t.Fatalf("action %q result = %+v, want not_live without a provider mutation", action, result)
		}
		if result.Path == "" || result.Risk == "" {
			t.Fatalf("action %q result = %+v, want declared path and risk for the non-live boundary", action, result)
		}
	}
	previouslyBlocked := report.Capabilities.WriteActions["update_issue"]
	if !strings.Contains(previouslyBlocked.Reason, "provider mutation was not run") {
		t.Fatalf("update_issue result = %+v, want honest non-live reason", previouslyBlocked)
	}
	stage := findWriteSweepStage(t, report, "update_issue")
	if stage.Passed || !strings.HasPrefix(stage.Error, "not_live: ") {
		t.Fatalf("update_issue stage = %+v, want explicit non-pass not_live stage", stage)
	}
}

func findWriteSweepStage(t *testing.T, report Report, action string) StageResult {
	t.Helper()
	want := "write_sweep_" + action
	for _, stage := range report.Stages {
		if stage.Name == want {
			return stage
		}
	}
	t.Fatalf("report omitted stage %q", want)
	return StageResult{}
}
