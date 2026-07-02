// Stage implementations for the source-stage pipeline (stages 0-11) of the
// certification harness, per docs/architecture/connector-certification-design.md
// §A "Stage list: Source stages" and SPEC.md §1.6. Proven end-to-end against
// the built-in "sample" connector (T/B-14); no CLI wiring — Run is exercised
// directly from Go tests until a later phase adds `pm connectors certify`.
package certify

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// credentialName / warehouseCredentialName / connection names used for every
// certify run. These are internal implementation details of the ephemeral
// workdir, not user-facing configuration.
const (
	sourceCredentialName    = "cert-source"
	warehouseCredentialName = "cert-warehouse"
	fileCredentialName      = "cert-file"
	liveConnectionName      = "cert_live"
	captureConnectionPrefix = "cert_capture_"
)

// noDefsBundleReason is recorded on the fixture_conformance stage: wave0 ships
// zero defs bundles (internal/connectors/defs contains only defs.go until
// Wave F golden migrations land), so every connector -- sample included --
// has no bundle to check yet. This is a deliberate, documented skip: the real
// Tier-0 fixture-conformance integration is a Wave F / V-21 concern and must
// not import internal/connectors/conformance (owned by a parallel task) nor
// internal/connectors/defs (whose tree is empty in wave0 regardless).
const noDefsBundleReason = "skipped: no defs bundle (wave0 ships zero defs bundles; real Tier-0 fixture-conformance integration lands with golden migrations in Wave F / V-21)"

// sabotage lets a test flip the expected envelope kind for a named stage so
// the pipeline can prove a stage failure surfaces correctly end-to-end
// (TEST-PLAN.md §4 "sabotage test -> passed=false, failing stage named").
// It is not used by production code paths.
type sabotage struct {
	stage     string
	wrongKind string
}

// SabotageExpectedKind registers a wrong expected-kind override for the named
// stage on r's next Run call, for self-test purposes only (see
// TestSourceStagesSabotageFailsNamedStage). It is exported for use by
// certify_test but is not part of the certify.Runner's real operating mode.
func SabotageExpectedKind(r *Runner, stage, wrongKind string) {
	r.sabotage = &sabotage{stage: stage, wrongKind: wrongKind}
}

// LastWorkdir returns the ephemeral root directory used by r's most recent
// Run call (for cleanup-verification tests only).
func LastWorkdir(r *Runner) string {
	return r.lastWorkdir
}

// runContext threads the harness, options, secret values, and per-run
// bookkeeping through the stage functions below.
type runContext struct {
	ctx      context.Context
	harness  *Harness
	opts     Options
	sabotage *sabotage
	root     string

	// capturePath is the JSONL file written from the stage-5 live read,
	// replayed through the built-in file connector by stages 6/7/10.
	capturePath           string
	captureFileRegistered bool

	// incremental/resume bookkeeping (stages 8/9).
	incrementalConnection  string
	incrementalRun1Cursor  string
	incrementalRun1Records int
}

// stageFunc executes one certification stage, appending its StageResult to
// rep.Stages and mutating rep.Capabilities as appropriate. Returning an error
// stops the pipeline (used only for unrecoverable setup failures, e.g. an
// unreadable capture file); a stage that merely fails its own assertion
// still returns nil and records Passed=false on the StageResult.
type stageFunc func(rc *runContext, rep *Report) error

