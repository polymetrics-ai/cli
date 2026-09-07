package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func sourceFoundationRegisterFixture(t *testing.T) (string, sourceFoundationDemandInputs) {
	t.Helper()
	repo, universe, _, _ := sourceFoundationRequirementsFixture(t)
	_, pin := sourceFoundationObligationsFixture(t, repo, universe)
	inputs, err := observeSourceFoundationDemandInputs(t.Context(), repo, sourceFoundationAssessmentsPath, universe, pin)
	if err != nil {
		t.Fatal(err)
	}
	return repo, inputs
}

func TestSourceFoundationRegisterCurrentKnownRelation(t *testing.T) {
	repo, _, _, _ := sourceFoundationRequirementsFixture(t)
	_, cohort := sourceInventoryFixture(t, []string{"source.b", "source.a"}, 2)
	path := filepath.Join(repo, "source.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var source map[string]any
	if err := json.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}
	for _, value := range source["rest"].(map[string]any)["operations"].([]any) {
		row := value.(map[string]any)
		if row["id"] == "source.b" {
			row["method"] = "put"
			row["source_operation"].(map[string]any)["summary"] = "Update webhook headers"
		}
	}
	raw, err = json.Marshal(source)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	cohort.Inventories[0].SHA256 = sourceBytesHash(raw)
	universe, err := buildSourceFoundationUniverse(t.Context(), repo, cohort, nil)
	if err != nil {
		t.Fatal(err)
	}
	document := sourceFoundationAssessmentDocumentFixture(t, universe)
	additional := document.Assessments[0]
	additional.Key = universe.manifest.SourceOperations[1].Source.Key
	additional.Requirements = append([]sourceFoundationRequirement{}, additional.Requirements...)
	ref, required := sourceRegistrationDemand(universe.manifest.SourceOperations[1].Facts)
	if !required || additional.Key.ID != "source.b" {
		t.Fatal("actual current source did not establish the extra known obligation")
	}
	additional.Requirements[0].SourceRefs = []sourceFactRef{ref}
	document.Assessments = append(document.Assessments, additional)
	writeSourceFoundationAssessmentFixture(t, repo, document)
	_, pin := sourceFoundationObligationsFixture(t, repo, universe)
	inputs, err := observeSourceFoundationDemandInputs(t.Context(), repo, sourceFoundationAssessmentsPath, universe, pin)
	if err != nil {
		t.Fatal(err)
	}
	got, err := buildSourceFoundationRegister(t.Context(), repo, inputs)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Known) != 2 || got.Known[0].Identity.Key.ID != "source.a" || got.Known[1].Identity.Key.ID != "source.b" ||
		got.Known[1].Identity.Lane != "sync_transport" || got.Known[1].HistoricalReceiver || got.Known[1].SourceState != "mapped_unproven" {
		t.Fatal("actual register erased current source-owned update/header obligation outside the historical seed")
	}
}

func TestSourceFoundationRegisterObservation(t *testing.T) {
	repo, inputs := sourceFoundationRegisterFixture(t)
	before, err := json.Marshal(inputs.assessments.universe.manifest)
	if err != nil {
		t.Fatal(err)
	}
	got, err := buildSourceFoundationRegister(t.Context(), repo, inputs)
	if err != nil {
		t.Fatal(err)
	}
	if got.SchemaVersion != 1 || got.Kind != "foundation_demand_register" || got.SourceSHA256 != sourceBytesHash(before) ||
		got.Coverage.UniverseCount != 14 || len(got.Coverage.Assessed) != 1 || len(got.Coverage.Unassessed) != 13 ||
		len(got.Known) != 1 || got.Known[0].Identity.Key.ID != "source.a" || !got.Known[0].HistoricalReceiver || got.Known[0].SentryRegistration ||
		len(got.Requirements) != 1 || got.Requirements[0].Status != "unresolved" || len(got.Facets) != 4 || len(got.Examples) != 35 || len(got.Adopters) != 0 {
		t.Fatal("actual aggregate projection lost independent source/known/aspect/example scope")
	}
	if !reflect.DeepEqual(got.Known[0].SourceDocumentPins, inputs.baseline.Obligations[0].SourceDocumentPins) ||
		!reflect.DeepEqual(got.Known[0].SourceRefs, inputs.baseline.Obligations[0].SourceRefs) {
		t.Fatal("known obligation source custody lost")
	}
	if len(got.Proofs) != 2 || got.Proofs[0].Status != "current" || got.Proofs[1].Status != "current" ||
		!got.Checks.StructuralValid || !got.Checks.CoverageAccounted || !got.Checks.RequiredReconciliation || !got.Checks.SelectedReuseProofComplete ||
		got.Checks.SelectedReuseRequirements != 0 || got.Checks.UnresolvedRequirements != 1 || got.Checks.Authority != "authoring_consistency_only; no lane authority or checkpoint acceptance" {
		t.Fatal("available narrow proof observations and unresolved requirement accounting were conflated")
	}
	after, err := json.Marshal(inputs.assessments.universe.manifest)
	if err != nil || string(before) != string(after) {
		t.Fatal("foundation register mutated source/lane authority")
	}
	second, err := buildSourceFoundationRegister(t.Context(), repo, inputs)
	if err != nil || !reflect.DeepEqual(got, second) {
		t.Fatal("same retained inputs produced different complete register observations")
	}
	raw, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	if err := validateSourceFoundationRegister(t.Context(), repo, raw, inputs); err != nil {
		t.Fatalf("actual complete closed register/independent membership reader: %v", err)
	}
}

