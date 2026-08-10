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
	"reflect"
	"strconv"
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

func TestBatchPlanRejectsArtifactKindsWithoutParsers(t *testing.T) {
	for _, kind := range []string{"api_blueprint", "graphql", "asyncapi", "wsdl"} {
		t.Run(kind, func(t *testing.T) {
			record := batchLedgerRecordFixture("acme", 23, "2026-08-06")
			record["artifact_kind"] = kind
			ledger := writeBatchLedger(t, []map[string]any{record})
			out := filepath.Join(t.TempDir(), "batch.json")
			var stdout, stderr bytes.Buffer

			code := run([]string{"batch", "plan", "--ledger", ledger, "--out", out, "--connector", "acme"}, &stdout, &stderr)
			if code == 0 {
				t.Fatalf("batch plan accepted unsupported artifact kind %q; stdout=%s stderr=%s", kind, stdout.String(), stderr.String())
			}
			if !strings.Contains(stderr.String(), "artifact_kind is "+strconv.Quote(kind)) {
				t.Fatalf("batch plan stderr = %q, want unsupported artifact kind %q", stderr.String(), kind)
			}
		})
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

func TestBatchPlanCanonicalizesProviderReferenceURL(t *testing.T) {
	record := batchLedgerRecordFixture("acme", 23, "2026-08-06")
	record["provider_reference_url"] = " https://example.test/reference "
	ledger := writeBatchLedger(t, []map[string]any{record})
	manifestPath := filepath.Join(t.TempDir(), "batch.json")
	var stdout, stderr bytes.Buffer

	if code := run([]string{"batch", "plan", "--ledger", ledger, "--out", manifestPath, "--connector", "acme"}, &stdout, &stderr); code != 0 {
		t.Fatalf("batch plan exit = %d, want 0; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	raw, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	if !strings.Contains(string(raw), `"provider_reference_url": "https://example.test/reference"`) {
		t.Fatalf("planned manifest = %s, want canonical provider reference URL", raw)
	}
}

func TestReadBatchManifestCanonicalizesProviderReferenceURL(t *testing.T) {
	path := writeBatchManifestFixture(t, "cli-surface")
	manifest, err := readBatchManifest(path)
	if err != nil {
		t.Fatalf("read fixture manifest: %v", err)
	}
	manifest.Connectors[0].ProviderReferenceURL = " https://example.test/reference "
	if err := writeBatchManifest(path, manifest); err != nil {
		t.Fatalf("write manifest with padded reference: %v", err)
	}
	manifest, err = readBatchManifest(path)
	if err != nil {
		t.Fatalf("read manifest with padded reference: %v", err)
	}
	if got := manifest.Connectors[0].ProviderReferenceURL; got != "https://example.test/reference" {
		t.Fatalf("provider reference URL = %q, want canonical URL", got)
	}
}

func TestReadBatchManifestAllowsDocumentedPartnerAccess(t *testing.T) {
	path := writeBatchManifestFixture(t, "cli-surface")
	manifest, err := readBatchManifest(path)
	if err != nil {
		t.Fatalf("read public manifest: %v", err)
	}
	manifest.Connectors[0].AccessModel = "partner_gated"
	if err := writeBatchManifest(path, manifest); err != nil {
		t.Fatalf("write documented partner manifest: %v", err)
	}
	if _, err := readBatchManifest(path); err != nil {
		t.Fatalf("read documented partner manifest: %v", err)
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
	setBatchManifestOperationCounts(t, manifestPath, 3, 1, 2)
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
	var operations struct {
		Operations []engine.OperationSpec `json:"operations"`
	}
	if err := json.Unmarshal(opsRaw, &operations); err != nil {
		t.Fatalf("decode operations.json: %v", err)
	}
	if len(operations.Operations) != 1 || operations.Operations[0].REST == nil || operations.Operations[0].REST.Method != http.MethodDelete {
		t.Fatalf("operations.json = %s, want typed metadata for the documented DELETE", opsRaw)
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
	if len(cli.Commands) != 3 {
		t.Fatalf("generated CLI commands = %+v, want stream, write, and blocked artifact operation", cli.Commands)
	}
	retainedTargets := map[string]string{}
	for _, command := range cli.Commands {
		switch command.Path {
		case "widget list":
			retainedTargets[command.Path] = command.Stream
		case "widget create":
			retainedTargets[command.Path] = command.Write
		}
	}
	if !reflect.DeepEqual(retainedTargets, map[string]string{
		"widget list":   "widgets",
		"widget create": "create_widget",
	}) {
		t.Fatalf("generated CLI commands = %+v, want existing ETL and reverse-ETL command paths retained", cli.Commands)
	}
	if len(cli.Commands[0].APISurface) != 1 || len(cli.Commands[1].APISurface) != 1 || len(cli.Commands[2].APISurface) != 1 {
		t.Fatalf("generated CLI command endpoint bindings = %+v, want one per command", cli.Commands)
	}
	var blocked int
	for _, command := range cli.Commands {
		if command.Availability == materializeAvailabilityNotImplemented {
			blocked++
		}
	}
	if blocked != 1 {
		t.Fatalf("generated CLI commands = %+v, want one named-dependency command", cli.Commands)
	}

	if reportRaw, err := os.ReadFile(reportPath); err != nil || !strings.Contains(string(reportRaw), `"cli-surface"`) || !strings.Contains(string(reportRaw), `"included"`) {
		t.Fatalf("materialize report = %q, want included candidate report (err=%v)", reportRaw, err)
	}
}

func TestBatchMaterializeUsesExactExistingSurfaceEvidence(t *testing.T) {
	sourceDefsRoot := t.TempDir()
	fsys := cliSurfaceBundleFS(validCLISurfaceJSON())
	fsys["cli-surface/api_surface.json"] = &fstest.MapFile{Data: []byte(`{
		"api": "test API v1",
		"scope": "The preserved provider inventory was exhaustively counted from the official reference.",
		"endpoints": [
			{ "method": "GET", "path": "/widgets", "covered_by": { "stream": "widgets" } },
			{ "method": "POST", "path": "/widgets", "covered_by": { "write": "create_widget" } },
			{ "method": "DELETE", "path": "/widgets/{id}", "operation": { "model": "destructive_action", "status": "blocked", "risk": "high", "blocked_by_default": true, "reason": "The official provider operation requires explicit approval." } }
		]
	}`)}
	writeBatchBundle(t, sourceDefsRoot, fsys)
	manifestPath := writeBatchManifestFixture(t, "cli-surface")
	setBatchManifestOperationCounts(t, manifestPath, 3, 1, 2)
	artifactDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(artifactDir, "cli-surface.txt"), []byte("official provider reference index"), 0o644); err != nil {
		t.Fatalf("write official artifact: %v", err)
	}
	defsRoot := t.TempDir()
	reportPath := filepath.Join(t.TempDir(), "materialize.json")
	var stdout, stderr bytes.Buffer

	code := run([]string{"batch", "materialize", "--manifest", manifestPath, "--source-defs-root", sourceDefsRoot, "--defs-root", defsRoot, "--artifact-dir", artifactDir, "--retrieved-at", "2026-08-09", "--report", reportPath, "--existing-surface-evidence"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("batch materialize exit = %d, want 0; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	var report BatchMaterializeReport
	decodeJSONFile(t, reportPath, &report)
	if len(report.Included) != 1 || report.Included[0].EvidenceMode != "existing_surface_evidence" || report.Included[0].ArtifactOperations != 3 {
		t.Fatalf("materialize report = %+v, want exact existing-surface evidence inclusion", report)
	}
	var surface engine.APISurface
	decodeJSONFile(t, filepath.Join(defsRoot, "cli-surface", "api_surface.json"), &surface)
	if surface.OperationLedgerVersion != 2 || len(surface.Endpoints) != 3 || !strings.Contains(surface.Scope, "Exact existing-surface evidence fallback") {
		t.Fatalf("materialized surface = %+v, want v2 exact existing-surface evidence", surface)
	}
	for _, endpoint := range surface.Endpoints {
		if endpoint.Provenance == nil || endpoint.Provenance.SourceURL != "https://example.test/cli-surface.json" || !strings.HasPrefix(endpoint.Provenance.Coordinate, "existing-complete-surface[") {
			t.Fatalf("endpoint provenance = %+v, want cited exact existing-surface evidence", endpoint.Provenance)
		}
	}
	validation, err := validatePath(filepath.Join(defsRoot, "cli-surface"))
	if err != nil || len(validation.Findings) != 0 {
		t.Fatalf("validate exact existing-surface materialization = %+v, %v", validation.Findings, err)
	}

	setBatchManifestOperationCounts(t, manifestPath, 4, 2, 2)
	mismatchReportPath := filepath.Join(t.TempDir(), "mismatch.json")
	stdout.Reset()
	stderr.Reset()
	code = run([]string{"batch", "materialize", "--manifest", manifestPath, "--source-defs-root", sourceDefsRoot, "--defs-root", t.TempDir(), "--artifact-dir", artifactDir, "--retrieved-at", "2026-08-09", "--report", mismatchReportPath, "--existing-surface-evidence"}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("mismatched existing-surface materialize exit = %d, want 1; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	decodeJSONFile(t, mismatchReportPath, &report)
	if len(report.Included) != 0 || len(report.Dropped) != 1 || report.Dropped[0].Stage != "existing_surface_evidence" || !strings.Contains(report.Dropped[0].Reason, "3 operation(s), not the immutable ledger count 4") {
		t.Fatalf("mismatch materialize report = %+v, want exact-count refusal", report)
	}
}

func TestBatchMaterializeAllowsExactUnversionedOfficialReference(t *testing.T) {
	sourceDefsRoot := t.TempDir()
	fsys := cliSurfaceBundleFS(validCLISurfaceJSON())
	fsys["cli-surface/api_surface.json"] = &fstest.MapFile{Data: []byte(`{
		"api": "test API v1",
		"scope": "The preserved provider inventory was exhaustively counted from the official reference.",
		"endpoints": [
			{ "method": "GET", "path": "/widgets", "covered_by": { "stream": "widgets" } },
			{ "method": "POST", "path": "/widgets", "covered_by": { "write": "create_widget" } },
			{ "method": "DELETE", "path": "/widgets/{id}", "operation": { "model": "destructive_action", "status": "blocked", "risk": "high", "blocked_by_default": true, "reason": "The official provider operation requires explicit approval." } }
		]
	}`)}
	writeBatchBundle(t, sourceDefsRoot, fsys)
	manifestPath := writeBatchManifestFixture(t, "cli-surface")
	setBatchManifestOperationCounts(t, manifestPath, 3, 1, 2)

	rawManifest, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("read manifest fixture: %v", err)
	}
	var manifest map[string]any
	if err := json.Unmarshal(rawManifest, &manifest); err != nil {
		t.Fatalf("decode manifest fixture: %v", err)
	}
	connectors, ok := manifest["connectors"].([]any)
	if !ok || len(connectors) != 1 {
		t.Fatalf("manifest connectors = %#v, want one connector", manifest["connectors"])
	}
	connector, ok := connectors[0].(map[string]any)
	if !ok {
		t.Fatalf("manifest connector = %#v, want object", connectors[0])
	}
	artifact, ok := connector["artifact"].(map[string]any)
	if !ok {
		t.Fatalf("manifest artifact = %#v, want object", connector["artifact"])
	}
	artifact["url"] = "https://example.test/reference"
	artifact["kind"] = "html_reference"
	artifact["version"] = ""
	artifact["unversioned_official_reference"] = true
	rawManifest, err = json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Fatalf("encode unversioned manifest fixture: %v", err)
	}
	if err := os.WriteFile(manifestPath, append(rawManifest, '\n'), 0o644); err != nil {
		t.Fatalf("write unversioned manifest fixture: %v", err)
	}

	artifactDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(artifactDir, "cli-surface.txt"), []byte("official provider reference index"), 0o644); err != nil {
		t.Fatalf("write official artifact: %v", err)
	}
	defsRoot := t.TempDir()
	reportPath := filepath.Join(t.TempDir(), "materialize.json")
	var stdout, stderr bytes.Buffer

	code := run([]string{"batch", "materialize", "--manifest", manifestPath, "--source-defs-root", sourceDefsRoot, "--defs-root", defsRoot, "--artifact-dir", artifactDir, "--retrieved-at", "2026-08-09", "--report", reportPath, "--existing-surface-evidence"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("unversioned exact-evidence materialize exit = %d, want 0; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	var report BatchMaterializeReport
	decodeJSONFile(t, reportPath, &report)
	if len(report.Included) != 1 || report.Included[0].Artifact.Version != "provider-publishes-no-version-marker" {
		t.Fatalf("unversioned materialize report = %+v, want explicit no-version provenance", report)
	}
	var surface engine.APISurface
	decodeJSONFile(t, filepath.Join(defsRoot, "cli-surface", "api_surface.json"), &surface)
	if !strings.Contains(surface.Scope, "provider publishes no version marker") {
		t.Fatalf("unversioned surface scope = %q, want documented no-version marker", surface.Scope)
	}
	for _, endpoint := range surface.Endpoints {
		if endpoint.Provenance == nil || endpoint.Provenance.Version != "provider-publishes-no-version-marker" {
			t.Fatalf("unversioned endpoint provenance = %+v, want explicit no-version marker", endpoint.Provenance)
		}
	}

	withoutEvidenceReport := filepath.Join(t.TempDir(), "without-evidence.json")
	stdout.Reset()
	stderr.Reset()
	code = run([]string{"batch", "materialize", "--manifest", manifestPath, "--source-defs-root", sourceDefsRoot, "--defs-root", t.TempDir(), "--artifact-dir", artifactDir, "--retrieved-at", "2026-08-09", "--report", withoutEvidenceReport}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("unversioned ordinary materialize exit = %d, want 1; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	decodeJSONFile(t, withoutEvidenceReport, &report)
	if len(report.Included) != 0 || len(report.Dropped) != 1 || report.Dropped[0].Stage != "existing_surface_evidence" || !strings.Contains(report.Dropped[0].Reason, "requires --existing-surface-evidence") {
		t.Fatalf("unversioned ordinary materialize report = %+v, want exact-evidence refusal", report)
	}

	setBatchManifestOperationCounts(t, manifestPath, 4, 2, 2)
	mismatchReportPath := filepath.Join(t.TempDir(), "mismatch.json")
	stdout.Reset()
	stderr.Reset()
	code = run([]string{"batch", "materialize", "--manifest", manifestPath, "--source-defs-root", sourceDefsRoot, "--defs-root", t.TempDir(), "--artifact-dir", artifactDir, "--retrieved-at", "2026-08-09", "--report", mismatchReportPath, "--existing-surface-evidence"}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("mismatched unversioned materialize exit = %d, want 1; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	decodeJSONFile(t, mismatchReportPath, &report)
	if len(report.Included) != 0 || len(report.Dropped) != 1 || report.Dropped[0].Stage != "existing_surface_evidence" || !strings.Contains(report.Dropped[0].Reason, "3 operation(s), not the immutable ledger count 4") {
		t.Fatalf("mismatched unversioned report = %+v, want exact-count refusal", report)
	}
}

func TestReadBatchManifestRejectsUndeclaredUnversionedArtifact(t *testing.T) {
	manifestPath := writeBatchManifestFixture(t, "cli-surface")
	rawManifest, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("read manifest fixture: %v", err)
	}
	var manifest map[string]any
	if err := json.Unmarshal(rawManifest, &manifest); err != nil {
		t.Fatalf("decode manifest fixture: %v", err)
	}
	connectors := manifest["connectors"].([]any)
	artifact := connectors[0].(map[string]any)["artifact"].(map[string]any)
	artifact["version"] = ""
	rawManifest, err = json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Fatalf("encode ordinary unversioned manifest fixture: %v", err)
	}
	if err := os.WriteFile(manifestPath, append(rawManifest, '\n'), 0o644); err != nil {
		t.Fatalf("write ordinary unversioned manifest fixture: %v", err)
	}

	if _, err := readBatchManifest(manifestPath); err == nil || !strings.Contains(err.Error(), "artifact.version is required") {
		t.Fatalf("read ordinary unversioned manifest error = %v, want preserved version guard", err)
	}
}

func TestExistingSurfaceEvidenceInventoryMergesSharedEndpointBindings(t *testing.T) {
	bundle := engine.Bundle{
		Name: "example",
		Surface: &engine.APISurface{
			API: "Example API",
			Endpoints: []engine.SurfaceEndpoint{
				{Method: http.MethodGet, Path: "/widgets", CoveredBy: &engine.SurfaceCoverage{Stream: "widgets"}},
				{Method: http.MethodGet, Path: "/widgets", CoveredBy: &engine.SurfaceCoverage{Stream: "archived_widgets"}},
				{Method: http.MethodGet, Path: "/config"},
				{Method: http.MethodGet, Path: "/healthz"},
			},
		},
		Streams: []engine.StreamSpec{{Name: "widgets"}, {Name: "archived_widgets"}},
	}
	candidate := BatchManifestConnector{
		Connector:       "example",
		OperationsTotal: 4,
		Artifact:        BatchArtifact{URL: "https://example.test/reference", Kind: "html_reference", Version: "v1"},
	}
	inventory, err := existingSurfaceEvidenceInventory(bundle, candidate, "2026-08-09", strings.Repeat("a", 64), []byte("official reference index"))
	if err != nil {
		t.Fatalf("existing surface evidence inventory: %v", err)
	}
	if len(inventory.Endpoints) != 3 {
		t.Fatalf("direct evidence endpoint inventory = %+v, want three unique provider method/path rows", inventory.Endpoints)
	}
	surface, err := materializeAPISurface(bundle, candidate, "2026-08-09", strings.Repeat("a", 64), inventory.Endpoints, inventory.Sources)
	if err != nil {
		t.Fatalf("materialize shared direct-evidence bindings: %v", err)
	}
	var widgets *engine.SurfaceEndpoint
	for index := range surface.Endpoints {
		endpoint := &surface.Endpoints[index]
		if endpoint.Method == http.MethodGet && endpoint.Path == "/widgets" {
			widgets = endpoint
			break
		}
	}
	if widgets == nil || widgets.CoveredBy == nil {
		t.Fatalf("surface = %+v, want one merged widgets coverage entry", surface.Endpoints)
	}
	got := map[string]bool{}
	for _, stream := range widgets.CoveredBy.StreamTargets() {
		got[stream] = true
	}
	if !got["widgets"] || !got["archived_widgets"] {
		t.Fatalf("merged direct-evidence streams = %+v, want both preserved stream targets", got)
	}
}

func TestBatchMaterializeMapsUnsupportedOperationsAndFlagsSurfaceDiscrepancies(t *testing.T) {
	sourceDefsRoot := t.TempDir()
	fsys := cliSurfaceBundleFS(validCLISurfaceJSON())
	fsys["cli-surface/api_surface.json"] = &fstest.MapFile{Data: []byte(`{
		"api": "test API v1",
		"endpoints": [
			{ "method": "GET", "path": "/widgets", "covered_by": { "stream": "widgets" } },
			{ "method": "POST", "path": "/widgets", "covered_by": { "write": "create_widget" } },
			{ "method": "GET", "path": "/legacy", "covered_by": { "stream": "widgets" } }
		]
	}`)}
	writeBatchBundle(t, sourceDefsRoot, fsys)
	defsRoot := t.TempDir()
	manifestPath := writeBatchManifestFixture(t, "cli-surface")
	setBatchManifestOperationCounts(t, manifestPath, 3, 1, 2)
	artifactDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(artifactDir, "cli-surface.json"), []byte(`{
		"openapi": "3.0.0",
		"paths": {
			"/widgets": {
				"get": {"summary": "List widgets"},
				"post": {"summary": "Create widget"}
			},
			"/new": {
				"delete": {"summary": "Delete a widget"}
			}
		}
	}`), 0o644); err != nil {
		t.Fatalf("write artifact: %v", err)
	}
	reportPath := filepath.Join(t.TempDir(), "materialize.json")
	var stdout, stderr bytes.Buffer

	code := run([]string{"batch", "materialize", "--manifest", manifestPath, "--source-defs-root", sourceDefsRoot, "--defs-root", defsRoot, "--artifact-dir", artifactDir, "--retrieved-at", "2026-08-06", "--report", reportPath}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("batch materialize exit = %d, want complete inventory materialization; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}

	bundleDir := filepath.Join(defsRoot, "cli-surface")
	var surface engine.APISurface
	decodeJSONFile(t, filepath.Join(bundleDir, "api_surface.json"), &surface)
	if len(surface.Endpoints) != 4 {
		t.Fatalf("materialized endpoints = %d, want three artifact operations plus one source discrepancy", len(surface.Endpoints))
	}
	var discrepancy, unsupported *engine.SurfaceEndpoint
	for i := range surface.Endpoints {
		endpoint := &surface.Endpoints[i]
		if endpoint.Discrepancy == "present-in-surface-absent-from-artifact" {
			discrepancy = endpoint
		}
		if endpoint.Method == http.MethodDelete && endpoint.Path == "/new" {
			unsupported = endpoint
		}
	}
	if discrepancy == nil || discrepancy.Method != http.MethodGet || discrepancy.Path != "/legacy" || discrepancy.CoveredBy == nil {
		t.Fatalf("source-only endpoint = %+v, want retained executable endpoint with exact discrepancy", discrepancy)
	}
	if unsupported == nil || unsupported.Operation == nil || !strings.HasPrefix(unsupported.Operation.Notes, "named_dependency=") {
		t.Fatalf("unsupported endpoint = %+v, want blocked operation with named dependency", unsupported)
	}

	var operations struct {
		Operations []engine.OperationSpec `json:"operations"`
	}
	decodeJSONFile(t, filepath.Join(bundleDir, "operations.json"), &operations)
	if len(operations.Operations) != 1 || operations.Operations[0].REST == nil {
		t.Fatalf("generated operations = %+v, want one typed metadata operation", operations.Operations)
	}

	var cli engine.CLISurface
	decodeJSONFile(t, filepath.Join(bundleDir, "cli_surface.json"), &cli)
	if len(cli.Commands) != 3 {
		t.Fatalf("generated commands = %d, want stream, write, and not-implemented artifact operation", len(cli.Commands))
	}
	var notImplemented int
	for _, command := range cli.Commands {
		if command.Availability == "not_implemented" {
			notImplemented++
			if !strings.HasPrefix(command.Notes, "named_dependency=") {
				t.Fatalf("not-implemented command = %+v, want named dependency note", command)
			}
		}
	}
	if notImplemented != 1 {
		t.Fatalf("not-implemented commands = %d, want one", notImplemented)
	}
	validation, err := validatePath(bundleDir)
	if err != nil {
		t.Fatalf("validate materialized bundle: %v", err)
	}
	if len(validation.Findings) != 0 {
		t.Fatalf("materialized bundle findings = %+v, want zero", validation.Findings)
	}
}

func TestBatchMaterializeRetainsMissingExecutableCoverageAsDiscrepancy(t *testing.T) {
	sourceDefsRoot := t.TempDir()
	writeBatchBundle(t, sourceDefsRoot, cliSurfaceBundleFS(validCLISurfaceJSON()))
	defsRoot := t.TempDir()
	manifestPath := writeBatchManifestFixture(t, "cli-surface")
	setBatchManifestOperationCounts(t, manifestPath, 1, 1, 0)
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
	if code != 0 {
		t.Fatalf("batch materialize exit = %d, want source coverage retained as discrepancy; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	var surface engine.APISurface
	decodeJSONFile(t, filepath.Join(defsRoot, "cli-surface", "api_surface.json"), &surface)
	if len(surface.Endpoints) != 3 {
		t.Fatalf("materialized endpoints = %d, want artifact endpoint plus two retained source rows", len(surface.Endpoints))
	}
	var discrepancies int
	for _, endpoint := range surface.Endpoints {
		if endpoint.Discrepancy == materializeSurfaceDiscrepancy {
			discrepancies++
		}
	}
	if discrepancies != 2 {
		t.Fatalf("materialized surface = %+v, want two source-only discrepancy rows", surface.Endpoints)
	}
	if _, err := os.Stat(filepath.Join(defsRoot, "cli-surface", "operations.json")); err != nil {
		t.Fatalf("materializer did not write generated operations catalog: %v", err)
	}
}

func TestBatchMaterializePreservesImplementedDirectReadCoverage(t *testing.T) {
	sourceDefsRoot := t.TempDir()
	writeBatchBundle(t, sourceDefsRoot, directReadCLISurfaceBundleFS(validDirectReadCLISurfaceJSON()))
	defsRoot := t.TempDir()
	manifestPath := writeBatchManifestFixture(t, "cli-surface")
	setBatchManifestOperationCounts(t, manifestPath, 4, 2, 2)
	artifactDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(artifactDir, "cli-surface.json"), []byte(`{
		"openapi": "3.0.0",
		"paths": {
			"/widgets": { "get": {"summary": "List widgets"}, "post": {"summary": "Create widget"} },
			"/widgets/{path}": { "get": {"summary": "Read widget metadata"} },
			"/widgets/{id}": { "delete": {"summary": "Delete widget"} }
		}
	}`), 0o644); err != nil {
		t.Fatalf("write artifact: %v", err)
	}
	reportPath := filepath.Join(t.TempDir(), "materialize.json")
	var stdout, stderr bytes.Buffer

	if code := run([]string{"batch", "materialize", "--manifest", manifestPath, "--source-defs-root", sourceDefsRoot, "--defs-root", defsRoot, "--artifact-dir", artifactDir, "--retrieved-at", "2026-08-06", "--report", reportPath}, &stdout, &stderr); code != 0 {
		t.Fatalf("batch materialize exit = %d, want 0; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}

	var cli engine.CLISurface
	bundleDir := filepath.Join(defsRoot, "cli-surface")
	decodeJSONFile(t, filepath.Join(bundleDir, "cli_surface.json"), &cli)
	var preserved bool
	for _, command := range cli.Commands {
		if command.Path == "widget read" && command.Intent == "direct_read" && command.Availability == "implemented" {
			preserved = true
		}
	}
	if !preserved {
		t.Fatalf("materialized commands = %+v, want implemented direct-read command retained for its covered endpoint", cli.Commands)
	}
	validation, err := validatePath(bundleDir)
	if err != nil {
		t.Fatalf("validate materialized direct-read bundle: %v", err)
	}
	if len(validation.Findings) != 0 {
		t.Fatalf("materialized direct-read findings = %+v, want zero", validation.Findings)
	}
}

func TestBatchMaterializeRetainsSourceCoverageWhenArtifactIdentityDiffers(t *testing.T) {
	tests := []struct {
		name            string
		sourceMethod    string
		sourcePath      string
		artifactMethod  string
		artifactPath    string
		wantDiscrepancy bool
	}{
		{
			name:           "exact method and path",
			sourceMethod:   http.MethodGet,
			sourcePath:     "/widgets",
			artifactMethod: http.MethodGet,
			artifactPath:   "/widgets",
		},
		{
			name:            "trailing slash is distinct",
			sourceMethod:    http.MethodGet,
			sourcePath:      "/widgets",
			artifactMethod:  http.MethodGet,
			artifactPath:    "/widgets/",
			wantDiscrepancy: true,
		},
		{
			name:           "method case is equivalent",
			sourceMethod:   "get",
			sourcePath:     "/widgets",
			artifactMethod: http.MethodGet,
			artifactPath:   "/widgets",
		},
		{
			name:           "artifact method case is equivalent",
			sourceMethod:   http.MethodGet,
			sourcePath:     "/widgets",
			artifactMethod: "get",
			artifactPath:   "/widgets",
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
			if err != nil {
				t.Fatalf("materialize exact endpoint: %v", err)
			}
			if test.wantDiscrepancy {
				if len(surface.Endpoints) != 2 {
					t.Fatalf("materialized endpoints = %+v, want artifact operation plus retained source discrepancy", surface.Endpoints)
				}
				var found bool
				for _, endpoint := range surface.Endpoints {
					if endpoint.Method == test.sourceMethod && endpoint.Path == test.sourcePath && endpoint.Discrepancy == materializeSurfaceDiscrepancy {
						found = true
					}
				}
				if !found {
					t.Fatalf("materialized endpoints = %+v, want exact source discrepancy", surface.Endpoints)
				}
				return
			}
			if len(surface.Endpoints) != 1 || surface.Endpoints[0].CoveredBy == nil || surface.Endpoints[0].CoveredBy.Stream != "widgets" {
				t.Fatalf("materialized endpoints = %+v, want exact cited stream coverage", surface.Endpoints)
			}
		})
	}
}

func TestMaterializeAPISurfacePreservesMultiSourceProvenanceAlternatives(t *testing.T) {
	bundle := engine.Bundle{
		Surface: &engine.APISurface{
			API: "Example API",
			Endpoints: []engine.SurfaceEndpoint{{
				Method:    http.MethodGet,
				Path:      "/widgets",
				CoveredBy: &engine.SurfaceCoverage{Stream: "widgets"},
			}},
		},
		Streams: []engine.StreamSpec{{Name: "widgets"}},
	}
	candidate := BatchManifestConnector{Connector: "cli-surface", Artifact: BatchArtifact{
		Kind:    "openapi",
		URL:     "https://provider.example/openapi.json",
		Version: "3.1.0",
	}}
	rootHash := strings.Repeat("a", 64)
	refHash := strings.Repeat("b", 64)
	surface, err := materializeAPISurface(
		bundle,
		candidate,
		"2026-08-08",
		rootHash,
		[]batchArtifactEndpoint{{
			Method:           http.MethodGet,
			Path:             "/widgets",
			SourceURL:        "https://provider.example/paths/widgets.yaml",
			SourceKind:       "referenced-document",
			SourceVersion:    "3.1.0",
			SourceRetrieved:  "2026-08-08",
			SourceSHA256:     refHash,
			SourceCoordinate: "paths[\"/widgets\"].get",
			Alternatives: []batchArtifactEndpointAlternative{{
				SourceURL:        "https://provider.example/reference",
				SourceKind:       "official-reference",
				SourceVersion:    "2026-08-08",
				SourceRetrieved:  "2026-08-08",
				SourceSHA256:     rootHash,
				SourceCoordinate: "https://provider.example/reference#GET /widgets",
			}},
		}},
		[]batchArtifactSource{{
			URL:       "https://provider.example/openapi.json",
			Kind:      "openapi",
			Version:   "3.1.0",
			Retrieved: "2026-08-08",
			SHA256:    rootHash,
		}, {
			URL:       "https://provider.example/paths/widgets.yaml",
			Kind:      "referenced-document",
			Version:   "3.1.0",
			Retrieved: "2026-08-08",
			SHA256:    refHash,
		}, {
			URL:       "https://provider.example/reference",
			Kind:      "official-reference",
			Version:   "2026-08-08",
			Retrieved: "2026-08-08",
			SHA256:    rootHash,
		}},
	)
	if err != nil {
		t.Fatalf("materialize multi-source surface: %v", err)
	}
	if len(surface.Artifacts) != 3 || len(surface.Endpoints) != 1 {
		t.Fatalf("surface = %+v, want three cited sources and one canonical endpoint", surface)
	}
	provenance := surface.Endpoints[0].Provenance
	if provenance == nil || provenance.SourceURL != "https://provider.example/paths/widgets.yaml" || provenance.SourceKind != "referenced-document" || provenance.SHA256 != refHash || provenance.Coordinate == "" {
		t.Fatalf("primary provenance = %+v, want normalized referenced-document citation", provenance)
	}
	if len(provenance.Alternatives) != 1 || provenance.Alternatives[0].SourceURL != "https://provider.example/reference" || provenance.Alternatives[0].Coordinate == "" {
		t.Fatalf("alternative provenance = %+v, want preserved official-reference disagreement", provenance.Alternatives)
	}
	catalogSurface := engine.APISurface{Endpoints: []engine.SurfaceEndpoint{{
		Method:     http.MethodGet,
		Path:       "/widgets",
		Provenance: provenance,
		Operation:  &engine.SurfaceOperation{Reason: "documented read", Risk: "low"},
	}}}
	operations, err := materializeOperationCatalog(bundle, catalogSurface, candidate)
	if err != nil {
		t.Fatalf("materialize operation catalog: %v", err)
	}
	if len(operations) != 1 || operations[0].SourceURL != "https://provider.example/paths/widgets.yaml" {
		t.Fatalf("operation provenance = %+v, want referenced-document source URL", operations)
	}
}

func TestMaterializeAPISurfaceDisambiguatesIdenticalReferenceArtifacts(t *testing.T) {
	bundle := engine.Bundle{Surface: &engine.APISurface{API: "Example API"}}
	candidate := BatchManifestConnector{Connector: "cli-surface", Artifact: BatchArtifact{
		Kind:    "openapi",
		URL:     "https://provider.example/openapi.json",
		Version: "3.1.0",
	}}
	rootHash := strings.Repeat("a", 64)
	sharedReferenceHash := strings.Repeat("b", 64)
	firstReferenceURL := "https://provider.example/references/first.json"
	secondReferenceURL := "https://provider.example/references/second.json"
	surface, err := materializeAPISurface(
		bundle,
		candidate,
		"2026-08-08",
		rootHash,
		[]batchArtifactEndpoint{{
			Method:           http.MethodGet,
			Path:             "/widgets",
			SourceURL:        firstReferenceURL,
			SourceKind:       "official-reference",
			SourceVersion:    "3.1.0",
			SourceRetrieved:  "2026-08-08",
			SourceSHA256:     sharedReferenceHash,
			SourceCoordinate: "https://provider.example/references/first.json#/paths/~1widgets/get",
		}, {
			Method:           http.MethodPost,
			Path:             "/widgets",
			SourceURL:        secondReferenceURL,
			SourceKind:       "official-reference",
			SourceVersion:    "3.1.0",
			SourceRetrieved:  "2026-08-08",
			SourceSHA256:     sharedReferenceHash,
			SourceCoordinate: "https://provider.example/references/second.json#/paths/~1widgets/post",
		}},
		[]batchArtifactSource{{
			URL:       candidate.Artifact.URL,
			Kind:      candidate.Artifact.Kind,
			Version:   candidate.Artifact.Version,
			Retrieved: "2026-08-08",
			SHA256:    rootHash,
		}, {
			URL:       firstReferenceURL,
			Kind:      "official-reference",
			Version:   "3.1.0",
			Retrieved: "2026-08-08",
			SHA256:    sharedReferenceHash,
		}, {
			URL:       secondReferenceURL,
			Kind:      "official-reference",
			Version:   "3.1.0",
			Retrieved: "2026-08-08",
			SHA256:    sharedReferenceHash,
		}},
	)
	if err != nil {
		t.Fatalf("materialize duplicate-content references: %v", err)
	}
	artifactIDs := map[string]string{}
	for _, artifact := range surface.Artifacts {
		artifactIDs[artifact.URL] = artifact.ID
	}
	if artifactIDs[firstReferenceURL] == "" || artifactIDs[secondReferenceURL] == "" || artifactIDs[firstReferenceURL] == artifactIDs[secondReferenceURL] {
		t.Fatalf("reference artifact IDs = %+v, want distinct IDs for distinct URLs", artifactIDs)
	}
	if provenance := engine.ValidateSurfaceProvenance(&surface); provenance.Status != engine.SurfaceProvenanceComplete || len(provenance.Issues) != 0 {
		t.Fatalf("surface provenance = %+v, want complete distinct-reference evidence", provenance)
	}
}

func TestMaterializeAPISurfaceMatchesMethodsCaseInsensitively(t *testing.T) {
	bundle := engine.Bundle{
		Surface: &engine.APISurface{
			API: "Example API",
			Endpoints: []engine.SurfaceEndpoint{{
				Method:    " get ",
				Path:      "/widgets",
				CoveredBy: &engine.SurfaceCoverage{Stream: "widgets"},
			}},
		},
		Streams: []engine.StreamSpec{{Name: "widgets"}},
	}
	candidate := BatchManifestConnector{Connector: "cli-surface", Artifact: BatchArtifact{
		Kind:    "openapi",
		URL:     "https://provider.example/openapi.json",
		Version: "3.1.0",
	}}
	surface, err := materializeAPISurface(bundle, candidate, "2026-08-08", strings.Repeat("a", 64), []batchArtifactEndpoint{{
		Method: http.MethodGet,
		Path:   "/widgets",
	}})
	if err != nil {
		t.Fatalf("materialize case-variant endpoint: %v", err)
	}
	if len(surface.Endpoints) != 1 {
		t.Fatalf("materialized endpoints = %+v, want one reconciled endpoint", surface.Endpoints)
	}
	endpoint := surface.Endpoints[0]
	if endpoint.Method != http.MethodGet || endpoint.CoveredBy == nil || endpoint.CoveredBy.Stream != "widgets" || endpoint.Discrepancy != "" {
		t.Fatalf("reconciled endpoint = %+v, want artifact method with existing classifier and no discrepancy", endpoint)
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

func TestParseBatchSwaggerArtifactPrefixesBasePath(t *testing.T) {
	artifact := []byte(`{
		"swagger": "2.0",
		"basePath": "/api",
		"paths": {
			"/widgets": {
				"get": {"summary": "List widgets", "responses": {"200": {"description": "OK"}}}
			}
		}
	}`)

	endpoints, err := parseBatchOpenAPIArtifact(artifact)
	if err != nil {
		t.Fatalf("parse Swagger basePath artifact: %v", err)
	}
	if len(endpoints) != 1 || endpoints[0].Method != http.MethodGet || endpoints[0].Path != "/api/widgets" {
		t.Fatalf("parsed endpoints = %+v, want GET /api/widgets", endpoints)
	}
}

func TestParseBatchOpenAPIArtifactPrefixesServerBasePath(t *testing.T) {
	artifact := []byte(`{
		"openapi": "3.1.0",
		"servers": [{"url": "https://{account}.provider.example/api/3"}],
		"paths": {
			"/widgets": {
				"get": {"summary": "List widgets", "responses": {"200": {"description": "OK"}}}
			}
		}
	}`)

	endpoints, err := parseBatchOpenAPIArtifact(artifact)
	if err != nil {
		t.Fatalf("parse OpenAPI server base-path artifact: %v", err)
	}
	if len(endpoints) != 1 || endpoints[0].Method != http.MethodGet || endpoints[0].Path != "/api/3/widgets" {
		t.Fatalf("parsed endpoints = %+v, want GET /api/3/widgets", endpoints)
	}
}

func TestParseBatchOpenAPIArtifactPrefixesServerBasePathWithPathVariable(t *testing.T) {
	artifact := []byte(`{
		"openapi": "3.1.0",
		"servers": [{"url": "https://api.provider.example/api/v100/rest/spaces/{space_id}"}],
		"paths": {
			"/widgets": {
				"get": {"summary": "List widgets", "responses": {"200": {"description": "OK"}}}
			}
		}
	}`)

	endpoints, err := parseBatchOpenAPIArtifact(artifact)
	if err != nil {
		t.Fatalf("parse OpenAPI templated server base-path artifact: %v", err)
	}
	if len(endpoints) != 1 || endpoints[0].Method != http.MethodGet || endpoints[0].Path != "/api/v100/rest/spaces/{space_id}/widgets" {
		t.Fatalf("parsed endpoints = %+v, want GET /api/v100/rest/spaces/{space_id}/widgets", endpoints)
	}
}

func TestParseBatchOpenAPIArtifactRequiresOpenAPI3OrSwagger2(t *testing.T) {
	for _, test := range []struct {
		name string
		doc  string
	}{
		{name: "openapi2", doc: `openapi: 2.0.0
paths:
  /widgets:
    get: {}`},
		{name: "swagger3", doc: `swagger: "3.0"
paths:
  /widgets:
    get: {}`},
		{name: "both", doc: `openapi: 3.0.3
swagger: "2.0"
paths:
  /widgets:
    get: {}`},
	} {
		t.Run(test.name, func(t *testing.T) {
			if _, err := parseBatchOpenAPIArtifact([]byte(test.doc)); err == nil || !strings.Contains(err.Error(), "OpenAPI 3.x or Swagger 2.0") {
				t.Fatalf("parse error = %v, want strict version rejection", err)
			}
		})
	}
}

func TestParseBatchGoogleDiscoveryArtifactEnumeratesNestedMethods(t *testing.T) {
	artifact := []byte(`{
		"kind": "discovery#restDescription",
		"name": "example",
		"version": "v1",
		"methods": {
			"health": {"id": "example.health", "path": "v1/health", "httpMethod": "GET", "description": "Check health"}
		},
		"resources": {
			"widgets": {
				"methods": {
					"get": {"id": "example.widgets.get", "path": "v1/{+name}", "httpMethod": "GET", "description": "Get widget"}
				},
				"resources": {
					"items": {
						"methods": {
							"create": {"id": "example.widgets.items.create", "path": "v1/{+parent}/items", "httpMethod": "POST", "description": "Create item"}
						}
					}
				}
			}
		}
	}`)
	source := batchArtifactSource{URL: "https://example.test/$discovery/rest?version=v1", Kind: "openapi", Version: "v1", Retrieved: "2026-08-09"}
	inventory, err := parseBatchGoogleDiscoveryArtifact(artifact, source)
	if err != nil {
		t.Fatalf("parse Google Discovery artifact: %v", err)
	}
	if len(inventory.Endpoints) != 3 {
		t.Fatalf("endpoint count = %d, want 3", len(inventory.Endpoints))
	}
	got := make(map[string]batchArtifactEndpoint, len(inventory.Endpoints))
	for _, endpoint := range inventory.Endpoints {
		got[endpoint.Method+" "+endpoint.Path] = endpoint
	}
	for _, key := range []string{"GET /v1/health", "GET /v1/{name}", "POST /v1/{parent}/items"} {
		endpoint, ok := got[key]
		if !ok {
			t.Fatalf("endpoints = %+v, missing %s", inventory.Endpoints, key)
		}
		if endpoint.SourceURL != source.URL || endpoint.SourceKind != source.Kind || endpoint.SourceVersion != source.Version || endpoint.SourceRetrieved != source.Retrieved || endpoint.SourceCoordinate == "" {
			t.Fatalf("endpoint provenance = %+v, want exact discovery source", endpoint)
		}
	}
}

func TestMaterializeAPISurfaceRetainsDuplicateCoveredStreamBindings(t *testing.T) {
	bundle := engine.Bundle{
		Name: "example",
		Surface: &engine.APISurface{
			API: "Example API",
			Endpoints: []engine.SurfaceEndpoint{
				{Method: http.MethodGet, Path: "/widgets", CoveredBy: &engine.SurfaceCoverage{Stream: "widgets"}},
				{Method: http.MethodGet, Path: "/widgets", CoveredBy: &engine.SurfaceCoverage{Stream: "archived_widgets"}},
			},
		},
		Streams: []engine.StreamSpec{{Name: "widgets"}, {Name: "archived_widgets"}},
	}
	candidate := BatchManifestConnector{Connector: "example", Artifact: BatchArtifact{URL: "https://example.test/discovery", Kind: "openapi", Version: "v1"}}
	surface, err := materializeAPISurface(bundle, candidate, "2026-08-09", strings.Repeat("a", 64), []batchArtifactEndpoint{{
		Method:           http.MethodGet,
		Path:             "/v1/widgets",
		Summary:          "List widgets",
		SourceURL:        candidate.Artifact.URL,
		SourceKind:       candidate.Artifact.Kind,
		SourceVersion:    candidate.Artifact.Version,
		SourceRetrieved:  "2026-08-09",
		SourceSHA256:     strings.Repeat("a", 64),
		SourceCoordinate: "methods[\"list\"]",
	}})
	if err != nil {
		t.Fatalf("materialize duplicate stream bindings: %v", err)
	}
	if len(surface.Endpoints) != 1 || surface.Endpoints[0].CoveredBy == nil {
		t.Fatalf("surface endpoints = %+v, want one merged binding", surface.Endpoints)
	}
	if surface.Endpoints[0].Path != "/widgets" {
		t.Fatalf("materialized endpoint path = %q, want connector-relative /widgets", surface.Endpoints[0].Path)
	}
	got := map[string]bool{}
	for _, stream := range surface.Endpoints[0].CoveredBy.StreamTargets() {
		got[stream] = true
	}
	if !got["widgets"] || !got["archived_widgets"] {
		t.Fatalf("surface stream bindings = %+v, want both source streams", got)
	}
}

func TestMaterializeAPISurfaceUsesExistingDynamicServerBaseSuffix(t *testing.T) {
	bundle := engine.Bundle{
		Name: "example",
		Surface: &engine.APISurface{
			API: "Example API",
			Endpoints: []engine.SurfaceEndpoint{{
				Method:    http.MethodGet,
				Path:      "/spaces/{space_id}/widgets",
				CoveredBy: &engine.SurfaceCoverage{Stream: "widgets"},
			}},
		},
		Streams: []engine.StreamSpec{{Name: "widgets"}},
	}
	candidate := BatchManifestConnector{Connector: "example", Artifact: BatchArtifact{URL: "https://example.test/openapi.json", Kind: "openapi", Version: "v1"}}
	surface, err := materializeAPISurface(bundle, candidate, "2026-08-09", strings.Repeat("a", 64), []batchArtifactEndpoint{{
		Method:           http.MethodGet,
		Path:             "/api/v100/rest/spaces/{space_id}/widgets",
		Summary:          "List widgets",
		SourceURL:        candidate.Artifact.URL,
		SourceKind:       candidate.Artifact.Kind,
		SourceVersion:    candidate.Artifact.Version,
		SourceRetrieved:  "2026-08-09",
		SourceSHA256:     strings.Repeat("a", 64),
		SourceCoordinate: "paths[\"/widgets\"].get",
	}})
	if err != nil {
		t.Fatalf("materialize dynamic server base suffix: %v", err)
	}
	if len(surface.Endpoints) != 1 || surface.Endpoints[0].Path != "/spaces/{space_id}/widgets" || surface.Endpoints[0].CoveredBy == nil || surface.Endpoints[0].CoveredBy.Stream != "widgets" || surface.Endpoints[0].Discrepancy != "" {
		t.Fatalf("surface endpoints = %+v, want one classified connector-relative dynamic-base endpoint", surface.Endpoints)
	}
}

func TestMaterializeOperationCatalogKeepsEmptyOperationsArray(t *testing.T) {
	operations, err := materializeOperationCatalog(engine.Bundle{Name: "example"}, engine.APISurface{}, BatchManifestConnector{Connector: "example"})
	if err != nil {
		t.Fatalf("materialize empty operation catalog: %v", err)
	}
	if operations == nil {
		t.Fatal("materialized empty operation catalog is nil, want JSON []")
	}
}

func TestMaterializeOperationCatalogUsesArtifactURLWithoutEndpointProvenance(t *testing.T) {
	operations, err := materializeOperationCatalog(
		engine.Bundle{Name: "example"},
		engine.APISurface{Endpoints: []engine.SurfaceEndpoint{{
			Method:    http.MethodGet,
			Path:      "/widgets",
			Operation: &engine.SurfaceOperation{Model: "direct_read", Risk: "low", Reason: "documented read"},
		}}},
		BatchManifestConnector{Connector: "example", Artifact: BatchArtifact{URL: "https://provider.example/openapi.json"}},
	)
	if err != nil {
		t.Fatalf("materialize operation catalog: %v", err)
	}
	if len(operations) != 1 || operations[0].SourceURL != "https://provider.example/openapi.json" {
		t.Fatalf("operations = %+v, want artifact source URL fallback", operations)
	}
}

func TestMaterializeOperationsFollowEndpointModels(t *testing.T) {
	provenance := &engine.SurfaceProvenance{SourceURL: "https://provider.example/openapi.json"}
	surface := engine.APISurface{Endpoints: []engine.SurfaceEndpoint{
		{
			Method:     http.MethodPost,
			Path:       "/flows/batch/read",
			Provenance: provenance,
			Operation:  &engine.SurfaceOperation{Model: "direct_read", Risk: "medium", Reason: "bounded query"},
		},
		{
			Method:     http.MethodGet,
			Path:       "/files/diff",
			Provenance: provenance,
			Operation:  &engine.SurfaceOperation{Model: "binary_read", Risk: "medium", Reason: "binary diff"},
		},
		{
			Method:     http.MethodPost,
			Path:       "/reports/search",
			Provenance: provenance,
			Operation:  &engine.SurfaceOperation{Model: "direct_read", Risk: "medium", Reason: "untyped query"},
		},
		{
			Method:     http.MethodGet,
			Path:       "/canonical/read",
			Provenance: provenance,
			Operation:  &engine.SurfaceOperation{Model: "direct_read", Risk: "low", Reason: "canonical typed read"},
		},
	}}
	bundle := engine.Bundle{
		Name: "acme",
		Operations: []engine.OperationSpec{
			{
				ID:           "acme.flow.read",
				Kind:         "rest_read",
				Summary:      "Read flows",
				Risk:         "medium",
				Approval:     "none",
				OutputPolicy: "json_redacted",
				REST: &engine.RESTOperationSpec{
					Method:     http.MethodPost,
					Path:       "/flows/batch/read",
					MaxBytes:   1024,
					BodySchema: json.RawMessage(`{"type":"object"}`),
				},
			},
			{
				ID:           "acme.diff.download",
				Kind:         "binary_download",
				Summary:      "Download diff",
				Risk:         "medium",
				Approval:     "none",
				OutputPolicy: "binary_file_bounded",
				Binary:       &engine.BinaryOperationSpec{Method: http.MethodGet, Path: "/files/diff", MaxBytes: 1024},
			},
			{
				ID:            "acme.post.flows-batch-read",
				Kind:          "rest_write",
				Summary:       "Stale generated write",
				Risk:          "medium",
				Approval:      "not implemented",
				OutputPolicy:  "json",
				MutationClass: "create",
				REST:          &engine.RESTOperationSpec{Method: http.MethodPost, Path: "/flows/batch/read", MaxBytes: 1024},
			},
			{
				ID:           "acme.get.files-diff",
				Kind:         "rest_read",
				Summary:      "Stale generated read",
				Risk:         "medium",
				Approval:     "not implemented",
				OutputPolicy: "json",
				REST:         &engine.RESTOperationSpec{Method: http.MethodGet, Path: "/files/diff", MaxBytes: 1024},
			},
			{
				ID:           "acme.get.canonical-read",
				Kind:         "rest_read",
				Summary:      "Canonical read",
				Risk:         "low",
				Approval:     "none",
				OutputPolicy: "json_redacted",
				REST:         &engine.RESTOperationSpec{Method: http.MethodGet, Path: "/canonical/read", MaxBytes: 2048},
			},
		},
	}
	candidate := BatchManifestConnector{Connector: "acme", Artifact: BatchArtifact{URL: "https://provider.example/openapi.json"}}

	operations, err := materializeOperationCatalog(bundle, surface, candidate)
	if err != nil {
		t.Fatalf("materializeOperationCatalog: %v", err)
	}
	operationsByID := make(map[string]engine.OperationSpec, len(operations))
	for _, operation := range operations {
		operationsByID[operation.ID] = operation
	}
	if _, ok := operationsByID["acme.post.flows-batch-read"]; ok {
		t.Fatal("stale generated rest_write operation was retained for direct_read endpoint")
	}
	if _, ok := operationsByID["acme.get.files-diff"]; ok {
		t.Fatal("stale generated rest_read operation was retained for binary_read endpoint")
	}
	if operation, ok := operationsByID["acme.post.reports-search"]; !ok || operation.Kind != "composite" {
		t.Fatalf("untyped POST direct-read operation = %+v, want composite metadata", operation)
	}
	if operation, ok := operationsByID["acme.get.canonical-read"]; !ok || operation.REST == nil || operation.REST.MaxBytes != 2048 {
		t.Fatalf("canonical typed read = %+v, want retained source contract", operation)
	}

	cli, err := materializeCLISurface(bundle, surface, candidate, operations)
	if err != nil {
		t.Fatalf("materializeCLISurface: %v", err)
	}
	commands := make(map[string]engine.CLICommand, len(cli.Commands))
	for _, command := range cli.Commands {
		commands[command.Path] = command
	}
	for path, want := range map[string]struct {
		intent     string
		operation  string
		dependency string
	}{
		"api post flows batch read": {intent: "direct_read", operation: "acme.flow.read", dependency: "engine.direct_read_operation_contract"},
		"api get files diff":        {intent: "binary_download", operation: "acme.diff.download", dependency: "engine.binary_download_operation_contract"},
		"api post reports search":   {intent: "direct_read", operation: "acme.post.reports-search", dependency: "engine.direct_read_operation_contract"},
		"api get canonical read":    {intent: "direct_read", operation: "acme.get.canonical-read", dependency: "engine.direct_read_executor"},
	} {
		command, ok := commands[path]
		if !ok {
			t.Fatalf("command %q was not materialized", path)
		}
		if command.Intent != want.intent || command.Operation != want.operation || !strings.Contains(command.Notes, want.dependency) {
			t.Fatalf("command %q = %+v, want intent=%q operation=%q dependency=%q", path, command, want.intent, want.operation, want.dependency)
		}
	}
}

func TestMaterializeCLISurfaceUsesPluralWriteCoverage(t *testing.T) {
	bundle := engine.Bundle{
		Name: "acme",
		Writes: []engine.WriteAction{
			{Name: "set_banner", RecordSchema: json.RawMessage(`{"type":"object","properties":{}}`)},
			{Name: "clear_banner", RecordSchema: json.RawMessage(`{"type":"object","properties":{}}`)},
		},
	}
	surface := engine.APISurface{Endpoints: []engine.SurfaceEndpoint{{
		Method:    http.MethodPut,
		Path:      "/announcement-banner",
		CoveredBy: &engine.SurfaceCoverage{Writes: []string{"set_banner", "clear_banner"}},
	}}}
	if err := ensureMaterializedCoverage(bundle, surface); err != nil {
		t.Fatalf("ensureMaterializedCoverage: %v", err)
	}

	cli, err := materializeCLISurface(bundle, surface, BatchManifestConnector{Connector: "acme", Artifact: BatchArtifact{URL: "https://provider.example/openapi.json"}}, nil)
	if err != nil {
		t.Fatalf("materializeCLISurface: %v", err)
	}
	for _, write := range []string{"set_banner", "clear_banner"} {
		path := materializedCommandPath(write, "apply")
		found := false
		for _, command := range cli.Commands {
			if command.Path == path && command.Write == write && len(command.APISurface) == 1 && command.APISurface[0] == (engine.CLISurfaceEndpointRef{Method: http.MethodPut, Path: "/announcement-banner"}) {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("write command %q did not retain plural coverage reference: %+v", write, cli.Commands)
		}
	}
}

func TestMaterializeCLISurfacePreservesExistingWriteContractAndPropagatesRedactions(t *testing.T) {
	bundle := engine.Bundle{
		Name: "acme",
		Writes: []engine.WriteAction{{
			Name: "change_visibility",
			RecordSchema: json.RawMessage(`{
				"type": "object",
				"required": ["receipt_handle", "visibility_timeout"],
				"properties": {
					"receipt_handle": {"type": "string"},
					"visibility_timeout": {"type": "integer"}
				}
			}`),
			RedactFields: []string{"action_secret", "receipt_handle"},
		}},
		CLISurface: &engine.CLISurface{Commands: []engine.CLICommand{{
			Path:         "message change-visibility",
			Intent:       "reverse_etl",
			Availability: "implemented",
			Write:        "change_visibility",
			Flags: []engine.CLIFlag{
				{Name: "receipt-handle", Type: "string", MapsTo: "record.receipt_handle", Required: true},
				{Name: "visibility-timeout", Type: "integer", MapsTo: "record.visibility_timeout", Required: true},
				{Name: "legacy-only", Type: "string", MapsTo: "record.legacy_only"},
			},
			RedactFields: []string{"existing_secret", "receipt_handle"},
			Examples:     []string{"pm acme message change-visibility --receipt-handle receipt --visibility-timeout 30 --legacy-only value --preview"},
		}}},
	}
	surface := engine.APISurface{Endpoints: []engine.SurfaceEndpoint{{
		Method:    http.MethodPost,
		Path:      "/messages/visibility",
		CoveredBy: &engine.SurfaceCoverage{Write: "change_visibility"},
	}}}

	cli, err := materializeCLISurface(bundle, surface, BatchManifestConnector{Connector: "acme", Artifact: BatchArtifact{URL: "https://provider.example/openapi.json"}}, nil)
	if err != nil {
		t.Fatalf("materializeCLISurface: %v", err)
	}
	if len(cli.Commands) != 1 {
		t.Fatalf("materialized commands = %+v, want one write command", cli.Commands)
	}
	command := cli.Commands[0]
	if command.Path != "message change-visibility" {
		t.Fatalf("command path = %q, want retained existing path", command.Path)
	}
	if got, want := command.RedactFields, []string{"action_secret", "existing_secret", "receipt_handle"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("command redact_fields = %v, want %v", got, want)
	}
	if got, want := command.Flags, []engine.CLIFlag{
		{Name: "receipt-handle", Type: "string", MapsTo: "record.receipt_handle", Required: true},
		{Name: "visibility-timeout", Type: "integer", MapsTo: "record.visibility_timeout", Required: true},
		{Name: "legacy-only", Type: "string", MapsTo: "record.legacy_only"},
	}; !reflect.DeepEqual(got, want) {
		t.Fatalf("command flags = %+v, want retained executable contract %+v", got, want)
	}
	if got, want := command.Examples, []string{"pm acme message change-visibility --receipt-handle receipt --visibility-timeout 30 --legacy-only value --preview"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("command examples = %v, want retained runnable example %v", got, want)
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

func TestParseBatchOpenAPIArtifactMapsTopLevelWebhooks(t *testing.T) {
	artifact := []byte(`{
		"openapi": "3.1.0",
		"paths": {"/widgets": {"get": {"summary": "List widgets"}}},
		"webhooks": {
			"widget.created": {"post": {"summary": "Widget created"}},
			"widget.deleted": {"post": {"summary": "Widget deleted"}}
		}
	}`)

	endpoints, err := parseBatchOpenAPIArtifact(artifact)
	if err != nil {
		t.Fatalf("parse top-level webhook artifact: %v", err)
	}
	if len(endpoints) != 3 {
		t.Fatalf("parsed endpoints = %+v, want HTTP operation plus two webhook operations", endpoints)
	}
	var webhookCount int
	for _, endpoint := range endpoints {
		if endpoint.Method == "WEBHOOK" {
			webhookCount++
			if endpoint.SourceCoordinate == "" || !strings.HasPrefix(endpoint.SourceCoordinate, "webhooks[") {
				t.Fatalf("webhook endpoint = %+v, want exact webhooks coordinate", endpoint)
			}
		}
	}
	if webhookCount != 2 {
		t.Fatalf("parsed endpoints = %+v, want two visible webhook operations", endpoints)
	}
}

func TestParseBatchOpenAPIArtifactAcceptsOperationReferenceWithLocalMetadata(t *testing.T) {
	artifact := []byte(`{
		"openapi": "3.1.0",
		"paths": {
			"/widgets": {
				"delete": {
					"operationId": "customDelete",
					"$ref": "#/components/operations/customRequest"
				}
			}
		},
		"components": {
			"operations": {
				"customRequest": {
					"summary": "Send a custom request",
					"responses": {"200": {"description": "ok"}}
				}
			}
		}
	}`)

	endpoints, err := parseBatchOpenAPIArtifact(artifact)
	if err != nil {
		t.Fatalf("parse operation reference artifact: %v", err)
	}
	if len(endpoints) != 1 {
		t.Fatalf("endpoint count = %d, want 1", len(endpoints))
	}
	endpoint := endpoints[0]
	if endpoint.Method != http.MethodDelete || endpoint.Path != "/widgets" || endpoint.Summary != "customDelete" {
		t.Fatalf("endpoint = %+v, want DELETE /widgets with local operation metadata", endpoint)
	}
}

func TestParseBatchOpenAPIArtifactIgnoresNonRequestOperationMetadata(t *testing.T) {
	artifact := []byte(`{
		"openapi": "3.1.0",
		"paths": {
			"/widgets": {
				"post": {
					"summary": "Create widget",
					"examples": {"request": {"value": {"name": "example"}}},
					"callbacks": {"delivery": {"{$request.body#/callback}": {"post": {"responses": {"200": {"description": "OK"}}}}}},
					"responses": {"201": {"description": "Created"}}
				}
			}
		}
	}`)

	endpoints, err := parseBatchOpenAPIArtifact(artifact)
	if err != nil {
		t.Fatalf("parse OpenAPI operation metadata: %v", err)
	}
	if len(endpoints) != 1 || endpoints[0].Method != http.MethodPost || endpoints[0].Path != "/widgets" || endpoints[0].Summary != "Create widget" {
		t.Fatalf("parsed endpoints = %+v, want the provider request without callback deliveries", endpoints)
	}
}

func TestParseBatchOpenAPIArtifactAcceptsJSONUnicodeEscapes(t *testing.T) {
	artifact := []byte(`{
		"openapi": "3.1.0",
		"paths": {
			"/messages": {
				"post": {
					"summary": "Send a message \ud83d\udc4d",
					"responses": {"201": {"description": "Created"}}
				}
			}
		}
	}`)

	endpoints, err := parseBatchOpenAPIArtifact(artifact)
	if err != nil {
		t.Fatalf("parse JSON Unicode escape artifact: %v", err)
	}
	if len(endpoints) != 1 || endpoints[0].Method != http.MethodPost || endpoints[0].Path != "/messages" {
		t.Fatalf("parsed endpoints = %+v, want POST /messages", endpoints)
	}
}

func TestParseBatchOpenAPIArtifactIgnoresProviderNeutralFreeTier(t *testing.T) {
	artifact := []byte(`{
		"swagger": "2.0",
		"paths": {
			"/quotes": {
				"get": {
					"summary": "Get quotes",
					"freeTier": null,
					"responses": {"200": {"description": "OK"}}
				}
			}
		}
	}`)

	endpoints, err := parseBatchOpenAPIArtifact(artifact)
	if err != nil {
		t.Fatalf("parse provider neutral field artifact: %v", err)
	}
	if len(endpoints) != 1 || endpoints[0].Method != http.MethodGet || endpoints[0].Path != "/quotes" {
		t.Fatalf("parsed endpoints = %+v, want GET /quotes", endpoints)
	}
}

func TestMaterializedOperationIDSanitizesNonHTTPMethodLabels(t *testing.T) {
	used := map[string]bool{}
	if got, want := materializedOperationID("leadfeeder", "*", "/accounts/{account_id}/users", used), "leadfeeder.operation.accounts-account-id-users"; got != want {
		t.Fatalf("wildcard method ID = %q, want %q", got, want)
	}
	if got, want := materializedOperationID("leadfeeder", "POST/PATCH/DELETE", "/all", used), "leadfeeder.post-patch-delete.all"; got != want {
		t.Fatalf("aggregate method ID = %q, want %q", got, want)
	}
}

func TestParseBatchOpenAPIArtifactResolvesExternalPathItemReferences(t *testing.T) {
	artifact := []byte(`{
		"swagger": "2.0",
		"paths": {"/widgets": {"$ref": "paths/widgets.yaml#/widgets"}}
	}`)
	external := []byte(`widgets:
  get:
    summary: List widgets
`)
	inventory, err := parseBatchOpenAPIArtifactAt(artifact, "https://provider.example/openapi.yaml", func(rawURL string) ([]byte, error) {
		if rawURL != "https://provider.example/paths/widgets.yaml" {
			t.Fatalf("external reference URL = %q, want normalized provider URL", rawURL)
		}
		return external, nil
	})
	if err != nil {
		t.Fatalf("parse external path-item artifact: %v", err)
	}
	if len(inventory.Endpoints) != 1 || inventory.Endpoints[0].Method != "GET" || inventory.Endpoints[0].Path != "/widgets" {
		t.Fatalf("parsed endpoints = %+v, want resolved GET /widgets", inventory.Endpoints)
	}
	if len(inventory.Sources) != 2 || inventory.Sources[1].URL != "https://provider.example/paths/widgets.yaml" {
		t.Fatalf("parsed sources = %+v, want root plus referenced document", inventory.Sources)
	}
	if inventory.Endpoints[0].SourceCoordinate != "#/widgets/get" {
		t.Fatalf("external endpoint coordinate = %q, want source-local pointer", inventory.Endpoints[0].SourceCoordinate)
	}
}

func TestParseBatchOpenAPIArtifactStillFailsClosedForUnsupportedOperationContainers(t *testing.T) {
	tests := []struct {
		name     string
		artifact string
		want     string
	}{
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

func TestParseBatchPostmanArtifactNormalizesNestedRequestsAndDeduplicates(t *testing.T) {
	artifact := []byte(`{
		"info": {"schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"},
		"item": [
			{"name": "Widgets", "item": [
				{"name": "List widgets", "request": {"method": "GET", "url": {"raw": "{{base_url}}/widgets"}}},
				{"name": "List widgets example", "request": {"method": "GET", "url": {"path": ["widgets"]}}},
				{"name": "Update widget", "request": {"method": "PATCH", "url": "{{base_url}}/widgets/:widget_id"}}
			]}
		]
	}`)
	inventory, err := parseBatchPostmanArtifact(artifact, batchArtifactSource{URL: "https://provider.example/collection.json", Retrieved: "2026-08-08"})
	if err != nil {
		t.Fatalf("parse Postman artifact: %v", err)
	}
	if len(inventory.Endpoints) != 2 {
		t.Fatalf("Postman endpoints = %+v, want duplicate-normalized GET plus PATCH", inventory.Endpoints)
	}
	if inventory.Endpoints[0].Path != "/widgets" || inventory.Endpoints[1].Path != "/widgets/{widget_id}" {
		t.Fatalf("Postman paths = %+v, want normalized connector paths", inventory.Endpoints)
	}
	for _, endpoint := range inventory.Endpoints {
		if endpoint.SourceCoordinate == "" || endpoint.SourceURL == "" {
			t.Fatalf("Postman endpoint provenance = %+v, want collection coordinate and URL", endpoint)
		}
	}
}

func TestParseBatchHTMLReferenceTraversesOfficialMachineSource(t *testing.T) {
	rootURL := "https://provider.example/reference"
	machineURL := "https://provider.example/openapi.json"
	root := []byte(`<html><body><p>GET /widgets</p><a href="/openapi.json">OpenAPI export</a></body></html>`)
	machine := []byte(`{"openapi":"3.1.0","paths":{"/widgets":{"post":{"summary":"Create widget"}}}}`)
	inventory, err := parseBatchHTMLReference(root, batchArtifactSource{URL: rootURL, Kind: "html_reference", Retrieved: "2026-08-08"}, func(rawURL string) ([]byte, error) {
		if rawURL != machineURL {
			t.Fatalf("traversed URL = %q, want official machine source", rawURL)
		}
		return machine, nil
	})
	if err != nil {
		t.Fatalf("parse HTML reference: %v", err)
	}
	if len(inventory.Endpoints) != 2 || inventory.Endpoints[0].Path != "/widgets" || inventory.Endpoints[1].Path != "/widgets" {
		t.Fatalf("HTML/source endpoints = %+v, want explicit GET plus linked POST", inventory.Endpoints)
	}
	if len(inventory.Sources) != 2 {
		t.Fatalf("HTML/source provenance = %+v, want root and linked source", inventory.Sources)
	}
}

func TestCompleteBatchHTMLReferenceRootAvoidsTraversalWhenCountIsMet(t *testing.T) {
	artifact := []byte(`<html><body>
		GET /projects
		POST /projects
		<a href="/unrelated-reference">unrelated reference</a>
	</body></html>`)
	source := batchArtifactSource{URL: "https://provider.example/reference", Kind: "html_reference", Version: "v1", Retrieved: "2026-08-09"}
	candidate := BatchManifestConnector{Connector: "example", OperationsTotal: 2}

	inventory, complete, err := completeBatchHTMLReferenceRoot(candidate, artifact, source)
	if err != nil {
		t.Fatalf("parse complete root reference: %v", err)
	}
	if !complete {
		t.Fatal("complete root reference = false, want true without traversing unrelated links")
	}
	if len(inventory.Endpoints) != 2 {
		t.Fatalf("root inventory endpoints = %+v, want two documented requests", inventory.Endpoints)
	}
}

func TestParseBatchHTMLReferenceSkipsStaticAssetCandidates(t *testing.T) {
	rootURL := "https://provider.example/reference"
	machineURL := "https://provider.example/openapi.json"
	root := []byte(`<html><head><link href="/build/css/api-doc.css" rel="stylesheet"></head><body><a href="/openapi.json">OpenAPI export</a></body></html>`)
	machine := []byte(`{"openapi":"3.1.0","paths":{"/widgets":{"get":{"summary":"List widgets"}}}}`)

	inventory, err := parseBatchHTMLReference(root, batchArtifactSource{URL: rootURL, Kind: "html_reference", Retrieved: "2026-08-09"}, func(rawURL string) ([]byte, error) {
		if rawURL != machineURL {
			t.Fatalf("traversed URL = %q, want only official machine source", rawURL)
		}
		return machine, nil
	})
	if err != nil {
		t.Fatalf("parse HTML reference: %v", err)
	}
	if len(inventory.Endpoints) != 1 || inventory.Endpoints[0].Path != "/widgets" {
		t.Fatalf("HTML/source endpoints = %+v, want linked machine operation only", inventory.Endpoints)
	}
}

func TestIsLikelyBatchReferenceLinkAllowsOnlyHTTPSchemesOrRelativeReferences(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want bool
	}{
		{
			name: "data URI is not a reference document",
			raw:  "data:application/openapi+json,{}",
			want: false,
		},
		{
			name: "vbscript URI is not a reference document",
			raw:  "vbscript:msgbox('openapi')",
			want: false,
		},
		{
			name: "mailto URI is not a reference document",
			raw:  "mailto:openapi@example.test",
			want: false,
		},
		{
			name: "HTTPS OpenAPI document remains eligible",
			raw:  "https://provider.example/openapi.json",
			want: true,
		},
		{
			name: "relative OpenAPI document remains eligible",
			raw:  "/reference/openapi.json",
			want: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := isLikelyBatchReferenceLink(test.raw); got != test.want {
				t.Fatalf("isLikelyBatchReferenceLink(%q) = %t, want %t", test.raw, got, test.want)
			}
		})
	}
}

func TestParseBatchHTMLReferenceSkipsFeedCandidates(t *testing.T) {
	rootURL := "https://provider.example/reference"
	machineURL := "https://provider.example/openapi.json"
	root := []byte(`<html><head><link href="/reference/rss.xml" rel="alternate" type="application/rss+xml"></head><body><a href="/openapi.json">OpenAPI export</a></body></html>`)
	machine := []byte(`{"openapi":"3.1.0","paths":{"/widgets":{"get":{"summary":"List widgets"}}}}`)

	inventory, err := parseBatchHTMLReference(root, batchArtifactSource{URL: rootURL, Kind: "html_reference", Retrieved: "2026-08-09"}, func(rawURL string) ([]byte, error) {
		if rawURL != machineURL {
			t.Fatalf("traversed URL = %q, want only official machine source", rawURL)
		}
		return machine, nil
	})
	if err != nil {
		t.Fatalf("parse HTML reference: %v", err)
	}
	if len(inventory.Endpoints) != 1 || inventory.Endpoints[0].Path != "/widgets" {
		t.Fatalf("HTML/source endpoints = %+v, want linked machine operation only", inventory.Endpoints)
	}
}

func TestParseBatchHTMLReferenceRecognizesMarkdownCodeOperations(t *testing.T) {
	root := []byte("<mark style=\"color:green;\">`GET`</mark> `/widgets`\n<mark style=\"color:blue;\">`POST`</mark> `/widgets`")

	inventory, err := parseBatchHTMLReference(root, batchArtifactSource{URL: "https://provider.example/reference.md", Kind: "html_reference", Retrieved: "2026-08-09"}, nil)
	if err != nil {
		t.Fatalf("parse Markdown HTML reference: %v", err)
	}
	if len(inventory.Endpoints) != 2 || inventory.Endpoints[0].Method != http.MethodGet || inventory.Endpoints[0].Path != "/widgets" || inventory.Endpoints[1].Method != http.MethodPost || inventory.Endpoints[1].Path != "/widgets" {
		t.Fatalf("Markdown code endpoints = %+v, want GET and POST /widgets", inventory.Endpoints)
	}
}

func TestNormalizeBatchHTMLOperationPathConvertsAnglePlaceholders(t *testing.T) {
	if got, want := normalizeBatchHTMLOperationPath("/pypi/&lt;project&gt;/&lt;version&gt;/json"), "/pypi/{project}/{version}/json"; got != want {
		t.Fatalf("normalized HTML path = %q, want %q", got, want)
	}
}

func TestParseBatchHTMLReferencePrefersStructuralRoutesOverRequestExamples(t *testing.T) {
	root := []byte(`<html><body>
		<p>GET /discovery/v2/attractions</p>
		<p>GET /discovery/v2/attractions/{id}</p>
		<p>GET /discovery/v2/events/{id}/images</p>
		<p>GET /users/{id}</p>
		<pre>GET /discovery/v2/attractions.json</pre>
		<pre>GET /discovery/v2/attractions/K8vZ9175BhV.json</pre>
		<pre>GET /discovery/v2/events/0B004F0401BD55E5/images.json</pre>
		<pre>GET /users/me</pre>
	</body></html>`)

	inventory, err := parseBatchHTMLReference(root, batchArtifactSource{URL: "https://provider.example/reference", Kind: "html_reference", Retrieved: "2026-08-09"}, nil)
	if err != nil {
		t.Fatalf("parse HTML reference: %v", err)
	}
	paths := map[string]bool{}
	for _, endpoint := range inventory.Endpoints {
		paths[endpoint.Path] = true
	}
	for _, path := range []string{
		"/discovery/v2/attractions",
		"/discovery/v2/attractions/{id}",
		"/discovery/v2/events/{id}/images",
		"/users/{id}",
		"/users/me",
	} {
		if !paths[path] {
			t.Fatalf("HTML routes = %+v, want %q", inventory.Endpoints, path)
		}
	}
	if len(paths) != 5 || paths["/discovery/v2/attractions.json"] || paths["/discovery/v2/attractions/K8vZ9175BhV.json"] || paths["/discovery/v2/events/0B004F0401BD55E5/images.json"] {
		t.Fatalf("HTML routes = %+v, want structural routes without concrete request examples", inventory.Endpoints)
	}
}

func TestTicketmasterGeneratedSurfaceExcludesReferenceExamples(t *testing.T) {
	root := filepath.Join("..", "..", "internal", "connectors", "defs", "ticketmaster")
	var surface engine.APISurface
	decodeJSONFile(t, filepath.Join(root, "api_surface.json"), &surface)
	if len(surface.Endpoints) != 19 || !strings.Contains(surface.Scope, "13 documented operations") {
		t.Fatalf("Ticketmaster surface = %d endpoint(s), scope %q, want 19 endpoints from 13 documented operations", len(surface.Endpoints), surface.Scope)
	}
	for _, endpoint := range surface.Endpoints {
		if endpoint.Path == "/discovery/v2/attractions/K8vZ9175BhV.json" {
			t.Fatalf("Ticketmaster surface retains concrete example endpoint %+v", endpoint)
		}
	}

	var operations struct {
		Operations []json.RawMessage `json:"operations"`
	}
	decodeJSONFile(t, filepath.Join(root, "operations.json"), &operations)
	if len(operations.Operations) != 15 {
		t.Fatalf("Ticketmaster operations = %d, want 15 after suppressing request examples", len(operations.Operations))
	}

	var cli struct {
		Commands []json.RawMessage `json:"commands"`
	}
	decodeJSONFile(t, filepath.Join(root, "cli_surface.json"), &cli)
	if len(cli.Commands) != 19 {
		t.Fatalf("Ticketmaster CLI commands = %d, want 19 after suppressing request examples", len(cli.Commands))
	}
}

func TestParseBatchHTMLReferenceDoesNotScanMachineDocumentProse(t *testing.T) {
	rootURL := "https://provider.example/reference"
	machineURL := "https://provider.example/openapi.json"
	root := []byte(`<html><body><a href="/openapi.json">OpenAPI export</a></body></html>`)
	machine := []byte(`{"openapi":"3.1.0","paths":{"/widgets":{"get":{"description":"Use GET /internal-preview"}}}}`)
	inventory, err := parseBatchHTMLReference(root, batchArtifactSource{URL: rootURL, Kind: "html_reference", Retrieved: "2026-08-08"}, func(rawURL string) ([]byte, error) {
		if rawURL != machineURL {
			t.Fatalf("traversed URL = %q, want official machine source", rawURL)
		}
		return machine, nil
	})
	if err != nil {
		t.Fatalf("parse HTML reference: %v", err)
	}
	if len(inventory.Endpoints) != 1 || inventory.Endpoints[0].Method != http.MethodGet || inventory.Endpoints[0].Path != "/widgets" {
		t.Fatalf("machine inventory = %+v, want only declared machine operation", inventory.Endpoints)
	}
}

func TestParseBatchHTMLReferenceTraversesMarkdownMachineSource(t *testing.T) {
	rootURL := "https://provider.example/docs/llms.txt"
	machineURL := "https://provider.example/docs/openapi.json"
	root := []byte("GET /widgets\n[OpenAPI export](/docs/openapi.json)\n")
	machine := []byte(`{"openapi":"3.1.0","paths":{"/widgets":{"post":{"summary":"Create widget"}}}}`)
	inventory, err := parseBatchHTMLReference(root, batchArtifactSource{URL: rootURL, Kind: "openapi_fragments", Retrieved: "2026-08-08"}, func(rawURL string) ([]byte, error) {
		if rawURL != machineURL {
			t.Fatalf("traversed URL = %q, want Markdown-linked machine source", rawURL)
		}
		return machine, nil
	})
	if err != nil {
		t.Fatalf("parse Markdown reference: %v", err)
	}
	if len(inventory.Endpoints) != 2 || len(inventory.Sources) != 2 {
		t.Fatalf("Markdown/source inventory = %+v, want explicit GET plus linked POST and two sources", inventory)
	}
}

func TestParseBatchHTMLReferenceResolvesNestedLinksAgainstCurrentPage(t *testing.T) {
	rootURL := "https://provider.example/docs/index"
	referenceURL := "https://provider.example/docs/v1/reference"
	machineURL := "https://provider.example/docs/v1/openapi.json"
	root := []byte(`<html><body><a href="v1/reference">Reference</a></body></html>`)
	reference := []byte(`<html><body><a href="openapi.json">OpenAPI export</a></body></html>`)
	machine := []byte(`{"openapi":"3.1.0","paths":{"/widgets":{"get":{"summary":"List widgets"}}}}`)
	wrongMachine := []byte(`{"openapi":"3.1.0","paths":{"/wrong":{"get":{"summary":"Wrong document"}}}}`)
	fetched := []string{}
	inventory, err := parseBatchHTMLReference(root, batchArtifactSource{URL: rootURL, Kind: "html_reference", Retrieved: "2026-08-08"}, func(rawURL string) ([]byte, error) {
		fetched = append(fetched, rawURL)
		switch rawURL {
		case referenceURL:
			return reference, nil
		case machineURL:
			return machine, nil
		case "https://provider.example/docs/openapi.json":
			return wrongMachine, nil
		default:
			t.Fatalf("traversed URL = %q, want nested official reference", rawURL)
			return nil, nil
		}
	})
	if err != nil {
		t.Fatalf("parse nested HTML reference: %v", err)
	}
	if len(fetched) != 2 || fetched[0] != referenceURL || fetched[1] != machineURL {
		t.Fatalf("traversed URLs = %+v, want %q then %q", fetched, referenceURL, machineURL)
	}
	if len(inventory.Endpoints) != 1 || inventory.Endpoints[0].Path != "/widgets" || inventory.Endpoints[0].SourceURL != machineURL {
		t.Fatalf("nested reference inventory = %+v, want machine endpoint from current-page-relative link", inventory)
	}
}

func TestParseBatchHTMLReferenceStripsLinkedDocumentFragment(t *testing.T) {
	rootURL := "https://provider.example/reference"
	machineURL := "https://provider.example/openapi.json"
	root := []byte(`<html><body><p>GET /widgets</p><a href="/openapi.json#/paths/~1widgets">OpenAPI export</a></body></html>`)
	machine := []byte(`{"openapi":"3.1.0","paths":{"/widgets":{"post":{"summary":"Create widget"}}}}`)
	inventory, err := parseBatchHTMLReference(root, batchArtifactSource{URL: rootURL, Kind: "html_reference", Retrieved: "2026-08-08"}, func(rawURL string) ([]byte, error) {
		if rawURL != machineURL {
			t.Fatalf("traversed URL = %q, want fragment-free machine source", rawURL)
		}
		return machine, nil
	})
	if err != nil {
		t.Fatalf("parse fragment-linked HTML reference: %v", err)
	}
	if len(inventory.Endpoints) != 2 || len(inventory.Sources) != 2 || inventory.Endpoints[1].Method != http.MethodPost || inventory.Endpoints[1].SourceURL != machineURL {
		t.Fatalf("fragment-linked inventory = %+v, want explicit and linked machine operations", inventory)
	}
}

func TestParseBatchHTMLReferenceFailsClosedForMalformedSelectedLink(t *testing.T) {
	root := []byte(`<html><body><p>GET /widgets</p><a href="/openapi%zz.json">OpenAPI export</a></body></html>`)
	_, err := parseBatchHTMLReference(root, batchArtifactSource{URL: "https://provider.example/reference", Kind: "html_reference", Retrieved: "2026-08-08"}, nil)
	var unknown *batchArtifactInventoryUnknownError
	if !errors.As(err, &unknown) || !strings.Contains(err.Error(), "malformed selected link") {
		t.Fatalf("parse error = %v, want malformed-link unknown inventory", err)
	}
}

func TestParseBatchHTMLReferenceFailsClosedForSelectedLinkRejectedByAdmission(t *testing.T) {
	root := []byte(`<html><body><p>GET /widgets</p><a href="/openapi.json?api_key=fixture;v=1">OpenAPI export</a></body></html>`)
	_, err := parseBatchHTMLReference(root, batchArtifactSource{URL: "https://provider.example/reference", Kind: "html_reference", Retrieved: "2026-08-08"}, nil)
	var unknown *batchArtifactInventoryUnknownError
	if !errors.As(err, &unknown) || !strings.Contains(err.Error(), "failed URL admission") {
		t.Fatalf("parse error = %v, want admission failure unknown inventory", err)
	}
}

func TestParseBatchHTMLReferenceFailsClosedForSelectedOffHostLink(t *testing.T) {
	root := []byte(`<html><body><p>GET /widgets</p><a href="https://api.provider.example/openapi.json">OpenAPI export</a></body></html>`)
	_, err := parseBatchHTMLReference(root, batchArtifactSource{URL: "https://docs.provider.example/reference", Kind: "html_reference", Retrieved: "2026-08-08"}, nil)
	var unknown *batchArtifactInventoryUnknownError
	if !errors.As(err, &unknown) || !strings.Contains(err.Error(), "trusted provider host") {
		t.Fatalf("parse error = %v, want off-host selected-link unknown inventory", err)
	}
}

func TestParseBatchHTMLReferenceSkipsOffHostNonMachineNavigation(t *testing.T) {
	rootURL := "https://docs.provider.example/reference"
	machineURL := "https://docs.provider.example/openapi.json"
	root := []byte(`<html><body><a href="https://support.provider.example/api/reference">Support navigation</a><a href="/openapi.json">OpenAPI export</a></body></html>`)
	machine := []byte(`{"openapi":"3.1.0","paths":{"/widgets":{"get":{"summary":"List widgets"}}}}`)

	inventory, err := parseBatchHTMLReference(root, batchArtifactSource{URL: rootURL, Kind: "html_reference", Retrieved: "2026-08-09"}, func(rawURL string) ([]byte, error) {
		if rawURL != machineURL {
			t.Fatalf("traversed URL = %q, want only same-host machine source", rawURL)
		}
		return machine, nil
	})
	if err != nil {
		t.Fatalf("parse HTML reference: %v", err)
	}
	if len(inventory.Endpoints) != 1 || inventory.Endpoints[0].Path != "/widgets" {
		t.Fatalf("HTML/source endpoints = %+v, want linked machine operation only", inventory.Endpoints)
	}
}

func TestParseBatchHTMLReferenceFailsClosedForSelectedLinkFailures(t *testing.T) {
	rootURL := "https://provider.example/reference"
	machineURL := "https://provider.example/openapi.json"
	root := []byte(`<html><body><p>GET /widgets</p><a href="/openapi.json">OpenAPI export</a></body></html>`)
	tests := []struct {
		name     string
		raw      []byte
		fetchErr error
		want     string
	}{
		{
			name:     "incomplete response",
			fetchErr: batchArtifactInventoryUnknown("artifact response is incomplete: received HTTP 206 Partial Content"),
			want:     "HTTP 206",
		},
		{
			name: "malformed machine artifact",
			raw:  []byte(`{"openapi":"3.1.0","paths":[]}`),
			want: "machine-readable artifact",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := parseBatchHTMLReference(root, batchArtifactSource{URL: rootURL, Kind: "html_reference", Retrieved: "2026-08-08"}, func(rawURL string) ([]byte, error) {
				if rawURL != machineURL {
					t.Fatalf("traversed URL = %q, want official machine source", rawURL)
				}
				if test.fetchErr != nil {
					return nil, test.fetchErr
				}
				return test.raw, nil
			})
			var unknown *batchArtifactInventoryUnknownError
			if !errors.As(err, &unknown) || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("parse error = %v, want unknown inventory containing %q", err, test.want)
			}
		})
	}
}

func TestParseBatchHTMLReferenceFailsClosedAtDocumentLimit(t *testing.T) {
	var root strings.Builder
	root.WriteString("<html><body>GET /widgets")
	for index := 0; index < maxBatchArtifactReferenceDocuments; index++ {
		root.WriteString(`<a href="/reference/`)
		root.WriteString(strconv.Itoa(index))
		root.WriteString(`">Reference</a>`)
	}
	root.WriteString("</body></html>")

	_, err := parseBatchHTMLReference([]byte(root.String()), batchArtifactSource{URL: "https://provider.example/reference", Kind: "html_reference", Retrieved: "2026-08-08"}, func(string) ([]byte, error) {
		return []byte("GET /linked"), nil
	})
	var unknown *batchArtifactInventoryUnknownError
	if !errors.As(err, &unknown) || !strings.Contains(err.Error(), "document limit") {
		t.Fatalf("parse error = %v, want unknown inventory document-limit failure", err)
	}
}

func TestParseBatchHTMLReferenceSharesBudgetWithLinkedOpenAPIReferences(t *testing.T) {
	var root strings.Builder
	root.WriteString("<html><body>")
	for index := 0; index < maxBatchArtifactReferenceDocuments-2; index++ {
		root.WriteString(`<a href="/reference/`)
		root.WriteString(strconv.Itoa(index))
		root.WriteString(`">Reference</a>`)
	}
	root.WriteString(`<a href="/machine.openapi.json">OpenAPI export</a></body></html>`)
	machine := []byte(`{"openapi":"3.1.0","paths":{"/widgets":{"$ref":"/paths/widgets.yaml#/widgets"}}}`)
	_, err := parseBatchHTMLReference([]byte(root.String()), batchArtifactSource{URL: "https://provider.example/reference", Kind: "html_reference", Retrieved: "2026-08-08"}, func(rawURL string) ([]byte, error) {
		switch {
		case rawURL == "https://provider.example/machine.openapi.json":
			return machine, nil
		case strings.HasPrefix(rawURL, "https://provider.example/reference/"):
			return []byte(`<html><body></body></html>`), nil
		case rawURL == "https://provider.example/paths/widgets.yaml":
			t.Fatal("external path-item reference fetched after shared document budget was exhausted")
		default:
			t.Fatalf("traversed URL = %q, want bounded official reference", rawURL)
		}
		return nil, nil
	})
	var unknown *batchArtifactInventoryUnknownError
	if !errors.As(err, &unknown) || !strings.Contains(err.Error(), "reference limit") {
		t.Fatalf("parse error = %v, want shared-reference-budget failure", err)
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
	if err := validateBatchArtifactURL("https://analyticsdata.googleapis.com/$discovery/rest?version=v1beta"); err != nil {
		t.Fatalf("validateBatchArtifactURL accepted official versioned discovery URL: %v", err)
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

func TestParseBatchReferenceURLRejectsUnsafeQueryParameters(t *testing.T) {
	for _, key := range []string{"api_key", "api-key", "apikey", "access-key", "auth-key", "key", "sig", "signature", "subscription-key"} {
		t.Run(key, func(t *testing.T) {
			_, err := parseBatchReferenceURL("https://example.test/openapi.json?" + key + "=fixture")
			if err == nil || !strings.Contains(err.Error(), "non-sensitive") {
				t.Fatalf("parseBatchReferenceURL(%q) error = %v, want unsafe-query rejection", key, err)
			}
		})
	}
}

func TestParseBatchReferenceURLAcceptsNonSensitiveQueryParameters(t *testing.T) {
	for _, key := range []string{"api-version", "format", "lang", "locale", "version"} {
		t.Run(key, func(t *testing.T) {
			if _, err := parseBatchReferenceURL("https://example.test/openapi.json?" + key + "=fixture"); err != nil {
				t.Fatalf("parseBatchReferenceURL(%q) error = %v, want non-sensitive query admission", key, err)
			}
		})
	}
}

func TestMergeBatchArtifactInventoriesPreservesFallbackAlternatives(t *testing.T) {
	primary := batchArtifactInventory{Endpoints: []batchArtifactEndpoint{{
		Method:           http.MethodGet,
		Path:             "/widgets",
		SourceURL:        "https://provider.example/openapi.json",
		SourceKind:       "openapi",
		SourceCoordinate: `paths["/widgets"].get`,
	}}}
	fallback := batchArtifactInventory{Endpoints: []batchArtifactEndpoint{{
		Method:           http.MethodGet,
		Path:             "/widgets",
		SourceURL:        "https://provider.example/reference/first.json",
		SourceKind:       "official-reference",
		SourceCoordinate: `#/paths/~1widgets/get`,
		Alternatives: []batchArtifactEndpointAlternative{{
			SourceURL:        "https://provider.example/reference/second.json",
			SourceKind:       "official-reference",
			SourceCoordinate: `#/paths/~1widgets/get`,
		}, {
			SourceURL:        "https://provider.example/reference/first.json",
			SourceKind:       "official-reference",
			SourceCoordinate: `#/paths/~1widgets/get`,
		}},
	}}}
	merged := mergeBatchArtifactInventories(primary, fallback)
	if len(merged.Endpoints) != 1 || len(merged.Endpoints[0].Alternatives) != 2 {
		t.Fatalf("merged endpoints = %+v, want both deduplicated fallback citations", merged.Endpoints)
	}
	if merged.Endpoints[0].Alternatives[0].SourceURL != "https://provider.example/reference/first.json" || merged.Endpoints[0].Alternatives[1].SourceURL != "https://provider.example/reference/second.json" {
		t.Fatalf("merged alternatives = %+v, want fallback primary and secondary citations", merged.Endpoints[0].Alternatives)
	}
}

func TestParseBatchReferenceURLRejectsMalformedQuery(t *testing.T) {
	_, err := parseBatchReferenceURL("https://example.test/openapi.json?api_key=fixture;v=1")
	if err == nil || !strings.Contains(err.Error(), "well-formed") {
		t.Fatalf("parseBatchReferenceURL malformed query error = %v, want rejection", err)
	}
}

func TestReadBatchMaterializeArtifactCacheAcceptsOfficialReferenceText(t *testing.T) {
	artifactDir := t.TempDir()
	path := filepath.Join(artifactDir, "cli-surface.txt")
	if err := os.WriteFile(path, []byte("GET /widgets\n"), 0o644); err != nil {
		t.Fatalf("write text artifact cache: %v", err)
	}
	raw, err := readBatchMaterializeArtifactCache(artifactDir, "cli-surface")
	if err != nil {
		t.Fatalf("read text artifact cache: %v", err)
	}
	if string(raw) != "GET /widgets\n" {
		t.Fatalf("cached text = %q, want exact provider reference source", raw)
	}
}

func TestBatchArtifactSourceFetcherRejectsSymlinkedCacheComponent(t *testing.T) {
	artifactDir := t.TempDir()
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(artifactDir, "acme")); err != nil {
		t.Fatalf("symlink cache component: %v", err)
	}
	fetch := batchArtifactSourceFetcher(batchMaterializeOptions{artifactDir: artifactDir}, BatchManifestConnector{Connector: "acme"})
	if _, err := fetch("https://provider.example/openapi.yaml"); err == nil || !strings.Contains(err.Error(), "non-symlink directory") {
		t.Fatalf("fetch error = %v, want symlinked cache rejection", err)
	}
	entries, err := os.ReadDir(outside)
	if err != nil {
		t.Fatalf("read external directory: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("external cache directory entries = %+v, want no write through symlink", entries)
	}
}

func TestBatchArtifactSourceFetcherCachedReferencesOnlyRefusesCacheMiss(t *testing.T) {
	artifactDir := t.TempDir()
	fetch := batchArtifactSourceFetcher(batchMaterializeOptions{
		artifactDir:          artifactDir,
		cachedReferencesOnly: true,
	}, BatchManifestConnector{Connector: "acme"})
	if _, err := fetch("https://provider.example/openapi.yaml"); err == nil || !strings.Contains(err.Error(), "cached-references-only") {
		t.Fatalf("cache-only fetch error = %v, want cache-miss refusal without a network fetch", err)
	}
}

func TestParseBatchMaterializeOptionsCachedReferencesOnlyRequiresArtifactDir(t *testing.T) {
	args := []string{
		"--manifest", "manifest.json",
		"--source-defs-root", "source-defs",
		"--retrieved-at", "2026-08-09",
		"--report", "report.json",
		"--cached-references-only",
	}
	if _, err := parseBatchMaterializeOptions(args); err == nil || !strings.Contains(err.Error(), "requires --artifact-dir") {
		t.Fatalf("cache-only options error = %v, want artifact cache requirement", err)
	}
	args = append(args[:len(args)-1], "--artifact-dir", "artifacts", "--cached-references-only")
	opts, err := parseBatchMaterializeOptions(args)
	if err != nil {
		t.Fatalf("parse cache-only options: %v", err)
	}
	if !opts.cachedReferencesOnly || opts.artifactDir != "artifacts" {
		t.Fatalf("cache-only options = %+v, want enabled artifact cache", opts)
	}
}

func TestParseBatchManifestArtifactRequiresCompleteReconciledInventory(t *testing.T) {
	tests := []struct {
		name       string
		primaryRaw string
		total      int
	}{
		{
			name:       "primary and fallback remain short",
			primaryRaw: `{"openapi":"3.1.0","paths":{"/widgets":{"get":{}}}}`,
			total:      3,
		},
		{
			name:       "fallback only remains short",
			primaryRaw: `not an OpenAPI document`,
			total:      2,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			artifactDir := t.TempDir()
			candidate := BatchManifestConnector{
				Connector:            "acme",
				OperationsTotal:      test.total,
				ProviderReferenceURL: "https://provider.example/reference",
				Artifact:             BatchArtifact{URL: "https://provider.example/acme.json", Kind: "openapi", Version: "3.1.0"},
			}
			referencePath, err := batchArtifactReferenceCachePath(artifactDir, candidate.Connector, candidate.ProviderReferenceURL)
			if err != nil {
				t.Fatalf("reference cache path: %v", err)
			}
			if err := os.WriteFile(referencePath, []byte("GET /gadgets\n"), 0o644); err != nil {
				t.Fatalf("write reference cache: %v", err)
			}
			_, err = parseBatchManifestArtifact(batchMaterializeOptions{artifactDir: artifactDir, retrievedAt: "2026-08-08"}, candidate, []byte(test.primaryRaw))
			var unknown *batchArtifactInventoryUnknownError
			if !errors.As(err, &unknown) || !strings.Contains(err.Error(), "below the ledger") {
				t.Fatalf("parse error = %v, want incomplete reconciled inventory rejection", err)
			}
		})
	}
}

func TestParseBatchManifestArtifactSharesBudgetWithFallback(t *testing.T) {
	artifactDir := t.TempDir()
	candidate := BatchManifestConnector{
		Connector:            "acme",
		OperationsTotal:      maxBatchArtifactReferenceDocuments,
		ProviderReferenceURL: "https://provider.example/reference",
		Artifact:             BatchArtifact{URL: "https://provider.example/acme.json", Kind: "openapi", Version: "3.1.0"},
	}
	var paths strings.Builder
	paths.WriteString(`{"openapi":"3.1.0","paths":{`)
	for index := 0; index < maxBatchArtifactReferenceDocuments-1; index++ {
		if index > 0 {
			paths.WriteByte(',')
		}
		paths.WriteString(strconv.Quote("/widgets/" + strconv.Itoa(index)))
		paths.WriteString(`:{"$ref":"references/`)
		paths.WriteString(strconv.Itoa(index))
		paths.WriteString(`.yaml#/item"}`)
		referenceURL := "https://provider.example/references/" + strconv.Itoa(index) + ".yaml"
		referencePath, err := batchArtifactReferenceCachePath(artifactDir, candidate.Connector, referenceURL)
		if err != nil {
			t.Fatalf("reference cache path %d: %v", index, err)
		}
		if err := os.WriteFile(referencePath, []byte("item:\n  get: {}\n"), 0o644); err != nil {
			t.Fatalf("write reference cache %d: %v", index, err)
		}
	}
	paths.WriteString(`}}`)

	_, err := parseBatchManifestArtifact(batchMaterializeOptions{artifactDir: artifactDir, retrievedAt: "2026-08-08"}, candidate, []byte(paths.String()))
	var unknown *batchArtifactInventoryUnknownError
	if !errors.As(err, &unknown) || !strings.Contains(err.Error(), "64-document") {
		t.Fatalf("parse error = %v, want candidate-wide traversal-budget rejection", err)
	}
}

func TestStagedFloatExternalReferenceCoordinatesAreSourceLocal(t *testing.T) {
	stageRoot := filepath.Join("..", "..", ".planning", "phases", "persistiq-artifact-materialize-pilot-r1", "generalization-validation-2026-08-08")
	var surface struct {
		Endpoints []struct {
			Method     string `json:"method"`
			Path       string `json:"path"`
			Provenance struct {
				Artifact   string `json:"artifact"`
				SourceURL  string `json:"source_url"`
				SourceKind string `json:"source_kind"`
				Coordinate string `json:"coordinate"`
			} `json:"provenance"`
		} `json:"endpoints"`
	}
	decodeJSONFile(t, filepath.Join(stageRoot, "generated-defs", "float", "api_surface.json"), &surface)
	documents := map[string]batchArtifactDocument{}
	coordinates := map[string]string{}
	artifacts := map[string]string{}
	externalEndpoints := 0
	for _, endpoint := range surface.Endpoints {
		if endpoint.Provenance.SourceKind != "referenced-document" {
			continue
		}
		externalEndpoints++
		coordinate := endpoint.Provenance.Coordinate
		if !strings.HasPrefix(coordinate, "#/") {
			t.Fatalf("external endpoint %s %s coordinate = %q, want source-local pointer", endpoint.Method, endpoint.Path, coordinate)
		}
		document, ok := documents[endpoint.Provenance.SourceURL]
		if !ok {
			cachePath, err := batchArtifactReferenceCachePath(filepath.Join(stageRoot, "artifacts"), "float", endpoint.Provenance.SourceURL)
			if err != nil {
				t.Fatalf("reference cache path for %q: %v", endpoint.Provenance.SourceURL, err)
			}
			raw, err := readBoundedArtifactFile(cachePath)
			if err != nil {
				t.Fatalf("read reference source %q: %v", endpoint.Provenance.SourceURL, err)
			}
			document, err = parseBatchArtifactDocument(raw, batchArtifactSource{URL: endpoint.Provenance.SourceURL})
			if err != nil {
				t.Fatalf("parse reference source %q: %v", endpoint.Provenance.SourceURL, err)
			}
			documents[endpoint.Provenance.SourceURL] = document
		}
		pointerAndMethod := strings.TrimPrefix(coordinate, "#")
		separator := strings.LastIndex(pointerAndMethod, "/")
		if separator <= 0 {
			t.Fatalf("external endpoint %s %s coordinate = %q, want pointer and method", endpoint.Method, endpoint.Path, coordinate)
		}
		node, err := resolveBatchArtifactJSONPointer(document.Root, pointerAndMethod[:separator], coordinate)
		if err != nil {
			t.Fatalf("resolve external endpoint %s %s coordinate %q: %v", endpoint.Method, endpoint.Path, coordinate, err)
		}
		fields, err := batchYAMLFields(node)
		if err != nil {
			t.Fatalf("external endpoint %s %s coordinate %q fields: %v", endpoint.Method, endpoint.Path, coordinate, err)
		}
		if _, ok := fields[pointerAndMethod[separator+1:]]; !ok {
			t.Fatalf("external endpoint %s %s coordinate = %q, want method in source document", endpoint.Method, endpoint.Path, coordinate)
		}
		key := batchArtifactEndpointKey(endpoint.Method, endpoint.Path)
		coordinates[key] = coordinate
		artifacts[key] = endpoint.Provenance.Artifact
	}
	if externalEndpoints == 0 {
		t.Fatal("staged Float surface has no external-reference endpoints")
	}

	var mapping struct {
		Operations []struct {
			Method     string `json:"method"`
			Path       string `json:"path"`
			Provenance struct {
				Artifact   string `json:"artifact"`
				Coordinate string `json:"coordinate"`
			} `json:"provenance"`
		} `json:"operations"`
	}
	decodeJSONFile(t, filepath.Join(stageRoot, "reports", "float-operation-mapping-rerun-2.json"), &mapping)
	for _, operation := range mapping.Operations {
		key := batchArtifactEndpointKey(operation.Method, operation.Path)
		if want, ok := coordinates[key]; ok && operation.Provenance.Coordinate != want {
			t.Fatalf("Float operation mapping %s %s coordinate = %q, want %q", operation.Method, operation.Path, operation.Provenance.Coordinate, want)
		}
		if want, ok := artifacts[key]; ok && operation.Provenance.Artifact != want {
			t.Fatalf("Float operation mapping %s %s artifact = %q, want %q", operation.Method, operation.Path, operation.Provenance.Artifact, want)
		}
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
	setBatchManifestOperationCounts(t, manifestPath, 3, 2, 1)
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
	setBatchManifestOperationCounts(t, manifestPath, 3, 1, 2)
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

func decodeJSONFile(t *testing.T, path string, value any) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if err := json.Unmarshal(raw, value); err != nil {
		t.Fatalf("decode %s: %v", path, err)
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

func setBatchManifestOperationCounts(t *testing.T, path string, total, read, write int) {
	t.Helper()
	manifest, err := readBatchManifest(path)
	if err != nil {
		t.Fatalf("read manifest fixture: %v", err)
	}
	if len(manifest.Connectors) != 1 {
		t.Fatalf("manifest fixture connectors = %d, want one", len(manifest.Connectors))
	}
	manifest.Connectors[0].OperationsTotal = total
	manifest.Connectors[0].OperationsRead = read
	manifest.Connectors[0].OperationsWrite = write
	if err := writeBatchManifest(path, manifest); err != nil {
		t.Fatalf("write manifest fixture counts: %v", err)
	}
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
