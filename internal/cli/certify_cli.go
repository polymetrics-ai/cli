package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
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

func (defaultCertifyCommandRuntime) RunSingle(ctx context.Context, root string, opts certify.Options) (certify.Report, error) {
	opts.ArtifactDir = filepath.Join(root, ".polymetrics")
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
	var (
		creds certify.CredsFile
		names []string
		err   error
	)
	if credsPath != "" {
		creds, err = certify.LoadCredsFile(credsPath)
		if err != nil {
			return nil, err
		}
		if err := certify.ValidateBatchConstraints(creds); err != nil {
			return nil, usageErrorf("%v", err)
		}
		names = creds.ConnectorNames()
	} else {
		names, err = sweepTargetConnectors(root, "")
		if err != nil {
			return nil, err
		}
	}
	if len(names) == 0 {
		return nil, usageErrorf("pm connectors certify --sweep found no ledger to sweep under %s (pass --credentials-file, or run a live/batch certify first)", root)
	}
	results := make(map[string]certify.SweepResult, len(names))
	for _, name := range names {
		workspace, err := os.MkdirTemp("", "pm-certify-sweep-"+name+"-")
		if err != nil {
			return nil, fmt.Errorf("certify: create sweep workspace: %w", err)
		}
		ledgerRoot, err := certify.DurableLedgerRoot(root, name)
		if err != nil {
			_ = os.RemoveAll(workspace)
			return nil, fmt.Errorf("certify: validate sweep ledger root: %w", err)
		}
		var credential *certify.CredentialRef
		if entry, ok := creds.Connectors[name]; ok {
			ref := entry.Credential
			credential = &ref
		}
		result, sweepErr := certify.NewSweeper(certify.SweeperOptions{
			LedgerRoot:    ledgerRoot,
			WorkspaceRoot: workspace,
			Connector:     name,
			Credential:    credential,
			OlderThan:     olderThan,
		}).Sweep(ctx)
		_ = os.RemoveAll(workspace)
		if sweepErr != nil {
			return nil, fmt.Errorf("certify: sweep %s: %w", name, sweepErr)
		}
		results[name] = result
	}
	return results, nil
}

