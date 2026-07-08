package certify_test

import (
	"context"
	"os"
	"testing"

	"polymetrics.ai/internal/connectors/certify"
)

// TestSourceStagesAgainstSample drives certify.Runner.Run end-to-end against
// the built-in "sample" connector (certification design implementation-order
// step 2 / PLAN.md T-14 / SPEC.md §1.6): stages 0-11 in an ephemeral --root
// workdir mirroring the Makefile "smoke" recipe flags (Makefile:41), proving
// the source-stage pipeline without any CLI wiring (Runner drives cli.Run
// in-process via the harness built in T/B-12).
func TestSourceStagesAgainstSample(t *testing.T) {
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
	if rep.Connector != "sample" {
		t.Errorf("Report.Connector = %q, want sample", rep.Connector)
	}
	if rep.Kind != "ConnectorCertification" {
		t.Errorf("Report.Kind = %q, want ConnectorCertification", rep.Kind)
	}
	if rep.SchemaVersion != 1 {
		t.Errorf("Report.SchemaVersion = %d, want 1", rep.SchemaVersion)
	}
	if rep.StartedAt.IsZero() || rep.CompletedAt.IsZero() {
		t.Errorf("Report timestamps not set: started=%v completed=%v", rep.StartedAt, rep.CompletedAt)
	}
	if rep.CompletedAt.Before(rep.StartedAt) {
		t.Errorf("CompletedAt %v before StartedAt %v", rep.CompletedAt, rep.StartedAt)
	}

	// --- stage 0: preflight ---
	preflight := mustStage(t, rep, "preflight")
	if !preflight.Passed {
		t.Errorf("preflight stage failed: %+v", preflight)
	}

	// --- stage 1: fixture_conformance — sample has no defs bundle in wave0 ---
	fixture := mustStage(t, rep, "fixture_conformance")
	if fixture.Passed {
		t.Errorf("fixture_conformance stage: want skipped (Passed=false with skip semantics), got Passed=true: %+v", fixture)
	}
	if fixture.Error == "" || !containsAny(fixture.Error, "skip", "no defs bundle", "no bundle") {
		t.Errorf("fixture_conformance stage Error = %q, want a skip-with-reason mentioning no defs bundle", fixture.Error)
	}

	// --- stage 2: manual_json ---
	manual := mustStage(t, rep, "manual_json")
	if !manual.Passed {
		t.Errorf("manual_json stage failed: %+v", manual)
	}
	if manual.CLI.Kind != "Connector" {
		t.Errorf("manual_json stage CLI.Kind = %q, want Connector", manual.CLI.Kind)
	}
	if manual.CLI.ExitCode != 0 {
		t.Errorf("manual_json stage CLI.ExitCode = %d, want 0", manual.CLI.ExitCode)
	}

	// --- stage 3: credentials add/test ---
	credAdd := mustStage(t, rep, "credentials_add")
	if !credAdd.Passed || credAdd.CLI.Kind != "Credential" {
		t.Errorf("credentials_add stage = %+v, want Passed=true Kind=Credential", credAdd)
	}
	credTest := mustStage(t, rep, "credentials_test")
	if !credTest.Passed || credTest.CLI.Kind != "CredentialTest" {
		t.Errorf("credentials_test stage = %+v, want Passed=true Kind=CredentialTest", credTest)
	}

	// --- stage 4: catalog ---
	catalog := mustStage(t, rep, "catalog")
	if !catalog.Passed || catalog.CLI.Kind != "Catalog" {
		t.Errorf("catalog stage = %+v, want Passed=true Kind=Catalog", catalog)
	}
	if rep.Capabilities.Catalog.Streams < 1 {
		t.Errorf("Capabilities.Catalog.Streams = %d, want >=1", rep.Capabilities.Catalog.Streams)
	}

	// --- stage 5: etl_full_refresh_append (live) ---
	fullAppend := mustStage(t, rep, "etl_full_refresh_append")
	if !fullAppend.Passed {
		t.Errorf("etl_full_refresh_append stage failed: %+v", fullAppend)
	}
	fraMode, ok := rep.Capabilities.SyncModes["full_refresh_append"]
	if !ok {
		t.Fatalf("Capabilities.SyncModes missing full_refresh_append: %+v", rep.Capabilities.SyncModes)
	}
	if fraMode.Result != "pass" || fraMode.DataSource != "live" {
		t.Errorf("SyncModes[full_refresh_append] = %+v, want {pass live}", fraMode)
	}
	if rep.Capabilities.Read.Records <= 0 {
		t.Errorf("Capabilities.Read.Records = %d, want >0", rep.Capabilities.Read.Records)
	}

	// --- stage 6: etl_full_refresh_overwrite (capture replay) ---
	overwrite := mustStage(t, rep, "etl_full_refresh_overwrite")
	if !overwrite.Passed {
		t.Errorf("etl_full_refresh_overwrite stage failed: %+v", overwrite)
	}
	froMode, ok := rep.Capabilities.SyncModes["full_refresh_overwrite"]
	if !ok || froMode.Result != "pass" || froMode.DataSource != "capture" {
		t.Errorf("SyncModes[full_refresh_overwrite] = %+v, want {pass capture}", froMode)
	}

	// --- stage 7: etl_full_refresh_overwrite_deduped (capture replay, PK dedup) ---
	dedupOverwrite := mustStage(t, rep, "etl_full_refresh_overwrite_deduped")
	if !dedupOverwrite.Passed {
		t.Errorf("etl_full_refresh_overwrite_deduped stage failed: %+v", dedupOverwrite)
	}
	frodMode, ok := rep.Capabilities.SyncModes["full_refresh_overwrite_deduped"]
	if !ok || frodMode.Result != "pass" || frodMode.DataSource != "capture" {
		t.Errorf("SyncModes[full_refresh_overwrite_deduped] = %+v, want {pass capture}", frodMode)
	}

	// --- stage 8: etl_incremental_append (live) ---
	incAppend := mustStage(t, rep, "etl_incremental_append")
	if !incAppend.Passed {
		t.Errorf("etl_incremental_append stage failed: %+v", incAppend)
	}
	iaMode, ok := rep.Capabilities.SyncModes["incremental_append"]
	if !ok || iaMode.Result != "pass" || iaMode.DataSource != "live" || !iaMode.CursorAdvanced {
		t.Errorf("SyncModes[incremental_append] = %+v, want {pass live cursor_advanced=true}", iaMode)
	}

	// --- stage 9: resume (live run 2) ---
	resume := mustStage(t, rep, "resume")
	if !resume.Passed {
		t.Errorf("resume stage failed: %+v", resume)
	}
	if rep.Capabilities.Resume.Result != "pass" {
		t.Errorf("Capabilities.Resume.Result = %q, want pass", rep.Capabilities.Resume.Result)
	}

	// --- stage 10: etl_incremental_append_deduped (capture replay) ---
	incDedup := mustStage(t, rep, "etl_incremental_append_deduped")
	if !incDedup.Passed {
		t.Errorf("etl_incremental_append_deduped stage failed: %+v", incDedup)
	}
	iadMode, ok := rep.Capabilities.SyncModes["incremental_append_deduped"]
	if !ok || iadMode.Result != "pass" || iadMode.DataSource != "capture" {
		t.Errorf("SyncModes[incremental_append_deduped] = %+v, want {pass capture}", iadMode)
	}

	// --- stage 11: query_contract ---
	queryContract := mustStage(t, rep, "query_contract")
	if !queryContract.Passed || queryContract.CLI.Kind != "QueryResult" {
		t.Errorf("query_contract stage = %+v, want Passed=true Kind=QueryResult", queryContract)
	}

	// json_contract meta-stage: aggregates kind+exit assertions across all
	// stages (certification design §A stage 21).
	if rep.Capabilities.JSONContract.Result != "pass" {
		t.Errorf("Capabilities.JSONContract.Result = %q, want pass", rep.Capabilities.JSONContract.Result)
	}
	if rep.Capabilities.JSONContract.StagesChecked < 10 {
		t.Errorf("Capabilities.JSONContract.StagesChecked = %d, want >=10", rep.Capabilities.JSONContract.StagesChecked)
	}

	// secret_redaction: scans ALL captured stage output for the planted
	// PM_SAMPLE_TOKEN value (exact/base64/URL-encoded forms per
	// THREAT-MODEL §1/§7); must never appear.
	if rep.Capabilities.SecretRedaction.Result != "pass" {
		t.Errorf("Capabilities.SecretRedaction.Result = %q, want pass", rep.Capabilities.SecretRedaction.Result)
	}
	for _, stage := range rep.Stages {
		if hits := certify.ScanForSecrets(stage.CLI.ArgvRedacted, []string{"sample-cert-token"}); len(hits) != 0 {
			t.Errorf("stage %s argv leaked secret: %v (%q)", stage.Name, hits, stage.CLI.ArgvRedacted)
		}
	}

	// Every stage that actually invokes the CLI recorded an argv string and
	// exit code (json_contract meta-stage assertion, stages[].cli shape).
	// fixture_conformance is a documented skip-only exception (no CLI call
	// -- see noDefsBundleReason); flow_roundtrip/schedule_roundtrip
	// (stages_glue.go) are aggregate meta-stages that summarize their own
	// named sub-stages (flow_plan/flow_preview/flow_run/flow_status,
	// schedule_create/schedule_list/schedule_install/schedule_remove) rather
	// than issuing a CLI call of their own -- each of those sub-stages DOES
	// carry its own CLI.ArgvRedacted, checked implicitly since they are
	// themselves entries in rep.Stages.
	metaStagesWithoutDirectCLICall := map[string]bool{
		"fixture_conformance": true,
		"flow_roundtrip":      true,
		"schedule_roundtrip":  true,
		// write stages 12-17 (stages_write.go) are skipped entirely in
		// this test (Options.Write is false) — each records a documented
		// skip with no CLI call, exactly like fixture_conformance above.
		"write_plan_preview":       true,
		"write_create":             true,
		"write_verify":             true,
		"write_cleanup":            true,
		"cleanup_verify":           true,
		"approval_idempotency":     true,
		"write_sweep_all_pairings": true,
	}
	for _, stage := range rep.Stages {
		if metaStagesWithoutDirectCLICall[stage.Name] {
			continue
		}
		if stage.CLI.ArgvRedacted == "" {
			t.Errorf("stage %s: CLI.ArgvRedacted empty, want a recorded pm invocation", stage.Name)
		}
	}
}

