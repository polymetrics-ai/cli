package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"
	"testing/fstest"

	"polymetrics.ai/internal/connectors/engine"
)

const (
	renderedReferenceArtifactURL  = "https://fixtures.polymetrics.invalid/reference/widgets.html"
	renderedReferencePublishedURL = "https://docs.polymetrics.invalid/reference/widgets"
	renderedReferenceCitationURL  = "https://docs.polymetrics.invalid/reference/widgets#list-widgets"
)

func sourceImportRenderedReferenceDocument(t *testing.T, citationURL string) (map[string]any, []byte) {
	t.Helper()
	page := []byte(`{"operation":"GET /widgets","section":"list-widgets"}`)
	digest := sha256.Sum256(page)
	return map[string]any{
		"id":           "reference",
		"kind":         "rendered_reference",
		"content_type": "application/json",
		"artifact": map[string]any{
			"source_url": renderedReferenceArtifactURL,
			"sha256":     hex.EncodeToString(digest[:]),
			"bytes":      len(page),
		},
		"published_source": map[string]any{
			"source_url":  renderedReferencePublishedURL,
			"capture_url": renderedReferenceArtifactURL,
			"sha256":      hex.EncodeToString(digest[:]),
			"bytes":       len(page),
			"adapter":     "fixture-rendered-reference-capture-v1",
		},
		"operations": []any{map[string]any{
			"id":              "fixture.rest.reference.list-widgets",
			"protocol":        "rest",
			"method":          "GET",
			"path":            "/widgets",
			"operation_id":    "list-widgets",
			"source_location": "#list-widgets",
			"citation_url":    citationURL,
		}},
	}, page
}

func sourceImportV3RenderedReferenceLock(t *testing.T, citationURL string) ([]byte, []byte) {
	t.Helper()
	document, page := sourceImportRenderedReferenceDocument(t, citationURL)
	lock := map[string]any{
		"schema_version": 3,
		"connector":      "fixture",
		"rest": map[string]any{
			"retrieval": "hermetic rendered-reference fixture capture",
			"coverage_confidence": map[string]any{
				"level": "documented",
				"basis": "The provider reference publishes this endpoint as one captured JSON document.",
			},
			"source_documents": []any{document},
		},
		"counts": map[string]any{"rest": 1, "graphql_query": 0, "graphql_mutation": 0, "total": 1},
	}
	raw, err := json.Marshal(lock)
	if err != nil {
		t.Fatal(err)
	}
	return raw, page
}

func sourceImportV3BundleLock(t *testing.T) ([]byte, []byte) {
	t.Helper()
	archive := []byte("fixture archive bytes")
	return sourceImportV3BundleLockWithCapture(t, "application/zip", archive)
}

func sourceImportV3BundleLockWithCapture(t *testing.T, contentType string, archive []byte) ([]byte, []byte) {
	t.Helper()
	digest := sha256.Sum256(archive)
	lock := map[string]any{
		"schema_version": 3,
		"connector":      "fixture",
		"rest": map[string]any{
			"retrieval": "hermetic bundle fixture capture",
			"coverage_confidence": map[string]any{
				"level": "machine-readable",
				"basis": "The archive is retained as the immutable evidence for the listed API operations.",
			},
			"source_documents": []any{map[string]any{
				"id":           "bundle",
				"kind":         "bundle",
				"content_type": contentType,
				"artifact": map[string]any{
					"source_url": "https://fixtures.polymetrics.invalid/reference/api.zip",
					"sha256":     hex.EncodeToString(digest[:]),
					"bytes":      len(archive),
				},
				"published_source": map[string]any{
					"source_url":  "https://docs.polymetrics.invalid/reference/api-archive",
					"capture_url": "https://fixtures.polymetrics.invalid/reference/api.zip",
					"sha256":      hex.EncodeToString(digest[:]),
					"bytes":       len(archive),
					"adapter":     "fixture-bundle-capture-v1",
				},
				"operations": []any{map[string]any{
					"id":              "fixture.rest.bundle.list-widgets",
					"protocol":        "rest",
					"method":          "GET",
					"path":            "/widgets",
					"operation_id":    "list-widgets",
					"source_location": "widgets.openapi.json#/paths/~1widgets/get",
				}},
			}},
		},
		"counts": map[string]any{"rest": 1, "graphql_query": 0, "graphql_mutation": 0, "total": 1},
	}
	raw, err := json.Marshal(lock)
	if err != nil {
		t.Fatal(err)
	}
	return raw, archive
}