func newCertifyCobraCommand(ctx context.Context, root string, stdout io.Writer, jsonOut bool, runtime certifyCommandRuntime) *cobra.Command {
	var flags certifyCommandFlags
	cmd := newConnectorsActionCobraCommand("certify [connector]", func(_ *cobra.Command, args []string) error {
		if err := validateCertifyBooleanFlags(flags); err != nil {
			return markCobraLegacyError(err)
		}
		if firstArgIsHelp(args) {
			return markCobraLegacyError(writeManual("connectors", stdout, jsonOut))
		}
		switch {
		case lastString(flags.Sweeps) == "true" && lastString(flags.Alls) == "true":
			return usageErrorf("--all and --sweep are not supported together")
		case lastString(flags.Sweeps) == "true":
			if err := validateCertifySweepFlags(flags); err != nil {
				return markCobraLegacyError(err)
			}
			if err := validateCertifySweepAgeFlag(flags); err != nil {
				return markCobraLegacyError(err)
			}
			return markCobraLegacyError(runCertifySweep(ctx, root, flags, stdout, jsonOut, runtime))
		case lastString(flags.Alls) == "true":
			if err := validateCertifyBatchFlags(flags); err != nil {
				return markCobraLegacyError(err)
			}
			if err := validateCertifyParallelFlag(flags); err != nil {
				return markCobraLegacyError(err)
			}
			return markCobraLegacyError(runCertifyBatch(ctx, root, flags, stdout, jsonOut, runtime))
		default:
			if len(args) != 1 {
				return usageErrorf("pm connectors certify <connector> | --all --credentials-file <file> | --sweep")
			}
			if err := validateCertifySingleFlags(flags); err != nil {
				return markCobraLegacyError(err)
			}
			// Preserve connector-validation precedence (and its existing span)
			// while keeping valid-connector credential controls ahead of telemetry.
			if safety.ValidateIdentifier(args[0], "connector") == nil {
				if _, err := certifyOptionsFromFlags(args[0], flags); err != nil {
					return markCobraLegacyError(err)
				}
			}
			return markCobraLegacyError(runCertifySingle(ctx, root, args[0], flags, stdout, jsonOut, runtime))
		}
	})
	// Certification can load credentials and perform writes. Unlike the legacy
	// connector inspection actions, it must fail closed on every unknown flag
	// so a typo cannot silently bypass a safety control.
	cmd.FParseErrWhitelist.UnknownFlags = false
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
	if saveErr := rep.Save(saveDir); saveErr != nil {
		if certify.ExitCodeFor(rep) != 3 {
			return fmt.Errorf("certify: persist certification report: %w", saveErr)
		}
		rep.Stages = append(rep.Stages, certify.StageResult{
			Name:   "report_persistence",
			Tier:   0,
			Passed: false,
			Error:  "report persistence failed; leaked resources still require cleanup",
		})
	}

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
	if err := certify.ValidateCredentialReference(connector, certify.CredentialRef{FromEnv: secretEnv, Config: config}); err != nil {
		return certify.Options{}, usageErrorf("%v", err)
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

func prevalidateCertifySafetyArgs(args []string) error {
	return prevalidateCertifyArgs(args, true)
}

func prevalidateCertifySyntaxArgs(args []string) error {
	return prevalidateCertifyArgs(args, false)
}

func prevalidateCertifyArgs(args []string, validateSweepCreds bool) error {
	if len(args) < 2 || args[0] != "connectors" || args[1] != "certify" {
		return nil
	}

	if err := validateCertifyRequiredValueArgs(args, 2); err != nil {
		return err
	}

	args = normalizeStringArraySpaceValues(args, 2, certifyFlagNames)
	boolNames := map[string]bool{
		"keep-workdir": true, "write": true, "full": true,
		"all": true, "resume": true, "sweep": true,
	}
	known := map[string]bool{
		"from-env": true, "config": true, "stream": true, "skip": true,
		"keep-workdir": true, "write": true, "full": true, "all": true,
		"credentials-file": true, "parallel": true, "resume": true,
		"sweep": true, "older-than": true, "help": true,
	}
	unsupported := map[string]bool{
		"credential": true, "limit": true, "modes": true, "record": true,
		"replay": true, "allow-production-writes": true, "rate-limit": true,
		"budget": true, "live-all-modes": true,
	}

	var (
		sweepEnabled bool
		credsPath    string
	)
	for _, arg := range args[2:] {
		if arg == "--" {
			break
		}
		if arg == "-h" {
			continue
		}
		if !strings.HasPrefix(arg, "-") {
			continue
		}
		if !strings.HasPrefix(arg, "--") || arg == "--" {
			return usageErrorf("unknown shorthand flag: %s", arg)
		}
		nameValue := strings.TrimPrefix(arg, "--")
		name, value, assigned := strings.Cut(nameValue, "=")
		if name == "" || strings.HasPrefix(name, "-") {
			return usageErrorf("unknown flag: %s", arg)
		}
		if unsupported[name] {
			return usageErrorf("--%s is not supported; refusing to run certification with a no-op control", name)
		}
		if !known[name] {
			return usageErrorf("unknown flag: --%s", name)
		}
		if !assigned {
			value = "true"
		}
		if boolNames[name] && value != "true" && value != "false" {
			return usageErrorf("--%s must be true or false, got %q", name, value)
		}
		switch name {
		case "parallel":
			parallel, err := strconv.Atoi(value)
			if err != nil {
				return validationErrorf("invalid --parallel %q, want integer", value)
			}
			if parallel < 1 || parallel > certify.MaxParallel {
				return usageErrorf("--parallel must be between 1 and %d", certify.MaxParallel)
			}
		case "older-than":
			age, err := time.ParseDuration(value)
			if err != nil || age <= 0 || age > maxCertifySweepAge {
				return usageErrorf("--older-than must be greater than zero and no more than 8760h")
			}
		case "sweep":
			sweepEnabled = value == "true"
		case "credentials-file":
			credsPath = value
		}
	}

	if validateSweepCreds && sweepEnabled && credsPath != "" && credsPath != "true" {
		creds, err := certify.LoadCredsFile(credsPath)
		if err != nil {
			return usageErrorf("%v", err)
		}
		if err := certify.ValidateBatchConstraints(creds); err != nil {
			return usageErrorf("%v", err)
		}
	}
	return nil
}

var certifyRequiredValueFlagNames = map[string]struct{}{
	"from-env":         {},
	"config":           {},
	"stream":           {},
	"credentials-file": {},
	"parallel":         {},
	"older-than":       {},
}

func validateCertifyRequiredValueArgs(args []string, start int) error {
	for i := start; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			break
		}
		if !strings.HasPrefix(arg, "--") || arg == "--" {
			continue
		}
		nameValue := strings.TrimPrefix(arg, "--")
		name, value, assigned := strings.Cut(nameValue, "=")
		if _, required := certifyRequiredValueFlagNames[name]; !required {
			continue
		}
		if assigned {
			if strings.TrimSpace(value) == "" {
				return usageErrorf("--%s requires a value", name)
			}
			continue
		}
		if i+1 >= len(args) || certifyValueArgMissing(name, args[i+1]) {
			return usageErrorf("--%s requires a value", name)
		}
		i++
	}
	return nil
}

