package main

import (
	"encoding/json"
	"strconv"
	"testing"
)

// TestDeclarationAdmissionMappingReaderR3RejectsCrossVersionAndVariantDrift
// is the frozen-review matrix for the mapping-only reader. It deliberately
// varies structural fields only: bytes, digests, capture metadata, and opaque
// source_operation/source_contract leaves remain non-binding to mapping.
func TestDeclarationAdmissionMappingReaderR3RejectsCrossVersionAndVariantDrift(t *testing.T) {
	ordinaryTests := []struct {
		name   string
		mutate func(map[string]any)
	}{
		{
			name: "root operations_found",
			mutate: func(lock map[string]any) {
				lock["operations_found"] = map[string]any{"rest": 1, "graphql_query": 0, "graphql_mutation": 0, "total": 1}
			},
		},
		{
			name: "root coverage_confidence",
			mutate: func(lock map[string]any) {
				lock["coverage_confidence"] = map[string]any{"level": "ordinary", "basis": "foreign variant"}
			},
		},
		{
			name: "empty source_kind discriminator",
			mutate: func(lock map[string]any) {
				lock["rest"].(map[string]any)["source_kind"] = ""
			},
		},
		{
			name: "null source_kind discriminator",
			mutate: func(lock map[string]any) {
				lock["rest"].(map[string]any)["source_kind"] = nil
			},
		},
		{
			name: "empty operation_counts",
			mutate: func(lock map[string]any) {
				lock["rest"].(map[string]any)["operation_counts"] = map[string]any{}
			},
		},
		{
			name: "null operation_counts",
			mutate: func(lock map[string]any) {
				lock["rest"].(map[string]any)["operation_counts"] = nil
			},
		},
		{
			name: "empty supplements",
			mutate: func(lock map[string]any) {
				lock["rest"].(map[string]any)["supplements"] = []any{}
			},
		},
		{
			name: "null supplements",
			mutate: func(lock map[string]any) {
				lock["rest"].(map[string]any)["supplements"] = nil
			},
		},
		{
			name: "empty operation source_url",
			mutate: func(lock map[string]any) {
				declarationAdmissionR3LegacyOperation(lock)["source_url"] = ""
			},
		},
		{
			name: "null operation source_url",
			mutate: func(lock map[string]any) {
				declarationAdmissionR3LegacyOperation(lock)["source_url"] = nil
			},
		},
	}
	for _, version := range []int{1, 2} {
		for _, test := range ordinaryTests {
			version, test := version, test
			t.Run("ordinary schema v"+strconv.Itoa(version)+" rejects "+test.name, func(t *testing.T) {
				lock := declarationAdmissionR3LegacyOrdinaryLock(version)
				test.mutate(lock)
				if err := declarationAdmissionR3Parse(lock); err == nil {
					t.Fatal("mapping reader accepted cross-version or variant field drift")
				}
			})
		}
	}

	tests := []struct {
		name   string
		lock   func() map[string]any
		mutate func(map[string]any)
	}{
		{
			name: "legacy source_reference rejects empty and null operation citation_url",
			lock: declarationAdmissionR3LegacyReferenceLock,
			mutate: func(lock map[string]any) {
				declarationAdmissionR3LegacyOperation(lock)["citation_url"] = ""
			},
		},
		{
			name: "legacy source_reference rejects null operation citation_url",
			lock: declarationAdmissionR3LegacyReferenceLock,
			mutate: func(lock map[string]any) {
				declarationAdmissionR3LegacyOperation(lock)["citation_url"] = nil
			},
		},
		{
			name: "legacy source_reference rejects null operation citation_binding",
			lock: declarationAdmissionR3LegacyReferenceLock,
			mutate: func(lock map[string]any) {
				declarationAdmissionR3LegacyOperation(lock)["citation_binding"] = nil
			},
		},
		{
			name: "schema v3 rejects foreign top-level operations_found",
			lock: func() map[string]any { return declarationAdmissionR3V3Lock("openapi") },
			mutate: func(lock map[string]any) {
				lock["operations_found"] = map[string]any{"rest": 1, "graphql_query": 0, "graphql_mutation": 0, "total": 1}
			},
		},
		{
			name: "schema v3 rejects foreign top-level coverage_confidence",
			lock: func() map[string]any { return declarationAdmissionR3V3Lock("openapi") },
			mutate: func(lock map[string]any) {
				lock["coverage_confidence"] = map[string]any{"level": "foreign", "basis": "foreign variant"}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			lock := test.lock()
			test.mutate(lock)
			if err := declarationAdmissionR3Parse(lock); err == nil {
				t.Fatal("mapping reader accepted cross-version or variant field drift")
			}
		})
	}

	for _, kind := range []string{"openapi", "rendered_reference", "bundle", "source_reference"} {
		for _, value := range []any{"", nil} {
			kind, value := kind, value
			t.Run("schema v3 "+kind+" rejects operation source_url variant "+declarationAdmissionR3ValueName(value), func(t *testing.T) {
				lock := declarationAdmissionR3V3Lock(kind)
				declarationAdmissionR3V3Operation(lock)["source_url"] = value
				if err := declarationAdmissionR3Parse(lock); err == nil {
					t.Fatal("mapping reader accepted v3 operation source_url variant drift")
				}
			})
		}
	}
}

