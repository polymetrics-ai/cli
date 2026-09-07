package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestSourceFoundationProofRecordOrder150(t *testing.T) {
	for _, reverse := range []bool{false, true} {
		name := "original_valid_order"
		if reverse {
			name = "reversed_same_valid_records"
		}
		t.Run(name, func(t *testing.T) {
			repo, document := sourceFoundationProofObservationFixture(t)
			if len(document.Records) != 2 {
				t.Fatal("expected the two original reviewed records")
			}
			want := map[string]sourceFoundationProofRecord{}
			for _, record := range document.Records {
				want[record.ID] = record
			}
			// The actual Check capture includes the Mode owner as an input.
			// Reversing these same records reaches input-before-owner parsing.
			overlap := false
			for _, input := range document.Records[1].Inputs {
				for _, owner := range document.Records[0].OwnerSymbols {
					overlap = overlap || input.Path == owner.File
				}
			}
			if !overlap {
				t.Fatal("actual input-to-later-owner overlap not reached")
			}
			if reverse {
				document.Records[0], document.Records[1] = document.Records[1], document.Records[0]
			}
			writeSourceFoundationObservationFixture(t, repo, document)
			got, err := readSourceFoundationProofObservations(t.Context(), repo)
			if err != nil {
				t.Fatalf("same valid proof records must remain current: %v", err)
			}
			if len(got) != len(want) {
				t.Fatalf("observations=%d, want two exact records", len(got))
			}
			for _, proof := range got {
				if !reflect.DeepEqual(proof.record, want[proof.record.ID]) || proof.status != "current" {
					t.Errorf("record identity or current assertion changed for %s: %s", proof.record.ID, proof.status)
				}
				delete(want, proof.record.ID)
			}
			if len(want) != 0 {
				t.Fatal("record omitted or duplicated")
			}
		})
	}
}

func TestSourceFoundationDecisionRequiredCondition150(t *testing.T) {
	for _, mode := range []string{"explicit_empty", "absent", "null"} {
		t.Run(mode, func(t *testing.T) {
			repo, universe, document, _ := sourceFoundationRequirementsFixture(t)
			want := []sourceFoundationDecision{{ID: "cli-batch1-vercel-inbound-sync-decision-r1", State: "pending", Condition: ""}}
			document.Assessments[0].Requirements[0].DecisionRefs = want
			raw := writeSourceFoundationAssessmentFixture(t, repo, document)
			observed, err := observeSourceFoundationAssessments(t.Context(), repo, sourceFoundationAssessmentsPath, universe)
			if err != nil {
				t.Fatal(err)
			}
			control, err := buildSourceFoundationRequirements(t.Context(), repo, observed)
			if err != nil || len(control) != 1 || !reflect.DeepEqual(control[0].DecisionRefs, want) {
				t.Fatalf("explicit-empty pending decision control: %+v, %v", control, err)
			}
			if mode == "explicit_empty" {
				return
			}
			var wire map[string]any
			if err := json.Unmarshal(raw, &wire); err != nil {
				t.Fatal(err)
			}
			decision := wire["assessments"].([]any)[0].(map[string]any)["requirements"].([]any)[0].(map[string]any)["decision_refs"].([]any)[0].(map[string]any)
			if mode == "absent" {
				delete(decision, "condition")
			} else {
				decision["condition"] = nil
			}
			raw, err = json.Marshal(wire)
			if err != nil {
				t.Fatal(err)
			}
			proofWrite(t, repo, sourceFoundationAssessmentsPath, raw)
			observed, err = observeSourceFoundationAssessments(t.Context(), repo, sourceFoundationAssessmentsPath, universe)
			if err != nil {
				return
			}
			got, err := buildSourceFoundationRequirements(t.Context(), repo, observed)
			if err == nil {
				t.Fatalf("required non-null condition %s accepted: %+v", mode, got)
			}
		})
	}
}

func sourceFoundationUnavailableFault150(t *testing.T, repo, fault string) {
	t.Helper()
	name := sourceFoundationProofPath
	if fault == "missing_dependency" || fault == "stale_dependency" {
		name = "go.sum"
	}
	if fault == "missing_selected_record" {
		raw, err := os.ReadFile(filepath.Join(repo, name))
		if err != nil {
			t.Fatal(err)
		}
		document, err := decodeSourceFoundationProofDocument(t.Context(), raw, 2)
		if err != nil {
			t.Fatal(err)
		}
		document.Records = []sourceFoundationProofRecord{}
		writeSourceFoundationObservationFixture(t, repo, document)
	} else if fault == "stale_dependency" {
		proofWrite(t, repo, name, []byte("readable changed dependency"))
	} else if err := os.Remove(filepath.Join(repo, name)); err != nil {
		t.Fatal(err)
	}
}

