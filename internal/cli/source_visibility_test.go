package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
	"polymetrics.ai/internal/connectors/manifestindex"
	"reflect"
	"sort"
	"strings"
	"sync/atomic"
	"testing"
)

// These requests enter the real public router and use the compiled inventory.
// Expected identities and outcomes are retained-source facts, not output-derived.
func TestSourceVisibilityPublicSelection162(t *testing.T) {
	for _, tc := range []struct {
		name, connector, id, lane, code string
		exit                            int
	}{
		{"mapped_remove", "asana", "asana.rest.removeCustomFieldSettingForGoal", "direct_write", "source_mapping_unproven", 1},
		{"excluded_read", "asana", "asana.rest.removeCustomFieldSettingForGoal", "direct_read", "source_lane_not_applicable", 3},
		{"scoped_gap", "vercel", "vercel.rest.createWebhook", "sync_transport", "missing_foundation", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var out, diag bytes.Buffer
			got := Run([]string{"connectors", "inspect", tc.connector, "--inventory", "primary", "--source-id", tc.id, "--lane", tc.lane, "--preflight", "--json", "--root", t.TempDir()}, &out, &diag)
			var response map[string]json.RawMessage
			if err := json.Unmarshal(out.Bytes(), &response); err != nil {
				t.Fatalf("public JSON: %v; stderr=%s", err, &diag)
			}
			if got != tc.exit || !bytes.Contains(response["source_selection"], []byte(tc.code)) || !bytes.Contains(response["source_selection"], []byte(tc.id)) || !strings.Contains(diag.String(), "error:") {
				t.Fatalf("source selection got exit=%d kind=%s selection=%s stderr=%s; want exit=%d code=%s id=%s", got, response["kind"], response["source_selection"], &diag, tc.exit, tc.code, tc.id)
			}
		})
	}
}

func TestSourceVisibilityCompletePublicMembership162(t *testing.T) {
	raw, err := os.ReadFile("../../cmd/connectorgen/testdata/source_lanes/batch1-expected-ids.json")
	if err != nil {
		t.Fatal(err)
	}
	var expected struct {
		Keys []connectors.SourceOperationKey `json:"keys"`
	}
	if err = json.Unmarshal(raw, &expected); err != nil {
		t.Fatal(err)
	}
	want := map[connectors.SourceCellSelection]bool{}
	names := map[string]bool{}
	for _, key := range expected.Keys {
		names[key.Connector] = true
		for _, lane := range []connectors.SourceLane{"direct_read", "direct_write", "binary_download", "binary_upload", "etl", "reverse_etl", "sync_transport"} {
			selection := connectors.SourceCellSelection{Source: key, Lane: lane}
			if want[selection] {
				t.Fatal("independent fixture duplicate")
			}
			want[selection] = true
		}
	}
	got := map[connectors.SourceCellSelection]bool{}
	ordered := []string{}
	for name := range names {
		ordered = append(ordered, name)
	}
	sort.Strings(ordered)
	for _, name := range ordered {
		var out, diag bytes.Buffer
		if code := Run([]string{"connectors", "inspect", name, "--sources", "--json", "--root", t.TempDir()}, &out, &diag); code != 0 {
			t.Fatalf("catalog %s: %d %s", name, code, &diag)
		}
		var response struct {
			Kind             string                      `json:"kind"`
			Catalog          connectors.SourceVisibility `json:"catalog"`
			ExecutionChecked bool                        `json:"execution_checked"`
		}
		if err = json.Unmarshal(out.Bytes(), &response); err != nil {
			t.Fatal(err)
		}
		if response.Kind != "ConnectorSourceCatalog" || response.ExecutionChecked {
			t.Fatal("catalog became execution")
		}
		for _, operation := range response.Catalog.Operations {
			if !operation.Observed || operation.IdentityCitation < 0 || operation.IdentityCitation >= len(response.Catalog.Citations) {
				t.Fatal("uncited operation")
			}
			for _, cell := range operation.Cells {
				selection := connectors.SourceCellSelection{Source: operation.Source, Lane: cell.Lane}
				if got[selection] {
					t.Fatalf("duplicate %+v", selection)
				}
				got[selection] = true
			}
		}
	}
	if !reflect.DeepEqual(got, want) {
		for key := range want {
			if !got[key] {
				t.Errorf("missing %+v", key)
			}
		}
		for key := range got {
			if !want[key] {
				t.Errorf("unexpected %+v", key)
			}
		}
	}
}