// Run executes stages 0-11 (source stages) against exactly one connector in
// a fresh os.MkdirTemp root, mirroring the Makefile "smoke" target's
// --root/--json flag pattern (Makefile:41). See stages_source.go doc comment
// for stage-list scope.
func (r *Runner) Run(ctx context.Context) (Report, error) {
	if r.opts.Connector == "" {
		return Report{}, fmt.Errorf("certify: Options.Connector is required")
	}
	if ctx == nil {
		return Report{}, fmt.Errorf("certify: nil context")
	}

	root, err := os.MkdirTemp("", "pm-certify-"+r.opts.Connector+"-")
	if err != nil {
		return Report{}, fmt.Errorf("certify: create ephemeral workdir: %w", err)
	}
	r.lastWorkdir = root
	if !r.opts.KeepWork {
		defer func() { _ = os.RemoveAll(root) }()
	}

	secretValues := make([]string, 0, len(r.opts.SecretEnv))
	for _, envName := range r.opts.SecretEnv {
		if v := os.Getenv(envName); v != "" {
			secretValues = append(secretValues, v)
		}
	}

	rc := &runContext{
		ctx:      ctx,
		harness:  NewHarness(root, WithSecrets(secretValues...)),
		opts:     r.opts,
		sabotage: r.sabotage,
		root:     root,
	}

	rep := Report{
		Kind:          "ConnectorCertification",
		SchemaVersion: 1,
		Connector:     r.opts.Connector,
		Mode:          "live",
		StartedAt:     time.Now().UTC(),
		Passed:        true,
	}
	rep.Capabilities.SyncModes = map[string]SyncModeResult{}

	stages := []stageFunc{
		stagePreflight,
		stageFixtureConformance,
		stageManualJSON,
		stageCredentialsAdd,
		stageCredentialsTest,
		stageCatalog,
		stageFullRefreshAppend,
		stageFullRefreshOverwrite,
		stageFullRefreshOverwriteDeduped,
		stageIncrementalAppend,
		stageResume,
		stageIncrementalAppendDeduped,
		stageQueryContract,
	}

	for _, stage := range stages {
		if err := stage(rc, &rep); err != nil {
			rep.CompletedAt = time.Now().UTC()
			return rep, err
		}
	}

	finalizeJSONContract(&rep)
	finalizeSecretRedaction(&rep, secretValues)
	rep.CompletedAt = time.Now().UTC()
	rep.Passed = allStagesPassed(rep.Stages)
	return rep, nil
}

// recordStage runs body, timing it, and appends a StageResult to rep.Stages.
// body returns (passed, cliInfo, errMessage); tier is the certification-design
// §"Tiers" table value (0 fixture, 1 replay/capture, 2 live).
func recordStage(rep *Report, name string, tier int, run func() (bool, CLIStageInfo, string)) StageResult {
	start := time.Now()
	passed, cli, errMsg := run()
	stage := StageResult{
		Name:       name,
		Tier:       tier,
		Passed:     passed,
		DurationMS: time.Since(start).Milliseconds(),
		Error:      errMsg,
		CLI:        cli,
	}
	rep.Stages = append(rep.Stages, stage)
	return stage
}

// expectKind returns the expected envelope kind for name, substituting the
// registered sabotage override (self-test only) when it targets this stage.
func (rc *runContext) expectKind(stageName, kind string) string {
	if rc.sabotage != nil && rc.sabotage.stage == stageName {
		return rc.sabotage.wrongKind
	}
	return kind
}

// assertKind runs h.MustKind and converts a mismatch into the (passed, error)
// pair recordStage expects, given an already-obtained CLIResult.
func assertKind(rc *runContext, stageName string, res CLIResult, wantKind string, wantExit int) (bool, string) {
	kind := rc.expectKind(stageName, wantKind)
	if err := rc.harness.MustKind(res, kind, wantExit); err != nil {
		return false, err.Error()
	}
	return true, ""
}

func cliInfoFrom(res CLIResult) CLIStageInfo {
	return CLIStageInfo{ArgvRedacted: res.ArgvRedacted, ExitCode: res.ExitCode, Kind: res.Kind}
}

// streamName is the source stream certified: Options.Stream if set, else
// "customers" (sample's first stream with a cursor field, matching design
// §A command spec's default "first stream with a cursor field, else first").
func (rc *runContext) streamName() string {
	if rc.opts.Stream != "" {
		return rc.opts.Stream
	}
	return "customers"
}

// cursorField is the cursor field name for the certified stream. Wave0 only
// certifies "sample", whose sole cursor-bearing stream is "customers" with
// cursor field "updated_at"; a future multi-connector Runner would resolve
// this from the connector's live Catalog() instead of hardcoding it.
func (rc *runContext) cursorField() string {
	return "updated_at"
}

// queryRowCount runs `pm query run --table <table> --json` and returns the
// row count, used by the overwrite stage to assert truncate semantics.
func queryRowCount(rc *runContext, table string) (int, error) {
	res := rc.harness.Run("query", "run", "--table", table, "--json")
	if res.ExitCode != 0 || res.Kind != "QueryResult" {
		return 0, fmt.Errorf("query --table %s: exit=%d kind=%q", table, res.ExitCode, res.Kind)
	}
	count, _ := res.Envelope["count"].(float64)
	return int(count), nil
}