func sourceImportV3UnavailableLock(t *testing.T) []byte {
	t.Helper()
	evidence := []byte(`{"status":"no-public-api-reference"}`)
	digest := sha256.Sum256(evidence)
	lock := map[string]any{
		"schema_version": 3,
		"connector":      "fixture",
		"rest": map[string]any{
			"retrieval": "hermetic unavailable-source fixture capture",
			"coverage_confidence": map[string]any{
				"level": "unavailable",
				"basis": "The provider has no usable API reference for this connector.",
			},
			"source_documents": []any{map[string]any{
				"id":                 "unavailable",
				"kind":               "unavailable",
				"content_type":       "application/json",
				"unavailable_reason": "provider has no usable API reference",
				"artifact": map[string]any{
					"source_url": "https://fixtures.polymetrics.invalid/reference/unavailable.json",
					"sha256":     hex.EncodeToString(digest[:]),
					"bytes":      len(evidence),
				},
				"published_source": map[string]any{
					"source_url":  "https://docs.polymetrics.invalid/reference/availability",
					"capture_url": "https://fixtures.polymetrics.invalid/reference/unavailable.json",
					"sha256":      hex.EncodeToString(digest[:]),
					"bytes":       len(evidence),
					"adapter":     "fixture-unavailable-capture-v1",
				},
				"operations": []any{},
			}},
		},
		"counts": map[string]any{"rest": 0, "graphql_query": 0, "graphql_mutation": 0, "total": 0},
	}
	raw, err := json.Marshal(lock)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestSourceImportVersion3RenderedReferenceProjectsCapturedCitation(t *testing.T) {
	raw, page := sourceImportV3RenderedReferenceLock(t, renderedReferenceCitationURL)
	lock, err := parseSourceImportLock(raw, "fixture")
	if err != nil {
		t.Fatalf("parse rendered-reference lock: %v", err)
	}
	requested := map[string]int{}
	result, err := importSourceLockResult(context.Background(), lock, sourceImportFetchFunc(func(_ context.Context, sourceURL string) ([]byte, error) {
		requested[sourceURL]++
		if sourceURL != renderedReferenceArtifactURL {
			t.Fatalf("rendered-reference import fetched provenance URL %q", sourceURL)
		}
		return page, nil
	}), defaultSourceImportLimits())
	if err != nil {
		t.Fatalf("import rendered-reference lock: %v", err)
	}
	if len(result.Operations) != 1 || result.Operations[0].SourceID != "fixture.rest.reference.list-widgets" || result.Operations[0].Method != "get" || result.Operations[0].Path != "/widgets" {
		t.Fatalf("rendered-reference operation = %#v", result.Operations)
	}
	if requested[renderedReferencePublishedURL] != 0 || requested[renderedReferenceCitationURL] != 0 {
		t.Fatalf("rendered-reference provenance was fetched: %#v", requested)
	}
	projected, err := json.Marshal(result.Operations[0])
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(projected), `"citation_url":"`+renderedReferenceCitationURL+`"`) {
		t.Fatalf("rendered-reference citation was omitted from projection: %s", projected)
	}
	if findings := validateSourceDescriptorAgainstLock("fixture", "sources/fixture-operation-descriptor.json", lock, sourceImportDescriptorDocument{SchemaVersion: 3, Operations: result.Operations}); len(findings) != 0 {
		t.Fatalf("rendered-reference source projection findings = %+v", findings)
	}
}

