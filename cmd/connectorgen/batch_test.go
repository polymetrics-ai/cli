package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

func TestBatchPlanRejectsMissingRetrievalDate(t *testing.T) {
	ledger := writeBatchLedger(t, []map[string]any{
		batchLedgerRecordFixture("acme", 23, ""),
	})
	out := filepath.Join(t.TempDir(), "batch.json")
	var stdout, stderr bytes.Buffer

	code := run([]string{"batch", "plan", "--ledger", ledger, "--out", out, "--connector", "acme"}, &stdout, &stderr)
	if code == 0 {
		t.Fatalf("batch plan exit = 0, want missing retrieval-date rejection; stdout=%s stderr=%s", stdout.String(), stderr.String())
	}
	if !strings.Contains(stderr.String(), "retrieved_at") {
		t.Fatalf("batch plan stderr = %q, want retrieval-date evidence error", stderr.String())
	}
}

func TestBatchPlanWritesDeterministicEvidenceManifest(t *testing.T) {
	ledger := writeBatchLedger(t, []map[string]any{
		batchLedgerRecordFixture("bravo", 43, "2026-08-06"),
		batchLedgerRecordFixture("alpha", 23, "2026-08-05"),
	})
	root := t.TempDir()
	first := filepath.Join(root, "first.json")
	second := filepath.Join(root, "second.json")

	for _, out := range []string{first, second} {
		var stdout, stderr bytes.Buffer
		code := run([]string{"batch", "plan", "--ledger", ledger, "--out", out, "--size", "2"}, &stdout, &stderr)
		if code != 0 {
			t.Fatalf("batch plan exit = %d, want 0; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
		}
	}

	firstRaw, err := os.ReadFile(first)
	if err != nil {
		t.Fatalf("read first manifest: %v", err)
	}
	secondRaw, err := os.ReadFile(second)
	if err != nil {
		t.Fatalf("read second manifest: %v", err)
	}
	if string(firstRaw) != string(secondRaw) {
		t.Fatalf("manifest is not deterministic:\nfirst=%s\nsecond=%s", firstRaw, secondRaw)
	}

	var manifest struct {
		Connectors []struct {
			Connector       string `json:"connector"`
			OperationsTotal int    `json:"operations_total"`
			Artifact        struct {
				URL         string `json:"url"`
				Version     string `json:"version"`
				RetrievedAt string `json:"retrieved_at"`
			} `json:"artifact"`
		} `json:"connectors"`
	}
	if err := json.Unmarshal(firstRaw, &manifest); err != nil {
		t.Fatalf("decode manifest: %v", err)
	}
	if len(manifest.Connectors) != 2 {
		t.Fatalf("manifest connectors = %d, want 2", len(manifest.Connectors))
	}
	if got := manifest.Connectors[0]; got.Connector != "alpha" || got.OperationsTotal != 23 || got.Artifact.URL == "" || got.Artifact.Version == "" || got.Artifact.RetrievedAt != "2026-08-05" {
		t.Fatalf("first manifest connector = %+v, want preserved alpha evidence", got)
	}
}

func TestBatchGateContinuesAfterConnectorFailure(t *testing.T) {
	defsRoot := t.TempDir()
	writeBatchBundle(t, defsRoot, cliSurfaceBundleFS(validCLISurfaceJSON()))
	if err := os.MkdirAll(filepath.Join(defsRoot, "broken"), 0o755); err != nil {
		t.Fatalf("mkdir broken bundle: %v", err)
	}
	manifestPath := writeBatchManifestFixture(t, "cli-surface", "broken")
	reportPath := filepath.Join(t.TempDir(), "report.json")
	var stdout, stderr bytes.Buffer

	code := run([]string{"batch", "gate", "--manifest", manifestPath, "--defs-root", defsRoot, "--report", reportPath}, &stdout, &stderr)
	if code == 0 {
		t.Fatalf("batch gate exit = 0, want a nonzero exit for the broken connector; stdout=%s stderr=%s", stdout.String(), stderr.String())
	}

	report := readBatchGateReportFixture(t, reportPath)
	if len(report.Included) != 1 || report.Included[0].Connector != "cli-surface" {
		t.Fatalf("included = %+v, want cli-surface to proceed despite broken sibling", report.Included)
	}
	if len(report.Dropped) != 1 || report.Dropped[0].Connector != "broken" || report.Dropped[0].Stage != "validate" {
		t.Fatalf("dropped = %+v, want named validate-stage broken drop", report.Dropped)
	}
}

func TestBatchGateUsesRuntimePreflightForEveryImplementedCommand(t *testing.T) {
	defsRoot := t.TempDir()
	writeBatchBundle(t, defsRoot, cliSurfaceBundleFS(validCLISurfaceJSON()))
	manifestPath := writeBatchManifestFixture(t, "cli-surface")
	reportPath := filepath.Join(t.TempDir(), "report.json")
	var stdout, stderr bytes.Buffer

	code := run([]string{"batch", "gate", "--manifest", manifestPath, "--defs-root", defsRoot, "--report", reportPath}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("batch gate exit = %d, want 0; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	report := readBatchGateReportFixture(t, reportPath)
	if len(report.Included) != 1 || report.Included[0].CommandsChecked != 2 {
		t.Fatalf("included = %+v, want both implemented commands to run through runtime preflight", report.Included)
	}
}

func TestBatchGateDropsSurfaceSyncDrift(t *testing.T) {
	defsRoot := t.TempDir()
	writeBatchBundle(t, defsRoot, batchSurfaceSyncDriftBundleFS())
	manifestPath := writeBatchManifestFixture(t, "cli-surface")
	reportPath := filepath.Join(t.TempDir(), "report.json")
	var stdout, stderr bytes.Buffer

	code := run([]string{"batch", "gate", "--manifest", manifestPath, "--defs-root", defsRoot, "--report", reportPath}, &stdout, &stderr)
	if code == 0 {
		t.Fatalf("batch gate exit = 0, want surface-sync drift failure; stdout=%s stderr=%s", stdout.String(), stderr.String())
	}
	report := readBatchGateReportFixture(t, reportPath)
	if len(report.Dropped) != 1 || report.Dropped[0].Stage != "surface_sync" || !strings.Contains(report.Dropped[0].Reason, "flag_maps_to") {
		t.Fatalf("dropped = %+v, want surface-sync field-drift drop", report.Dropped)
	}
}

func TestBatchGateDropsRedactingOutputPolicy(t *testing.T) {
	defsRoot := t.TempDir()
	writeBatchBundle(t, defsRoot, directWriteCLISurfaceBundleFS())
	manifestPath := writeBatchManifestFixture(t, "cli-surface")
	reportPath := filepath.Join(t.TempDir(), "report.json")
	var stdout, stderr bytes.Buffer

	code := run([]string{"batch", "gate", "--manifest", manifestPath, "--defs-root", defsRoot, "--report", reportPath}, &stdout, &stderr)
	if code == 0 {
		t.Fatalf("batch gate exit = 0, want no-redaction failure; stdout=%s stderr=%s", stdout.String(), stderr.String())
	}
	report := readBatchGateReportFixture(t, reportPath)
	if len(report.Dropped) != 1 || report.Dropped[0].Stage != "output_policy" || !strings.Contains(report.Dropped[0].Reason, "json_redacted") {
		t.Fatalf("dropped = %+v, want json_redacted output-policy drop", report.Dropped)
	}
}

func TestBatchGateDropsWriteRedactFields(t *testing.T) {
	defsRoot := t.TempDir()
	fsys := cliSurfaceBundleFS(validCLISurfaceJSON())
	fsys["cli-surface/writes.json"] = &fstest.MapFile{Data: []byte(strings.Replace(
		string(fsys["cli-surface/writes.json"].Data),
		`"risk": "creates a widget"`,
		`"redact_fields": ["token"], "risk": "creates a widget"`,
		1,
	))}
	writeBatchBundle(t, defsRoot, fsys)
	manifestPath := writeBatchManifestFixture(t, "cli-surface")
	reportPath := filepath.Join(t.TempDir(), "report.json")
	var stdout, stderr bytes.Buffer

	code := run([]string{"batch", "gate", "--manifest", manifestPath, "--defs-root", defsRoot, "--report", reportPath}, &stdout, &stderr)
	if code == 0 {
		t.Fatalf("batch gate exit = 0, want redact_fields failure; stdout=%s stderr=%s", stdout.String(), stderr.String())
	}
	report := readBatchGateReportFixture(t, reportPath)
	if len(report.Dropped) != 1 || report.Dropped[0].Stage != "output_policy" || !strings.Contains(report.Dropped[0].Reason, `write action "create_widget" declares redact_fields`) {
		t.Fatalf("dropped = %+v, want write redact_fields output-policy drop", report.Dropped)
	}
}

func writeBatchLedger(t *testing.T, records []map[string]any) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "ledger.json")
	raw, err := json.MarshalIndent(map[string]any{
		"schema_version": "cli-provider-artifact-sweep-r1",
		"created_at":     "2026-08-05T00:00:00Z",
		"records":        records,
	}, "", "  ")
	if err != nil {
		t.Fatalf("encode ledger: %v", err)
	}
	if err := os.WriteFile(path, append(raw, '\n'), 0o644); err != nil {
		t.Fatalf("write ledger: %v", err)
	}
	return path
}

func batchLedgerRecordFixture(name string, total int, retrievedAt string) map[string]any {
	return map[string]any{
		"connector":             name,
		"status":                "done",
		"operations_total":      total,
		"operations_read":       total - 3,
		"operations_write":      3,
		"artifact_url":          "https://example.test/" + name + ".openapi.json",
		"artifact_kind":         "openapi",
		"artifact_version":      "1.0.0",
		"retrieved_at":          retrievedAt,
		"auth_model":            "api_key",
		"access_model":          "public",
		"evidence_source":       "provider-published test OpenAPI",
		"counting_note":         "exact operation inventory",
		"processed_at":          "2026-08-05T00:00:00Z",
		"scope_in_current_defs": true,
	}
}

func writeBatchManifestFixture(t *testing.T, names ...string) string {
	t.Helper()
	connectors := make([]BatchManifestConnector, 0, len(names))
	for i, name := range names {
		connectors = append(connectors, BatchManifestConnector{
			Connector:       name,
			OperationsTotal: 10 + i,
			OperationsRead:  7 + i,
			OperationsWrite: 3,
			Artifact: BatchArtifact{
				URL:         "https://example.test/" + name + ".json",
				Kind:        "openapi",
				Version:     "1.0.0",
				RetrievedAt: "2026-08-05",
			},
			AuthModel:       "api_key",
			AccessModel:     "public",
			EvidenceSource:  "provider-published test OpenAPI",
			CountingNote:    "exact operation inventory",
			ProcessedAt:     "2026-08-05T00:00:00Z",
			SelectionReason: "test evidence",
		})
	}
	path := filepath.Join(t.TempDir(), "batch.json")
	if err := writeBatchManifest(path, BatchManifest{
		SchemaVersion: batchManifestSchemaVersion,
		SourceLedger:  BatchLedgerSource{SchemaVersion: "test", CreatedAt: "2026-08-05T00:00:00Z", RecordCount: len(connectors)},
		Selection:     BatchSelection{Mode: "explicit", RequestedSize: len(connectors), Criteria: "test"},
		Connectors:    connectors,
	}); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	return path
}

type batchGateReportFixture struct {
	Included []struct {
		Connector       string `json:"connector"`
		CommandsChecked int    `json:"commands_checked"`
	} `json:"included"`
	Dropped []struct {
		Connector string `json:"connector"`
		Stage     string `json:"stage"`
		Reason    string `json:"reason"`
	} `json:"dropped"`
}

func readBatchGateReportFixture(t *testing.T, path string) batchGateReportFixture {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read batch gate report: %v", err)
	}
	var report batchGateReportFixture
	if err := json.Unmarshal(raw, &report); err != nil {
		t.Fatalf("decode batch gate report: %v", err)
	}
	return report
}

