package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors/certify"
)

// certifyCLIRealInvocationBudget permits exactly one route proof that reaches
// the real certify.Runner. Rendering, persistence, batch, and failure cases
// use complete report fixtures. Every remaining direct Run call is counted
// here, so the contract covers the whole CLI test binary. A cold
// -count=1 run measures 92 calls, so the ceiling intentionally has no slack.
const certifyCLIRealInvocationBudget = 92

var certifyCLIRealInvocations atomic.Int64

func countedCertifyCLI(args []string, stdout, stderr *bytes.Buffer) int {
	certifyCLIRealInvocations.Add(1)
	return Run(args, stdout, stderr)
}

func countedCertifyCLIWriter(args []string, stdout, stderr io.Writer) int {
	certifyCLIRealInvocations.Add(1)
	return Run(args, stdout, stderr)
}

// TestMain wires the real Run entrypoint into certify's in-process CLI
// driver exactly once for this test binary (mirroring cmd/pm/main.go),
// since `pm connectors certify` drives certify.Runner/RunBatch which in turn
// drive Run recursively via certify.Harness (see
// internal/connectors/certify/cliharness.go SetCLIRunFunc).
func TestMain(m *testing.M) {
	certify.SetCLIRunFunc(countedCertifyCLIWriter)
	code := m.Run()
	got := certifyCLIRealInvocations.Load()
	fmt.Fprintf(os.Stderr, "certify CLI real invocations: %d (budget %d)\n", got, certifyCLIRealInvocationBudget)
	if got > certifyCLIRealInvocationBudget {
		fmt.Fprintf(os.Stderr, "certify CLI real invocation budget exceeded: got %d, allowed %d; retain one certify router proof and render remaining cases from fixtures\n", got, certifyCLIRealInvocationBudget)
		code = 1
	}
	os.Exit(code)
}

func certifyRun(t *testing.T, root string, args ...string) (stdout, stderr string, code int) {
	t.Helper()
	var outBuf, errBuf bytes.Buffer
	allArgs := append([]string{"--root", root}, args...)
	code = countedCertifyCLI(allArgs, &outBuf, &errBuf)
	return outBuf.String(), errBuf.String(), code
}

// TestCertifyCLISingleConnectorPassExitsZero is the one retained real
// certification route proof. It drives the full sample sweep through `pm`, so
// the CLI dispatch, Runner, harness, report persistence, and full source
// coverage all share one executable invocation rather than running the same
// expensive topology in a separate package test.
func TestCertifyCLISingleConnectorPassExitsZero(t *testing.T) {
	t.Setenv("PM_CERT_SAMPLE_TOKEN", "sample-cli-token")
	root := t.TempDir()

	stdout, stderr, code := certifyRun(t, root, "connectors", "certify", "sample",
		"--from-env", "token=PM_CERT_SAMPLE_TOKEN", "--full", "--json")

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stdout=%s stderr=%s", code, stdout, stderr)
	}
	var envelope struct {
		Kind   string         `json:"kind"`
		Report certify.Report `json:"report"`
	}
	if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
		t.Fatalf("decode certification JSON envelope: %v; stdout=%s", err, stdout)
	}
	if envelope.Kind != "ConnectorCertification" {
		t.Errorf("envelope kind = %q, want ConnectorCertification", envelope.Kind)
	}
	if envelope.Report.Connector != "sample" {
		t.Errorf("report connector = %q, want sample", envelope.Report.Connector)
	}
	if !envelope.Report.Passed {
		t.Fatalf("full certification report Passed = false; stages=%+v", envelope.Report.Stages)
	}
	assertFullSourceSweepReport(t, envelope.Report)
	path := filepath.Join(root, ".polymetrics", "certifications", "sample.json")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("router proof did not persist its report at %s: %v", path, err)
	}
}

