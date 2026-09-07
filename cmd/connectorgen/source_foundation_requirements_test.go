package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func sourceFoundationRequirementsFixture(t *testing.T) (string, sourceFoundationUniverse, sourceFoundationAssessmentDocument, sourceFoundationProofDocument) {
	t.Helper()
	repo, proofs := sourceFoundationProofObservationFixture(t)
	writeSourceFoundationObservationFixture(t, repo, proofs)
	// Both fixtures retain these literal source rows; the actual source reader
	// still verifies the expected file digest, not a synthetic observed object.
	_, cohort := sourceInventoryFixture(t, []string{"source.b", "source.a"}, 2)
	universe, err := buildSourceFoundationUniverse(t.Context(), repo, cohort, nil)
	if err != nil {
		t.Fatal(err)
	}
	document := sourceFoundationAssessmentDocumentFixture(t, universe)
	document.Assessments[0].Requirements[0].ProofIDs = []string{proofs.Records[0].ID}
	writeSourceFoundationAssessmentFixture(t, repo, document)
	return repo, universe, document, proofs
}

func TestSourceFoundationRequirementObservation(t *testing.T) {
	repo, universe, document, proofs := sourceFoundationRequirementsFixture(t)
	observed, err := observeSourceFoundationAssessments(t.Context(), repo, sourceFoundationAssessmentsPath, universe)
	if err != nil {
		t.Fatal(err)
	}
	got, err := buildSourceFoundationRequirements(t.Context(), repo, observed)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Status != "unresolved" || got[0].Statement != document.Assessments[0].Requirements[0].Statement ||
		len(got[0].Proofs) != 1 || got[0].Proofs[0].Status != "current" ||
		got[0].Proofs[0].Assertion != proofs.Records[0].Assertion.Statement ||
		!reflect.DeepEqual(got[0].Proofs[0].Limitations, proofs.Records[0].Limitations) ||
		!reflect.DeepEqual(got[0].SourceRefs, document.Assessments[0].Requirements[0].SourceRefs) {
		t.Fatal("real source requirement/current narrow assertion separation lost")
	}
}

func TestSourceFoundationRequirementUnrelatedAtlasSelector(t *testing.T) {
	repo, universe, document, proofs := sourceFoundationRequirementsFixture(t)
	observed, err := observeSourceFoundationAssessments(t.Context(), repo, sourceFoundationAssessmentsPath, universe)
	if err != nil {
		t.Fatal(err)
	}
	key, facts, annotation := sourceBindingFixture099F(t, "stream", nil)
	cell := sourceBindingOutcome099F(t, key, facts, annotation, 1)
	assessment := document.Assessments[0]
	assessment.Key, assessment.Lane = key, cell.Lane
	requirement := &assessment.Requirements[0]
	requirement.Statement = proofs.Records[0].Assertion.Statement
	requirement.Assessment = "existing_shared_capability"
	requirement.SourceRefs = []sourceFactRef{facts.Refs["responses"]}
	requirement.FitBindings = annotation.IntendedBindings
	requirement.AffectedArtifacts = []string{annotation.IntendedBindings[0].Artifact}
	observed.universe.manifest.SourceOperations = []sourceLaneManifestRow{{Source: retainedSourceOperation{Key: key}, Facts: facts, Lanes: []sourceLaneCell{cell}}}
	observed.authored = []sourceFoundationCellAssessment{assessment}
	// The canonical stream is real, and its source binding is accepted, but
	// transport.sync-contract.v1 declares sync_transport.json selectors. It
	// cannot acquire a stream adopter merely by borrowing the vocabulary test.
	if got, err := buildSourceFoundationRequirements(t.Context(), repo, observed); err == nil {
		t.Errorf("actual requirement consumer accepted unrelated Atlas selector: %+v", got)
	}
	observed.authored[0].Requirements[0].Assessment = "unresolved"
	if got, err := buildSourceFoundationRequirements(t.Context(), repo, observed); err != nil || len(got) != 1 || got[0].Status != "unresolved" {
		t.Fatalf("narrow proof must remain observable without fabricated reuse: %v", err)
	}
}

