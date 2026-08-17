package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors/certify"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
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

func TestCertificationEvidenceReportImportsDefinitionBoundHTTPProofWithoutSecrets(t *testing.T) {
	const canary = "cert-evidence-import-header-query-body-canary"
	root := t.TempDir()
	reportPath := writeCompletedReadEvidenceReport(t, root, "github")
	proofPath := writeImportedEvidenceProof(t, root, "github", canary)

	var stdout, stderr bytes.Buffer
	code := run([]string{
		"certification-evidence", "report", "--connector", "github", "--report", reportPath,
		"--external-proof", proofPath, "--record-prefix", "github_read_import", "--repo-root", root,
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("report importer exit=%d stderr=%q", code, stderr.String())
	}
	if got := stdout.String(); got != "wrote report evidence records: 2\n" {
		t.Fatalf("report importer stdout=%q", got)
	}
	items, err := loadAcceptedEvidence(root, []string{"github"})
	if err != nil {
		t.Fatalf("loadAcceptedEvidence() = %v", err)
	}
	if got := matchingCapabilityEvidence(items, "github", "operation:rest_read"); len(got) != 1 {
		t.Fatalf("matchingCapabilityEvidence() = %#v, want one imported GitHub read record", got)
	}
	if got := matchingCapabilityEvidence(items, "github", "operation:graphql_query"); len(got) != 1 {
		t.Fatalf("matchingCapabilityEvidence() = %#v, want one imported GitHub GraphQL read record", got)
	}
	if len(items) != 2 || items[0].Provider != "github_api" || len(items[0].Proof.HTTPExchanges) != 1 {
		t.Fatalf("imported GitHub evidence = %#v", items)
	}
	raw, err := os.ReadFile(filepath.Join(root, acceptedEvidenceDirectory, "github_read_import-capability_operation_rest_read.json"))
	if err != nil {
		t.Fatalf("read emitted evidence: %v", err)
	}
	if bytes.Contains(raw, []byte(canary)) {
		t.Fatal("emitted evidence retained the planted header/query/body secret")
	}
	response := items[0].Proof.HTTPExchanges[0].Response.Body
	if response.Encoding != "json" || !isSanitizedJSONValueMust(t, response.Value) {
		t.Fatalf("response body = %#v, want fully fingerprinted JSON", response)
	}
}

func TestCertificationEvidenceReportUsesSecondConnectorDefinitionWithoutSharedBranch(t *testing.T) {
	root := t.TempDir()
	reportPath := writeCompletedReadEvidenceReport(t, root, "xero")
	proofPath := writeImportedEvidenceProof(t, root, "xero", "cert-evidence-xero-canary")
	var stdout, stderr bytes.Buffer
	code := run([]string{
		"certification-evidence", "report", "--connector", "xero", "--report", reportPath,
		"--external-proof", proofPath, "--record-prefix", "xero_read_import", "--repo-root", root,
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("second connector report importer exit=%d stderr=%q", code, stderr.String())
	}
	raw, err := os.ReadFile(filepath.Join(root, acceptedEvidenceDirectory, "xero_read_import-capability_operation_rest_read.json"))
	if err != nil {
		t.Fatalf("read second connector evidence: %v", err)
	}
	var evidence acceptedEvidence
	if err := decodeStrictJSON(raw, &evidence); err != nil || validateAcceptedEvidence(evidence) != nil || evidence.Provider != "xero_api" || evidence.FunctionKind != "operation:rest_read" {
		t.Fatalf("second connector evidence = %#v, decode/validate error=%v", evidence, err)
	}
}

func TestCertificationEvidenceReportRequiresFreshCompletedStages(t *testing.T) {
	root := t.TempDir()
	reportPath := writeCompletedReadEvidenceReport(t, root, "github")
	var report certify.Report
	raw, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &report); err != nil {
		t.Fatal(err)
	}
	report.Stages[0].Resumed = true
	raw, err = json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(reportPath, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	proofPath := writeImportedEvidenceProof(t, root, "github", "cert-evidence-resumed-canary")
	var stdout, stderr bytes.Buffer
	code := run([]string{
		"certification-evidence", "report", "--connector", "github", "--report", reportPath,
		"--external-proof", proofPath, "--record-prefix", "github_resumed", "--repo-root", root,
	}, &stdout, &stderr)
	if code != 1 || !strings.Contains(stderr.String(), "not freshly completed") {
		t.Fatalf("resumed report importer exit=%d stderr=%q", code, stderr.String())
	}
	items, err := loadAcceptedEvidence(root, []string{"github"})
	if err != nil || len(items) != 0 {
		t.Fatalf("resumed report emitted evidence=%#v err=%v", items, err)
	}
}

func TestCertificationEvidenceReportRefusesIncompleteBindingsWithoutPartialPublication(t *testing.T) {
	root := t.TempDir()
	reportPath := writeCompletedReadEvidenceReport(t, root, "github")
	var report certify.Report
	raw, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &report); err != nil {
		t.Fatal(err)
	}
	for index := range report.Stages {
		if strings.HasPrefix(report.Stages[index].Name, "graphql_") {
			report.Stages[index].Resumed = true
			break
		}
	}
	raw, err = json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(reportPath, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	proofPath := writeImportedEvidenceProof(t, root, "github", "cert-evidence-incomplete-binding-canary")
	var stdout, stderr bytes.Buffer
	code := run([]string{
		"certification-evidence", "report", "--connector", "github", "--report", reportPath,
		"--external-proof", proofPath, "--record-prefix", "github_incomplete_binding", "--repo-root", root,
	}, &stdout, &stderr)
	if code != 1 || !strings.Contains(stderr.String(), "not freshly completed") {
		t.Fatalf("incomplete binding importer exit=%d stderr=%q", code, stderr.String())
	}
	items, err := loadAcceptedEvidence(root, []string{"github"})
	if err != nil || len(items) != 0 {
		t.Fatalf("incomplete binding emitted partial evidence=%#v err=%v", items, err)
	}
}

func TestCertificationEvidenceReportRefusesUnverifiedFullParityScope(t *testing.T) {
	root := t.TempDir()
	reportPath := writeCompletedReadEvidenceReport(t, root, "github")
	var report certify.Report
	raw, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &report); err != nil {
		t.Fatal(err)
	}
	for index := range report.Stages {
		if report.Stages[index].Name == "full_parity" {
			report.Stages[index].Passed = false
			break
		}
	}
	raw, err = json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(reportPath, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	proofPath := writeImportedEvidenceProof(t, root, "github", "cert-evidence-unverified-scope-canary")
	var stdout, stderr bytes.Buffer
	code := run([]string{
		"certification-evidence", "report", "--connector", "github", "--report", reportPath,
		"--external-proof", proofPath, "--record-prefix", "github_unverified_scope", "--repo-root", root,
	}, &stdout, &stderr)
	if code != 1 || !strings.Contains(stderr.String(), "does not prove the accepted full-parity credential scope") {
		t.Fatalf("unverified scope importer exit=%d stderr=%q", code, stderr.String())
	}
	items, err := loadAcceptedEvidence(root, []string{"github"})
	if err != nil || len(items) != 0 {
		t.Fatalf("unverified scope emitted evidence=%#v err=%v", items, err)
	}
}

func TestCertificationEvidenceReportRefusesProofFromDifferentRun(t *testing.T) {
	root := t.TempDir()
	reportPath := writeCompletedReadEvidenceReport(t, root, "github")
	proofPath := writeImportedEvidenceProof(t, root, "github", "cert-evidence-wrong-run-canary")
	var report certify.Report
	raw, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &report); err != nil {
		t.Fatal(err)
	}
	report.StartedAt = report.StartedAt.Add(time.Second)
	raw, err = json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(reportPath, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := run([]string{
		"certification-evidence", "report", "--connector", "github", "--report", reportPath,
		"--external-proof", proofPath, "--record-prefix", "github_wrong_run", "--repo-root", root,
	}, &stdout, &stderr)
	if code != 1 || !strings.Contains(stderr.String(), "does not belong to the completed report run") {
		t.Fatalf("wrong-run importer exit=%d stderr=%q", code, stderr.String())
	}
	items, err := loadAcceptedEvidence(root, []string{"github"})
	if err != nil || len(items) != 0 {
		t.Fatalf("wrong-run proof emitted evidence=%#v err=%v", items, err)
	}
}

var evidenceImportTestStartedAt = time.Date(2026, time.August, 17, 0, 0, 0, 0, time.UTC)

func writeCompletedReadEvidenceReport(t *testing.T, root, connector string) string {
	t.Helper()
	bundle, err := engine.Load(defs.FS, connector)
	if err != nil {
		t.Fatalf("load %s bundle: %v", connector, err)
	}
	stages := make([]certify.StageResult, 0, len(bundle.Certification.DirectReadCandidates))
	for _, candidate := range bundle.Certification.DirectReadCandidates {
		stages = append(stages, certify.StageResult{Name: candidate.StageName, Passed: true, Status: "passed"})
	}
	if bundle.Certification.GraphQL != nil {
		for _, candidate := range bundle.Certification.GraphQL.LiveCandidates {
			stages = append(stages, certify.StageResult{Name: candidate.StageName, Passed: true, Status: "passed"})
		}
	}
	stages = append(stages, certify.StageResult{Name: "full_parity", Passed: true, Status: "passed"})
	report := certify.Report{
		Kind: "ConnectorCertification", Connector: connector, Passed: true,
		StartedAt: evidenceImportTestStartedAt, CompletedAt: evidenceImportTestStartedAt.Add(time.Second), Stages: stages,
	}
	raw, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, connector+"-report.json")
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func writeImportedEvidenceProof(t *testing.T, root, connector, canary string) string {
	t.Helper()
	body := []byte(`{"account":"` + canary + `"}`)
	path, err := certify.WriteExternalProof(root, certify.ExternalProofInput{
		Connector: connector, RunID: fmt.Sprintf("external-%d", evidenceImportTestStartedAt.UTC().UnixNano()), BinarySHA256: strings.Repeat("a", 64),
		Command: []string{"pm", "connectors", "certify", connector, "--from-env", "token=PM_CERT_TOKEN"},
		Stdout:  "completed", ExitCode: 0, Passed: true, FullParity: true, PreparedValues: []string{canary},
		HTTPExchanges: []certify.ObservedHTTPExchange{{
			Request: certify.ObservedHTTPRequest{Method: http.MethodGet, Target: "https://api.example.test/resource?token=" + canary,
				Headers: http.Header{"Authorization": {"Bearer " + canary}}, Body: certify.ObservedBody{Complete: true}},
			Response: certify.ObservedHTTPResponse{Status: http.StatusOK, Body: certify.ObservedBody{Bytes: body, OriginalBytes: len(body), Complete: true}},
		}},
	})
	if err != nil {
		t.Fatalf("write external proof: %v", err)
	}
	return path
}

func isSanitizedJSONValueMust(t *testing.T, raw json.RawMessage) bool {
	t.Helper()
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		t.Fatalf("decode response proof body: %v", err)
	}
	return isSanitizedJSONValue(value)
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
