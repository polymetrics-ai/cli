package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBatch1DeferredManifestReconcilesPublishedCensus(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "docs", "architecture", "batch1-source-operation-mapping-manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := decodeBatch1DeferredManifest(raw)
	if err != nil {
		t.Fatalf("decode manifest: %v", err)
	}
	stats, err := manifest.reconcilePublishedCensus()
	if err != nil {
		t.Fatalf("reconcile manifest: %v", err)
	}
	if stats.SourceOperations != 4341 || stats.Runnable != 846 || stats.Declarable != 1585 || stats.Deferred != 1910 {
		t.Fatalf("census = %+v, want source_operations=4341 runnable=846 declarable=1585 deferred=1910", stats)
	}
	if got := len(manifest.deferredRecords()); got != 1910 {
		t.Fatalf("deferred records = %d, want 1910", got)
	}
}

func TestBatch1DeferredManifestRejectsUnsafeOrAmbiguousDeferredRecords(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*batch1DeferredManifest)
		want   string
	}{
		{
			name: "duplicate source identity",
			mutate: func(manifest *batch1DeferredManifest) {
				manifest.Records = append(manifest.Records, manifest.Records[0])
			},
			want: "duplicate record_key",
		},
		{
			name: "missing provider citation",
			mutate: func(manifest *batch1DeferredManifest) {
				manifest.Records[0].Source.URL = ""
			},
			want: "source_url",
		},
		{
			name: "generic HTTP target",
			mutate: func(manifest *batch1DeferredManifest) {
				manifest.Records[0].CanonicalTarget.Endpoint.Path = "https://provider.example/v1/widgets"
			},
			want: "connector-relative",
		},
		{
			name: "policy-only foundation",
			mutate: func(manifest *batch1DeferredManifest) {
				manifest.Records[0].MissingImplementation.Foundation = "delete"
			},
			want: "policy-only",
		},
		{
			name: "multiple primary gaps",
			mutate: func(manifest *batch1DeferredManifest) {
				manifest.Records[0].MissingImplementation.AdditionalFoundation = "typed_request_body"
			},
			want: "exactly one",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest := batch1DeferredFixture(t)
			tt.mutate(&manifest)
			if err := manifest.validate(); err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("validate error = %v, want %q", err, tt.want)
			}
		})
	}
}

func batch1DeferredFixture(t *testing.T) batch1DeferredManifest {
	t.Helper()
	const fixture = `{
  "schema_version": 1,
  "invariants": {
    "source_operations": 1,
    "runnable": 0,
    "declarable": 0,
    "deferred": 1,
    "every_source_operation_exactly_once": true,
    "every_record_has_cli_path": true,
    "deferred_has_exactly_one_concrete_missing_component": true,
    "zero_denominator_forbidden": true,
    "policy_only_terms_forbidden_as_components": ["delete", "risk"]
  },
  "records": [{
    "provider": "acme",
    "record_key": "acme:widgets.get",
    "mapping_state": "deferred",
    "declaration_state": "deferred",
    "lane": "direct_read",
    "intended_cli_path": {"path": "widgets view", "source": "manifest_reservation", "current_availability": "reserved"},
    "canonical_target": {"operation_key": "rest:get:/v1/widgets/{id}", "endpoint": {"method": "GET", "path": "/v1/widgets/{id}"}},
    "source": {
      "protocol": "rest",
      "operation_id": "acme.rest.widgets.get",
      "provider_operation_id": "widgets.get",
      "method": "GET",
      "path": "/v1/widgets/{id}",
      "source_lock": "internal/connectors/defs/acme/sources/acme-operation-source-lock.json",
      "source_url": "https://provider.example/openapi.json",
      "source_location": "paths[\"/v1/widgets/{id}\"].get"
    },
    "missing_implementation": {
      "component": "source_descriptor",
      "foundation": "provider_dialect_foundation",
      "evidence": "provider parser has no closed descriptor",
      "projection_prerequisite": {"kind": "runtime_deferred_command_projection", "same_cli_path_required": true, "status": "required"}
    }
  }]
}`
	var manifest batch1DeferredManifest
	if err := json.Unmarshal([]byte(fixture), &manifest); err != nil {
		t.Fatal(err)
	}
	if err := manifest.validate(); err != nil {
		t.Fatalf("fixture: %v", err)
	}
	return manifest
}
