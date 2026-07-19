// Stage implementations for the create-then-cleanup write protocol (design
// docs/architecture/connector-certification-design.md §A "Write stages":
// 12 write_plan_preview, 13 write_create, 14 write_verify, 15 write_cleanup,
// 16 cleanup_verify, 17 approval_idempotency; §C "Create-then-cleanup write
// protocol").
//
// Two write-pairing sources feed this pipeline:
//
//  1. A curated WritePairing for the connector under test (pairing.go
//     builtinPairings), when the connector exposes a real writes.json action
//     with a safe cleanup (e.g. github's create_label/delete_label). The
//     source-side connection created by stageCatalog (stage 4) is reused for
//     the live read-back in write_verify/cleanup_verify.
//  2. The sample/outbox self-test path (design §C prove-against note: "if
//     no live creds, the stage self-test uses the sample/outbox reverse-ETL
//     path the Makefile smoke target already exercises"), used automatically
//     whenever the connector under test has no curated WritePairing (e.g.
//     "sample" itself, which has no write capability at all). This exercises
//     the exact same write-protocol machinery — plan/preview/create/verify/
//     cleanup/cleanup_verify/ledger/idempotency — against the built-in
//     "outbox" destination connector instead of a live third-party API.
package certify

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const (
	writeOutboxCredentialName = "cert-outbox"
	writeReverseSelfTestName  = "cert_write_selftest"
)

// writeContext carries write-stage bookkeeping across stages 12-17.
type writeContext struct {
	pairing   WritePairing
	connector string
	tag       string
	runID8    string

	planID        string
	approvalToken string

	// selfTest is true when no curated WritePairing exists for the
	// connector under test, so the sample/outbox reverse-ETL path is used
	// instead of the connector's own writes.json actions.
	selfTest bool

	// seedRecordFields names the fields written into write_plan_preview's
	// seeded source table (the live-pairing path only), so
	// stageWriteCleanupLive can seed its own cleanup record consistently
	// without recomputing field ordering.
	seedRecordFields []string

	ledger *Ledger

	previewPassed    bool
	createPassed     bool
	cleanupAttempted bool
}

// approvalTokenLinePattern matches the human-readable "Approval token: X"
// line `pm reverse plan` prints to stdout (Makefile:66 smoke recipe uses the
// identical pattern via awk).
var approvalTokenLinePattern = regexp.MustCompile(`Approval token:\s*(\S+)`)
var planIDLinePattern = regexp.MustCompile(`Created reverse plan (\S+)`)

// --- stage 12: write_plan_preview ---

func stageWritePlanPreview(rc *runContext, rep *Report) error {
	if !rc.opts.Write {
		skipWriteStages(rc, rep, "skipped: write testing disabled (Options.Write is false)")
		return nil
	}

	wc := &writeContext{connector: rc.opts.Connector, runID8: NewRunID8()}
	pairings := PairingsFor(rc.opts.Connector)
	if len(pairings) > 0 {
		wc.pairing = pairings[0]
	} else {
		wc.selfTest = true
		wc.pairing = WritePairing{
			Create:       "create",
			Cleanup:      "delete",
			CleanupKind:  "delete",
			VerifyStream: "records",
			VerifyField:  "tag",
		}
	}
	wc.tag = NewTag(rc.opts.Connector, wc.runID8)

	ledgerRoot := rc.root
	if rc.opts.ArtifactDir != "" {
		durableRoot, dirErr := secureDir(rc.opts.ArtifactDir, certificationsDirName, "ledger", rc.opts.Connector)
		if dirErr != nil {
			recordStage(rc, rep, "write_plan_preview", 2, func() (bool, CLIStageInfo, string) {
				return false, CLIStageInfo{}, fmt.Sprintf("write_plan_preview: create durable ledger root: %v", dirErr)
			})
			return nil
		}
		ledgerRoot = durableRoot
	}
	ledger, err := NewLedger(ledgerRoot)
	if err != nil {
		recordStage(rc, rep, "write_plan_preview", 2, func() (bool, CLIStageInfo, string) {
			return false, CLIStageInfo{}, fmt.Sprintf("write_plan_preview: create ledger: %v", err)
		})
		return nil
	}
	wc.ledger = ledger
	rc.write = wc

	if wc.selfTest {
		stageWritePlanPreviewSelfTest(rc, rep, wc)
		return nil
	}
	stageWritePlanPreviewLive(rc, rep, wc)
	return nil
}