func TestSourceFoundationRequirementClassificationAdmission(t *testing.T) {
	for _, scenario := range []string{"vocabulary borrowed for executor", "local configuration without binding", "absence inferred from no proof", "provider limitation without clause", "no demand against applicable source", "no demand without exclusion", "proof from unrelated contract"} {
		t.Run(scenario, func(t *testing.T) {
			repo, universe, document, proofs := sourceFoundationRequirementsFixture(t)
			control, err := observeSourceFoundationAssessments(t.Context(), repo, sourceFoundationAssessmentsPath, universe)
			if err != nil {
				t.Fatal(err)
			}
			if got, err := buildSourceFoundationRequirements(t.Context(), repo, control); err != nil || len(got) != 1 || got[0].Status != "unresolved" {
				t.Fatalf("real unresolved requirement/current vocabulary control: %v", err)
			}
			requirement := &document.Assessments[0].Requirements[0]
			switch scenario {
			case "vocabulary borrowed for executor":
				requirement.Assessment = "existing_shared_capability"
				requirement.Statement = "A receiver executes this source and durably acknowledges every event"
			case "local configuration without binding":
				requirement.Assessment = "connector_local_configuration"
				requirement.Statement = proofs.Records[0].Assertion.Statement
			case "absence inferred from no proof":
				requirement.Assessment = "absent_shared_foundation"
				requirement.ProofIDs = []string{}
			case "provider limitation without clause":
				requirement.Assessment = "provider_limitation"
			case "no demand against applicable source":
				document.Assessments[0].Lane = "direct_read"
				requirement.Assessment = "no_demand_for_this_cell"
				ref := universe.manifest.SourceOperations[0].Facts.Refs["method"]
				requirement.SourceExclusion = &ref
			case "no demand without exclusion":
				requirement.Assessment = "no_demand_for_this_cell"
			case "proof from unrelated contract":
				requirement.ProofIDs = []string{proofs.Records[1].ID}
			}
			writeSourceFoundationAssessmentFixture(t, repo, document)
			observed, err := observeSourceFoundationAssessments(t.Context(), repo, sourceFoundationAssessmentsPath, universe)
			if err != nil {
				t.Fatalf("negative did not reach classification consumer: %v", err)
			}
			if got, err := buildSourceFoundationRequirements(t.Context(), repo, observed); err == nil {
				t.Errorf("actual requirement consumer accepted %s: %+v", scenario, got)
			}
		})
	}
}

func TestSourceFoundationRequirementSourceExclusion(t *testing.T) {
	repo, universe, document, _ := sourceFoundationRequirementsFixture(t)
	requirement := &document.Assessments[0].Requirements[0]
	requirement.Assessment = "no_demand_for_this_cell"
	requirement.ProofIDs = []string{}
	ref := universe.manifest.SourceOperations[0].Facts.Refs["source_operation"]
	requirement.SourceExclusion = &ref
	requirement.AtlasLookup.Candidates[0].Disposition = "no_demand_for_this_cell"
	writeSourceFoundationAssessmentFixture(t, repo, document)
	observed, err := observeSourceFoundationAssessments(t.Context(), repo, sourceFoundationAssessmentsPath, universe)
	if err != nil {
		t.Fatal(err)
	}
	got, err := buildSourceFoundationRequirements(t.Context(), repo, observed)
	if err != nil || len(got) != 1 || got[0].Status != "no_demand_for_this_cell" || len(got[0].Proofs) != 0 ||
		len(got[0].SourceRefs) != 1 || got[0].SourceRefs[0] != ref {
		t.Fatalf("independent actual source exclusion not retained: %v", err)
	}
}

