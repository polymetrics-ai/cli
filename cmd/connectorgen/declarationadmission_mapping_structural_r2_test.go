package main

import (
	"encoding/json"
	"strings"
	"testing"
)

// These R2 cases deliberately exercise the mapping-only reader directly. The
// source-import reader is intentionally stricter about retained artifacts;
// mapping admission must ignore that retention representation while keeping
// executable source identity and inventory structure closed.
func TestDeclarationAdmissionMappingReaderR2SourceReferenceStructure(t *testing.T) {
	tests := []struct {
		name   string
		lock   func(*testing.T) map[string]any
		mutate func(map[string]any)
		want   string
	}{
		{
			name: "accepts legacy source reference with malformed retention and enriched operation",
			lock: declarationAdmissionR2LegacyReferenceLock,
			mutate: func(lock map[string]any) {
				rest := declarationAdmissionR2REST(lock)
				rest["sha256"] = map[string]any{"ignored": "retention"}
				rest["bytes"] = "ignored-retention-byte-count"
				rest["operations"].([]any)[0].(map[string]any)["source_operation"] = map[string]any{
					"request": map[string]any{"schema": "enriched source operation is opaque to mapping admission"},
				}
			},
		},
		{
			name:   "rejects legacy source reference on schema v1",
			lock:   declarationAdmissionR2LegacyReferenceLock,
			mutate: func(lock map[string]any) { lock["schema_version"] = 1 },
			want:   "schema version 2",
		},
		{
			name:   "rejects legacy lower case method",
			lock:   declarationAdmissionR2LegacyReferenceLock,
			mutate: func(lock map[string]any) { declarationAdmissionR2LegacyOperation(lock, 0)["method"] = "get" },
			want:   "non-canonical REST operation identity",
		},
		{
			name: "rejects legacy unsupported method",
			lock: declarationAdmissionR2LegacyReferenceLock,
			mutate: func(lock map[string]any) {
				declarationAdmissionR2LegacyOperation(lock, 0)["method"] = "TRACE"
				declarationAdmissionR2REST(lock)["operation_counts"] = map[string]any{"TRACE": 1, "POST": 1}
			},
			want: "unsupported HTTP method",
		},
		{
			name: "rejects legacy control character operation identity",
			lock: declarationAdmissionR2LegacyReferenceLock,
			mutate: func(lock map[string]any) {
				declarationAdmissionR2LegacyOperation(lock, 0)["id"] = "fixture.rest.get.widgets\u0000"
			},
			want: "non-canonical REST operation identity",
		},
		{
			name: "rejects legacy overlong operation identity",
			lock: declarationAdmissionR2LegacyReferenceLock,
			mutate: func(lock map[string]any) {
				declarationAdmissionR2LegacyOperation(lock, 0)["id"] = strings.Repeat("a", 1025)
			},
			want: "non-canonical REST operation identity",
		},
		{
			name: "rejects legacy control character provider operation ID",
			lock: declarationAdmissionR2LegacyReferenceLock,
			mutate: func(lock map[string]any) {
				declarationAdmissionR2LegacyOperation(lock, 0)["operation_id"] = "list\u0000widgets"
			},
			want: "invalid provider operation ID",
		},
		{
			name: "rejects legacy overlong provider operation ID",
			lock: declarationAdmissionR2LegacyReferenceLock,
			mutate: func(lock map[string]any) {
				declarationAdmissionR2LegacyOperation(lock, 0)["operation_id"] = strings.Repeat("p", 1025)
			},
			want: "invalid provider operation ID",
		},
		{
			name: "rejects legacy control character source location",
			lock: declarationAdmissionR2LegacyReferenceLock,
			mutate: func(lock map[string]any) {
				declarationAdmissionR2LegacyOperation(lock, 1)["source_location"] = "custom\nreference"
			},
			want: "non-canonical REST operation identity",
		},
		{
			name: "rejects legacy supplement location not bound to its operations",
			lock: declarationAdmissionR2LegacyReferenceLock,
			mutate: func(lock map[string]any) {
				declarationAdmissionR2REST(lock)["supplements"].([]any)[0].(map[string]any)["source_location"] = "different location"
			},
			want: "citation location",
		},
		{
			name: "rejects legacy supplement control character location",
			lock: declarationAdmissionR2LegacyReferenceLock,
			mutate: func(lock map[string]any) {
				declarationAdmissionR2REST(lock)["supplements"].([]any)[0].(map[string]any)["source_location"] = "custom\nreference"
			},
			want: "invalid mapping evidence",
		},
		{
			name: "rejects legacy supplement count not bound to its operations",
			lock: declarationAdmissionR2LegacyReferenceLock,
			mutate: func(lock map[string]any) {
				declarationAdmissionR2REST(lock)["supplements"].([]any)[0].(map[string]any)["operation_count"] = 2
			},
			want: "operation count",
		},
		{
			name: "rejects legacy source reference operation citation URL",
			lock: declarationAdmissionR2LegacyReferenceLock,
			mutate: func(lock map[string]any) {
				declarationAdmissionR2LegacyOperation(lock, 0)["citation_url"] = "https://docs.polymetrics.invalid/fixture/openapi#widgets"
			},
			want: "must not declare a citation URL",
		},
		{
			name: "rejects legacy duplicate route",
			lock: declarationAdmissionR2LegacyReferenceLock,
			mutate: func(lock map[string]any) {
				declarationAdmissionR2LegacyOperation(lock, 1)["method"] = "GET"
				declarationAdmissionR2LegacyOperation(lock, 1)["path"] = "/widgets"
				declarationAdmissionR2REST(lock)["operation_counts"] = map[string]any{"GET": 2}
			},
			want: "duplicates REST route",
		},
		{
			name: "rejects legacy operations found mismatch",
			lock: declarationAdmissionR2LegacyReferenceLock,
			mutate: func(lock map[string]any) {
				lock["operations_found"] = map[string]any{"rest": 1, "graphql_query": 0, "graphql_mutation": 0, "total": 1}
			},
			want: "operations_found",
		},
		{
			name:   "rejects legacy coverage confidence that does not repeat source kind",
			lock:   declarationAdmissionR2LegacyReferenceLock,
			mutate: func(lock map[string]any) { lock["coverage_confidence"].(map[string]any)["level"] = "partial" },
			want:   "coverage confidence",
		},
		{
			name: "accepts v3 source reference with malformed retention and enriched operation",
			lock: declarationAdmissionR2V3ReferenceLock,
			mutate: func(lock map[string]any) {
				document := declarationAdmissionR2Document(lock, 0)
				reference := document["source_reference"].(map[string]any)
				reference["sha256"] = map[string]any{"ignored": "retention"}
				reference["bytes"] = "ignored-retention-byte-count"
				document["artifact"] = map[string]any{
					"source_url": "", "sha256": map[string]any{"ignored": "retention"}, "bytes": "ignored-retention-byte-count",
				}
				document["published_source"] = map[string]any{
					"source_url": "", "capture_url": "https://captures.polymetrics.invalid/fixture", "sha256": map[string]any{"ignored": "retention"}, "bytes": "ignored-retention-byte-count", "adapter": false,
				}
				document["operations"].([]any)[0].(map[string]any)["source_operation"] = []any{"opaque", map[string]any{"payload": true}}
			},
		},
		{
			name:   "rejects v3 lower case method",
			lock:   declarationAdmissionR2V3ReferenceLock,
			mutate: func(lock map[string]any) { declarationAdmissionR2DocumentOperation(lock, 0, 0)["method"] = "get" },
			want:   "non-canonical REST operation identity",
		},
		{
			name:   "rejects v3 unsupported method",
			lock:   declarationAdmissionR2V3ReferenceLock,
			mutate: func(lock map[string]any) { declarationAdmissionR2DocumentOperation(lock, 0, 0)["method"] = "TRACE" },
			want:   "unsupported HTTP method",
		},
		{
			name: "rejects v3 control character operation identity",
			lock: declarationAdmissionR2V3ReferenceLock,
			mutate: func(lock map[string]any) {
				declarationAdmissionR2DocumentOperation(lock, 0, 0)["id"] = "fixture.rest.source.get\u0000"
			},
			want: "non-canonical REST operation identity",
		},
		{
			name: "rejects v3 overlong operation identity",
			lock: declarationAdmissionR2V3ReferenceLock,
			mutate: func(lock map[string]any) {
				declarationAdmissionR2DocumentOperation(lock, 0, 0)["id"] = strings.Repeat("a", 1025)
			},
			want: "non-canonical REST operation identity",
		},
		{
			name: "rejects v3 control character provider operation ID",
			lock: declarationAdmissionR2V3ReferenceLock,
			mutate: func(lock map[string]any) {
				declarationAdmissionR2DocumentOperation(lock, 0, 0)["operation_id"] = "source\u0000operation"
			},
			want: "invalid provider operation ID",
		},
		{
			name: "rejects v3 overlong provider operation ID",
			lock: declarationAdmissionR2V3ReferenceLock,
			mutate: func(lock map[string]any) {
				declarationAdmissionR2DocumentOperation(lock, 0, 0)["operation_id"] = strings.Repeat("p", 1025)
			},
			want: "invalid provider operation ID",
		},
		{
			name: "rejects v3 control character source location",
			lock: declarationAdmissionR2V3ReferenceLock,
			mutate: func(lock map[string]any) {
				declarationAdmissionR2DocumentOperation(lock, 0, 0)["source_location"] = "paths\nwidgets"
			},
			want: "non-canonical REST operation identity",
		},
		{
			name: "rejects v3 source reference mixed with an artifact",
			lock: declarationAdmissionR2V3ReferenceLock,
			mutate: func(lock map[string]any) {
				declarationAdmissionR2Document(lock, 0)["artifact"] = map[string]any{"source_url": "https://docs.polymetrics.invalid/fixture/other"}
			},
			want: "cannot mix",
		},
		{
			name: "rejects v3 source reference without coverage confidence",
			lock: declarationAdmissionR2V3ReferenceLock,
			mutate: func(lock map[string]any) {
				delete(declarationAdmissionR2REST(lock), "coverage_confidence")
			},
			want: "require coverage confidence",
		},
		{
			name: "rejects zero operation v3 source reference",
			lock: declarationAdmissionR2V3ReferenceLock,
			mutate: func(lock map[string]any) {
				declarationAdmissionR2Document(lock, 0)["operations"] = []any{}
				declarationAdmissionR2Counts(lock)["rest"] = 0
				declarationAdmissionR2Counts(lock)["total"] = 0
			},
			want: "has no operations",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			lock := testCase.lock(t)
			testCase.mutate(lock)
			err := declarationAdmissionR2Parse(t, lock)
			if testCase.want == "" {
				if err != nil {
					t.Fatalf("mapping reader rejected structurally valid reference: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), testCase.want) {
				t.Fatalf("mapping reader error = %v, want containing %q", err, testCase.want)
			}
		})
	}
}

func TestDeclarationAdmissionMappingReaderR2UnavailableStructure(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(map[string]any)
		want   string
	}{
		{
			name: "accepts unavailable reason without retained evidence",
			mutate: func(lock map[string]any) {
				document := declarationAdmissionR2Document(lock, 0)
				delete(document, "artifact")
				delete(document, "published_source")
			},
		},
		{
			name: "accepts canonical optional unavailable source URL with malformed retained evidence",
			mutate: func(lock map[string]any) {
				document := declarationAdmissionR2Document(lock, 0)
				document["artifact"].(map[string]any)["sha256"] = map[string]any{"ignored": true}
				document["published_source"].(map[string]any)["bytes"] = "ignored"
			},
		},
		{
			name:   "rejects unavailable reason with wrong type",
			mutate: func(lock map[string]any) { declarationAdmissionR2Document(lock, 0)["unavailable_reason"] = 42 },
			want:   "invalid reason",
		},
		{
			name:   "rejects empty unavailable reason",
			mutate: func(lock map[string]any) { declarationAdmissionR2Document(lock, 0)["unavailable_reason"] = "" },
			want:   "invalid reason",
		},
		{
			name: "rejects untrimmed unavailable reason",
			mutate: func(lock map[string]any) {
				declarationAdmissionR2Document(lock, 0)["unavailable_reason"] = " provider unavailable"
			},
			want: "invalid reason",
		},
		{
			name: "rejects control character unavailable reason",
			mutate: func(lock map[string]any) {
				declarationAdmissionR2Document(lock, 0)["unavailable_reason"] = "provider\nunavailable"
			},
			want: "invalid reason",
		},
		{
			name: "rejects overlong unavailable reason",
			mutate: func(lock map[string]any) {
				declarationAdmissionR2Document(lock, 0)["unavailable_reason"] = strings.Repeat("a", 1025)
			},
			want: "invalid reason",
		},
		{
			name: "rejects noncanonical optional unavailable source URL",
			mutate: func(lock map[string]any) {
				declarationAdmissionR2Document(lock, 0)["published_source"].(map[string]any)["source_url"] = "http://docs.polymetrics.invalid/reference/availability"
			},
			want: "invalid source URL",
		},
		{
			name:   "rejects unavailable document without coverage confidence",
			mutate: func(lock map[string]any) { delete(declarationAdmissionR2REST(lock), "coverage_confidence") },
			want:   "require coverage confidence",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			lock := declarationAdmissionR2UnavailableLock(t)
			testCase.mutate(lock)
			err := declarationAdmissionR2Parse(t, lock)
			if testCase.want == "" {
				if err != nil {
					t.Fatalf("mapping reader rejected structurally valid unavailable document: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), testCase.want) {
				t.Fatalf("mapping reader error = %v, want containing %q", err, testCase.want)
			}
		})
	}
}

func TestDeclarationAdmissionMappingReaderR2V3EnvelopeStructure(t *testing.T) {
	tests := []struct {
		name   string
		lock   func(*testing.T) map[string]any
		mutate func(map[string]any)
		want   string
	}{
		{
			name: "accepts rendered reference with malformed retention",
			lock: declarationAdmissionR2RenderedReferenceLock,
			mutate: func(lock map[string]any) {
				document := declarationAdmissionR2Document(lock, 0)
				document["artifact"].(map[string]any)["bytes"] = "ignored"
				document["published_source"].(map[string]any)["sha256"] = map[string]any{"ignored": true}
			},
		},
		{
			name:   "rejects blank retrieval",
			lock:   declarationAdmissionR2RenderedReferenceLock,
			mutate: func(lock map[string]any) { declarationAdmissionR2REST(lock)["retrieval"] = "" },
			want:   "retrieval",
		},
		{
			name:   "rejects control character retrieval",
			lock:   declarationAdmissionR2RenderedReferenceLock,
			mutate: func(lock map[string]any) { declarationAdmissionR2REST(lock)["retrieval"] = "captured\nreference" },
			want:   "retrieval",
		},
		{
			name:   "rejects overlong retrieval",
			lock:   declarationAdmissionR2RenderedReferenceLock,
			mutate: func(lock map[string]any) { declarationAdmissionR2REST(lock)["retrieval"] = strings.Repeat("r", 1025) },
			want:   "retrieval",
		},
		{
			name: "rejects no source documents",
			lock: declarationAdmissionR2RenderedReferenceLock,
			mutate: func(lock map[string]any) {
				declarationAdmissionR2REST(lock)["source_documents"] = []any{}
				declarationAdmissionR2Counts(lock)["rest"] = 0
				declarationAdmissionR2Counts(lock)["total"] = 0
			},
			want: "no v3 REST source documents",
		},
		{
			name: "rejects too many source documents",
			lock: declarationAdmissionR2UnavailableLock,
			mutate: func(lock map[string]any) {
				rest := declarationAdmissionR2REST(lock)
				documents := rest["source_documents"].([]any)
				for len(documents) <= defaultSourceImportDocuments {
					documents = append(documents, documents[0])
				}
				rest["source_documents"] = documents
			},
			want: "document count exceeds",
		},
		{
			name:   "rejects unsorted aggregate OpenAPI versions",
			lock:   declarationAdmissionR2OpenAPILock,
			mutate: func(lock map[string]any) { declarationAdmissionR2REST(lock)["openapi"] = []any{"3.1.0", "3.0.3"} },
			want:   "OpenAPI versions are not sorted",
		},
		{
			name:   "rejects duplicate aggregate OpenAPI version",
			lock:   declarationAdmissionR2OpenAPILock,
			mutate: func(lock map[string]any) { declarationAdmissionR2REST(lock)["openapi"] = []any{"3.0.3", "3.0.3"} },
			want:   "invalid or duplicate",
		},
		{
			name:   "rejects unsupported aggregate OpenAPI version",
			lock:   declarationAdmissionR2OpenAPILock,
			mutate: func(lock map[string]any) { declarationAdmissionR2REST(lock)["openapi"] = []any{"2.0.0"} },
			want:   "invalid or duplicate",
		},
		{
			name: "rejects OpenAPI document missing source form pin",
			lock: declarationAdmissionR2OpenAPILock,
			mutate: func(lock map[string]any) {
				delete(declarationAdmissionR2Document(lock, 0)["artifact"].(map[string]any), "openapi")
			},
			want: "OpenAPI version outside",
		},
		{
			name: "rejects OpenAPI document whose pin is absent from aggregate inventory",
			lock: declarationAdmissionR2OpenAPILock,
			mutate: func(lock map[string]any) {
				declarationAdmissionR2Document(lock, 0)["artifact"].(map[string]any)["openapi"] = "3.1.0"
			},
			want: "outside the aggregate",
		},
		{
			name:   "rejects aggregate version without OpenAPI document",
			lock:   declarationAdmissionR2RenderedReferenceLock,
			mutate: func(lock map[string]any) { declarationAdmissionR2REST(lock)["openapi"] = []any{"3.0.3"} },
			want:   "require an OpenAPI document",
		},
		{
			name: "rejects rendered source form pin",
			lock: declarationAdmissionR2RenderedReferenceLock,
			mutate: func(lock map[string]any) {
				declarationAdmissionR2Document(lock, 0)["artifact"].(map[string]any)["openapi"] = "3.0.3"
			},
			want: "must not declare an OpenAPI or Swagger version",
		},
		{
			name: "accepts zero operation rendered reference",
			lock: declarationAdmissionR2RenderedReferenceLock,
			mutate: func(lock map[string]any) {
				declarationAdmissionR2Document(lock, 0)["operations"] = []any{}
				declarationAdmissionR2Counts(lock)["rest"] = 0
				declarationAdmissionR2Counts(lock)["total"] = 0
			},
		},
		{
			name: "rejects zero operation OpenAPI document",
			lock: declarationAdmissionR2OpenAPILock,
			mutate: func(lock map[string]any) {
				declarationAdmissionR2Document(lock, 0)["operations"] = []any{}
				declarationAdmissionR2Counts(lock)["rest"] = 0
				declarationAdmissionR2Counts(lock)["total"] = 0
			},
			want: "has no operations",
		},
		{
			name:   "rejects missing rendered coverage confidence",
			lock:   declarationAdmissionR2RenderedReferenceLock,
			mutate: func(lock map[string]any) { delete(declarationAdmissionR2REST(lock), "coverage_confidence") },
			want:   "require coverage confidence",
		},
		{
			name: "rejects invalid coverage confidence level",
			lock: declarationAdmissionR2RenderedReferenceLock,
			mutate: func(lock map[string]any) {
				declarationAdmissionR2REST(lock)["coverage_confidence"].(map[string]any)["level"] = "\n"
			},
			want: "invalid v3 REST coverage confidence",
		},
		{
			name: "rejects invalid coverage confidence basis",
			lock: declarationAdmissionR2RenderedReferenceLock,
			mutate: func(lock map[string]any) {
				declarationAdmissionR2REST(lock)["coverage_confidence"].(map[string]any)["basis"] = strings.Repeat("b", 4097)
			},
			want: "invalid v3 REST coverage confidence",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			lock := testCase.lock(t)
			testCase.mutate(lock)
			err := declarationAdmissionR2Parse(t, lock)
			if testCase.want == "" {
				if err != nil {
					t.Fatalf("mapping reader rejected structurally valid v3 lock: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), testCase.want) {
				t.Fatalf("mapping reader error = %v, want containing %q", err, testCase.want)
			}
		})
	}
}

func TestDeclarationAdmissionMappingReaderR2KeepsRenderedReferenceProjection(t *testing.T) {
	lock := declarationAdmissionR2RenderedReferenceLock(t)
	parsed, err := parseDeclarationAdmissionSourceLock(declarationAdmissionR2Marshal(t, lock), "fixture")
	if err != nil {
		t.Fatalf("parse rendered-reference mapping lock: %v", err)
	}
	operation, found := parsed.Operations["fixture.rest.reference.list-widgets"]
	if !found {
		t.Fatalf("rendered-reference operation was not projected: %#v", parsed.Operations)
	}
	if operation.SourceReference || operation.CitationURL != renderedReferenceCitationURL || operation.PublishedSourceURL != renderedReferencePublishedURL || operation.SourceURL != renderedReferencePublishedURL {
		t.Fatalf("rendered-reference mapping projection = %#v", operation)
	}
}

func declarationAdmissionR2LegacyReferenceLock(t *testing.T) map[string]any {
	t.Helper()
	return map[string]any{
		"schema_version": 2,
		"connector":      "fixture",
		"rest": map[string]any{
			"source_url":       "https://docs.polymetrics.invalid/fixture/openapi",
			"sha256":           "retention-is-ignored-by-mapping-admission",
			"bytes":            "retention-is-ignored-by-mapping-admission",
			"source_kind":      sourceImportLegacySourceReferenceKind,
			"operation_counts": map[string]any{"GET": 1, "POST": 1},
			"supplements": []any{map[string]any{
				"source_url":      "https://docs.polymetrics.invalid/fixture/custom",
				"sha256":          map[string]any{"retention": "ignored"},
				"bytes":           -7,
				"source_location": "custom reference > endpoint",
				"operation_count": 1,
			}},
			"operations": []any{
				map[string]any{
					"id": "fixture.rest.get.widgets", "protocol": "rest", "method": "GET", "path": "/widgets",
					"operation_id": "listWidgets", "source_location": `paths["/widgets"].get`, "source_url": "https://docs.polymetrics.invalid/fixture/openapi",
				},
				map[string]any{
					"id": "fixture.rest.post.custom", "protocol": "rest", "method": "POST", "path": "/custom",
					"operation_id": "createCustom", "source_location": "custom reference > endpoint", "source_url": "https://docs.polymetrics.invalid/fixture/custom",
				},
			},
		},
		"operations_found":    map[string]any{"rest": 2, "graphql_query": 0, "graphql_mutation": 0, "total": 2},
		"coverage_confidence": map[string]any{"level": sourceImportLegacySourceReferenceKind, "basis": "The provider documents a complete machine-readable specification with one rendered dynamic supplement."},
		"counts":              map[string]any{"rest": 2, "graphql_query": 0, "graphql_mutation": 0, "total": 2},
	}
}

func declarationAdmissionR2V3ReferenceLock(t *testing.T) map[string]any {
	t.Helper()
	return declarationAdmissionR2LockMap(t, sourceImportV3SourceReferenceLock(t, "fixture", "source", "https://docs.polymetrics.invalid/fixture/openapi", strings.Repeat("a", 64), 512, "GET", "/widgets"))
}

func declarationAdmissionR2UnavailableLock(t *testing.T) map[string]any {
	t.Helper()
	return declarationAdmissionR2LockMap(t, sourceImportV3UnavailableLock(t))
}

func declarationAdmissionR2RenderedReferenceLock(t *testing.T) map[string]any {
	t.Helper()
	raw, _ := sourceImportV3RenderedReferenceLock(t, renderedReferenceCitationURL)
	return declarationAdmissionR2LockMap(t, raw)
}

func declarationAdmissionR2OpenAPILock(t *testing.T) map[string]any {
	t.Helper()
	return map[string]any{
		"schema_version": 3,
		"connector":      "fixture",
		"rest": map[string]any{
			"retrieval": "hermetic OpenAPI mapping-only fixture",
			"openapi":   []any{"3.0.3"},
			"source_documents": []any{map[string]any{
				"id": "primary",
				"artifact": map[string]any{
					"source_url": "https://fixtures.polymetrics.invalid/reference/openapi.json", "sha256": "retention-is-ignored", "bytes": -7, "openapi": "3.0.3",
				},
				"published_source": map[string]any{
					"source_url": "https://docs.polymetrics.invalid/reference/openapi", "capture_url": 123, "sha256": map[string]any{"ignored": true}, "bytes": "ignored", "adapter": false,
				},
				"operations": []any{map[string]any{
					"id": "fixture.rest.primary.list-widgets", "protocol": "rest", "method": "GET", "path": "/widgets", "operation_id": "listWidgets", "source_location": `paths["/widgets"].get`,
				}},
			}},
		},
		"counts": map[string]any{"rest": 1, "graphql_query": 0, "graphql_mutation": 0, "total": 1},
	}
}

func declarationAdmissionR2LockMap(t *testing.T, raw []byte) map[string]any {
	t.Helper()
	var lock map[string]any
	if err := json.Unmarshal(raw, &lock); err != nil {
		t.Fatalf("decode mapping lock fixture: %v", err)
	}
	return lock
}

func declarationAdmissionR2Parse(t *testing.T, lock map[string]any) error {
	t.Helper()
	raw := declarationAdmissionR2Marshal(t, lock)
	connector, _ := lock["connector"].(string)
	_, err := parseDeclarationAdmissionSourceLock(raw, connector)
	return err
}

func declarationAdmissionR2Marshal(t *testing.T, lock map[string]any) []byte {
	t.Helper()
	raw, err := json.Marshal(lock)
	if err != nil {
		t.Fatalf("encode mapping lock fixture: %v", err)
	}
	return raw
}

func declarationAdmissionR2REST(lock map[string]any) map[string]any {
	return lock["rest"].(map[string]any)
}

func declarationAdmissionR2Counts(lock map[string]any) map[string]any {
	return lock["counts"].(map[string]any)
}

func declarationAdmissionR2LegacyOperation(lock map[string]any, index int) map[string]any {
	return declarationAdmissionR2REST(lock)["operations"].([]any)[index].(map[string]any)
}

func declarationAdmissionR2Document(lock map[string]any, index int) map[string]any {
	return declarationAdmissionR2REST(lock)["source_documents"].([]any)[index].(map[string]any)
}

func declarationAdmissionR2DocumentOperation(lock map[string]any, documentIndex, operationIndex int) map[string]any {
	return declarationAdmissionR2Document(lock, documentIndex)["operations"].([]any)[operationIndex].(map[string]any)
}
