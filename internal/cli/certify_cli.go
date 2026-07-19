package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"polymetrics.ai/internal/connectors/certify"
	"polymetrics.ai/internal/safety"
	"polymetrics.ai/internal/telemetry"
)

type certifyCommandRuntime interface {
	RunSingle(ctx context.Context, root string, opts certify.Options) (certify.Report, error)
	LoadCredsFile(path string) (certify.CredsFile, error)
	RunBatch(ctx context.Context, root string, credsFile certify.CredsFile, resume bool) (certify.BatchReport, error)
	Sweep(ctx context.Context, root, credsPath string, olderThan time.Duration) (map[string]certify.SweepResult, error)
}

type defaultCertifyCommandRuntime struct{}

type certifyCommandFlags struct {
	Credentials          []string
	FromEnv              []string
	Configs              []string
	Streams              []string
	Limits               []string
	Modes                []string
	Skips                []string
	RateLimits           []string
	Budgets              []string
	Records              []string
	Replays              []string
	LiveAllModes         []string
	AllowProductionWrite []string
	KeepWorkdirs         []string
	Writes               []string
	Fulls                []string
	Alls                 []string
	CredentialsFiles     []string
	Parallels            []string
	Resumes              []string
	Sweeps               []string
	OlderThans           []string
}

func (defaultCertifyCommandRuntime) RunSingle(ctx context.Context, _ string, opts certify.Options) (certify.Report, error) {
	return certify.NewRunner(opts).Run(ctx)
}

func (defaultCertifyCommandRuntime) LoadCredsFile(path string) (certify.CredsFile, error) {
	return certify.LoadCredsFile(path)
}

func (defaultCertifyCommandRuntime) RunBatch(ctx context.Context, root string, credsFile certify.CredsFile, resume bool) (certify.BatchReport, error) {
	return certify.RunBatch(ctx, certify.BatchOptions{
		CredsFile:     credsFile,
		RunnerFactory: certify.DefaultRunnerFactory,
		BatchDir:      filepath.Join(root, ".polymetrics"),
		Resume:        resume,
	})
}

func (defaultCertifyCommandRuntime) Sweep(ctx context.Context, root, credsPath string, olderThan time.Duration) (map[string]certify.SweepResult, error) {
	connectors, err := sweepTargetConnectors(root, credsPath)
	if err != nil {
		return nil, err
	}
	if len(connectors) == 0 {
		return nil, usageErrorf("pm connectors certify --sweep found no ledger to sweep under %s (pass --credentials-file, or run a live/batch certify first)", root)
	}
	results := make(map[string]certify.SweepResult, len(connectors))
	for _, name := range connectors {
		ledgerRoot := filepath.Join(root, ".polymetrics", "certifications", "ledger", name)
		result, err := certify.NewSweeper(certify.SweeperOptions{Root: ledgerRoot, OlderThan: olderThan}).Sweep(ctx)
		if err != nil {
			return nil, fmt.Errorf("certify: sweep %s: %w", name, err)
		}
		results[name] = result
	}
	return results, nil
}

func newCertifyCobraCommand(ctx context.Context, root string, stdout io.Writer, jsonOut bool, runtime certifyCommandRuntime) *cobra.Command {
	var flags certifyCommandFlags
	cmd := newConnectorsActionCobraCommand("certify [connector]", func(_ *cobra.Command, args []string) error {
		if firstArgIsHelp(args) {
			return markCobraLegacyError(writeManual("connectors", stdout, jsonOut))
		}
		switch {
		case lastString(flags.Sweeps) == "true" && lastString(flags.Alls) == "true":
			return usageErrorf("--all and --sweep are not supported together")
		case lastString(flags.Sweeps) == "true":
			return markCobraLegacyError(runCertifySweep(ctx, root, flags, stdout, jsonOut, runtime))
		case lastString(flags.Alls) == "true":
			return markCobraLegacyError(runCertifyBatch(ctx, root, flags, stdout, jsonOut, runtime))
		default:
			if len(args) != 1 {
				return usageErrorf("pm connectors certify <connector> | --all --credentials-file <file> | --sweep")
			}
			return markCobraLegacyError(runCertifySingle(ctx, root, args[0], flags, stdout, jsonOut, runtime))
		}
	})
	setManualHelp(cmd, "connectors", stdout, jsonOut)
	addCertifyFlags(cmd, &flags)
	return cmd
}

