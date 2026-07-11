package certify

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestCatalogStreamSpecsFromStreams(t *testing.T) {
	streams := []any{
		map[string]any{
			"name":          "issues",
			"primary_key":   []any{"node_id"},
			"cursor_fields": []any{"updated_at"},
		},
		map[string]any{
			"name":          "pull_requests",
			"primary_key":   []any{"id"},
			"cursor_fields": []any{"merged_at", "updated_at"},
		},
		map[string]any{
			"name": "no_keys",
		},
	}

	specs := catalogStreamSpecsFromStreams(streams)
	if len(specs) != 3 {
		t.Fatalf("len(specs) = %d, want 3", len(specs))
	}
	if specs[0] != (streamSpec{Name: "issues", PrimaryKey: "node_id", CursorField: "updated_at"}) {
		t.Fatalf("specs[0] = %+v", specs[0])
	}
	if specs[1] != (streamSpec{Name: "pull_requests", PrimaryKey: "id", CursorField: "merged_at"}) {
		t.Fatalf("specs[1] = %+v", specs[1])
	}
	if specs[2] != (streamSpec{Name: "no_keys"}) {
		t.Fatalf("specs[2] = %+v", specs[2])
	}
}

func TestFullSweepNamesAreStreamScoped(t *testing.T) {
	rc := &runContext{opts: Options{Stream: "customers"}}
	if got := rc.liveConnectionName(); got != liveConnectionName {
		t.Fatalf("default liveConnectionName = %q, want %q", got, liveConnectionName)
	}
	if got := rc.fileCredentialName(); got != fileCredentialName {
		t.Fatalf("default fileCredentialName = %q, want %q", got, fileCredentialName)
	}
	if got := rc.captureConnectionName("full_refresh_overwrite"); got != captureConnectionPrefix+"full_refresh_overwrite" {
		t.Fatalf("default captureConnectionName = %q", got)
	}

	rc.currentStream = "pull requests"
	if got := rc.liveConnectionName(); got != "cert_live_pull_requests" {
		t.Fatalf("stream-scoped liveConnectionName = %q", got)
	}
	if got := rc.fileCredentialName(); got != "cert-file_pull_requests" {
		t.Fatalf("stream-scoped fileCredentialName = %q", got)
	}
	if got := rc.captureConnectionName("incremental_append_deduped"); got != "cert_capture_incremental_append_deduped_pull_requests" {
		t.Fatalf("stream-scoped captureConnectionName = %q", got)
	}
}

func TestFullSweepStreamSpecsPreserveCatalogStreamsWithoutPKOrCursor(t *testing.T) {
	rc := &runContext{catalogStreamSpecs: []streamSpec{{Name: "branches"}}}
	specs := rc.fullSweepStreamSpecs()
	if len(specs) != 1 {
		t.Fatalf("len(specs) = %d, want 1", len(specs))
	}
	if specs[0].Name != "branches" || specs[0].PrimaryKey != "" || specs[0].CursorField != "" {
		t.Fatalf("spec = %+v, want branches with empty primary key and cursor", specs[0])
	}
}

func TestFullSweepStreamSpecsFallbackToSelectedStream(t *testing.T) {
	rc := &runContext{opts: Options{Stream: "customers"}}
	specs := rc.fullSweepStreamSpecs()
	if len(specs) != 1 {
		t.Fatalf("len(specs) = %d, want 1", len(specs))
	}
	if specs[0].Name != "customers" || specs[0].PrimaryKey != "id" || specs[0].CursorField != "updated_at" {
		t.Fatalf("fallback spec = %+v", specs[0])
	}
}

func TestInspectStreamSpecsSeedBootstrapCursor(t *testing.T) {
	envelope := map[string]any{
		"manifest": map[string]any{
			"streams": []any{
				map[string]any{
					"name":          "attachments",
					"primary_key":   []any{"id"},
					"cursor_fields": []any{"updatedAt"},
				},
			},
		},
	}

	specs := streamSpecsFromInspectEnvelope(envelope)
	if len(specs) != 1 {
		t.Fatalf("len(specs) = %d, want 1", len(specs))
	}
	rc := &runContext{opts: Options{Connector: "twenty", Stream: "attachments"}, catalogStreamSpecs: specs}
	if got := rc.cursorField(); got != "updatedAt" {
		t.Fatalf("cursorField() = %q, want Twenty catalog cursor updatedAt before bootstrap connection", got)
	}
}

func TestFullSweepArtifactNamesAreBoundedForLongStream(t *testing.T) {
	const stream = "message_channel_message_association_message_folders"
	rc := &runContext{
		root:          t.TempDir(),
		opts:          Options{Connector: "twenty"},
		currentStream: stream,
	}
	rc.capturePath = filepath.Join(rc.root, rc.captureFileName()+".jsonl")

	components := map[string]string{
		"live_connection":    rc.liveConnectionName(),
		"live_table":         rc.liveTableName(),
		"capture_connection": rc.captureConnectionName("full_refresh_overwrite_deduped"),
		"capture_stream":     rc.captureStreamName(),
		"capture_table":      rc.captureTableName("cert_overwrite_deduped"),
		"flow_name":          rc.flowName(),
		"flow_table":         rc.flowTable(),
		"flow_connection":    rc.flowConnectionName(),
	}
	for label, component := range components {
		if component == "" {
			t.Fatalf("%s is empty", label)
		}
		if strings.ContainsAny(component, `/\\`) {
			t.Fatalf("%s = %q contains a path separator", label, component)
		}
		if len(component) > 96 {
			t.Fatalf("%s length = %d, want <= 96: %q", label, len(component), component)
		}
	}

	rawName := rc.captureConnectionName("full_refresh_overwrite_deduped") + "__" + rc.captureStreamName() + "__" + rc.captureTableName("cert_overwrite_deduped") + ".jsonl.tmp"
	if len(rawName) > 255 {
		t.Fatalf("raw capture path component length = %d, want <= 255: %q", len(rawName), rawName)
	}
}