// stageWritePlanPreviewSelfTest drives the plan/preview half of the
// sample/outbox self-test path: register the outbox destination credential,
// create a reverse plan mapping sample_customers rows tagged with wc.tag
// into the outbox, then assert the redaction gate on `reverse preview
// --json`.
func stageWritePlanPreviewSelfTest(rc *runContext, rep *Report, wc *writeContext) {
	recordStage(rc, rep, "write_outbox_credentials_add", 1, func() (bool, CLIStageInfo, string) {
		outboxDir := filepath.Join(rc.root, ".polymetrics", "outbox")
		res := rc.run("credentials", "add", writeOutboxCredentialName, "--connector", "outbox",
			"--config", "path="+outboxDir, "--json")
		passed, errMsg := assertKind(rc, "write_outbox_credentials_add", res, "Credential", 0)
		return passed, cliInfoFrom(res), errMsg
	})

	seedTable := "cert_write_seed_" + rc.opts.Connector
	seedStage := recordStage(rc, rep, "write_seed_table", 1, func() (bool, CLIStageInfo, string) {
		if err := seedTaggedSourceTable(rc, seedTable, wc.tag); err != nil {
			return false, CLIStageInfo{}, fmt.Sprintf("write_seed_table: %v", err)
		}
		return true, CLIStageInfo{}, ""
	})
	if !seedStage.Passed {
		return
	}

	stage := recordStage(rc, rep, "write_plan_preview", 1, func() (bool, CLIStageInfo, string) {
		planRes := rc.run("reverse", "plan", writeReverseSelfTestName,
			"--source-table", seedTable,
			"--destination", "outbox:"+writeOutboxCredentialName,
			"--map", "id:external_id",
			"--map", "tag:tag",
			"--action", wc.pairing.Create)
		if planRes.ExitCode != 0 {
			return false, cliInfoFrom(planRes), fmt.Sprintf("write_plan_preview: reverse plan exit=%d stderr=%s", planRes.ExitCode, planRes.Stderr)
		}
		planID := firstMatch(planIDLinePattern, planRes.Stdout)
		token := firstMatch(approvalTokenLinePattern, planRes.Stdout)
		if planID == "" || token == "" {
			return false, cliInfoFrom(planRes), "write_plan_preview: could not parse plan id/approval token"
		}

		previewRes := rc.run("reverse", "preview", planID, "--json")
		passed, errMsg := assertKind(rc, "write_plan_preview", previewRes, "ReversePlanPreview", 0)
		if !passed {
			return false, cliInfoFrom(previewRes), errMsg
		}
		passed, info, reason := checkPlanPreviewRedaction(previewRes, token)
		if !passed {
			return false, info, reason
		}
		wc.planID = planID
		wc.approvalToken = token
		wc.previewPassed = true
		return true, info, ""
	})
	_ = stage
}

// stageWritePlanPreviewLive drives the plan/preview half of the protocol
// against a real connector's own writes.json action (github's
// create_label/delete_label, etc.). It seeds a dedicated one-row warehouse
// table with the generated record's field values (reverse plan --map only
// renames existing columns, per seedGeneratedSourceTable's doc comment),
// then plans a reverse ETL from that seed table into the connector using
// the source-side credential (sourceCredentialName) stageCredentialsAdd
// already registered.
func stageWritePlanPreviewLive(rc *runContext, rep *Report, wc *writeContext) {
	seedTable := "cert_write_seed_" + rc.opts.Connector
	schema, err := writeActionRecordSchema(rc.opts.Connector, wc.pairing.Create)
	if err != nil {
		recordStage(rc, rep, "write_plan_preview", 2, func() (bool, CLIStageInfo, string) {
			return false, CLIStageInfo{}, fmt.Sprintf("write_plan_preview: %v", err)
		})
		return
	}
	record, err := GenerateRecordWithOverrides(schema, wc.tag, wc.runID8, wc.pairing.Overrides)
	if err != nil {
		recordStage(rc, rep, "write_plan_preview", 2, func() (bool, CLIStageInfo, string) {
			return false, CLIStageInfo{}, fmt.Sprintf("write_plan_preview: generate record: %v", err)
		})
		return
	}
	primaryKey := wc.pairing.IDField
	if _, ok := record[primaryKey]; !ok {
		// The pairing's IDField (e.g. github create_label's "name") is not
		// itself a required field of the CREATE action's own schema in
		// every case; fall back to the tag under that field name so the
		// seed row still has something unique to key on.
		record[primaryKey] = wc.tag
	}
	wc.seedRecordFields = fieldNames(record)

	seedStage := recordStage(rc, rep, "write_seed_table", 2, func() (bool, CLIStageInfo, string) {
		if err := seedGeneratedSourceTable(rc, seedTable, primaryKey, record); err != nil {
			return false, CLIStageInfo{}, fmt.Sprintf("write_seed_table: %v", err)
		}
		return true, CLIStageInfo{}, ""
	})
	if !seedStage.Passed {
		return
	}

	mapArgs := make([]string, 0, len(record)*2)
	for _, field := range wc.seedRecordFields {
		mapArgs = append(mapArgs, "--map", field+":"+field)
	}

	recordStage(rc, rep, "write_plan_preview", 2, func() (bool, CLIStageInfo, string) {
		planRes := rc.run(append([]string{
			"reverse", "plan", writeReverseSelfTestName,
			"--source-table", seedTable,
			"--destination", rc.opts.Connector + ":" + sourceCredentialName,
			"--action", wc.pairing.Create,
		}, mapArgs...)...)
		if planRes.ExitCode != 0 {
			return false, cliInfoFrom(planRes), fmt.Sprintf("write_plan_preview: reverse plan exit=%d stderr=%s", planRes.ExitCode, planRes.Stderr)
		}
		planID := firstMatch(planIDLinePattern, planRes.Stdout)
		token := firstMatch(approvalTokenLinePattern, planRes.Stdout)
		if planID == "" || token == "" {
			return false, cliInfoFrom(planRes), "write_plan_preview: could not parse plan id/approval token"
		}

		previewRes := rc.run("reverse", "preview", planID, "--json")
		passed, errMsg := assertKind(rc, "write_plan_preview", previewRes, "ReversePlanPreview", 0)
		if !passed {
			return false, cliInfoFrom(previewRes), errMsg
		}
		passed, info, reason := checkPlanPreviewRedaction(previewRes, token)
		if !passed {
			return false, info, reason
		}
		wc.planID = planID
		wc.approvalToken = token
		wc.previewPassed = true
		return true, info, ""
	})
}