func addCertifyFlags(cmd *cobra.Command, flags *certifyCommandFlags) {
	for _, spec := range []struct {
		name   string
		target *[]string
	}{
		{"from-env", &flags.FromEnv}, {"config", &flags.Configs}, {"stream", &flags.Streams},
		{"skip", &flags.Skips}, {"keep-workdir", &flags.KeepWorkdirs}, {"write", &flags.Writes},
		{"full", &flags.Fulls}, {"all", &flags.Alls}, {"credentials-file", &flags.CredentialsFiles},
		{"parallel", &flags.Parallels}, {"resume", &flags.Resumes}, {"sweep", &flags.Sweeps},
		{"older-than", &flags.OlderThans},
	} {
		addConnectorsStringArrayFlag(cmd, spec.target, spec.name, "connector certification option")
	}
	for _, spec := range []struct {
		name   string
		target *[]string
	}{
		{"credential", &flags.Credentials}, {"limit", &flags.Limits}, {"modes", &flags.Modes},
		{"record", &flags.Records}, {"replay", &flags.Replays},
		{"allow-production-writes", &flags.AllowProductionWrite}, {"rate-limit", &flags.RateLimits},
		{"budget", &flags.Budgets}, {"live-all-modes", &flags.LiveAllModes},
	} {
		addConnectorsStringArrayFlag(cmd, spec.target, spec.name, "unsupported connector certification option")
		if flag := cmd.Flags().Lookup(spec.name); flag != nil {
			flag.Hidden = true
		}
	}
}

// --- single-connector mode ---

func runCertifySingle(ctx context.Context, root, connector string, flags certifyCommandFlags, stdout io.Writer, jsonOut bool, runtime certifyCommandRuntime) (err error) {
	ctx, span := telemetry.StartSpan(ctx, "pm.certify.connector",
		telemetry.StringAttr("pm.certify.connector", connector),
		telemetry.StringAttr("pm.certify.mode", "single"),
	)
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetAttributes(telemetry.StringAttr("pm.certify.status", "failed"))
		} else {
			span.SetAttributes(telemetry.StringAttr("pm.certify.status", "ok"))
		}
		span.End()
	}()

	if err := safety.ValidateIdentifier(connector, "connector"); err != nil {
		return validationErrorf("%v", err)
	}

	if err := validateCertifySingleFlags(flags); err != nil {
		return err
	}

	opts, err := certifyOptionsFromFlags(connector, flags)
	if err != nil {
		return err
	}

	rep, err := runtime.RunSingle(ctx, root, opts)
	if err != nil {
		return fmt.Errorf("certify: %w", err)
	}

	rep.PMVersion = version
	saveDir := filepath.Join(root, ".polymetrics")
	_ = rep.Save(saveDir) // best-effort: a report-persistence failure must not mask the certification result itself.

	if err := writeCertifyReport(stdout, jsonOut, rep); err != nil {
		return err
	}

	return exitForReport(rep)
}

// certifyOptionsFromFlags builds certify.Options from the implemented
// single-connector controls. validateCertifySingleFlags must run first so no
// declared-but-unimplemented option can reach the runner as a no-op.
func certifyOptionsFromFlags(connector string, flags certifyCommandFlags) (certify.Options, error) {
	secretEnv := map[string]string{}
	for _, spec := range flags.FromEnv {
		field, env, ok := strings.Cut(spec, "=")
		if !ok || field == "" || env == "" {
			return certify.Options{}, usageErrorf("invalid --from-env %q, want field=ENV", spec)
		}
		secretEnv[field] = env
	}

	config, err := keyValues(flags.Configs)
	if err != nil {
		return certify.Options{}, err
	}

	write := lastString(flags.Writes) == "true"
	if skipWrite(flags) {
		write = false
	}

	return certify.Options{
		Connector: connector,
		Stream:    lastString(flags.Streams),
		Config:    config,
		SecretEnv: secretEnv,
		KeepWork:  lastString(flags.KeepWorkdirs) == "true",
		Write:     write,
		Full:      lastString(flags.Fulls) == "true",
	}, nil
}

