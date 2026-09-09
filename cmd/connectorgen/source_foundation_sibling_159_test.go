package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// WR-158-01 is coverage, not a claim of a pre-existing production failure.
func TestSourceFoundationRealOperationSibling159(t *testing.T) {
	for _, missing := range []bool{false, true} {
		t.Run(map[bool]string{false: "materialized", true: "canonical_intended"}[missing], func(t *testing.T) {
			repo, baseline, assessment := sourceFoundationCombinedRetainedProfile155(t, missing, "sibling159")
			control := sourceFoundationAdmissionCommand155(t, repo, baseline)
			var positive sourceFoundationRegister
			if err := json.Unmarshal(control, &positive); err != nil {
				t.Fatal(err)
			}
			want := 2
			if missing {
				want = 3
			}
			wantRequirements := want
			if missing {
				wantRequirements++
			}
			if len(positive.Requirements) != wantRequirements || len(positive.Adopters) != want {
				t.Fatal("legitimate operation lost body/auth/local fits")
			}
			for _, adopter := range positive.Adopters {
				if adopter.Identity.Key.ID != "source.a" || adopter.Binding.CanonicalID != "operation:widgets.create" {
					t.Fatal("positive adopter borrowed second operation")
				}
				sourceFoundationExpectedSelectors156(t, repo, adopter.Requirement, sourceFoundationMechanismFit{Selectors: adopter.Selectors})
			}
			var cohort sourceLaneCohort
			raw, err := os.ReadFile(filepath.Join(repo, sourceLaneCohortPath))
			if err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(raw, &cohort); err != nil {
				t.Fatal(err)
			}
			var annotations struct {
				Annotations []sourceSemanticAnnotation `json:"annotations"`
			}
			raw, err = os.ReadFile(filepath.Join(repo, sourceLaneAnnotationsPath))
			if err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(raw, &annotations); err != nil {
				t.Fatal(err)
			}
			if len(annotations.Annotations) != 2 || annotations.Annotations[1].Key.ID != "source.b" {
				t.Fatal("missing independent sibling source binding")
			}
			sibling := annotations.Annotations[1].IntendedBindings[0]
			if sibling.ID != "widgets.remove" || sibling.Pointer != "/operations/1" || sibling.CanonicalPointer != "/operations/1/operation" {
				t.Fatal("second canonical selector is not real")
			}
			for _, mechanism := range []int{0, 1} {
				t.Run([]string{"body", "auth"}[mechanism], func(t *testing.T) {
					universe, err := buildSourceFoundationUniverse(t.Context(), repo, cohort, annotations.Annotations)
					if err != nil {
						t.Fatal(err)
					}
					other := universe.manifest.SourceOperations[1]
					borrowed := assessment.Assessments[0].Requirements[mechanism]
					borrowed.ID = "sibling-borrowed-proof"
					borrowed.FitBindings = []sourceLaneTargetRef{sibling}
					borrowed.SourceRefs = []sourceFactRef{other.Facts.Refs["summary"]}
					// The sibling has no body and explicitly empty auth requirements. Its
					// own valid facts are supplied; it cannot borrow source.a's mechanism.
					if mechanism == 1 {
						borrowed.SourceRefs = append(borrowed.SourceRefs, other.Facts.Refs["security"], other.Facts.Refs["security_schemes"])
					}
					candidate := assessment
					candidate.Assessments = append([]sourceFoundationCellAssessment{}, assessment.Assessments...)
					if missing {
						candidate.Assessments[1].Requirements = append(append([]sourceFoundationRequirement{}, assessment.Assessments[1].Requirements...), borrowed)
					} else {
						candidate.Assessments = append(candidate.Assessments, sourceFoundationCellAssessment{Key: other.Source.Key, Lane: "direct_write", NextOwner: "polymetrics.ai/internal/connectors/engine", Requirements: []sourceFoundationRequirement{borrowed}})
					}
					writeSourceFoundationAssessmentFixture(t, repo, candidate)
					inputs, err := observeSourceFoundationDemandInputs(t.Context(), repo, sourceFoundationAssessmentsPath, universe, baseline)
					if err != nil {
						t.Fatalf("valid sibling setup failed before mechanism consumer: %v", err)
					}
					wantError := []string{"exact structured body source/selector fit missing", "exact static header auth source/selector fit missing"}[mechanism]
					requirements, err := buildSourceFoundationRequirements(t.Context(), repo, inputs.assessments)
					if err == nil || !strings.Contains(err.Error(), wantError) || len(requirements) != 0 {
						t.Fatalf("requirements borrowed same-file proof: %v", err)
					}
					register, err := buildSourceFoundationRegister(t.Context(), repo, inputs)
					if err == nil || !strings.Contains(err.Error(), wantError) || len(register.Adopters) != 0 {
						t.Fatalf("register emitted false sibling adopters: %v", err)
					}
					var out, diag bytes.Buffer
					if code := runSourceDemandsPolicy(t.Context(), []string{"source-demands", "--repo", repo}, &out, &diag, baseline); code == 0 || out.Len() != 0 || !strings.Contains(diag.String(), wantError) {
						t.Fatalf("command did not reach exact sibling mechanism refusal: %s", diag.String())
					}
					writeSourceFoundationAssessmentFixture(t, repo, assessment)
				})
			}
		})
	}
}