// fieldNames returns m's keys, sorted for deterministic --map argv ordering
// (stable across runs, easier to read in a recorded stages[].cli.argv
// entry).
func fieldNames(m map[string]any) []string {
	names := make([]string, 0, len(m))
	for k := range m {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}

// checkPlanPreviewRedaction is the stage-12 redaction gate itself (design §A
// "assert --json output has NO approval token"): the parsed envelope's
// plan.approval_token field must be absent/empty, AND — belt and suspenders
// — the raw JSON text must not contain the actual token value verbatim,
// base64, or URL-encoded.
func checkPlanPreviewRedaction(res CLIResult, token string) (bool, CLIStageInfo, string) {
	if plan, ok := res.Envelope["plan"].(map[string]any); ok {
		if v, present := plan["approval_token"]; present {
			if s, _ := v.(string); s != "" {
				return false, cliInfoFrom(res), "write_plan_preview: --json preview leaked a non-empty approval_token field"
			}
		}
	}
	if hits := ScanForSecrets(res.Stdout, []string{token}); len(hits) != 0 {
		return false, cliInfoFrom(res), fmt.Sprintf("write_plan_preview: --json preview output contains sensitive material (%d match)", len(hits))
	}
	return true, cliInfoFrom(res), ""
}

func firstMatch(re *regexp.Regexp, s string) string {
	m := re.FindStringSubmatch(s)
	if len(m) < 2 {
		return ""
	}
	return m[1]
}

// --- stage 13: write_create ---

func stageWriteCreate(rc *runContext, rep *Report) error {
	wc := rc.write
	if wc == nil {
		// write_plan_preview already recorded the skip/failure reason;
		// nothing further to do.
		if !rc.opts.Write {
			skipStage(rc, rep, "write_create", "skipped: write testing disabled (Options.Write is false)")
		} else {
			skipStage(rc, rep, "write_create", "skipped: write_plan_preview did not produce a plan")
		}
		return nil
	}
	if wc.planID == "" || wc.approvalToken == "" || !wc.previewPassed {
		wc.planID = ""
		wc.approvalToken = ""
		skipStage(rc, rep, "write_create", "skipped: write_plan_preview did not complete successfully")
		return nil
	}

	// Write-ahead: record BEFORE any live write is attempted (design §C).
	entityHint := wc.pairing.VerifyStream
	if wc.selfTest {
		entityHint = "outbox_record"
	} else if entityHint == "" {
		entityHint = wc.pairing.Create
	}
	if err := wc.ledger.RecordPlanned(LedgerEntry{
		Action:        wc.pairing.Create,
		CleanupAction: wc.pairing.Cleanup,
		Tag:           wc.tag,
		Connector:     rc.opts.Connector,
		RunID:         wc.runID8,
		EntityHint:    entityHint,
		VerifyField:   wc.pairing.VerifyField,
	}); err != nil {
		recordStage(rc, rep, "write_create", 2, func() (bool, CLIStageInfo, string) {
			return false, CLIStageInfo{}, fmt.Sprintf("write_create: write-ahead ledger record: %v", err)
		})
		return nil
	}

	stage := recordStage(rc, rep, "write_create", 2, func() (bool, CLIStageInfo, string) {
		res := rc.run("reverse", "run", wc.planID, "--approve", wc.approvalToken, "--json")
		passed, errMsg := assertKind(rc, "write_create", res, "ReverseRun", 0)
		if !passed {
			return false, cliInfoFrom(res), errMsg
		}
		run, _ := res.Envelope["run"].(map[string]any)
		succeeded, _ := run["records_succeeded"].(float64)
		failed, _ := run["records_failed"].(float64)
		if succeeded < 1 {
			return false, cliInfoFrom(res), fmt.Sprintf("write_create: records_succeeded=%v, want >=1", run["records_succeeded"])
		}
		if failed != 0 {
			return false, cliInfoFrom(res), fmt.Sprintf("write_create: records_failed=%v, want 0", run["records_failed"])
		}
		return true, cliInfoFrom(res), ""
	})
	wc.createPassed = stage.Passed

	result := "pass"
	reason := ""
	if !stage.Passed {
		result = "fail"
		reason = stage.Error
	}
	if rep.Capabilities.WriteActions == nil {
		rep.Capabilities.WriteActions = map[string]WriteActionResult{}
	}
	rep.Capabilities.WriteActions[wc.pairing.Create] = WriteActionResult{
		Result:  result,
		Cleanup: wc.pairing.Cleanup,
		Tag:     wc.tag,
		Reason:  reason,
	}
	// design §C failure semantics: "create fails -> stage fail, no leak" —
	// deliberately no ledger.RecordCleaned and no Leaks append here.
	return nil
}

// --- stage 14: write_verify (read-back) ---

func stageWriteVerify(rc *runContext, rep *Report) error {
	wc := rc.write
	if wc == nil || !wc.createPassed {
		skipStage(rc, rep, "write_verify", "skipped: write_create did not succeed, nothing to verify")
		return nil
	}

	if wc.selfTest {
		recordStage(rc, rep, "write_verify", 1, func() (bool, CLIStageInfo, string) {
			found, err := outboxRecordTagPresent(rc.root, wc.tag)
			if err != nil {
				return false, CLIStageInfo{}, fmt.Sprintf("write_verify: read outbox records: %v", err)
			}
			if !found {
				// design §A stage 14: "else unverified warning" — a
				// missing read-back is a warning, not a hard stage
				// failure, since some connectors/backends have no
				// reliable read-your-write guarantee.
				markWriteActionUnverified(rep, wc.pairing.Create, "write_verify: tag not found in outbox read-back")
				return true, CLIStageInfo{}, ""
			}
			return true, CLIStageInfo{}, ""
		})
		return nil
	}

	recordStage(rc, rep, "write_verify_connection_create", 2, func() (bool, CLIStageInfo, string) {
		res := rc.run("connections", "create", "cert_write_verify",
			"--source", rc.opts.Connector+":"+sourceCredentialName,
			"--destination", "warehouse:"+warehouseCredentialName,
			"--stream", wc.pairing.VerifyStream,
			"--primary-key", wc.pairing.IDField,
			"--sync-mode", "full_refresh_overwrite",
			"--table", "cert_write_verify",
			"--json")
		passed, errMsg := assertKind(rc, "write_verify_connection_create", res, "Connection", 0)
		return passed, cliInfoFrom(res), errMsg
	})

	recordStage(rc, rep, "write_verify", 2, func() (bool, CLIStageInfo, string) {
		res := rc.run("etl", "run", "--connection", "cert_write_verify", "--stream", wc.pairing.VerifyStream, "--json")
		passed, errMsg := assertKind(rc, "write_verify", res, "ETLRun", 0)
		if !passed {
			return false, cliInfoFrom(res), errMsg
		}
		found, ferr := queryFieldContainsValue(rc, "cert_write_verify", wc.pairing.VerifyField, wc.tag)
		if ferr != nil {
			return false, cliInfoFrom(res), fmt.Sprintf("write_verify: %v", ferr)
		}
		if !found {
			markWriteActionUnverified(rep, wc.pairing.Create, "write_verify: tag not found in live read-back of "+wc.pairing.VerifyStream)
		}
		return true, cliInfoFrom(res), ""
	})
	return nil
}

func markWriteActionUnverified(rep *Report, action, reason string) {
	if rep.Capabilities.WriteActions == nil {
		return
	}
	entry := rep.Capabilities.WriteActions[action]
	entry.Verify = "unverified"
	if entry.Reason == "" {
		entry.Reason = reason
	}
	rep.Capabilities.WriteActions[action] = entry
}

// outboxRecordTagPresent reads the outbox table stageWritePlanPreviewSelfTest
// wrote into and reports whether any record's "tag" field equals tag.
func outboxRecordTagPresent(root, tag string) (bool, error) {
	rows, err := readOutboxJSONL(root)
	if err != nil {
		return false, err
	}
	for _, row := range rows {
		if v, _ := row["tag"].(string); v == tag {
			return true, nil
		}
	}
	return false, nil
}

// outboxLastActionForTag returns the most recently appended record's
// "_outbox_action" value among rows whose "tag" field equals tag, so
// cleanup_verify can tell a tombstone (delete-shaped) append from the
// original create append in an append-only backend (design §C cleanup
// mechanics adapted for outbox's self-test path).
func outboxLastActionForTag(root, tag string) (string, error) {
	rows, err := readOutboxJSONL(root)
	if err != nil {
		return "", err
	}
	last := ""
	for _, row := range rows {
		if v, _ := row["tag"].(string); v != tag {
			continue
		}
		if action, _ := row["_outbox_action"].(string); action != "" {
			last = action
		}
	}
	return last, nil
}

// seedTaggedSourceTable materializes a dedicated one-row warehouse table
// named table containing exactly {"id": tag, "tag": tag}, via the built-in
// file connector's normal ETL path (rather than hand-writing warehouse
// JSONL directly), so a `reverse plan --map tag:tag` has an actual tag
// STRING VALUE to carry into the destination — `reverse plan`'s --map only
// renames existing columns 1:1 (app.mapReverseRecords), it cannot inject a
// constant, so certify must seed a source row that already contains the tag
// under a mappable field name.
func seedTaggedSourceTable(rc *runContext, table, tag string) error {
	return seedGeneratedSourceTable(rc, table, "id", map[string]any{"id": tag, "tag": tag})
}

// seedGeneratedSourceTable materializes a dedicated one-row warehouse table
// named table containing exactly record, via the built-in file connector's
// normal ETL path (rather than hand-writing warehouse JSONL directly), so a
// `reverse plan --map` has actual field VALUES to carry into the
// destination — `reverse plan`'s --map only renames existing columns 1:1
// (app.mapReverseRecords), it cannot inject a constant, so certify must seed
// a source row that already contains every generated field under its own
// name. primaryKey must be a key present in record.
func seedGeneratedSourceTable(rc *runContext, table, primaryKey string, record map[string]any) error {
	if _, ok := record[primaryKey]; !ok {
		return fmt.Errorf("seed record missing primary key field %q: %+v", primaryKey, record)
	}
	seedPath := filepath.Join(rc.root, table+"_seed.jsonl")
	if err := writeCaptureFile(seedPath, []any{record}); err != nil {
		return fmt.Errorf("write seed file: %w", err)
	}
	credName := "cert-write-seed-file-" + table
	credRes := rc.run("credentials", "add", credName, "--connector", "file", "--config", "path="+seedPath, "--json")
	if credRes.ExitCode != 0 {
		return fmt.Errorf("seed credentials add: exit=%d stderr=%s", credRes.ExitCode, credRes.Stderr)
	}
	streamName := strings.TrimSuffix(filepath.Base(seedPath), filepath.Ext(seedPath))
	connName := "cert_write_seed_conn_" + table
	connRes := rc.run("connections", "create", connName,
		"--source", "file:"+credName,
		"--destination", "warehouse:"+warehouseCredentialName,
		"--stream", streamName,
		"--primary-key", primaryKey,
		"--sync-mode", "full_refresh_overwrite",
		"--table", table,
		"--json")
	if connRes.ExitCode != 0 {
		return fmt.Errorf("seed connection create: exit=%d stderr=%s", connRes.ExitCode, connRes.Stderr)
	}
	runRes := rc.run("etl", "run", "--connection", connName, "--stream", streamName, "--json")
	if runRes.ExitCode != 0 || runRes.Kind != "ETLRun" {
		return fmt.Errorf("seed etl run: exit=%d kind=%q", runRes.ExitCode, runRes.Kind)
	}
	return nil
}

func readOutboxJSONL(root string) ([]map[string]any, error) {
	// The outbox destination table name is the reverse plan's OWN Name
	// (app.RunReverseETL: Write{Table: plan.Name}), not a fixed "records"
	// table — write_create/write_cleanup both plan under
	// writeReverseSelfTestName so their writes land in this one file.
	path := filepath.Join(root, ".polymetrics", "outbox", writeReverseSelfTestName+".jsonl")
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var rows []map[string]any
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var row map[string]any
		if err := json.Unmarshal([]byte(line), &row); err != nil {
			return nil, err
		}
		rows = append(rows, row)
	}
	return rows, nil
}

