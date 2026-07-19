// Package certify: sweeper.go implements `pm connectors certify --sweep`
// (design docs/architecture/connector-certification-design.md §C "Orphan
// sweeper: ledger entries without cleaned_at + optional live scan of
// VerifyStreams for aged pm-cert-<slug>-* tags; cleanup through the same
// plan/approve/run path"). CI batch jobs run this as a trailing step; a
// nightly sweep job backs it up.
package certify

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// MaxSweepAge is the largest orphan age accepted by any sweep entrypoint.
const MaxSweepAge = 365 * 24 * time.Hour

// SweeperOptions configures a Sweeper.
type SweeperOptions struct {
	// Root is the legacy combined root. New callers set LedgerRoot and
	// WorkspaceRoot separately so durable authority never doubles as CLI state.
	Root          string
	LedgerRoot    string
	WorkspaceRoot string
	Connector     string
	Credential    *CredentialRef
	// OlderThan is the minimum age (measured from LedgerEntry.PlannedAt) an
	// uncleaned entry must have before the sweeper will touch it
	// (certification design §A/§B "--older-than 24h" default), so an
	// in-flight certify run's own not-yet-cleaned entries are never swept
	// out from under it.
	OlderThan time.Duration
}

// SweepResult summarizes one Sweep call.
type SweepResult struct {
	// Scanned is the total number of ledger entries examined.
	Scanned int
	// Cleaned lists the tags the sweeper successfully cleaned and marked
	// RecordCleaned this pass.
	Cleaned []string
	// Skipped lists tags left alone this pass (too recent, or no known
	// cleanup mechanism), each with a reason.
	Skipped map[string]string
	// Failed lists tags whose cleanup was attempted but did not succeed,
	// each with the failure reason — these remain uncleaned in the ledger
	// for a future sweep pass.
	Failed map[string]string
}

// Sweeper drives orphan cleanup for aged, unledgered-as-cleaned certify write
// records.
type Sweeper struct {
	opts SweeperOptions
}

// NewSweeper constructs a Sweeper for opts.
func NewSweeper(opts SweeperOptions) *Sweeper {
	return &Sweeper{opts: opts}
}

// Sweep loads the ledger under s.opts.Root, finds every uncleaned entry
// older than s.opts.OlderThan, attempts cleanup for each via the same
// CLI-driven mechanics the write stages use, and records success back into
// the ledger.
func (s *Sweeper) Sweep(ctx context.Context) (SweepResult, error) {
	if s.opts.OlderThan <= 0 || s.opts.OlderThan > MaxSweepAge {
		return SweepResult{}, fmt.Errorf("certify: sweep age must be greater than zero and no more than 8760h")
	}
	if s.opts.Credential != nil && s.opts.Connector == "" {
		return SweepResult{}, fmt.Errorf("certify: sweep credential reference requires a selected connector")
	}
	if s.opts.Connector != "" {
		ref := CredentialRef{}
		if s.opts.Credential != nil {
			ref = *s.opts.Credential
		}
		if err := ValidateCredentialReference(s.opts.Connector, ref); err != nil {
			return SweepResult{}, err
		}
	}
	ledgerRoot := s.opts.LedgerRoot
	if ledgerRoot == "" {
		ledgerRoot = s.opts.Root
	}
	workspaceRoot := s.opts.WorkspaceRoot
	if workspaceRoot == "" {
		workspaceRoot = s.opts.Root
	}
	entries, err := LoadLedger(ledgerRoot)
	if err != nil {
		return SweepResult{}, fmt.Errorf("certify: sweeper load ledger: %w", err)
	}

	ledger, err := NewLedger(ledgerRoot)
	if err != nil {
		return SweepResult{}, fmt.Errorf("certify: sweeper open ledger: %w", err)
	}

	result := SweepResult{
		Scanned: len(entries.All()),
		Skipped: map[string]string{},
		Failed:  map[string]string{},
	}

	threshold := time.Now().UTC().Add(-s.opts.OlderThan)
	harness := NewHarness(workspaceRoot, WithContext(ctx))
	prepared := false

	for _, status := range entries.Uncleaned() {
		if status.PlannedAt.After(threshold) {
			result.Skipped[status.Tag] = "not yet aged past --older-than threshold"
			continue
		}
		if err := ctx.Err(); err != nil {
			return result, err
		}
		if s.opts.Connector != "" && status.Connector != s.opts.Connector {
			result.Failed[status.Tag] = "ledger connector does not match selected sweep connector"
			continue
		}
		if reason := validateSweepStatus(status); reason != "" {
			result.Failed[status.Tag] = reason
			continue
		}
		if !prepared {
			if err := prepareSweepWorkspace(harness, status.Connector, s.opts.Credential); err != nil {
				return result, err
			}
			prepared = true
		}

		ok, reason := sweepCleanTag(harness, status)
		if !ok {
			result.Failed[status.Tag] = reason
			continue
		}
		if err := ledger.RecordCleaned(status.Tag); err != nil {
			result.Failed[status.Tag] = fmt.Sprintf("cleaned but failed to record in ledger: %v", err)
			continue
		}
		result.Cleaned = append(result.Cleaned, status.Tag)
	}

	return result, nil
}

