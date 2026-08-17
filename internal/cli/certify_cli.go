package cli

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors/certify"
	"polymetrics.ai/internal/safety"
)

const certificationExternalChildEnv = "PM_CERTIFICATION_EXTERNAL_CHILD"

// certificationExternalRuntimeObservationEnv requests one secret-safe
// self-observation artifact from a fresh external-proof child. It is an
// integration-test seam, not a command-line option and never skips a proof.
const certificationExternalRuntimeObservationEnv = "PM_CERTIFICATION_EXTERNAL_RUNTIME_OBSERVATION"

// runCertify dispatches `pm connectors certify ...` (certification design §A
// command spec): a single connector by name, `--all --credentials-file`
// batch mode (§B), or `--sweep` orphan cleanup (§C). Purely additive to the
// existing `connectors` subcommand family in cli.go — no other connectors
// subcommand's behavior changes.
func runCertify(ctx context.Context, root string, args []string, stdout, stderr io.Writer, jsonOut bool) error {
	flags := parseFlags(args)
	positionals := flags.values["_"]
	if flags.first("external-proof") == "true" && os.Getenv(certificationExternalChildEnv) != "1" {
		if flags.first("all") == "true" || flags.first("sweep") == "true" || len(positionals) != 1 {
			return usageErrorf("pm connectors certify --external-proof requires one connector")
		}
		opts, err := certifyOptionsFromFlags(positionals[0], flags)
		if err != nil {
			return err
		}
		if err := rejectCertificationSecretArgv(args, opts); err != nil {
			return err
		}
		if flags.first("full-parity") != "true" && flags.first("direct-read-only") != "true" {
			return usageErrorf("pm connectors certify --external-proof requires --full-parity or --direct-read-only")
		}
		childArgs, childEnv, preparedValues, err := prepareExternalCertifyCredentialInput(args, flags, opts)
		if err != nil {
			return err
		}
		return runExternalCertifyChild(ctx, root, childArgs, stdout, stderr, preparedValues, childEnv)
	}

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
	// A certification workdir is intentionally ephemeral. Its write-ahead
	// lifecycle ledger is not: persist it where --sweep discovers it so a
	// crash or rate-limit restart can reconcile tagged resources rather than
	// replaying a create from scratch.
	opts.LedgerRoot = filepath.Join(root, ".polymetrics", "certifications", "ledger", connector)
	if opts.Full {
		opts.Resume = flags.first("resume") == "true"
		opts.DirectReadCheckpointPath = filepath.Join(root, ".polymetrics", "certifications", "progress", connector+"-direct-read.json")
	}
	externalProof := flags.first("external-proof") == "true"
	if externalProof {
		if os.Getenv(certificationExternalChildEnv) != "1" {
			return errors.New("external certification proof must execute in a fresh child process")
		}
		if flags.first("full-parity") != "true" && !opts.DirectReadOnly {
			return usageErrorf("pm connectors certify --external-proof requires --full-parity or --direct-read-only")
		}
		if err := rejectCertificationSecretArgv(os.Args[1:], opts); err != nil {
			return err
		}
		opts.ObserveHTTP = true
		if observationPath := os.Getenv(certificationExternalRuntimeObservationEnv); observationPath != "" {
			opts.RuntimeObservation = func(input certify.RuntimeObservationInput) error {
				return writeExternalRuntimeObservation(observationPath, root, input)
			}
		}
	}

	runner := certify.NewRunner(opts)
	rep, err := runner.Run(ctx)
	if err != nil {
		return fmt.Errorf("certify: %w", err)
	}

	rep.PMVersion = version
	saveDir := filepath.Join(root, ".polymetrics")
	_ = rep.Save(saveDir) // best-effort: a report-persistence failure must not mask the certification result itself.

	var rendered bytes.Buffer
	if err := writeCertifyReport(&rendered, jsonOut, rep); err != nil {
		return err
	}
	if externalProof {
		if !rep.Passed || exitCodeForReport(rep) != 0 {
			diagnostic, err := certificationFailedReportDiagnostic(rep, certificationPreparedValues(opts))
			if err != nil {
				return errors.New("certify external proof: external proof requires a completed successful process; fingerprint-redacted report diagnostic unavailable")
			}
			return fmt.Errorf("certify external proof: external proof requires a completed successful process; fingerprint-redacted diagnostic: %s", diagnostic)
		}
		binarySHA256, err := currentBinarySHA256()
		if err != nil {
			return err
		}
		flowReferences, err := certificationFlowRoundTripReferences(rep, certificationPreparedValues(opts))
		if err != nil {
			return fmt.Errorf("certify external proof: %w", err)
		}
		runID := fmt.Sprintf("external-%d", rep.StartedAt.UTC().UnixNano())
		if _, err := certify.WriteExternalProof(root, certify.ExternalProofInput{
			Connector:               connector,
			RunID:                   runID,
			BinarySHA256:            binarySHA256,
			Command:                 append([]string(nil), os.Args...),
			Stdout:                  rendered.String(),
			ExitCode:                exitCodeForReport(rep),
			Passed:                  rep.Passed,
			FullParity:              rep.FullParityVerified(),
			PreparedValues:          certificationPreparedValues(opts),
			HTTPExchanges:           runner.ObservedHTTPExchanges(),
			FlowRoundTripReferences: flowReferences,
		}); err != nil {
			return fmt.Errorf("certify external proof: %w", err)
		}
	}
	if _, err := io.Copy(stdout, &rendered); err != nil {
		return err
	}

	return exitForReport(rep)
}