func rejectUnsupportedCertifyControls(flags certifyCommandFlags) error {
	for _, control := range []struct {
		name   string
		values []string
	}{
		{"credential", flags.Credentials},
		{"limit", flags.Limits},
		{"modes", flags.Modes},
		{"record", flags.Records},
		{"replay", flags.Replays},
		{"allow-production-writes", flags.AllowProductionWrite},
		{"rate-limit", flags.RateLimits},
		{"budget", flags.Budgets},
		{"live-all-modes", flags.LiveAllModes},
	} {
		if len(control.values) != 0 {
			return usageErrorf("--%s is not supported; refusing to run certification with a no-op control", control.name)
		}
	}
	return nil
}

func validateCertifySingleFlags(flags certifyCommandFlags) error {
	if err := rejectUnsupportedCertifyControls(flags); err != nil {
		return err
	}
	if err := validateCertifySkips(flags.Skips); err != nil {
		return err
	}
	return rejectCertifyModeControls("single", []certifyModeControl{
		{"credentials-file", flags.CredentialsFiles},
		{"parallel", flags.Parallels},
		{"resume", flags.Resumes},
		{"older-than", flags.OlderThans},
	})
}

func validateCertifyBatchFlags(flags certifyCommandFlags) error {
	if err := rejectUnsupportedCertifyControls(flags); err != nil {
		return err
	}
	if err := validateCertifySkips(flags.Skips); err != nil {
		return err
	}
	if err := rejectCertifyModeControls("batch", []certifyModeControl{
		{"from-env", flags.FromEnv},
		{"config", flags.Configs},
		{"stream", flags.Streams},
		{"keep-workdir", flags.KeepWorkdirs},
		{"full", flags.Fulls},
		{"older-than", flags.OlderThans},
	}); err != nil {
		return err
	}
	if len(flags.Writes) > 0 && lastString(flags.Writes) != "false" {
		return usageErrorf("--write=%s is not supported in batch mode; only --write=false may override credential-file writes", lastString(flags.Writes))
	}
	return nil
}

func validateCertifySweepFlags(flags certifyCommandFlags) error {
	if err := rejectUnsupportedCertifyControls(flags); err != nil {
		return err
	}
	if len(flags.Skips) > 0 {
		return usageErrorf("--skip is not supported in sweep mode")
	}
	return rejectCertifyModeControls("sweep", []certifyModeControl{
		{"from-env", flags.FromEnv},
		{"config", flags.Configs},
		{"stream", flags.Streams},
		{"keep-workdir", flags.KeepWorkdirs},
		{"write", flags.Writes},
		{"full", flags.Fulls},
		{"parallel", flags.Parallels},
		{"resume", flags.Resumes},
	})
}

type certifyModeControl struct {
	name   string
	values []string
}

func rejectCertifyModeControls(mode string, controls []certifyModeControl) error {
	for _, control := range controls {
		if len(control.values) > 0 {
			return usageErrorf("--%s is not supported in %s certification mode", control.name, mode)
		}
	}
	return nil
}

func validateCertifySkips(values []string) error {
	for _, skip := range parseCSVFlags(values) {
		if skip != "write" {
			return usageErrorf("--skip=%s is not supported; only --skip=write is implemented", skip)
		}
	}
	return nil
}

func skipWrite(flags certifyCommandFlags) bool {
	for _, skip := range parseCSVFlags(flags.Skips) {
		if skip == "write" {
			return true
		}
	}
	return false
}

func batchWriteDisabled(flags certifyCommandFlags) bool {
	return (len(flags.Writes) > 0 && lastString(flags.Writes) == "false") || skipWrite(flags)
}

func disableCredentialFileWrites(credsFile *certify.CredsFile) {
	for name, entry := range credsFile.Connectors {
		entry.Write = false
		credsFile.Connectors[name] = entry
	}
}

// --- batch mode (--all --credentials-file) ---