func TestSourceFoundationRegisterIndependentOutputOracle(t *testing.T) {
	repo, inputs := sourceFoundationRegisterFixture(t)
	control, err := buildSourceFoundationRegister(t.Context(), repo, inputs)
	if err != nil {
		t.Fatal(err)
	}
	controlRaw, err := json.Marshal(control)
	if err != nil {
		t.Fatal(err)
	}
	if err := validateSourceFoundationRegister(t.Context(), repo, controlRaw, inputs); err != nil {
		t.Fatalf("complete retained source/report control: %v", err)
	}
	for _, scenario := range []string{"omitted known obligation", "wrong known source identity", "broadened assertion", "implemented source state", "omitted example", "invented adopter", "missing source facet", "changed baseline pin", "changed source digest", "null required relations", "invented completion", "omitted available proof", "fabricated input pin"} {
		t.Run(scenario, func(t *testing.T) {
			var candidate sourceFoundationRegister
			if err := json.Unmarshal(controlRaw, &candidate); err != nil {
				t.Fatal(err)
			}
			switch scenario {
			case "omitted known obligation":
				candidate.Known = []sourceFoundationKnownState{}
			case "wrong known source identity":
				candidate.Known[0].Identity.Key.ID = "source.b"
			case "broadened assertion":
				candidate.Requirements[0].Proofs[0].Assertion = "Every executor and mode is implemented"
			case "implemented source state":
				candidate.Requirements[0].SourceState = "implemented"
			case "omitted example":
				candidate.Examples = candidate.Examples[1:]
			case "invented adopter":
				candidate.Adopters = []sourceFoundationAdoption{{AtlasID: "transport.sync-contract.v1", Identity: candidate.Known[0].Identity, Relationship: "existing_shared_capability"}}
			case "missing source facet":
				candidate.Facets = candidate.Facets[1:]
			case "changed baseline pin":
				candidate.Baseline.SHA256 = sourceBytesHash([]byte("replacement baseline"))
			case "changed source digest":
				candidate.SourceSHA256 = sourceBytesHash([]byte("replacement source manifest"))
			case "null required relations":
				candidate.Adopters = nil
			case "invented completion":
				candidate.Checks.SelectedReuseRequirements = 1
				candidate.Checks.UnresolvedRequirements = 0
			case "omitted available proof":
				candidate.Proofs = candidate.Proofs[1:]
			case "fabricated input pin":
				candidate.Inputs = []sourceArtifactPin{{Path: "fiction.json", SHA256: sourceBytesHash(nil), Bytes: 0}}
			}
			// Complete JSON and self-consistent membership summaries remain
			// valid. Refusal must compare these claims with independent inputs.
			raw, err := json.Marshal(candidate)
			if err != nil {
				t.Fatal(err)
			}
			if err := validateSourceFoundationRegister(t.Context(), repo, raw, inputs); err == nil {
				t.Errorf("actual closed register reader accepted %s", scenario)
			}
		})
	}
}

func TestSourceFoundationRegisterRequiredZeroFields(t *testing.T) {
	repo, inputs := sourceFoundationRegisterFixture(t)
	control, err := buildSourceFoundationRegister(t.Context(), repo, inputs)
	if err != nil {
		t.Fatal(err)
	}
	controlRaw, err := json.Marshal(control)
	if err != nil {
		t.Fatal(err)
	}
	if err := validateSourceFoundationRegister(t.Context(), repo, controlRaw, inputs); err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"selected_reuse_requirements", "sentry_registration"} {
		for _, mutation := range []string{"absent", "null"} {
			t.Run(field+"/"+mutation, func(t *testing.T) {
				var wire map[string]any
				if err := json.Unmarshal(controlRaw, &wire); err != nil {
					t.Fatal(err)
				}
				object := wire["checks"].(map[string]any)
				if field == "sentry_registration" {
					object = wire["known_obligations"].([]any)[0].(map[string]any)
				}
				if mutation == "absent" {
					delete(object, field)
				} else {
					object[field] = nil
				}
				raw, err := json.Marshal(wire)
				if err != nil {
					t.Fatal(err)
				}
				if err := validateSourceFoundationRegister(t.Context(), repo, raw, inputs); err == nil {
					t.Errorf("actual register reader accepted %s required zero field %s", mutation, field)
				}
			})
		}
	}
}