// sweepCleanTag attempts to clean up one uncleaned ledger entry, dispatching
// on its EntityHint/Connector shape. Two mechanisms are supported today:
//
//   - "outbox_record" entity hint (the sample/outbox self-test write path,
//     stages_write.go): append a tombstone record via a fresh reverse-ETL
//     plan/run against the outbox connector, mirroring
//     stageWriteCleanupSelfTest.
//   - a connector with a curated WritePairing (pairing.go builtinPairings)
//     whose Create action matches status.Action: run that pairing's real
//     Cleanup action against the connector, mirroring
//     stageWriteCleanupLive. This requires the connector's source
//     credential (sourceCredentialName) to already exist under root, which
//     is true whenever the ledger entry came from a real certify run
//     against that connector in this same project.
//
// Any other shape is left alone (not cleaned) with a reason recorded in
// result.Skipped/Failed by the caller, since certify must never guess at an
// unsafe cleanup mechanism for an unrecognized entry.
func sweepCleanTag(h *Harness, status LedgerStatus) (bool, string) {
	if reason := validateSweepStatus(status); reason != "" {
		return false, reason
	}
	if status.EntityHint == "outbox_record" {
		if status.Action != "create" {
			return false, "outbox cleanup authority does not match the recorded action"
		}
		return sweepCleanOutboxRecord(h, status)
	}

	pairing, ok := matchPairingForAction(status.Connector, status.Action)
	if !ok {
		return false, fmt.Sprintf("no known cleanup mechanism for connector %q action %q", status.Connector, status.Action)
	}
	if status.CleanupAction != "" && status.CleanupAction != pairing.Cleanup {
		return false, "ledger cleanup action does not match the curated pairing"
	}
	if status.EntityHint != pairing.VerifyStream || (status.VerifyField != "" && status.VerifyField != pairing.VerifyField) {
		return false, "ledger verification provenance does not match the curated pairing"
	}
	return sweepCleanViaPairing(h, status, pairing)
}

func validateSweepStatus(status LedgerStatus) string {
	if status.Connector == "" || status.Action == "" || status.Tag == "" || status.RunID == "" || status.PlannedAt.IsZero() {
		return "ledger entry is missing required cleanup provenance"
	}
	if !validLedgerTag(status.Connector, status.RunID, status.Tag) {
		return "ledger tag is not bound to the recorded connector"
	}
	return ""
}

func prepareSweepWorkspace(h *Harness, connector string, credential *CredentialRef) error {
	initRes := h.Run("init", "--json")
	if initRes.ExitCode != 0 || initRes.Kind != "InitResult" {
		return fmt.Errorf("certify: initialize sweep workspace: exit=%d kind=%q", initRes.ExitCode, initRes.Kind)
	}
	if credential == nil || connector == "sample" {
		return nil
	}
	args := []string{"credentials", "add", sourceCredentialName, "--connector", connector, "--json"}
	fields := make([]string, 0, len(credential.FromEnv))
	for field := range credential.FromEnv {
		fields = append(fields, field)
	}
	sort.Strings(fields)
	for _, field := range fields {
		args = append(args, "--from-env", field+"="+credential.FromEnv[field])
	}
	keys := make([]string, 0, len(credential.Config))
	for key := range credential.Config {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		args = append(args, "--config", key+"="+credential.Config[key])
	}
	res := h.Run(args...)
	if res.ExitCode != 0 || res.Kind != "Credential" {
		return fmt.Errorf("certify: provision sweep credential reference: exit=%d kind=%q", res.ExitCode, res.Kind)
	}
	return nil
}

