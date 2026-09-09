package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"polymetrics.ai/internal/connectors"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestSourceVisibilityPublicGeneration162(t *testing.T) {
	repo, _, _ := sourceFoundationCombinedRetainedProfile155(t, false, "sibling159")
	hooks := filepath.Join(repo, "internal/connectors/hooks")
	if err := os.MkdirAll(hooks, 0o755); err != nil {
		t.Fatal(err)
	}
	var out, diag bytes.Buffer
	if code := runGenAt([]string{"gen"}, &out, &diag, hooks); code != 0 {
		t.Fatalf("actual generation failed: %d %s", code, &diag)
	}
	raw, err := os.ReadFile(filepath.Join(repo, "internal/connectors/manifestindex/index_gen.go"))
	if err != nil {
		t.Fatal(err)
	}
	// The current generator's valid execution entry is a positive setup control.
	if !strings.Contains(string(raw), `Connector: "acme"`) || !strings.Contains(string(raw), `Executor: "api_engine.v1"`) {
		t.Fatal("real execution index control missing")
	}
	// Inspect actual generated Go literals, independently of the runtime decoder.
	parsed, err := parser.ParseFile(token.NewFileSet(), "index_gen.go", raw, 0)
	if err != nil {
		t.Fatal(err)
	}
	var payloads []string
	ast.Inspect(parsed, func(n ast.Node) bool {
		kv, ok := n.(*ast.KeyValueExpr)
		if !ok {
			return true
		}
		key, ok := kv.Key.(*ast.Ident)
		if !ok || key.Name != "Payload" {
			return true
		}
		lit, ok := kv.Value.(*ast.BasicLit)
		if !ok {
			t.Fatal("payload not literal")
		}
		value, err := strconv.Unquote(lit.Value)
		if err != nil {
			t.Fatal(err)
		}
		if value != "" {
			payloads = append(payloads, value)
		}
		return true
	})
	if len(payloads) != 1 {
		t.Fatalf("expected exactly acme payload, got %d", len(payloads))
	}
	var visibility connectors.SourceVisibility
	if err := json.Unmarshal(sourceFixtureRaw174(t, payloads[0]), &visibility); err != nil {
		t.Fatal(err)
	}
	if visibility.Connector != "acme" || visibility.OperationCount != 2 || visibility.CellCount != 14 {
		t.Fatal("complete generated identity missing")
	}
	var ids []string
	for _, operation := range visibility.Operations {
		ids = append(ids, operation.Source.ID)
		var lanes []string
		for _, cell := range operation.Cells {
			lanes = append(lanes, string(cell.Lane))
		}
		if !reflect.DeepEqual(lanes, []string{"direct_read", "direct_write", "binary_download", "binary_upload", "etl", "reverse_etl", "sync_transport"}) {
			t.Fatalf("generated lanes differ: %v", lanes)
		}
	}
	if !reflect.DeepEqual(ids, []string{"source.a", "source.b"}) {
		t.Fatalf("generated source identities differ: %v", ids)
	}

}