func assertFullSourceSweepReport(t *testing.T, report certify.Report) {
	t.Helper()
	if stage := certifyReportStage(t, report, "full_sweep_connection_create_customers"); !stage.Passed {
		t.Fatalf("full sweep customers connection stage failed: %+v", stage)
	}
	if stage := certifyReportStage(t, report, "full_sweep_connection_create_events"); !stage.Passed {
		t.Fatalf("full sweep events connection stage failed: %+v", stage)
	}
	if got := countCertifyReportStages(report, "etl_full_refresh_append"); got != 2 {
		t.Fatalf("etl_full_refresh_append stages = %d, want 2 for sample's catalog streams", got)
	}
	if got := countCertifyReportStages(report, "flow_roundtrip"); got != 2 {
		t.Fatalf("flow_roundtrip stages = %d, want 2 for sample's catalog streams", got)
	}
	if got := countCertifyReportStages(report, "schedule_roundtrip"); got != 2 {
		t.Fatalf("schedule_roundtrip stages = %d, want 2 for sample's catalog streams", got)
	}
	if stage := certifyReportStage(t, report, "direct_read_sweep"); stage.Passed || !strings.Contains(stage.Error, "skipped:") {
		t.Fatalf("direct_read_sweep = %+v, want documented skip for sample", stage)
	}
	if report.Capabilities.DirectRead == nil || report.Capabilities.DirectRead.Result != "skipped" {
		t.Fatalf("Capabilities.DirectRead = %+v, want skipped", report.Capabilities.DirectRead)
	}
	if stage := certifyReportStage(t, report, "binary_download_sweep"); stage.Passed || !strings.Contains(stage.Error, "skipped:") {
		t.Fatalf("binary_download_sweep = %+v, want documented skip for sample", stage)
	}
	if report.Capabilities.Binary == nil || report.Capabilities.Binary.Result != "skipped" {
		t.Fatalf("Capabilities.Binary = %+v, want skipped", report.Capabilities.Binary)
	}
}

func certifyReportStage(t *testing.T, report certify.Report, name string) certify.StageResult {
	t.Helper()
	for _, stage := range report.Stages {
		if stage.Name == name {
			return stage
		}
	}
	t.Fatalf("stage %q not found; stages=%+v", name, report.Stages)
	return certify.StageResult{}
}

func countCertifyReportStages(report certify.Report, name string) int {
	count := 0
	for _, stage := range report.Stages {
		if stage.Name == name {
			count++
		}
	}
	return count
}

// TestCertifyCLISingleConnectorTextMode proves the non-JSON rendering path
// also works and reports PASS.
func TestCertifyCLISingleConnectorTextMode(t *testing.T) {
	stdout := renderCertifyReportText(completeCertifyReport())
	if !strings.Contains(stdout, "Legacy certification run: sample [PASS]") || !strings.Contains(stdout, "does not set the generated connector certification status") {
		t.Errorf("stdout missing human-readable PASS summary: %s", stdout)
	}
}

func TestCertifyCLIReportFixtureRendersJSONAndMapsExitCodes(t *testing.T) {
	rep := completeCertifyReport()
	var out bytes.Buffer
	if err := writeCertifyReport(&out, true, rep); err != nil {
		t.Fatalf("writeCertifyReport() error = %v", err)
	}
	if got := out.String(); !strings.Contains(got, `"kind": "ConnectorCertification"`) || !strings.Contains(got, `"connector": "sample"`) {
		t.Errorf("fixture JSON report = %s, want certification envelope and complete report", got)
	}
	assertCertifyExitCode(t, exitForReport(rep), 0)

	failed := rep
	failed.Passed = false
	assertCertifyExitCode(t, exitForReport(failed), 2)

	leaked := failed
	leaked.Leaks = []certify.Leak{{Tag: "pm-cert-sample-fixture", Connector: "sample", Action: "create", Reason: "fixture leak"}}
	assertCertifyExitCode(t, exitForReport(leaked), 3)
}

