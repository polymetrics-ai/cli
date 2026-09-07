package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func sourceFoundationCommandFixture(t *testing.T) (string, sourceArtifactPin) {
	t.Helper()
	repo, inputs := sourceFoundationRegisterFixture(t)
	_, cohort := sourceInventoryFixture(t, []string{"source.b", "source.a"}, 2)
	raw, err := json.Marshal(cohort)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, sourceLaneCohortPath), raw, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, sourceLaneAnnotationsPath), []byte(`{"schema_version":1,"annotations":[]}`), 0600); err != nil {
		t.Fatal(err)
	}
	var source, diagnostic bytes.Buffer
	if code := runSourceLanesContext(t.Context(), []string{"source-lanes", "--repo", repo}, &source, &diagnostic); code != 0 {
		t.Fatalf("actual source-lanes fixture generation: code%d %s", code, diagnostic.String())
	}
	if err := os.WriteFile(filepath.Join(repo, sourceLaneManifestPath), source.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	return repo, inputs.baselinePin
}

func sourceFoundationBoundCommandFixture(t *testing.T) (string, sourceArtifactPin, string) {
	t.Helper()
	repo, _, _, _ := sourceFoundationRequirementsFixture(t)
	key, facts, _ := sourceBindingFixture099F(t, "body", nil)
	for _, files := range []map[string][]byte{facts.bindings.Authoring, facts.bindings.Artifacts} {
		for name, raw := range files {
			path := filepath.Join(repo, name)
			if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, raw, 0600); err != nil {
				t.Fatal(err)
			}
		}
	}
	_, cohort := sourceInventoryFixture(t, []string{"source.b", "source.a"}, 2)
	cohort.Inventories[0].Connector = key.Connector
	sourceRaw, err := os.ReadFile(filepath.Join(repo, "source.json"))
	if err != nil {
		t.Fatal(err)
	}
	var retained map[string]any
	if err := json.Unmarshal(sourceRaw, &retained); err != nil {
		t.Fatal(err)
	}
	retained["connector"] = key.Connector
	sourceRaw, err = json.Marshal(retained)
	if err != nil {
		t.Fatal(err)
	}
	proofWrite(t, repo, "source.json", sourceRaw)
	cohort.Inventories[0].SHA256 = sourceBytesHash(sourceRaw)
	universe, err := buildSourceFoundationUniverse(t.Context(), repo, cohort, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, exists := universe.manifest.SourceOperations[0].Facts.bindings.Bundles[key.Connector]; !exists {
		t.Fatal("actual source reader did not load the valid canonical execution bundle")
	}
	assessment := sourceFoundationAssessmentDocumentFixture(t, universe)
	assessment.Assessments[0].Key = universe.manifest.SourceOperations[0].Source.Key
	writeSourceFoundationAssessmentFixture(t, repo, assessment)
	_, baseline := sourceFoundationObligationsFixture(t, repo, universe)
	raw, err := json.Marshal(cohort)
	if err != nil {
		t.Fatal(err)
	}
	proofWrite(t, repo, sourceLaneCohortPath, raw)
	proofWrite(t, repo, sourceLaneAnnotationsPath, []byte(`{"schema_version":1,"annotations":[]}`))
	var manifest, diagnostic bytes.Buffer
	if code := runSourceLanesContext(t.Context(), []string{"source-lanes", "--repo", repo}, &manifest, &diagnostic); code != 0 {
		t.Fatalf("actual bound source generation: %s", diagnostic.String())
	}
	proofWrite(t, repo, sourceLaneManifestPath, manifest.Bytes())
	return repo, baseline, "internal/connectors/defs/" + key.Connector + "/spec.json"
}

