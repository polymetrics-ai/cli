package engine

import (
	"strings"
	"testing"
	"testing/fstest"

	"polymetrics.ai/internal/connectors"
)

func TestBundleLoadsDefinitionOwnedPollingWatermarkDescriptor(t *testing.T) {
	fsys := pollingWatermarkBundleFS("acme")
	fsys["acme/polling_watermark.json"] = &fstest.MapFile{Data: []byte(validPollingWatermarkDefinitionJSON)}

	bundle, err := Load(fsys, "acme")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if bundle.PollingWatermark == nil {
		t.Fatal("bundle omitted polling_watermark.json declaration")
	}
	definition := New(bundle, nil).Definition()
	if definition.PollingWatermark == nil {
		t.Fatal("Definition omitted polling watermark declaration")
	}
	definition.PollingWatermark.Source.Modes[0] = "changed"
	if got := bundle.PollingWatermark.Source.Modes[0]; got != "incremental_upsert" {
		t.Fatalf("definition alias mutated authored descriptor mode = %q", got)
	}
}

func TestBundleRejectsUnsafePollingWatermarkDefinition(t *testing.T) {
	fsys := pollingWatermarkBundleFS("acme")
	invalid := strings.Replace(validPollingWatermarkDefinitionJSON, `"codec": "rfc3339_nano"`, `"codec": "float64"`, 1)
	fsys["acme/polling_watermark.json"] = &fstest.MapFile{Data: []byte(invalid)}

	_, err := Load(fsys, "acme")
	if err == nil || !strings.Contains(err.Error(), "cursor codec must preserve values losslessly") {
		t.Fatalf("Load error = %v, want non-lossless cursor refusal", err)
	}
}

func TestBundleLoadsNonImplementedPollingWatermarkWithoutExecutionClaim(t *testing.T) {
	fsys := pollingWatermarkBundleFS("acme")
	fsys["acme/polling_watermark.json"] = &fstest.MapFile{Data: []byte(`{
  "status": "planned",
  "reason": "native polling executor has not been registered"
}`)}

	bundle, err := Load(fsys, "acme")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if bundle.PollingWatermark == nil || bundle.PollingWatermark.Status != connectors.PollingWatermarkStatusPlanned {
		t.Fatalf("planned polling declaration = %+v, want retained non-executable declaration", bundle.PollingWatermark)
	}
}

func TestBundleRejectsPollingWatermarkForAPIBundle(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	fsys["acme/polling_watermark.json"] = &fstest.MapFile{Data: []byte(validPollingWatermarkDefinitionJSON)}

	_, err := Load(fsys, "acme")
	if err == nil {
		t.Fatal("Load succeeded, want API-bundle polling watermark refusal")
	}
	if got, want := err.Error(), `load bundle acme: polling_watermark.json requires metadata integration_type "database"`; got != want {
		t.Fatalf("Load error = %q, want %q", got, want)
	}
}

func pollingWatermarkBundleFS(name string) fstest.MapFS {
	fsys := fullValidBundleFS(name)
	fsys[name+"/metadata.json"] = &fstest.MapFile{Data: []byte(metadataWithIntegrationType(name, "database"))}
	return fsys
}

const validPollingWatermarkDefinitionJSON = `{
  "status": "implemented",
  "source": {
    "executor": {"family": "native_database", "id": "acme-polling-source-v1"},
    "object": {"kind": "relation"},
    "read": {
      "kind": "keyset",
      "max_page_size": 100,
      "max_pages": 2,
      "max_requests": 2,
      "stable_traversal": true,
      "predicate": "lexicographic_tuple"
    },
    "snapshot": {"kind": "transaction_snapshot"},
    "cursor": {"codec": "rfc3339_nano", "type": "timestamp", "precision": "nanosecond"},
    "ordering": {
      "watermark": {"catalog_field": "updated_at", "ascending": true},
      "tie_breaker": {"catalog_field": "id", "ascending": true, "unique": true}
    },
    "mutation": {"mutable": true, "commit_ordered": true, "bounded_overlap": true},
    "identity": {"engine": "acme-native", "account_scope": "account", "object_scope": "widgets"},
    "schema_compatibility": "exact_fingerprint",
    "delete_visibility": "hard_delete_invisible",
    "modes": ["incremental_upsert", "incremental_dedupe_history"]
  },
  "target": {
    "executor": {"family": "native_database", "id": "acme-polling-apply-v1"},
    "max_batch_records": 100,
    "staging": "staging_replace_supported",
    "stable_key_mapping": ["id"],
    "conditional_order_fence": true,
    "transaction": "required",
    "partial_result": "rollback",
    "retry_safe_close_and_insert": true,
    "validity_window": "supported",
    "strategies": ["merge", "dedupe_history"]
  }
}`

var _ = connectors.PollingWatermarkStatusImplemented
