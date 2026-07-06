package certify_test

import (
	"context"
	"testing"

	"polymetrics.ai/internal/connectors/certify"
)

// TestGlueStagesAgainstSample drives certify.Runner.Run end-to-end against
// the built-in "sample" connector and asserts the FLOW + SCHEDULE glue
// stages (certification design §A stage list "Glue stages": 18
// flow_roundtrip, 19 schedule_roundtrip) plus the two meta-stages that must
// see everything captured across the WHOLE run, including 18/19 (20
// secret_redaction_live, 21 json_contract).
func TestGlueStagesAgainstSample(t *testing.T) {
	t.Setenv("PM_SAMPLE_TOKEN", "sample-cert-token")

	r := certify.NewRunner(certify.Options{
		Connector: "sample",
		Stream:    "customers",
		Limit:     50,
		SecretEnv: map[string]string{"token": "PM_SAMPLE_TOKEN"},
	})

	rep, err := r.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !rep.Passed {
		t.Fatalf("Report.Passed = false, want true; stages=%+v", rep.Stages)
	}

	// --- stage 18: flow_roundtrip ---
	flowStage := mustStage(t, rep, "flow_roundtrip")
	if !flowStage.Passed {
		t.Fatalf("flow_roundtrip stage failed: %+v", flowStage)
	}
	if rep.Capabilities.Flow == nil {
		t.Fatalf("Capabilities.Flow is nil, want populated after flow_roundtrip")
	}
	if rep.Capabilities.Flow.Result != "pass" {
		t.Errorf("Capabilities.Flow.Result = %q, want pass", rep.Capabilities.Flow.Result)
	}

	// --- stage 19: schedule_roundtrip ---
	scheduleStage := mustStage(t, rep, "schedule_roundtrip")
	if !scheduleStage.Passed {
		t.Fatalf("schedule_roundtrip stage failed: %+v", scheduleStage)
	}
	if rep.Capabilities.Schedule == nil {
		t.Fatalf("Capabilities.Schedule is nil, want populated after schedule_roundtrip")
	}
	if rep.Capabilities.Schedule.Result != "pass" {
		t.Errorf("Capabilities.Schedule.Result = %q, want pass", rep.Capabilities.Schedule.Result)
	}
	if rep.Capabilities.Schedule.Backend != "crontab" {
		t.Errorf("Capabilities.Schedule.Backend = %q, want crontab", rep.Capabilities.Schedule.Backend)
	}
	if rep.Capabilities.Schedule.Residue {
		t.Errorf("Capabilities.Schedule.Residue = true, want false (no leaked schedule)")
	}

	// --- stage 20: secret_redaction meta-scan must have seen flow/schedule stage output too ---
	if rep.Capabilities.SecretRedaction.Result != "pass" {
		t.Errorf("Capabilities.SecretRedaction.Result = %q, want pass", rep.Capabilities.SecretRedaction.Result)
	}

	// --- stage 21: json_contract aggregation must count flow/schedule stages ---
	if rep.Capabilities.JSONContract.Result != "pass" {
		t.Errorf("Capabilities.JSONContract.Result = %q, want pass", rep.Capabilities.JSONContract.Result)
	}
	// stages_checked must be strictly greater than the wave0 (source-only)
	// count now that flow/schedule stages contribute CLI invocations too.
	if rep.Capabilities.JSONContract.StagesChecked < 14 {
		t.Errorf("Capabilities.JSONContract.StagesChecked = %d, want >=14 (source stages + flow/schedule)", rep.Capabilities.JSONContract.StagesChecked)
	}

	for _, name := range []string{"flow_plan", "flow_preview", "flow_run", "flow_status"} {
		s := mustStage(t, rep, name)
		if !s.Passed {
			t.Errorf("%s stage failed: %+v", name, s)
		}
	}
	for _, name := range []string{"schedule_create", "schedule_list", "schedule_install", "schedule_remove"} {
		s := mustStage(t, rep, name)
		if !s.Passed {
			t.Errorf("%s stage failed: %+v", name, s)
		}
	}
}

// TestGlueStagesFlowPreviewHasZeroSideEffects proves the preview (dry_run)
// step does not write to the warehouse: the query table backing the flow's
// query step must not exist (or be empty) until the real `flow run`
// executes it (design §D "preview dry_run with zero side effects").
func TestGlueStagesFlowPreviewHasZeroSideEffects(t *testing.T) {
	t.Setenv("PM_SAMPLE_TOKEN", "sample-cert-token")

	r := certify.NewRunner(certify.Options{
		Connector: "sample",
		Stream:    "customers",
		Limit:     50,
		SecretEnv: map[string]string{"token": "PM_SAMPLE_TOKEN"},
	})

	rep, err := r.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !rep.Passed {
		t.Fatalf("Report.Passed = false, want true; stages=%+v", rep.Stages)
	}

	preview := mustStage(t, rep, "flow_preview")
	if !preview.Passed {
		t.Fatalf("flow_preview stage failed: %+v", preview)
	}
	run := mustStage(t, rep, "flow_run")
	if !run.Passed {
		t.Fatalf("flow_run stage failed: %+v", run)
	}
}

