package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"polymetrics.ai/internal/connectors/engine"
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

func TestBatchNamespaceRendersContextualHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer

	code := run([]string{"batch"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("bare batch exit = %d, want 0; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "connectorgen batch materialize") || stderr.Len() != 0 {
		t.Fatalf("bare batch output = stdout %q stderr %q, want contextual stdout help only", stdout.String(), stderr.String())
	}
}

func TestBatchRejectsInvalidSubcommand(t *testing.T) {
	var stdout, stderr bytes.Buffer

	code := run([]string{"batch", "unknown"}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("invalid batch subcommand exit = %d, want 2; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	if stdout.Len() != 0 || !strings.Contains(stderr.String(), "unknown subcommand") {
		t.Fatalf("invalid batch output = stdout %q stderr %q, want usage error on stderr", stdout.String(), stderr.String())
	}
}

func TestBatchGateContinuesAfterConnectorFailure(t *testing.T) {
	defsRoot := t.TempDir()
	writeBatchBundle(t, defsRoot, batchGateBundleFS(t, validCLISurfaceJSON()))
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
	writeBatchBundle(t, defsRoot, batchGateBundleFS(t, validCLISurfaceJSON()))
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

func TestBatchGateDropsSurfaceWithoutImplementedCommands(t *testing.T) {
	defsRoot := t.TempDir()
	writeBatchBundle(t, defsRoot, batchGateBundleFS(t, strings.ReplaceAll(validCLISurfaceJSON(), `"availability": "implemented"`, `"availability": "planned"`)))
	manifestPath := writeBatchManifestFixture(t, "cli-surface")
	reportPath := filepath.Join(t.TempDir(), "report.json")
	var stdout, stderr bytes.Buffer

	code := run([]string{"batch", "gate", "--manifest", manifestPath, "--defs-root", defsRoot, "--report", reportPath}, &stdout, &stderr)
	if code == 0 {
		t.Fatalf("batch gate exit = 0, want no-implemented-command failure; stdout=%s stderr=%s", stdout.String(), stderr.String())
	}
	report := readBatchGateReportFixture(t, reportPath)
	if len(report.Included) != 0 || len(report.Dropped) != 1 || report.Dropped[0].Stage != "runtime_preflight" || !strings.Contains(report.Dropped[0].Reason, "no implemented command") {
		t.Fatalf("report = %+v, want runtime-preflight drop for zero implemented commands", report)
	}
}

func TestBatchGateSelectsOneManifestCandidateForIndividualPreflight(t *testing.T) {
	defsRoot := t.TempDir()
	writeBatchBundle(t, defsRoot, batchGateBundleFS(t, validCLISurfaceJSON()))
	if err := os.MkdirAll(filepath.Join(defsRoot, "broken"), 0o755); err != nil {
		t.Fatalf("mkdir broken bundle: %v", err)
	}
	manifestPath := writeBatchManifestFixture(t, "cli-surface", "broken")
	reportPath := filepath.Join(t.TempDir(), "report.json")
	var stdout, stderr bytes.Buffer

	code := run([]string{"batch", "gate", "--manifest", manifestPath, "--defs-root", defsRoot, "--connector", "cli-surface", "--report", reportPath}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("individual batch gate exit = %d, want 0; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	report := readBatchGateReportFixture(t, reportPath)
	if len(report.Included) != 1 || report.Included[0].Connector != "cli-surface" || len(report.Dropped) != 0 {
		t.Fatalf("individual gate report = %+v, want only cli-surface included", report)
	}
}

func TestBatchGateDropsSurfaceSyncDrift(t *testing.T) {
	defsRoot := t.TempDir()
	writeBatchBundle(t, defsRoot, batchSurfaceSyncDriftBundleFS(t))
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
	fsys := directWriteCLISurfaceBundleFS()
	fsys["cli-surface/api_surface.json"] = &fstest.MapFile{Data: []byte(batchGateDirectWriteSurface(t))}
	writeBatchBundle(t, defsRoot, fsys)
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
	fsys := batchGateBundleFS(t, validCLISurfaceJSON())
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

func TestBatchGateRejectsProtocolMetadataCoverage(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		wantReason string
	}{
		{name: http.MethodOptions, method: http.MethodOptions, wantReason: "covered_by"},
		{name: http.MethodTrace, method: http.MethodTrace, wantReason: "covered_by"},
		{name: "noncanonical TRACE", method: "trace", wantReason: "exact canonical TRACE"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defsRoot := t.TempDir()
			endpoints := append(batchGateDefaultEndpoints(batchGateFixtureArtifactURL), engine.SurfaceEndpoint{
				Method:     test.method,
				Path:       "/protocol-metadata",
				Provenance: batchGateFixtureProvenance(batchGateFixtureArtifactURL),
				CoveredBy:  &engine.SurfaceCoverage{Stream: "widgets"},
			})
			writeBatchBundle(t, defsRoot, batchGateBundleFSWithSurface(t, validCLISurfaceJSON(), batchGateV2Surface(t, batchGateFixtureArtifactURL, endpoints)))
			manifestPath := writeBatchManifestFixture(t, "cli-surface")
			reportPath := filepath.Join(t.TempDir(), "report.json")
			var stdout, stderr bytes.Buffer

			code := run([]string{"batch", "gate", "--manifest", manifestPath, "--defs-root", defsRoot, "--report", reportPath}, &stdout, &stderr)
			if code == 0 {
				t.Fatalf("batch gate exit = 0, want protocol coverage rejection; stdout=%s stderr=%s", stdout.String(), stderr.String())
			}
			report := readBatchGateReportFixture(t, reportPath)
			if report.ProvenanceRefusals != 0 || len(report.Included) != 0 || len(report.Dropped) != 1 || report.Dropped[0].Stage != "api_surface" || !strings.Contains(report.Dropped[0].Reason, test.method) || !strings.Contains(report.Dropped[0].Reason, test.wantReason) {
				t.Fatalf("report = %+v, want method-specific protocol coverage drop", report)
			}
		})
	}
}

func TestBatchGateClassifiesProtocolMetadataExclusions(t *testing.T) {
	defsRoot := t.TempDir()
	endpoints := append(batchGateDefaultEndpoints(batchGateFixtureArtifactURL),
		engine.SurfaceEndpoint{
			Method:     http.MethodOptions,
			Path:       "/protocol-metadata",
			Provenance: batchGateFixtureProvenance(batchGateFixtureArtifactURL),
			Operation:  batchProtocolMetadataOperation(http.MethodOptions),
		},
		engine.SurfaceEndpoint{
			Method:     http.MethodTrace,
			Path:       "/protocol-metadata",
			Provenance: batchGateFixtureProvenance(batchGateFixtureArtifactURL),
			Operation:  batchProtocolMetadataOperation(http.MethodTrace),
		},
	)
	writeBatchBundle(t, defsRoot, batchGateBundleFSWithSurface(t, validCLISurfaceJSON(), batchGateV2Surface(t, batchGateFixtureArtifactURL, endpoints)))
	manifestPath := writeBatchManifestFixture(t, "cli-surface")
	reportPath := filepath.Join(t.TempDir(), "report.json")
	var stdout, stderr bytes.Buffer

	code := run([]string{"batch", "gate", "--manifest", manifestPath, "--defs-root", defsRoot, "--report", reportPath}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("batch gate exit = %d, want protocol exclusions included; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	report := readBatchGateReportFixture(t, reportPath)
	if len(report.Included) != 1 || len(report.Dropped) != 0 || report.Included[0].DeclaredOperations != 4 || report.Included[0].OperationSplit != (BatchOperationSplit{Executable: 2, Excluded: 2}) {
		t.Fatalf("report = %+v, want two executable and two protocol-metadata exclusions", report)
	}
}

func TestBatchGateRequiresCompleteMatchedV2Provenance(t *testing.T) {
	tests := []struct {
		name         string
		bundleFS     func(t *testing.T) fstest.MapFS
		wantReason   string
		wantIncluded bool
	}{
		{
			name: "legacy v0 surface",
			bundleFS: func(t *testing.T) fstest.MapFS {
				t.Helper()
				return cliSurfaceBundleFS(validCLISurfaceJSON())
			},
			wantReason: "legacy v0",
		},
		{
			name: "missing endpoint provenance",
			bundleFS: func(t *testing.T) fstest.MapFS {
				t.Helper()
				endpoints := batchGateDefaultEndpoints(batchGateFixtureArtifactURL)
				endpoints[0].Provenance = nil
				return batchGateBundleFSWithSurface(t, validCLISurfaceJSON(), batchGateV2Surface(t, batchGateFixtureArtifactURL, endpoints))
			},
			wantReason: "provenance is required",
		},
		{
			name: "mismatched artifact URL",
			bundleFS: func(t *testing.T) fstest.MapFS {
				t.Helper()
				return batchGateBundleFSWithSurface(t, validCLISurfaceJSON(), batchGateV2Surface(t, "https://example.test/mismatched-artifact.json", batchGateDefaultEndpoints(batchGateFixtureArtifactURL)))
			},
			wantReason: "provenance artifact URL",
		},
		{
			name: "mismatched endpoint source URL",
			bundleFS: func(t *testing.T) fstest.MapFS {
				t.Helper()
				endpoints := batchGateDefaultEndpoints(batchGateFixtureArtifactURL)
				endpoints[0].Provenance.SourceURL = "https://example.test/other-provider-documentation.json"
				return batchGateBundleFSWithSurface(t, validCLISurfaceJSON(), batchGateV2Surface(t, batchGateFixtureArtifactURL, endpoints))
			},
			wantReason: "provenance source_url",
		},
		{
			name: "matched v2 provenance",
			bundleFS: func(t *testing.T) fstest.MapFS {
				t.Helper()
				return batchGateBundleFS(t, validCLISurfaceJSON())
			},
			wantIncluded: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defsRoot := t.TempDir()
			writeBatchBundle(t, defsRoot, test.bundleFS(t))
			manifestPath := writeBatchManifestFixture(t, "cli-surface")
			reportPath := filepath.Join(t.TempDir(), "report.json")
			var stdout, stderr bytes.Buffer

			code := run([]string{"batch", "gate", "--manifest", manifestPath, "--defs-root", defsRoot, "--report", reportPath}, &stdout, &stderr)
			report := readBatchGateReportFixture(t, reportPath)
			if test.wantIncluded {
				if code != 0 || report.ProvenanceRefusals != 0 || len(report.Included) != 1 || len(report.Dropped) != 0 {
					t.Fatalf("matched v2 gate = code %d report %+v, want included candidate; stdout=%s stderr=%s", code, report, stdout.String(), stderr.String())
				}
				return
			}
			if code == 0 || report.ProvenanceRefusals != 1 || len(report.Included) != 0 || len(report.Dropped) != 1 || report.Dropped[0].Stage != "provenance" || !strings.Contains(report.Dropped[0].Reason, test.wantReason) {
				t.Fatalf("provenance gate = code %d report %+v, want %q refusal; stdout=%s stderr=%s", code, report, test.wantReason, stdout.String(), stderr.String())
			}
		})
	}
}

func TestBatchGateReportsAggregateProvenanceRefusals(t *testing.T) {
	defsRoot := t.TempDir()
	for _, connector := range []string{"legacy-one", "legacy-two"} {
		writeNamedBatchBundle(t, defsRoot, connector, namedBatchBundleFS(t, connector, cliSurfaceBundleFS(validCLISurfaceJSON())))
	}
	manifestPath := writeBatchManifestFixture(t, "legacy-one", "legacy-two")
	reportPath := filepath.Join(t.TempDir(), "report.json")
	var stdout, stderr bytes.Buffer

	code := run([]string{"batch", "gate", "--manifest", manifestPath, "--defs-root", defsRoot, "--report", reportPath}, &stdout, &stderr)
	if code == 0 {
		t.Fatalf("batch gate exit = 0, want aggregate provenance refusals; stdout=%s stderr=%s", stdout.String(), stderr.String())
	}
	report := readBatchGateReportFixture(t, reportPath)
	if report.Candidates != 2 || report.ProvenanceRefusals != 2 || len(report.Included) != 0 || len(report.Dropped) != 2 {
		t.Fatalf("report = %+v, want two aggregate provenance refusals", report)
	}
	for _, drop := range report.Dropped {
		if drop.Stage != "provenance" {
			t.Fatalf("drop = %+v, want provenance stage", drop)
		}
	}
}

func TestBatchMaterializeGeneratesV2ProvenanceAndReachableSurface(t *testing.T) {
	sourceDefsRoot := t.TempDir()
	writeBatchBundle(t, sourceDefsRoot, cliSurfaceBundleFS(validCLISurfaceJSON()))
	defsRoot := t.TempDir()
	manifestPath := writeBatchManifestFixture(t, "cli-surface")
	artifactDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(artifactDir, "cli-surface.json"), []byte(`{
		"openapi": "3.0.0",
		"paths": {
			"/widgets": {
				"get": {"operationId": "listWidgets", "summary": "List widgets"},
				"post": {"operationId": "createWidget", "summary": "Create widget"}
			},
			"/widgets/{id}": {
				"delete": {"operationId": "deleteWidget", "summary": "Delete widget"}
			}
		}
	}`), 0o644); err != nil {
		t.Fatalf("write artifact: %v", err)
	}
	reportPath := filepath.Join(t.TempDir(), "materialize.json")
	var stdout, stderr bytes.Buffer

	code := run([]string{"batch", "materialize", "--manifest", manifestPath, "--source-defs-root", sourceDefsRoot, "--defs-root", defsRoot, "--artifact-dir", artifactDir, "--retrieved-at", "2026-08-06", "--report", reportPath}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("batch materialize exit = %d, want 0; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}

	surfaceRaw, err := os.ReadFile(filepath.Join(defsRoot, "cli-surface", "api_surface.json"))
	if err != nil {
		t.Fatalf("read generated api surface: %v", err)
	}
	var surface struct {
		OperationLedgerVersion int `json:"operation_ledger_version"`
		Artifacts              []struct {
			URL         string `json:"url"`
			RetrievedAt string `json:"retrieved_at"`
			SHA256      string `json:"sha256"`
		} `json:"artifacts"`
		Endpoints []struct {
			Method     string `json:"method"`
			Path       string `json:"path"`
			Provenance struct {
				Artifact  string `json:"artifact"`
				SourceURL string `json:"source_url"`
			} `json:"provenance"`
			CoveredBy any `json:"covered_by"`
			Operation any `json:"operation"`
		} `json:"endpoints"`
	}
	if err := json.Unmarshal(surfaceRaw, &surface); err != nil {
		t.Fatalf("decode generated api surface: %v", err)
	}
	if surface.OperationLedgerVersion != 2 || len(surface.Artifacts) != 1 || surface.Artifacts[0].URL != "https://example.test/cli-surface.json" || surface.Artifacts[0].RetrievedAt != "2026-08-06" || len(surface.Artifacts[0].SHA256) != 64 {
		t.Fatalf("generated provenance = %+v, want v2 artifact evidence", surface)
	}
	if len(surface.Endpoints) != 3 {
		t.Fatalf("generated endpoints = %d, want 3 artifact operations", len(surface.Endpoints))
	}
	for _, endpoint := range surface.Endpoints {
		if endpoint.Provenance.Artifact == "" || endpoint.Provenance.SourceURL != "https://example.test/cli-surface.json" {
			t.Fatalf("endpoint provenance = %+v, want artifact-local citation", endpoint)
		}
	}
	if surface.Endpoints[2].Method != "DELETE" || surface.Endpoints[2].Operation == nil || surface.Endpoints[2].CoveredBy != nil {
		t.Fatalf("defaulted DELETE endpoint = %+v, want explicitly blocked operation", surface.Endpoints[2])
	}

	opsRaw, err := os.ReadFile(filepath.Join(defsRoot, "cli-surface", "operations.json"))
	if err != nil {
		t.Fatalf("read generated operations: %v", err)
	}
	if string(opsRaw) != "{\n  \"operations\": []\n}\n" {
		t.Fatalf("operations.json = %s, want explicit no-direct-executor catalog", opsRaw)
	}

	cliRaw, err := os.ReadFile(filepath.Join(defsRoot, "cli-surface", "cli_surface.json"))
	if err != nil {
		t.Fatalf("read generated cli surface: %v", err)
	}
	var cli struct {
		Commands []struct {
			Path         string `json:"path"`
			Intent       string `json:"intent"`
			Availability string `json:"availability"`
			Stream       string `json:"stream"`
			Write        string `json:"write"`
			APISurface   []struct {
				Method string `json:"method"`
				Path   string `json:"path"`
			} `json:"api_surface"`
		} `json:"commands"`
	}
	if err := json.Unmarshal(cliRaw, &cli); err != nil {
		t.Fatalf("decode generated cli surface: %v", err)
	}
	if len(cli.Commands) != 2 || cli.Commands[0].Intent != "reverse_etl" || cli.Commands[0].Write != "create_widget" || cli.Commands[1].Intent != "etl" || cli.Commands[1].Stream != "widgets" {
		t.Fatalf("generated CLI commands = %+v, want reachable stream/write commands", cli.Commands)
	}
	if len(cli.Commands[0].APISurface) != 1 || len(cli.Commands[1].APISurface) != 1 {
		t.Fatalf("generated CLI command endpoint bindings = %+v, want one per command", cli.Commands)
	}

	if reportRaw, err := os.ReadFile(reportPath); err != nil || !strings.Contains(string(reportRaw), `"cli-surface"`) || !strings.Contains(string(reportRaw), `"included"`) {
		t.Fatalf("materialize report = %q, want included candidate report (err=%v)", reportRaw, err)
	}
}

func TestBatchMaterializeDropsMissingExecutableCoverageWithoutMutatingBundle(t *testing.T) {
	sourceDefsRoot := t.TempDir()
	writeBatchBundle(t, sourceDefsRoot, cliSurfaceBundleFS(validCLISurfaceJSON()))
	defsRoot := t.TempDir()
	manifestPath := writeBatchManifestFixture(t, "cli-surface")
	artifactDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(artifactDir, "cli-surface.json"), []byte(`{
		"openapi": "3.0.0",
		"paths": {"/different": {"get": {"summary": "Different endpoint"}}}
	}`), 0o644); err != nil {
		t.Fatalf("write artifact: %v", err)
	}
	reportPath := filepath.Join(t.TempDir(), "materialize.json")
	var stdout, stderr bytes.Buffer

	code := run([]string{"batch", "materialize", "--manifest", manifestPath, "--source-defs-root", sourceDefsRoot, "--defs-root", defsRoot, "--artifact-dir", artifactDir, "--retrieved-at", "2026-08-06", "--report", reportPath}, &stdout, &stderr)
	if code == 0 {
		t.Fatalf("batch materialize exit = 0, want missing coverage drop; stdout=%s stderr=%s", stdout.String(), stderr.String())
	}
	reportRaw, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("read materialize report: %v", err)
	}
	if !strings.Contains(string(reportRaw), `"coverage"`) || !strings.Contains(string(reportRaw), `"cli-surface"`) {
		t.Fatalf("materialize report = %s, want named coverage drop", reportRaw)
	}
	surfaceRaw, err := os.ReadFile(filepath.Join(sourceDefsRoot, "cli-surface", "api_surface.json"))
	if err != nil {
		t.Fatalf("read original api surface: %v", err)
	}
	if strings.Contains(string(surfaceRaw), `"operation_ledger_version": 2`) {
		t.Fatalf("materializer mutated dropped bundle api surface: %s", surfaceRaw)
	}
	if _, err := os.Stat(filepath.Join(sourceDefsRoot, "cli-surface", "operations.json")); !os.IsNotExist(err) {
		t.Fatalf("materializer wrote operations.json for dropped bundle: %v", err)
	}
}

func TestBatchMaterializeRequiresExactArtifactEndpointIdentity(t *testing.T) {
	tests := []struct {
		name           string
		sourceMethod   string
		sourcePath     string
		artifactMethod string
		artifactPath   string
		wantErr        bool
	}{
		{
			name:           "exact method and path",
			sourceMethod:   http.MethodGet,
			sourcePath:     "/widgets",
			artifactMethod: http.MethodGet,
			artifactPath:   "/widgets",
		},
		{
			name:           "trailing slash is distinct",
			sourceMethod:   http.MethodGet,
			sourcePath:     "/widgets",
			artifactMethod: http.MethodGet,
			artifactPath:   "/widgets/",
			wantErr:        true,
		},
		{
			name:           "method case is distinct",
			sourceMethod:   "get",
			sourcePath:     "/widgets",
			artifactMethod: http.MethodGet,
			artifactPath:   "/widgets",
			wantErr:        true,
		},
		{
			name:           "artifact method case is distinct",
			sourceMethod:   http.MethodGet,
			sourcePath:     "/widgets",
			artifactMethod: "get",
			artifactPath:   "/widgets",
			wantErr:        true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bundle := engine.Bundle{
				Surface: &engine.APISurface{
					API: "Example API",
					Endpoints: []engine.SurfaceEndpoint{{
						Method: test.sourceMethod,
						Path:   test.sourcePath,
						CoveredBy: &engine.SurfaceCoverage{
							Stream: "widgets",
						},
					}},
				},
				Streams: []engine.StreamSpec{{Name: "widgets"}},
			}
			surface, err := materializeAPISurface(
				bundle,
				BatchManifestConnector{Artifact: BatchArtifact{
					Kind:    "openapi",
					URL:     "https://example.test/openapi.json",
					Version: "1.0.0",
				}},
				"2026-08-06",
				"sha256",
				[]batchArtifactEndpoint{{Method: test.artifactMethod, Path: test.artifactPath}},
			)
			if test.wantErr {
				if err == nil || !strings.Contains(err.Error(), "absent from the cited artifact") {
					t.Fatalf("materialize error = %v, want exact-identity coverage rejection", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("materialize exact endpoint: %v", err)
			}
			if len(surface.Endpoints) != 1 || surface.Endpoints[0].CoveredBy == nil || surface.Endpoints[0].CoveredBy.Stream != "widgets" {
				t.Fatalf("materialized endpoints = %+v, want exact cited stream coverage", surface.Endpoints)
			}
		})
	}
}

func TestBatchMaterializeExcludesProtocolMetadataOperations(t *testing.T) {
	bundle := engine.Bundle{
		Surface: &engine.APISurface{
			API: "Example API",
			Endpoints: []engine.SurfaceEndpoint{
				{Method: http.MethodOptions, Path: "/diagnostics", CoveredBy: &engine.SurfaceCoverage{Stream: "diagnostics"}},
				{Method: http.MethodTrace, Path: "/diagnostics", CoveredBy: &engine.SurfaceCoverage{Stream: "diagnostics"}},
			},
		},
	}
	artifactEndpoints := []batchArtifactEndpoint{
		{Method: http.MethodOptions, Path: "/diagnostics"},
		{Method: http.MethodTrace, Path: "/diagnostics"},
	}

	surface, err := materializeAPISurface(
		bundle,
		BatchManifestConnector{Artifact: BatchArtifact{
			Kind:    "openapi",
			URL:     "https://example.test/openapi.json",
			Version: "1.0.0",
		}},
		"2026-08-06",
		"sha256",
		artifactEndpoints,
	)
	if err != nil {
		t.Fatalf("materialize protocol metadata: %v", err)
	}
	split, err := batchSurfaceSplit(&surface)
	if err != nil {
		t.Fatalf("split protocol metadata: %v", err)
	}
	if split != (BatchOperationSplit{Excluded: 2}) || split.total() != len(artifactEndpoints) {
		t.Fatalf("operation split = %+v, want two counted exclusions", split)
	}
	for _, endpoint := range surface.Endpoints {
		if endpoint.CoveredBy != nil || endpoint.Operation == nil {
			t.Fatalf("protocol endpoint = %+v, want excluded operation metadata only", endpoint)
		}
		if endpoint.Operation.Model == "disallowed" {
			t.Fatalf("protocol operation = %+v, want non-disallowed protocol metadata", endpoint.Operation)
		}
		if endpoint.Operation.Model != "local_workflow" || !endpoint.Operation.BlockedByDefault {
			t.Fatalf("protocol operation = %+v, want blocked protocol metadata", endpoint.Operation)
		}
		if !strings.Contains(endpoint.Operation.Reason, endpoint.Method) || !strings.Contains(endpoint.Operation.Reason, "protocol metadata") {
			t.Fatalf("protocol operation reason = %q, want method-specific metadata rationale", endpoint.Operation.Reason)
		}
	}
}

func TestBatchMaterializeDropsPreexistingDestinationBundle(t *testing.T) {
	sourceDefsRoot := t.TempDir()
	writeBatchBundle(t, sourceDefsRoot, cliSurfaceBundleFS(validCLISurfaceJSON()))
	defsRoot := t.TempDir()
	destination := writeBatchBundle(t, defsRoot, cliSurfaceBundleFS(validCLISurfaceJSON()))
	before, err := os.ReadFile(filepath.Join(destination, "api_surface.json"))
	if err != nil {
		t.Fatalf("read destination api surface before materialization: %v", err)
	}
	manifestPath := writeBatchManifestFixture(t, "cli-surface")
	artifactDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(artifactDir, "cli-surface.json"), []byte(`{
		"openapi": "3.0.0",
		"paths": {"/widgets": {"get": {"summary": "List widgets"}, "post": {"summary": "Create widget"}}}
	}`), 0o644); err != nil {
		t.Fatalf("write artifact: %v", err)
	}
	reportPath := filepath.Join(t.TempDir(), "materialize.json")
	var stdout, stderr bytes.Buffer

	code := run([]string{"batch", "materialize", "--manifest", manifestPath, "--source-defs-root", sourceDefsRoot, "--defs-root", defsRoot, "--artifact-dir", artifactDir, "--retrieved-at", "2026-08-06", "--report", reportPath}, &stdout, &stderr)
	if code == 0 {
		t.Fatalf("batch materialize exit = 0, want destination-collision drop; stdout=%s stderr=%s", stdout.String(), stderr.String())
	}
	report := readBatchGateReportFixture(t, reportPath)
	if len(report.Included) != 0 || len(report.Dropped) != 1 || report.Dropped[0].Stage != "bundle_collision" {
		t.Fatalf("materialize report = %+v, want named destination collision", report)
	}
	after, err := os.ReadFile(filepath.Join(destination, "api_surface.json"))
	if err != nil {
		t.Fatalf("read destination api surface after materialization: %v", err)
	}
	if string(after) != string(before) {
		t.Fatalf("materializer changed pre-existing destination bundle:\nbefore=%s\nafter=%s", before, after)
	}
}

func TestBatchMaterializeRejectsDestinationInsideSourceBundle(t *testing.T) {
	sourceDefsRoot := t.TempDir()
	source := writeBatchBundle(t, sourceDefsRoot, cliSurfaceBundleFS(validCLISurfaceJSON()))
	defsRoot := filepath.Join(source, "batch-output")
	manifestPath := writeBatchManifestFixture(t, "cli-surface")
	reportPath := filepath.Join(t.TempDir(), "materialize.json")
	var stdout, stderr bytes.Buffer

	code := run([]string{"batch", "materialize", "--manifest", manifestPath, "--source-defs-root", sourceDefsRoot, "--defs-root", defsRoot, "--retrieved-at", "2026-08-06", "--report", reportPath}, &stdout, &stderr)
	if code == 0 {
		t.Fatalf("batch materialize exit = 0, want source/destination overlap drop; stdout=%s stderr=%s", stdout.String(), stderr.String())
	}
	report := readBatchGateReportFixture(t, reportPath)
	if len(report.Included) != 0 || len(report.Dropped) != 1 || report.Dropped[0].Stage != "bundle_path" || !strings.Contains(report.Dropped[0].Reason, "must not overlap") {
		t.Fatalf("materialize report = %+v, want source/destination overlap drop", report)
	}
	if _, err := os.Stat(filepath.Join(source, "batch-output")); !os.IsNotExist(err) {
		t.Fatalf("materializer created a destination inside the source bundle: %v", err)
	}
}

func TestParseBatchOpenAPIArtifactAcceptsYAMLResponseStatusKeys(t *testing.T) {
	artifact := []byte(`openapi: 3.0.3
paths:
  /widgets:
    get:
      summary: List widgets
      responses:
        200:
          description: OK
`)

	endpoints, err := parseBatchOpenAPIArtifact(artifact)
	if err != nil {
		t.Fatalf("parse YAML artifact: %v", err)
	}
	if len(endpoints) != 1 || endpoints[0].Method != "GET" || endpoints[0].Path != "/widgets" || endpoints[0].Summary != "List widgets" {
		t.Fatalf("parsed endpoints = %+v, want cited GET /widgets", endpoints)
	}
}

func TestParseBatchOpenAPIArtifactResolvesLocalPathItemReferencesAndTrace(t *testing.T) {
	artifact := []byte(`{
		"openapi": "3.1.0",
		"paths": {
			"/widgets": {"$ref": "#/components/pathItems/widgets"},
			"/diagnostics": {"trace": {"summary": "Trace diagnostics"}}
		},
		"components": {
			"pathItems": {
				"widgets": {"get": {"operationId": "listWidgets"}}
			}
		}
	}`)

	endpoints, err := parseBatchOpenAPIArtifact(artifact)
	if err != nil {
		t.Fatalf("parse local-ref artifact: %v", err)
	}
	if len(endpoints) != 2 || endpoints[0].Method != "TRACE" || endpoints[0].Path != "/diagnostics" || endpoints[1].Method != "GET" || endpoints[1].Path != "/widgets" {
		t.Fatalf("parsed endpoints = %+v, want TRACE and local-ref GET operations", endpoints)
	}
}

func TestParseBatchOpenAPIArtifactFailsClosedForUnsupportedOperationContainers(t *testing.T) {
	tests := []struct {
		name     string
		artifact string
		want     string
	}{
		{
			name: "webhooks",
			artifact: `{
				"openapi": "3.1.0",
				"paths": {"/widgets": {"get": {"summary": "List widgets"}}},
				"webhooks": {"widget.created": {"post": {"summary": "Widget created"}}}
			}`,
			want: "top-level webhooks",
		},
		{
			name: "external path item reference",
			artifact: `{
				"openapi": "3.1.0",
				"paths": {"/widgets": {"$ref": "other.yaml#/components/pathItems/widgets"}}
			}`,
			want: "external path-item reference",
		},
		{
			name: "noncanonical operation key",
			artifact: `{
				"openapi": "3.1.0",
				"paths": {"/widgets": {"GET": {"summary": "List widgets"}}}
			}`,
			want: "unsupported path-item field \"GET\"",
		},
		{
			name: "multiple YAML documents",
			artifact: `openapi: 3.1.0
paths:
  /widgets:
    get:
      summary: List widgets
---
openapi: 3.1.0
paths:
  /other:
    get:
      summary: List other widgets
`,
			want: "multiple YAML documents",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := parseBatchOpenAPIArtifact([]byte(test.artifact))
			if err == nil || !strings.Contains(err.Error(), "artifact operation inventory is unknown") || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("parse error = %v, want concrete unknown-inventory reason containing %q", err, test.want)
			}
		})
	}
}

func TestBatchArtifactURLAndDestinationGuards(t *testing.T) {
	for _, raw := range []string{
		"https://user@example.test/openapi.json",
		"https://example.test/openapi.json?token=value",
		"https://example.test/openapi.json#access_token=value",
		"https://127.0.0.1/openapi.json",
	} {
		if err := validateBatchArtifactURL(raw); err == nil {
			t.Fatalf("validateBatchArtifactURL(%q) succeeded, want rejection", raw)
		}
	}
	privateLookup := func(context.Context, string) ([]net.IPAddr, error) {
		return []net.IPAddr{{IP: net.ParseIP("169.254.169.254")}}, nil
	}
	if _, err := batchArtifactPublicAddresses(context.Background(), "artifact.example", privateLookup); err == nil {
		t.Fatal("private resolved destination was accepted")
	}
	redirectURL, err := url.Parse("https://artifact.example/redirected.json")
	if err != nil {
		t.Fatalf("parse redirect URL: %v", err)
	}
	if err := newBatchArtifactHTTPClient(privateLookup).CheckRedirect(&http.Request{URL: redirectURL}, nil); err == nil {
		t.Fatal("redirect to private resolved destination was accepted")
	}
	dialed := ""
	local, remote := net.Pipe()
	t.Cleanup(func() {
		if err := remote.Close(); err != nil {
			t.Errorf("close remote pipe: %v", err)
		}
	})
	connection, err := dialBatchArtifactAddress(context.Background(), "tcp", "artifact.example:443", func(context.Context, string) ([]net.IPAddr, error) {
		return []net.IPAddr{{IP: net.ParseIP("8.8.8.8")}}, nil
	}, func(_ context.Context, _ string, address string) (net.Conn, error) {
		dialed = address
		return local, nil
	})
	if err != nil {
		t.Fatalf("dial public destination: %v", err)
	}
	t.Cleanup(func() {
		if err := connection.Close(); err != nil {
			t.Errorf("close batch artifact connection: %v", err)
		}
	})
	if dialed != "8.8.8.8:443" {
		t.Fatalf("dialed %q, want validated public address", dialed)
	}
}

func TestBatchArtifactResponseRejectsIncompleteInventories(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		headers  http.Header
		contains string
	}{
		{
			name:     "partial content status",
			status:   http.StatusPartialContent,
			headers:  http.Header{},
			contains: "HTTP 206",
		},
		{
			name:   "content range header",
			status: http.StatusOK,
			headers: http.Header{
				"Content-Range": []string{"bytes 0-1023/4096"},
			},
			contains: "Content-Range",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateBatchArtifactResponse(&http.Response{StatusCode: test.status, Header: test.headers})
			var unknown *batchArtifactInventoryUnknownError
			if !errors.As(err, &unknown) || !strings.Contains(err.Error(), "incomplete") || !strings.Contains(err.Error(), test.contains) {
				t.Fatalf("response error = %v, want incomplete unknown inventory containing %q", err, test.contains)
			}
			if stage := batchArtifactDropStage(err, "artifact_fetch"); stage != "artifact_inventory_unknown" {
				t.Fatalf("drop stage = %q, want artifact_inventory_unknown", stage)
			}
		})
	}
	if err := validateBatchArtifactResponse(&http.Response{StatusCode: http.StatusOK, Header: http.Header{}}); err != nil {
		t.Fatalf("complete artifact response rejected: %v", err)
	}
}