// TestSourceStagesSabotageFailsNamedStage proves a deliberately-wrong
// expected envelope kind flips exactly the sabotaged stage to failed and the
// overall report to Passed=false, naming the failing stage (PLAN.md T-14 /
// TEST-PLAN.md §4 "sabotage test -> passed=false, failing stage named").
func TestSourceStagesSabotageFailsNamedStage(t *testing.T) {
	t.Setenv("PM_SAMPLE_TOKEN", "sample-cert-token")

	r := certify.NewRunner(certify.Options{
		Connector: "sample",
		Stream:    "customers",
		Limit:     50,
		SecretEnv: map[string]string{"token": "PM_SAMPLE_TOKEN"},
	})
	certify.SabotageExpectedKind(r, "catalog", "NotTheRealKind")

	rep, err := r.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if rep.Passed {
		t.Fatalf("Report.Passed = true, want false after sabotage")
	}

	catalog := mustStage(t, rep, "catalog")
	if catalog.Passed {
		t.Errorf("sabotaged catalog stage Passed = true, want false")
	}
	if catalog.Error == "" {
		t.Errorf("sabotaged catalog stage Error is empty, want a message naming the mismatch")
	}

	// Every other stage that ran before the sabotaged one should still be
	// unaffected (proves the sabotage is scoped, not a global break).
	preflight := mustStage(t, rep, "preflight")
	if !preflight.Passed {
		t.Errorf("preflight stage should be unaffected by catalog sabotage: %+v", preflight)
	}
}