// queryFieldContainsValue runs `pm query run --table <table> --json` and
// reports whether any row's field equals value.
func queryFieldContainsValue(rc *runContext, table, field, value string) (bool, error) {
	res := rc.run("query", "run", "--table", table, "--json")
	if res.ExitCode != 0 || res.Kind != "QueryResult" {
		return false, fmt.Errorf("query --table %s: exit=%d kind=%q", table, res.ExitCode, res.Kind)
	}
	rows, _ := res.Envelope["rows"].([]any)
	for _, row := range rows {
		m, ok := row.(map[string]any)
		if !ok {
			continue
		}
		if v, _ := m[field].(string); v == value {
			return true, nil
		}
	}
	return false, nil
}

// --- stage 15: write_cleanup ---

func stageWriteCleanup(rc *runContext, rep *Report) error {
	wc := rc.write
	if wc == nil || !wc.createPassed {
		skipStage(rc, rep, "write_cleanup", "skipped: write_create did not succeed, nothing to clean up")
		return nil
	}
	wc.cleanupAttempted = true

	var cleanupPassed bool
	if wc.selfTest {
		cleanupPassed = stageWriteCleanupSelfTest(rc, rep, wc)
	} else {
		cleanupPassed = stageWriteCleanupLive(rc, rep, wc)
	}
	// design §C failure semantics: "Create ok + cleanup/verify fails ->
	// leaked_resource" — a failed cleanup CLI call itself (not just a later
	// cleanup_verify miss) already means certify cannot prove the tagged
	// entity was removed, so the leak must be recorded here rather than
	// waiting for stageCleanupVerify (which would only catch cases where
	// the cleanup call reported success but the entity is still present).
	if !cleanupPassed {
		recordLeak(rep, wc, "write_cleanup: cleanup CLI call failed, entity may still exist")
	}
	return nil
}

