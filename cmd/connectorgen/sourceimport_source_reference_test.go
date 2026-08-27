package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors/conformance"
	"polymetrics.ai/internal/connectors/engine"
)

const (
	outreachOpenAPIURL          = "https://api.outreach.io/api/v2/schema/openapi.json"
	outreachOpenAPISHA256       = "d1f697f6558fda68cd6d8059044e323c20849aeebf303e15c43e0eb9875e2ef6"
	outreachCustomURL           = "https://developers.outreach.io/api/custom-objects"
	outreachCustomSHA256        = "2e74714a933b74cb9a83ddbdb18eeb0b9d045115102ed7465021a45db19e3dda"
	outreachReferenceLockSHA256 = "f733248bfd484625b8f2bae3490b3211f7e158ab375d3c8de5ede83b1f369f89"
)

// outreachReferenceLockCompressed is gzip+base64 of the exact 100124-byte
// Outreach lock from candidate commit
// 18248d233e6abd9d7ec03075a225cf35ee2f5399. The decoded fixture digest is
// asserted below so this proof cannot silently become a hand-reduced inventory.
const outreachReferenceLockCompressed = `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`

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
	if result.DescriptorSchemaVersion != 3 || len(result.Operations) != 259 {
		t.Fatalf("source-reference descriptor = version %d operations %#v", result.DescriptorSchemaVersion, result.Operations)
	}
	operations := make(map[string]sourceOperationDescriptor, len(result.Operations))
	sourceCounts := map[string]int{}
	for _, operation := range result.Operations {
		operations[operation.SourceID] = operation
		sourceCounts[operation.Source.URL]++
	}
	if sourceCounts[outreachOpenAPIURL] != 253 || sourceCounts[outreachCustomURL] != 6 {
		t.Fatalf("Outreach source inventory counts = %#v, want primary=253 supplemental=6", sourceCounts)
	}
	main := operations["outreach.rest.get.ApiV2Prospects93"]
	custom := operations["outreach.rest.post.ApiV2CustomObjectsObjectName167"]
	if main.SourceID != "outreach.rest.get.ApiV2Prospects93" || main.Source.URL != outreachOpenAPIURL || main.Source.SHA256 != outreachOpenAPISHA256 || main.Source.Bytes != 1384297 || main.Source.Location != `.paths["/prospects"].get` {
		t.Fatalf("main source citation = %#v", main.Source)
	}
	if custom.SourceID != "outreach.rest.post.ApiV2CustomObjectsObjectName167" || custom.Source.URL != outreachCustomURL || custom.Source.SHA256 != outreachCustomSHA256 || custom.Source.Bytes != 422602 || custom.Source.Location != "Custom Objects via API: generic CRUD endpoint declaration" {
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

// TestSourceImportReferenceEncodesEveryDeclaredLaneWithoutPromotion isolates
// the six-lane JSON encoding. The commands are purpose-built declarations for
// this encoding test; the exact retained-Outreach proof above exercises the
// real 259-row inventory and current bundle separately.
func TestSourceImportReferenceEncodesEveryDeclaredLaneWithoutPromotion(t *testing.T) {
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
	if len(input.Operations) != 259 {
		t.Fatalf("operation-evidence source citation count = %d, want 259", len(input.Operations))
	}
	operations := make(map[string]operationEvidenceSourceOperation, len(input.Operations))
	for _, operation := range input.Operations {
		operations[operation.ID] = operation
	}
	prospects, ok := operations["outreach.rest.get.ApiV2Prospects93"]
	if !ok || prospects.Trace.URL != outreachOpenAPIURL || prospects.Trace.SHA256 != outreachOpenAPISHA256 {
		t.Fatalf("prospects source citation = %#v", prospects)
	}
	custom, ok := operations["outreach.rest.post.ApiV2CustomObjectsObjectName167"]
	if !ok || custom.Trace.URL != outreachCustomURL || custom.Trace.SHA256 != outreachCustomSHA256 {
		t.Fatalf("custom-object source citation = %#v", custom)
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
	row := projectOperationEvidenceRow(root, "outreach", prospects, bundle, nil, operationEvidenceWebsiteRow{}, conformance.Report{}, operationEvidenceCrosswalk{}, operationEvidenceDisposition{})
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
	if got := stdout.String(); !strings.Contains(got, "259 operation(s)") || !strings.Contains(got, "writes=0 cli=0") {
		t.Fatalf("source-import reference output = %q", got)
	}
	stdout.Reset()
	stderr.Reset()
	args = append(args, "--check")
	if code := runSourceImportWithFetcher(args, &stdout, &stderr, fetcher); code != 0 {
		t.Fatalf("source-import reference check exit=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
	if called || !strings.Contains(stdout.String(), "259 operation(s)") {
		t.Fatalf("source-import reference check fetch/output = called=%t stdout=%q", called, stdout.String())
	}
}

// TestExactOutreachReferenceProjectsAgainstCurrentMainBundle keeps the real
// retained 259-row inventory separate from the synthetic six-lane encoding
// unit test. It uses the checked-in Outreach bundle as the canonical mapping
// target, confirms the cited prospect remains visible in every evidence lane,
// and proves source-reference projection cannot rewrite the existing bundle.
func TestExactOutreachReferenceProjectsAgainstCurrentMainBundle(t *testing.T) {
	// Do not parallelize this proof: it reads the checked-in bundle whose bytes
	// are intentionally compared before and after projection.
	raw := sourceImportOutreachReferenceLock(t)
	lock, err := parseSourceImportLock(raw, "outreach")
	if err != nil {
		t.Fatalf("parse exact Outreach source lock: %v", err)
	}
	result, err := importSourceLockResult(context.Background(), lock, nil, defaultSourceImportLimits())
	if err != nil {
		t.Fatalf("import exact Outreach source lock: %v", err)
	}
	if len(result.Operations) != 259 {
		t.Fatalf("exact Outreach descriptor count = %d, want 259", len(result.Operations))
	}
	bundleDir := filepath.Join("..", "..", "internal", "connectors", "defs", "outreach")
	tracked := []string{"api_surface.json", "writes.json"}
	before := make(map[string][]byte, len(tracked))
	for _, name := range tracked {
		bytes, err := os.ReadFile(filepath.Join(bundleDir, name))
		if err != nil {
			t.Fatalf("read current Outreach %s: %v", name, err)
		}
		before[name] = bytes
	}
	for _, check := range []bool{false, true} {
		stats, err := projectSourceDescriptorToBundle(bundleDir, result, check)
		if err != nil {
			t.Fatalf("project exact Outreach source reference check=%t: %v", check, err)
		}
		if stats != (sourceProjectionStats{}) {
			t.Fatalf("exact Outreach source reference changed current bundle check=%t: %+v", check, stats)
		}
	}
	for name, want := range before {
		got, err := os.ReadFile(filepath.Join(bundleDir, name))
		if err != nil {
			t.Fatalf("read current Outreach %s after projection: %v", name, err)
		}
		if !bytes.Equal(got, want) {
			t.Fatalf("exact Outreach source reference rewrote current %s", name)
		}
	}

	lockPath := filepath.Join(t.TempDir(), "outreach-operation-source-lock.json")
	if err := os.WriteFile(lockPath, raw, 0o644); err != nil {
		t.Fatalf("write exact Outreach evidence lock: %v", err)
	}
	input, err := readOperationEvidenceSourceLock(lockPath, "outreach")
	if err != nil {
		t.Fatalf("read exact Outreach operation evidence: %v", err)
	}
	if len(input.Operations) != 259 {
		t.Fatalf("exact Outreach evidence rows = %d, want 259", len(input.Operations))
	}
	var prospects operationEvidenceSourceOperation
	for _, operation := range input.Operations {
		if operation.ID == "outreach.rest.get.ApiV2Prospects93" {
			prospects = operation
			break
		}
	}
	if prospects.ID == "" || prospects.Trace.URL != outreachOpenAPIURL || prospects.Trace.SHA256 != outreachOpenAPISHA256 || prospects.Trace.Location != `.paths["/prospects"].get` {
		t.Fatalf("exact Outreach prospects evidence trace = %#v", prospects)
	}
	bundle, err := engine.Load(os.DirFS(filepath.Join("..", "..", "internal", "connectors", "defs")), "outreach")
	if err != nil {
		t.Fatalf("load current Outreach bundle: %v", err)
	}
	row := projectOperationEvidenceRow(filepath.Join("..", ".."), "outreach", prospects, bundle, nil, operationEvidenceWebsiteRow{}, conformance.Report{}, operationEvidenceCrosswalk{}, operationEvidenceDisposition{})
	if row.Canonical.State != "mapped" || row.Canonical.Method != "GET" || row.Canonical.Path != "/api/v2/prospects" {
		t.Fatalf("exact Outreach canonical mapping = %#v", row.Canonical)
	}
	if !row.hasGap(sourceContractUnavailableFoundation) || row.Runtime.Enabled {
		t.Fatalf("exact Outreach source-reference runtime disposition = %#v", row)
	}
	for _, class := range operationEvidenceClasses {
		if row.Classifications[class].Enabled {
			t.Fatalf("exact Outreach %s lane was falsely enabled: %#v", class, row.Classifications[class])
		}
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

func TestSourceReferenceProjectionDoesNotDowngradeExistingDirectReadSurface(t *testing.T) {
	t.Parallel()
	lock, err := parseSourceImportLock(sourceImportV3SourceReferenceLock(t, "outreach", "prospects-source", outreachOpenAPIURL, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", 512, "GET", "/api/v2/prospects"), "outreach")
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
	apiPath := filepath.Join(bundleDir, "api_surface.json")
	writesBefore := []byte(`{"actions":[]}`)
	cliBefore := []byte(`{"commands":[{"path":"prospects list","summary":"existing direct read","intent":"direct_read","availability":"implemented","api_surface":[{"method":"GET","path":"/api/v2/prospects"}]}]}`)
	apiBefore := []byte(`{"api":"Outreach REST API v2","docs":"https://developers.outreach.io/api","endpoints":[{"method":"GET","path":"/api/v2/prospects","covered_by":{"stream":"prospects","direct_read":"prospects list"}}]}`)
	for path, raw := range map[string][]byte{writesPath: writesBefore, cliPath: cliBefore, apiPath: apiBefore} {
		if err := os.WriteFile(path, raw, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, check := range []bool{false, true} {
		stats, err := projectSourceDescriptorToBundle(bundleDir, result, check)
		if err != nil {
			t.Fatalf("project source reference check=%t: %v", check, err)
		}
		if stats != (sourceProjectionStats{}) {
			t.Fatalf("source-reference GET changed declaration check=%t: %+v", check, stats)
		}
		for path, want := range map[string][]byte{writesPath: writesBefore, cliPath: cliBefore, apiPath: apiBefore} {
			got, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, want) {
				t.Fatalf("source-reference GET changed %s check=%t:\n%s", filepath.Base(path), check, got)
			}
		}
	}
}

func TestSourceImportLegacyByteBackedLocksRejectReferenceOnlyFields(t *testing.T) {
	t.Parallel()
	fields := []struct {
		name   string
		mutate func(map[string]any)
	}{
		{"root operations_found", func(lock map[string]any) {
			lock["operations_found"] = map[string]any{"rest": 2, "graphql_query": 0, "graphql_mutation": 0, "total": 2}
		}},
		{"root coverage_confidence", func(lock map[string]any) {
			lock["coverage_confidence"] = map[string]any{"level": "source_reference", "basis": "reference only"}
		}},
		{"rest source_kind", func(lock map[string]any) {
			lock["rest"].(map[string]any)["source_kind"] = sourceImportLegacySourceReferenceKind
		}},
		{"rest operation_counts", func(lock map[string]any) {
			lock["rest"].(map[string]any)["operation_counts"] = map[string]any{"GET": 2}
		}},
		{"rest supplements", func(lock map[string]any) { lock["rest"].(map[string]any)["supplements"] = []any{} }},
		{"operation source_url", func(lock map[string]any) {
			lock["rest"].(map[string]any)["operations"].([]any)[0].(map[string]any)["source_url"] = "https://provider.example.invalid/reference"
		}},
	}
	for _, version := range []int{1, 2} {
		for _, field := range fields {
			version, field := version, field
			t.Run("v"+strconv.Itoa(version)+"/"+field.name, func(t *testing.T) {
				lock := sourceImportLegacyByteBackedWireFixture(version)
				field.mutate(lock)
				raw, err := json.Marshal(lock)
				if err != nil {
					t.Fatal(err)
				}
				_, err = parseSourceImportLock(raw, "alpha")
				if err == nil {
					t.Fatalf("byte-backed v%d %s was accepted", version, field.name)
				}
				if version == 2 && field.name == "rest source_kind" {
					if !strings.Contains(err.Error(), "source-reference") {
						t.Fatalf("schema-v2 source-kind discriminator error = %v, want closed source-reference contract", err)
					}
					return
				}
				if !strings.Contains(err.Error(), "unknown field") {
					t.Fatalf("byte-backed v%d %s error = %v, want closed unknown-field rejection", version, field.name, err)
				}
			})
		}
	}
}

func sourceImportLegacyByteBackedWireFixture(version int) map[string]any {
	return map[string]any{
		"schema_version": version,
		"connector":      "alpha",
		"rest": map[string]any{
			"source_url": "https://fixtures.polymetrics.invalid/alpha-openapi.json",
			"sha256":     strings.Repeat("a", 64),
			"bytes":      1,
			"openapi":    "3.0.3",
			"operations": []any{map[string]any{
				"id":              "alpha.rest.get.widgets",
				"protocol":        "rest",
				"method":          "GET",
				"path":            "/widgets",
				"source_location": `paths["/widgets"].get`,
			}},
		},
		"counts": map[string]any{"rest": 1, "graphql_query": 0, "graphql_mutation": 0, "total": 1},
	}
}

func TestSourceImportV3SourceReferenceRejectsClosedOperationIdentityViolations(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		mutate func(map[string]any)
	}{
		{"protocol", func(operation map[string]any) { operation["protocol"] = "graphql" }},
		{"lowercase method", func(operation map[string]any) { operation["method"] = "get" }},
		{"unsupported method", func(operation map[string]any) { operation["method"] = "TRACE" }},
		{"untrimmed id", func(operation map[string]any) { operation["id"] = " fixture.rest.widgets.get" }},
		{"control location", func(operation map[string]any) { operation["source_location"] = "paths[\"/widgets\"].get\x1b" }},
		{"untrimmed provider operation id", func(operation map[string]any) { operation["operation_id"] = " widgets-get" }},
		{"per-operation citation URL", func(operation map[string]any) {
			operation["citation_url"] = "https://docs.polymetrics.invalid/fixture/operation"
		}},
		{"per-operation citation binding", func(operation map[string]any) {
			operation["citation_binding"] = map[string]any{}
		}},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			var lock map[string]any
			if err := json.Unmarshal(sourceImportV3SourceReferenceLock(t, "fixture", "fixture-source", "https://docs.polymetrics.invalid/fixture/openapi", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", 512, "GET", "/widgets"), &lock); err != nil {
				t.Fatal(err)
			}
			operation := lock["rest"].(map[string]any)["source_documents"].([]any)[0].(map[string]any)["operations"].([]any)[0].(map[string]any)
			tt.mutate(operation)
			raw, err := json.Marshal(lock)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := parseSourceImportLock(raw, "fixture"); err == nil {
				t.Fatalf("v3 source-reference operation %s was accepted", tt.name)
			}
		})
	}
}

func TestSourceImportV3SourceReferenceRejectsDuplicateOperationIdentityAndRoute(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name   string
		want   string
		mutate func(map[string]any)
	}{
		{
			name: "identity",
			want: "duplicates",
			mutate: func(duplicate map[string]any) {
				duplicate["path"] = "/other-widgets"
			},
		},
		{
			name: "route",
			want: "route GET /widgets occurs",
			mutate: func(duplicate map[string]any) {
				duplicate["id"] = "fixture.rest.widgets-alias.get"
			},
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			var lock map[string]any
			if err := json.Unmarshal(sourceImportV3SourceReferenceLock(t, "fixture", "fixture-source", "https://docs.polymetrics.invalid/fixture/openapi", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", 512, "GET", "/widgets"), &lock); err != nil {
				t.Fatal(err)
			}
			document := lock["rest"].(map[string]any)["source_documents"].([]any)[0].(map[string]any)
			operations := document["operations"].([]any)
			duplicate := make(map[string]any, len(operations[0].(map[string]any)))
			for key, value := range operations[0].(map[string]any) {
				duplicate[key] = value
			}
			tt.mutate(duplicate)
			document["operations"] = append(operations, duplicate)
			raw, err := json.Marshal(lock)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := parseSourceImportLock(raw, "fixture"); err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("v3 source-reference duplicate %s error = %v", tt.name, err)
			}
		})
	}
}

func sourceImportOutreachReferenceLock(t *testing.T) []byte {
	t.Helper()
	compressed, err := base64.StdEncoding.DecodeString(outreachReferenceLockCompressed)
	if err != nil {
		t.Fatalf("decode exact Outreach source-lock fixture: %v", err)
	}
	reader, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		t.Fatalf("open exact Outreach source-lock fixture: %v", err)
	}
	raw, err := io.ReadAll(reader)
	if closeErr := reader.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		t.Fatalf("read exact Outreach source-lock fixture: %v", err)
	}
	if got := sha256.Sum256(raw); hex.EncodeToString(got[:]) != outreachReferenceLockSHA256 {
		t.Fatalf("exact Outreach source-lock fixture digest = %x, want %s", got, outreachReferenceLockSHA256)
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