func TestSourceDemandsCommandBindingInputCustody(t *testing.T) {
	repo, baseline, name := sourceFoundationBoundCommandFixture(t)
	path := filepath.Join(repo, name)
	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	before, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	var output, diagnostic bytes.Buffer
	args := []string{"source-demands", "--repo", repo}
	if code := runSourceDemandsPolicy(t.Context(), args, &output, &diagnostic, baseline); code != 0 {
		t.Fatalf("actual source/canonical/bundle control: %s", diagnostic.String())
	}
	var register sourceFoundationRegister
	if err := json.Unmarshal(output.Bytes(), &register); err != nil {
		t.Fatal(err)
	}
	t.Run("consumed declaration pin", func(t *testing.T) {
		want := sourceArtifactPin{Path: name, SHA256: sourceBytesHash(original), Bytes: int64(len(original))}
		found := false
		for _, pin := range register.Inputs {
			found = found || pin == want
		}
		if !found {
			t.Error("actual command omitted the consumed execution declaration input pin")
		}
	})
	t.Run("replacement after source observation", func(t *testing.T) {
		fired := false
		observer := func(event sourceProofReadEvent) {
			if fired || event.Path != sourceFoundationAssessmentsPath || event.Phase != "initial" || !event.Success {
				return
			}
			fired = true
			proofWrite(t, repo, name+".replacement", original)
			if err := os.Rename(path+".replacement", path); err != nil {
				t.Fatal(err)
			}
		}
		output.Reset()
		code := runSourceDemandsObserved(t.Context(), args, &output, &diagnostic, baseline, observer)
		after, err := os.Stat(path)
		if err != nil || !fired || os.SameFile(before, after) {
			t.Fatal("actual successful source/assessment read did not establish the replacement inode")
		}
		current, err := os.ReadFile(path)
		if err != nil || !bytes.Equal(current, original) {
			t.Fatal("later declaration replacement was rewritten")
		}
		if code == 0 || output.Len() != 0 {
			t.Error("actual command emitted stale reconciliation after declaration replacement")
		}
	})
}

func TestSourceDemandsCommandOutputAndCancellation(t *testing.T) {
	repo, baseline := sourceFoundationCommandFixture(t)
	args := []string{"source-demands", "--repo", repo}
	var expected, diagnostic bytes.Buffer
	if code := runSourceDemandsPolicy(t.Context(), args, &expected, &diagnostic, baseline); code != 0 {
		t.Fatalf("complete real command control: %s", diagnostic.String())
	}
	before := sourceBindingFixtureSnapshot(t, repo)
	for _, scenario := range []struct {
		name  string
		limit int
		err   error
		code  int
	}{{"complete", -1, nil, 0}, {"short", expected.Len() / 2, nil, 1}, {"partial error", expected.Len() / 2, io.ErrClosedPipe, 1}, {"complete error", -1, io.ErrClosedPipe, 1}} {
		t.Run(scenario.name, func(t *testing.T) {
			writer := &sourceLaneObservedWriter{limit: scenario.limit, err: scenario.err}
			code := runSourceDemandsPolicy(t.Context(), args, writer, &diagnostic, baseline)
			if code != scenario.code || writer.calls != 1 || !bytes.Equal(writer.attempted, expected.Bytes()) {
				t.Fatal("actual complete-register output frontier lost error or attempted byte identity")
			}
			count := expected.Len()
			if scenario.limit >= 0 {
				count = scenario.limit
			}
			if !bytes.Equal(writer.written, expected.Bytes()[:count]) || !reflect.DeepEqual(before, sourceBindingFixtureSnapshot(t, repo)) {
				t.Fatal("output outcome changed the expected prefix or retained input tree")
			}
		})
	}
	for _, phase := range []string{"initial", "final"} {
		t.Run("cancel after actual "+phase+" read", func(t *testing.T) {
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			fired := false
			observer := func(event sourceProofReadEvent) {
				if !fired && event.Phase == phase && event.Success {
					fired = true
					cancel()
				}
			}
			var output bytes.Buffer
			if code := runSourceDemandsObserved(ctx, args, &output, &diagnostic, baseline, observer); code != 1 || !fired || output.Len() != 0 {
				t.Fatal("actual read cancellation did not refuse output")
			}
			if !reflect.DeepEqual(before, sourceBindingFixtureSnapshot(t, repo)) {
				t.Fatal("cancellation changed the retained input tree")
			}
		})
	}
}

