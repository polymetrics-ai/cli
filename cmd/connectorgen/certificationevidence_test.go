package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors/certify"
)

func TestCertificationEvidencePostgresTransportPromotesOnlyCompletedModes(t *testing.T) {
	root := t.TempDir()
	reportPath := filepath.Join(root, "postgres-report.json")
	completed := postgresTransportEvidenceTestReport()
	raw, err := json.Marshal(struct {
		Kind   string         `json:"kind"`
		Report certify.Report `json:"report"`
	}{Kind: "ConnectorCertification", Report: completed})
	if err != nil {
		t.Fatalf("marshal report: %v", err)
	}
	if err := os.WriteFile(reportPath, raw, 0o600); err != nil {
		t.Fatalf("write report: %v", err)
	}
	t.Setenv("PM_POSTGRES_CERTIFICATION_EVIDENCE_TEST_PASSWORD", "live-test-secret")

	var stdout, stderr bytes.Buffer
	code := run([]string{
		"certification-evidence", "transport", "--connector", "postgres",
		"--report", reportPath,
		"--binary-sha", strings.Repeat("a", 64),
		"--from-env", "password=PM_POSTGRES_CERTIFICATION_EVIDENCE_TEST_PASSWORD",
		"--run-id", "postgres_transport_test",
		"--record-prefix", "postgres_transport_test",
		"--repo-root", root,
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("certification-evidence exit=%d stderr=%q", code, stderr.String())
	}
	if got := stdout.String(); got != "wrote declared transport evidence records: 12\n" {
		t.Fatalf("certification-evidence stdout=%q", got)
	}
	items, err := loadAcceptedEvidence(root, []string{"postgres"})
	if err != nil {
		t.Fatalf("loadAcceptedEvidence() error = %v", err)
	}
	if len(items) != 12 {
		t.Fatalf("accepted PostgreSQL transport evidence=%d, want 12", len(items))
	}
	for _, item := range items {
		if item.Scope != evidenceScopeSyncMode || item.Connector != "postgres" ||
			(item.Primitive != "database_read_into_warehouse" && item.Primitive != "database_write_from_warehouse") {
			t.Fatalf("accepted PostgreSQL evidence = %#v, want exact sync-mode proof", item)
		}
	}
}

func TestCertificationEvidencePostgresTransportRejectsUnexecutedMode(t *testing.T) {
	root := t.TempDir()
	report := postgresTransportEvidenceTestReport()
	report.Capabilities.DeclaredTransport.Modes[0].CheckpointCommitted = false
	raw, err := json.Marshal(struct {
		Kind   string         `json:"kind"`
		Report certify.Report `json:"report"`
	}{Kind: "ConnectorCertification", Report: report})
	if err != nil {
		t.Fatalf("marshal report: %v", err)
	}
	reportPath := filepath.Join(root, "postgres-report.json")
	if err := os.WriteFile(reportPath, raw, 0o600); err != nil {
		t.Fatalf("write report: %v", err)
	}
	t.Setenv("PM_POSTGRES_CERTIFICATION_EVIDENCE_TEST_PASSWORD", "live-test-secret")
	var stdout, stderr bytes.Buffer
	code := run([]string{
		"certification-evidence", "transport", "--connector", "postgres",
		"--report", reportPath,
		"--binary-sha", strings.Repeat("a", 64),
		"--from-env", "password=PM_POSTGRES_CERTIFICATION_EVIDENCE_TEST_PASSWORD",
		"--run-id", "postgres_transport_test",
		"--record-prefix", "postgres_transport_test",
		"--repo-root", root,
	}, &stdout, &stderr)
	if code != 1 || !strings.Contains(stderr.String(), "lacks completed declared target/read/checkpoint proof") {
		t.Fatalf("certification-evidence rejected report exit=%d stderr=%q", code, stderr.String())
	}
	items, err := loadAcceptedEvidence(root, []string{"postgres"})
	if err != nil {
		t.Fatalf("loadAcceptedEvidence() after rejected report: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("rejected PostgreSQL report wrote %d evidence records", len(items))
	}
}

func postgresTransportEvidenceTestReport() certify.Report {
	expected := []struct {
		mode     string
		strategy string
	}{
		{"full_overwrite", "replace"},
		{"full_append", "append"},
		{"incremental_append", "append"},
		{"incremental_upsert", "merge"},
		{"incremental_dedupe", "dedupe"},
		{"incremental_dedupe_history", "dedupe_history"},
	}
	modes := make([]certify.DeclaredTransportModeResult, 0, len(expected))
	for _, item := range expected {
		modes = append(modes, certify.DeclaredTransportModeResult{
			Mode: item.mode, ApplyStrategy: item.strategy, RecordsRead: 5, RecordsLoaded: 5,
			CheckpointCommitted: true, TargetNamespace: "polymetrics_cert", TargetRelation: "target_" + item.mode,
		})
	}
	return certify.Report{
		Kind: "ConnectorCertification", Connector: "postgres", Passed: true, CompletedAt: time.Date(2026, time.August, 17, 0, 0, 0, 0, time.UTC),
		Capabilities: certify.Capabilities{DeclaredTransport: &certify.DeclaredTransportResult{
			Result: "pass", SourceExecutor: "postgres_polling_watermark",
			DestinationExecutor: "postgres_managed_target", Modes: modes,
		}},
	}
}