// Literal source/operation/command coordinates determine the exact expected
// diagnostic wire sequence. Counts are consequences, never copied observations.
func sourceFoundationMissingDiagnostics160(t *testing.T, got sourceLaneManifest, sibling bool) {
	t.Helper()
	if err := sourceFoundationMissingDiagnosticError160(got, sibling); err != nil {
		t.Fatal(err)
	}
}

func sourceFoundationMissingDiagnosticError160(got sourceLaneManifest, sibling bool) error {
	row := func(id, lane, stage, code, pointer, severity string) sourceLaneDiagnostic {
		return sourceLaneDiagnostic{Key: sourceOperationKey{Connector: "acme", Inventory: "primary", ID: id}, Lanes: []string{lane}, Stage: stage, Code: code, Pointer: pointer, Owner: "acme", Severity: severity}
	}
	prefix := "internal/connectors/defs/acme/"
	want := []sourceLaneDiagnostic{
		row("source.a", "direct_write", "proof", "proof_unavailable", "", "deficit"),
		row("source.a", "direct_write", "reference", "execution_generation_mismatch", prefix+"operations.json#/operations/0", "error"),
		row("source.a", "direct_write", "reference", "target_absent", prefix+"cli_surface.json#/commands/0", "deficit"),
		row("source.a", "reverse_etl", "proof", "proof_unavailable", "", "deficit"),
	}
	if sibling {
		want = append(want,
			row("source.b", "direct_write", "proof", "proof_unavailable", "", "deficit"),
			row("source.b", "direct_write", "reference", "execution_generation_mismatch", prefix+"operations.json#/operations/1", "error"),
			row("source.b", "direct_write", "reference", "target_absent", prefix+"cli_surface.json#/commands/1", "deficit"),
			row("source.b", "reverse_etl", "proof", "proof_unavailable", "", "deficit"))
	} else {
		want = append(want, row("source.b", "direct_read", "proof", "proof_unavailable", "", "deficit"))
	}
	validation := sourceLaneValidation{Status: "invalid"}
	for _, d := range want {
		if d.Severity == "error" {
			validation.Errors++
		} else {
			validation.Deficits++
		}
	}
	if !reflect.DeepEqual(got.Diagnostics, want) || got.Validation != validation {
		return fmt.Errorf("independently expected diagnostic tuples mismatch: want=%+v got=%+v validation=%+v", want, got.Diagnostics, got.Validation)
	}
	return nil
}

func TestSourceFoundationDiagnosticOracle160(t *testing.T) {
	repo, _, _ := sourceFoundationCombinedRetainedProfile155(t, true, "sibling159")
	raw, err := os.ReadFile(filepath.Join(repo, sourceLaneManifestPath))
	if err != nil {
		t.Fatal(err)
	}
	var original sourceLaneManifest
	if err := json.Unmarshal(raw, &original); err != nil {
		t.Fatal(err)
	}
	if err := sourceFoundationMissingDiagnosticError160(original, true); err != nil {
		t.Fatal(err)
	}
	for _, fault := range []string{"missing", "extra", "wrong_source", "wrong_pointer", "wrong_severity"} {
		t.Run(fault, func(t *testing.T) {
			got := original
			got.Diagnostics = append([]sourceLaneDiagnostic{}, original.Diagnostics...)
			switch fault {
			case "missing":
				got.Diagnostics = got.Diagnostics[1:]
			case "extra":
				got.Diagnostics = append(got.Diagnostics, got.Diagnostics[0])
			case "wrong_source":
				got.Diagnostics[5].Key.ID = "source.a"
			case "wrong_pointer":
				got.Diagnostics[5].Pointer = "internal/connectors/defs/acme/operations.json#/operations/0"
			case "wrong_severity":
				got.Diagnostics[5].Severity = "deficit"
			}
			if sourceFoundationMissingDiagnosticError160(got, true) == nil {
				t.Fatal("diagnostic oracle accepted independent tuple corruption")
			}
		})
	}
}
