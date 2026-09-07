package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
)

func TestSourceFoundationMissingCLIAdmission155(t *testing.T) {
	for _, missing := range []bool{false, true} {
		t.Run(map[bool]string{false: "materialized_control", true: "actual_missing_cli"}[missing], func(t *testing.T) {
			repo, baseline, _ := sourceFoundationCombinedRetained155(t, missing)
			original, err := os.ReadFile(filepath.Join(repo, sourceLaneManifestPath))
			if err != nil {
				t.Fatal(err)
			}
			var out, diag bytes.Buffer
			code := runSourceDemandsPolicy(t.Context(), []string{"source-demands", "--repo", repo}, &out, &diag, baseline)
			if code != 0 {
				t.Fatalf("actual authoring consumer must preserve complete source and local work: %s", diag.String())
			}
			var got sourceFoundationRegister
			if json.Unmarshal(out.Bytes(), &got) != nil || got.Coverage.UniverseCount != 14 || len(got.Known) != 1 || len(got.Coverage.Assessed) != 1 || len(got.Coverage.Unassessed) != 13 {
				t.Fatal("independent14-cell source/K coverage lost")
			}
			want := 2
			if missing {
				want = 3
			}
			if len(got.Requirements) != want || len(got.Adopters) != want {
				t.Fatal("body/auth/local aspects lost")
			}
			after, err := os.ReadFile(filepath.Join(repo, sourceLaneManifestPath))
			if err != nil || !bytes.Equal(after, original) {
				t.Fatal("authoring consumer rewrote source producer truth")
			}
		})
	}
}

func sourceFoundationAdmissionCommand155(t *testing.T, repo string, baseline sourceArtifactPin) []byte {
	t.Helper()
	var out, diag bytes.Buffer
	if code := runSourceDemandsPolicy(t.Context(), []string{"source-demands", "--repo", repo}, &out, &diag, baseline); code != 0 {
		t.Fatal(diag.String())
	}
	return out.Bytes()
}