func TestDeclarationAdmissionMappingReaderR3LegacyReferenceValidatesSourceFormPins(t *testing.T) {
	valid := declarationAdmissionR3LegacyReferenceLock()
	if err := declarationAdmissionR3Parse(valid); err != nil {
		t.Fatalf("mapping reader rejected valid source-form pins while retention is malformed: %v", err)
	}

	withoutPins := declarationAdmissionR3LegacyReferenceLock()
	for _, artifact := range []map[string]any{
		declarationAdmissionR3REST(withoutPins),
		declarationAdmissionR3REST(withoutPins)["supplements"].([]any)[0].(map[string]any),
	} {
		delete(artifact, "openapi")
		delete(artifact, "swagger")
	}
	if err := declarationAdmissionR3Parse(withoutPins); err != nil {
		t.Fatalf("mapping reader rejected absent optional source-form pins: %v", err)
	}

	for _, test := range []struct {
		name   string
		mutate func(map[string]any)
	}{
		{
			name: "accepts primary OpenAPI 3.1 and supplement Swagger 2.0",
			mutate: func(lock map[string]any) {
				declarationAdmissionR3REST(lock)["openapi"] = "3.1.0"
			},
		},
		{
			name: "accepts primary Swagger 2.0 and supplement OpenAPI 3.0",
			mutate: func(lock map[string]any) {
				rest := declarationAdmissionR3REST(lock)
				delete(rest, "openapi")
				rest["swagger"] = "2.0"
				supplement := declarationAdmissionR3Supplement(lock)
				delete(supplement, "swagger")
				supplement["openapi"] = "3.0.3"
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			lock := declarationAdmissionR3LegacyReferenceLock()
			test.mutate(lock)
			if err := declarationAdmissionR3Parse(lock); err != nil {
				t.Fatalf("mapping reader rejected supported source-form pins while retention is malformed: %v", err)
			}
		})
	}

	tests := []struct {
		name   string
		mutate func(map[string]any)
	}{
		{
			name:   "primary openapi wrong JSON type",
			mutate: func(lock map[string]any) { declarationAdmissionR3REST(lock)["openapi"] = 3 },
		},
		{
			name:   "supplement openapi wrong JSON type",
			mutate: func(lock map[string]any) { declarationAdmissionR3Supplement(lock)["openapi"] = 3 },
		},
		{
			name:   "primary unsupported openapi",
			mutate: func(lock map[string]any) { declarationAdmissionR3REST(lock)["openapi"] = "2.0.0" },
		},
		{
			name:   "supplement unsupported openapi",
			mutate: func(lock map[string]any) { declarationAdmissionR3Supplement(lock)["openapi"] = "2.0.0" },
		},
		{
			name: "primary unsupported swagger",
			mutate: func(lock map[string]any) {
				rest := declarationAdmissionR3REST(lock)
				delete(rest, "openapi")
				rest["swagger"] = "3.0"
			},
		},
		{
			name: "supplement unsupported swagger",
			mutate: func(lock map[string]any) {
				supplement := declarationAdmissionR3Supplement(lock)
				delete(supplement, "swagger")
				supplement["swagger"] = "3.0"
			},
		},
		{
			name:   "primary simultaneous openapi and swagger",
			mutate: func(lock map[string]any) { declarationAdmissionR3REST(lock)["swagger"] = "2.0" },
		},
		{
			name:   "supplement simultaneous openapi and swagger",
			mutate: func(lock map[string]any) { declarationAdmissionR3Supplement(lock)["openapi"] = "3.1.0" },
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			lock := declarationAdmissionR3LegacyReferenceLock()
			test.mutate(lock)
			if err := declarationAdmissionR3Parse(lock); err == nil {
				t.Fatal("mapping reader accepted invalid source-form pin")
			}
		})
	}
}

