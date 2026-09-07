package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestSourceFoundationDecisionCommandParity150(t *testing.T) {
	rawCases, err := os.ReadFile("testdata/source-foundation-decision-150.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Name  string `json:"name"`
		Wire  string `json:"wire"`
		Valid bool   `json:"valid"`
	}
	if err := json.Unmarshal(rawCases, &cases); err != nil {
		t.Fatal(err)
	}
	for _, test := range cases {
		t.Run(test.Name, func(t *testing.T) {
			repo, baseline := sourceFoundationCommandFixture(t)
			args := []string{"source-demands", "--repo", repo}
			var out, diag bytes.Buffer
			if code := runSourceDemandsPolicy(t.Context(), args, &out, &diag, baseline); code != 0 {
				t.Fatalf("healthy source/assessment/register command frontier: %s", diag.String())
			}
			var control sourceFoundationRegister
			if err := json.Unmarshal(out.Bytes(), &control); err != nil || len(control.Requirements) != 1 {
				t.Fatalf("healthy complete register: %v", err)
			}
			raw, err := os.ReadFile(filepath.Join(repo, sourceFoundationAssessmentsPath))
			if err != nil {
				t.Fatal(err)
			}
			needle := []byte(`"decision_refs":[]`)
			if bytes.Count(raw, needle) != 1 {
				t.Fatal("exact single fixture decision array not found")
			}
			raw = bytes.Replace(raw, needle, []byte(`"decision_refs":`+test.Wire), 1)
			proofWrite(t, repo, sourceFoundationAssessmentsPath, raw)
			out.Reset()
			diag.Reset()
			code := runSourceDemandsPolicy(t.Context(), args, &out, &diag, baseline)
			if !test.Valid {
				if code == 0 || out.Len() != 0 || diag.Len() == 0 {
					t.Fatalf("invalid decision accepted or emitted partial register: code=%d bytes=%d diagnostic=%s", code, out.Len(), diag.String())
				}
				return
			}
			var got sourceFoundationRegister
			if code != 0 || json.Unmarshal(out.Bytes(), &got) != nil || len(got.Requirements) != 1 {
				t.Fatalf("valid decision refused: code=%d diagnostic=%s", code, diag.String())
			}
			var want []sourceFoundationDecision
			if err := json.Unmarshal([]byte(test.Wire), &want); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got.Requirements[0].DecisionRefs, want) ||
				!reflect.DeepEqual(got.Coverage, control.Coverage) || got.Requirements[0].Identity != control.Requirements[0].Identity ||
				got.Requirements[0].NextOwner != control.Requirements[0].NextOwner || len(got.Adopters) != 0 {
				t.Fatal("valid decision normalized or changed source/owner/capability")
			}
		})
	}
}

func TestSourceFoundationDecisionSavedRegister150(t *testing.T) {
	repo, universe, document, _ := sourceFoundationRequirementsFixture(t)
	document.Assessments[0].Requirements[0].DecisionRefs = []sourceFoundationDecision{{
		ID: "cli-batch1-vercel-inbound-sync-decision-r1", State: "pending", Condition: "",
	}}
	writeSourceFoundationAssessmentFixture(t, repo, document)
	_, pin := sourceFoundationObligationsFixture(t, repo, universe)
	inputs, err := observeSourceFoundationDemandInputs(t.Context(), repo, sourceFoundationAssessmentsPath, universe, pin)
	if err != nil {
		t.Fatal(err)
	}
	control, err := buildSourceFoundationRegister(t.Context(), repo, inputs)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(control)
	if err != nil {
		t.Fatal(err)
	}
	if err := validateSourceFoundationRegister(t.Context(), repo, raw, inputs); err != nil {
		t.Fatal(err)
	}
	for _, mutation := range []string{"missing", "null"} {
		t.Run(mutation, func(t *testing.T) {
			var wire map[string]any
			if err := json.Unmarshal(raw, &wire); err != nil {
				t.Fatal(err)
			}
			decision := wire["requirements"].([]any)[0].(map[string]any)["decision_refs"].([]any)[0].(map[string]any)
			if decision["condition"] != "" {
				t.Fatal("explicit-empty decision not reached in saved register")
			}
			if mutation == "missing" {
				delete(decision, "condition")
			} else {
				decision["condition"] = nil
			}
			// Summary counts remain independently correct: this mutation changes
			// only nested wire presence, not the number or identity of demands.
			changed, err := json.Marshal(wire)
			if err != nil {
				t.Fatal(err)
			}
			if err := validateSourceFoundationRegister(t.Context(), repo, changed, inputs); err == nil {
				t.Fatal("saved register erased required nested decision condition")
			}
		})
	}
}