func runCertifyBatch(ctx context.Context, root string, flags certifyCommandFlags, stdout io.Writer, jsonOut bool, runtime certifyCommandRuntime) (err error) {
	ctx, span := telemetry.StartSpan(ctx, "pm.certify.batch", telemetry.StringAttr("pm.certify.mode", "batch"))
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetAttributes(telemetry.StringAttr("pm.certify.status", "failed"))
		} else {
			span.SetAttributes(telemetry.StringAttr("pm.certify.status", "ok"))
		}
		span.End()
	}()

	if err := validateCertifyBatchFlags(flags); err != nil {
		return err
	}

	credsPath := lastString(flags.CredentialsFiles)
	if credsPath == "" {
		return usageErrorf("pm connectors certify --all requires --credentials-file <file>")
	}

	credsFile, err := runtime.LoadCredsFile(credsPath)
	if err != nil {
		return err
	}
	if batchWriteDisabled(flags) {
		disableCredentialFileWrites(&credsFile)
	}
	if err := certify.ValidateBatchConstraints(credsFile); err != nil {
		return usageErrorf("%v", err)
	}

	parallel, err := parseIntFlag("parallel", lastString(flags.Parallels), 0)
	if err != nil {
		return err
	}
	if parallel > 0 {
		credsFile.Defaults.Parallel = parallel
	}

	batch, err := runtime.RunBatch(ctx, root, credsFile, lastString(flags.Resumes) == "true")
	if err != nil {
		return fmt.Errorf("certify: batch run failed: %w", err)
	}

	if err := writeCertifyBatchReport(stdout, jsonOut, batch); err != nil {
		return err
	}

	return exitForBatch(batch)
}

// --- sweep mode (--sweep) ---

func runCertifySweep(ctx context.Context, root string, flags certifyCommandFlags, stdout io.Writer, jsonOut bool, runtime certifyCommandRuntime) error {
	if err := validateCertifySweepFlags(flags); err != nil {
		return err
	}

	olderThan := 24 * time.Hour
	if raw := lastString(flags.OlderThans); raw != "" {
		d, err := time.ParseDuration(raw)
		if err != nil {
			return usageErrorf("invalid --older-than %q: %v", raw, err)
		}
		olderThan = d
	}

	results, err := runtime.Sweep(ctx, root, lastString(flags.CredentialsFiles), olderThan)
	if err != nil {
		return err
	}

	if err := writeCertifySweepReport(stdout, jsonOut, results); err != nil {
		return err
	}

	return exitForSweep(results)
}

// sweepTargetConnectors lists the connectors to sweep: every entry in
// credsPath's CredsFile if given, else every subdirectory already present
// under <root>/.polymetrics/certifications/ledger/ (certification design §C
// "Ledger copied into .polymetrics/certifications/ledger/ even on crash").
func sweepTargetConnectors(root, credsPath string) ([]string, error) {
	if credsPath != "" {
		cf, err := certify.LoadCredsFile(credsPath)
		if err != nil {
			return nil, err
		}
		return cf.ConnectorNames(), nil
	}
	ledgerRoot := filepath.Join(root, ".polymetrics", "certifications", "ledger")
	return listSubdirNames(ledgerRoot), nil
}

// listSubdirNames returns the names of dir's immediate subdirectories, or an
// empty slice if dir doesn't exist (never an error — an absent ledger root
// just means "nothing to sweep yet").
func listSubdirNames(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	return names
}

// --- output rendering ---

func writeCertifyReport(stdout io.Writer, jsonOut bool, rep certify.Report) error {
	if jsonOut {
		return writeJSON(stdout, envelope{"kind": "ConnectorCertification", "report": rep})
	}
	fmt.Fprint(stdout, renderCertifyReportText(rep))
	return nil
}

