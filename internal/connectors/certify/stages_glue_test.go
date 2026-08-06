package certify_test

import (
	"context"
	"strings"
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

	r, driver := scriptedSampleRunner(t, certify.Options{
		Connector: "sample",
		Stream:    "customers",
		Limit:     50,
		SecretEnv: map[string]string{"token": "PM_SAMPLE_TOKEN"},
	})

	rep, err := r.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	driver.assertProtocol(t)
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

	r, driver := scriptedSampleRunner(t, certify.Options{
		Connector: "sample",
		Stream:    "customers",
		Limit:     50,
		SecretEnv: map[string]string{"token": "PM_SAMPLE_TOKEN"},
	})

	rep, err := r.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	driver.assertProtocol(t)
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

func TestGlueStagesFlowStatusRejectsMismatchedIdentity(t *testing.T) {
	t.Setenv("PM_SAMPLE_TOKEN", "sample-cert-token")

	r, driver := scriptedSampleRunner(t, certify.Options{
		Connector: "sample",
		Stream:    "customers",
		Limit:     50,
		SecretEnv: map[string]string{"token": "PM_SAMPLE_TOKEN"},
	})
	driver.statusFlow = "cert_flow_other"

	rep, err := r.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	driver.assertProtocol(t)
	if rep.Passed {
		t.Fatal("Report.Passed = true, want false after mismatched flow status identity")
	}
	status := mustStage(t, rep, "flow_status")
	if status.Passed {
		t.Fatalf("flow_status stage Passed = true, want false: %+v", status)
	}
	if !strings.Contains(status.Error, `flow="cert_flow_other"`) {
		t.Errorf("flow_status error = %q, want mismatched flow identity", status.Error)
	}
}

func TestGlueStagesScheduleListRejectsMismatchedDefinition(t *testing.T) {
	for _, tt := range []struct {
		name   string
		mutate func(*scriptedCLI)
		want   string
	}{
		{
			name: "cron",
			mutate: func(driver *scriptedCLI) {
				driver.scheduleCron = "0 4 * * *"
			},
			want: `cron="0 4 * * *"`,
		},
		{
			name: "flow",
			mutate: func(driver *scriptedCLI) {
				driver.scheduleFlow = "cert_flow_other"
			},
			want: `flow="cert_flow_other"`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("PM_SAMPLE_TOKEN", "sample-cert-token")

			r, driver := scriptedSampleRunner(t, certify.Options{
				Connector: "sample",
				Stream:    "customers",
				Limit:     50,
				SecretEnv: map[string]string{"token": "PM_SAMPLE_TOKEN"},
			})
			tt.mutate(driver)

			rep, err := r.Run(context.Background())
			if err != nil {
				t.Fatalf("Run() error = %v", err)
			}
			driver.assertProtocol(t)
			if rep.Passed {
				t.Fatal("Report.Passed = true, want false after mismatched schedule definition")
			}
			list := mustStage(t, rep, "schedule_list")
			if list.Passed {
				t.Fatalf("schedule_list stage Passed = true, want false: %+v", list)
			}
			if !strings.Contains(list.Error, tt.want) {
				t.Errorf("schedule_list error = %q, want mismatched %s", list.Error, tt.name)
			}
		})
	}
}

func TestGlueStagesScheduleInstallRejectsMismatchedCommand(t *testing.T) {
	for _, tt := range []struct {
		name   string
		mutate func(*scriptedCLI)
		want   string
	}{
		{
			name: "cron",
			mutate: func(driver *scriptedCLI) {
				driver.installedScheduleCron = "0 4 * * *"
			},
			want: "cron",
		},
		{
			name: "flow",
			mutate: func(driver *scriptedCLI) {
				driver.installedScheduleFlow = "cert_flow_other"
			},
			want: "expected flow run payload",
		},
		{
			name: "sentinel delimiter",
			mutate: func(driver *scriptedCLI) {
				driver.installedScheduleGap = ""
			},
			want: "sentinel must be preceded by whitespace",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("PM_SAMPLE_TOKEN", "sample-cert-token")

			r, driver := scriptedSampleRunner(t, certify.Options{
				Connector: "sample",
				Stream:    "customers",
				Limit:     50,
				SecretEnv: map[string]string{"token": "PM_SAMPLE_TOKEN"},
			})
			tt.mutate(driver)

			rep, err := r.Run(context.Background())
			if err != nil {
				t.Fatalf("Run() error = %v", err)
			}
			driver.assertProtocol(t)
			if rep.Passed {
				t.Fatal("Report.Passed = true, want false after mismatched schedule command")
			}
			install := mustStage(t, rep, "schedule_install")
			if install.Passed {
				t.Fatalf("schedule_install stage Passed = true, want false: %+v", install)
			}
			if !strings.Contains(install.Error, tt.want) {
				t.Errorf("schedule_install error = %q, want mismatched %s", install.Error, tt.name)
			}
		})
	}
}