func TestSourceImportVersion3SwaggerTwoProjectsWithoutOpenAPIVersionInventory(t *testing.T) {
	swagger := []byte(`{"swagger":"2.0","info":{"title":"fixture","version":"1"},"paths":{"/widgets":{"get":{"operationId":"listWidgets","responses":{"200":{"description":"ok"}}}}}}`)
	digest := sha256.Sum256(swagger)
	lockRaw, err := json.Marshal(map[string]any{
		"schema_version": 3,
		"connector":      "fixture",
		"rest": map[string]any{
			"retrieval": "hermetic Swagger 2.0 fixture capture",
			"openapi":   []any{},
			"source_documents": []any{map[string]any{
				"id": "swagger",
				"artifact": map[string]any{
					"source_url": "https://fixtures.polymetrics.invalid/reference/swagger.json",
					"sha256":     hex.EncodeToString(digest[:]),
					"bytes":      len(swagger),
					"swagger":    "2.0",
				},
				"published_source": map[string]any{
					"source_url":  "https://docs.polymetrics.invalid/reference/swagger",
					"capture_url": "https://fixtures.polymetrics.invalid/reference/swagger.json",
					"sha256":      hex.EncodeToString(digest[:]),
					"bytes":       len(swagger),
					"adapter":     "fixture-swagger-capture-v1",
				},
				"operations": []any{map[string]any{
					"id":              "fixture.rest.swagger.list-widgets",
					"protocol":        "rest",
					"method":          "GET",
					"path":            "/widgets",
					"operation_id":    "listWidgets",
					"source_location": `paths["/widgets"].get`,
				}},
			}},
		},
		"counts": map[string]any{"rest": 1, "graphql_query": 0, "graphql_mutation": 0, "total": 1},
	})
	if err != nil {
		t.Fatal(err)
	}
	lock, err := parseSourceImportLock(lockRaw, "fixture")
	if err != nil {
		t.Fatalf("parse Swagger 2.0 v3 lock: %v", err)
	}
	result, err := importSourceLockResult(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) {
		return swagger, nil
	}), defaultSourceImportLimits())
	if err != nil {
		t.Fatalf("import Swagger 2.0 v3 lock: %v", err)
	}
	if len(result.Operations) != 1 || result.Operations[0].Source.Form != "swagger" || result.Operations[0].Source.Version != "2.0" {
		t.Fatalf("Swagger 2.0 operation projection = %#v", result.Operations)
	}
	if findings := validateSourceDescriptorAgainstLock("fixture", "sources/fixture-operation-descriptor.json", lock, sourceImportDescriptorDocument{SchemaVersion: 3, Operations: result.Operations}); len(findings) != 0 {
		t.Fatalf("Swagger 2.0 source projection findings = %+v", findings)
	}
}

func TestSourceImportVersion3RenderedReferenceRetainsCapturedEvidenceWithoutOperations(t *testing.T) {
	raw, page := sourceImportV3RenderedReferenceLock(t, renderedReferenceCitationURL)
	var wire map[string]any
	if err := json.Unmarshal(raw, &wire); err != nil {
		t.Fatal(err)
	}
	sidecar := []byte(`{"navigation":"reference index"}`)
	sidecarDigest := sha256.Sum256(sidecar)
	wire["rest"].(map[string]any)["source_documents"] = append(wire["rest"].(map[string]any)["source_documents"].([]any), map[string]any{
		"id":           "sidecar",
		"kind":         "rendered_reference",
		"content_type": "application/json",
		"artifact": map[string]any{
			"source_url": "https://fixtures.polymetrics.invalid/reference/navigation.json",
			"sha256":     hex.EncodeToString(sidecarDigest[:]),
			"bytes":      len(sidecar),
		},
		"published_source": map[string]any{
			"source_url":  "https://docs.polymetrics.invalid/reference",
			"capture_url": "https://fixtures.polymetrics.invalid/reference/navigation.json",
			"sha256":      hex.EncodeToString(sidecarDigest[:]),
			"bytes":       len(sidecar),
			"adapter":     "fixture-rendered-reference-capture-v1",
		},
		"operations": []any{},
	})
	raw, err := json.Marshal(wire)
	if err != nil {
		t.Fatal(err)
	}
	lock, err := parseSourceImportLock(raw, "fixture")
	if err != nil {
		t.Fatalf("parse rendered-reference lock with evidence-only sidecar: %v", err)
	}
	result, err := importSourceLockResult(context.Background(), lock, sourceImportFetchFunc(func(_ context.Context, sourceURL string) ([]byte, error) {
		switch sourceURL {
		case renderedReferenceArtifactURL:
			return page, nil
		case "https://fixtures.polymetrics.invalid/reference/navigation.json":
			return sidecar, nil
		default:
			t.Fatalf("unexpected source fetch %q", sourceURL)
			return nil, nil
		}
	}), defaultSourceImportLimits())
	if err != nil {
		t.Fatalf("import rendered-reference lock with evidence-only sidecar: %v", err)
	}
	if len(result.Operations) != 1 || result.Operations[0].SourceID != "fixture.rest.reference.list-widgets" {
		t.Fatalf("rendered-reference sidecar changed operation projection: %#v", result.Operations)
	}
	if findings := validateSourceDescriptorAgainstLock("fixture", "sources/fixture-operation-descriptor.json", lock, sourceImportDescriptorDocument{SchemaVersion: 3, Operations: result.Operations}); len(findings) != 0 {
		t.Fatalf("rendered-reference sidecar source projection findings = %+v", findings)
	}
}