func stageWriteCleanupSelfTest(rc *runContext, rep *Report, wc *writeContext) bool {
	seedTable := "cert_write_seed_" + rc.opts.Connector
	stage := recordStage(rc, rep, "write_cleanup", 1, func() (bool, CLIStageInfo, string) {
		// The reverse plan's Name determines the outbox destination table
		// (app.RunReverseETL: Write{Table: plan.Name}) — cleanup MUST reuse
		// the exact same plan name as write_create's plan so both the
		// create and the tombstone append land in the SAME outbox table,
		// letting outboxLastActionForTag see both and tell them apart.
		planRes := rc.run("reverse", "plan", writeReverseSelfTestName,
			"--source-table", seedTable,
			"--destination", "outbox:"+writeOutboxCredentialName,
			"--map", "id:external_id",
			"--map", "tag:tag",
			"--action", wc.pairing.Cleanup)
		if planRes.ExitCode != 0 {
			return false, cliInfoFrom(planRes), fmt.Sprintf("write_cleanup: reverse plan exit=%d stderr=%s", planRes.ExitCode, planRes.Stderr)
		}
		planID := firstMatch(planIDLinePattern, planRes.Stdout)
		token := firstMatch(approvalTokenLinePattern, planRes.Stdout)
		if planID == "" || token == "" {
			return false, cliInfoFrom(planRes), "write_cleanup: could not parse cleanup plan id/approval token"
		}
		return runApprovedReverse(rc, "write_cleanup", planID, token)
	})

	return stage.Passed
}

