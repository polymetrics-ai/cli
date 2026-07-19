package certify

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestReviewCorrectionPreviewFailureBlocksCreateAndLedger(t *testing.T) {
	oldRun := cliRun
	oldRunContext := cliRunContext
	cliRunContext = func(_ context.Context, args []string, stdout, stderr io.Writer, _ CLIInvocationOptions) int {
		return cliRun(args, stdout, stderr)
	}
	t.Cleanup(func() {
		cliRun = oldRun
		cliRunContext = oldRunContext
	})
	var runs atomic.Int64
	cliRun = func(args []string, stdout, _ io.Writer) int {
		if hasArgs(args, "reverse", "run") {
			runs.Add(1)
			fmt.Fprint(stdout, `{"kind":"ReverseRun","run":{"records_succeeded":1,"records_failed":0}}`)
		}
		return 0
	}

	ledger, err := NewLedger(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	rc := &runContext{
		ctx:     context.Background(),
		harness: NewHarness(t.TempDir()),
		opts:    Options{Connector: "sample", Write: true},
		root:    t.TempDir(),
		write: &writeContext{
			pairing:       WritePairing{Create: "create", Cleanup: "delete", VerifyStream: "records"},
			tag:           NewTag("sample", "12345678"),
			planID:        "plan-after-failed-preview",
			approvalToken: "approval-marker",
			ledger:        ledger,
		},
	}
	var rep Report
	if err := stageWriteCreate(rc, &rep); err != nil {
		t.Fatal(err)
	}
	if runs.Load() != 0 {
		t.Fatalf("reverse execution count=%d, want zero after failed/unvalidated preview", runs.Load())
	}
	entries, err := LoadLedger(filepath.Dir(ledger.Path()))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries.Uncleaned()) != 0 {
		t.Fatalf("write-ahead entries=%d, want zero when preview did not succeed", len(entries.Uncleaned()))
	}
}

