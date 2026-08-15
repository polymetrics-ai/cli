package certify

import (
	"context"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
)

func TestFullWriteSweepFailsForDeliberatelyBrokenDeclaredAction(t *testing.T) {
	assertBrokenWriteActionFailsCertification(t, "create_label")
}

func TestFullWriteSweepFailsForDeliberatelyBrokenPreviouslyBlockedAction(t *testing.T) {
	assertBrokenWriteActionFailsCertification(t, "update_issue")
}

func TestFullWriteSweepExercisesAll607DeclaredGitHubWriteActions(t *testing.T) {
	productionProbe, err := certificationWriteActionProbe("github")
	if err != nil {
		t.Fatal(err)
	}
	probed := make(map[string]int)
	rc := &runContext{
		ctx:  context.Background(),
		opts: Options{Connector: "github", Full: true, Write: true},
		writeActionProbe: func(ctx context.Context, connector, action string) error {
			if connector != "github" {
				t.Fatalf("probe connector = %q, want github", connector)
			}
			probed[action]++
			return productionProbe(ctx, connector, action)
		},
	}
	report := Report{Passed: true}
	if err := stageWriteSweepAllPairings(rc, &report); err != nil {
		t.Fatalf("stageWriteSweepAllPairings() error = %v", err)
	}
	if got := len(probed); got != 607 {
		t.Fatalf("probed actions = %d, want all 607 declared GitHub write actions", got)
	}
	if got := len(report.Capabilities.WriteActions); got != 607 {
		t.Fatalf("reported write actions = %d, want 607", got)
	}
	for action, calls := range probed {
		if calls != 1 {
			t.Fatalf("probe calls for %q = %d, want exactly 1", action, calls)
		}
		if result := report.Capabilities.WriteActions[action]; result.Result != "pass" {
			t.Fatalf("action %q result = %+v, want pass", action, result)
		}
	}
	previouslyBlocked := report.Capabilities.WriteActions["update_issue"]
	if !strings.Contains(previouslyBlocked.Reason, "provider mutation not run") {
		t.Fatalf("update_issue result = %+v, want honest non-live reason", previouslyBlocked)
	}
	curatedLifecycleCreates := 0
	for _, result := range report.Capabilities.WriteActions {
		if result.Cleanup != "" && result.Verify != "" {
			curatedLifecycleCreates++
		}
	}
	if curatedLifecycleCreates != 3 {
		t.Fatalf("curated live lifecycle create actions = %d, want 3", curatedLifecycleCreates)
	}
	t.Logf("github_write_coverage declared=607 definition_prepared=607 selected_live_pair_actions=2 remaining_non_live_actions=605 curated_lifecycle_create_actions=%d", curatedLifecycleCreates)
}

func assertBrokenWriteActionFailsCertification(t *testing.T, brokenAction string) {
	t.Helper()
	bundle, err := engine.Load(defs.FS, "github")
	if err != nil {
		t.Fatal(err)
	}
	broken := false
	for i := range bundle.Writes {
		if bundle.Writes[i].Name == brokenAction {
			bundle.Writes[i].RecordSchema = []byte(`{"type":`)
			broken = true
			break
		}
	}
	if !broken {
		t.Fatalf("GitHub definition omitted action %q", brokenAction)
	}
	const brokenMessage = "compile record_schema"
	var probed []string
	rc := &runContext{
		ctx:  context.Background(),
		opts: Options{Connector: "github", Full: true, Write: true},
		writeActionProbe: func(_ context.Context, connector, action string) error {
			if connector != "github" {
				t.Fatalf("probe connector = %q, want github", connector)
			}
			probed = append(probed, action)
			if action == brokenAction {
				return probeCertificationWriteBundle(context.Background(), bundle, action)
			}
			return nil
		},
	}
	report := Report{Passed: true}
	if err := stageWriteSweepAllPairings(rc, &report); err != nil {
		t.Fatalf("stageWriteSweepAllPairings() error = %v", err)
	}
	report.Passed = allStagesPassed(report.Stages)

	if report.Passed {
		t.Fatal("certification report passed after a deliberately broken write action")
	}
	if got := ExitCodeFor(report); got == 0 {
		t.Fatalf("ExitCodeFor(report) = %d, want non-zero certification failure", got)
	}
	result, ok := report.Capabilities.WriteActions[brokenAction]
	if !ok {
		t.Fatalf("report omitted broken action %q", brokenAction)
	}
	if result.Result != "fail" || !strings.Contains(result.Reason, brokenMessage) {
		t.Fatalf("broken action result = %+v, want fail naming deliberate break", result)
	}
	stage := findWriteSweepStage(t, report, brokenAction)
	if stage.Passed || !strings.Contains(stage.Error, brokenAction) || !strings.Contains(stage.Error, brokenMessage) {
		t.Fatalf("broken action stage = %+v, want named failure", stage)
	}
	if !containsCoverageAction(probed, brokenAction) {
		t.Fatalf("probe did not execute broken action %q; probed %d actions", brokenAction, len(probed))
	}

	intact := &runContext{ctx: context.Background(), opts: Options{Connector: "github", Full: true, Write: true}}
	intactReport := Report{Passed: true}
	if err := stageWriteSweepAllPairings(intact, &intactReport); err != nil {
		t.Fatalf("intact stageWriteSweepAllPairings() error = %v", err)
	}
	intactReport.Passed = allStagesPassed(intactReport.Stages)
	if !intactReport.Passed || ExitCodeFor(intactReport) != 0 {
		t.Fatalf("intact certification report failed: stages=%+v", failedStages(intactReport.Stages))
	}
	intactResult := intactReport.Capabilities.WriteActions[brokenAction]
	if intactResult.Result != "pass" {
		t.Fatalf("intact action %q result = %+v, want pass", brokenAction, intactResult)
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

func failedStages(stages []StageResult) []StageResult {
	failed := make([]StageResult, 0)
	for _, stage := range stages {
		if !stage.Passed {
			failed = append(failed, stage)
		}
	}
	return failed
}

func containsCoverageAction(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