func TestSourceFoundationRequirementExactBindingFit(t *testing.T) {
	for _, local := range []bool{false, true} {
		name := "shared existing binding"
		if local {
			name = "local missing artifact"
		}
		t.Run(name, func(t *testing.T) {
			repo, universe, document, proofs := sourceFoundationRequirementsFixture(t)
			observed, err := observeSourceFoundationAssessments(t.Context(), repo, sourceFoundationAssessmentsPath, universe)
			if err != nil {
				t.Fatal(err)
			}
			// Exercise the existing canonical producer, engine bundle loader,
			// source normalization and exact binding consumer. No hand-authored
			// accepted target is substituted into the classifier's result.
			key, facts, annotation := sourceBindingREST118A(t, "[]", "{}")
			wantReferences := 1
			if local {
				delete(facts.bindings.Artifacts, annotation.IntendedBindings[0].Artifact)
				wantReferences = 0
			}
			cell := sourceBindingOutcome099F(t, key, facts, annotation, wantReferences)
			assessment := document.Assessments[0]
			assessment.Key, assessment.Lane = key, cell.Lane
			requirement := &assessment.Requirements[0]
			requirement.Statement = proofs.Records[1].Assertion.Statement
			requirement.ProofIDs = []string{proofs.Records[1].ID}
			requirement.AtlasLookup.Candidates[0].AtlasID = proofs.Records[1].AtlasID
			requirement.AtlasLookup.Candidates[0].Contract = proofs.Records[1].Contract
			requirement.Assessment = "existing_shared_capability"
			if local {
				requirement.Assessment = "connector_local_configuration"
			}
			requirement.SourceRefs = []sourceFactRef{facts.Refs["responses"]}
			requirement.FitBindings = annotation.IntendedBindings
			requirement.AffectedArtifacts = []string{annotation.IntendedBindings[0].Artifact}
			universe.manifest.SourceOperations = []sourceLaneManifestRow{{Source: retainedSourceOperation{Key: key}, Facts: facts, Lanes: []sourceLaneCell{cell}}}
			observed.universe = universe
			observed.authored = []sourceFoundationCellAssessment{assessment}
			got, err := buildSourceFoundationRequirements(t.Context(), repo, observed)
			if err != nil || len(got) != 1 || got[0].Status != requirement.Assessment || len(got[0].Proofs) != 1 ||
				got[0].Proofs[0].Assertion != requirement.Statement || len(got[0].Proofs[0].Limitations) == 0 {
				t.Fatalf("exact narrow assertion/binding result: %v", err)
			}
			// The result retains only the quoted conditional check-status assertion.
			// It does not certify this read operation or promote its source lane.
			if cell.State != "mapped_unproven" || len(cell.ProofRefs) != 0 {
				t.Fatal("requirement result changed lane proof authority")
			}
		})
	}
}

func TestSourceFoundationRequirementDecisionAdmission(t *testing.T) {
	for _, scenario := range []string{"invented owner", "invented approval", "duplicate decision", "unconditional exposure", "exposure without condition"} {
		t.Run(scenario, func(t *testing.T) {
			repo, universe, document, _ := sourceFoundationRequirementsFixture(t)
			requirement := &document.Assessments[0].Requirements[0]
			requirement.DecisionRefs = []sourceFoundationDecision{{ID: "cli-batch1-vercel-inbound-sync-decision-r1", State: "pending"},
				{ID: "cli-plan-webhook-receiver-foundation-r1-decision-webhook-exposure-product-boundary", State: "conditional", Condition: "If receiver approval requires an exposed endpoint"}}
			writeSourceFoundationAssessmentFixture(t, repo, document)
			observed, err := observeSourceFoundationAssessments(t.Context(), repo, sourceFoundationAssessmentsPath, universe)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := buildSourceFoundationRequirements(t.Context(), repo, observed); err != nil {
				t.Fatalf("existing pending/conditional owners control: %v", err)
			}
			switch scenario {
			case "invented owner":
				requirement.DecisionRefs[0].ID = "new-approved-receiver"
			case "invented approval":
				requirement.DecisionRefs[0].State = "approved"
			case "duplicate decision":
				requirement.DecisionRefs = append(requirement.DecisionRefs, requirement.DecisionRefs[0])
			case "unconditional exposure":
				requirement.DecisionRefs[1].State = "pending"
			case "exposure without condition":
				requirement.DecisionRefs[1].Condition = ""
			}
			writeSourceFoundationAssessmentFixture(t, repo, document)
			observed, err = observeSourceFoundationAssessments(t.Context(), repo, sourceFoundationAssessmentsPath, universe)
			if err != nil {
				t.Fatalf("did not reach requirement decision consumer: %v", err)
			}
			if _, err := buildSourceFoundationRequirements(t.Context(), repo, observed); err == nil {
				t.Errorf("accepted %s", scenario)
			}
		})
	}
}