// assertNoDuplicatePKs runs `pm query run --table <table> --json` and fails
// if any "id" value (the certify-fixed primary key field, see
// createCaptureConnection's --primary-key id) repeats across rows.
func assertNoDuplicatePKs(rc *runContext, table string) error {
	res := rc.harness.Run("query", "run", "--table", table, "--json")
	if res.ExitCode != 0 || res.Kind != "QueryResult" {
		return fmt.Errorf("query --table %s: exit=%d kind=%q", table, res.ExitCode, res.Kind)
	}
	rows, _ := res.Envelope["rows"].([]any)
	seen := map[string]bool{}
	for _, row := range rows {
		m, ok := row.(map[string]any)
		if !ok {
			continue
		}
		id, _ := m["id"].(string)
		if id == "" {
			continue
		}
		if seen[id] {
			return fmt.Errorf("duplicate primary key %q found in table %s (dedup failed)", id, table)
		}
		seen[id] = true
	}
	return nil
}

// --- stage 0: preflight ---

func stagePreflight(rc *runContext, rep *Report) error {
	recordStage(rep, "init", 2, func() (bool, CLIStageInfo, string) {
		res := rc.harness.Run("init", "--json")
		passed, errMsg := assertKind(rc, "init", res, "InitResult", 0)
		return passed, cliInfoFrom(res), errMsg
	})

	recordStage(rep, "preflight", 2, func() (bool, CLIStageInfo, string) {
		res := rc.harness.Run("connectors", "list", "--json")
		passed, errMsg := assertKind(rc, "preflight", res, "ConnectorList", 0)
		if !passed {
			return false, cliInfoFrom(res), errMsg
		}
		found := false
		if list, ok := res.Envelope["connectors"].([]any); ok {
			for _, item := range list {
				m, ok := item.(map[string]any)
				if !ok {
					continue
				}
				if name, _ := m["name"].(string); name == rc.opts.Connector {
					found = true
					break
				}
			}
		}
		if !found {
			return false, cliInfoFrom(res), fmt.Sprintf("preflight: connector %q not present in registry list", rc.opts.Connector)
		}
		if len(rc.opts.SecretEnv) == 0 {
			return true, cliInfoFrom(res), ""
		}
		for field, envName := range rc.opts.SecretEnv {
			if os.Getenv(envName) == "" {
				return false, cliInfoFrom(res), fmt.Sprintf("preflight: secret env %s (field %s) is empty", envName, field)
			}
		}
		return true, cliInfoFrom(res), ""
	})
	return nil
}

// --- stage 1: fixture_conformance (skip-with-reason: wave0 has no bundles) ---

func stageFixtureConformance(_ *runContext, rep *Report) error {
	stage := recordStage(rep, "fixture_conformance", 0, func() (bool, CLIStageInfo, string) {
		return false, CLIStageInfo{}, noDefsBundleReason
	})
	_ = stage
	return nil
}

// --- stage 2: manual_json ---

func stageManualJSON(rc *runContext, rep *Report) error {
	recordStage(rep, "manual_json", 2, func() (bool, CLIStageInfo, string) {
		res := rc.harness.Run("connectors", "inspect", rc.opts.Connector, "--json")
		passed, errMsg := assertKind(rc, "manual_json", res, "Connector", 0)
		if !passed {
			return false, cliInfoFrom(res), errMsg
		}
		if hits := ScanForSecrets(res.Stdout, secretValuesFromEnv(rc.opts.SecretEnv)); len(hits) != 0 {
			return false, cliInfoFrom(res), fmt.Sprintf("manual_json: secret value leaked in connectors inspect output: %v", hits)
		}
		return true, cliInfoFrom(res), ""
	})
	return nil
}

// --- stage 3: credentials add/test ---

func stageCredentialsAdd(rc *runContext, rep *Report) error {
	recordStage(rep, "credentials_add", 2, func() (bool, CLIStageInfo, string) {
		args := []string{"credentials", "add", sourceCredentialName, "--connector", rc.opts.Connector, "--json"}
		for field, envName := range rc.opts.SecretEnv {
			args = append(args, "--from-env", field+"="+envName)
		}
		for k, v := range rc.opts.Config {
			args = append(args, "--config", k+"="+v)
		}
		res := rc.harness.Run(args...)
		passed, errMsg := assertKind(rc, "credentials_add", res, "Credential", 0)
		return passed, cliInfoFrom(res), errMsg
	})

	recordStage(rep, "warehouse_credentials_add", 2, func() (bool, CLIStageInfo, string) {
		warehouseDir := filepath.Join(rc.root, ".polymetrics", "warehouse")
		res := rc.harness.Run("credentials", "add", warehouseCredentialName, "--connector", "warehouse",
			"--config", "path="+warehouseDir, "--json")
		passed, errMsg := assertKind(rc, "warehouse_credentials_add", res, "Credential", 0)
		return passed, cliInfoFrom(res), errMsg
	})
	return nil
}