// TestSourceStagesSecretLeakInStdoutFailsSecretRedactionNamingStage is the
// M2 regression test (SECURITY-REVIEW.md): finalizeSecretRedaction must scan
// EVERY stage's raw captured stdout, not just the already-redacted argv
// string. Before the fix, secret_redaction scanned only
// stage.CLI.ArgvRedacted, so a secret leaking into a stage's stdout (the
// realistic leak surface — e.g. an `etl run` verbose/error message) was
// silently missed and secret_redaction still reported "pass". The
// SabotageStdoutLeak seam plants a distinctive secret value into the
// etl_full_refresh_append stage's captured stdout (a live "etl run" stage,
// the highest-risk call per the security review) without touching that
// stage's own pass/fail outcome, so this test isolates the redaction
// capability's own scan completeness.
func TestSourceStagesSecretLeakInStdoutFailsSecretRedactionNamingStage(t *testing.T) {
	// The planted "leak" must be a value the harness is actually watching
	// for (Options.SecretEnv), not an arbitrary string — ScanForSecrets only
	// ever matches against the run's own known secret values, exactly as it
	// would need to for a REAL leaked credential to be caught.
	const knownSecret = "sample-cert-token"
	t.Setenv("PM_SAMPLE_TOKEN", knownSecret)

	r := certify.NewRunner(certify.Options{
		Connector: "sample",
		Stream:    "customers",
		Limit:     50,
		SecretEnv: map[string]string{"token": "PM_SAMPLE_TOKEN"},
	})
	certify.SabotageStdoutLeak(r, "etl_full_refresh_append", knownSecret)

	rep, err := r.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	// The sabotaged stage's OWN outcome (envelope kind/exit code) is
	// unaffected: only secret_redaction should notice the planted leak.
	fullAppend := mustStage(t, rep, "etl_full_refresh_append")
	if !fullAppend.Passed {
		t.Fatalf("etl_full_refresh_append stage Passed = false, want true (sabotage plants a stdout leak, not a stage failure): %+v", fullAppend)
	}

	if rep.Capabilities.SecretRedaction.Result != "fail" {
		t.Fatalf("Capabilities.SecretRedaction.Result = %q, want fail (a secret was planted in etl_full_refresh_append's stdout)", rep.Capabilities.SecretRedaction.Result)
	}
	if !containsAny(rep.Capabilities.SecretRedaction.Reason, "etl_full_refresh_append") {
		t.Errorf("Capabilities.SecretRedaction.Reason = %q, want it to name the offending stage etl_full_refresh_append", rep.Capabilities.SecretRedaction.Reason)
	}
	if rep.Passed {
		t.Errorf("Report.Passed = true, want false: secret_redaction failing should fail the overall report")
	}
}