func assertSourceFoundationUnavailableRegister150(t *testing.T, got, control sourceFoundationRegister) {
	t.Helper()
	if !reflect.DeepEqual(got.Coverage, control.Coverage) || !reflect.DeepEqual(got.Known, control.Known) ||
		len(got.Requirements) != 1 || len(got.Adopters) != 0 {
		t.Fatalf("optional evidence changed independent U/A/K/demand/adopters: coverage=%+v requirements=%d", got.Coverage, len(got.Requirements))
	}
	r, want := got.Requirements[0], control.Requirements[0]
	if r.Identity != want.Identity || r.ID != want.ID || r.Status != "unresolved" || r.NextOwner != want.NextOwner ||
		!reflect.DeepEqual(r.SourceRefs, want.SourceRefs) || len(r.MissingEvidence) == 0 {
		t.Fatalf("unavailable proof erased source identity/owner/evidence or remained current: %+v", r)
	}
	raw, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var wire map[string]json.RawMessage
	if err := json.Unmarshal(raw, &wire); err != nil {
		t.Fatal(err)
	}
	var requested []string
	var issues []map[string]string
	if json.Unmarshal(wire["requested_proof_ids"], &requested) != nil || len(requested) != 1 || requested[0] != want.Proofs[0].ID ||
		json.Unmarshal(wire["proof_issues"], &issues) != nil || len(issues) == 0 || issues[0]["id"] != requested[0] {
		t.Fatal("unavailable proof lost exact requested ID or typed issue")
	}
	absent := issues[0]["code"] == "proof_document_missing" || issues[0]["code"] == "proof_record_missing"
	if absent && len(r.Proofs) != 0 || !absent && (len(r.Proofs) != 1 || r.Proofs[0].Status == "current") {
		t.Fatal("absent evidence fabricated a record or present unavailable evidence became current")
	}
}

func TestSourceFoundationUnavailableRegister150(t *testing.T) {
	for _, fault := range []string{"missing_document", "missing_selected_record", "missing_dependency", "stale_dependency"} {
		t.Run(fault, func(t *testing.T) {
			repo, inputs := sourceFoundationRegisterFixture(t)
			control, err := buildSourceFoundationRegister(t.Context(), repo, inputs)
			if err != nil || control.Coverage.UniverseCount != 14 || len(control.Requirements) != 1 || control.Requirements[0].Status != "unresolved" {
				t.Fatalf("healthy actual fourteen-cell register control: %v", err)
			}
			sourceFoundationUnavailableFault150(t, repo, fault)
			got, err := buildSourceFoundationRegister(t.Context(), repo, inputs)
			if err != nil {
				t.Fatalf("optional evidence erased independent register: %v", err)
			}
			assertSourceFoundationUnavailableRegister150(t, got, control)
		})
	}
}

func TestSourceFoundationUnavailableCommand150(t *testing.T) {
	for _, fault := range []string{"missing_document", "missing_selected_record", "missing_dependency", "stale_dependency"} {
		t.Run(fault, func(t *testing.T) {
			repo, baseline := sourceFoundationCommandFixture(t)
			var out, diag bytes.Buffer
			args := []string{"source-demands", "--repo", repo}
			if code := runSourceDemandsPolicy(t.Context(), args, &out, &diag, baseline); code != 0 {
				t.Fatalf("healthy actual command control: %s", diag.String())
			}
			var control sourceFoundationRegister
			if err := json.Unmarshal(out.Bytes(), &control); err != nil {
				t.Fatal(err)
			}
			if control.Coverage.UniverseCount != 14 || len(control.Requirements) != 1 {
				t.Fatal("actual full source/register control not reached")
			}
			sourceFoundationUnavailableFault150(t, repo, fault)
			out.Reset()
			diag.Reset()
			if code := runSourceDemandsPolicy(t.Context(), args, &out, &diag, baseline); code != 0 {
				t.Fatalf("optional proof blocks actual command: code=%d bytes=%d error=%s", code, out.Len(), diag.String())
			}
			var got sourceFoundationRegister
			if err := json.Unmarshal(out.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			assertSourceFoundationUnavailableRegister150(t, got, control)
		})
	}
}