func TestSourceImportVersion3RenderedReferenceProjectsYAMLPathFragment(t *testing.T) {
	fragment := []byte("get:\n  operationId: listAudit\n  responses:\n    '200':\n      description: listed audit records\n")
	digest := sha256.Sum256(fragment)
	document := map[string]any{
		"id":           "audit-path-fragment",
		"kind":         "rendered_reference",
		"content_type": "application/yaml",
		"artifact": map[string]any{
			"source_url": "https://fixtures.polymetrics.invalid/reference/audit.yml",
			"sha256":     hex.EncodeToString(digest[:]),
			"bytes":      len(fragment),
		},
		"published_source": map[string]any{
			"source_url":  "https://docs.polymetrics.invalid/public-api/audit",
			"capture_url": "https://fixtures.polymetrics.invalid/reference/audit.yml",
			"sha256":      hex.EncodeToString(digest[:]),
			"bytes":       len(fragment),
			"adapter":     "fixture-yaml-path-fragment-capture-v1",
		},
		"operations": []any{map[string]any{
			"id":              "fixture.rest.reference.list-audit",
			"protocol":        "rest",
			"method":          "GET",
			"path":            "/audit",
			"operation_id":    "listAudit",
			"source_location": "#/get",
			"citation_url":    "https://docs.polymetrics.invalid/public-api/audit#list-audit",
		}},
	}
	raw, err := json.Marshal(map[string]any{
		"schema_version": 3,
		"connector":      "fixture",
		"rest": map[string]any{
			"retrieval": "hermetic rendered-reference YAML fragment fixture capture",
			"coverage_confidence": map[string]any{
				"level": "documented",
				"basis": "The captured YAML document is an OpenAPI path fragment, not a standalone OpenAPI description.",
			},
			"source_documents": []any{document},
		},
		"counts": map[string]any{"rest": 1, "graphql_query": 0, "graphql_mutation": 0, "total": 1},
	})
	if err != nil {
		t.Fatal(err)
	}
	lock, err := parseSourceImportLock(raw, "fixture")
	if err != nil {
		t.Fatalf("parse YAML path-fragment lock: %v", err)
	}
	result, err := importSourceLockResult(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) {
		return fragment, nil
	}), defaultSourceImportLimits())
	if err != nil {
		t.Fatalf("import YAML path-fragment lock: %v", err)
	}
	if len(result.Operations) != 1 || result.Operations[0].Source.Form != "rendered_reference" || result.Operations[0].Source.CitationURL != "https://docs.polymetrics.invalid/public-api/audit#list-audit" {
		t.Fatalf("YAML path-fragment operation projection = %#v", result.Operations)
	}
}

