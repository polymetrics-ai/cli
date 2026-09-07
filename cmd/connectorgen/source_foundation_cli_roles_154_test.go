package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// Only test metadata is extended. These records must remain unreviewed; this
// helper does not manufacture original execution evidence or review authority.
func sourceFoundationAddFixtureRole154(t *testing.T, repo, name string) {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(repo, sourceFoundationProofPath))
	if err != nil {
		t.Fatal(err)
	}
	var doc sourceFoundationProofDocument
	if err = json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	input, err := os.ReadFile(filepath.Join(repo, name))
	if err != nil {
		t.Fatal(err)
	}
	r := &doc.Records[0]
	pin := sourceFoundationProofInput{Path: name, SHA256: sourceBytesHash(input), Bytes: int64(len(input)), Role: "fixture"}
	r.Inputs = append(r.Inputs, pin)
	sort.Slice(r.Inputs, func(i, j int) bool { return r.Inputs[i].Path < r.Inputs[j].Path })
	raw, err = os.ReadFile(filepath.Join(repo, r.Capture.Path))
	if err != nil {
		t.Fatal(err)
	}
	var capture map[string]any
	if err = json.Unmarshal(raw, &capture); err != nil {
		t.Fatal(err)
	}
	capture["inputs"].(map[string]any)[name] = map[string]any{"sha256": pin.SHA256, "bytes": pin.Bytes, "snapshot": "test-only-unreviewed-overlap"}
	raw, err = json.Marshal(capture)
	if err != nil {
		t.Fatal(err)
	}
	proofWrite(t, repo, r.Capture.Path, raw)
	r.Capture.SHA256, r.Capture.Bytes = sourceBytesHash(raw), int64(len(raw))
	writeSourceFoundationObservationFixture(t, repo, doc)
}

func TestSourceFoundationRequiredRoleOverlap154(t *testing.T) {
	for _, name := range []string{sourceLaneCohortPath, sourceLaneAnnotationsPath, sourceDemandAtlasPath} {
		t.Run(name, func(t *testing.T) {
			repo, baseline := sourceFoundationCommandFixture(t)
			sourceFoundationAddFixtureRole154(t, repo, name)
			var out, diag bytes.Buffer
			args := []string{"source-demands", "--repo", repo}
			if code := runSourceDemandsPolicy(t.Context(), args, &out, &diag, baseline); code != 0 {
				t.Fatal(diag.String())
			}
			if err := os.Remove(filepath.Join(repo, name)); err != nil {
				t.Fatal(err)
			}
			out.Reset()
			diag.Reset()
			if code := runSourceDemandsPolicy(t.Context(), args, &out, &diag, baseline); code == 0 || out.Len() != 0 {
				t.Fatal("required input absence softened into optional evidence")
			}
			if _, err := os.Stat(filepath.Join(repo, name)); !os.IsNotExist(err) {
				t.Fatal("missing required input was created")
			}
		})
	}
}

func TestSourceFoundationSavedCheckFixtureRole154(t *testing.T) {
	repo, baseline := sourceFoundationCommandFixture(t)
	name := "candidate.json"
	raw := []byte("{\"deliberately_invalid_saved_register\":true}\n")
	proofWrite(t, repo, name, raw)
	sourceFoundationAddFixtureRole154(t, repo, name)
	var out, diag bytes.Buffer
	// The ordinary producer must first admit the altered, explicitly unreviewed
	// fixture metadata; otherwise a preload error could falsely pass this test.
	args := []string{"source-demands", "--repo", repo}
	if code := runSourceDemandsPolicy(t.Context(), args, &out, &diag, baseline); code != 0 {
		t.Fatal(diag.String())
	}
	before, err := os.Stat(filepath.Join(repo, name))
	if err != nil {
		t.Fatal(err)
	}
	events := []sourceProofReadEvent{}
	out.Reset()
	diag.Reset()
	code := runSourceDemandsObserved(t.Context(), append(args, "--check", name), &out, &diag, baseline, func(e sourceProofReadEvent) {
		if e.Path == name {
			events = append(events, e)
		}
	})
	if code == 0 || out.Len() != 0 || !strings.Contains(diag.String(), "saved register invalid or changed") {
		t.Fatalf("actual saved-check consumer not reached: %s", diag.String())
	}
	if len(events) != 1 || !events[0].Success || events[0].Phase != "initial" || events[0].Bytes != int64(len(raw)) {
		t.Fatalf("saved byte role reread or failed: %+v", events)
	}
	after, err := os.Stat(filepath.Join(repo, name))
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(repo, name))
	if err != nil || !bytes.Equal(got, raw) || !os.SameFile(before, after) || before.ModTime() != after.ModTime() {
		t.Fatal("check changed original candidate")
	}
}
