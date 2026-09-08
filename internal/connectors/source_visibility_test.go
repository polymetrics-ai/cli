package connectors_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/manifestindex"
)

func sourceArtifact162(t *testing.T, name string) connectors.SourceVisibilityArtifact {
	t.Helper()
	for _, e := range manifestindex.GeneratedEntries() {
		if e.Connector == name {
			return e.SourceVisibility
		}
	}
	t.Fatal("generated fixture absent")
	return connectors.SourceVisibilityArtifact{}
}
func sourceMutated162(t *testing.T, a connectors.SourceVisibilityArtifact, mutate func(map[string]any)) connectors.SourceVisibilityArtifact {
	t.Helper()
	var v map[string]any
	if err := json.Unmarshal(sourceRaw174(t, a.Payload), &v); err != nil {
		t.Fatal(err)
	}
	mutate(v)
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(raw)
	a.Payload = sourceGzip174(t, raw)
	a.Encoding = "gzip"
	a.Bytes = len(raw)
	a.SHA256 = hex.EncodeToString(sum[:])
	return a
}
func TestSourceVisibilityStrictSelected162(t *testing.T) {
	original := sourceArtifact162(t, "vercel")
	v, err := connectors.DecodeSourceVisibility(t.Context(), original)
	if err != nil {
		t.Fatal(err)
	}
	gapIndex := -1
	for i, o := range v.Operations {
		if o.Source.ID == "vercel.rest.createWebhook" {
			gapIndex = i
		}
	}
	if gapIndex < 0 || v.Operations[gapIndex].Cells[6].Capability.Kind != "known_gap" {
		t.Fatal("actual scoped gap positive absent")
	}
	cell := func(m map[string]any) map[string]any {
		return m["operations"].([]any)[gapIndex].(map[string]any)["cells"].([]any)[6].(map[string]any)
	}
	tests := []struct {
		name   string
		mutate func(map[string]any)
	}{
		{"wrong_key_same_count", func(m map[string]any) {
			m["operations"].([]any)[gapIndex].(map[string]any)["source"].(map[string]any)["id"] = "vercel.rest.wrongWebhook"
		}},
		{"duplicate_cell", func(m map[string]any) {
			c := m["operations"].([]any)[gapIndex].(map[string]any)["cells"].([]any)
			c[6] = c[0]
		}},
		{"missing_identity_citation", func(m map[string]any) { delete(m["operations"].([]any)[0].(map[string]any), "identity_citation") }},
		{"missing_required_observed", func(m map[string]any) { delete(m["operations"].([]any)[0].(map[string]any), "observed") }},
		{"unknown_capability", func(m map[string]any) { cell(m)["capability"].(map[string]any)["atlas_id"] = "not.real.v1" }},
		{"unknown_candidate", func(m map[string]any) { cell(m)["capability"].(map[string]any)["candidate_ids"] = []any{"not.real.v1"} }},
		{"wrong_gap_owner", func(m map[string]any) {
			cell(m)["capability"].(map[string]any)["atlas_id"] = "runtime.direct-execution.v1"
		}},
		{"forged_gap", func(m map[string]any) {
			cell(m)["capability"].(map[string]any)["gap_id"] = "invented-receiver"
			cell(m)["gap_refs"] = []any{"invented-receiver"}
		}},
		{"missing_gap_citations", func(m map[string]any) { cell(m)["lane_facts"] = []any{} }},
		{"fake_implemented", func(m map[string]any) { cell(m)["source_state"] = "implemented" }},
		{"executable_addition", func(m map[string]any) { cell(m)["runnable"] = true }},
		{"unknown_document", func(m map[string]any) { m["citations"].([]any)[0].(map[string]any)["document_id"] = "unknown-document" }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := sourceMutated162(t, original, tc.mutate)
			_, err := connectors.DecodeSourceVisibility(t.Context(), a)
			var selected *connectors.SourceVisibilityDataError
			if !errors.As(err, &selected) {
				t.Fatalf("real selected decoder accepted %s: %v", tc.name, err)
			}
		})
	}
}
func TestSourceVisibilityLazyAndDefensive162(t *testing.T) {
	a := sourceArtifact162(t, "asana")
	loads := 0
	r, err := connectors.NewLazyRegistryWithEntries([]connectors.LazyRegistryEntry{{Metadata: connectors.Metadata{Name: "asana"}, SourceVisibility: a}}, func(context.Context, string) (connectors.Connector, error) { loads++; return connectors.Sample{}, nil })
	if err != nil {
		t.Fatal(err)
	}
	first, err := r.SourceVisibility(t.Context(), "asana")
	if err != nil {
		t.Fatal(err)
	}
	original := first.Operations[0].Source.ID
	first.Operations[0].Source.ID = "mutated"
	first.Citations[0].Pointer = "/mutated"
	second, err := r.SourceVisibility(t.Context(), "asana")
	if err != nil || second.Operations[0].Source.ID != original || second.Citations[0].Pointer == "/mutated" || loads != 0 {
		t.Fatalf("metadata alias/construction: %v loads=%d", err, loads)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err = r.SourceVisibility(ctx, "asana")
	if !errors.Is(err, context.Canceled) || loads != 0 {
		t.Fatalf("canceled metadata: %v loads=%d", err, loads)
	}
}

func TestSourceVisibilitySelectedHeaderOwnership162(t *testing.T) {
	artifact := sourceArtifact162(t, "vercel")
	registry, err := connectors.NewLazyRegistryWithEntries([]connectors.LazyRegistryEntry{{Metadata: connectors.Metadata{Name: "asana"}, SourceVisibility: artifact}}, func(context.Context, string) (connectors.Connector, error) {
		t.Fatal("source selection constructed execution")
		return nil, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = registry.SourceVisibility(t.Context(), "asana")
	var data *connectors.SourceVisibilityDataError
	if !errors.As(err, &data) || data.Connector != "asana" {
		t.Fatalf("selected metadata returned another connector: %v", err)
	}
}

func TestSourceVisibilityResourceAndText163(t *testing.T) {
	original := sourceArtifact162(t, "asana")
	if _, err := connectors.DecodeSourceVisibility(t.Context(), original); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name   string
		mutate func(map[string]any)
	}{
		{"unsafe_method", func(m map[string]any) { m["operations"].([]any)[0].(map[string]any)["method"] = "GET\nforged" }},
		{"unsafe_path", func(m map[string]any) { m["operations"].([]any)[0].(map[string]any)["path"] = "/ok\x1b[31m" }},
		{"unsafe_citation_section", func(m map[string]any) { m["citations"].([]any)[0].(map[string]any)["section"] = "ok\nforged" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := connectors.DecodeSourceVisibility(t.Context(), sourceMutated162(t, original, tc.mutate))
			var data *connectors.SourceVisibilityDataError
			if !errors.As(err, &data) {
				t.Fatalf("unsafe metadata accepted: %v", err)
			}
		})
	}
	for _, tc := range []struct {
		name   string
		mutate func(*connectors.SourceVisibilityArtifact)
	}{
		{"over_budget", func(a *connectors.SourceVisibilityArtifact) { a.Bytes = connectors.SourceVisibilityConnectorLimit + 1 }},
		{"length_mismatch", func(a *connectors.SourceVisibilityArtifact) { a.Bytes++ }},
		{"digest_mismatch", func(a *connectors.SourceVisibilityArtifact) { a.SHA256 = strings.Repeat("0", 64) }},
		{"key_mismatch", func(a *connectors.SourceVisibilityArtifact) { a.KeySHA256 = strings.Repeat("0", 64) }},
		{"duplicate_member", func(a *connectors.SourceVisibilityArtifact) {
			raw := []byte(strings.Replace(string(sourceRaw174(t, a.Payload)), `"schema_version":1`, `"schema_version":1,"schema_version":1`, 1))
			a.Payload = sourceGzip174(t, raw)
			a.Bytes = len(raw)
			h := sha256.Sum256(raw)
			a.SHA256 = hex.EncodeToString(h[:])
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a := original
			tc.mutate(&a)
			_, err := connectors.DecodeSourceVisibility(t.Context(), a)
			var data *connectors.SourceVisibilityDataError
			if !errors.As(err, &data) {
				t.Fatalf("malformed payload accepted: %v", err)
			}
		})
	}
}

func TestSourceVisibilityCoverageAndUnknowns163(t *testing.T) {
	github := sourceArtifact162(t, "github")
	if github.Coverage != "not_in_cohort" {
		t.Fatal("independent non-cohort control changed")
	}
	loads := 0
	resolver := func(context.Context, string) (connectors.Connector, error) { loads++; return connectors.Sample{}, nil }
	r, err := connectors.NewLazyRegistryWithEntries([]connectors.LazyRegistryEntry{{Metadata: connectors.Metadata{Name: "github"}, SourceVisibility: github}, {Metadata: connectors.Metadata{Name: "asana"}, SourceVisibility: connectors.SourceVisibilityArtifact{SchemaVersion: 1, Connector: "asana", Coverage: "in_cohort"}}}, resolver)
	if err != nil {
		t.Fatal(err)
	}
	v, err := r.SourceVisibility(t.Context(), "github")
	if err != nil || v.Coverage != "not_in_cohort" || len(v.Operations) != 0 || loads != 0 {
		t.Fatalf("unrelated damaged payload blocked selected non-cohort metadata: %v", err)
	}
	_, err = r.PreflightSource(t.Context(), connectors.SourceCellSelection{Source: connectors.SourceOperationKey{Connector: "github", Inventory: "primary", ID: "unknown"}, Lane: "direct_read"})
	var input *connectors.SourceSelectionInputError
	if !errors.As(err, &input) || input.Code != "source_cell_not_found" {
		t.Fatalf("non-cohort invented source cell: %v", err)
	}
	_, err = r.SourceVisibility(t.Context(), "absent")
	if !errors.As(err, &input) {
		t.Fatalf("unknown connector: %v", err)
	}
	if _, err = r.SourceVisibility(nil, "github"); err == nil { //nolint:staticcheck // Deliberately verify the nil-context refusal, not a valid caller.
		t.Fatal("nil context accepted")
	}
	var nilRegistry *connectors.Registry
	if _, err = nilRegistry.SourceVisibility(t.Context(), "github"); err == nil {
		t.Fatal("nil registry accepted")
	}
	compatibility, err := connectors.NewLazyRegistry([]connectors.Metadata{{Name: "asana"}}, resolver)
	if err != nil {
		t.Fatal(err)
	}
	_, err = compatibility.SourceVisibility(t.Context(), "asana")
	var data *connectors.SourceVisibilityDataError
	if !errors.As(err, &data) || loads != 0 {
		t.Fatalf("unsupplied compatibility metadata claimed cohort coverage: %v loads=%d", err, loads)
	}
}
