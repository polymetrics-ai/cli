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
	"time"
)

// SweeperOptions configures a Sweeper.
type SweeperOptions struct {
	// Root is the project root (an existing `pm init`-ed directory) whose
	// certify-ledger.jsonl the sweeper scans and whose CLI surface it drives
	// to perform cleanup.
	Root string
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
	entries, err := LoadLedger(s.opts.Root)
	if err != nil {
		return SweepResult{}, fmt.Errorf("certify: sweeper load ledger: %w", err)
	}

	ledger, err := NewLedger(s.opts.Root)
	if err != nil {
		return SweepResult{}, fmt.Errorf("certify: sweeper open ledger: %w", err)
	}

	result := SweepResult{
		Scanned: len(entries.All()),
		Skipped: map[string]string{},
		Failed:  map[string]string{},
	}

	threshold := time.Now().UTC().Add(-s.opts.OlderThan)
	harness := NewHarness(s.opts.Root)

	for _, status := range entries.Uncleaned() {
		if status.PlannedAt.After(threshold) {
			result.Skipped[status.Tag] = "not yet aged past --older-than threshold"
			continue
		}
		if err := ctx.Err(); err != nil {
			return result, err
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
	if status.EntityHint == "outbox_record" {
		return sweepCleanOutboxRecord(h, status)
	}

	pairing, ok := matchPairingForAction(status.Connector, status.Action)
	if !ok {
		return false, fmt.Sprintf("no known cleanup mechanism for connector %q action %q", status.Connector, status.Action)
	}
	return sweepCleanViaPairing(h, status, pairing)
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
		return false, fmt.Sprintf("provision outbox credential: exit=%d stderr=%s", credRes.ExitCode, credRes.Stderr)
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
	planRes := h.Run("reverse", "plan", writeReverseSelfTestName,
		"--source-table", tombstoneSourceTable(h, status),
		"--destination", "outbox:"+writeOutboxCredentialName,
		"--map", "tag:tag",
		"--action", "delete")
	if planRes.ExitCode != 0 {
		return false, fmt.Sprintf("sweep plan: exit=%d stderr=%s", planRes.ExitCode, planRes.Stderr)
	}
	planID := firstMatch(planIDLinePattern, planRes.Stdout)
	token := firstMatch(approvalTokenLinePattern, planRes.Stdout)
	if planID == "" || token == "" {
		return false, "sweep: could not parse plan id/approval token"
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
func tombstoneSourceTable(h *Harness, status LedgerStatus) string {
	table := "cert_sweep_source"
	_ = h.Run("credentials", "add", "cert-sweep-warehouse", "--connector", "warehouse",
		"--config", "path="+filepath.Join(h.root, ".polymetrics", "warehouse"), "--json")
	// Seed the row via the file connector so the write goes through the
	// normal ETL path rather than hand-writing warehouse JSONL directly.
	seedPath := filepath.Join(h.root, "cert_sweep_seed.jsonl")
	_ = writeSweepSeedFile(seedPath, status.Tag)
	_ = h.Run("credentials", "add", "cert-sweep-seed-file", "--connector", "file",
		"--config", "path="+seedPath, "--json")
	_ = h.Run("connections", "create", "cert_sweep_seed_conn",
		"--source", "file:cert-sweep-seed-file",
		"--destination", "warehouse:cert-sweep-warehouse",
		"--stream", "cert_sweep_seed",
		"--primary-key", "tag",
		"--sync-mode", "full_refresh_overwrite",
		"--table", table,
		"--json")
	_ = h.Run("etl", "run", "--connection", "cert_sweep_seed_conn", "--stream", "cert_sweep_seed", "--json")
	return table
}

func writeSweepSeedFile(path, tag string) error {
	return writeCaptureFile(path, []any{map[string]any{"tag": tag}})
}

// sweepCleanViaPairing runs pairing's real Cleanup action against a
// connector that already has a live source credential registered in this
// project (sourceCredentialName), mirroring stageWriteCleanupLive.
func sweepCleanViaPairing(h *Harness, status LedgerStatus, pairing WritePairing) (bool, string) {
	table := "cert_sweep_" + status.Connector
	planRes := h.Run("reverse", "plan", "cert_sweep_cleanup_"+status.Tag,
		"--source-table", table,
		"--destination", status.Connector+":"+sourceCredentialName,
		"--map", pairing.IDField+":"+pairing.IDField,
		"--action", pairing.Cleanup)
	if planRes.ExitCode != 0 {
		return false, fmt.Sprintf("sweep plan: exit=%d stderr=%s", planRes.ExitCode, planRes.Stderr)
	}
	planID := firstMatch(planIDLinePattern, planRes.Stdout)
	token := firstMatch(approvalTokenLinePattern, planRes.Stdout)
	if planID == "" || token == "" {
		return false, "sweep: could not parse plan id/approval token"
	}
	runRes := h.Run("reverse", "run", planID, "--approve", token, "--json")
	if runRes.ExitCode != 0 || runRes.Kind != "ReverseRun" {
		return false, fmt.Sprintf("sweep cleanup run: exit=%d kind=%q", runRes.ExitCode, runRes.Kind)
	}
	return true, ""
}
