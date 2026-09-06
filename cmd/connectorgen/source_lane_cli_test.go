package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func sourceLaneCLIRepository(t *testing.T) string {
	t.Helper()
	root, cohort := sourceInventoryFixture(t, []string{"source.a", "source.b"}, 2)
	raw, err := json.Marshal(cohort)
	if err != nil {
		t.Fatal(err)
	}
	proofWrite(t, root, "data/connector-canon/batch1-source-lane-cohort.json", raw)
	proofWrite(t, root, "data/connector-canon/batch1-source-lane-annotations.json", []byte(`{"schema_version":1,"annotations":[]}`))
	proofDocument(t, root, []sourceLaneProofRecord{})
	return root
}

func TestSourceLaneCLIReadOnlyRoundTrip(t *testing.T) {
	root := sourceLaneCLIRepository(t)
	before := sourceBindingFixtureSnapshot(t, root)
	var out, diagnostics bytes.Buffer
	if code := runContext(context.Background(), []string{"source-lanes", "--repo", root}, &out, &diagnostics); code != 0 {
		t.Fatalf("generate code%d: %s", code, &diagnostics)
	}
	first := append([]byte(nil), out.Bytes()...)
	if len(first) == 0 || first[len(first)-1] != '\n' {
		t.Fatal("manifest missing final newline")
	}
	var result sourceLaneManifest
	if err := json.Unmarshal(first, &result); err != nil {
		t.Fatal(err)
	}
	if result.SourceTotals.Primary != 2 || result.SourceTotals.Cells != 14 || result.SourceOperations[0].Source.Key.ID != "source.a" || result.SourceOperations[1].Source.Key.ID != "source.b" {
		t.Fatal("CLI did not return exact retained records")
	}
	out.Reset()
	diagnostics.Reset()
	if code := run([]string{"source-lanes", "--repo", root}, &out, &diagnostics); code != 0 || !bytes.Equal(first, out.Bytes()) {
		t.Fatalf("same inputs produced different full bytes, code%d", code)
	}
	if after := sourceBindingFixtureSnapshot(t, root); !reflect.DeepEqual(before, after) {
		t.Fatal("generation wrote files")
	}
	proofWrite(t, root, "candidate.json", first)
	before = sourceBindingFixtureSnapshot(t, root)
	out.Reset()
	diagnostics.Reset()
	if code := run([]string{"source-lanes", "--check", "--manifest", "candidate.json", "--repo", root}, &out, &diagnostics); code != 0 {
		t.Fatalf("exact candidate check failed%d: %s", code, &diagnostics)
	}
	if out.Len() != 0 {
		t.Fatal("check unexpectedly emitted manifest bytes")
	}
	if after := sourceBindingFixtureSnapshot(t, root); !reflect.DeepEqual(before, after) {
		t.Fatal("check wrote files")
	}
	result.SourceOperations[0].Source.Key.ID = "same-count-other"
	wrong, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	proofWrite(t, root, "candidate.json", wrong)
	before = sourceBindingFixtureSnapshot(t, root)
	diagnostics.Reset()
	if code := run([]string{"source-lanes", "--check", "--manifest", "candidate.json", "--repo", root}, io.Discard, &diagnostics); code != 1 || !strings.Contains(diagnostics.String(), "manifest_source_unexpected") {
		t.Fatalf("same-count substitution passed or lost precise error: code%d %s", code, &diagnostics)
	}
	if after := sourceBindingFixtureSnapshot(t, root); !reflect.DeepEqual(before, after) {
		t.Fatal("failed check changed candidate")
	}
}

func TestSourceLaneCLIArguments(t *testing.T) {
	for _, args := range [][]string{{"--unknown"}, {"--repo"}, {"--repo", "a", "--repo", "b"}, {"--check", "--check"}, {"--manifest", "candidate.json"}, {"--check", "--manifest", "../escape"}, {"--check", "--manifest", "/absolute"}, {"extra"}, {"--help", "--unknown"}} {
		var out, err bytes.Buffer
		if code := run(append([]string{"source-lanes"}, args...), &out, &err); code != 2 || out.Len() != 0 {
			t.Fatalf("invalid args%v: code%d stdout%q stderr%q", args, code, &out, &err)
		}
	}
	var out, err bytes.Buffer
	if code := run([]string{"source-lanes", "--repo", "/nonexistent/source-lane-help", "--help"}, &out, &err); code != 0 || !strings.Contains(out.String(), "authoring-only") || !strings.Contains(out.String(), "--check") {
		t.Fatalf("help performed input work or lacks contract: code%d %s %s", code, &out, &err)
	}
}

type sourceLaneShortWriter struct{}

func (sourceLaneShortWriter) Write(p []byte) (int, error) { return len(p) / 2, nil }

func TestSourceLaneCLIInputAndOutputFailure(t *testing.T) {
	root := sourceLaneCLIRepository(t)
	var out, diagnostics bytes.Buffer
	if code := run([]string{"source-lanes", "--repo", root}, sourceLaneShortWriter{}, &diagnostics); code != 1 {
		t.Fatalf("short stdout accepted: code%d", code)
	}
	if err := os.Remove(filepath.Join(root, "source.json")); err != nil {
		t.Fatal(err)
	}
	diagnostics.Reset()
	if code := run([]string{"source-lanes", "--repo", root}, &out, &diagnostics); code != 1 {
		t.Fatalf("missing source reported success%d", code)
	}
	var result sourceLaneManifest
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Validation.Status != "invalid" || result.SourceTotals.Primary != 2 || result.SourceTotals.Cells != 14 {
		t.Fatal("source failure suppressed complete diagnostic universe")
	}
	out.Reset()
	diagnostics.Reset()
	proofWrite(t, root, "data/connector-canon/batch1-source-lane-cohort.json", []byte(`{"schema_version":1,"inventories":[]}`))
	if code := run([]string{"source-lanes", "--repo", root}, &out, &diagnostics); code != 1 || out.Len() != 0 {
		t.Fatalf("invalid anchor fabricated manifest: code%d output%q", code, &out)
	}
}
