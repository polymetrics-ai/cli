package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors/conformance"
	"polymetrics.ai/internal/connectors/engine"
)

const (
	outreachOpenAPIURL    = "https://api.outreach.io/api/v2/schema/openapi.json"
	outreachOpenAPISHA256 = "d1f697f6558fda68cd6d8059044e323c20849aeebf303e15c43e0eb9875e2ef6"
	outreachCustomURL     = "https://developers.outreach.io/api/custom-objects"
	outreachCustomSHA256  = "2e74714a933b74cb9a83ddbdb18eeb0b9d045115102ed7465021a45db19e3dda"
)

// TestSourceImportOutreachReferenceProjectsCitedOperationsWithoutFetching is
// the retained-Outreach vertical proof. The lock is declaration evidence only:
// its hashes cite the unavailable captures but must neither trigger a network
// request nor promote a command to executable.
func TestSourceImportOutreachReferenceProjectsCitedOperationsWithoutFetching(t *testing.T) {
	t.Parallel()
	raw := sourceImportOutreachReferenceLock(t)
	lock, err := parseSourceImportLock(raw, "outreach")
	if err != nil {
		t.Fatalf("parse cited-only Outreach source lock: %v", err)
	}
	called := false
	result, err := importSourceLockResult(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) {
		called = true
		return nil, nil
	}), defaultSourceImportLimits())
	if err != nil {
		t.Fatalf("import cited-only Outreach source lock: %v", err)
	}
	if called {
		t.Fatal("cited-only source reference attempted to fetch a provider document")
	}
	if result.DescriptorSchemaVersion != 3 || len(result.Operations) != 2 {
		t.Fatalf("source-reference descriptor = version %d operations %#v", result.DescriptorSchemaVersion, result.Operations)
	}
	main, custom := result.Operations[0], result.Operations[1]
	if main.SourceID != "outreach.rest.get.ApiV2Prospects" || main.Source.URL != outreachOpenAPIURL || main.Source.SHA256 != outreachOpenAPISHA256 || main.Source.Bytes != 1384297 || main.Source.Location != `.paths["/prospects"].get` {
		t.Fatalf("main source citation = %#v", main.Source)
	}
	if custom.SourceID != "outreach.rest.post.ApiV2CustomObjects" || custom.Source.URL != outreachCustomURL || custom.Source.SHA256 != outreachCustomSHA256 || custom.Source.Bytes != 422602 || custom.Source.Location != "Custom Objects via API: generic CRUD endpoint declaration" {
		t.Fatalf("supplement source citation = %#v", custom.Source)
	}
	for _, operation := range result.Operations {
		if !operation.Runtime.MergeBlocked || !sourceOperationHasFoundationGap(operation, "source_contract_unavailable") {
			t.Fatalf("source reference operation %q was not kept at source_contract_unavailable: %#v", operation.SourceID, operation.Runtime)
		}
		if len(operation.Request.Path) != 0 || len(operation.Request.Query) != 0 || len(operation.Request.Header) != 0 || operation.Request.Body != nil || len(operation.Responses) != 0 || operation.Output.Class != "" {
			t.Fatalf("source reference operation %q invented an execution contract: %#v", operation.SourceID, operation)
		}
	}
	encoded, err := marshalSourceImportResult(result)
	if err != nil {
		t.Fatalf("marshal cited-only Outreach descriptors: %v", err)
	}
	if !bytes.Contains(encoded, []byte(`"source_contract_unavailable"`)) || !bytes.Contains(encoded, []byte(outreachCustomURL)) {
		t.Fatalf("cited-only descriptor omitted gap or supplemental citation:\n%s", encoded)
	}
}