func stageWriteCleanupLive(rc *runContext, rep *Report, wc *writeContext) bool {
	// Build the cleanup record: at minimum the IDField identifying which
	// entity to clean (the same value write_create's created entity used),
	// plus any other required fields the Cleanup action's own schema
	// declares — generated fresh so a required-but-not-identifying field
	// (rare for delete/close actions, but not impossible) is still present.
	record := map[string]any{wc.pairing.IDField: wc.tag}
	if schema, err := writeActionRecordSchema(rc.opts.Connector, wc.pairing.Cleanup); err == nil {
		if generated, gerr := GenerateRecordWithOverrides(schema, wc.tag, wc.runID8, wc.pairing.Overrides); gerr == nil {
			for k, v := range generated {
				record[k] = v
			}
		}
	}

	seedTable := "cert_write_seed_cleanup_" + rc.opts.Connector
	seedStage := recordStage(rc, rep, "write_cleanup_seed_table", 2, func() (bool, CLIStageInfo, string) {
		if err := seedGeneratedSourceTable(rc, seedTable, wc.pairing.IDField, record); err != nil {
			return false, CLIStageInfo{}, fmt.Sprintf("write_cleanup_seed_table: %v", err)
		}
		return true, CLIStageInfo{}, ""
	})
	if !seedStage.Passed {
		return false
	}

	mapArgs := make([]string, 0, len(record)*2)
	for _, field := range fieldNames(record) {
		mapArgs = append(mapArgs, "--map", field+":"+field)
	}

	stage := recordStage(rc, rep, "write_cleanup", 2, func() (bool, CLIStageInfo, string) {
		planRes := rc.run(append([]string{
			"reverse", "plan", writeReverseSelfTestName + "_cleanup",
			"--source-table", seedTable,
			"--destination", rc.opts.Connector + ":" + sourceCredentialName,
			"--action", wc.pairing.Cleanup,
		}, mapArgs...)...)
		if planRes.ExitCode != 0 {
			return false, cliInfoFrom(planRes), fmt.Sprintf("write_cleanup: reverse plan exit=%d stderr=%s", planRes.ExitCode, planRes.Stderr)
		}
		planID := firstMatch(planIDLinePattern, planRes.Stdout)
		token := firstMatch(approvalTokenLinePattern, planRes.Stdout)
		if planID == "" || token == "" {
			return false, cliInfoFrom(planRes), "write_cleanup: could not parse cleanup plan id/approval token"
		}
		return runApprovedReverse(rc, "write_cleanup", planID, token)
	})

	return stage.Passed
}