func TestSourceVisibilityGenerationPreservesOutputs162(t *testing.T) {
	repo, _, _ := sourceFoundationCombinedRetainedProfile155(t, false, "sibling159")
	hooks := filepath.Join(repo, "internal/connectors/hooks")
	if err := os.MkdirAll(hooks, 0o755); err != nil {
		t.Fatal(err)
	}
	var out, diag bytes.Buffer
	if code := runGenAt([]string{"gen"}, &out, &diag, hooks); code != 0 {
		t.Fatalf("valid generation: %d %s", code, &diag)
	}
	outputs := []string{"internal/connectors/hooks/hookset/hookset_gen.go", "internal/connectors/manifestindex/index_gen.go", "internal/connectors/source_capabilities_gen.go"}
	preserved := map[string][]byte{}
	for _, name := range outputs {
		raw, err := os.ReadFile(filepath.Join(repo, name))
		if err != nil {
			t.Fatal(err)
		}
		preserved[name] = append(raw, []byte("\n// retained output sentinel\n")...)
		if err = os.WriteFile(filepath.Join(repo, name), preserved[name], 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(repo, sourceLaneCohortPath), []byte(`{"schema_version":1}`), 0o644); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	diag.Reset()
	if code := runGenAt([]string{"gen"}, &out, &diag, hooks); code == 0 {
		t.Fatal("invalid source cohort accepted")
	}
	for _, name := range outputs {
		raw, err := os.ReadFile(filepath.Join(repo, name))
		if err != nil || !bytes.Equal(raw, preserved[name]) {
			t.Errorf("invalid source changed prior generated output %s: %v", name, err)
		}
	}
}

// Existing execution-index tests now provide the mandatory retained source
// inputs explicitly. Their independent alpha execution assertions stay intact.
func sourceVisibilityGenerationRoot162(t *testing.T) string {
	t.Helper()
	repo, _, _ := sourceFoundationCombinedRetainedProfile155(t, false, "sibling159")
	return repo
}

func TestSourceVisibilityProjectionOracles162(t *testing.T) {
	repo, _, _ := sourceFoundationCombinedRetainedProfile155(t, false, "sibling159")
	raw, err := os.ReadFile(filepath.Join(repo, sourceLaneCohortPath))
	if err != nil {
		t.Fatal(err)
	}
	var cohort sourceLaneCohort
	if err = json.Unmarshal(raw, &cohort); err != nil {
		t.Fatal(err)
	}
	cohortHash := sourceBytesHash(raw)
	var out, diag bytes.Buffer
	if code := runSourceLanesContext(t.Context(), []string{"source-lanes", "--repo", repo}, &out, &diag); code != 0 {
		t.Fatalf("real report control: %d %s", code, &diag)
	}
	original := append([]byte{}, out.Bytes()...)
	root, err := os.OpenRoot(repo)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = root.Close() }()
	cache := sourceProofFileCache{ctx: t.Context(), root: root, limits: sourceProofFileLimits{UniqueBytes: 64 << 20, Files: 1}, files: map[string]*sourceProofFile{}}
	atlas, err := readSourceFoundationAtlas(&cache)
	if err != nil {
		t.Fatal(err)
	}
	fresh := func() sourceLaneManifest {
		var v sourceLaneManifest
		if err := json.Unmarshal(original, &v); err != nil {
			t.Fatal(err)
		}
		return v
	}
	positive, err := projectSourceVisibility(fresh(), cohort, cohortHash, atlas)
	if err != nil || positive["acme"].Payload == "" {
		t.Fatalf("real projection positive: %v", err)
	}
	var projected connectors.SourceVisibility
	if err := json.Unmarshal(sourceFixtureRaw174(t, positive["acme"].Payload), &projected); err != nil {
		t.Fatal(err)
	}
	matched := 0
	for _, o := range projected.Operations {
		if o.Source.ID != "source.a" {
			continue
		}
		for _, c := range o.Cells {
			if c.Lane != "direct_write" {
				continue
			}
			for _, ref := range c.Admitted {
				if ref.SchemaRole != "request" || ref.SourceSchema == nil || len(ref.FieldMappings) != 1 || ref.FieldMappings[0].TargetKind != "schema" || ref.FieldMappings[0].TargetPointer == nil || *ref.FieldMappings[0].TargetPointer != "" {
					t.Fatalf("admitted diagnostic schema/field lineage lost: %+v", ref)
				}
				citation := projected.Citations[*ref.SourceSchema]
				if !strings.HasSuffix(citation.Pointer, "/requestBody/content/application~1json/schema") || projected.Citations[ref.FieldMappings[0].SourceCitation] != citation {
					t.Fatalf("sorted schema citation changed source identity: %+v", citation)
				}
				matched++
			}
		}
	}
	if matched != 2 {
		t.Fatalf("expected both admitted operation and command diagnostic links, got %d", matched)
	}
	for _, tc := range []struct {
		name   string
		mutate func(*sourceLaneManifest)
		want   []string
	}{
		{"same_count_wrong_key", func(v *sourceLaneManifest) { v.SourceOperations[1].Source.Key.ID = "source.other" }, []string{"missing", "source.b", "unexpected", "source.other"}},
		{"hidden_delete", func(v *sourceLaneManifest) { v.SourceOperations = v.SourceOperations[:1] }, []string{"missing", "source.b"}},
		{"duplicate_source", func(v *sourceLaneManifest) { v.SourceOperations[1] = v.SourceOperations[0] }, []string{"duplicate"}},
		{"forged_fact_hash", func(v *sourceLaneManifest) {
			r := v.SourceOperations[0].Facts.Refs["source_operation"]
			r.ValueSHA256 = sourceBytesHash([]byte("forged"))
			v.SourceOperations[0].Facts.Refs["source_operation"] = r
		}, nil},
		{"sibling_fact_pointer", func(v *sourceLaneManifest) {
			v.SourceOperations[0].Facts.Refs["source_operation"] = v.SourceOperations[1].Facts.Refs["source_operation"]
		}, nil},
		{"forged_lane_citation", func(v *sourceLaneManifest) {
			v.SourceOperations[0].Lanes[1].FactRefs[0].ValueSHA256 = sourceBytesHash([]byte("forged"))
		}, nil},
		{"sibling_lane_citation", func(v *sourceLaneManifest) {
			v.SourceOperations[0].Lanes[1].FactRefs[0] = v.SourceOperations[1].Facts.Refs["method"]
		}, nil},
		{"missing_exclusion_citation", func(v *sourceLaneManifest) { v.SourceOperations[0].Lanes[0].FactRefs = nil }, nil},
		{"fake_implemented", func(v *sourceLaneManifest) {
			v.SourceOperations[0].Lanes[1].State = "implemented"
			v.SourceOperations[0].Lanes[1].Reason.Code = "behavior_proven"
		}, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			v := fresh()
			tc.mutate(&v)
			_, err := projectSourceVisibility(v, cohort, cohortHash, atlas)
			if err == nil {
				t.Fatalf("projection accepted %s", tc.name)
			}
			for _, want := range tc.want {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("projection error %q missing required identity detail %q", err, want)
				}
			}
		})
	}
}

func sourceFixtureGzip174(t *testing.T, raw []byte) string {
	t.Helper()
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	if _, err := w.Write(raw); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return b.String()
}

func sourceFixtureRaw174(t *testing.T, payload string) []byte {
	t.Helper()
	r, err := gzip.NewReader(strings.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}
	raw, err := io.ReadAll(io.LimitReader(r, connectors.SourceVisibilityConnectorLimit+1))
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Close(); err != nil {
		t.Fatal(err)
	}
	if len(raw) > connectors.SourceVisibilityConnectorLimit {
		t.Fatal("fixture over budget")
	}
	return raw
}