func TestSourceVisibilityDoesNotOpenRoot162(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, ".polymetrics"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".polymetrics", "config.yaml"), []byte("invalid: ["), 0o600); err != nil {
		t.Fatal(err)
	}
	var out, diag bytes.Buffer
	if code := Run([]string{"version", "--root", root}, &out, &diag); code == 0 || !strings.Contains(diag.String(), "config.yaml") {
		t.Fatalf("actual config read control absent: %d %s", code, &diag)
	}
	out.Reset()
	diag.Reset()
	code := Run([]string{"connectors", "inspect", "asana", "--inventory", "primary", "--source-id", "asana.rest.removeCustomFieldSettingForGoal", "--lane", "direct_write", "--preflight", "--json", "--root", root}, &out, &diag)
	if code != 1 || !strings.Contains(out.String(), "source_mapping_unproven") || strings.Contains(diag.String(), "config.yaml") {
		t.Fatalf("source selection opened project root: %d %s %s", code, &out, &diag)
	}
}

func TestSourceVisibilityManualParity162(t *testing.T) {
	for _, args := range [][]string{{"help", "connectors"}, {"connectors"}, {"connectors", "inspect", "--help"}, {"connectors", "inspect", "asana", "--sources", "--help"}} {
		var out, diag bytes.Buffer
		if code := Run(append(args, "--root", t.TempDir()), &out, &diag); code != 0 || !strings.Contains(out.String(), "--source-id") || !strings.Contains(out.String(), "--preflight") {
			t.Errorf("source manual %v: %d %s", args, code, &diag)
		}
	}
}

type sourceApprovalReader162 struct{ reads int }

func (r *sourceApprovalReader162) Read([]byte) (int, error) { r.reads++; return 0, io.EOF }

func TestSourceVisibilityCLIBoundaries162(t *testing.T) {
	entries := []connectors.LazyRegistryEntry{}
	for _, e := range manifestindex.GeneratedEntries() {
		if e.Connector == "asana" || e.Connector == "vercel" || e.Connector == "github" {
			entries = append(entries, connectors.LazyRegistryEntry{Metadata: e.Metadata, SourceVisibility: e.SourceVisibility})
		}
	}
	loads, opens := 0, 0
	approval := &sourceApprovalReader162{}
	registry, err := connectors.NewLazyRegistryWithEntries(entries, func(ctx context.Context, name string) (connectors.Connector, error) {
		loads++
		b, e := engine.Load(defs.FS, name)
		if e != nil {
			return nil, e
		}
		return engine.New(b, nil), nil
	})
	if err != nil {
		t.Fatal(err)
	}
	stop := errors.New("actual-app-open-boundary")
	open := func(string) (*app.App, error) { opens++; return nil, stop }
	openers := appOpeners{open: open, reverse: open, registry: registry, approvalReader: approval, mode: appOpenerTestOverride}
	selection := []string{"connectors", "inspect", "asana", "--inventory", "primary", "--source-id", "asana.rest.removeCustomFieldSettingForGoal", "--lane", "direct_write", "--preflight"}
	for _, tc := range []struct {
		name string
		args []string
		exit int
	}{
		{"mapped", selection, 1},
		{"unknown_flag", append(append([]string{}, selection...), "--arbitrary"), 2},
		{"duplicate_id", append(append([]string{}, selection...), "--source-id", "another"), 2},
		{"approval_input_refused", append(append([]string{}, selection...), "--approval-token-stdin"), 2},
		{"missing_tuple", []string{"connectors", "inspect", "asana", "--source-id", "x"}, 2},
		{"bare_value", []string{"connectors", "inspect", "asana", "--inventory", "--source-id", "x", "--lane", "direct_write"}, 2},
		{"conflicting_forms", append(append([]string{}, selection...), "--sources"), 2},
		{"other_action", []string{"connectors", "catalog", "--sources"}, 2},
		{"false_presence", []string{"connectors", "inspect", "asana", "--sources=false"}, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var out, diag bytes.Buffer
			args := append(append([]string{}, tc.args...), "--root", t.TempDir())
			if code := run(args, &out, &diag, openers); code != tc.exit || out.Len() != 0 || loads != 0 || opens != 0 || approval.reads != 0 {
				t.Fatalf("source crossed boundary: code=%d loads=%d opens=%d reads=%d stdout=%s stderr=%s", code, loads, opens, approval.reads, &out, &diag)
			}
		})
	}
	var out, diag bytes.Buffer
	code := run([]string{"github", "issue", "list", "--credential", "fixture", "--root", t.TempDir()}, &out, &diag, openers)
	if code == 0 || !strings.Contains(diag.String(), stop.Error()) || loads != 1 || opens != 1 {
		t.Fatalf("real normal positive did not fire boundaries: %d %s loads=%d opens=%d", code, &diag, loads, opens)
	}
}