func matchPairingForAction(connector, action string) (WritePairing, bool) {
	for _, p := range PairingsFor(connector) {
		if p.Create == action {
			return p, true
		}
	}
	return WritePairing{}, false
}

// sweepCleanOutboxRecord appends a tombstone record for status.Tag to the
// outbox, provisioning the outbox credential/table idempotently first
// (credentials add is itself idempotent-by-name at the CLI layer — a
// pre-existing credential of the same name is simply reused).
func sweepCleanOutboxRecord(h *Harness, status LedgerStatus) (bool, string) {
	outboxDir := filepath.Join(h.root, ".polymetrics", "outbox")
	credRes := h.Run("credentials", "add", writeOutboxCredentialName, "--connector", "outbox",
		"--config", "path="+outboxDir, "--json")
	if credRes.ExitCode != 0 && credRes.Kind != "Credential" {
		return false, fmt.Sprintf("provision outbox credential: exit=%d kind=%q", credRes.ExitCode, credRes.Kind)
	}

	// A sweeper run has no live source table to plan a reverse ETL from (the
	// original certify run's ephemeral workdir is long gone), so it writes
	// the tombstone directly through the outbox connector's own Write path
	// via a minimal one-record file->outbox reverse plan sourced from a
	// throwaway single-row table this call creates. The plan MUST reuse
	// writeReverseSelfTestName (not a tag-derived name) since the outbox
	// destination table is keyed by the reverse plan's own Name
	// (app.RunReverseETL: Write{Table: plan.Name}) — this is the same table
	// stages_write.go's write_create/write_cleanup wrote the original
	// create/cleanup records into.
	sourceTable, err := tombstoneSourceTable(h, status)
	if err != nil {
		return false, fmt.Sprintf("prepare sweep tombstone source: %v", err)
	}
	planRes := h.Run("reverse", "plan", writeReverseSelfTestName,
		"--source-table", sourceTable,
		"--destination", "outbox:"+writeOutboxCredentialName,
		"--map", "tag:tag",
		"--action", "delete")
	if planRes.ExitCode != 0 {
		return false, fmt.Sprintf("sweep plan: exit=%d kind=%q", planRes.ExitCode, planRes.Kind)
	}
	planID := firstMatch(planIDLinePattern, planRes.Stdout)
	token := firstMatch(approvalTokenLinePattern, planRes.Stdout)
	if planID == "" || token == "" {
		return false, "sweep: could not parse plan id/approval token"
	}
	if ok, reason := sweepPreviewPlan(h, planID, token); !ok {
		return false, reason
	}
	runRes := h.Run("reverse", "run", planID, "--approve", token, "--json")
	if runRes.ExitCode != 0 || runRes.Kind != "ReverseRun" {
		return false, fmt.Sprintf("sweep cleanup run: exit=%d kind=%q", runRes.ExitCode, runRes.Kind)
	}
	return true, ""
}

// tombstoneSourceTable ensures a one-row warehouse table containing exactly
// status.Tag exists (so the reverse plan above has something to read),
// returning its name. This keeps the sweeper self-contained: it never
// depends on the original certify run's ephemeral capture files still
// existing on disk.
func tombstoneSourceTable(h *Harness, status LedgerStatus) (string, error) {
	table := "cert_sweep_source"
	seedPath := filepath.Join(h.root, "cert_sweep_seed.jsonl")
	if err := writeSweepSeedFile(seedPath, status.Tag); err != nil {
		return "", err
	}
	invocations := [][]string{
		{"credentials", "add", "cert-sweep-warehouse", "--connector", "warehouse", "--config", "path=" + filepath.Join(h.root, ".polymetrics", "warehouse"), "--json"},
		{"credentials", "add", "cert-sweep-seed-file", "--connector", "file", "--config", "path=" + seedPath, "--json"},
		{"connections", "create", "cert_sweep_seed_conn", "--source", "file:cert-sweep-seed-file", "--destination", "warehouse:cert-sweep-warehouse", "--stream", "cert_sweep_seed", "--primary-key", "tag", "--sync-mode", "full_refresh_overwrite", "--table", table, "--json"},
		{"etl", "run", "--connection", "cert_sweep_seed_conn", "--stream", "cert_sweep_seed", "--json"},
	}
	for _, invocation := range invocations {
		res := h.Run(invocation...)
		if res.ExitCode != 0 {
			return "", fmt.Errorf("%s: exit=%d kind=%q", strings.Join(invocation[:2], " "), res.ExitCode, res.Kind)
		}
	}
	return table, nil
}

