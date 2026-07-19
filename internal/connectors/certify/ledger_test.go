package certify_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors/certify"
)

// TestLedgerRecordPlannedWritesBeforeCreate proves the write-ahead leak
// ledger (design §C "before any live write, append {action, tag, connector,
// entity_hint, planned_at} to certify-ledger.jsonl"): RecordPlanned must
// persist an entry to disk synchronously, so a crash immediately afterward
// still leaves a durable trail naming the tag.
func TestLedgerRecordPlannedWritesBeforeCreate(t *testing.T) {
	dir := t.TempDir()
	ledger, err := certify.NewLedger(dir)
	if err != nil {
		t.Fatalf("NewLedger() error = %v", err)
	}

	if err := ledger.RecordPlanned(certify.LedgerEntry{
		Action:     "create_label",
		Tag:        "pm-cert-github-ab12cd34-1751450000",
		Connector:  "github",
		EntityHint: "label",
	}); err != nil {
		t.Fatalf("RecordPlanned() error = %v", err)
	}

	path := filepath.Join(dir, "certify-ledger.jsonl")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read ledger file: %v", err)
	}
	if !strings.Contains(string(raw), "pm-cert-github-ab12cd34-1751450000") {
		t.Errorf("ledger file missing planted tag: %s", raw)
	}
	if !strings.Contains(string(raw), `"planned_at"`) {
		t.Errorf("ledger entry missing planned_at: %s", raw)
	}
	if strings.Contains(string(raw), `"cleaned_at"`) {
		t.Errorf("ledger entry has cleaned_at before cleanup ran: %s", raw)
	}
}

// TestLedgerRecordCleanedMarksEntry proves that after a verified cleanup,
// RecordCleaned appends a {tag, cleaned_at} record the sweeper/loader can
// join back to the original planned entry (design §C "after verified
// cleanup append {tag, cleaned_at}").
func TestLedgerRecordCleanedMarksEntry(t *testing.T) {
	dir := t.TempDir()
	ledger, err := certify.NewLedger(dir)
	if err != nil {
		t.Fatalf("NewLedger() error = %v", err)
	}
	tag := "pm-cert-github-ab12cd34-1751450000"
	if err := ledger.RecordPlanned(certify.LedgerEntry{Action: "create_label", Tag: tag, Connector: "github"}); err != nil {
		t.Fatalf("RecordPlanned() error = %v", err)
	}
	if err := ledger.RecordCleaned(tag); err != nil {
		t.Fatalf("RecordCleaned() error = %v", err)
	}

	entries, err := certify.LoadLedger(dir)
	if err != nil {
		t.Fatalf("LoadLedger() error = %v", err)
	}
	status, ok := entries.StatusFor(tag)
	if !ok {
		t.Fatalf("StatusFor(%q) not found in loaded ledger", tag)
	}
	if !status.Cleaned {
		t.Errorf("StatusFor(%q).Cleaned = false, want true after RecordCleaned", tag)
	}
}

// TestLoadLedgerUncleanedEntries proves the sweeper's core query: entries
// with a planned_at but no cleaned_at are the ones that need attention
// (design §C "Orphan sweeper: ledger entries without cleaned_at").
func TestLoadLedgerUncleanedEntries(t *testing.T) {
	dir := t.TempDir()
	ledger, err := certify.NewLedger(dir)
	if err != nil {
		t.Fatalf("NewLedger() error = %v", err)
	}
	leaked := "pm-cert-github-leak0001-1751450000"
	cleaned := "pm-cert-github-clean001-1751450001"
	if err := ledger.RecordPlanned(certify.LedgerEntry{Action: "create_label", Tag: leaked, Connector: "github"}); err != nil {
		t.Fatalf("RecordPlanned(leaked) error = %v", err)
	}
	if err := ledger.RecordPlanned(certify.LedgerEntry{Action: "create_label", Tag: cleaned, Connector: "github"}); err != nil {
		t.Fatalf("RecordPlanned(cleaned) error = %v", err)
	}
	if err := ledger.RecordCleaned(cleaned); err != nil {
		t.Fatalf("RecordCleaned(cleaned) error = %v", err)
	}

	entries, err := certify.LoadLedger(dir)
	if err != nil {
		t.Fatalf("LoadLedger() error = %v", err)
	}
	uncleaned := entries.Uncleaned()
	if len(uncleaned) != 1 {
		t.Fatalf("Uncleaned() returned %d entries, want 1: %+v", len(uncleaned), uncleaned)
	}
	if uncleaned[0].Tag != leaked {
		t.Errorf("Uncleaned()[0].Tag = %q, want %q", uncleaned[0].Tag, leaked)
	}
}

// TestLedgerCopyToReportDir proves the ledger is copied into
// .polymetrics/certifications/ledger/ even on crash (design §C "Ledger
// copied into .polymetrics/certifications/ledger/ even on crash").
func TestLedgerCopyToReportDir(t *testing.T) {
	workDir := t.TempDir()
	reportDir := t.TempDir()

	ledger, err := certify.NewLedger(workDir)
	if err != nil {
		t.Fatalf("NewLedger() error = %v", err)
	}
	if err := ledger.RecordPlanned(certify.LedgerEntry{Action: "create_label", Tag: "pm-cert-github-copy0001-1751450000", Connector: "github"}); err != nil {
		t.Fatalf("RecordPlanned() error = %v", err)
	}

	if err := ledger.CopyTo(reportDir, "github"); err != nil {
		t.Fatalf("CopyTo() error = %v", err)
	}

	copied := filepath.Join(reportDir, "certifications", "ledger", "github", "certify-ledger.jsonl")
	if _, err := os.Stat(copied); err != nil {
		t.Fatalf("expected copied ledger at %s: %v", copied, err)
	}
}

// TestLedgerRecordPlannedIsAppendOnly proves multiple RecordPlanned calls
// accumulate rather than overwrite (a real certify run stages many write
// pairs across one connector run).
func TestLedgerRecordPlannedIsAppendOnly(t *testing.T) {
	dir := t.TempDir()
	ledger, err := certify.NewLedger(dir)
	if err != nil {
		t.Fatalf("NewLedger() error = %v", err)
	}
	for i := 0; i < 3; i++ {
		runID := fmt.Sprintf("append%02d", i)
		tag := fmt.Sprintf("pm-cert-github-%s-%d", runID, 1751450000+i)
		if err := ledger.RecordPlanned(certify.LedgerEntry{Action: "create_label", Tag: tag, Connector: "github"}); err != nil {
			t.Fatalf("RecordPlanned() call %d error = %v", i, err)
		}
	}
	entries, err := certify.LoadLedger(dir)
	if err != nil {
		t.Fatalf("LoadLedger() error = %v", err)
	}
	if len(entries.All()) != 3 {
		t.Fatalf("All() returned %d entries, want 3", len(entries.All()))
	}
}
