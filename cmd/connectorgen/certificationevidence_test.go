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

func TestCertificationEvidencePostgresTransportCapabilityWritePromotesOnlyCompleteAggregateProfile(t *testing.T) {
	root := t.TempDir()
	reportPath := filepath.Join(root, "postgres-report.json")
	raw, err := json.Marshal(struct {
		Kind   string         `json:"kind"`
		Report certify.Report `json:"report"`
	}{Kind: "ConnectorCertification", Report: postgresTransportEvidenceTestReport()})
	if err != nil {
		t.Fatalf("marshal report: %v", err)
	}
	if err := os.WriteFile(reportPath, raw, 0o600); err != nil {
		t.Fatalf("write report: %v", err)
	}
	t.Setenv("PM_POSTGRES_CERTIFICATION_EVIDENCE_TEST_PASSWORD", "live-test-secret")

	var stdout, stderr bytes.Buffer
	code := run([]string{
		"certification-evidence", "transport-capability-write", "--connector", "postgres",
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
	if got := stdout.String(); got != "wrote declared transport capability evidence records: 1\n" {
		t.Fatalf("certification-evidence stdout=%q", got)
	}
	items, err := loadAcceptedEvidence(root, []string{"postgres"})
	if err != nil {
		t.Fatalf("loadAcceptedEvidence() error = %v", err)
	}
	if len(items) != 1 || items[0].Scope != evidenceScopeCapability || items[0].FunctionKind != "capability:write" {
		t.Fatalf("accepted PostgreSQL capability evidence=%#v, want one capability:write record", items)
	}
}

func TestCertificationEvidencePostgresTransportCapabilityWriteRejectsIncompleteAggregateProfile(t *testing.T) {
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
		"certification-evidence", "transport-capability-write", "--connector", "postgres",
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
		t.Fatalf("rejected PostgreSQL capability report wrote %d evidence records", len(items))
	}
}

func TestCertificationEvidencePostgresChangeCapturePromotesOnlyReceiptBackedBinaryProof(t *testing.T) {
	root := t.TempDir()
	reportPath := filepath.Join(root, "postgres-cdc-report.json")
	raw, err := json.Marshal(postgresChangeCaptureEvidenceTestReport())
	if err != nil {
		t.Fatalf("marshal change-capture report: %v", err)
	}
	if err := os.WriteFile(reportPath, raw, 0o600); err != nil {
		t.Fatalf("write change-capture report: %v", err)
	}
	t.Setenv("PM_POSTGRES_CERTIFICATION_EVIDENCE_TEST_PASSWORD", "live-test-secret")

	var stdout, stderr bytes.Buffer
	code := run([]string{
		"certification-evidence", "change-capture", "--connector", "postgres",
		"--report", reportPath,
		"--binary-sha", strings.Repeat("a", 64),
		"--from-env", "password=PM_POSTGRES_CERTIFICATION_EVIDENCE_TEST_PASSWORD",
		"--run-id", "postgres_cdc_test",
		"--record-prefix", "postgres_cdc_test",
		"--repo-root", root,
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("certification-evidence exit=%d stderr=%q", code, stderr.String())
	}
	if got := stdout.String(); got != "wrote change-capture evidence records: 2\n" {
		t.Fatalf("certification-evidence stdout=%q", got)
	}
	items, err := loadAcceptedEvidence(root, []string{"postgres"})
	if err != nil {
		t.Fatalf("loadAcceptedEvidence() error = %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("accepted PostgreSQL change-capture evidence=%d, want 2", len(items))
	}
	seenCapability, seenMode := false, false
	for _, item := range items {
		switch {
		case item.Scope == evidenceScopeCapability && item.FunctionKind == "capability:cdc":
			seenCapability = true
		case item.Scope == evidenceScopeSyncMode && item.SyncMode == "change_capture" && item.Primitive == "database_read_into_warehouse":
			seenMode = true
		default:
			t.Fatalf("accepted PostgreSQL evidence = %#v, want only CDC capability and database-read mode proof", item)
		}
	}
	if !seenCapability || !seenMode {
		t.Fatalf("change-capture evidence coverage capability=%t mode=%t, want both", seenCapability, seenMode)
	}
}

func TestCertificationEvidencePostgresChangeCaptureRejectsAcknowledgementBeforeReceipt(t *testing.T) {
	root := t.TempDir()
	report := postgresChangeCaptureEvidenceTestReport()
	report.AcknowledgedAfterWarehouseReceipt = false
	raw, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("marshal change-capture report: %v", err)
	}
	reportPath := filepath.Join(root, "postgres-cdc-report.json")
	if err := os.WriteFile(reportPath, raw, 0o600); err != nil {
		t.Fatalf("write change-capture report: %v", err)
	}
	t.Setenv("PM_POSTGRES_CERTIFICATION_EVIDENCE_TEST_PASSWORD", "live-test-secret")

	var stdout, stderr bytes.Buffer
	code := run([]string{
		"certification-evidence", "change-capture", "--connector", "postgres",
		"--report", reportPath,
		"--binary-sha", strings.Repeat("a", 64),
		"--from-env", "password=PM_POSTGRES_CERTIFICATION_EVIDENCE_TEST_PASSWORD",
		"--run-id", "postgres_cdc_test",
		"--record-prefix", "postgres_cdc_test",
		"--repo-root", root,
	}, &stdout, &stderr)
	if code != 1 || !strings.Contains(stderr.String(), "requires bounded staging, independent warehouse read-back, and acknowledgement after receipt") {
		t.Fatalf("certification-evidence rejected report exit=%d stderr=%q", code, stderr.String())
	}
	items, err := loadAcceptedEvidence(root, []string{"postgres"})
	if err != nil {
		t.Fatalf("loadAcceptedEvidence() after rejected report: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("rejected PostgreSQL change-capture report wrote %d evidence records", len(items))
	}
}

func postgresChangeCaptureEvidenceTestReport() changeCaptureEvidenceReport {
	return changeCaptureEvidenceReport{
		Kind:                              "ChangeCaptureCertification",
		Connector:                         "postgres",
		CompletedAt:                       time.Date(2026, time.August, 17, 0, 0, 0, 0, time.UTC),
		BoundedDurableStaging:             true,
		WarehouseReceiptPersisted:         true,
		IndependentWarehouseReadback:      true,
		AcknowledgedAfterWarehouseReceipt: true,
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