// TestSourceImportV3SourceReferenceUsesTheSameClosedProjection exercises the
// explicit schema-v3 form independently of Outreach so the reader is shared,
// not a connector-name exception.
func TestSourceImportV3SourceReferenceUsesTheSameClosedProjection(t *testing.T) {
	t.Parallel()
	raw := sourceImportV3SourceReferenceLock(t, "fixture", "fixture-source", "https://docs.polymetrics.invalid/fixture/openapi", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", 512, "GET", "/widgets")
	lock, err := parseSourceImportLock(raw, "fixture")
	if err != nil {
		t.Fatalf("parse v3 source reference: %v", err)
	}
	result, err := importSourceLockResult(context.Background(), lock, nil, defaultSourceImportLimits())
	if err != nil {
		t.Fatalf("import v3 source reference: %v", err)
	}
	if len(result.Operations) != 1 {
		t.Fatalf("v3 source reference operations = %#v", result.Operations)
	}
	operation := result.Operations[0]
	if operation.SourceID != "fixture.rest.fixture-source.get" || operation.Source.URL != "https://docs.polymetrics.invalid/fixture/openapi" || operation.Source.SHA256 != "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" || !sourceOperationHasFoundationGap(operation, "source_contract_unavailable") {
		t.Fatalf("v3 source-reference operation = %#v", operation)
	}
}

func TestSourceReferenceDigestIsProvenanceNotAnExecutionGate(t *testing.T) {
	t.Parallel()
	for _, digest := range []string{
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
	} {
		lock, err := parseSourceImportLock(sourceImportV3SourceReferenceLock(t, "fixture", "fixture-source", "https://docs.polymetrics.invalid/fixture/openapi", digest, 512, "GET", "/widgets"), "fixture")
		if err != nil {
			t.Fatalf("parse source reference digest %q: %v", digest, err)
		}
		result, err := importSourceLockResult(context.Background(), lock, nil, defaultSourceImportLimits())
		if err != nil {
			t.Fatalf("import source reference digest %q: %v", digest, err)
		}
		operation := result.Operations[0]
		if operation.Source.SHA256 != digest || !sourceOperationHasFoundationGap(operation, sourceContractUnavailableFoundation) {
			t.Fatalf("digest %q changed reference provenance/execution state: %#v", digest, operation)
		}
	}
}

func TestSourceImportSourceReferenceRejectsUnsupportedAndUnsafeKinds(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		mutate func(map[string]any)
		want   string
	}{
		{
			name: "unsupported legacy source kind",
			mutate: func(lock map[string]any) {
				lock["rest"].(map[string]any)["source_kind"] = "unbounded_generic_http"
			},
			want: "unsupported source-reference kind",
		},
		{
			name: "unsafe v3 citation URL",
			mutate: func(lock map[string]any) {
				document := lock["rest"].(map[string]any)["source_documents"].([]any)[0].(map[string]any)
				document["source_reference"].(map[string]any)["source_url"] = "https://docs.polymetrics.invalid/fixture/openapi?access_token=not-a-citation"
			},
			want: "source reference",
		},
		{
			name: "unsupported v3 document kind",
			mutate: func(lock map[string]any) {
				document := lock["rest"].(map[string]any)["source_documents"].([]any)[0].(map[string]any)
				document["kind"] = "generic_http"
			},
			want: "unsupported kind",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			var lock map[string]any
			if tt.name == "unsupported legacy source kind" {
				if err := json.Unmarshal(sourceImportOutreachReferenceLock(t), &lock); err != nil {
					t.Fatal(err)
				}
			} else if err := json.Unmarshal(sourceImportV3SourceReferenceLock(t, "fixture", "fixture-source", "https://docs.polymetrics.invalid/fixture/openapi", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", 512, "GET", "/widgets"), &lock); err != nil {
				t.Fatal(err)
			}
			tt.mutate(lock)
			raw, err := json.Marshal(lock)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := parseSourceImportLock(raw, firstNonEmpty(lock["connector"].(string))); err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("source reference error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestSourceImportReferenceFeedsOperationEvidenceWithoutEnabledClassification(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	lockPath := filepath.Join(root, "outreach-operation-source-lock.json")
	if err := os.WriteFile(lockPath, sourceImportOutreachReferenceLock(t), 0o644); err != nil {
		t.Fatalf("write Outreach reference lock: %v", err)
	}
	input, err := readOperationEvidenceSourceLock(lockPath, "outreach")
	if err != nil {
		t.Fatalf("read cited-only operation evidence lock: %v", err)
	}
	if len(input.Operations) != 2 || input.Operations[1].Trace.URL != outreachCustomURL || input.Operations[1].Trace.SHA256 != outreachCustomSHA256 {
		t.Fatalf("operation-evidence source citations = %#v", input.Operations)
	}
	commands := make([]engine.CLICommand, 0, len(operationEvidenceClasses))
	for _, class := range operationEvidenceClasses {
		commands = append(commands, engine.CLICommand{
			Path:         class + " widgets",
			Intent:       class,
			Availability: "implemented",
			APISurface:   []engine.CLISurfaceEndpointRef{{Method: "GET", Path: "/api/v2/prospects"}},
		})
	}
	bundle := engine.Bundle{Name: "outreach", CLISurface: &engine.CLISurface{Commands: commands}}
	row := projectOperationEvidenceRow(root, "outreach", input.Operations[0], bundle, nil, operationEvidenceWebsiteRow{}, conformance.Report{}, operationEvidenceCrosswalk{}, operationEvidenceDisposition{})
	if !row.hasGap("source_contract_unavailable") {
		t.Fatalf("source-reference operation evidence omitted source_contract_unavailable: %#v", row)
	}
	for _, class := range operationEvidenceClasses {
		if !row.Classifications[class].Declared || row.Classifications[class].Enabled {
			t.Fatalf("source-reference %s classification = %#v, want declared but not enabled", class, row.Classifications[class])
		}
	}
}

func TestRunSourceImportReferenceChecksWithoutRetainedArtifactOrSurfaceWrite(t *testing.T) {
	t.Parallel()
	defsDir := filepath.Join(t.TempDir(), "defs")
	sourcesDir := filepath.Join(defsDir, "outreach", "sources")
	if err := os.MkdirAll(sourcesDir, 0o755); err != nil {
		t.Fatalf("create source-reference fixture directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourcesDir, "outreach-operation-source-lock.json"), sourceImportOutreachReferenceLock(t), 0o644); err != nil {
		t.Fatalf("write source-reference fixture lock: %v", err)
	}
	output := filepath.Join(sourcesDir, "outreach-operation-descriptor.json")
	called := false
	fetcher := sourceImportFetchFunc(func(context.Context, string) ([]byte, error) {
		called = true
		return nil, nil
	})
	var stdout, stderr bytes.Buffer
	args := []string{"source-import", "outreach", "--defs", defsDir, "--out", output}
	if code := runSourceImportWithFetcher(args, &stdout, &stderr, fetcher); code != 0 {
		t.Fatalf("source-import reference exit=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
	if called {
		t.Fatal("source-import reference attempted to fetch a provider document")
	}
	if _, err := os.Stat(filepath.Join(sourcesDir, "artifacts")); !os.IsNotExist(err) {
		t.Fatalf("source-import reference created a retained-artifact directory: %v", err)
	}
	if got := stdout.String(); !strings.Contains(got, "2 operation(s)") || !strings.Contains(got, "writes=0 cli=0") {
		t.Fatalf("source-import reference output = %q", got)
	}
	stdout.Reset()
	stderr.Reset()
	args = append(args, "--check")
	if code := runSourceImportWithFetcher(args, &stdout, &stderr, fetcher); code != 0 {
		t.Fatalf("source-import reference check exit=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
	if called || !strings.Contains(stdout.String(), "2 operation(s)") {
		t.Fatalf("source-import reference check fetch/output = called=%t stdout=%q", called, stdout.String())
	}
}

func TestSourceReferenceProjectionDoesNotMaterializeAnExistingWriteOrCommand(t *testing.T) {
	t.Parallel()
	lock, err := parseSourceImportLock(sourceImportV3SourceReferenceLock(t, "fixture", "fixture-source", "https://docs.polymetrics.invalid/fixture/openapi", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", 512, "POST", "/widgets"), "fixture")
	if err != nil {
		t.Fatalf("parse source reference: %v", err)
	}
	result, err := importSourceLockResult(context.Background(), lock, nil, defaultSourceImportLimits())
	if err != nil {
		t.Fatalf("import source reference: %v", err)
	}
	bundleDir := t.TempDir()
	writesPath := filepath.Join(bundleDir, "writes.json")
	cliPath := filepath.Join(bundleDir, "cli_surface.json")
	writesBefore := []byte(`{"schema_version":1,"actions":[{"name":"widgets","kind":"custom","method":"POST","path":"/widgets","body_type":"json","body_fields":["name"],"record_schema":{"type":"object","additionalProperties":false,"properties":{"name":{"type":"string"}}},"risk":"standard"}]}`)
	cliBefore := []byte(`{"schema_version":1,"commands":[{"path":"widgets create","summary":"existing closed command","intent":"reverse_etl","availability":"implemented","write":"widgets","flags":[{"name":"name","type":"string","maps_to":"record.name"}]}]}`)
	if err := os.WriteFile(writesPath, writesBefore, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cliPath, cliBefore, 0o644); err != nil {
		t.Fatal(err)
	}
	stats, err := projectSourceDescriptorToBundle(bundleDir, result, false)
	if err != nil {
		t.Fatalf("project source reference: %v", err)
	}
	if stats != (sourceProjectionStats{}) {
		t.Fatalf("source-reference projection materialized a declaration: %+v", stats)
	}
	writesAfter, err := os.ReadFile(writesPath)
	if err != nil {
		t.Fatal(err)
	}
	cliAfter, err := os.ReadFile(cliPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(writesBefore, writesAfter) || !bytes.Equal(cliBefore, cliAfter) {
		t.Fatalf("source-reference projection changed declarations:\nwrites=%s\ncli=%s", writesAfter, cliAfter)
	}
}

func sourceImportOutreachReferenceLock(t *testing.T) []byte {
	t.Helper()
	lock := map[string]any{
		"schema_version": 2,
		"connector":      "outreach",
		"rest": map[string]any{
			"source_url":  outreachOpenAPIURL,
			"source_kind": "complete_machine_readable_specification_with_rendered_dynamic_supplement",
			"sha256":      outreachOpenAPISHA256,
			"bytes":       1384297,
			"openapi":     "3.0.3",
			"operation_counts": map[string]any{
				"GET":  1,
				"POST": 1,
			},
			"supplements": []any{map[string]any{
				"source_url":      outreachCustomURL,
				"source_location": "Custom Objects via API: generic CRUD endpoint declaration",
				"operation_count": 1,
				"sha256":          outreachCustomSHA256,
				"bytes":           422602,
			}},
			"operations": []any{
				map[string]any{
					"id":              "outreach.rest.get.ApiV2Prospects",
					"protocol":        "rest",
					"method":          "GET",
					"path":            "/api/v2/prospects",
					"source_location": `.paths["/prospects"].get`,
					"source_url":      outreachOpenAPIURL,
				},
				map[string]any{
					"id":              "outreach.rest.post.ApiV2CustomObjects",
					"protocol":        "rest",
					"method":          "POST",
					"path":            "/api/v2/customObjects/{objectName}",
					"source_location": "Custom Objects via API: generic CRUD endpoint declaration",
					"source_url":      outreachCustomURL,
				},
			},
		},
		"counts":           map[string]any{"rest": 2, "graphql_query": 0, "graphql_mutation": 0, "total": 2},
		"operations_found": map[string]any{"rest": 2, "graphql_query": 0, "graphql_mutation": 0, "total": 2},
		"coverage_confidence": map[string]any{
			"level": "complete_machine_readable_specification_with_rendered_dynamic_supplement",
			"basis": "The canonical OpenAPI citation is supplemented by the provider's documented dynamic Custom Objects routes.",
		},
	}
	raw, err := json.Marshal(lock)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func sourceImportV3SourceReferenceLock(t *testing.T, connector, id, sourceURL, digest string, size int64, method, path string) []byte {
	t.Helper()
	lock := map[string]any{
		"schema_version": 3,
		"connector":      connector,
		"rest": map[string]any{
			"retrieval": "declaration-only source-reference fixture",
			"coverage_confidence": map[string]any{
				"level": "source_reference",
				"basis": "The provider operation inventory and canonical citation are retained while the byte-backed contract is unavailable.",
			},
			"source_documents": []any{map[string]any{
				"id":   id,
				"kind": "source_reference",
				"source_reference": map[string]any{
					"source_url": sourceURL,
					"sha256":     digest,
					"bytes":      size,
					"openapi":    "3.0.3",
				},
				"operations": []any{map[string]any{
					"id":              connector + ".rest." + id + "." + strings.ToLower(method),
					"protocol":        "rest",
					"method":          method,
					"path":            path,
					"operation_id":    id + "-operation",
					"source_location": `paths["` + path + `"].` + strings.ToLower(method),
				}},
			}},
		},
		"counts": map[string]any{"rest": 1, "graphql_query": 0, "graphql_mutation": 0, "total": 1},
	}
	raw, err := json.Marshal(lock)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}