func TestReviewCorrectionCleanupAndSweepRequirePreviewBeforeRun(t *testing.T) {
	oldRun := cliRun
	oldRunContext := cliRunContext
	cliRunContext = func(_ context.Context, args []string, stdout, stderr io.Writer, _ CLIInvocationOptions) int {
		return cliRun(args, stdout, stderr)
	}
	t.Cleanup(func() {
		cliRun = oldRun
		cliRunContext = oldRunContext
	})
	var calls []string
	cliRun = func(args []string, stdout, _ io.Writer) int {
		switch {
		case hasArgs(args, "reverse", "plan"):
			calls = append(calls, "plan")
			fmt.Fprint(stdout, "Created reverse plan cleanup-plan\nApproval token: approval-marker\n")
		case hasArgs(args, "reverse", "preview"):
			calls = append(calls, "preview")
			fmt.Fprint(stdout, `{"kind":"ReversePlanPreview","plan":{}}`)
		case hasArgs(args, "reverse", "run"):
			calls = append(calls, "run")
			fmt.Fprint(stdout, `{"kind":"ReverseRun","run":{"records_succeeded":1,"records_failed":0}}`)
		default:
			fmt.Fprint(stdout, `{"kind":"Credential"}`)
		}
		return 0
	}

	ledger, err := NewLedger(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	rc := &runContext{ctx: context.Background(), harness: NewHarness(t.TempDir()), root: t.TempDir(), opts: Options{Connector: "sample"}}
	wc := &writeContext{
		pairing:      WritePairing{Create: "create", Cleanup: "delete"},
		tag:          NewTag("sample", "12345678"),
		selfTest:     true,
		createPassed: true,
		ledger:       ledger,
	}
	var rep Report
	stageWriteCleanupSelfTest(rc, &rep, wc)
	assertPlanPreviewRunOrder(t, calls)

	calls = nil
	status := LedgerStatus{
		Tag:        NewTag("sample", "87654321"),
		Connector:  "sample",
		Action:     "create",
		EntityHint: "outbox_record",
		PlannedAt:  time.Now().Add(-48 * time.Hour),
	}
	ok, reason := sweepCleanOutboxRecord(NewHarness(t.TempDir()), status)
	if !ok {
		t.Fatalf("sweep cleanup failed: %s", reason)
	}
	assertPlanPreviewRunOrder(t, calls)
}

func TestReviewCorrectionFailedSweepPreviewBlocksRun(t *testing.T) {
	oldRun := cliRun
	oldRunContext := cliRunContext
	var runCalls atomic.Int64
	cliRunContext = func(_ context.Context, args []string, stdout, stderr io.Writer, _ CLIInvocationOptions) int {
		return cliRun(args, stdout, stderr)
	}
	t.Cleanup(func() {
		cliRun = oldRun
		cliRunContext = oldRunContext
	})
	cliRun = func(args []string, stdout, _ io.Writer) int {
		switch {
		case hasArgs(args, "reverse", "plan"):
			fmt.Fprint(stdout, "Created reverse plan cleanup-plan\nApproval token: approval-marker\n")
		case hasArgs(args, "reverse", "preview"):
			fmt.Fprint(stdout, `{"kind":"WrongPreview"}`)
		case hasArgs(args, "reverse", "run"):
			runCalls.Add(1)
			fmt.Fprint(stdout, `{"kind":"ReverseRun"}`)
		default:
			fmt.Fprint(stdout, `{"kind":"Credential"}`)
		}
		return 0
	}
	status := LedgerStatus{
		Tag:        "pm-cert-sample-87654321-1700000000",
		Connector:  "sample",
		Action:     "create",
		RunID:      "87654321",
		EntityHint: "outbox_record",
		PlannedAt:  time.Now().Add(-48 * time.Hour),
	}
	if ok, _ := sweepCleanOutboxRecord(NewHarness(t.TempDir()), status); ok {
		t.Fatal("sweep succeeded after a mismatched preview")
	}
	if runCalls.Load() != 0 {
		t.Fatalf("failed sweep preview reached reverse execution %d times", runCalls.Load())
	}
}

func TestReviewCorrectionSweepPreparationFailureBlocksPlanning(t *testing.T) {
	oldRun := cliRun
	oldRunContext := cliRunContext
	t.Cleanup(func() {
		cliRun = oldRun
		cliRunContext = oldRunContext
	})
	var planCalls atomic.Int64
	cliRun = func(args []string, stdout, _ io.Writer) int {
		if hasArgs(args, "reverse", "plan") {
			planCalls.Add(1)
		}
		if hasArgs(args, "credentials", "add") && flagValue(args, "--connector") == "outbox" {
			fmt.Fprint(stdout, `{"kind":"Credential"}`)
			return 0
		}
		return 1
	}
	cliRunContext = func(_ context.Context, args []string, stdout, stderr io.Writer, _ CLIInvocationOptions) int {
		return cliRun(args, stdout, stderr)
	}
	status := LedgerStatus{
		Tag: "pm-cert-sample-87654321-1700000000", Connector: "sample", Action: "create", CleanupAction: "delete",
		RunID: "87654321", EntityHint: "outbox_record", PlannedAt: time.Now().Add(-48 * time.Hour),
	}
	if ok, _ := sweepCleanOutboxRecord(NewHarness(t.TempDir()), status); ok {
		t.Fatal("sweep succeeded after source preparation failure")
	}
	if planCalls.Load() != 0 {
		t.Fatalf("source preparation failure reached reverse planning %d times", planCalls.Load())
	}
}

func TestReviewCorrectionSecretHitsAndArgvAreOpaque(t *testing.T) {
	marker := "approval-and-credential-marker"
	hits := ScanForSecrets("prefix "+marker+" suffix", []string{marker})
	if len(hits) == 0 {
		t.Fatal("expected planted marker detection")
	}
	if strings.Contains(fmt.Sprint(hits), marker) {
		t.Fatal("secret detection metadata copied the matched value")
	}
	argv := redactArgv([]string{"reverse", "run", "plan", "--approve", marker, "--json"}, nil)
	if strings.Contains(argv, marker) || !strings.Contains(argv, "--approve ***") {
		t.Fatalf("approval argv was not semantically redacted: %q", argv)
	}
	configArgv := redactArgv([]string{"credentials", "add", "x", "--config", "api_token=" + marker}, nil)
	if strings.Contains(configArgv, marker) || !strings.Contains(configArgv, "api_token=***") {
		t.Fatalf("secret-bearing config argv was not semantically redacted: %q", configArgv)
	}

	reason := redactSecretsInText("encoded marker: "+secretForms(marker)[1], []string{marker})
	rep := Report{Stages: []StageResult{{Error: reason, CLI: CLIStageInfo{ArgvRedacted: argv + " " + configArgv}}}}
	raw, err := json.Marshal(rep)
	if err != nil {
		t.Fatal(err)
	}
	for _, form := range secretForms(marker) {
		if form != "" && strings.Contains(string(raw), form) {
			t.Fatal("serialized report retained planted sensitive material")
		}
	}
}

func TestReviewCorrectionReportWritesAreRestrictiveAndNoFollow(t *testing.T) {
	dir := t.TempDir()
	rep := Report{Kind: "ConnectorCertification", SchemaVersion: 1, Connector: "sample", StartedAt: time.Now().UTC()}
	if err := rep.Save(dir); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, certificationsDirName, "sample.json")
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("report mode=%#o, want 0600", info.Mode().Perm())
	}

	dir = t.TempDir()
	certDir := filepath.Join(dir, certificationsDirName)
	if err := os.MkdirAll(certDir, 0o700); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(t.TempDir(), "outside.json")
	if err := os.WriteFile(target, []byte("unchanged"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(certDir, "sample.json")); err != nil {
		t.Fatal(err)
	}
	if err := rep.Save(dir); err == nil {
		t.Fatal("Save succeeded through a symlink, want rejection")
	}
	raw, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != "unchanged" {
		t.Fatal("Save modified symlink target")
	}
	if _, err := LoadReport(filepath.Join(certDir, "sample.json")); err == nil {
		t.Fatal("LoadReport followed a symlink")
	}
}