func stageCredentialsTest(rc *runContext, rep *Report) error {
	recordStage(rep, "credentials_test", 2, func() (bool, CLIStageInfo, string) {
		// Gotcha #5 (design doc §"Load-bearing facts"): live credential
		// validation must go through `pm credentials test`, which resolves
		// vault/secret-backed values, rather than `pm etl check --connector`
		// (which only builds RuntimeConfig from --config and never resolves
		// credential Secrets).
		res := rc.harness.Run("credentials", "test", sourceCredentialName, "--json")
		passed, errMsg := assertKind(rc, "credentials_test", res, "CredentialTest", 0)
		if !passed {
			return false, cliInfoFrom(res), errMsg
		}
		if hits := ScanForSecrets(res.Stdout, secretValuesFromEnv(rc.opts.SecretEnv)); len(hits) != 0 {
			return false, cliInfoFrom(res), fmt.Sprintf("credentials_test: secret value leaked in output: %v", hits)
		}
		rep.Capabilities.Check.Result = "pass"
		return true, cliInfoFrom(res), ""
	})
	return nil
}

// --- stage 4: catalog ---

func stageCatalog(rc *runContext, rep *Report) error {
	stream := rc.streamName()
	recordStage(rep, "connection_create", 2, func() (bool, CLIStageInfo, string) {
		res := rc.harness.Run("connections", "create", liveConnectionName,
			"--source", rc.opts.Connector+":"+sourceCredentialName,
			"--destination", "warehouse:"+warehouseCredentialName,
			"--stream", stream,
			"--primary-key", "id",
			"--cursor", rc.cursorField(),
			"--sync-mode", "full_refresh_append",
			"--table", "cert_live_"+stream,
			"--json")
		passed, errMsg := assertKind(rc, "connection_create", res, "Connection", 0)
		return passed, cliInfoFrom(res), errMsg
	})

	recordStage(rep, "catalog", 2, func() (bool, CLIStageInfo, string) {
		res := rc.harness.Run("catalog", "refresh", "--connection", liveConnectionName, "--json")
		passed, errMsg := assertKind(rc, "catalog", res, "Catalog", 0)
		if !passed {
			return false, cliInfoFrom(res), errMsg
		}
		streams, count := catalogStreams(res.Envelope)
		rep.Capabilities.Catalog.Result = "pass"
		rep.Capabilities.Catalog.Streams = count
		if count < 1 {
			return false, cliInfoFrom(res), "catalog: expected at least one stream"
		}
		if !catalogHasPKAndCursor(streams, stream) {
			return false, cliInfoFrom(res), fmt.Sprintf("catalog: stream %q missing primary_key/cursor_fields", stream)
		}
		return true, cliInfoFrom(res), ""
	})
	return nil
}

func catalogStreams(env map[string]any) ([]any, int) {
	catalogObj, ok := env["catalog"].(map[string]any)
	if !ok {
		return nil, 0
	}
	inner, ok := catalogObj["catalog"].(map[string]any)
	if !ok {
		return nil, 0
	}
	streams, _ := inner["streams"].([]any)
	return streams, len(streams)
}

func catalogHasPKAndCursor(streams []any, name string) bool {
	for _, item := range streams {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if n, _ := m["name"].(string); n != name {
			continue
		}
		pk, _ := m["primary_key"].([]any)
		cursor, _ := m["cursor_fields"].([]any)
		return len(pk) > 0 && len(cursor) > 0
	}
	return false
}

// --- stage 5: etl_full_refresh_append (LIVE) ---

