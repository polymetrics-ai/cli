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

type sourceLaneObservedWriter struct {
	limit              int
	err                error
	calls              int
	attempted, written []byte
}

func (w *sourceLaneObservedWriter) Write(raw []byte) (int, error) {
	w.calls++
	w.attempted = append(w.attempted, raw...)
	n := len(raw)
	if w.limit >= 0 {
		n = min(n, w.limit)
	}
	w.written = append(w.written, raw[:n]...)
	return n, w.err
}

func TestSourceLaneCLIOutputBoundaryWitness(t *testing.T) {
	root := sourceLaneCLIRepository(t)
	proofWrite(t, root, "candidate.json", []byte("retained owned candidate\n"))
	var expected, diagnostics bytes.Buffer
	if code := runContext(context.Background(), []string{"source-lanes", "--repo", root}, &expected, &diagnostics); code != 0 {
		t.Fatalf("positive generation failed%d: %s", code, &diagnostics)
	}
	var document sourceLaneManifest
	if err := json.Unmarshal(expected.Bytes(), &document); err != nil {
		t.Fatal(err)
	}
	if len(document.SourceOperations) != 2 || document.SourceOperations[0].Source.Key.ID != "source.a" || document.SourceOperations[1].Source.Key.ID != "source.b" || document.SourceTotals.Cells != 14 {
		t.Fatal("positive writer did not receive independently expected source rows/cells")
	}
	before := sourceBindingFixtureSnapshot(t, root)
	candidate := filepath.Join(root, "candidate.json")
	identity, err := os.Stat(candidate)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name     string
		limit    int
		err      error
		wantCode int
	}{
		{"complete success", -1, nil, 0},
		{"short successful return", expected.Len() / 2, nil, 1},
		{"partial write error", expected.Len() / 2, io.ErrClosedPipe, 1},
		{"complete write with completion error", -1, io.ErrClosedPipe, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			writer := &sourceLaneObservedWriter{limit: tc.limit, err: tc.err}
			diagnostics.Reset()
			code := runContext(context.Background(), []string{"source-lanes", "--repo", root}, writer, &diagnostics)
			if code != tc.wantCode || writer.calls != 1 || !bytes.Equal(writer.attempted, expected.Bytes()) {
				t.Fatalf("wrong reached output boundary: code%d calls%d attempted%d expected%d", code, writer.calls, len(writer.attempted), expected.Len())
			}
			n := expected.Len()
			if tc.limit >= 0 {
				n = min(n, tc.limit)
			}
			if !bytes.Equal(writer.written, expected.Bytes()[:n]) {
				t.Fatal("actual written prefix differs from retained expected document")
			}
			if tc.wantCode != 0 && !strings.Contains(diagnostics.String(), "output_write_failed") {
				t.Fatalf("missing output failure diagnostic: %s", &diagnostics)
			}
			if after := sourceBindingFixtureSnapshot(t, root); !reflect.DeepEqual(before, after) {
				t.Fatal("output outcome changed input/candidate tree")
			}
			after, err := os.Stat(candidate)
			if err != nil || !os.SameFile(identity, after) {
				t.Fatalf("output outcome replaced owned candidate: %v", err)
			}
		})
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var canceled bytes.Buffer
	diagnostics.Reset()
	if code := runContext(ctx, []string{"source-lanes", "--repo", root}, &canceled, &diagnostics); code != 1 || canceled.Len() != 0 || !strings.Contains(diagnostics.String(), "canceled") {
		t.Fatalf("cancellation emitted success/partial manifest: code%d bytes%d %s", code, canceled.Len(), &diagnostics)
	}
	if after := sourceBindingFixtureSnapshot(t, root); !reflect.DeepEqual(before, after) {
		t.Fatal("canceled generation changed retained tree")
	}
	after, err := os.Stat(candidate)
	if err != nil || !os.SameFile(identity, after) {
		t.Fatalf("canceled generation replaced owned candidate: %v", err)
	}
}