func TestSourceImportVersion3RenderedReferenceRejectsStandaloneOpenAPIDescription(t *testing.T) {
	standalone := []byte(`{"openapi":"3.0.3","info":{"title":"fixture","version":"1"},"paths":{"/audit":{"get":{"operationId":"listAudit","responses":{"200":{"description":"listed audit records"}}}}}}`)
	digest := sha256.Sum256(standalone)
	lockRaw, err := json.Marshal(map[string]any{
		"schema_version": 3,
		"connector":      "fixture",
		"rest": map[string]any{
			"retrieval": "standalone OpenAPI fixture incorrectly declared as rendered reference",
			"coverage_confidence": map[string]any{
				"level": "documented",
				"basis": "The source declaration must not use rendered_reference for a standalone OpenAPI document.",
			},
			"source_documents": []any{map[string]any{
				"id":           "standalone",
				"kind":         "rendered_reference",
				"content_type": "application/json",
				"artifact": map[string]any{
					"source_url": "https://fixtures.polymetrics.invalid/reference/standalone.json",
					"sha256":     hex.EncodeToString(digest[:]),
					"bytes":      len(standalone),
				},
				"published_source": map[string]any{
					"source_url":  "https://docs.polymetrics.invalid/public-api/audit",
					"capture_url": "https://fixtures.polymetrics.invalid/reference/standalone.json",
					"sha256":      hex.EncodeToString(digest[:]),
					"bytes":       len(standalone),
					"adapter":     "fixture-standalone-openapi-capture-v1",
				},
				"operations": []any{map[string]any{
					"id":              "fixture.rest.reference.list-audit",
					"protocol":        "rest",
					"method":          "GET",
					"path":            "/audit",
					"operation_id":    "listAudit",
					"source_location": "#/paths/~1audit/get",
					"citation_url":    "https://docs.polymetrics.invalid/public-api/audit#list-audit",
				}},
			}},
		},
		"counts": map[string]any{"rest": 1, "graphql_query": 0, "graphql_mutation": 0, "total": 1},
	})
	if err != nil {
		t.Fatal(err)
	}
	lock, err := parseSourceImportLock(lockRaw, "fixture")
	if err != nil {
		t.Fatalf("parse incorrectly typed standalone OpenAPI lock: %v", err)
	}
	_, err = importSourceLockResult(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) {
		return standalone, nil
	}), defaultSourceImportLimits())
	if err == nil || !strings.Contains(err.Error(), "must use kind openapi") {
		t.Fatalf("standalone OpenAPI declared as rendered reference error = %v", err)
	}
}

func TestSourceImportVersion3MixedOpenAPIAndRenderedReferenceKeepsOpenAPIProjectionBytes(t *testing.T) {
	openAPI := []byte(`{"openapi":"3.0.3","info":{"title":"fixture","version":"1"},"paths":{"/alpha":{"get":{"operationId":"shared","responses":{"200":{"description":"ok"}}}}}}`)
	openAPILock, err := parseSourceImportLock(sourceImportV3FixtureLock(t, "fixture", []sourceImportV3FixtureDocument{{ID: "alpha", Path: "/alpha", Artifact: openAPI}}), "fixture")
	if err != nil {
		t.Fatalf("parse OpenAPI-only v3 lock: %v", err)
	}
	openAPIResult, err := importSourceLockResult(context.Background(), openAPILock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) { return openAPI, nil }), defaultSourceImportLimits())
	if err != nil {
		t.Fatalf("import OpenAPI-only v3 lock: %v", err)
	}
	baseline, err := json.Marshal(openAPIResult.Operations[0])
	if err != nil {
		t.Fatal(err)
	}
	baselineDigest := sha256.Sum256(baseline)
	if got := hex.EncodeToString(baselineDigest[:]); got != "b4f02389465992c5daf69c2a98e989058c449cc8a54fa892a016e2b9d865e4e7" {
		t.Fatalf("OpenAPI-only v3 projection changed: SHA-256 %s", got)
	}

	var mixed map[string]any
	if err := json.Unmarshal(sourceImportV3FixtureLock(t, "fixture", []sourceImportV3FixtureDocument{{ID: "alpha", Path: "/alpha", Artifact: openAPI}}), &mixed); err != nil {
		t.Fatal(err)
	}
	referenceDocument, page := sourceImportRenderedReferenceDocument(t, renderedReferenceCitationURL)
	rest := mixed["rest"].(map[string]any)
	rest["source_documents"] = append(rest["source_documents"].([]any), referenceDocument)
	rest["coverage_confidence"] = map[string]any{"level": "documented", "basis": "The second document is a captured provider reference."}
	mixed["counts"] = map[string]any{"rest": 2, "graphql_query": 0, "graphql_mutation": 0, "total": 2}
	raw, err := json.Marshal(mixed)
	if err != nil {
		t.Fatal(err)
	}
	lock, err := parseSourceImportLock(raw, "fixture")
	if err != nil {
		t.Fatalf("parse mixed v3 lock: %v", err)
	}
	result, err := importSourceLockResult(context.Background(), lock, sourceImportFetchFunc(func(_ context.Context, sourceURL string) ([]byte, error) {
		switch sourceURL {
		case "https://fixtures.polymetrics.invalid/alpha.openapi.json":
			return openAPI, nil
		case renderedReferenceArtifactURL:
			return page, nil
		default:
			t.Fatalf("mixed lock fetched provenance URL %q", sourceURL)
			return nil, nil
		}
	}), defaultSourceImportLimits())
	if err != nil {
		t.Fatalf("import mixed v3 lock: %v", err)
	}
	if len(result.Operations) != 2 {
		t.Fatalf("mixed v3 imported operations = %#v", result.Operations)
	}
	for _, operation := range result.Operations {
		if operation.SourceID != "fixture.rest.alpha.shared" {
			continue
		}
		actual, marshalErr := json.Marshal(operation)
		if marshalErr != nil {
			t.Fatal(marshalErr)
		}
		if !bytes.Equal(actual, baseline) {
			t.Fatalf("OpenAPI operation projection changed in mixed lock:\n got %s\nwant %s", actual, baseline)
		}
	}
	if findings := validateSourceDescriptorAgainstLock("fixture", "sources/fixture-operation-descriptor.json", lock, sourceImportDescriptorDocument{SchemaVersion: 3, Operations: result.Operations}); len(findings) != 0 {
		t.Fatalf("mixed source projection findings = %+v", findings)
	}
}

