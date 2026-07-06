package connectors

import (
	"encoding/json"
	"testing"
)

// fakeDefinitionConnector is the minimal Tier-3-style fake used to exercise
// DefinitionProvider without depending on the engine package (avoids an
// import cycle: engine imports connectors, so connectors' own tests cannot
// import engine).
type fakeDefinitionConnector struct {
	def Definition
}

func (f fakeDefinitionConnector) Name() string { return f.def.Name }
func (f fakeDefinitionConnector) Metadata() Metadata {
	return Metadata{Name: f.def.Name, DisplayName: f.def.DisplayName}
}
func (f fakeDefinitionConnector) Definition() Definition { return f.def }

// TestDefinitionProviderRoundTrip asserts DefinitionProvider is implementable
// and Definition() returns the exact struct handed in, unmodified.
func TestDefinitionProviderRoundTrip(t *testing.T) {
	want := Definition{
		Name:            "acme",
		DisplayName:     "Acme",
		Description:     "Acme connector.",
		IntegrationType: "api",
		DocsURL:         "https://example.com/docs",
		ReleaseStage:    "beta",
		Capabilities:    Capabilities{Check: true, Read: true},
		Spec:            json.RawMessage(`{"type":"object"}`),
		Streams: []StreamSummary{
			{Name: "widgets", PrimaryKey: []string{"id"}, SyncModes: []string{"full_refresh_append"}},
		},
		WriteActions: []WriteActionInfo{
			{Name: "update_widget", Kind: "update", Method: "PATCH", Path: "/widgets/{id}", Risk: "medium"},
		},
		Risk: RiskSpec{Read: "low", Write: "medium"},
	}

	var provider DefinitionProvider = fakeDefinitionConnector{def: want}
	got := provider.Definition()

	if got.Name != want.Name || got.DisplayName != want.DisplayName {
		t.Fatalf("Definition() name/display_name = %q/%q, want %q/%q", got.Name, got.DisplayName, want.Name, want.DisplayName)
	}
	if len(got.Streams) != 1 || got.Streams[0].Name != "widgets" {
		t.Fatalf("Definition().Streams = %+v, want one stream named widgets", got.Streams)
	}
	if len(got.WriteActions) != 1 || got.WriteActions[0].Name != "update_widget" {
		t.Fatalf("Definition().WriteActions = %+v, want one action named update_widget", got.WriteActions)
	}
}

// TestDefinitionJSONShape locks the wire shape: field names/omitempty
// behavior per API-CONTRACT.md §1, so downstream JSON consumers (CLI --json,
// certify) don't silently drift.
func TestDefinitionJSONShape(t *testing.T) {
	def := Definition{
		Name:            "acme",
		DisplayName:     "Acme",
		IntegrationType: "api",
		ReleaseStage:    "ga",
		Capabilities:    Capabilities{Check: true},
		Spec:            json.RawMessage(`{}`),
		Streams: []StreamSummary{
			{Name: "widgets", SyncModes: []string{"full_refresh_append", "full_refresh_overwrite"}},
		},
		Risk: RiskSpec{Read: "low"},
	}

	raw, err := json.Marshal(def)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	for _, field := range []string{"name", "display_name", "integration_type", "release_stage", "capabilities", "spec", "streams", "risk"} {
		if _, ok := decoded[field]; !ok {
			t.Fatalf("Definition JSON missing field %q; got keys %v", field, keysOf(decoded))
		}
	}
	// write_actions/icon/description/docs_url are omitempty and unset here.
	for _, field := range []string{"write_actions", "icon", "description", "docs_url"} {
		if _, ok := decoded[field]; ok {
			t.Fatalf("Definition JSON unexpectedly present unset omitempty field %q", field)
		}
	}

	streams, ok := decoded["streams"].([]any)
	if !ok || len(streams) != 1 {
		t.Fatalf("Definition JSON streams = %v, want one entry", decoded["streams"])
	}
	streamObj, ok := streams[0].(map[string]any)
	if !ok {
		t.Fatalf("Definition JSON streams[0] not an object: %v", streams[0])
	}
	if _, ok := streamObj["sync_modes"]; !ok {
		t.Fatalf("StreamSummary JSON missing sync_modes; got %v", streamObj)
	}
	if _, ok := streamObj["primary_key"]; ok {
		t.Fatalf("StreamSummary JSON unexpectedly present unset omitempty primary_key: %v", streamObj)
	}
}

func keysOf(m map[string]any) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