func stageFullRefreshAppend(rc *runContext, rep *Report) error {
	stream := rc.streamName()
	table := "cert_live_" + stream
	var capturePath string

	stage := recordStage(rep, "etl_full_refresh_append", 2, func() (bool, CLIStageInfo, string) {
		res := rc.harness.Run("etl", "run", "--connection", liveConnectionName, "--stream", stream, "--json")
		passed, errMsg := assertKind(rc, "etl_full_refresh_append", res, "ETLRun", 0)
		if !passed {
			return false, cliInfoFrom(res), errMsg
		}
		read, _ := runInt(res.Envelope, "records_read")
		rep.Capabilities.Read.Result = "pass"
		rep.Capabilities.Read.Stream = stream
		rep.Capabilities.Read.Records = read
		rep.Capabilities.SyncModes["full_refresh_append"] = SyncModeResult{Result: "pass", DataSource: "live"}
		if read <= 0 {
			rep.Capabilities.SyncModes["full_refresh_append"] = SyncModeResult{Result: "passed_empty", DataSource: "live", Reason: "records_read was 0"}
		}
		return true, cliInfoFrom(res), ""
	})
	if !stage.Passed {
		return nil
	}

	// The captured JSONL for stages 6/7/10 (capture-replay via the built-in
	// file connector) is this stage's live output, queried back out via
	// `pm query run` and stripped of _polymetrics_* bookkeeping fields.
	capturePath = filepath.Join(rc.root, "capture_"+stream+".jsonl")
	recordStage(rep, "capture_write", 1, func() (bool, CLIStageInfo, string) {
		res := rc.harness.Run("query", "run", "--table", table, "--json")
		passed, errMsg := assertKind(rc, "capture_write", res, "QueryResult", 0)
		if !passed {
			return false, cliInfoFrom(res), errMsg
		}
		rows, _ := res.Envelope["rows"].([]any)
		if err := writeCaptureFile(capturePath, rows); err != nil {
			return false, cliInfoFrom(res), fmt.Sprintf("capture_write: %v", err)
		}
		return true, cliInfoFrom(res), ""
	})
	rc.capturePath = capturePath
	return nil
}

func runInt(env map[string]any, field string) (int, bool) {
	runObj, ok := env["run"].(map[string]any)
	if !ok {
		return 0, false
	}
	v, ok := runObj[field].(float64)
	if !ok {
		return 0, false
	}
	return int(v), true
}

// writeCaptureFile strips _polymetrics_* bookkeeping fields from each row and
// writes the remainder as JSONL, suitable for replay through the built-in
// `file` connector (capture-replay stages 6/7/10).
func writeCaptureFile(path string, rows []any) error {
	var b strings.Builder
	for _, row := range rows {
		m, ok := row.(map[string]any)
		if !ok {
			continue
		}
		clean := map[string]any{}
		for k, v := range m {
			if strings.HasPrefix(k, "_polymetrics") {
				continue
			}
			clean[k] = v
		}
		line, err := json.Marshal(clean)
		if err != nil {
			return fmt.Errorf("marshal capture row: %w", err)
		}
		b.Write(line)
		b.WriteByte('\n')
	}
	return os.WriteFile(path, []byte(b.String()), 0o600)
}

// --- capture-replay connection setup shared by stages 6/7/10 ---

// setupCaptureConnection registers (once per certify run) the cert-file
// credential and a warehouse-destination connection for the given sync mode,
// returning the destination table name.
func (rc *runContext) setupCaptureConnection(rep *Report, mode, table string) (bool, string) {
	if rc.captureFileRegistered {
		return true, ""
	}
	res := rc.harness.Run("credentials", "add", fileCredentialName, "--connector", "file",
		"--config", "path="+rc.capturePath, "--json")
	if passed, errMsg := assertKind(rc, "capture_credentials_add", res, "Credential", 0); !passed {
		recordStage(rep, "capture_credentials_add", 1, func() (bool, CLIStageInfo, string) {
			return false, cliInfoFrom(res), errMsg
		})
		return false, errMsg
	}
	recordStage(rep, "capture_credentials_add", 1, func() (bool, CLIStageInfo, string) {
		return true, cliInfoFrom(res), ""
	})
	rc.captureFileRegistered = true
	return true, ""
}

func captureConnectionName(mode string) string {
	return captureConnectionPrefix + mode
}

func (rc *runContext) createCaptureConnection(rep *Report, stageName, mode, table string) (bool, CLIStageInfo, string) {
	res := rc.harness.Run("connections", "create", captureConnectionName(mode),
		"--source", "file:"+fileCredentialName,
		"--destination", "warehouse:"+warehouseCredentialName,
		"--stream", rc.captureStreamName(),
		"--primary-key", "id",
		"--cursor", rc.cursorField(),
		"--sync-mode", mode,
		"--table", table,
		"--json")
	passed, errMsg := assertKind(rc, stageName, res, "Connection", 0)
	return passed, cliInfoFrom(res), errMsg
}