// certificationFlowRoundTripReferences turns the successful in-process flow
// read-back stages into safe proof references. Accepted external evidence must
// name the completed plan, preview, execution, and status/read-back steps
// rather than treating a source-only HTTP transcript as a complete workflow.
func certificationFlowRoundTripReferences(rep certify.Report, preparedValues []string) ([]string, error) {
	required := []string{"flow_plan", "flow_preview", "flow_run", "flow_status"}
	for _, name := range required {
		stagePassed := false
		var failed *certify.StageResult
		for index := range rep.Stages {
			stage := &rep.Stages[index]
			if stage.Name != name {
				continue
			}
			if stage.Passed {
				stagePassed = true
				break
			}
			failed = stage
		}
		if stagePassed {
			continue
		}
		if failed == nil {
			aggregate := certificationStageResult(rep, "flow_roundtrip")
			if aggregate == nil {
				return nil, fmt.Errorf("full external certification did not complete flow round-trip stage %q: no passing stage result", name)
			}
			redacted, err := certificationStageDiagnostic(*aggregate, preparedValues)
			if err != nil {
				return nil, fmt.Errorf("full external certification did not complete flow round-trip stage %q: fingerprint-redacted aggregate diagnostic unavailable", name)
			}
			predecessor := certificationStageResult(rep, "etl_full_refresh_append")
			if predecessor == nil || predecessor.Passed {
				return nil, fmt.Errorf("full external certification did not complete flow round-trip stage %q: no named stage result; flow_roundtrip aggregate: %s", name, redacted)
			}
			predecessorRedacted, err := certificationStageDiagnostic(*predecessor, preparedValues)
			if err != nil {
				return nil, fmt.Errorf("full external certification did not complete flow round-trip stage %q: fingerprint-redacted prerequisite diagnostic unavailable", name)
			}
			return nil, fmt.Errorf("full external certification did not complete flow round-trip stage %q: no named stage result; flow_roundtrip aggregate: %s; etl_full_refresh_append prerequisite: %s", name, redacted, predecessorRedacted)
		}
		redacted, err := certificationStageDiagnostic(*failed, preparedValues)
		if err != nil {
			return nil, fmt.Errorf("full external certification did not complete flow round-trip stage %q: fingerprint-redacted diagnostic unavailable", name)
		}
		return nil, fmt.Errorf("full external certification did not complete flow round-trip stage %q: %s", name, redacted)
	}
	return required, nil
}