func runApprovedReverse(rc *runContext, stageName, planID, token string) (bool, CLIStageInfo, string) {
	previewRes := rc.run("reverse", "preview", planID, "--json")
	passed, errMsg := assertKind(rc, stageName, previewRes, "ReversePlanPreview", 0)
	if !passed {
		return false, cliInfoFrom(previewRes), errMsg
	}
	passed, info, reason := checkPlanPreviewRedaction(previewRes, token)
	if !passed {
		return false, info, reason
	}
	runRes := rc.run("reverse", "run", planID, "--approve", token, "--json")
	passed, errMsg = assertKind(rc, stageName, runRes, "ReverseRun", 0)
	return passed, cliInfoFrom(runRes), errMsg
}

// --- stage 16: cleanup_verify ---

func stageCleanupVerify(rc *runContext, rep *Report) error {
	wc := rc.write
	if wc == nil || !wc.createPassed {
		skipStage(rc, rep, "cleanup_verify", "skipped: write_create did not succeed, nothing to verify cleaned")
		return nil
	}

	gone, err := cleanupEntityGone(rc, wc)
	if err != nil {
		recordStage(rc, rep, "cleanup_verify", tierFor(wc), func() (bool, CLIStageInfo, string) {
			return false, CLIStageInfo{}, fmt.Sprintf("cleanup_verify: %v", err)
		})
		recordLeak(rep, wc, "cleanup_verify: could not determine entity state: "+err.Error())
		return nil
	}

	recordStage(rc, rep, "cleanup_verify", tierFor(wc), func() (bool, CLIStageInfo, string) {
		if !gone {
			return false, CLIStageInfo{}, "cleanup_verify: entity still present after cleanup (leaked_resource)"
		}
		return true, CLIStageInfo{}, ""
	})

	if !gone {
		recordLeak(rep, wc, "cleanup_verify: entity still present after cleanup")
		return nil
	}

	cleanupCallFailed := latestStageFailed(rep.Stages, "write_cleanup")
	if cleanupCallFailed {
		removeLeakForWrite(rep, wc)
		if entry, ok := rep.Capabilities.WriteActions[wc.pairing.Create]; ok {
			entry.Result = "fail"
			entry.Reason = "write_cleanup failed, but cleanup_verify proved the entity absent"
			rep.Capabilities.WriteActions[wc.pairing.Create] = entry
		}
	}

	if err := wc.ledger.RecordCleaned(wc.tag); err != nil {
		recordStage(rc, rep, "cleanup_verify_ledger", tierFor(wc), func() (bool, CLIStageInfo, string) {
			return false, CLIStageInfo{}, fmt.Sprintf("cleanup_verify: record verified cleanup in ledger: %v", err)
		})
		return nil
	}

	if entry, ok := rep.Capabilities.WriteActions[wc.pairing.Create]; ok {
		if cleanupCallFailed {
			entry.Result = "fail"
			entry.Reason = "write_cleanup failed, but cleanup_verify proved the entity absent"
		} else {
			entry.Result = "pass"
			if entry.Verify == "" {
				entry.Verify = "read_back"
			}
		}
		rep.Capabilities.WriteActions[wc.pairing.Create] = entry
	}
	return nil
}

func tierFor(wc *writeContext) int {
	if wc.selfTest {
		return 1
	}
	return 2
}

// cleanupEntityGone determines whether the tagged entity is gone after
// cleanup. For the self-test path (append-only outbox), "gone" is
// interpreted as "the last record appended for this tag reflects the
// cleanup action, not the original create action" — outbox never actually
// deletes rows, so this is the closest available analogue, exactly proving
// the write-protocol MACHINERY (ledger + stage sequencing + verification)
// without depending on real delete semantics. cleanupVerifySabotage forces
// this to report "still looks like the original create" for self-tests only
// (SabotageCleanupVerifyEntityStillPresent).
func cleanupEntityGone(rc *runContext, wc *writeContext) (bool, error) {
	if wc.selfTest {
		if rc.cleanupVerifySabotage {
			return false, nil
		}
		lastAction, err := outboxLastActionForTag(rc.root, wc.tag)
		if err != nil {
			return false, err
		}
		return lastAction == wc.pairing.Cleanup, nil
	}

	if rc.cleanupVerifySabotage {
		return false, nil
	}
	// Re-run the verify connection's ETL so cert_write_verify reflects the
	// POST-cleanup state — write_verify (stage 14) populated this table
	// once, before cleanup ran, so querying it without refreshing first
	// would only ever see stale pre-cleanup data (always "still present").
	refreshRes := rc.run("etl", "run", "--connection", "cert_write_verify", "--stream", wc.pairing.VerifyStream, "--json")
	if refreshRes.ExitCode != 0 || refreshRes.Kind != "ETLRun" {
		return false, fmt.Errorf("cleanup_verify: refresh cert_write_verify: exit=%d kind=%q", refreshRes.ExitCode, refreshRes.Kind)
	}
	found, err := queryFieldContainsValue(rc, "cert_write_verify", wc.pairing.VerifyField, wc.tag)
	if err != nil {
		return false, err
	}
	return !found, nil
}