func writeBatchBundle(t *testing.T, defsRoot string, fsys fstest.MapFS) string {
	t.Helper()
	const fixturePrefix = "cli-surface/"
	for source, file := range fsys {
		rel, ok := strings.CutPrefix(source, fixturePrefix)
		if !ok {
			t.Fatalf("fixture file %q does not start with %q", source, fixturePrefix)
		}
		path := filepath.Join(defsRoot, "cli-surface", filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
		}
		if err := os.WriteFile(path, file.Data, 0o644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}
	return filepath.Join(defsRoot, "cli-surface")
}

func batchSurfaceSyncDriftBundleFS() fstest.MapFS {
	fsys := cliSurfaceBundleFS(`{
		"tagline": "Work with CLI Surface from the command line.",
		"usage": "pm cli-surface <command> [flags]",
		"commands": [
			{
				"path": "artifact download",
				"summary": "Download an artifact",
				"intent": "binary_download",
				"availability": "implemented",
				"operation": "cli-surface.artifact.download",
				"source_cli_path": "clis artifact download",
				"api_surface": [{ "method": "GET", "path": "/artifacts/{id}" }],
				"flags": [{ "name": "id", "type": "string", "maps_to": "query.id" }],
				"examples": ["pm cli-surface artifact download --id w_1 --dest-root ./downloads"]
			}
		]
	}`)
	fsys["cli-surface/api_surface.json"] = &fstest.MapFile{Data: []byte(`{
		"api": "test API v1",
		"operation_ledger_version": 1,
		"endpoints": [
			{ "method": "GET", "path": "/widgets", "covered_by": { "stream": "widgets" } },
			{ "method": "POST", "path": "/widgets", "covered_by": { "write": "create_widget" } },
			{ "method": "GET", "path": "/artifacts/{id}", "operation": { "model": "binary_read", "status": "blocked", "risk": "low", "blocked_by_default": true, "reason": "typed binary operation metadata" } }
		]
	}`)}
	fsys["cli-surface/operations.json"] = &fstest.MapFile{Data: []byte(`{
		"operations": [{
			"id": "cli-surface.artifact.download",
			"kind": "binary_download",
			"summary": "Download an artifact",
			"risk": "low",
			"approval": "none",
			"output_policy": "binary_file_bounded",
			"binary": { "method": "GET", "path": "/artifacts/{id}", "max_bytes": 1024 }
		}]
	}`)}
	return fsys
}