func TestReviewCorrectionCredsFileStrictBoundary(t *testing.T) {
	validHeader := "version: 1\nconnectors:\n  sample:\n    credential:\n      from_env:\n        token: PM_CERT_SAMPLE_TOKEN\n"
	cases := []struct {
		name string
		raw  string
	}{
		{"unknown field", "version: 1\nconnectorz: {}\n"},
		{"unsupported version", "version: 2\nconnectors:\n  sample: {}\n"},
		{"empty jobs", "version: 1\nconnectors: {}\n"},
		{"traversal connector", "version: 1\nconnectors:\n  ../sample: {}\n"},
		{"invalid env reference", strings.Replace(validHeader, "PM_CERT_SAMPLE_TOKEN", "bad-env-name", 1)},
		{"secret config injection", "version: 1\nconnectors:\n  github:\n    credential:\n      config:\n        token: planted-marker\n"},
		{"explicit parallel zero", "version: 1\ndefaults:\n  parallel: 0\nconnectors:\n  sample: {}\n"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "creds.yaml")
			if err := os.WriteFile(path, []byte(tc.raw), 0o600); err != nil {
				t.Fatal(err)
			}
			if _, err := LoadCredsFile(path); err == nil {
				t.Fatal("LoadCredsFile succeeded, want strict rejection")
			}
		})
	}

	t.Run("symlink", func(t *testing.T) {
		realPath := filepath.Join(t.TempDir(), "real.yaml")
		if err := os.WriteFile(realPath, []byte(validHeader), 0o600); err != nil {
			t.Fatal(err)
		}
		linkPath := filepath.Join(t.TempDir(), "creds.yaml")
		if err := os.Symlink(realPath, linkPath); err != nil {
			t.Fatal(err)
		}
		if _, err := LoadCredsFile(linkPath); err == nil {
			t.Fatal("LoadCredsFile followed a symlink")
		}
	})

	t.Run("oversize", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "creds.yaml")
		raw := append([]byte(validHeader), make([]byte, (1<<20)+1)...)
		if err := os.WriteFile(path, raw, 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := LoadCredsFile(path); err == nil {
			t.Fatal("LoadCredsFile accepted oversized input")
		}
	})
}