func TestSourceFoundationRequirementProviderClauseAndAspects(t *testing.T) {
	repo, _, _, _ := sourceFoundationRequirementsFixture(t)
	_, cohort := sourceInventoryFixture(t, []string{"source.b", "source.a"}, 2)
	name := filepath.Join(repo, "source.json")
	raw, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	var retained map[string]any
	if err := json.Unmarshal(raw, &retained); err != nil {
		t.Fatal(err)
	}
	const clause = "This fixture endpoint cannot deliver event notifications."
	for _, value := range retained["rest"].(map[string]any)["operations"].([]any) {
		value.(map[string]any)["source_operation"].(map[string]any)["description"] = clause
	}
	raw, err = json.Marshal(retained)
	if err != nil {
		t.Fatal(err)
	}
	proofWrite(t, repo, "source.json", raw)
	cohort.Inventories[0].SHA256 = sourceBytesHash(raw)
	universe, err := buildSourceFoundationUniverse(t.Context(), repo, cohort, nil)
	if err != nil {
		t.Fatal(err)
	}
	document := sourceFoundationAssessmentDocumentFixture(t, universe)
	facts := universe.manifest.SourceOperations[0].Facts
	ref := sourceBindingCitation099F(t, facts, facts.Refs["source_operation"].Pointer+"/description")
	ref.DocumentID = facts.Refs["source_operation"].DocumentID
	limitation := document.Assessments[0].Requirements[0]
	limitation.ID, limitation.Statement, limitation.Assessment = "provider-events", clause, "provider_limitation"
	limitation.SourceRefs, limitation.ProviderClause = []sourceFactRef{ref}, &ref
	limitation.AtlasLookup.Candidates = append([]sourceFoundationLookupCandidate{}, limitation.AtlasLookup.Candidates...)
	limitation.AtlasLookup.Candidates[0].Disposition = "provider_limitation"
	limitation.AtlasLookup.Candidates[0].Rationale = "The retained endpoint clause restricts delivery regardless of transport mode vocabulary."
	document.Assessments[0].Requirements = append(document.Assessments[0].Requirements, limitation)
	writeSourceFoundationAssessmentFixture(t, repo, document)
	observed, err := observeSourceFoundationAssessments(t.Context(), repo, sourceFoundationAssessmentsPath, universe)
	if err != nil {
		t.Fatal(err)
	}
	got, err := buildSourceFoundationRequirements(t.Context(), repo, observed)
	if err != nil || len(got) != 2 {
		t.Fatalf("two separately classified requirements: %v", err)
	}
	if got[0].ID != "provider-events" || got[0].Status != "provider_limitation" || got[0].Statement != clause ||
		got[0].ProviderClause == nil || *got[0].ProviderClause != ref || got[1].ID != "transport-fit" || got[1].Status != "unresolved" ||
		got[0].Identity != got[1].Identity {
		t.Fatalf("lost exact clause or collapsed distinct aspects: %+v", got)
	}
	// Equal clause bytes on a different operation are not this source's evidence.
	otherFacts := universe.manifest.SourceOperations[1].Facts
	other := sourceBindingCitation099F(t, otherFacts, otherFacts.Refs["source_operation"].Pointer+"/description")
	other.DocumentID = otherFacts.Refs["source_operation"].DocumentID
	document.Assessments[0].Requirements[1].SourceRefs = []sourceFactRef{other}
	document.Assessments[0].Requirements[1].ProviderClause = &other
	writeSourceFoundationAssessmentFixture(t, repo, document)
	if _, err := observeSourceFoundationAssessments(t.Context(), repo, sourceFoundationAssessmentsPath, universe); err == nil {
		t.Fatal("copied sibling provider clause accepted")
	}
}