func TestSourceVisibilityDoesNotGateRealRead162(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		if r.URL.Path != "/repos/octocat/fixture/issues" || r.URL.Query().Get("per_page") != "100" {
			t.Errorf("wrong actual request: %s", r.URL.String())
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `[{"id":1,"node_id":"first","number":1,"state":"open","title":"one","user":{"login":"octocat","id":1},"updated_at":"2026-01-01T00:00:00Z"},{"id":2,"node_id":"second","number":2,"state":"open","title":"two","user":{"login":"octocat","id":1},"updated_at":"2026-01-01T00:00:00Z"},{"id":3,"node_id":"third","number":3,"state":"open","title":"three","user":{"login":"octocat","id":1},"updated_at":"2026-01-01T00:00:00Z"}]`)
	}))
	defer server.Close()
	root := t.TempDir()
	for _, args := range [][]string{{"init"}, {"credentials", "add", "public-fixture", "--connector", "github", "--config", "owner=octocat", "--config", "repo=fixture", "--config", "public_access=true", "--config", "base_url=" + server.URL}} {
		var out, diag bytes.Buffer
		if code := Run(append(args, "--root", root), &out, &diag); code != 0 {
			t.Fatalf("hermetic public-access fixture: %d %s", code, &diag)
		}
	}
	var entry connectors.LazyRegistryEntry
	for _, e := range manifestindex.GeneratedEntries() {
		if e.Connector == "github" {
			entry = connectors.LazyRegistryEntry{Metadata: e.Metadata, SourceVisibility: connectors.SourceVisibilityArtifact{SchemaVersion: 1, Connector: "github", Coverage: "in_cohort", Payload: "damaged", Bytes: 7}}
		}
	}
	loads := 0
	registry, err := connectors.NewLazyRegistryWithEntries([]connectors.LazyRegistryEntry{entry}, func(ctx context.Context, name string) (connectors.Connector, error) {
		loads++
		b, e := engine.Load(defs.FS, name)
		if e != nil {
			return nil, e
		}
		return engine.New(b, nil), nil
	})
	if err != nil {
		t.Fatal(err)
	}
	var out, diag bytes.Buffer
	if code := runWithPreflightRegistry([]string{"connectors", "inspect", "github", "--sources", "--root", root, "--json"}, &out, &diag, registry); code != 1 || !strings.Contains(out.String(), "source_visibility_invalid") || loads != 0 || requests.Load() != 0 {
		t.Fatalf("malformed source selection crossed execution: %d %s", code, &diag)
	}
	out.Reset()
	diag.Reset()
	code := runWithPreflightRegistry([]string{"github", "issue", "list", "--credential", "public-fixture", "--limit", "2", "--root", root, "--json"}, &out, &diag, registry)
	var response struct {
		Count   int `json:"count"`
		Records []struct {
			NodeID string `json:"node_id"`
		} `json:"records"`
	}
	if err = json.Unmarshal(out.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if code != 0 || loads != 1 || requests.Load() != 1 || response.Count != 2 || len(response.Records) != 2 || response.Records[0].NodeID != "first" || response.Records[1].NodeID != "second" {
		t.Fatalf("source metadata gated/corrupted actual read: %d loads=%d requests=%d count=%d stdout=%s stderr=%s", code, loads, requests.Load(), response.Count, &out, &diag)
	}
}

func TestSourceVisibilityHumanRefusal162(t *testing.T) {
	var out, diag bytes.Buffer
	code := Run([]string{"connectors", "inspect", "vercel", "--inventory", "primary", "--source-id", "vercel.rest.createWebhook", "--lane", "sync_transport", "--preflight", "--root", t.TempDir()}, &out, &diag)
	if code != 1 || out.Len() != 0 {
		t.Fatalf("text refusal emitted success: %d %s", code, &out)
	}
	for _, want := range []string{"vercel.rest.createWebhook", "POST /v1/webhooks", "transport.sync-contract.v1", "cli-webhook-event-surface-foundation-r1", "existing_foundation_demand", "/rest/operations/381/source_operation"} {
		if !strings.Contains(diag.String(), want) {
			t.Errorf("text refusal omitted %q: %s", want, &diag)
		}
	}
}