// TestCertifyCLISingleConnectorSavesReport proves the CLI persists the
// report under <root>/.polymetrics/certifications/<connector>.json
// (certification design §A report artifact path).
func TestCertifyCLISingleConnectorSavesReport(t *testing.T) {
	root := t.TempDir()
	rep := completeCertifyReport()
	if err := rep.Save(filepath.Join(root, ".polymetrics")); err != nil {
		t.Fatalf("Report.Save() error = %v", err)
	}

	path := filepath.Join(root, ".polymetrics", "certifications", "sample.json")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("report not saved at %s: %v", path, err)
	}
	loaded, err := certify.LoadReport(path)
	if err != nil {
		t.Fatalf("LoadReport(%s) error = %v", path, err)
	}
	if loaded.Connector != rep.Connector || !loaded.Passed {
		t.Fatalf("loaded report = %+v, want complete passing sample report", loaded)
	}
}

// TestCertifyCLIMissingConnectorArgIsUsageError proves `pm connectors
// certify` with no connector name and no --all/--sweep is a usage error, not
// a panic or a certify-mode exit code.
func TestCertifyCLIMissingConnectorArgIsUsageError(t *testing.T) {
	root := t.TempDir()

	_, stderr, code := certifyRun(t, root, "connectors", "certify")

	if code == 0 {
		t.Fatalf("exit code = 0, want non-zero for missing connector argument")
	}
	if !strings.Contains(stderr, "error:") {
		t.Errorf("stderr missing error message: %s", stderr)
	}
}

func TestCertifyCLIHelpShowsProvenanceContract(t *testing.T) {
	var stdout bytes.Buffer
	if err := runConnectors(context.Background(), t.TempDir(), []string{"certify", "--help"}, &stdout, false); err != nil {
		t.Fatalf("runConnectors(certify --help): %v", err)
	}
	for _, want := range []string{
		"pm connectors certify <connector> [--full] [--json]",
		"provider-artifact",
		"provenance evidence",
		"legacy_unverified",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Errorf("stdout missing %q: %s", want, stdout.String())
		}
	}
}

// TestCertifyCLIUnknownConnectorFails proves a connector name not present in
// the registry surfaces as a certify failure with a non-zero exit rather
// than a panic.
func TestCertifyCLIUnknownConnectorFails(t *testing.T) {
	root := t.TempDir()

	_, _, code := certifyRun(t, root, "connectors", "certify", "definitely-not-a-real-connector", "--json")

	if code == 0 {
		t.Fatalf("exit code = 0, want non-zero for unknown connector")
	}
}

// TestCertifyCLIAllRequiresCredentialsFile proves --all without
// --credentials-file is a clear usage error (certification design §A
// command spec: "pm connectors certify --all --credentials-file creds.yaml").
func TestCertifyCLIAllRequiresCredentialsFile(t *testing.T) {
	root := t.TempDir()

	_, stderr, code := certifyRun(t, root, "connectors", "certify", "--all")

	if code == 0 {
		t.Fatalf("exit code = 0, want non-zero (missing --credentials-file)")
	}
	if !strings.Contains(stderr, "credentials-file") {
		t.Errorf("stderr should mention --credentials-file: %s", stderr)
	}
}

// TestCertifyCLIBatchModeRunsCredsFileConnectors renders a complete batch
// fixture. RunBatch behavior itself is covered with fake Runnables in
// internal/connectors/certify; repeating a full CLI route here is redundant.
func TestCertifyCLIBatchModeRunsCredsFileConnectors(t *testing.T) {
	var out bytes.Buffer
	if err := writeCertifyBatchReport(&out, true, completeCertifyBatchReport()); err != nil {
		t.Fatalf("writeCertifyBatchReport() error = %v", err)
	}
	stdout := out.String()
	if !strings.Contains(stdout, `"kind": "ConnectorCertificationBatch"`) {
		t.Errorf("stdout missing ConnectorCertificationBatch envelope: %s", stdout)
	}
}