func TestDeclarationAdmissionMappingReaderR3AcceptsClosedVariants(t *testing.T) {
	tests := []struct {
		name string
		lock func() map[string]any
	}{
		{
			name: "ordinary schema v1 with opaque source contract and operation",
			lock: func() map[string]any { return declarationAdmissionR3LegacyOrdinaryLock(1) },
		},
		{
			name: "ordinary schema v2 with opaque source contract and operation",
			lock: func() map[string]any { return declarationAdmissionR3LegacyOrdinaryLock(2) },
		},
		{
			name: "legacy source reference with malformed retention",
			lock: declarationAdmissionR3LegacyReferenceLock,
		},
	}
	for _, kind := range []string{"openapi", "rendered_reference", "bundle", "source_reference", "unavailable"} {
		kind := kind
		tests = append(tests, struct {
			name string
			lock func() map[string]any
		}{
			name: "schema v3 " + kind + " document with malformed retention",
			lock: func() map[string]any { return declarationAdmissionR3V3Lock(kind) },
		})
	}
	tests = append(tests, struct {
		name string
		lock func() map[string]any
	}{
		name: "schema v3 accepts absent OpenAPI kind as its compatibility default",
		lock: func() map[string]any {
			lock := declarationAdmissionR3V3Lock("openapi")
			delete(lock["rest"].(map[string]any)["source_documents"].([]any)[0].(map[string]any), "kind")
			return lock
		},
	})
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := declarationAdmissionR3Parse(test.lock()); err != nil {
				t.Fatalf("mapping reader rejected closed supported variant: %v", err)
			}
		})
	}
}

func declarationAdmissionR3Parse(lock map[string]any) error {
	raw, err := json.Marshal(lock)
	if err != nil {
		return err
	}
	_, err = parseDeclarationAdmissionSourceLock(raw, "fixture")
	return err
}

func declarationAdmissionR3LegacyOrdinaryLock(version int) map[string]any {
	return map[string]any{
		"schema_version": version,
		"connector":      "fixture",
		"source_contract": map[string]any{
			"opaque": []any{"mapping reader must not decode this provider leaf"},
		},
		"rest": map[string]any{
			"source_url": "https://docs.polymetrics.invalid/fixture/openapi.json",
			"sha256":     map[string]any{"malformed": "retention is non-binding"},
			"bytes":      "malformed-retention-byte-count",
			"openapi":    "3.0.3",
			"operations": []any{map[string]any{
				"id": "fixture.rest.list-widgets", "protocol": "rest", "method": "GET", "path": "/widgets",
				"operation_id": "listWidgets", "deprecated": false, "source_location": `paths["/widgets"].get`,
				"source_operation": map[string]any{"opaque": []any{"provider-specific", true}},
			}},
		},
		"counts": map[string]any{"rest": 1, "graphql_query": 0, "graphql_mutation": 0, "total": 1},
	}
}

func declarationAdmissionR3LegacyReferenceLock() map[string]any {
	return map[string]any{
		"schema_version":   2,
		"connector":        "fixture",
		"operations_found": map[string]any{"rest": 2, "graphql_query": 0, "graphql_mutation": 0, "total": 2},
		"coverage_confidence": map[string]any{
			"level": sourceImportLegacySourceReferenceKind,
			"basis": "provider sources are cited without a retained execution contract",
		},
		"source_contract": []any{"opaque", map[string]any{"provider": "leaf"}},
		"rest": map[string]any{
			"source_kind":      sourceImportLegacySourceReferenceKind,
			"source_url":       "https://docs.polymetrics.invalid/fixture/openapi.json",
			"sha256":           map[string]any{"malformed": "retention is non-binding"},
			"bytes":            "malformed-retention-byte-count",
			"openapi":          "3.0.3",
			"operation_counts": map[string]any{"GET": 1, "POST": 1},
			"supplements": []any{map[string]any{
				"source_url":      "https://docs.polymetrics.invalid/fixture/supplement",
				"sha256":          []any{"malformed", "retention"},
				"bytes":           map[string]any{"malformed": true},
				"swagger":         "2.0",
				"source_location": "Supplemental provider reference",
				"operation_count": 1,
			}},
			"operations": []any{
				map[string]any{
					"id": "fixture.rest.list-widgets", "protocol": "rest", "method": "GET", "path": "/widgets",
					"operation_id": "listWidgets", "deprecated": false, "source_location": `paths["/widgets"].get`,
					"source_url":       "https://docs.polymetrics.invalid/fixture/openapi.json",
					"source_operation": []any{"opaque provider operation"},
				},
				map[string]any{
					"id": "fixture.rest.create-widget", "protocol": "rest", "method": "POST", "path": "/widgets",
					"operation_id": "createWidget", "deprecated": false, "source_location": "Supplemental provider reference",
					"source_url": "https://docs.polymetrics.invalid/fixture/supplement",
				},
			},
		},
		"counts": map[string]any{"rest": 2, "graphql_query": 0, "graphql_mutation": 0, "total": 2},
	}
}