func TestSourceFoundationAdmissionTruth155(t *testing.T) {
	for _, missing := range []bool{false, true} {
		t.Run(map[bool]string{false: "materialized", true: "missing"}[missing], func(t *testing.T) {
			repo, baseline, _ := sourceFoundationCombinedRetained155(t, missing)
			original, err := os.ReadFile(filepath.Join(repo, sourceLaneManifestPath))
			if err != nil {
				t.Fatal(err)
			}
			var source sourceLaneManifest
			if json.Unmarshal(original, &source) != nil {
				t.Fatal("source report invalid JSON")
			}
			raw := sourceFoundationAdmissionCommand155(t, repo, baseline)
			var got sourceFoundationRegister
			if json.Unmarshal(raw, &got) != nil {
				t.Fatal("register invalid JSON")
			}
			if !reflect.DeepEqual(got.SourceAdmission.CurrentValidation, source.Validation) || !reflect.DeepEqual(got.SourceAdmission.CurrentDiagnostics, source.Diagnostics) {
				t.Fatal("actual current validation/diagnostics changed")
			}
			if missing {
				if got.SourceAdmission.Kind != "canonical_intended_missing_cli" || len(got.SourceAdmission.MissingCLI) != 1 {
					t.Fatal("missing exact authoring admission")
				}
				g := got.SourceAdmission.MissingCLI[0]
				if g.PartialGeneration == g.CanonicalGeneration || len(g.Affected) != 1 || len(g.DiagnosticIndices) != 2 || g.Affected[0].Binding.ID != "widgets create" || !reflect.DeepEqual(g.Affected[0].LocalRequirementIDs, []string{"local-command-artifact"}) {
					t.Fatal("missing CLI admission coordinates wrong")
				}
				for _, pin := range got.Inputs {
					if pin.Path == g.Path {
						t.Fatal("intended CLI pin promoted to actual input")
					}
				}
				var out, diag bytes.Buffer
				if code := runSourceLanesContext(t.Context(), []string{"source-lanes", "--repo", repo, "--check", "--manifest", sourceLaneManifestPath}, &out, &diag); code != 1 {
					t.Fatalf("invalid source-lanes check exit=%d: %s", code, diag.String())
				}
			} else if got.SourceAdmission.Kind != "current_valid" || len(got.SourceAdmission.MissingCLI) != 0 {
				t.Fatal("valid source needed exception")
			}
			if second := sourceFoundationAdmissionCommand155(t, repo, baseline); !bytes.Equal(raw, second) {
				t.Fatal("whole generated output nondeterministic")
			}
			project, err := repoRoot()
			if err != nil {
				t.Fatal(err)
			}
			candidate := filepath.Join(t.TempDir(), "candidate.json")
			if err = os.WriteFile(candidate, raw, 0600); err != nil {
				t.Fatal(err)
			}
			result, err := exec.CommandContext(t.Context(), "python3", filepath.Join(project, ".planning/phases/cp13-foundation-demand-register/register-schema-check.py"), "--document", candidate).CombinedOutput()
			t.Logf("actual full source admission schema: %s", result)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestSourceFoundationAdmissionCompanion155(t *testing.T) {
	for _, fault := range []string{"missing_row", "missing_path", "unresolved_companion", "all_unresolved"} {
		t.Run(fault, func(t *testing.T) {
			repo, baseline, assessment := sourceFoundationCombinedRetained155(t, true)
			sourceFoundationAdmissionCommand155(t, repo, baseline)
			local := &assessment.Assessments[0].Requirements[2]
			switch fault {
			case "missing_row":
				assessment.Assessments[0].Requirements = assessment.Assessments[0].Requirements[:2]
			case "missing_path":
				local.AffectedArtifacts = []string{"internal/connectors/defs/acme/operations.json"}
			case "unresolved_companion":
				local.Assessment = "unresolved"
				local.ProofIDs = []string{"missing-proof"}
			case "all_unresolved":
				for i := range assessment.Assessments[0].Requirements {
					r := &assessment.Assessments[0].Requirements[i]
					r.Assessment = "unresolved"
					r.ProofIDs = []string{"missing-proof"}
				}
			}
			writeSourceFoundationAssessmentFixture(t, repo, assessment)
			var out, diag bytes.Buffer
			code := runSourceDemandsPolicy(t.Context(), []string{"source-demands", "--repo", repo}, &out, &diag, baseline)
			if fault == "all_unresolved" {
				var got sourceFoundationRegister
				if code != 0 || json.Unmarshal(out.Bytes(), &got) != nil || len(got.Requirements) != 3 || len(got.Adopters) != 0 {
					t.Fatalf("unresolved work accounting lost: %s", diag.String())
				}
				for _, r := range got.Requirements {
					if r.Status != "unresolved" {
						t.Fatal("missing proof authorized reuse")
					}
				}
			} else if code == 0 || out.Len() != 0 {
				t.Fatal("missing/unverified companion authorized resolved shared fit")
			}
		})
	}
}

func TestSourceFoundationAdmissionMalformedSibling155(t *testing.T) {
	for _, fault := range []string{"missing_schema", "missing_spec", "changed_operation", "extra_schema", "symlink_cli", "directory_cli", "empty_cli"} {
		t.Run(fault, func(t *testing.T) {
			repo, baseline, _ := sourceFoundationCombinedRetained155(t, true)
			sourceFoundationAdmissionCommand155(t, repo, baseline)
			prefix := "internal/connectors/defs/acme/"
			switch fault {
			case "missing_schema":
				if err := os.Remove(filepath.Join(repo, prefix+"schemas/request.json")); err != nil {
					t.Fatal(err)
				}
			case "missing_spec":
				if err := os.Remove(filepath.Join(repo, prefix+"spec.json")); err != nil {
					t.Fatal(err)
				}
			case "changed_operation":
				proofWrite(t, repo, prefix+"operations.json", []byte(`{"operations":[]}`))
			case "extra_schema":
				proofWrite(t, repo, prefix+"schemas/extra.json", []byte(`{"type":"string"}`))
			case "symlink_cli":
				if err := os.Symlink("operations.json", filepath.Join(repo, prefix+"cli_surface.json")); err != nil {
					t.Fatal(err)
				}
			case "directory_cli":
				if err := os.Mkdir(filepath.Join(repo, prefix+"cli_surface.json"), 0700); err != nil {
					t.Fatal(err)
				}
			case "empty_cli":
				proofWrite(t, repo, prefix+"cli_surface.json", []byte{})
			}
			var out, diag bytes.Buffer
			if code := runSourceDemandsPolicy(t.Context(), []string{"source-demands", "--repo", repo}, &out, &diag, baseline); code == 0 || out.Len() != 0 {
				t.Fatal("unrelated/malformed artifact admitted as CLI absence")
			}
		})
	}
}

func TestSourceFoundationAdmissionForgedSource155(t *testing.T) {
	for _, fault := range []string{"status", "counts", "pointer", "severity", "owner", "duplicate", "remove", "key"} {
		t.Run(fault, func(t *testing.T) {
			repo, baseline, _ := sourceFoundationCombinedRetained155(t, true)
			sourceFoundationAdmissionCommand155(t, repo, baseline)
			raw, err := os.ReadFile(filepath.Join(repo, sourceLaneManifestPath))
			if err != nil {
				t.Fatal(err)
			}
			var source sourceLaneManifest
			if json.Unmarshal(raw, &source) != nil {
				t.Fatal("source JSON")
			}
			index := -1
			for i, d := range source.Diagnostics {
				if d.Code == "execution_generation_mismatch" {
					index = i
				}
			}
			if index < 0 {
				t.Fatal("actual reference error absent")
			}
			switch fault {
			case "status":
				source.Validation.Status = "valid"
			case "counts":
				source.Validation.Errors = 0
			case "pointer":
				source.Diagnostics[index].Pointer = "unrelated"
			case "severity":
				source.Diagnostics[index].Severity = "deficit"
			case "owner":
				source.Diagnostics[index].Owner = "unrelated"
			case "duplicate":
				source.Diagnostics = append(source.Diagnostics, source.Diagnostics[index])
			case "remove":
				source.Diagnostics = append(source.Diagnostics[:index], source.Diagnostics[index+1:]...)
			case "key":
				source.SourceOperations[0].Source.Key.ID = "same-count-substitute"
			}
			raw, err = json.Marshal(source)
			if err != nil {
				t.Fatal(err)
			}
			proofWrite(t, repo, sourceLaneManifestPath, raw)
			var out, diag bytes.Buffer
			if code := runSourceDemandsPolicy(t.Context(), []string{"source-demands", "--repo", repo}, &out, &diag, baseline); code == 0 || out.Len() != 0 {
				t.Fatal("saved diagnostic/source forgery supplied admission authority")
			}
		})
	}
}

func TestSourceFoundationAdmissionValidManifestSibling155(t *testing.T) {
	repo, baseline, assessment := sourceFoundationCombinedRetained155(t, true)
	var envelope struct {
		SchemaVersion int                        `json:"schema_version"`
		Annotations   []sourceSemanticAnnotation `json:"annotations"`
	}
	raw, err := os.ReadFile(filepath.Join(repo, sourceLaneAnnotationsPath))
	if err != nil || json.Unmarshal(raw, &envelope) != nil {
		t.Fatal("annotations")
	}
	command := envelope.Annotations[0].IntendedBindings[1]
	envelope.Annotations[0].IntendedBindings = []sourceLaneTargetRef{command}
	raw, err = json.Marshal(envelope)
	if err != nil {
		t.Fatal(err)
	}
	proofWrite(t, repo, sourceLaneAnnotationsPath, raw)
	for i := range assessment.Assessments[0].Requirements {
		assessment.Assessments[0].Requirements[i].FitBindings = []sourceLaneTargetRef{command}
	}
	writeSourceFoundationAssessmentFixture(t, repo, assessment)
	var source, diag bytes.Buffer
	if code := runSourceLanesContext(t.Context(), []string{"source-lanes", "--repo", repo}, &source, &diag); code != 0 {
		t.Fatal(diag.String())
	}
	proofWrite(t, repo, sourceLaneManifestPath, source.Bytes())
	raw = sourceFoundationAdmissionCommand155(t, repo, baseline)
	var got sourceFoundationRegister
	if json.Unmarshal(raw, &got) != nil || got.SourceAdmission.CurrentValidation.Status != "valid" || got.SourceAdmission.Kind != "canonical_intended_missing_cli" || len(got.SourceAdmission.MissingCLI) != 1 || len(got.Adopters) != 3 {
		t.Fatal("valid-manifest intended fit skipped its physical witness")
	}
	for _, r := range got.Requirements {
		if len(r.MechanismFits) != 1 || r.MechanismFits[0].DeclarationState != "canonical_intended" {
			t.Fatal("valid manifest mislabeled missing command materialized")
		}
	}
}

func TestSourceFoundationAdmissionCustody155(t *testing.T) {
	for _, fault := range []string{"same_bytes_new_inode", "changed_source", "cli_appears", "extra_schema", "cancel_initial", "cancel_final"} {
		t.Run(fault, func(t *testing.T) {
			repo, baseline, _ := sourceFoundationCombinedRetained155(t, true)
			sourceFoundationAdmissionCommand155(t, repo, baseline)
			name := "internal/connectors/defs/acme/source.lock.json"
			full := filepath.Join(repo, name)
			original, err := os.ReadFile(full)
			if err != nil {
				t.Fatal(err)
			}
			before, err := os.Stat(full)
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			fired := false
			var event sourceProofReadEvent
			var out, diag bytes.Buffer
			code := runSourceDemandsObserved(ctx, []string{"source-demands", "--repo", repo}, &out, &diag, baseline, func(e sourceProofReadEvent) {
				if fired || !e.Success {
					return
				}
				if fault == "cancel_initial" {
					if e.Path != name || e.Phase != "initial" {
						return
					}
				} else if e.Phase != "final" {
					return
				}
				fired = true
				event = e
				switch fault {
				case "same_bytes_new_inode":
					proofWrite(t, repo, "replacement.json", original)
					if err := os.Rename(filepath.Join(repo, "replacement.json"), full); err != nil {
						t.Fatal(err)
					}
				case "changed_source":
					proofWrite(t, repo, name, append(append([]byte{}, original...), byte(' ')))
				case "cli_appears":
					proofWrite(t, repo, "internal/connectors/defs/acme/cli_surface.json", []byte(`{"actor":true}`))
				case "extra_schema":
					proofWrite(t, repo, "internal/connectors/defs/acme/schemas/actor.json", []byte(`{"type":"string"}`))
				case "cancel_initial", "cancel_final":
					cancel()
				}
			})
			if !fired || event.Bytes == 0 || event.Before == nil || event.After == nil || code == 0 || out.Len() != 0 {
				t.Fatalf("actual custody frontier not reached/refused: %+v code=%d %s", event, code, diag.String())
			}
			after, err := os.Stat(full)
			if err != nil {
				t.Fatal(err)
			}
			if fault == "same_bytes_new_inode" && os.SameFile(before, after) {
				t.Fatal("replacement inode not reached")
			}
			if fault == "changed_source" && !os.SameFile(before, after) {
				t.Fatal("same-inode mutation not reached")
			}
			want := original
			if fault == "changed_source" {
				want = append(append([]byte{}, original...), byte(' '))
			}
			actual, err := os.ReadFile(full)
			if err != nil || !bytes.Equal(actual, want) {
				t.Fatal("actor source bytes modified")
			}
			if fault == "cli_appears" {
				actual, err = os.ReadFile(filepath.Join(repo, "internal/connectors/defs/acme/cli_surface.json"))
				if err != nil || string(actual) != `{"actor":true}` {
					t.Fatal("actor CLI file modified")
				}
			}
			if fault == "extra_schema" {
				actual, err = os.ReadFile(filepath.Join(repo, "internal/connectors/defs/acme/schemas/actor.json"))
				if err != nil || string(actual) != `{"type":"string"}` {
					t.Fatal("actor schema modified")
				}
			}
		})
	}
}

func TestSourceFoundationAdmissionSavedChecks155(t *testing.T) {
	repo, baseline, _ := sourceFoundationCombinedRetained155(t, true)
	raw := sourceFoundationAdmissionCommand155(t, repo, baseline)
	for _, name := range []string{sourceFoundationRegisterPath, "candidate.json"} {
		t.Run(name, func(t *testing.T) {
			proofWrite(t, repo, name, raw)
			before, err := os.Stat(filepath.Join(repo, name))
			if err != nil {
				t.Fatal(err)
			}
			var out, diag bytes.Buffer
			args := []string{"source-demands", "--repo", repo, "--check"}
			if name != sourceFoundationRegisterPath {
				args = append(args, name)
			}
			if code := runSourceDemandsPolicy(t.Context(), args, &out, &diag, baseline); code != 0 || out.Len() != 0 {
				t.Fatalf("actual saved check: %s", diag.String())
			}
			after, err := os.Stat(filepath.Join(repo, name))
			if err != nil || !os.SameFile(before, after) || before.Mode() != after.Mode() || before.ModTime() != after.ModTime() {
				t.Fatal("saved check changed file metadata")
			}
			actual, err := os.ReadFile(filepath.Join(repo, name))
			if err != nil || !bytes.Equal(actual, raw) {
				t.Fatal("saved check changed bytes")
			}
		})
	}
	for _, fault := range []string{"validation", "diagnostic", "intended_pin", "local_join", "declaration_state"} {
		t.Run(fault, func(t *testing.T) {
			var got sourceFoundationRegister
			if json.Unmarshal(raw, &got) != nil {
				t.Fatal("register")
			}
			switch fault {
			case "validation":
				got.SourceAdmission.CurrentValidation.Status = "valid"
			case "diagnostic":
				got.SourceAdmission.CurrentDiagnostics = []sourceLaneDiagnostic{}
			case "intended_pin":
				got.Inputs = append(got.Inputs, got.SourceAdmission.MissingCLI[0].IntendedCLI)
			case "local_join":
				got.SourceAdmission.MissingCLI[0].Affected[0].LocalRequirementIDs = []string{"forged"}
			case "declaration_state":
				got.Requirements[0].MechanismFits[0].DeclarationState = "materialized"
			}
			altered, err := json.Marshal(got)
			if err != nil {
				t.Fatal(err)
			}
			proofWrite(t, repo, "forged.json", altered)
			var out, diag bytes.Buffer
			if code := runSourceDemandsPolicy(t.Context(), []string{"source-demands", "--repo", repo, "--check", "forged.json"}, &out, &diag, baseline); code == 0 || out.Len() != 0 {
				t.Fatal("coherent saved admission forgery accepted")
			}
		})
	}
}

func TestSourceFoundationAdmissionWholeSurface155(t *testing.T) {
	for _, profile := range []string{"joined", "unjoined"} {
		t.Run(profile, func(t *testing.T) {
			repo, baseline, _ := sourceFoundationCombinedRetainedProfile155(t, true, profile)
			var out, diag bytes.Buffer
			code := runSourceDemandsPolicy(t.Context(), []string{"source-demands", "--repo", repo}, &out, &diag, baseline)
			if profile == "unjoined" {
				if code == 0 || out.Len() != 0 {
					t.Fatal("unjoined lost command hidden by one convenient source binding")
				}
				return
			}
			var got sourceFoundationRegister
			if code != 0 || json.Unmarshal(out.Bytes(), &got) != nil {
				t.Fatalf("complete two-command ownership refused: %s", diag.String())
			}
			if len(got.Requirements) != 4 || len(got.Adopters) != 4 || len(got.SourceAdmission.MissingCLI) != 1 || len(got.SourceAdmission.MissingCLI[0].Affected) != 2 {
				t.Fatal("complete lost command surface accounting collapsed")
			}
			want := map[string]string{"widgets create": "local-command-artifact", "widgets create-secondary": "local-secondary-command-artifact"}
			for _, a := range got.SourceAdmission.MissingCLI[0].Affected {
				if !reflect.DeepEqual(a.LocalRequirementIDs, []string{want[a.Binding.ID]}) {
					t.Fatal("command local ownership swapped")
				}
				delete(want, a.Binding.ID)
			}
			if len(want) != 0 {
				t.Fatal("declared command omitted")
			}
		})
	}
}

func TestSourceFoundationAdmissionIndependentIdentity155(t *testing.T) {
	repo, baseline, _ := sourceFoundationCombinedRetained155(t, true)
	sourceFoundationAdmissionCommand155(t, repo, baseline)
	raw, err := os.ReadFile(filepath.Join(repo, "source.json"))
	if err != nil {
		t.Fatal(err)
	}
	raw = bytes.ReplaceAll(raw, []byte(`"source.a"`), []byte(`"source.c"`))
	proofWrite(t, repo, "source.json", raw)
	var cohort sourceLaneCohort
	cohortRaw, err := os.ReadFile(filepath.Join(repo, sourceLaneCohortPath))
	if err != nil || json.Unmarshal(cohortRaw, &cohort) != nil {
		t.Fatal("cohort")
	}
	cohort.Inventories[0].SHA256 = sourceBytesHash(raw)
	cohort.Inventories[0].ExpectedIDs = []string{"source.c", "source.b"}
	cohortRaw, err = json.Marshal(cohort)
	if err != nil {
		t.Fatal(err)
	}
	proofWrite(t, repo, sourceLaneCohortPath, cohortRaw)
	for _, name := range []string{sourceLaneAnnotationsPath, sourceFoundationAssessmentsPath} {
		raw, err = os.ReadFile(filepath.Join(repo, name))
		if err != nil {
			t.Fatal(err)
		}
		raw = bytes.ReplaceAll(raw, []byte(`"source.a"`), []byte(`"source.c"`))
		proofWrite(t, repo, name, raw)
	}
	var source, diag bytes.Buffer
	if code := runSourceLanesContext(t.Context(), []string{"source-lanes", "--repo", repo}, &source, &diag); code != 1 || source.Len() == 0 {
		t.Fatalf("actual source substitution not produced: %s", diag.String())
	}
	proofWrite(t, repo, sourceLaneManifestPath, source.Bytes())
	var out bytes.Buffer
	diag.Reset()
	if code := runSourceDemandsPolicy(t.Context(), []string{"source-demands", "--repo", repo}, &out, &diag, baseline); code == 0 || out.Len() != 0 {
		t.Fatal("same-count source replacement manufactured independent baseline")
	}
}

func TestSourceFoundationAdmissionStandaloneCustody155(t *testing.T) {
	repo, _, inputs, _ := sourceFoundationCombinedFixture153(t, true)
	if _, err := buildSourceFoundationRegister(t.Context(), repo, inputs); err != nil {
		t.Fatal(err)
	}
	name := "internal/connectors/defs/acme/source.lock.json"
	full := filepath.Join(repo, name)
	before, err := os.Stat(full)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(full)
	if err != nil {
		t.Fatal(err)
	}
	proofWrite(t, repo, "actor.json", raw)
	if err := os.Rename(filepath.Join(repo, "actor.json"), full); err != nil {
		t.Fatal(err)
	}
	after, err := os.Stat(full)
	if err != nil || os.SameFile(before, after) {
		t.Fatal("replacement not reached")
	}
	if got, err := buildSourceFoundationRegister(t.Context(), repo, inputs); err == nil || got.Kind != "" {
		t.Fatal("standalone universe retained stale inode permission")
	}
	if got, err := buildSourceFoundationRequirements(t.Context(), repo, inputs.assessments); err == nil || len(got) != 0 {
		t.Fatal("standalone requirements retained stale inode permission")
	}
	actual, err := os.ReadFile(full)
	if err != nil || !bytes.Equal(actual, raw) {
		t.Fatal("actor-owned replacement changed")
	}
}

func TestSourceFoundationAdmissionDuplicateAnnotation156(t *testing.T) {
	repo, baseline, _ := sourceFoundationCombinedRetained155(t, true)
	sourceFoundationAdmissionCommand155(t, repo, baseline)
	raw, err := os.ReadFile(filepath.Join(repo, sourceLaneAnnotationsPath))
	if err != nil {
		t.Fatal(err)
	}
	var document struct {
		SchemaVersion int                        `json:"schema_version"`
		Annotations   []sourceSemanticAnnotation `json:"annotations"`
	}
	if json.Unmarshal(raw, &document) != nil || len(document.Annotations) != 1 {
		t.Fatal("original annotation control")
	}
	document.Annotations = append(document.Annotations, document.Annotations[0])
	raw, err = json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	proofWrite(t, repo, sourceLaneAnnotationsPath, raw)
	var out, diag bytes.Buffer
	if code := runSourceDemandsPolicy(t.Context(), []string{"source-demands", "--repo", repo}, &out, &diag, baseline); code == 0 || out.Len() != 0 {
		t.Fatal("duplicate current source annotation admitted missing CLI generation")
	}
	after, err := os.ReadFile(filepath.Join(repo, sourceLaneAnnotationsPath))
	if err != nil || !bytes.Equal(raw, after) {
		t.Fatal("refusal changed duplicate actor input")
	}
}