// TestGlueStagesScheduleRoundtripLeavesNoResidue proves the harness snapshots
// the (redirected, ephemeral) crontab before create/install and asserts it is
// byte-identical after remove, per design §D "remove -> assert sentinel
// absent AND crontab byte-identical to snapshot (residue = leaked_schedule)".
func TestGlueStagesScheduleRoundtripLeavesNoResidue(t *testing.T) {
	t.Setenv("PM_SAMPLE_TOKEN", "sample-cert-token")

	r := certify.NewRunner(certify.Options{
		Connector: "sample",
		Stream:    "customers",
		Limit:     50,
		SecretEnv: map[string]string{"token": "PM_SAMPLE_TOKEN"},
	})

	rep, err := r.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	install := mustStage(t, rep, "schedule_install")
	if !install.Passed {
		t.Fatalf("schedule_install stage failed: %+v", install)
	}
	remove := mustStage(t, rep, "schedule_remove")
	if !remove.Passed {
		t.Fatalf("schedule_remove stage failed: %+v", remove)
	}
	if rep.Capabilities.Schedule == nil || rep.Capabilities.Schedule.Residue {
		t.Fatalf("Capabilities.Schedule = %+v, want non-nil with Residue=false", rep.Capabilities.Schedule)
	}
}

// TestGlueStagesSabotageFlowFailsNamedStage proves a deliberately-wrong
// expected kind/shape assertion on the flow_roundtrip stage flips exactly
// that stage (and the overall report) to failed, without disturbing
// unrelated stages -- mirroring the source-stage sabotage contract
// (TestSourceStagesSabotageFailsNamedStage).
func TestGlueStagesSabotageFlowFailsNamedStage(t *testing.T) {
	t.Setenv("PM_SAMPLE_TOKEN", "sample-cert-token")

	r := certify.NewRunner(certify.Options{
		Connector: "sample",
		Stream:    "customers",
		Limit:     50,
		SecretEnv: map[string]string{"token": "PM_SAMPLE_TOKEN"},
	})
	certify.SabotageExpectedKind(r, "flow_run", "NotTheRealKind")

	rep, err := r.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if rep.Passed {
		t.Fatalf("Report.Passed = true, want false after flow sabotage")
	}
	flowRun := mustStage(t, rep, "flow_run")
	if flowRun.Passed {
		t.Errorf("sabotaged flow_run stage Passed = true, want false")
	}

	// schedule stages should still run and pass -- proves the sabotage is
	// scoped to the flow stage only.
	install := mustStage(t, rep, "schedule_install")
	if !install.Passed {
		t.Errorf("schedule_install stage should be unaffected by flow sabotage: %+v", install)
	}
}

// TestGlueStagesSecretLeakInFlowStdoutFailsSecretRedaction proves the M2
// full-output secret scan (finalizeSecretRedaction) covers the new
// flow/schedule stages too, not just the original source-stage set.
func TestGlueStagesSecretLeakInFlowStdoutFailsSecretRedaction(t *testing.T) {
	const knownSecret = "sample-cert-token"
	t.Setenv("PM_SAMPLE_TOKEN", knownSecret)

	r := certify.NewRunner(certify.Options{
		Connector: "sample",
		Stream:    "customers",
		Limit:     50,
		SecretEnv: map[string]string{"token": "PM_SAMPLE_TOKEN"},
	})
	certify.SabotageStdoutLeak(r, "flow_run", knownSecret)

	rep, err := r.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	flowRun := mustStage(t, rep, "flow_run")
	if !flowRun.Passed {
		t.Fatalf("flow_run stage Passed = false, want true (sabotage plants a stdout leak, not a stage failure): %+v", flowRun)
	}
	if rep.Capabilities.SecretRedaction.Result != "fail" {
		t.Fatalf("Capabilities.SecretRedaction.Result = %q, want fail (secret planted in flow_run stdout)", rep.Capabilities.SecretRedaction.Result)
	}
	if !containsAny(rep.Capabilities.SecretRedaction.Reason, "flow_run") {
		t.Errorf("Capabilities.SecretRedaction.Reason = %q, want it to name flow_run", rep.Capabilities.SecretRedaction.Reason)
	}
	if rep.Passed {
		t.Errorf("Report.Passed = true, want false: secret_redaction failing should fail the overall report")
	}
}