func recordLeak(rep *Report, wc *writeContext, reason string) {
	rep.Leaks = append(rep.Leaks, Leak{
		Tag:       wc.tag,
		Connector: wc.connector,
		Action:    wc.pairing.Create,
		Reason:    reason,
	})
	if rep.Capabilities.WriteActions == nil {
		rep.Capabilities.WriteActions = map[string]WriteActionResult{}
	}
	entry := rep.Capabilities.WriteActions[wc.pairing.Create]
	entry.Result = "leaked_resource"
	entry.Reason = reason
	rep.Capabilities.WriteActions[wc.pairing.Create] = entry
}

func removeLeakForWrite(rep *Report, wc *writeContext) {
	kept := rep.Leaks[:0]
	for _, leak := range rep.Leaks {
		if leak.Tag == wc.tag && leak.Connector == wc.connector && leak.Action == wc.pairing.Create {
			continue
		}
		kept = append(kept, leak)
	}
	rep.Leaks = kept
}

func latestStageFailed(stages []StageResult, name string) bool {
	for i := len(stages) - 1; i >= 0; i-- {
		if stages[i].Name == name {
			return !stages[i].Passed
		}
	}
	return false
}

// --- stage 17: approval_idempotency ---

func stageApprovalIdempotency(rc *runContext, rep *Report) error {
	wc := rc.write
	if wc == nil || !wc.createPassed || wc.planID == "" || wc.approvalToken == "" {
		skipStage(rc, rep, "approval_idempotency", "skipped: no consumed plan/token available to replay")
		return nil
	}

	recordStage(rc, rep, "approval_idempotency", tierFor(wc), func() (bool, CLIStageInfo, string) {
		res := rc.run("reverse", "run", wc.planID, "--approve", wc.approvalToken, "--json")
		if res.ExitCode == 0 {
			return false, cliInfoFrom(res), "approval_idempotency: replaying a consumed plan+token succeeded, want rejection"
		}
		if res.Kind != "Error" {
			return false, cliInfoFrom(res), fmt.Sprintf("approval_idempotency: replay kind=%q exit=%d, want an Error envelope", res.Kind, res.ExitCode)
		}
		return true, cliInfoFrom(res), ""
	})
	return nil
}

// --- shared skip helpers ---

// skipWriteStages records write_plan_preview's own skip AND leaves rc.write
// nil so every subsequent write stage (13-17) can cheaply detect "nothing to
// do here" and record its own scoped skip reason.
func skipWriteStages(rc *runContext, rep *Report, reason string) {
	skipStage(rc, rep, "write_plan_preview", reason)
}

// skipStage records a documented, non-failing skip for name (the same
// convention stageFixtureConformance established: Passed=false with a
// "skipped: ..." Error, exempted from allStagesPassed by name).
func skipStage(rc *runContext, rep *Report, name, reason string) {
	recordStage(rc, rep, name, 2, func() (bool, CLIStageInfo, string) {
		return false, CLIStageInfo{}, reason
	})
}

// writeActionRecordSchema looks up the record_schema for actionName among
// connector's curated writes.json (currently only github's is embedded here
// via builtinWriteSchemas; a future connector.DefinitionProvider-based
// lookup can replace this once more connectors expose Definition()).
func writeActionRecordSchema(connector, actionName string) ([]byte, error) {
	schemas, ok := builtinWriteSchemas[connector]
	if !ok {
		return nil, fmt.Errorf("no curated record_schema available for connector %q", connector)
	}
	schema, ok := schemas[actionName]
	if !ok {
		return nil, fmt.Errorf("no curated record_schema available for %q action %q", connector, actionName)
	}
	return schema, nil
}

// builtinWriteSchemas holds the minimal record_schema JSON for each curated
// WritePairing's Create/Cleanup action (mirroring the corresponding
// writes.json entries under internal/connectors/defs/<connector>/writes.json,
// design §C "Data generation from write action record_schema").
var builtinWriteSchemas = map[string]map[string][]byte{
	"github": {
		"create_label": []byte(`{
			"type": "object",
			"required": ["name"],
			"properties": {"name": {"type": "string"}, "color": {"type": "string"}, "description": {"type": "string"}}
		}`),
		"delete_label": []byte(`{
			"type": "object",
			"required": ["name"],
			"properties": {"name": {"type": "string"}}
		}`),
		"create_issue": []byte(`{
			"type": "object",
			"required": ["title"],
			"properties": {"title": {"type": "string"}, "body": {"type": "string"}}
		}`),
		"close_issue": []byte(`{
			"type": "object",
			"required": ["issue_number"],
			"properties": {"issue_number": {"type": "integer"}}
		}`),
		"create_milestone": []byte(`{
			"type": "object",
			"required": ["title"],
			"properties": {"title": {"type": "string"}}
		}`),
		"delete_milestone": []byte(`{
			"type": "object",
			"required": ["milestone_number"],
			"properties": {"milestone_number": {"type": "integer"}}
		}`),
	},
}