func TestSourceDemandsCommandCandidateCheckPreservesFile(t *testing.T) {
	repo, baseline := sourceFoundationCommandFixture(t)
	args := []string{"source-demands", "--repo", repo}
	var expected, diagnostic bytes.Buffer
	if code := runSourceDemandsPolicy(t.Context(), args, &expected, &diagnostic, baseline); code != 0 {
		t.Fatal("actual complete command control failed")
	}
	for _, scenario := range []string{"exact", "coherent forged coverage", "unknown field", "changed full bytes"} {
		t.Run(scenario, func(t *testing.T) {
			raw := append([]byte{}, expected.Bytes()...)
			var candidate map[string]any
			if err := json.Unmarshal(raw, &candidate); err != nil {
				t.Fatal(err)
			}
			switch scenario {
			case "coherent forged coverage":
				candidate["known_obligations"] = []any{}
			case "unknown field":
				candidate["implemented"] = true
			case "changed full bytes":
				raw = append(raw, ' ')
			}
			if scenario == "coherent forged coverage" || scenario == "unknown field" {
				var err error
				raw, err = json.Marshal(candidate)
				if err != nil {
					t.Fatal(err)
				}
			}
			path := filepath.Join(repo, "candidate.json")
			if err := os.WriteFile(path, raw, 0600); err != nil {
				t.Fatal(err)
			}
			before, err := os.Stat(path)
			if err != nil {
				t.Fatal(err)
			}
			var output bytes.Buffer
			code := runSourceDemandsPolicy(t.Context(), append(args, "--check", "candidate.json"), &output, &diagnostic, baseline)
			wantCode := 1
			if scenario == "exact" {
				wantCode = 0
			}
			if code != wantCode || output.Len() != 0 {
				t.Fatal("candidate check accepted forged output or wrote a report")
			}
			after, err := os.Stat(path)
			if err != nil || !os.SameFile(before, after) || before.Mode() != after.Mode() || before.ModTime() != after.ModTime() {
				t.Fatal("candidate check changed original file identity/metadata")
			}
			saved, err := os.ReadFile(path)
			if err != nil || !bytes.Equal(saved, raw) {
				t.Fatal("candidate check changed original full bytes")
			}
		})
	}
}