func TestBatchMaterializeConvertsLegacyExclusionsAndDerivesWriteFlags(t *testing.T) {
	sourceDefsRoot := t.TempDir()
	defsRoot := t.TempDir()
	fsys := cliSurfaceBundleFS(validCLISurfaceJSON())
	fsys["cli-surface/api_surface.json"] = &fstest.MapFile{Data: []byte(`{
		"api": "test API v1",
		"endpoints": [
			{ "method": "GET", "path": "/widgets", "covered_by": { "stream": "widgets" } },
			{ "method": "POST", "path": "/widgets", "covered_by": { "write": "create_widget" } },
			{ "method": "GET", "path": "/widgets/{id}", "excluded": { "category": "duplicate_of", "reason": "The widgets stream already returns the complete record." } }
		]
	}`)}
	writeBatchBundle(t, sourceDefsRoot, fsys)
	manifestPath := writeBatchManifestFixture(t, "cli-surface")
	artifactDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(artifactDir, "cli-surface.json"), []byte(`{
		"openapi": "3.0.0",
		"paths": {
			"/widgets": { "get": {"summary": "List widgets"}, "post": {"summary": "Create widget"} },
			"/widgets/{id}": { "get": {"summary": "Get widget"} }
		}
	}`), 0o644); err != nil {
		t.Fatalf("write artifact: %v", err)
	}
	reportPath := filepath.Join(t.TempDir(), "materialize.json")
	var stdout, stderr bytes.Buffer

	code := run([]string{"batch", "materialize", "--manifest", manifestPath, "--source-defs-root", sourceDefsRoot, "--defs-root", defsRoot, "--artifact-dir", artifactDir, "--retrieved-at", "2026-08-06", "--report", reportPath}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("batch materialize exit = %d, want 0; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	validation, err := validatePath(filepath.Join(defsRoot, "cli-surface"))
	if err != nil {
		t.Fatalf("validate materialized bundle: %v", err)
	}
	if len(validation.Findings) != 0 {
		t.Fatalf("materialized bundle findings = %+v, want no legacy classifier or write-flag drift", validation.Findings)
	}

	cliRaw, err := os.ReadFile(filepath.Join(defsRoot, "cli-surface", "cli_surface.json"))
	if err != nil {
		t.Fatalf("read generated CLI surface: %v", err)
	}
	if !strings.Contains(string(cliRaw), `"maps_to": "record.name"`) || !strings.Contains(string(cliRaw), `"required": true`) {
		t.Fatalf("generated CLI write flags = %s, want required record.name binding", cliRaw)
	}
	surfaceRaw, err := os.ReadFile(filepath.Join(defsRoot, "cli-surface", "api_surface.json"))
	if err != nil {
		t.Fatalf("read generated api surface: %v", err)
	}
	if strings.Contains(string(surfaceRaw), `"excluded"`) || !strings.Contains(string(surfaceRaw), `"operation"`) {
		t.Fatalf("generated v2 api surface = %s, want legacy exclusion converted to blocked operation", surfaceRaw)
	}
}

// TestBatchMaterializePluralOnlyWriteCoverage drives the real `batch
// materialize` command over a bundle whose covered_by rows use ONLY the plural
// `writes` spelling, so it exercises batchSurfaceSplit (the api_surface gate),
// materializeAPISurface/copyMaterializedClassifier, ensureMaterializedCoverage
// and materializeCLISurface in one pass.
//
// Without the plural migration each of those three fails a different way:
// batchSurfaceSplit reports "has an empty covered_by classifier",
// ensureMaterializedCoverage reports "has no matching endpoint in the cited
// artifact", and materializeCLISurface reports "has no materialized
// api_surface reference".
//
// POST /widgets carries TWO write actions in one array — the cardinality the
// shared foundation exists for, which no shipped bundle currently uses (all
// 544 jira/workday-rest arrays are singletons). PUT /widgets/{id} carries a
// single-element array, the form those bundles actually ship.
func TestBatchMaterializePluralOnlyWriteCoverage(t *testing.T) {
	sourceDefsRoot := t.TempDir()
	defsRoot := t.TempDir()
	fsys := cliSurfaceBundleFS(validCLISurfaceJSON())
	fsys["cli-surface/writes.json"] = &fstest.MapFile{Data: []byte(`{
		"actions": [
			{
				"name": "create_widget",
				"kind": "create",
				"method": "POST",
				"path": "/widgets",
				"record_schema": { "type": "object", "required": ["name"], "properties": { "name": { "type": "string" } } },
				"risk": "creates a widget"
			},
			{
				"name": "import_widget",
				"kind": "create",
				"method": "POST",
				"path": "/widgets",
				"record_schema": { "type": "object", "required": ["source_id"], "properties": { "source_id": { "type": "string" } } },
				"risk": "imports a widget from an external system"
			},
			{
				"name": "rename_widget",
				"kind": "update",
				"method": "PUT",
				"path": "/widgets/{id}",
				"path_fields": ["id"],
				"record_schema": { "type": "object", "required": ["id", "name"], "properties": { "id": { "type": "string" }, "name": { "type": "string" } } },
				"risk": "renames a widget"
			}
		]
	}`)}
	fsys["cli-surface/api_surface.json"] = &fstest.MapFile{Data: []byte(`{
		"api": "test API v1",
		"endpoints": [
			{ "method": "GET", "path": "/widgets", "covered_by": { "stream": "widgets" } },
			{ "method": "POST", "path": "/widgets", "covered_by": { "writes": ["create_widget", "import_widget"] } },
			{ "method": "PUT", "path": "/widgets/{id}", "covered_by": { "writes": ["rename_widget"] } }
		]
	}`)}
	writeBatchBundle(t, sourceDefsRoot, fsys)
	manifestPath := writeBatchManifestFixture(t, "cli-surface")
	artifactDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(artifactDir, "cli-surface.json"), []byte(`{
		"openapi": "3.0.0",
		"paths": {
			"/widgets": { "get": {"summary": "List widgets"}, "post": {"summary": "Create widget"} },
			"/widgets/{id}": { "put": {"summary": "Rename widget"} }
		}
	}`), 0o644); err != nil {
		t.Fatalf("write artifact: %v", err)
	}
	reportPath := filepath.Join(t.TempDir(), "materialize.json")
	var stdout, stderr bytes.Buffer

	code := run([]string{"batch", "materialize", "--manifest", manifestPath, "--source-defs-root", sourceDefsRoot, "--defs-root", defsRoot, "--artifact-dir", artifactDir, "--retrieved-at", "2026-08-06", "--report", reportPath}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("batch materialize exit = %d, want 0; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}

	validation, err := validatePath(filepath.Join(defsRoot, "cli-surface"))
	if err != nil {
		t.Fatalf("validate materialized bundle: %v", err)
	}
	if len(validation.Findings) != 0 {
		t.Fatalf("materialized bundle findings = %+v, want none for plural-only coverage", validation.Findings)
	}

	surfaceRaw, err := os.ReadFile(filepath.Join(defsRoot, "cli-surface", "api_surface.json"))
	if err != nil {
		t.Fatalf("read generated api surface: %v", err)
	}
	var surface engine.APISurface
	if err := json.Unmarshal(surfaceRaw, &surface); err != nil {
		t.Fatalf("decode generated api surface: %v", err)
	}
	// batchSurfaceSplit must count the plural rows as executable, not as an
	// empty classifier.
	split, err := batchSurfaceSplit(&surface)
	if err != nil {
		t.Fatalf("batchSurfaceSplit over plural-only coverage: %v", err)
	}
	if split.Executable != 3 {
		t.Errorf("split.Executable = %d, want 3 (1 stream + 2 plural write rows)", split.Executable)
	}

	writeTargets := map[string][]string{}
	for _, endpoint := range surface.Endpoints {
		if endpoint.CoveredBy == nil {
			continue
		}
		writeTargets[endpoint.Method+" "+endpoint.Path] = endpoint.CoveredBy.WriteTargets()
	}
	if got := writeTargets["POST /widgets"]; len(got) != 2 {
		t.Errorf("POST /widgets write targets = %v, want both plural actions preserved through materialization", got)
	}
	if got := writeTargets["PUT /widgets/{id}"]; len(got) != 1 {
		t.Errorf("PUT /widgets/{id} write targets = %v, want the single-element plural array preserved", got)
	}

	// materializeCLISurface must resolve an api_surface reference for EVERY
	// declared write action, including the two that share one endpoint.
	cliRaw, err := os.ReadFile(filepath.Join(defsRoot, "cli-surface", "cli_surface.json"))
	if err != nil {
		t.Fatalf("read generated CLI surface: %v", err)
	}
	var cli engine.CLISurface
	if err := json.Unmarshal(cliRaw, &cli); err != nil {
		t.Fatalf("decode generated CLI surface: %v", err)
	}
	byWrite := map[string]engine.CLICommand{}
	for _, cmd := range cli.Commands {
		if cmd.Write != "" {
			byWrite[cmd.Write] = cmd
		}
	}
	for _, action := range []string{"create_widget", "import_widget", "rename_widget"} {
		cmd, ok := byWrite[action]
		if !ok {
			t.Fatalf("generated CLI surface has no command for write %q; commands=%s", action, cliRaw)
		}
		if len(cmd.APISurface) == 0 {
			t.Errorf("write %q command has no api_surface reference", action)
		}
	}
	if ref := byWrite["create_widget"].APISurface; len(ref) != 1 || ref[0].Path != "/widgets" {
		t.Errorf("create_widget api_surface = %+v, want the shared POST /widgets row", ref)
	}
	if ref := byWrite["import_widget"].APISurface; len(ref) != 1 || ref[0].Path != "/widgets" {
		t.Errorf("import_widget api_surface = %+v, want the same shared POST /widgets row", ref)
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
	Candidates         int `json:"candidates"`
	ProvenanceRefusals int `json:"provenance_refusals"`
	Included           []struct {
		Connector          string              `json:"connector"`
		CommandsChecked    int                 `json:"commands_checked"`
		DeclaredOperations int                 `json:"declared_operations"`
		OperationSplit     BatchOperationSplit `json:"operation_split"`
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
	return writeNamedBatchBundle(t, defsRoot, "cli-surface", fsys)
}

func writeNamedBatchBundle(t *testing.T, defsRoot, name string, fsys fstest.MapFS) string {
	t.Helper()
	const fixturePrefix = "cli-surface/"
	for source, file := range fsys {
		rel, ok := strings.CutPrefix(source, fixturePrefix)
		if !ok {
			t.Fatalf("fixture file %q does not start with %q", source, fixturePrefix)
		}
		path := filepath.Join(defsRoot, name, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
		}
		if err := os.WriteFile(path, file.Data, 0o644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}
	return filepath.Join(defsRoot, name)
}

func namedBatchBundleFS(t *testing.T, name string, fsys fstest.MapFS) fstest.MapFS {
	t.Helper()
	clone := make(fstest.MapFS, len(fsys))
	for path, file := range fsys {
		copy := *file
		copy.Data = append([]byte(nil), file.Data...)
		clone[path] = &copy
	}
	metadata := clone["cli-surface/metadata.json"]
	updated := strings.Replace(string(metadata.Data), `"name": "cli-surface"`, `"name": "`+name+`"`, 1)
	if updated == string(metadata.Data) {
		t.Fatalf("fixture metadata did not contain cli-surface name")
	}
	metadata.Data = []byte(updated)
	return clone
}

const (
	batchGateFixtureArtifactID  = "cli-surface-artifact-2026-08-06"
	batchGateFixtureArtifactURL = "https://example.test/cli-surface.json"
	batchGateFixtureSHA         = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
)

func batchGateBundleFS(t *testing.T, cliSurface string) fstest.MapFS {
	t.Helper()
	return batchGateBundleFSWithSurface(t, cliSurface, batchGateV2Surface(t, batchGateFixtureArtifactURL, batchGateDefaultEndpoints(batchGateFixtureArtifactURL)))
}

func batchGateBundleFSWithSurface(t *testing.T, cliSurface, surface string) fstest.MapFS {
	t.Helper()
	fsys := cliSurfaceBundleFS(cliSurface)
	fsys["cli-surface/api_surface.json"] = &fstest.MapFile{Data: []byte(surface)}
	return fsys
}

func batchGateV2Surface(t *testing.T, artifactURL string, endpoints []engine.SurfaceEndpoint) string {
	t.Helper()
	raw, err := json.Marshal(engine.APISurface{
		API:                    "test API v2",
		OperationLedgerVersion: 2,
		Artifacts: []engine.SurfaceArtifact{{
			ID:          batchGateFixtureArtifactID,
			URL:         artifactURL,
			RetrievedAt: "2026-08-06",
			SHA256:      batchGateFixtureSHA,
		}},
		Endpoints: endpoints,
	})
	if err != nil {
		t.Fatalf("marshal batch gate v2 surface: %v", err)
	}
	return string(raw)
}

func batchGateFixtureProvenance(artifactURL string) *engine.SurfaceProvenance {
	return &engine.SurfaceProvenance{
		Artifact:  batchGateFixtureArtifactID,
		SourceURL: artifactURL,
	}
}

func batchGateDefaultEndpoints(artifactURL string) []engine.SurfaceEndpoint {
	return []engine.SurfaceEndpoint{
		{
			Method:     http.MethodGet,
			Path:       "/widgets",
			Provenance: batchGateFixtureProvenance(artifactURL),
			CoveredBy:  &engine.SurfaceCoverage{Stream: "widgets"},
		},
		{
			Method:     http.MethodPost,
			Path:       "/widgets",
			Provenance: batchGateFixtureProvenance(artifactURL),
			CoveredBy:  &engine.SurfaceCoverage{Write: "create_widget"},
		},
	}
}

func batchGateDirectWriteSurface(t *testing.T) string {
	t.Helper()
	endpoints := append(batchGateDefaultEndpoints(batchGateFixtureArtifactURL), engine.SurfaceEndpoint{
		Method:     http.MethodPost,
		Path:       "/widgets/{id}/archive",
		Provenance: batchGateFixtureProvenance(batchGateFixtureArtifactURL),
		Operation: &engine.SurfaceOperation{
			Model:            "destructive_action",
			Status:           "blocked",
			Risk:             "high",
			BlockedByDefault: true,
			Reason:           "Typed rest_write operation",
			Notes:            "Declared only through the direct-write executor.",
		},
	})
	return batchGateV2Surface(t, batchGateFixtureArtifactURL, endpoints)
}

func batchSurfaceSyncDriftBundleFS(t *testing.T) fstest.MapFS {
	t.Helper()
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
	endpoints := append(batchGateDefaultEndpoints(batchGateFixtureArtifactURL), engine.SurfaceEndpoint{
		Method:     http.MethodGet,
		Path:       "/artifacts/{id}",
		Provenance: batchGateFixtureProvenance(batchGateFixtureArtifactURL),
		Operation: &engine.SurfaceOperation{
			Model:            "binary_read",
			Status:           "blocked",
			Risk:             "low",
			BlockedByDefault: true,
			Reason:           "typed binary operation metadata",
		},
	})
	fsys["cli-surface/api_surface.json"] = &fstest.MapFile{Data: []byte(batchGateV2Surface(t, batchGateFixtureArtifactURL, endpoints))}
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