func certificationStageResult(rep certify.Report, name string) *certify.StageResult {
	for index := range rep.Stages {
		if rep.Stages[index].Name == name {
			return &rep.Stages[index]
		}
	}
	return nil
}

func certificationStageDiagnostic(stage certify.StageResult, preparedValues []string) (string, error) {
	diagnostic := fmt.Sprintf("stage=%q status=%q exit_code=%d kind=%q error=%s", stage.Name, stage.Status, stage.CLI.ExitCode, stage.CLI.Kind, stage.Error)
	return certify.RedactExternalProofDiagnostic(diagnostic, preparedValues)
}

// certificationFailedReportDiagnostic exposes the one actionable stage from a
// completed but unsuccessful external certification report. It deliberately
// redacts before returning because the report's Error fields may originate in
// provider responses and must never reach terminal output unfiltered.
func certificationFailedReportDiagnostic(rep certify.Report, preparedValues []string) (string, error) {
	for _, stage := range rep.Stages {
		if stage.Passed || stage.Status == "skipped" {
			continue
		}
		redacted, err := certificationStageDiagnostic(stage, preparedValues)
		if err != nil {
			return "", err
		}
		return "first non-passing stage: " + redacted, nil
	}
	return "report passed=false without a non-passing stage result", nil
}

func runExternalCertifyChild(ctx context.Context, root string, args []string, stdout, stderr io.Writer, preparedValues, childEnv []string) error {
	moduleRoot, err := certificationModuleRoot()
	if err != nil {
		return err
	}
	buildDir, err := os.MkdirTemp("", "pm-certify-external-")
	if err != nil {
		return fmt.Errorf("certify external proof: create build directory: %w", err)
	}
	defer func() { _ = os.RemoveAll(buildDir) }()
	binaryPath := filepath.Join(buildDir, "pm")
	build := exec.CommandContext(ctx, "go", "build", "-o", binaryPath, "./cmd/pm")
	build.Dir = moduleRoot
	build.Stderr = &bytes.Buffer{}
	if err := build.Run(); err != nil {
		return fmt.Errorf("certify external proof: build fresh pm binary: %w", err)
	}

	childArgs := append([]string{"connectors", "certify"}, args...)
	childArgs = append(childArgs, "--root", root)
	child := exec.CommandContext(ctx, binaryPath, childArgs...)
	child.Env = append(os.Environ(), certificationExternalChildEnv+"=1")
	child.Env = append(child.Env, childEnv...)
	var childStdout, childStderr bytes.Buffer
	child.Stdout = &childStdout
	child.Stderr = &childStderr
	err = child.Run()
	if relayErr := relayExternalCertifyChildOutput(stdout, stderr, childStdout.String(), childStderr.String(), preparedValues); relayErr != nil {
		return relayErr
	}
	if err == nil {
		return nil
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		connector := "unknown"
		if len(args) > 0 {
			connector = args[0]
		}
		return certifyExitErrorf(exitErr.ExitCode(), "external certification %s: exit %d", connector, exitErr.ExitCode())
	}
	return fmt.Errorf("certify external proof: run fresh pm binary: %w", err)
}

const certificationStdinSecretEnv = "PM_CERTIFICATION_STDIN_SECRET"

// prepareExternalCertifyCredentialInput converts the one supported stdin
// credential into a child-only environment value. The parent command line
// keeps only a field name; the raw stdin value is never written, serialized
// into argv, or handed to an ordinary credential profile.
func prepareExternalCertifyCredentialInput(args []string, flags parsedFlags, opts certify.Options) ([]string, []string, []string, error) {
	prepared := certificationPreparedValues(opts)
	stdinValues := flags.values["value-stdin"]
	if len(stdinValues) == 0 {
		return append([]string(nil), args...), nil, prepared, nil
	}
	if len(stdinValues) != 1 || flags.isBare("value-stdin") {
		return nil, nil, nil, usageErrorf("pm connectors certify --external-proof accepts exactly one --value-stdin field")
	}
	field := stdinValues[0]
	if err := safety.ValidateIdentifier(field, "stdin credential field"); err != nil {
		return nil, nil, nil, validationErrorf("%v", err)
	}
	payload, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("read stdin certification credential: %w", err)
	}
	value := strings.TrimRight(string(payload), "\r\n")
	if value == "" {
		return nil, nil, nil, errors.New("stdin certification credential is empty")
	}
	childArgs := replaceCertificationStdinArg(args, field)
	prepared = append(prepared, value)
	return childArgs, []string{certificationStdinSecretEnv + "=" + value}, prepared, nil
}