func declarationAdmissionR3V3Lock(kind string) map[string]any {
	document := map[string]any{
		"id": "primary",
		"artifact": map[string]any{
			"source_url": "https://captures.polymetrics.invalid/fixture/openapi.json",
			"sha256":     map[string]any{"malformed": "retention is non-binding"},
			"bytes":      "malformed-retention-byte-count",
		},
		"published_source": map[string]any{
			"source_url":  "https://docs.polymetrics.invalid/fixture/reference",
			"capture_url": map[string]any{"malformed": "retention is non-binding"},
			"sha256":      []any{"malformed", "retention"},
			"bytes":       "malformed-retention-byte-count",
			"adapter":     map[string]any{"malformed": "retention is non-binding"},
		},
	}
	operation := map[string]any{
		"id": "fixture.rest.list-widgets", "protocol": "rest", "method": "GET", "path": "/widgets",
		"operation_id": "listWidgets", "deprecated": false, "source_location": "#widgets",
		"source_operation": map[string]any{"opaque": []any{"provider-specific", true}},
	}
	rest := map[string]any{
		"retrieval":        "provider reference inventory",
		"openapi":          []any{},
		"source_documents": []any{document},
	}
	switch kind {
	case "openapi":
		document["kind"] = "openapi"
		document["artifact"].(map[string]any)["openapi"] = "3.0.3"
		rest["openapi"] = []any{"3.0.3"}
		document["operations"] = []any{operation}
	case "rendered_reference":
		document["kind"] = kind
		document["content_type"] = "text/html"
		operation["citation_url"] = "https://docs.polymetrics.invalid/fixture/reference#widgets"
		document["operations"] = []any{operation}
		rest["coverage_confidence"] = map[string]any{"level": "partial", "basis": "rendered provider reference"}
	case "bundle":
		document["kind"] = kind
		document["content_type"] = "application/gzip"
		document["operations"] = []any{operation}
		rest["coverage_confidence"] = map[string]any{"level": "partial", "basis": "provider archive inventory"}
	case "source_reference":
		document["kind"] = kind
		document["artifact"] = map[string]any{"source_url": "", "sha256": map[string]any{"malformed": true}, "bytes": "ignored"}
		document["published_source"] = map[string]any{"source_url": "", "capture_url": "ignored", "sha256": map[string]any{"malformed": true}, "bytes": "ignored", "adapter": "ignored"}
		document["source_reference"] = map[string]any{
			"source_url": "https://docs.polymetrics.invalid/fixture/reference",
			"sha256":     map[string]any{"malformed": "retention is non-binding"},
			"bytes":      "malformed-retention-byte-count",
			"openapi":    "3.1.0",
		}
		document["operations"] = []any{operation}
		rest["coverage_confidence"] = map[string]any{"level": "source_reference", "basis": "provider source reference without executable capture"}
	case "unavailable":
		document["kind"] = kind
		document["unavailable_reason"] = "provider has no usable public API reference"
		document["operations"] = []any{}
		rest["coverage_confidence"] = map[string]any{"level": "unavailable", "basis": "provider source is unavailable"}
	default:
		panic("unsupported R3 test document kind: " + kind)
	}
	count := 1
	if kind == "unavailable" {
		count = 0
	}
	return map[string]any{
		"schema_version":  3,
		"connector":       "fixture",
		"source_contract": map[string]any{"opaque": []any{"provider-specific", true}},
		"rest":            rest,
		"counts":          map[string]any{"rest": count, "graphql_query": 0, "graphql_mutation": 0, "total": count},
	}
}

func declarationAdmissionR3REST(lock map[string]any) map[string]any {
	return lock["rest"].(map[string]any)
}

func declarationAdmissionR3LegacyOperation(lock map[string]any) map[string]any {
	return declarationAdmissionR3REST(lock)["operations"].([]any)[0].(map[string]any)
}

func declarationAdmissionR3Supplement(lock map[string]any) map[string]any {
	return declarationAdmissionR3REST(lock)["supplements"].([]any)[0].(map[string]any)
}

func declarationAdmissionR3V3Operation(lock map[string]any) map[string]any {
	return declarationAdmissionR3REST(lock)["source_documents"].([]any)[0].(map[string]any)["operations"].([]any)[0].(map[string]any)
}

func declarationAdmissionR3ValueName(value any) string {
	if value == nil {
		return "null"
	}
	return "empty"
}