func TestSourceImportVersion3RenderedReferenceRejectsUnverifiableEvidenceAndCitations(t *testing.T) {
	t.Run("captured page hash mismatch", func(t *testing.T) {
		raw, page := sourceImportV3RenderedReferenceLock(t, renderedReferenceCitationURL)
		lock, err := parseSourceImportLock(raw, "fixture")
		if err != nil {
			t.Fatalf("parse rendered-reference lock: %v", err)
		}
		_, err = importSourceLockResult(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) {
			return append(append([]byte(nil), page...), '!'), nil
		}), defaultSourceImportLimits())
		if err == nil || !strings.Contains(err.Error(), "source-lock refresh required") {
			t.Fatalf("rendered-reference hash mismatch error = %v", err)
		}
	})

	for _, tc := range []struct {
		name     string
		citation string
		want     string
	}{
		{name: "missing operation citation", citation: "", want: "citation"},
		{name: "foreign operation citation", citation: "https://unrelated.polymetrics.invalid/reference/widgets#list-widgets", want: "citation"},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			raw, _ := sourceImportV3RenderedReferenceLock(t, tc.citation)
			if _, err := parseSourceImportLock(raw, "fixture"); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("rendered-reference validation error = %v, want %q", err, tc.want)
			}
		})
	}

	t.Run("missing coverage basis", func(t *testing.T) {
		raw, _ := sourceImportV3RenderedReferenceLock(t, renderedReferenceCitationURL)
		var lock map[string]any
		if err := json.Unmarshal(raw, &lock); err != nil {
			t.Fatal(err)
		}
		lock["rest"].(map[string]any)["coverage_confidence"].(map[string]any)["basis"] = ""
		raw, err := json.Marshal(lock)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := parseSourceImportLock(raw, "fixture"); err == nil || !strings.Contains(err.Error(), "coverage confidence") {
			t.Fatalf("missing coverage basis error = %v", err)
		}
	})

	for _, tc := range []struct {
		name   string
		mutate func(map[string]any, map[string]any)
	}{
		{name: "missing captured hash", mutate: func(artifact, _ map[string]any) { artifact["sha256"] = "" }},
		{name: "missing capture URL", mutate: func(_, published map[string]any) { published["capture_url"] = "" }},
		{name: "missing capture adapter", mutate: func(_, published map[string]any) { published["adapter"] = "" }},
		{name: "missing captured byte count", mutate: func(_, published map[string]any) { published["bytes"] = 0 }},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			raw, _ := sourceImportV3RenderedReferenceLock(t, renderedReferenceCitationURL)
			var lock map[string]any
			if err := json.Unmarshal(raw, &lock); err != nil {
				t.Fatal(err)
			}
			document := lock["rest"].(map[string]any)["source_documents"].([]any)[0].(map[string]any)
			tc.mutate(document["artifact"].(map[string]any), document["published_source"].(map[string]any))
			raw, err := json.Marshal(lock)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := parseSourceImportLock(raw, "fixture"); err == nil {
				t.Fatal("rendered-reference lock accepted incomplete captured evidence")
			}
		})
	}
}