// captureStreamName is the stream name the built-in `file` connector's
// Catalog synthesizes from the capture file's basename (no extension).
func (rc *runContext) captureStreamName() string {
	base := filepath.Base(rc.capturePath)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

// --- stage 6: etl_full_refresh_overwrite (CAPTURE) ---

func stageFullRefreshOverwrite(rc *runContext, rep *Report) error {
	if rc.capturePath == "" {
		recordStage(rep, "etl_full_refresh_overwrite", 1, func() (bool, CLIStageInfo, string) {
			return false, CLIStageInfo{}, "etl_full_refresh_overwrite: no capture available (etl_full_refresh_append did not produce one)"
		})
		return nil
	}
	table := "cert_overwrite_" + rc.streamName()
	mode := "full_refresh_overwrite"

	if ok, errMsg := rc.setupCaptureConnection(rep, mode, table); !ok {
		recordStage(rep, "etl_full_refresh_overwrite", 1, func() (bool, CLIStageInfo, string) {
			return false, CLIStageInfo{}, errMsg
		})
		return nil
	}

	recordStage(rep, "capture_connection_overwrite", 1, func() (bool, CLIStageInfo, string) {
		return rc.createCaptureConnection(rep, "capture_connection_overwrite", mode, table)
	})

	recordStage(rep, "etl_full_refresh_overwrite", 1, func() (bool, CLIStageInfo, string) {
		// Run twice: overwrite truncate semantics means the row count must
		// stay the same after a second run, not double.
		first := rc.harness.Run("etl", "run", "--connection", captureConnectionName(mode), "--stream", rc.captureStreamName(), "--json")
		if passed, errMsg := assertKind(rc, "etl_full_refresh_overwrite", first, "ETLRun", 0); !passed {
			return false, cliInfoFrom(first), errMsg
		}
		firstCount, err := queryRowCount(rc, table)
		if err != nil {
			return false, cliInfoFrom(first), err.Error()
		}

		second := rc.harness.Run("etl", "run", "--connection", captureConnectionName(mode), "--stream", rc.captureStreamName(), "--json")
		if passed, errMsg := assertKind(rc, "etl_full_refresh_overwrite", second, "ETLRun", 0); !passed {
			return false, cliInfoFrom(second), errMsg
		}
		secondCount, err := queryRowCount(rc, table)
		if err != nil {
			return false, cliInfoFrom(second), err.Error()
		}
		if secondCount != firstCount {
			return false, cliInfoFrom(second), fmt.Sprintf("etl_full_refresh_overwrite: row count changed across overwrite runs (want truncate semantics): run1=%d run2=%d", firstCount, secondCount)
		}
		rep.Capabilities.SyncModes[mode] = SyncModeResult{Result: "pass", DataSource: "capture"}
		return true, cliInfoFrom(second), ""
	})
	return nil
}

// --- stage 7: etl_full_refresh_overwrite_deduped (CAPTURE) ---

func stageFullRefreshOverwriteDeduped(rc *runContext, rep *Report) error {
	if rc.capturePath == "" {
		recordStage(rep, "etl_full_refresh_overwrite_deduped", 1, func() (bool, CLIStageInfo, string) {
			return false, CLIStageInfo{}, "etl_full_refresh_overwrite_deduped: no capture available"
		})
		return nil
	}
	table := "cert_overwrite_deduped_" + rc.streamName()
	mode := "full_refresh_overwrite_deduped"

	if ok, errMsg := rc.setupCaptureConnection(rep, mode, table); !ok {
		recordStage(rep, "etl_full_refresh_overwrite_deduped", 1, func() (bool, CLIStageInfo, string) {
			return false, CLIStageInfo{}, errMsg
		})
		return nil
	}

	recordStage(rep, "capture_connection_overwrite_deduped", 1, func() (bool, CLIStageInfo, string) {
		return rc.createCaptureConnection(rep, "capture_connection_overwrite_deduped", mode, table)
	})

	recordStage(rep, "etl_full_refresh_overwrite_deduped", 1, func() (bool, CLIStageInfo, string) {
		res := rc.harness.Run("etl", "run", "--connection", captureConnectionName(mode), "--stream", rc.captureStreamName(), "--json")
		if passed, errMsg := assertKind(rc, "etl_full_refresh_overwrite_deduped", res, "ETLRun", 0); !passed {
			return false, cliInfoFrom(res), errMsg
		}
		if err := assertNoDuplicatePKs(rc, table); err != nil {
			return false, cliInfoFrom(res), fmt.Sprintf("etl_full_refresh_overwrite_deduped: %v", err)
		}
		rep.Capabilities.SyncModes[mode] = SyncModeResult{Result: "pass", DataSource: "capture"}
		return true, cliInfoFrom(res), ""
	})
	return nil
}

// --- stage 8: etl_incremental_append (LIVE) + stage 9: resume (LIVE run 2) ---

func stageIncrementalAppend(rc *runContext, rep *Report) error {
	stream := rc.streamName()
	table := "cert_incremental_" + stream
	connName := "cert_incremental"

	recordStage(rep, "incremental_connection_create", 2, func() (bool, CLIStageInfo, string) {
		res := rc.harness.Run("connections", "create", connName,
			"--source", rc.opts.Connector+":"+sourceCredentialName,
			"--destination", "warehouse:"+warehouseCredentialName,
			"--stream", stream,
			"--primary-key", "id",
			"--cursor", rc.cursorField(),
			"--sync-mode", "incremental_append",
			"--table", table,
			"--json")
		passed, errMsg := assertKind(rc, "incremental_connection_create", res, "Connection", 0)
		return passed, cliInfoFrom(res), errMsg
	})

	stage := recordStage(rep, "etl_incremental_append", 2, func() (bool, CLIStageInfo, string) {
		res := rc.harness.Run("etl", "run", "--connection", connName, "--stream", stream, "--json")
		passed, errMsg := assertKind(rc, "etl_incremental_append", res, "ETLRun", 0)
		if !passed {
			return false, cliInfoFrom(res), errMsg
		}
		cursor := checkpointString(res.Envelope, "cursor")
		if cursor == "" {
			return false, cliInfoFrom(res), "etl_incremental_append: no cursor recorded on checkpoint"
		}
		rc.incrementalRun1Cursor = cursor
		read1, _ := runInt(res.Envelope, "records_read")
		rc.incrementalRun1Records = read1
		rep.Capabilities.SyncModes["incremental_append"] = SyncModeResult{Result: "pass", DataSource: "live", CursorAdvanced: true}
		return true, cliInfoFrom(res), ""
	})
	if !stage.Passed {
		return nil
	}
	rc.incrementalConnection = connName
	return nil
}

func stageResume(rc *runContext, rep *Report) error {
	if rc.incrementalConnection == "" {
		recordStage(rep, "resume", 2, func() (bool, CLIStageInfo, string) {
			return false, CLIStageInfo{}, "resume: incremental_append did not complete, nothing to resume"
		})
		return nil
	}
	stream := rc.streamName()

	recordStage(rep, "resume", 2, func() (bool, CLIStageInfo, string) {
		res := rc.harness.Run("etl", "run", "--connection", rc.incrementalConnection, "--stream", stream, "--json")
		passed, errMsg := assertKind(rc, "resume", res, "ETLRun", 0)
		if !passed {
			return false, cliInfoFrom(res), errMsg
		}
		cursor2 := checkpointString(res.Envelope, "cursor")
		read2, _ := runInt(res.Envelope, "records_read")
		if read2 > rc.incrementalRun1Records {
			return false, cliInfoFrom(res), fmt.Sprintf("resume: run2 records_read=%d exceeds run1=%d", read2, rc.incrementalRun1Records)
		}
		if compareCursorStrings(cursor2, rc.incrementalRun1Cursor) < 0 {
			return false, cliInfoFrom(res), fmt.Sprintf("resume: cursor regressed run1=%q run2=%q", rc.incrementalRun1Cursor, cursor2)
		}
		rep.Capabilities.Resume.Result = "pass"
		return true, cliInfoFrom(res), ""
	})
	return nil
}

func checkpointString(env map[string]any, field string) string {
	runObj, ok := env["run"].(map[string]any)
	if !ok {
		return ""
	}
	checkpoint, ok := runObj["checkpoint"].(map[string]any)
	if !ok {
		return ""
	}
	v, _ := checkpoint[field].(string)
	return v
}

// compareCursorStrings compares RFC3339 timestamps textually where possible,
// falling back to a plain string comparison. Sample's "updated_at" cursor
// values are RFC3339, so lexicographic comparison already agrees with
// chronological order; this helper documents that assumption for callers
// rather than depending on internal/app's unexported compareCursor.
func compareCursorStrings(a, b string) int {
	switch {
	case a == b:
		return 0
	case a > b:
		return 1
	default:
		return -1
	}
}

// --- stage 10: etl_incremental_append_deduped (CAPTURE) ---

func stageIncrementalAppendDeduped(rc *runContext, rep *Report) error {
	if rc.capturePath == "" {
		recordStage(rep, "etl_incremental_append_deduped", 1, func() (bool, CLIStageInfo, string) {
			return false, CLIStageInfo{}, "etl_incremental_append_deduped: no capture available"
		})
		return nil
	}
	table := "cert_incremental_deduped_" + rc.streamName()
	mode := "incremental_append_deduped"

	if ok, errMsg := rc.setupCaptureConnection(rep, mode, table); !ok {
		recordStage(rep, "etl_incremental_append_deduped", 1, func() (bool, CLIStageInfo, string) {
			return false, CLIStageInfo{}, errMsg
		})
		return nil
	}

	recordStage(rep, "capture_connection_incremental_deduped", 1, func() (bool, CLIStageInfo, string) {
		return rc.createCaptureConnection(rep, "capture_connection_incremental_deduped", mode, table)
	})

	recordStage(rep, "etl_incremental_append_deduped", 1, func() (bool, CLIStageInfo, string) {
		res := rc.harness.Run("etl", "run", "--connection", captureConnectionName(mode), "--stream", rc.captureStreamName(), "--json")
		if passed, errMsg := assertKind(rc, "etl_incremental_append_deduped", res, "ETLRun", 0); !passed {
			return false, cliInfoFrom(res), errMsg
		}
		if err := assertNoDuplicatePKs(rc, table); err != nil {
			return false, cliInfoFrom(res), fmt.Sprintf("etl_incremental_append_deduped: %v", err)
		}
		rep.Capabilities.SyncModes[mode] = SyncModeResult{Result: "pass", DataSource: "capture"}
		return true, cliInfoFrom(res), ""
	})
	return nil
}

// --- stage 11: query_contract ---

func stageQueryContract(rc *runContext, rep *Report) error {
	table := "cert_live_" + rc.streamName()
	recordStage(rep, "query_contract", 2, func() (bool, CLIStageInfo, string) {
		res := rc.harness.Run("query", "run", "--table", table, "--limit", "1", "--json")
		passed, errMsg := assertKind(rc, "query_contract", res, "QueryResult", 0)
		return passed, cliInfoFrom(res), errMsg
	})
	return nil
}

// --- finalization: json_contract + secret_redaction meta-stages ---

func finalizeJSONContract(rep *Report) {
	checked := 0
	allKindsGood := true
	for _, stage := range rep.Stages {
		if stage.CLI.Kind == "" {
			// fixture_conformance is a skip-only stage with no CLI call.
			continue
		}
		checked++
		var mismatch *KindMismatchError
		if stage.Error != "" && isKindMismatch(stage.Error, &mismatch) {
			allKindsGood = false
		}
	}
	rep.Capabilities.JSONContract = CapabilityResult{Result: "pass", StagesChecked: checked}
	if !allKindsGood {
		rep.Capabilities.JSONContract.Result = "fail"
	}
}

// isKindMismatch is a best-effort textual check: recordStage stores only the
// rendered error string (Report is a plain-data JSON type), so this checks
// for the KindMismatchError's rendered shape rather than an errors.As chain.
func isKindMismatch(msg string, _ **KindMismatchError) bool {
	return strings.Contains(msg, "cli result mismatch")
}

func finalizeSecretRedaction(rep *Report, secretValues []string) {
	result := "pass"
	for _, stage := range rep.Stages {
		if hits := ScanForSecrets(stage.CLI.ArgvRedacted, secretValues); len(hits) != 0 {
			result = "fail"
			break
		}
	}
	rep.Capabilities.SecretRedaction = CapabilityResult{Result: result}
}

func allStagesPassed(stages []StageResult) bool {
	for _, s := range stages {
		if s.Name == "fixture_conformance" {
			// A documented skip never fails the overall report.
			continue
		}
		if !s.Passed {
			return false
		}
	}
	return true
}

func secretValuesFromEnv(secretEnv map[string]string) []string {
	values := make([]string, 0, len(secretEnv))
	for _, envName := range secretEnv {
		if v := os.Getenv(envName); v != "" {
			values = append(values, v)
		}
	}
	return values
}