func renderCertifyReportText(rep certify.Report) string {
	var b strings.Builder
	status := "FAIL"
	if rep.Passed {
		status = "PASS"
	}
	fmt.Fprintf(&b, "Certification: %s [%s]\n", rep.Connector, status)
	fmt.Fprintf(&b, "  check:    %s\n", rep.Capabilities.Check.Result)
	fmt.Fprintf(&b, "  catalog:  %s (streams=%d)\n", rep.Capabilities.Catalog.Result, rep.Capabilities.Catalog.Streams)
	fmt.Fprintf(&b, "  read:     %s (stream=%s records=%d)\n", rep.Capabilities.Read.Result, rep.Capabilities.Read.Stream, rep.Capabilities.Read.Records)
	fmt.Fprintf(&b, "  resume:   %s\n", rep.Capabilities.Resume.Result)
	fmt.Fprintf(&b, "  redaction:%s\n", rep.Capabilities.SecretRedaction.Result)
	if len(rep.Leaks) > 0 {
		fmt.Fprintf(&b, "  LEAKED RESOURCES: %d\n", len(rep.Leaks))
		for _, leak := range rep.Leaks {
			fmt.Fprintf(&b, "    - %s (%s): %s\n", leak.Tag, leak.Connector, leak.Reason)
		}
	}
	for _, stage := range rep.Stages {
		if stage.Passed {
			continue
		}
		// A "skipped: ..." stage (e.g. fixture_conformance with no defs
		// bundle yet, or a write stage with Options.Write disabled) is a
		// documented skip, not a failure — Report.Passed itself already
		// treats these as non-failing (stages_source.go allStagesPassed),
		// so the text summary must not label them FAILED too.
		if stage.Name == "fixture_conformance" || strings.HasPrefix(stage.Error, "skipped: ") {
			fmt.Fprintf(&b, "  stage %s: skipped: %s\n", stage.Name, strings.TrimPrefix(stage.Error, "skipped: "))
			continue
		}
		fmt.Fprintf(&b, "  stage %s: FAILED: %s\n", stage.Name, stage.Error)
	}
	return b.String()
}

func writeCertifyBatchReport(stdout io.Writer, jsonOut bool, batch certify.BatchReport) error {
	if jsonOut {
		return writeJSON(stdout, envelope{"kind": "ConnectorCertificationBatch", "batch": batch, "matrix": batch.SummaryMatrix().Rows})
	}
	fmt.Fprint(stdout, renderBatchMatrixText(batch))
	return nil
}

// renderBatchMatrixText renders the certification design §B summary matrix
// as a tab-separated table, leak rows first (SummaryMatrix already sorts
// them that way).
func renderBatchMatrixText(batch certify.BatchReport) string {
	var b strings.Builder
	matrix := batch.SummaryMatrix()
	if leaks := batch.Leaks(); len(leaks) > 0 {
		fmt.Fprintf(&b, "LEAKED RESOURCES (%d):\n", len(leaks))
		for _, leak := range leaks {
			fmt.Fprintf(&b, "  - %s (%s): %s\n", leak.Tag, leak.Connector, leak.Reason)
		}
	}
	fmt.Fprintln(&b, "connector\tcheck\tcatalog\tread\tresume\twrite\tflow\tschedule\tredaction\tleaked")
	for _, row := range matrix.Rows {
		fmt.Fprintf(&b, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%v\n",
			row.Connector, row.Check, row.Catalog, row.Read, row.Resume, row.Write, row.Flow, row.Schedule, row.Redaction, row.Leaked)
	}
	fmt.Fprintf(&b, "exit_code: %d\n", batch.ExitCode)
	return b.String()
}

func writeCertifySweepReport(stdout io.Writer, jsonOut bool, results map[string]certify.SweepResult) error {
	if jsonOut {
		return writeJSON(stdout, envelope{"kind": "ConnectorCertificationSweep", "results": results})
	}
	var b strings.Builder
	for name, res := range results {
		fmt.Fprintf(&b, "%s: scanned=%d cleaned=%d skipped=%d failed=%d\n",
			name, res.Scanned, len(res.Cleaned), len(res.Skipped), len(res.Failed))
	}
	fmt.Fprint(stdout, b.String())
	return nil
}

// --- exit-code mapping (certification design §A: 0 pass / 1 usage-internal
// / 2 certification failures / 3 leaked resources) ---

func exitForReport(rep certify.Report) error {
	code := certify.ExitCodeFor(rep)
	if code == 0 {
		return nil
	}
	return certifyExitErrorf(code, "certification %s: exit %d", rep.Connector, code)
}

func exitForBatch(batch certify.BatchReport) error {
	if batch.ExitCode == 0 {
		return nil
	}
	return certifyExitErrorf(batch.ExitCode, "batch certification: exit %d", batch.ExitCode)
}

func exitForSweep(results map[string]certify.SweepResult) error {
	for _, res := range results {
		if len(res.Failed) > 0 {
			return certifyExitErrorf(3, "sweep: %d entries failed cleanup", len(res.Failed))
		}
	}
	return nil
}