func certifyValueArgMissing(name, next string) bool {
	if next == "--" || next == "-h" || strings.HasPrefix(next, "--") {
		return true
	}
	if strings.HasPrefix(next, "-") && name != "parallel" && name != "older-than" {
		return true
	}
	return false
}

func validateCertifyParallelFlag(flags certifyCommandFlags) error {
	if len(flags.Parallels) == 0 {
		return nil
	}
	parallel, err := parseIntFlag("parallel", lastString(flags.Parallels), 0)
	if err != nil {
		return err
	}
	if parallel < 1 || parallel > certify.MaxParallel {
		return usageErrorf("--parallel must be between 1 and %d", certify.MaxParallel)
	}
	return nil
}

func validateCertifySweepAgeFlag(flags certifyCommandFlags) error {
	raw := lastString(flags.OlderThans)
	if raw == "" {
		return nil
	}
	age, err := time.ParseDuration(raw)
	if err != nil {
		return usageErrorf("invalid --older-than %q: %v", raw, err)
	}
	if age <= 0 || age > maxCertifySweepAge {
		return usageErrorf("--older-than must be greater than zero and no more than 8760h")
	}
	return nil
}

func validateCertifyBooleanFlags(flags certifyCommandFlags) error {
	for _, control := range []struct {
		name   string
		values []string
	}{
		{"keep-workdir", flags.KeepWorkdirs},
		{"write", flags.Writes},
		{"full", flags.Fulls},
		{"all", flags.Alls},
		{"resume", flags.Resumes},
		{"sweep", flags.Sweeps},
	} {
		for _, value := range control.values {
			if value != "true" && value != "false" {
				return usageErrorf("--%s must be true or false, got %q", control.name, value)
			}
		}
	}
	return nil
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

	if err := validateCertifyParallelFlag(flags); err != nil {
		return err
	}
	parallel, err := parseIntFlag("parallel", lastString(flags.Parallels), 0)
	if err != nil {
		return err
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
	if len(flags.Parallels) > 0 {
		credsFile.Defaults.Parallel = parallel
	}

	batch, err := runtime.RunBatch(ctx, root, credsFile, lastString(flags.Resumes) == "true")
	if err != nil {
		if len(batch.Results) > 0 && batch.ExitCode != 0 {
			if err := writeCertifyBatchReport(stdout, jsonOut, batch); err != nil {
				return err
			}
			return exitForBatch(batch)
		}
		return fmt.Errorf("certify: batch run failed: %w", err)
	}

	if err := writeCertifyBatchReport(stdout, jsonOut, batch); err != nil {
		return err
	}

	return exitForBatch(batch)
}

// --- sweep mode (--sweep) ---

const maxCertifySweepAge = certify.MaxSweepAge

func runCertifySweep(ctx context.Context, root string, flags certifyCommandFlags, stdout io.Writer, jsonOut bool, runtime certifyCommandRuntime) error {
	if err := validateCertifySweepFlags(flags); err != nil {
		return err
	}

	if err := validateCertifySweepAgeFlag(flags); err != nil {
		return err
	}
	olderThan := 24 * time.Hour
	if raw := lastString(flags.OlderThans); raw != "" {
		olderThan, _ = time.ParseDuration(raw) // validated above
	}
	if olderThan <= 0 || olderThan > maxCertifySweepAge {
		return usageErrorf("--older-than must be greater than zero and no more than 8760h")
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
	ledgerRoot, err := certify.DurableLedgerBase(root)
	if err != nil {
		return nil, fmt.Errorf("certify: validate sweep ledger base: %w", err)
	}
	names, err := listSubdirNames(ledgerRoot)
	if err != nil {
		return nil, err
	}
	registry := appRegistry()
	for _, name := range names {
		if err := safety.ValidateIdentifier(name, "connector"); err != nil {
			return nil, validationErrorf("invalid durable ledger connector: %v", err)
		}
		if _, ok := registry.Get(name); !ok {
			return nil, validationErrorf("durable ledger connector %q is not registered locally", name)
		}
	}
	return names, nil
}

// listSubdirNames returns the names of dir's immediate subdirectories, or an
// empty slice if dir doesn't exist (never an error — an absent ledger root
// just means "nothing to sweep yet").
func listSubdirNames(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("certify: read durable ledger base: %w", err)
	}
	var names []string
	for _, entry := range entries {
		if entry.Type()&os.ModeSymlink != 0 {
			return nil, fmt.Errorf("certify: durable ledger connector %q must not be a symlink", entry.Name())
		}
		if entry.IsDir() {
			names = append(names, entry.Name())
		}
	}
	return names, nil
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