func TestSourceDemandsCommandObservation(t *testing.T) {
	repo, baseline := sourceFoundationCommandFixture(t)
	args := []string{"source-demands", "--repo", repo}
	var output, diagnostic bytes.Buffer
	if code := runSourceDemandsPolicy(t.Context(), args, &output, &diagnostic, baseline); code != 0 {
		t.Fatalf("actual command generation: code%d %s", code, diagnostic.String())
	}
	var register sourceFoundationRegister
	if err := json.Unmarshal(output.Bytes(), &register); err != nil {
		t.Fatal(err)
	}
	if register.Coverage.UniverseCount != 14 || len(register.Coverage.Assessed) != 1 || len(register.Coverage.Unassessed) != 13 ||
		len(register.Known) != 1 || len(register.Requirements) != 1 || len(register.Facets) != 4 || len(register.Examples) != 35 || len(register.Adopters) != 0 ||
		register.Requirements[0].Status != "unresolved" || register.Requirements[0].Proofs[0].Status != "current" {
		t.Fatal("actual command output lost bounded source/requirement observations")
	}
	sourceBytes, err := os.ReadFile(filepath.Join(repo, sourceLaneManifestPath))
	if err != nil {
		t.Fatal(err)
	}
	wantPin := sourceArtifactPin{Path: sourceLaneManifestPath, SHA256: sourceBytesHash(sourceBytes), Bytes: int64(len(sourceBytes))}
	found := false
	for _, pin := range register.Inputs {
		if pin == wantPin {
			found = true
		}
	}
	if !found || len(register.Proofs) != 2 || register.Checks.SelectedReuseRequirements != 0 || register.Checks.UnresolvedRequirements != 1 {
		t.Fatal("actual source file pin or separate proof/requirement accounting missing")
	}
	path := filepath.Join(repo, sourceFoundationRegisterPath)
	if err := os.WriteFile(path, output.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	before, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	var checkOutput bytes.Buffer
	if code := runSourceDemandsPolicy(t.Context(), append(args, "--check"), &checkOutput, &diagnostic, baseline); code != 0 || checkOutput.Len() != 0 {
		t.Fatalf("actual default check: code%d %s", code, diagnostic.String())
	}
	after, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	saved, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(saved, output.Bytes()) || !os.SameFile(before, after) || before.Mode() != after.Mode() || before.ModTime() != after.ModTime() {
		t.Fatal("check changed the retained full register bytes or file identity")
	}
}

func TestSourceDemandsCommandHelpAndClosedArguments(t *testing.T) {
	for _, args := range [][]string{
		{"source-demands", "--help"},
		{"source-demands", "--proofs", "other.json"},
		{"source-demands", "--cohort", "../outside.json"},
		{"source-demands", "--manifest"},
		{"source-demands", "--check", "--check"},
	} {
		t.Run(args[1]+"/"+args[len(args)-1], func(t *testing.T) {
			var output, diagnostic bytes.Buffer
			code := runContext(t.Context(), args, &output, &diagnostic)
			if args[1] == "--help" {
				if code != 0 || !bytes.Contains(output.Bytes(), []byte("relative-source-manifest")) || !bytes.Contains(output.Bytes(), []byte("not executable capability")) {
					t.Fatal("real main dispatch lost authoring command help")
				}
			} else if code != 2 || output.Len() != 0 {
				t.Fatal("invalid authoring arguments reached repository work or produced output")
			}
		})
	}
}

func TestSourceDemandsCommandFinalInputCustody(t *testing.T) {
	for _, scenario := range []string{"manifest inode replacement", "manifest same inode changed bytes", "cohort inode replacement"} {
		t.Run(scenario, func(t *testing.T) {
			repo, baseline := sourceFoundationCommandFixture(t)
			target := sourceLaneManifestPath
			if scenario == "cohort inode replacement" {
				target = sourceLaneCohortPath
			}
			path := filepath.Join(repo, target)
			original, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			before, err := os.Stat(path)
			if err != nil {
				t.Fatal(err)
			}
			changed := append(append([]byte{}, original...), ' ')
			fired := false
			observer := func(event sourceProofReadEvent) {
				if fired || event.Path != sourceFoundationAssessmentsPath || event.Phase != "initial" {
					return
				}
				if !event.Success || event.Bytes == 0 || event.Before == nil || event.After == nil {
					t.Fatal("fault did not follow a completed actual assessment read")
				}
				fired = true
				if scenario == "manifest same inode changed bytes" {
					if err := os.WriteFile(path, changed, 0600); err != nil {
						t.Fatal(err)
					}
				} else {
					changed = original
					replacement := path + ".replacement"
					if err := os.WriteFile(replacement, changed, 0600); err != nil {
						t.Fatal(err)
					}
					if err := os.Rename(replacement, path); err != nil {
						t.Fatal(err)
					}
				}
			}
			var output, diagnostic bytes.Buffer
			code := runSourceDemandsObserved(t.Context(), []string{"source-demands", "--repo", repo}, &output, &diagnostic, baseline, observer)
			if !fired {
				t.Fatal("actual command did not reach the disclosed completed-read fault frontier")
			}
			after, err := os.Stat(path)
			if err != nil {
				t.Fatal(err)
			}
			wantSameInode := scenario == "manifest same inode changed bytes"
			if os.SameFile(before, after) != wantSameInode {
				t.Fatal("fixture failed to establish the distinct inode/content substitution")
			}
			actual, err := os.ReadFile(path)
			if err != nil || !bytes.Equal(actual, changed) {
				t.Fatal("command rewrote or discarded the later replacement")
			}
			if code == 0 || output.Len() != 0 {
				t.Errorf("actual command emitted successful stale observations after %s", scenario)
			}
		})
	}
}