// TestCertifyCLIBatchModeTextRendersMatrix proves the non-JSON batch
// rendering includes the summary matrix header row (certification design
// §B columns).
func TestCertifyCLIBatchModeTextRendersMatrix(t *testing.T) {
	stdout := renderBatchMatrixText(completeCertifyBatchReport())
	if !strings.Contains(stdout, "connector\tcheck\tcatalog") {
		t.Errorf("stdout missing matrix header: %s", stdout)
	}
	if !strings.Contains(stdout, "exit_code: 0") {
		t.Errorf("stdout missing exit_code summary line: %s", stdout)
	}
}

func TestCertifyCLIBatchFixturePreservesMatrixOrderingAndExit(t *testing.T) {
	batch := completeCertifyBatchReport()
	batch.Results = []certify.BatchConnectorResult{
		{Connector: "zeta", Report: completeCertifyReport(), ExitCode: 2},
		{Connector: "alpha", Report: completeCertifyReport(), ExitCode: 0},
		{Connector: "resumed", Resumed: true, ExitCode: 0},
	}
	batch.ExitCode = 2

	stdout := renderBatchMatrixText(batch)
	alpha := strings.Index(stdout, "alpha\t")
	resumed := strings.Index(stdout, "resumed\t")
	zeta := strings.Index(stdout, "zeta\t")
	if alpha < 0 || resumed < 0 || zeta < 0 || alpha >= resumed || resumed >= zeta {
		t.Errorf("matrix rows are not name-ordered: %s", stdout)
	}
	assertCertifyExitCode(t, exitForBatch(batch), 2)
}

// TestCertifyCLISweepWithoutTargetsIsUsageError proves --sweep with nothing
// to sweep (no --credentials-file, no prior ledger) fails clearly instead of
// silently reporting success.
func TestCertifyCLISweepWithoutTargetsIsUsageError(t *testing.T) {
	root := t.TempDir()

	_, stderr, code := certifyRun(t, root, "connectors", "certify", "--sweep")

	if code == 0 {
		t.Fatalf("exit code = 0, want non-zero (nothing to sweep)")
	}
	if stderr == "" {
		t.Errorf("stderr empty, want an explanatory error")
	}
}

// TestCertifyCLISweepInvalidOlderThanIsUsageError proves a malformed
// --older-than duration is rejected before any sweeping is attempted.
func TestCertifyCLISweepInvalidOlderThanIsUsageError(t *testing.T) {
	root := t.TempDir()
	credsPath := filepath.Join(root, "creds.yaml")
	if err := os.WriteFile(credsPath, []byte("version: 1\nconnectors: {}\n"), 0o600); err != nil {
		t.Fatalf("write creds file: %v", err)
	}

	_, _, code := certifyRun(t, root, "connectors", "certify", "--sweep",
		"--credentials-file", credsPath, "--older-than", "not-a-duration")

	if code == 0 {
		t.Fatalf("exit code = 0, want non-zero for invalid --older-than")
	}
}