func TestGlueStagesFlowStepsRejectInvalidResults(t *testing.T) {
	for _, tt := range []struct {
		name   string
		stage  string
		mutate func(*scriptedCLI)
		want   string
	}{
		{
			name:  "run non-object entries",
			stage: "flow_run",
			mutate: func(driver *scriptedCLI) {
				driver.flowRunSteps = []any{nil, nil}
			},
			want: "step is not an object",
		},
		{
			name:  "run duplicate step",
			stage: "flow_run",
			mutate: func(driver *scriptedCLI) {
				driver.flowRunSteps = []any{
					map[string]any{"id": "cert_sync", "status": "ok"},
					map[string]any{"id": "cert_sync", "status": "ok"},
				}
			},
			want: "appears more than once",
		},
		{
			name:  "status non-object entries",
			stage: "flow_status",
			mutate: func(driver *scriptedCLI) {
				driver.flowStatusSteps = []any{nil, nil}
			},
			want: "step is not an object",
		},
		{
			name:  "status duplicate step",
			stage: "flow_status",
			mutate: func(driver *scriptedCLI) {
				driver.flowStatusSteps = []any{
					map[string]any{"id": "cert_sync", "status": "success"},
					map[string]any{"id": "cert_sync", "status": "success"},
				}
			},
			want: "appears more than once",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("PM_SAMPLE_TOKEN", "sample-cert-token")

			r, driver := scriptedSampleRunner(t, certify.Options{
				Connector: "sample",
				Stream:    "customers",
				Limit:     50,
				SecretEnv: map[string]string{"token": "PM_SAMPLE_TOKEN"},
			})
			tt.mutate(driver)

			rep, err := r.Run(context.Background())
			if err != nil {
				t.Fatalf("Run() error = %v", err)
			}
			driver.assertProtocol(t)
			if rep.Passed {
				t.Fatal("Report.Passed = true, want false after invalid flow step results")
			}
			stage := mustStage(t, rep, tt.stage)
			if stage.Passed {
				t.Fatalf("%s stage Passed = true, want false: %+v", tt.stage, stage)
			}
			if !strings.Contains(stage.Error, tt.want) {
				t.Errorf("%s error = %q, want %s", tt.stage, stage.Error, tt.want)
			}
		})
	}
}

// TestGlueStagesScheduleRoundtripLeavesNoResidue proves the harness snapshots
// the (redirected, ephemeral) crontab before create/install and asserts it is
// byte-identical after remove, per design §D "remove -> assert sentinel
// absent AND crontab byte-identical to snapshot (residue = leaked_schedule)".
func TestGlueStagesScheduleRoundtripLeavesNoResidue(t *testing.T) {
	t.Setenv("PM_SAMPLE_TOKEN", "sample-cert-token")

	r, driver := scriptedSampleRunner(t, certify.Options{
		Connector: "sample",
		Stream:    "customers",
		Limit:     50,
		SecretEnv: map[string]string{"token": "PM_SAMPLE_TOKEN"},
	})

	rep, err := r.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	driver.assertProtocol(t)

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

	r, driver := scriptedSampleRunner(t, certify.Options{
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
	driver.assertProtocol(t)
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

	r, driver := scriptedSampleRunner(t, certify.Options{
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
	driver.assertProtocol(t)

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