func TestReviewCorrectionLedgerLayoutAndAuthority(t *testing.T) {
	root := t.TempDir()
	ledger, err := NewLedger(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := ledger.RecordPlanned(LedgerEntry{Connector: "sample", Action: "create", Tag: NewTag("sample", "12345678")}); err != nil {
		t.Fatal(err)
	}
	if err := ledger.CopyTo(root, "sample"); err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(root, certificationsDirName, "ledger", "sample", ledgerFileName)
	if _, err := os.Stat(want); err != nil {
		t.Fatalf("durable sweep layout missing at %s: %v", want, err)
	}

	oldRun := cliRun
	oldRunContext := cliRunContext
	cliRunContext = func(_ context.Context, args []string, stdout, stderr io.Writer, _ CLIInvocationOptions) int {
		return cliRun(args, stdout, stderr)
	}
	t.Cleanup(func() {
		cliRun = oldRun
		cliRunContext = oldRunContext
	})
	var calls atomic.Int64
	cliRun = func(_ []string, _, _ io.Writer) int { calls.Add(1); return 1 }
	forged := LedgerStatus{Tag: "forged-tag", Connector: "github", Action: "create_label", EntityHint: "labels", PlannedAt: time.Now().Add(-48 * time.Hour)}
	if ok, _ := sweepCleanTag(NewHarness(t.TempDir()), forged); ok || calls.Load() != 0 {
		t.Fatalf("forged ledger cleanup reached effects: ok=%v calls=%d", ok, calls.Load())
	}

	linkRoot := t.TempDir()
	target := filepath.Join(t.TempDir(), ledgerFileName)
	if err := os.WriteFile(target, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(linkRoot, ledgerFileName)); err != nil {
		t.Fatal(err)
	}
	if _, err := NewLedger(linkRoot); err == nil {
		t.Fatal("NewLedger followed a symlink")
	}
	if _, err := LoadLedger(linkRoot); err == nil {
		t.Fatal("LoadLedger followed a symlink")
	}
}

func TestReviewCorrectionParallelSchedulesUseInvocationLocalCrontabs(t *testing.T) {
	oldRun := cliRun
	oldRunContext := cliRunContext
	cliRunContext = func(_ context.Context, args []string, stdout, stderr io.Writer, _ CLIInvocationOptions) int {
		return cliRun(args, stdout, stderr)
	}
	t.Cleanup(func() {
		cliRun = oldRun
		cliRunContext = oldRunContext
	})
	t.Setenv("PM_CRONTAB_FILE", "operator-crontab")

	rootA := t.TempDir()
	rootB := t.TempDir()
	aCreate := make(chan struct{})
	releaseA := make(chan struct{})
	bInstall := make(chan struct{})
	releaseB := make(chan struct{})
	var mismatch atomic.Int64
	var mu sync.Mutex
	names := map[string]string{}
	cliRun = func(args []string, stdout, _ io.Writer) int {
		root := flagValue(args, "--root")
		expected := filepath.Join(root, scheduleCrontabFileName)
		if hasArgs(args, "schedule", "create") {
			name := flagValue(args, "--name")
			mu.Lock()
			names[root] = name
			mu.Unlock()
			if root == rootA {
				close(aCreate)
				<-releaseA
			}
			fmt.Fprint(stdout, `{"kind":"Schedule"}`)
			return 0
		}
		if hasArgs(args, "schedule", "list") {
			mu.Lock()
			name := names[root]
			mu.Unlock()
			fmt.Fprintf(stdout, `{"kind":"ScheduleList","schedules":[{"name":%q}]}`, name)
			return 0
		}
		if hasArgs(args, "schedule", "install") {
			if root == rootB {
				close(bInstall)
				<-releaseB
			}
			name := args[2]
			_ = os.WriteFile(expected, []byte("0 3 * * * pm # pm-schedule-"+name+"\n"), 0o600)
			fmt.Fprint(stdout, `{"kind":"ScheduleInstall","backend":"crontab"}`)
			return 0
		}
		if hasArgs(args, "schedule", "remove") {
			_ = os.WriteFile(expected, nil, 0o600)
			fmt.Fprint(stdout, `{"kind":"ScheduleRemove"}`)
			return 0
		}
		return 1
	}
	cliRunContext = func(_ context.Context, args []string, stdout, stderr io.Writer, opts CLIInvocationOptions) int {
		root := flagValue(args, "--root")
		if opts.CrontabFile != filepath.Join(root, scheduleCrontabFileName) {
			mismatch.Add(1)
		}
		return cliRun(args, stdout, stderr)
	}

	run := func(root string, done chan<- struct{}) {
		rc := &runContext{ctx: context.Background(), harness: NewHarness(root), root: root, opts: Options{Connector: "sample"}}
		var rep Report
		_ = stageScheduleRoundtrip(rc, &rep)
		close(done)
	}
	doneA := make(chan struct{})
	doneB := make(chan struct{})
	go run(rootA, doneA)
	<-aCreate
	go run(rootB, doneB)
	<-bInstall
	close(releaseA)
	<-doneA
	close(releaseB)
	<-doneB
	if mismatch.Load() != 0 {
		t.Fatalf("parallel schedule invocations observed %d cross-run/operator crontab selections", mismatch.Load())
	}
	if got := os.Getenv("PM_CRONTAB_FILE"); got != "operator-crontab" {
		t.Fatalf("operator crontab env changed to %q", got)
	}
}

func TestReviewCorrectionCancelledRunnerHasNoEffects(t *testing.T) {
	oldRun := cliRun
	oldRunContext := cliRunContext
	cliRunContext = func(_ context.Context, args []string, stdout, stderr io.Writer, _ CLIInvocationOptions) int {
		return cliRun(args, stdout, stderr)
	}
	t.Cleanup(func() {
		cliRun = oldRun
		cliRunContext = oldRunContext
	})
	var calls atomic.Int64
	cliRun = func(_ []string, _, _ io.Writer) int { calls.Add(1); return 1 }
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := NewRunner(Options{Connector: "sample"}).Run(ctx)
	if err == nil || !strings.Contains(err.Error(), context.Canceled.Error()) {
		t.Fatalf("Run error=%v, want context canceled", err)
	}
	if calls.Load() != 0 {
		t.Fatalf("pre-cancelled runner invoked CLI %d times", calls.Load())
	}
}

func TestReviewCorrectionFailedPreflightGatesLaterStages(t *testing.T) {
	runner := NewRunner(Options{Connector: "sample"})
	SabotageExpectedKind(runner, "preflight", "WrongKind")
	rep, err := runner.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for _, stage := range rep.Stages {
		if stage.Name == "credentials_add" || stage.Name == "credentials_test" || stage.Name == "catalog" || stage.Name == "write_create" {
			t.Fatalf("failed preflight allowed later stage %q", stage.Name)
		}
	}
}

func TestReviewCorrectionBoundedCleanupUsesPlanPreviewRunAfterCancellation(t *testing.T) {
	oldRun := cliRun
	oldRunContext := cliRunContext
	t.Cleanup(func() {
		cliRun = oldRun
		cliRunContext = oldRunContext
	})
	var calls []string
	cliRun = func(args []string, stdout, _ io.Writer) int {
		switch {
		case hasArgs(args, "reverse", "plan"):
			calls = append(calls, "plan")
			fmt.Fprint(stdout, "Created reverse plan cleanup-plan\nApproval token: approval-marker\n")
		case hasArgs(args, "reverse", "preview"):
			calls = append(calls, "preview")
			fmt.Fprint(stdout, `{"kind":"ReversePlanPreview","plan":{}}`)
		case hasArgs(args, "reverse", "run"):
			calls = append(calls, "run")
			fmt.Fprint(stdout, `{"kind":"ReverseRun"}`)
		}
		return 0
	}
	cliRunContext = func(ctx context.Context, args []string, stdout, stderr io.Writer, _ CLIInvocationOptions) int {
		if err := ctx.Err(); err != nil {
			t.Errorf("bounded cleanup received canceled context: %v", err)
			return 1
		}
		return cliRun(args, stdout, stderr)
	}
	ledger, err := NewLedger(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	harness := NewHarness(t.TempDir(), WithContext(cancelled))
	rc := &runContext{
		ctx:     cancelled,
		harness: harness,
		root:    t.TempDir(),
		opts:    Options{Connector: "sample"},
		write: &writeContext{
			pairing:      WritePairing{Create: "create", Cleanup: "delete"},
			tag:          "pm-cert-sample-12345678-1700000000",
			runID8:       "12345678",
			selfTest:     true,
			createPassed: true,
			ledger:       ledger,
		},
	}
	var rep Report
	runBoundedCancellationCleanup(rc, &rep)
	assertPlanPreviewRunOrder(t, calls)
	if !rc.write.cleanupAttempted {
		t.Fatal("bounded cleanup was not marked attempted")
	}
}

func TestReviewCorrectionResumeRejectsFutureSchemaAndIdentityMismatch(t *testing.T) {
	dir := t.TempDir()
	certDir := filepath.Join(dir, certificationsDirName)
	if err := os.MkdirAll(certDir, 0o700); err != nil {
		t.Fatal(err)
	}
	base := map[string]any{
		"kind":                          "ConnectorCertification",
		"schema_version":                99,
		"connector":                     "sample",
		"started_at":                    time.Now().Add(-time.Minute).UTC(),
		"completed_at":                  time.Now().UTC(),
		"connector_manifest_hash":       "sha256:mismatched",
		"effective_options_fingerprint": "sha256:mismatched",
	}
	raw, _ := json.Marshal(base)
	if err := os.WriteFile(filepath.Join(certDir, "sample.json"), raw, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, ok := completedReport(dir, "sample"); ok {
		t.Fatal("resume accepted a future-schema, identity-incompatible report")
	}

	base["schema_version"] = ReportSchemaVersion
	raw, _ = json.Marshal(base)
	if err := os.WriteFile(filepath.Join(certDir, "sample.json"), raw, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, ok := completedReport(dir, "sample", resumeIdentity{ManifestHash: "sha256:current", Fingerprint: "sha256:current"}); ok {
		t.Fatal("resume accepted mismatched manifest/options identity")
	}
}

func hasArgs(args []string, want ...string) bool {
	if len(args) < len(want) {
		return false
	}
	for i := range want {
		if args[i] != want[i] {
			return false
		}
	}
	return true
}

func flagValue(args []string, name string) string {
	for i, arg := range args {
		if arg == name && i+1 < len(args) {
			return args[i+1]
		}
		if strings.HasPrefix(arg, name+"=") {
			return strings.TrimPrefix(arg, name+"=")
		}
	}
	return ""
}

func assertPlanPreviewRunOrder(t *testing.T, calls []string) {
	t.Helper()
	joined := strings.Join(calls, ",")
	if !strings.Contains(joined, "plan,preview,run") {
		t.Fatalf("cleanup call order=%v, want plan -> preview -> run", calls)
	}
}
