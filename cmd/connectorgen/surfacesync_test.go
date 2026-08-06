package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// writeSyncBundle lays down a minimal one-command bundle and returns its dir.
func writeSyncBundle(t *testing.T, cli, ops any) string {
	t.Helper()
	dir := t.TempDir()
	for name, doc := range map[string]any{"cli_surface.json": cli, "operations.json": ops} {
		data, err := json.MarshalIndent(doc, "", "  ")
		if err != nil {
			t.Fatalf("encode %s: %v", name, err)
		}
		if err := os.WriteFile(filepath.Join(dir, name), append(data, '\n'), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	return dir
}

func readSyncedCommand(t *testing.T, dir string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, "cli_surface.json"))
	if err != nil {
		t.Fatalf("read cli_surface.json: %v", err)
	}
	var doc struct {
		Commands []map[string]any `json:"commands"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("decode cli_surface.json: %v", err)
	}
	if len(doc.Commands) != 1 {
		t.Fatalf("cli_surface.json has %d commands, want 1", len(doc.Commands))
	}
	return doc.Commands[0]
}

func directReadBundle(apiSurface []any, outputPolicy string, mapsTo string) (any, any) {
	command := map[string]any{
		"path":         "issues view",
		"intent":       "direct_read",
		"availability": "implemented",
		"operation":    "acme.issue",
		"flags":        []any{map[string]any{"name": "issue-id", "type": "string"}},
	}
	if apiSurface != nil {
		command["api_surface"] = apiSurface
	}
	if outputPolicy != "" {
		command["output_policy"] = outputPolicy
	}
	if mapsTo != "" {
		command["flags"] = []any{map[string]any{"name": "issue-id", "type": "string", "maps_to": mapsTo}}
	}
	cli := map[string]any{"usage": "pm acme <command>", "commands": []any{command}}
	ops := map[string]any{"operations": []any{map[string]any{
		"id":   "acme.issue",
		"kind": "rest_read",
		"rest": map[string]any{"method": "GET", "path": "/issues/{issue_id}", "max_bytes": 1048576},
	}}}
	return cli, ops
}

func directWriteBundle(apiSurface []any, outputPolicy string, mapsTo string) (any, any) {
	command := map[string]any{
		"path":         "issues close",
		"intent":       "direct_write",
		"availability": "implemented",
		"operation":    "acme.issue.close",
		"flags":        []any{map[string]any{"name": "issue-id", "type": "string"}},
	}
	if apiSurface != nil {
		command["api_surface"] = apiSurface
	}
	if outputPolicy != "" {
		command["output_policy"] = outputPolicy
	}
	if mapsTo != "" {
		command["flags"] = []any{map[string]any{"name": "issue-id", "type": "string", "maps_to": mapsTo}}
	}
	cli := map[string]any{"usage": "pm acme <command>", "commands": []any{command}}
	ops := map[string]any{"operations": []any{map[string]any{
		"id":            "acme.issue.close",
		"kind":          "rest_write",
		"output_policy": "write_result_redacted",
		"rest": map[string]any{
			"method": "POST",
			"path":   "/issues/{issue_id}/close",
		},
	}}}
	return cli, ops
}

// TestSyncBundleReportsDivergentAPISurface is the check the tool documents but
// did not perform: a hand-edited or stale api_surface previously passed --check
// clean, because syncBundle only ever filled an ABSENT field. The executor
// reads the operation's own path, so help, docs and the website would advertise
// one endpoint while the command called another.
func TestSyncBundleReportsDivergentAPISurface(t *testing.T) {
	stale := []any{map[string]any{"method": "GET", "path": "/issues/{issue_id}/stale"}}
	cli, ops := directReadBundle(stale, "json_redacted", "path.issue_id")
	dir := writeSyncBundle(t, cli, ops)

	stats, err := syncBundle(dir, true)
	if err != nil {
		t.Fatalf("syncBundle check: %v", err)
	}
	if stats.Corrected.APISurface != 1 {
		t.Fatalf("corrected api_surface = %d, want 1 (stats: %+v)", stats.Corrected.APISurface, stats)
	}
	if stats.Filled.total() != 0 {
		t.Fatalf("filled = %s, want nothing filled", stats.Filled)
	}
	// Check mode must not write.
	if got := readSyncedCommand(t, dir)["api_surface"]; !endpointPathIs(got, "/issues/{issue_id}/stale") {
		t.Fatalf("check mode rewrote the bundle: %v", got)
	}

	if _, err := syncBundle(dir, false); err != nil {
		t.Fatalf("syncBundle write: %v", err)
	}
	if got := readSyncedCommand(t, dir)["api_surface"]; !endpointPathIs(got, "/issues/{issue_id}") {
		t.Fatalf("write mode did not correct api_surface: %v", got)
	}
}

func TestSyncBundleReportsDivergentFlagMapsTo(t *testing.T) {
	surface := []any{map[string]any{"method": "GET", "path": "/issues/{issue_id}"}}
	cli, ops := directReadBundle(surface, "json_redacted", "query.issue_id")
	dir := writeSyncBundle(t, cli, ops)

	stats, err := syncBundle(dir, true)
	if err != nil {
		t.Fatalf("syncBundle check: %v", err)
	}
	if stats.Corrected.FlagMapsTo != 1 || stats.Filled.FlagMapsTo != 0 {
		t.Fatalf("flag maps_to stats = filled %d / corrected %d, want 0 / 1", stats.Filled.FlagMapsTo, stats.Corrected.FlagMapsTo)
	}

	if _, err := syncBundle(dir, false); err != nil {
		t.Fatalf("syncBundle write: %v", err)
	}
	flags, _ := readSyncedCommand(t, dir)["flags"].([]any)
	if len(flags) != 1 {
		t.Fatalf("flags = %v, want one", flags)
	}
	flag, _ := flags[0].(map[string]any)
	if flag["maps_to"] != "path.issue_id" {
		t.Fatalf("maps_to = %v, want path.issue_id", flag["maps_to"])
	}
}

// A rest_write command has no author-selected output policy: its declared
// operation is the one source of truth for both the response contract and the
// endpoint the executor will call. Keeping either hand-edited means a command
// can pass help/docs review while its plan dispatches something else.
func TestSyncBundleDirectWriteDerivesOperationContract(t *testing.T) {
	stale := []any{map[string]any{"method": "POST", "path": "/issues/{issue_id}/close/stale"}}
	cli, ops := directWriteBundle(stale, "json_redacted", "query.issue_id")
	dir := writeSyncBundle(t, cli, ops)

	stats, err := syncBundle(dir, true)
	if err != nil {
		t.Fatalf("syncBundle check: %v", err)
	}
	if stats.Corrected.APISurface != 1 || stats.Corrected.OutputPolicy != 1 ||
		stats.Corrected.FlagMapsTo != 1 || stats.Filled.MaxBytes != 1 {
		t.Fatalf("direct_write stats = %+v, want endpoint/policy/flag corrections and max_bytes fill", stats)
	}

	if _, err := syncBundle(dir, false); err != nil {
		t.Fatalf("syncBundle write: %v", err)
	}
	command := readSyncedCommand(t, dir)
	if got := command["output_policy"]; got != "write_result_redacted" {
		t.Fatalf("output_policy = %v, want declared write_result_redacted", got)
	}
	if got := command["api_surface"]; !endpointPathIs(got, "/issues/{issue_id}/close") {
		t.Fatalf("api_surface = %v, want operation endpoint", got)
	}
	if got := command["flags"]; !flagMapsTo(got, "path.issue_id") {
		t.Fatalf("flags = %v, want path.issue_id", got)
	}
}

// An unsupported output_policy is divergence: the runtime refuses it. A
// supported one is a deliberate choice and is left exactly as authored.
func TestSyncBundleOutputPolicyCorrectsOnlyUnsupportedValues(t *testing.T) {
	surface := []any{map[string]any{"method": "GET", "path": "/issues/{issue_id}"}}

	t.Run("unsupported is corrected", func(t *testing.T) {
		cli, ops := directReadBundle(surface, "json", "path.issue_id")
		dir := writeSyncBundle(t, cli, ops)
		stats, err := syncBundle(dir, false)
		if err != nil {
			t.Fatalf("syncBundle: %v", err)
		}
		if stats.Corrected.OutputPolicy != 1 {
			t.Fatalf("corrected output_policy = %d, want 1", stats.Corrected.OutputPolicy)
		}
		if got := readSyncedCommand(t, dir)["output_policy"]; got != defaultDirectReadOutputPolicy {
			t.Fatalf("output_policy = %v, want %s", got, defaultDirectReadOutputPolicy)
		}
	})

	t.Run("supported is untouched", func(t *testing.T) {
		cli, ops := directReadBundle(surface, "clinical_json_redacted", "path.issue_id")
		dir := writeSyncBundle(t, cli, ops)
		stats, err := syncBundle(dir, true)
		if err != nil {
			t.Fatalf("syncBundle: %v", err)
		}
		if stats.total() != 0 {
			t.Fatalf("stats = %+v, want a clean bundle", stats)
		}
	})
}

// A binary download produces a file, so any output_policy on it is divergence.
func TestSyncBundleRemovesOutputPolicyFromBinaryDownload(t *testing.T) {
	cli := map[string]any{"usage": "pm acme <command>", "commands": []any{map[string]any{
		"path":          "artifact download",
		"intent":        "binary_download",
		"availability":  "implemented",
		"operation":     "acme.artifact",
		"output_policy": "json_redacted",
		"flags":         []any{map[string]any{"name": "artifact-id", "type": "string"}},
		"api_surface":   []any{map[string]any{"method": "GET", "path": "/artifacts/{artifact_id}"}},
	}}}
	ops := map[string]any{"operations": []any{map[string]any{
		"id":     "acme.artifact",
		"kind":   "binary_download",
		"binary": map[string]any{"method": "GET", "path": "/artifacts/{artifact_id}", "max_bytes": 104857600},
	}}}
	dir := writeSyncBundle(t, cli, ops)

	stats, err := syncBundle(dir, false)
	if err != nil {
		t.Fatalf("syncBundle: %v", err)
	}
	if stats.Corrected.OutputPolicy != 1 {
		t.Fatalf("corrected output_policy = %d, want 1", stats.Corrected.OutputPolicy)
	}
	command := readSyncedCommand(t, dir)
	if _, present := command["output_policy"]; present {
		t.Fatalf("binary download kept an output_policy: %v", command)
	}
	// Removing one key must not disturb the others.
	for _, want := range []string{"path", "intent", "availability", "operation", "flags", "api_surface"} {
		if _, ok := command[want]; !ok {
			t.Fatalf("removing output_policy dropped %q: %v", want, command)
		}
	}
	if got := command["flags"]; !flagMapsTo(got, "path.artifact_id") {
		t.Fatalf("binary download flag maps_to = %v, want path.artifact_id", got)
	}
}

// A present but non-positive rest.max_bytes is unusable, so it is corrected; a
// positive one is the operation's own declaration and is left alone.
func TestSyncBundleMaxBytesFillsAndCorrects(t *testing.T) {
	surface := []any{map[string]any{"method": "GET", "path": "/issues/{issue_id}"}}

	t.Run("absent is filled", func(t *testing.T) {
		cli, ops := directReadBundle(surface, "json_redacted", "path.issue_id")
		rest := ops.(map[string]any)["operations"].([]any)[0].(map[string]any)["rest"].(map[string]any)
		delete(rest, "max_bytes")
		dir := writeSyncBundle(t, cli, ops)
		stats, err := syncBundle(dir, true)
		if err != nil {
			t.Fatalf("syncBundle: %v", err)
		}
		if stats.Filled.MaxBytes != 1 || stats.Corrected.MaxBytes != 0 {
			t.Fatalf("max_bytes stats = filled %d / corrected %d, want 1 / 0", stats.Filled.MaxBytes, stats.Corrected.MaxBytes)
		}
	})

	t.Run("non-positive is corrected", func(t *testing.T) {
		cli, ops := directReadBundle(surface, "json_redacted", "path.issue_id")
		ops.(map[string]any)["operations"].([]any)[0].(map[string]any)["rest"].(map[string]any)["max_bytes"] = 0
		dir := writeSyncBundle(t, cli, ops)
		stats, err := syncBundle(dir, true)
		if err != nil {
			t.Fatalf("syncBundle: %v", err)
		}
		if stats.Corrected.MaxBytes != 1 || stats.Filled.MaxBytes != 0 {
			t.Fatalf("max_bytes stats = filled %d / corrected %d, want 0 / 1", stats.Filled.MaxBytes, stats.Corrected.MaxBytes)
		}
	})

	t.Run("positive is a deliberate declaration", func(t *testing.T) {
		cli, ops := directReadBundle(surface, "json_redacted", "path.issue_id")
		ops.(map[string]any)["operations"].([]any)[0].(map[string]any)["rest"].(map[string]any)["max_bytes"] = 4194304
		dir := writeSyncBundle(t, cli, ops)
		stats, err := syncBundle(dir, true)
		if err != nil {
			t.Fatalf("syncBundle: %v", err)
		}
		if stats.total() != 0 {
			t.Fatalf("stats = %+v, want a clean bundle", stats)
		}
	})
}

// A bundle that already agrees with its operations must report nothing, or
// --check would fail on every repository it is meant to protect.
func TestSyncBundleConsistentBundleIsClean(t *testing.T) {
	surface := []any{map[string]any{"method": "GET", "path": "/issues/{issue_id}"}}
	cli, ops := directReadBundle(surface, "json_redacted", "path.issue_id")
	dir := writeSyncBundle(t, cli, ops)

	stats, err := syncBundle(dir, true)
	if err != nil {
		t.Fatalf("syncBundle: %v", err)
	}
	if stats.total() != 0 {
		t.Fatalf("stats = %+v, want a clean bundle", stats)
	}
}

func TestSyncRuntimeOperationEndpointLedgerCreatesCompactProjection(t *testing.T) {
	root := t.TempDir()
	for path, file := range operationCLISurfaceBundleFS(validOperationCLISurfaceJSON(), validOperationsJSON()) {
		target := filepath.Join(root, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", target, err)
		}
		if err := os.WriteFile(target, file.Data, 0o644); err != nil {
			t.Fatalf("write %s: %v", target, err)
		}
	}

	stats, err := syncRuntimeOperationEndpointLedger(root, false)
	if err != nil {
		t.Fatalf("syncRuntimeOperationEndpointLedger: %v", err)
	}
	if !stats.Changed || stats.Entries != 1 {
		t.Fatalf("stats = %+v, want one written endpoint", stats)
	}

	path := filepath.Join(root, "operation_endpoint_ledger.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read runtime endpoint ledger: %v", err)
	}
	var ledger map[string][]map[string]any
	if err := json.Unmarshal(raw, &ledger); err != nil {
		t.Fatalf("decode runtime endpoint ledger: %v", err)
	}
	entries := ledger["cli-surface"]
	if len(entries) != 1 {
		t.Fatalf("ledger entries = %#v, want one", entries)
	}
	entry := entries[0]
	keys := make([]string, 0, len(entry))
	for key := range entry {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	if got, want := strings.Join(keys, ","), "kind,max_bytes,method,path"; got != want {
		t.Fatalf("ledger fields = %q, want %q", got, want)
	}
	if entry["method"] != "GET" || entry["path"] != "/widgets/{id}" || entry["kind"] != "rest_read" || entry["max_bytes"] != float64(1024) {
		t.Fatalf("ledger entry = %#v, want direct-read operation contract", entry)
	}

	stats, err = syncRuntimeOperationEndpointLedger(root, true)
	if err != nil {
		t.Fatalf("syncRuntimeOperationEndpointLedger --check: %v", err)
	}
	if stats.Changed {
		t.Fatalf("clean runtime endpoint ledger reported drift: %+v", stats)
	}
	if err := os.Remove(path); err != nil {
		t.Fatalf("remove runtime endpoint ledger: %v", err)
	}
	stats, err = syncRuntimeOperationEndpointLedger(root, true)
	if err != nil {
		t.Fatalf("syncRuntimeOperationEndpointLedger missing projection --check: %v", err)
	}
	if !stats.Changed {
		t.Fatalf("missing runtime endpoint ledger did not report drift: %+v", stats)
	}
	if err := os.WriteFile(path, []byte(`{"cli-surface":[{"unexpected":true}]}`), 0o644); err != nil {
		t.Fatalf("write malformed runtime endpoint ledger: %v", err)
	}
	stats, err = syncRuntimeOperationEndpointLedger(root, false)
	if err != nil {
		t.Fatalf("syncRuntimeOperationEndpointLedger malformed projection: %v", err)
	}
	if !stats.Changed {
		t.Fatalf("malformed runtime endpoint ledger did not report replacement: %+v", stats)
	}
}

func endpointPathIs(raw any, want string) bool {
	entries, ok := raw.([]any)
	if !ok || len(entries) != 1 {
		return false
	}
	endpoint, ok := entries[0].(map[string]any)
	return ok && endpoint["path"] == want
}

func flagMapsTo(raw any, want string) bool {
	flags, ok := raw.([]any)
	if !ok || len(flags) != 1 {
		return false
	}
	flag, ok := flags[0].(map[string]any)
	return ok && flag["maps_to"] == want
}
