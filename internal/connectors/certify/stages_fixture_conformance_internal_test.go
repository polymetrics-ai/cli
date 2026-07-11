package certify

import "testing"

func TestFixtureConformanceRunsForTwentyBundle(t *testing.T) {
	rc := &runContext{opts: Options{Connector: "twenty"}}
	var rep Report

	if err := stageFixtureConformance(rc, &rep); err != nil {
		t.Fatalf("stageFixtureConformance returned error: %v", err)
	}
	if len(rep.Stages) != 1 {
		t.Fatalf("len(rep.Stages) = %d, want 1", len(rep.Stages))
	}
	stage := rep.Stages[0]
	if stage.Name != "fixture_conformance" {
		t.Fatalf("stage.Name = %q, want fixture_conformance", stage.Name)
	}
	if !stage.Passed {
		t.Fatalf("fixture_conformance should run and pass for Twenty bundle fixtures, got error %q", stage.Error)
	}
	if stage.Error != "" {
		t.Fatalf("fixture_conformance Error = %q, want empty", stage.Error)
	}
}

func TestFixtureConformanceFailureFailsReport(t *testing.T) {
	failed := []StageResult{{Name: "fixture_conformance", Passed: false, Error: "fixture_conformance: conformance failed: write_request_shape:boom"}}
	if allStagesPassed(failed) {
		t.Fatalf("fixture_conformance conformance failure must fail the report")
	}

	skipped := []StageResult{{Name: "fixture_conformance", Passed: false, Error: noDefsBundleReason + " \"sample\""}}
	if !allStagesPassed(skipped) {
		t.Fatalf("fixture_conformance missing-fixture skip should not fail the report")
	}
}