func TestSourceVisibilityUnresolvedLaneParity163(t *testing.T) {
	var entry connectors.LazyRegistryEntry
	for _, e := range manifestindex.GeneratedEntries() {
		if e.Connector == "asana" {
			entry = connectors.LazyRegistryEntry{Metadata: e.Metadata, SourceVisibility: e.SourceVisibility}
		}
	}
	loads := 0
	registry, err := connectors.NewLazyRegistryWithEntries([]connectors.LazyRegistryEntry{entry}, func(context.Context, string) (connectors.Connector, error) {
		loads++
		return nil, errors.New("unexpected execution construction")
	})
	if err != nil {
		t.Fatal(err)
	}
	selected := connectors.SourceCellSelection{Source: connectors.SourceOperationKey{Connector: "asana", Inventory: "primary", ID: "asana.rest.addCustomFieldSettingForGoal"}, Lane: "binary_download"}
	_, err = app.PreflightConnectorSource(t.Context(), registry, selected)
	var typed *connectors.SourceSelectionError
	if !errors.As(err, &typed) {
		t.Fatal(err)
	}
	if typed.Selection != selected || typed.Code != "source_mapping_unproven" || typed.View.Cell.SourceReason.Code != "facts_unresolved" || len(typed.View.Cell.LaneFacts) != 0 || typed.View.Cell.Capability.AtlasID != nil || typed.View.LaneEvidenceScope != "operation_identity_only_lane_unresolved" || len(typed.View.Citations) == 0 {
		t.Fatalf("unresolved lane lost identity/uncertainty: %v", err)
	}
	var out, diag bytes.Buffer
	code := runWithPreflightRegistry([]string{"connectors", "inspect", "asana", "--inventory=primary", "--source-id=asana.rest.addCustomFieldSettingForGoal", "--lane=binary_download", "--preflight", "--json", "--root", t.TempDir()}, &out, &diag, registry)
	var envelope struct {
		Selection connectors.SourceSelectionError `json:"source_selection"`
	}
	if e := json.Unmarshal(out.Bytes(), &envelope); e != nil {
		t.Fatal(e)
	}
	if code != 1 || loads != 0 || !reflect.DeepEqual(envelope.Selection, *typed) {
		t.Fatalf("CLI/App source parity: exit=%d loads=%d stderr=%s", code, loads, &diag)
	}
}

type sourceOutputFault163 struct {
	remaining int
	cause     error
	bytes     bytes.Buffer
	fired     bool
}

func (w *sourceOutputFault163) Write(p []byte) (int, error) {
	n := min(len(p), w.remaining)
	_, _ = w.bytes.Write(p[:n])
	w.remaining -= n
	if w.remaining == 0 {
		w.fired = true
		return n, w.cause
	}
	return n, nil
}
func TestSourceVisibilityOutputErrors163(t *testing.T) {
	cell := []string{"connectors", "inspect", "asana", "--inventory", "primary", "--source-id", "asana.rest.addCustomFieldSettingForGoal", "--lane", "binary_download"}
	for _, scope := range []struct {
		name string
		args []string
	}{
		{"cell_text", cell},
		{"catalog_text", []string{"connectors", "inspect", "asana", "--sources"}},
		{"cell_json", append(append([]string{}, cell...), "--json")},
	} {
		t.Run(scope.name, func(t *testing.T) {
			args := append(append([]string{}, scope.args...), "--root", t.TempDir())
			var complete, diag bytes.Buffer
			if code := Run(args, &complete, &diag); code != 0 || !strings.Contains(complete.String(), "asana.rest.addCustomFieldSettingForGoal") {
				t.Fatalf("actual source output control failed: %d %s", code, &diag)
			}
			expected := append([]byte{}, complete.Bytes()...)
			stop := errors.New("source-output-write-cut")
			for _, cut := range []struct {
				name  string
				limit int
				cause error
			}{
				{"before_bytes", 0, stop}, {"partial_bytes", 37, stop}, {"full_bytes_completion_error", len(expected), stop}, {"short_nil_error", 37, nil},
			} {
				t.Run(cut.name, func(t *testing.T) {
					writer := &sourceOutputFault163{remaining: cut.limit, cause: cut.cause}
					var stderr bytes.Buffer
					code := Run(args, writer, &stderr)
					if !writer.fired || !bytes.Equal(writer.bytes.Bytes(), expected[:cut.limit]) {
						t.Fatalf("named output cut not reached with exact prefix: fired=%v bytes=%d want=%d", writer.fired, writer.bytes.Len(), cut.limit)
					}
					want := "source-output-write-cut"
					if cut.cause == nil {
						want = io.ErrShortWrite.Error()
					}
					if code == 0 || !strings.Contains(stderr.String(), want) {
						t.Fatalf("source output failure reported success/lost cause: code=%d stderr=%s", code, &stderr)
					}
				})
			}
		})
	}
}