func writeSweepSeedFile(path, tag string) error {
	return writeCaptureFile(path, []any{map[string]any{"tag": tag}})
}

// sweepCleanViaPairing runs pairing's real Cleanup action against a
// connector that already has a live source credential registered in this
// project (sourceCredentialName), mirroring stageWriteCleanupLive.
func sweepCleanViaPairing(h *Harness, status LedgerStatus, pairing WritePairing) (bool, string) {
	table := "cert_sweep_" + status.Connector
	fields, err := seedSweepCleanupTable(h, table, status, pairing)
	if err != nil {
		return false, fmt.Sprintf("sweep seed cleanup table: %v", err)
	}
	planArgs := []string{"reverse", "plan", "cert_sweep_cleanup_" + status.RunID,
		"--source-table", table,
		"--destination", status.Connector + ":" + sourceCredentialName,
		"--action", pairing.Cleanup}
	for _, field := range fields {
		planArgs = append(planArgs, "--map", field+":"+field)
	}
	planRes := h.Run(planArgs...)
	if planRes.ExitCode != 0 {
		return false, fmt.Sprintf("sweep plan: exit=%d kind=%q", planRes.ExitCode, planRes.Kind)
	}
	planID := firstMatch(planIDLinePattern, planRes.Stdout)
	token := firstMatch(approvalTokenLinePattern, planRes.Stdout)
	if planID == "" || token == "" {
		return false, "sweep: could not parse plan id/approval token"
	}
	if ok, reason := sweepPreviewPlan(h, planID, token); !ok {
		return false, reason
	}
	runRes := h.Run("reverse", "run", planID, "--approve", token, "--json")
	if runRes.ExitCode != 0 || runRes.Kind != "ReverseRun" {
		return false, fmt.Sprintf("sweep cleanup run: exit=%d kind=%q", runRes.ExitCode, runRes.Kind)
	}
	return true, ""
}

func seedSweepCleanupTable(h *Harness, table string, status LedgerStatus, pairing WritePairing) ([]string, error) {
	record := map[string]any{pairing.IDField: status.Tag}
	if schema, err := writeActionRecordSchema(status.Connector, pairing.Cleanup); err == nil {
		if generated, err := GenerateRecordWithOverrides(schema, status.Tag, status.RunID, pairing.Overrides); err == nil {
			for key, value := range generated {
				record[key] = value
			}
		}
	}
	seedPath := filepath.Join(h.root, table+".jsonl")
	if err := writeCaptureFile(seedPath, []any{record}); err != nil {
		return nil, err
	}
	warehouse := filepath.Join(h.root, ".polymetrics", "warehouse")
	for _, invocation := range [][]string{
		{"credentials", "add", "cert-sweep-warehouse", "--connector", "warehouse", "--config", "path=" + warehouse, "--json"},
		{"credentials", "add", "cert-sweep-file", "--connector", "file", "--config", "path=" + seedPath, "--json"},
		{"connections", "create", "cert_sweep_cleanup_conn", "--source", "file:cert-sweep-file", "--destination", "warehouse:cert-sweep-warehouse", "--stream", table, "--primary-key", pairing.IDField, "--sync-mode", "full_refresh_overwrite", "--table", table, "--json"},
		{"etl", "run", "--connection", "cert_sweep_cleanup_conn", "--stream", table, "--json"},
	} {
		res := h.Run(invocation...)
		if res.ExitCode != 0 {
			return nil, fmt.Errorf("%s: exit=%d", strings.Join(invocation[:2], " "), res.ExitCode)
		}
	}
	return fieldNames(record), nil
}

func sweepPreviewPlan(h *Harness, planID, token string) (bool, string) {
	previewRes := h.Run("reverse", "preview", planID, "--json")
	if err := h.MustKind(previewRes, "ReversePlanPreview", 0); err != nil {
		return false, fmt.Sprintf("sweep preview: %v", err)
	}
	if hits := ScanForSecrets(previewRes.Stdout+previewRes.Stderr, []string{token}); len(hits) != 0 {
		return false, fmt.Sprintf("sweep preview contains sensitive material (%d match)", len(hits))
	}
	return true, ""
}