// TestSourceStagesEphemeralWorkdirCleanedUp proves the Runner uses an
// ephemeral os.MkdirTemp root per run and does not leave it behind (unless
// Options.KeepWork is set) — certification design §A execution model.
func TestSourceStagesEphemeralWorkdirCleanedUp(t *testing.T) {
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
		t.Fatalf("Report.Passed = false, want true")
	}

	workdir := certify.LastWorkdir(r)
	if workdir == "" {
		t.Fatalf("LastWorkdir() = empty, want the ephemeral root used by Run")
	}
	if _, err := os.Stat(workdir); !os.IsNotExist(err) {
		t.Errorf("workdir %s still exists after Run() with KeepWork=false: err=%v", workdir, err)
	}
}

// mustStage looks up a named stage in rep.Stages or fails the test.
func mustStage(t *testing.T, rep certify.Report, name string) certify.StageResult {
	t.Helper()
	for _, s := range rep.Stages {
		if s.Name == name {
			return s
		}
	}
	t.Fatalf("stage %q not found in report; stages=%+v", name, rep.Stages)
	return certify.StageResult{}
}

func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if len(sub) == 0 {
			continue
		}
		if idx := indexOfSubstring(s, sub); idx >= 0 {
			return true
		}
	}
	return false
}

func indexOfSubstring(s, sub string) int {
	// Small local helper to avoid importing strings twice with different
	// aliasing concerns; behaves like strings.Index.
	n, m := len(s), len(sub)
	if m == 0 {
		return 0
	}
	for i := 0; i+m <= n; i++ {
		if s[i:i+m] == sub {
			return i
		}
	}
	return -1
}
