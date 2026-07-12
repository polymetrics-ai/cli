package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors/certify"
	"polymetrics.ai/internal/safety"
)

// runCertify dispatches `pm connectors certify ...` (certification design §A
// command spec): a single connector by name, `--all --credentials-file`
// batch mode (§B), or `--sweep` orphan cleanup (§C). Purely additive to the
// existing `connectors` subcommand family in cli.go — no other connectors
// subcommand's behavior changes.
func runCertify(ctx context.Context, root string, args []string, stdout io.Writer, jsonOut bool) error {
	flags := parseFlags(args)
	positionals := flags.values["_"]

	switch {
	case flags.first("sweep") == "true":
		return runCertifySweep(ctx, root, flags, stdout, jsonOut)
	case flags.first("all") == "true":
		return runCertifyBatch(ctx, root, flags, stdout, jsonOut)
	default:
		if len(positionals) != 1 {
			return usageErrorf("pm connectors certify <connector> | --all --credentials-file <file> | --sweep")
		}
		return runCertifySingle(ctx, root, positionals[0], flags, stdout, jsonOut)
	}
}

// --- single-connector mode ---

func runCertifySingle(ctx context.Context, root, connector string, flags parsedFlags, stdout io.Writer, jsonOut bool) error {
	if err := safety.ValidateIdentifier(connector, "connector"); err != nil {
		return validationErrorf("%v", err)
	}

	opts, err := certifyOptionsFromFlags(connector, flags)
	if err != nil {
		return err
	}

	runner := certify.NewRunner(opts)
	rep, err := runner.Run(ctx)
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

// certifyOptionsFromFlags builds certify.Options from `pm connectors certify
// <connector>` flags: --stream, --limit, --modes, --skip, --keep-workdir,
// --from-env (repeatable field=ENV), --config (repeatable key=value), --write
// (certification design §A command spec).
func certifyOptionsFromFlags(connector string, flags parsedFlags) (certify.Options, error) {
	limit, err := parseIntFlag("limit", flags.first("limit"), 50)
	if err != nil {
		return certify.Options{}, err
	}

	secretEnv := map[string]string{}
	for _, spec := range flags.values["from-env"] {
		field, env, ok := strings.Cut(spec, "=")
		if !ok || field == "" || env == "" {
			return certify.Options{}, usageErrorf("invalid --from-env %q, want field=ENV", spec)
		}
		secretEnv[field] = env
	}

	config, err := keyValues(flags.values["config"])
	if err != nil {
		return certify.Options{}, err
	}

	skip := parseCSVFlags(flags.values["skip"])
	write := flags.first("write") == "true"
	full := flags.first("full") == "true"
	for _, s := range skip {
		if s == "write" {
			write = false
		}
	}

	return certify.Options{
		Connector: connector,
		Stream:    flags.first("stream"),
		Limit:     limit,
		Modes:     parseCSVFlags(flags.values["modes"]),
		Config:    config,
		SecretEnv: secretEnv,
		KeepWork:  flags.first("keep-workdir") == "true",
		Write:     write,
		Full:      full,
	}, nil
}

// --- batch mode (--all --credentials-file) ---

func runCertifyBatch(ctx context.Context, root string, flags parsedFlags, stdout io.Writer, jsonOut bool) error {
	credsPath := flags.first("credentials-file")
	if credsPath == "" {
		return usageErrorf("pm connectors certify --all requires --credentials-file <file>")
	}

	cf, err := certify.LoadCredsFile(credsPath)
	if err != nil {
		return err
	}

	if parallel, perr := parseIntFlag("parallel", flags.first("parallel"), 0); perr != nil {
		return perr
	} else if parallel > 0 {
		cf.Defaults.Parallel = parallel
	}

	batchDir := filepath.Join(root, ".polymetrics")
	batch, err := certify.RunBatch(ctx, certify.BatchOptions{
		CredsFile:     cf,
		RunnerFactory: certify.DefaultRunnerFactory,
		BatchDir:      batchDir,
		Resume:        flags.first("resume") == "true",
	})
	if err != nil {
		return fmt.Errorf("certify: batch run failed: %w", err)
	}

	if err := writeCertifyBatchReport(stdout, jsonOut, batch); err != nil {
		return err
	}

	return exitForBatch(batch)
}

// --- sweep mode (--sweep) ---

func runCertifySweep(ctx context.Context, root string, flags parsedFlags, stdout io.Writer, jsonOut bool) error {
	olderThan := 24 * time.Hour
	if raw := flags.first("older-than"); raw != "" {
		d, err := time.ParseDuration(raw)
		if err != nil {
			return usageErrorf("invalid --older-than %q: %v", raw, err)
		}
		olderThan = d
	}

	connectors, err := sweepTargetConnectors(root, flags.first("credentials-file"))
	if err != nil {
		return err
	}
	if len(connectors) == 0 {
		return usageErrorf("pm connectors certify --sweep found no ledger to sweep under %s (pass --credentials-file, or run a live/batch certify first)", root)
	}

	results := make(map[string]certify.SweepResult, len(connectors))
	for _, name := range connectors {
		ledgerRoot := filepath.Join(root, ".polymetrics", "certifications", "ledger", name)
		sweeper := certify.NewSweeper(certify.SweeperOptions{Root: ledgerRoot, OlderThan: olderThan})
		res, err := sweeper.Sweep(ctx)
		if err != nil {
			return fmt.Errorf("certify: sweep %s: %w", name, err)
		}
		results[name] = res
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
	_, err := fmt.Fprint(stdout, renderCertifyReportText(rep))
	return err
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
	_, err := fmt.Fprint(stdout, renderBatchMatrixText(batch))
	return err
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
	_, err := fmt.Fprint(stdout, b.String())
	return err
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