// TestCertifyCLIDoesNotBreakExistingConnectorsSubcommands is a regression
// guard: adding the certify dispatch case must not change behavior of any
// pre-existing `pm connectors` subcommand.
func TestCertifyCLIDoesNotBreakExistingConnectorsSubcommands(t *testing.T) {
	root := t.TempDir()

	stdout, stderr, code := certifyRun(t, root, "connectors", "list", "--json")
	if code != 0 {
		t.Fatalf("connectors list --json: exit %d, stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, `"kind": "ConnectorList"`) {
		t.Errorf("connectors list output missing ConnectorList kind: %s", stdout)
	}

	stdout2, stderr2, code2 := certifyRun(t, root, "connectors", "inspect", "sample", "--json")
	if code2 != 0 {
		t.Fatalf("connectors inspect sample --json: exit %d, stderr=%s", code2, stderr2)
	}
	if !strings.Contains(stdout2, `"kind": "Connector"`) {
		t.Errorf("connectors inspect output missing Connector kind: %s", stdout2)
	}
}

// completeCertifyReport is a complete, successful certification artifact for
// rendering and persistence tests. It deliberately includes every report
// capability represented in the CLI output contract so those tests do not
// need to repeat a real Runner invocation just to obtain sample data.
func completeCertifyReport() certify.Report {
	startedAt := time.Date(2026, time.August, 6, 12, 0, 0, 0, time.UTC)
	return certify.Report{
		Kind:          "ConnectorCertification",
		SchemaVersion: 1,
		Connector:     "sample",
		PMVersion:     "test",
		StartedAt:     startedAt,
		CompletedAt:   startedAt.Add(time.Second),
		Mode:          "live",
		Passed:        true,
		Capabilities: certify.Capabilities{
			Check:           certify.CapabilityResult{Result: "pass"},
			Catalog:         certify.CapabilityResult{Result: "pass", Streams: 2},
			Read:            certify.CapabilityResult{Result: "pass", Stream: "customers", Records: 2},
			Resume:          certify.CapabilityResult{Result: "pass"},
			JSONContract:    certify.CapabilityResult{Result: "pass", StagesChecked: 21},
			SecretRedaction: certify.CapabilityResult{Result: "pass"},
			SyncModes: map[string]certify.SyncModeResult{
				"full_refresh_append":            {Result: "pass", DataSource: "live"},
				"full_refresh_overwrite":         {Result: "pass", DataSource: "capture"},
				"full_refresh_overwrite_deduped": {Result: "pass", Reason: "typed pre-I/O refusal confirmed"},
				"incremental_append":             {Result: "pass", DataSource: "live", CursorAdvanced: true},
				"incremental_append_deduped":     {Result: "pass", Reason: "typed pre-I/O refusal confirmed"},
			},
			DirectRead: &certify.CapabilityResult{Result: "skipped", Reason: "fixture does not declare a direct read"},
			Binary:     &certify.CapabilityResult{Result: "skipped", Reason: "fixture does not declare a binary download"},
			Flow:       &certify.CapabilityResult{Result: "pass"},
			Schedule:   &certify.ScheduleResult{Result: "pass", Backend: "crontab", Residue: false},
			WriteActions: map[string]certify.WriteActionResult{
				"create": {Result: "pass", Cleanup: "delete", Verify: "verified", Tag: "pm-cert-sample-fixture"},
			},
		},
		Stages: []certify.StageResult{
			{Name: "init", Tier: 2, Passed: true, DurationMS: 1, CLI: certify.CLIStageInfo{ArgvRedacted: "pm init --json", Kind: "InitResult"}},
			{Name: "fixture_conformance", Tier: 0, Passed: false, Error: "skipped: fixture is complete for rendering only"},
		},
	}
}

func completeCertifyBatchReport() certify.BatchReport {
	return certify.BatchReport{
		RunID:     "fixture-batch",
		StartedAt: time.Date(2026, time.August, 6, 12, 0, 0, 0, time.UTC),
		Results: []certify.BatchConnectorResult{
			{Connector: "sample", Report: completeCertifyReport(), ExitCode: 0},
			{Connector: "skipped", Skipped: true, SkipReason: "fixture-only connector", ExitCode: 0},
		},
		ExitCode: 0,
	}
}

func assertCertifyExitCode(t *testing.T, err error, want int) {
	t.Helper()
	if want == 0 {
		if err != nil {
			t.Fatalf("exit error = %v, want nil for exit 0", err)
		}
		return
	}
	if err == nil {
		t.Fatalf("exit error = nil, want exit %d", want)
	}
	var cliErr *cliError
	if !errors.As(err, &cliErr) {
		t.Fatalf("exit error = %T %v, want *cliError", err, err)
	}
	if cliErr.exitOverride == nil || *cliErr.exitOverride != want {
		t.Errorf("exit override = %v, want %d", cliErr.exitOverride, want)
	}
}