func replaceCertificationStdinArg(args []string, field string) []string {
	out := make([]string, 0, len(args)+2)
	for index := 0; index < len(args); index++ {
		arg := args[index]
		if arg == "--value-stdin" {
			index++
			continue
		}
		if strings.HasPrefix(arg, "--value-stdin=") {
			continue
		}
		out = append(out, arg)
	}
	return append(out, "--from-env", field+"="+certificationStdinSecretEnv)
}

// relayExternalCertifyChildOutput is the parent-side final credential boundary:
// the child has no terminal attached, so neither stream may reach the invoking
// process until both have been checked against the prepared credential values.
// On a refusal it writes neither raw stream; it returns only a separately
// fingerprint-redacted diagnostic so an external-proof failure stays
// diagnosable without exposing a credential.
func relayExternalCertifyChildOutput(stdout, stderr io.Writer, childStdout, childStderr string, preparedValues []string) error {
	if len(certify.ScanForSecrets(childStdout, preparedValues)) != 0 || len(certify.ScanForSecrets(childStderr, preparedValues)) != 0 {
		diagnostic, err := certify.RedactExternalProofDiagnostic(
			"external certification child stdout:\n"+childStdout+"\nexternal certification child stderr:\n"+childStderr,
			preparedValues,
		)
		if err != nil {
			return errors.New("external certification child output contained credential material; refusing to relay captured streams and diagnostic")
		}
		return fmt.Errorf("external certification child output contained credential material; refusing to relay captured streams; fingerprint-redacted diagnostic:\n%s", diagnostic)
	}
	if _, err := io.WriteString(stdout, childStdout); err != nil {
		return err
	}
	if _, err := io.WriteString(stderr, childStderr); err != nil {
		return err
	}
	return nil
}

func rejectCertificationSecretArgv(args []string, opts certify.Options) error {
	prepared := certificationPreparedValues(opts)
	if len(prepared) == 0 {
		return nil
	}
	if len(certify.ScanForSecrets(strings.Join(args, "\x00"), prepared)) != 0 {
		return validationErrorf("certification credentials must be supplied through --from-env or stdin, never process arguments")
	}
	return nil
}

func certificationPreparedValues(opts certify.Options) []string {
	values := make([]string, 0, len(opts.SecretEnv))
	for _, envName := range opts.SecretEnv {
		if value := os.Getenv(envName); value != "" {
			values = append(values, value)
		}
	}
	return values
}

func certificationModuleRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("certify external proof: determine working directory: %w", err)
	}
	for {
		if _, statErr := os.Stat(filepath.Join(dir, "go.mod")); statErr == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("certify external proof: could not find module root to build a fresh pm binary")
		}
		dir = parent
	}
}

func currentBinarySHA256() (string, error) {
	path, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("certify external proof: locate current binary: %w", err)
	}
	payload, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("certify external proof: read current binary: %w", err)
	}
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:]), nil
}