func TestSourceImportVersion3RenderedReferenceCitationNeedsPointerOrVerifiedCaptureBinding(t *testing.T) {
	t.Run("generic publication URL", func(t *testing.T) {
		raw, _ := sourceImportV3RenderedReferenceLock(t, renderedReferencePublishedURL)
		if _, err := parseSourceImportLock(raw, "fixture"); err == nil || !strings.Contains(err.Error(), "citation") {
			t.Fatalf("generic rendered publication citation error = %v, want operation-specific citation refusal", err)
		}
	})

	t.Run("capture extraction binding", func(t *testing.T) {
		raw, page := sourceImportV3RenderedReferenceLock(t, renderedReferencePublishedURL)
		var lock map[string]any
		if err := json.Unmarshal(raw, &lock); err != nil {
			t.Fatal(err)
		}
		document := lock["rest"].(map[string]any)["source_documents"].([]any)[0].(map[string]any)
		operation := document["operations"].([]any)[0].(map[string]any)
		digest := sha256.Sum256(page)
		operation["citation_binding"] = map[string]any{
			"capture_url":     renderedReferenceArtifactURL,
			"capture_sha256":  hex.EncodeToString(digest[:]),
			"capture_bytes":   len(page),
			"source_location": "#list-widgets",
		}
		raw, err := json.Marshal(lock)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := parseSourceImportLock(raw, "fixture"); err != nil {
			t.Fatalf("verified rendered capture binding rejected: %v", err)
		}

		operation["citation_binding"].(map[string]any)["capture_sha256"] = strings.Repeat("0", 64)
		raw, err = json.Marshal(lock)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := parseSourceImportLock(raw, "fixture"); err == nil || !strings.Contains(err.Error(), "citation") {
			t.Fatalf("mismatched rendered capture binding error = %v, want citation refusal", err)
		}
	})
}

func TestSourceImportVersion3BundleRejectsArchiveHashMismatch(t *testing.T) {
	raw, archive := sourceImportV3BundleLock(t)
	lock, err := parseSourceImportLock(raw, "fixture")
	if err != nil {
		t.Fatalf("parse bundle lock: %v", err)
	}
	result, err := importSourceLockResult(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) {
		return archive, nil
	}), defaultSourceImportLimits())
	if err != nil {
		t.Fatalf("import bundle lock: %v", err)
	}
	if len(result.Operations) != 1 || result.Operations[0].Source.Form != "bundle" {
		t.Fatalf("bundle operation projection = %#v", result.Operations)
	}
	_, err = importSourceLockResult(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) {
		return append(append([]byte(nil), archive...), '!'), nil
	}), defaultSourceImportLimits())
	if err == nil || !strings.Contains(err.Error(), "source-lock refresh required") {
		t.Fatalf("bundle archive hash mismatch error = %v", err)
	}
}

func TestSourceImportVersion3BundleProjectsGzipCapture(t *testing.T) {
	var archive bytes.Buffer
	writer := gzip.NewWriter(&archive)
	if _, err := writer.Write([]byte("fixture gzip archive bytes")); err != nil {
		t.Fatalf("write gzip fixture: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close gzip fixture: %v", err)
	}

	raw, captured := sourceImportV3BundleLockWithCapture(t, "application/x-gzip", archive.Bytes())
	lock, err := parseSourceImportLock(raw, "fixture")
	if err != nil {
		t.Fatalf("parse gzip bundle lock: %v", err)
	}
	result, err := importSourceLockResult(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) {
		return captured, nil
	}), defaultSourceImportLimits())
	if err != nil {
		t.Fatalf("import gzip bundle lock: %v", err)
	}
	if len(result.Operations) != 1 || result.Operations[0].Source.Form != "bundle" {
		t.Fatalf("gzip bundle operation projection = %#v", result.Operations)
	}
}