func exitCodeForReport(rep certify.Report) int {
	return certify.ExitCodeFor(rep)
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
	fullParity := flags.first("full-parity") == "true"
	writeOnly := flags.first("write-only") == "true"
	directReadOnly := flags.first("direct-read-only") == "true"
	write := flags.first("write") == "true" || fullParity || writeOnly
	full := flags.first("full") == "true" || fullParity || directReadOnly
	if writeOnly && fullParity {
		return certify.Options{}, usageErrorf("--write-only cannot be combined with --full-parity")
	}
	if directReadOnly && write {
		return certify.Options{}, usageErrorf("pm connectors certify --direct-read-only cannot be combined with --write")
	}
	for _, s := range skip {
		if s == "write" {
			if fullParity {
				return certify.Options{}, usageErrorf("--full-parity cannot skip write")
			}
			write = false
		}
	}

	return certify.Options{
		Connector:         connector,
		Stream:            flags.first("stream"),
		Limit:             limit,
		Modes:             parseCSVFlags(flags.values["modes"]),
		Config:            config,
		SecretEnv:         secretEnv,
		KeepWork:          flags.first("keep-workdir") == "true",
		Write:             write,
		WriteOnly:         writeOnly,
		Full:              full,
		RequireFullParity: fullParity,
		DirectReadOnly:    directReadOnly,
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
		sweeper := certify.NewSweeper(certify.SweeperOptions{Root: ledgerRoot, ProjectRoot: root, OlderThan: olderThan})
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
	_, _ = fmt.Fprint(stdout, renderCertifyReportText(rep))
	return nil
}

func renderCertifyReportText(rep certify.Report) string {
	var b strings.Builder
	status := "FAIL"
	if rep.Passed {
		status = "PASS"
	}
	if rep.Mode == "partial_live" {
		status = "PARTIAL"
	}
	fmt.Fprintf(&b, "Legacy certification run: %s [%s]\n", rep.Connector, status)
	fmt.Fprintln(&b, "  This run does not set the generated connector certification status.")
	fmt.Fprintf(&b, "  check:    %s\n", rep.Capabilities.Check.Result)
	fmt.Fprintf(&b, "  catalog:  %s (streams=%d)\n", rep.Capabilities.Catalog.Result, rep.Capabilities.Catalog.Streams)
	fmt.Fprintf(&b, "  read:     %s (stream=%s records=%d)\n", rep.Capabilities.Read.Result, rep.Capabilities.Read.Stream, rep.Capabilities.Read.Records)
	fmt.Fprintf(&b, "  resume:   %s\n", rep.Capabilities.Resume.Result)
	fmt.Fprintf(&b, "  redaction:%s\n", rep.Capabilities.SecretRedaction.Result)
	if directRead := rep.Capabilities.DirectRead; directRead != nil {
		fmt.Fprintf(&b, "  direct-read: %s", directRead.Result)
		if directRead.StagesChecked > 0 {
			fmt.Fprintf(&b, " (candidates=%d", directRead.StagesChecked)
			if directRead.ResumedStages > 0 {
				fmt.Fprintf(&b, " resumed=%d", directRead.ResumedStages)
			}
			fmt.Fprint(&b, ")")
		}
		if directRead.Reason != "" {
			fmt.Fprintf(&b, "; %s", directRead.Reason)
		}
		fmt.Fprintln(&b)
	}
	if surface := rep.Capabilities.Surface; surface != nil {
		fmt.Fprintf(&b, "  surface:  %s", surface.Result)
		if provenance := surface.Provenance; provenance != nil {
			fmt.Fprintf(&b, "; provenance: %s (ledger=%d artifacts=%d endpoints=%d cited=%d)",
				provenance.Status,
				provenance.LedgerVersion,
				provenance.ArtifactCount,
				provenance.EndpointCount,
				provenance.CitedEndpoints,
			)
			if provenance.Reason != "" {
				fmt.Fprintf(&b, "; provenance reason: %s", provenance.Reason)
			}
		}
		fmt.Fprintln(&b)
	}
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
		if strings.HasPrefix(stage.Error, "not_live: ") {
			fmt.Fprintf(&b, "  stage %s: not live: %s\n", stage.Name, strings.TrimPrefix(stage.Error, "not_live: "))
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
	_, _ = fmt.Fprint(stdout, renderBatchMatrixText(batch))
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
	_, _ = fmt.Fprint(stdout, b.String())
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