func TestSourceImportVersion3UnavailableSourceProjectsBlockingGap(t *testing.T) {
	lockRaw := sourceImportV3UnavailableLock(t)
	lock, err := parseSourceImportLock(lockRaw, "fixture")
	if err != nil {
		t.Fatalf("parse unavailable-source lock: %v", err)
	}
	descriptorRaw, err := json.Marshal(sourceImportDescriptorDocument{SchemaVersion: 3, Operations: []sourceOperationDescriptor{}})
	if err != nil {
		t.Fatal(err)
	}
	findings := checkSourceProjection(fstest.MapFS{
		"fixture/sources/fixture-operation-source-lock.json": &fstest.MapFile{Data: lockRaw},
		"fixture/sources/fixture-operation-descriptor.json":  &fstest.MapFile{Data: descriptorRaw},
	}, engine.Bundle{Name: lock.Connector})
	if len(findings) != 1 || !strings.Contains(findings[0].Message, "unavailable") || !strings.Contains(findings[0].Message, "https://docs.polymetrics.invalid/reference/availability") {
		t.Fatalf("unavailable source findings = %+v", findings)
	}
}

func TestSourceImportVersion3UnavailableSourceDoesNotRequireCapturedArtifact(t *testing.T) {
	raw := sourceImportV3UnavailableLock(t)
	var wire map[string]any
	if err := json.Unmarshal(raw, &wire); err != nil {
		t.Fatal(err)
	}
	document := wire["rest"].(map[string]any)["source_documents"].([]any)[0].(map[string]any)
	delete(document, "content_type")
	delete(document, "artifact")
	delete(document, "published_source")
	raw, err := json.Marshal(wire)
	if err != nil {
		t.Fatal(err)
	}
	lock, err := parseSourceImportLock(raw, "fixture")
	if err != nil {
		t.Fatalf("parse source-traced unavailable lock without capture: %v", err)
	}
	fetched := false
	_, err = importSourceLockResult(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) {
		fetched = true
		return nil, nil
	}), defaultSourceImportLimits())
	if err == nil || !strings.Contains(err.Error(), "unavailable") {
		t.Fatalf("unavailable import error = %v", err)
	}
	if fetched {
		t.Fatal("unavailable source attempted an artifact fetch")
	}
	descriptorRaw, err := json.Marshal(sourceImportDescriptorDocument{SchemaVersion: 3, Operations: []sourceOperationDescriptor{}})
	if err != nil {
		t.Fatal(err)
	}
	findings := checkSourceProjection(fstest.MapFS{
		"fixture/sources/fixture-operation-source-lock.json": &fstest.MapFile{Data: raw},
		"fixture/sources/fixture-operation-descriptor.json":  &fstest.MapFile{Data: descriptorRaw},
	}, engine.Bundle{Name: lock.Connector})
	if len(findings) != 1 || !strings.Contains(findings[0].Message, "unavailable") || !strings.Contains(findings[0].Message, "provider has no usable API reference") {
		t.Fatalf("unavailable source findings = %+v", findings)
	}
}

func TestSourceImportRenderedReferenceKeepsSchemaOneAndTwoLocksValid(t *testing.T) {
	for _, schemaVersion := range []int{1, 2} {
		schemaVersion := schemaVersion
		t.Run("schema version "+string(rune('0'+schemaVersion)), func(t *testing.T) {
			lock := sourceImportFixtureLock("fixture", "https://fixtures.polymetrics.invalid/legacy.openapi.json", []byte(`{"openapi":"3.0.3","info":{"title":"fixture","version":"1"},"paths":{}}`))
			lock.SchemaVersion = schemaVersion
			raw, err := json.Marshal(lock)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := parseSourceImportLock(raw, "fixture"); err != nil {
				t.Fatalf("schema %d lock no longer validates: %v", schemaVersion, err)
			}
		})
	}
}
