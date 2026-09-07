package main

import "testing"

func TestSourceFoundationExactMechanism153(t *testing.T) {
	for _, local := range []bool{false, true} {
		name := "shared"
		if local {
			name = "local"
		}
		t.Run(name, func(t *testing.T) {
			repo, universe, document, proofs := sourceFoundationRequirementsFixture(t)
			observed, err := observeSourceFoundationAssessments(t.Context(), repo, sourceFoundationAssessmentsPath, universe)
			if err != nil {
				t.Fatal(err)
			}
			proofObservations, err := readSourceFoundationProofObservations(t.Context(), repo)
			if err != nil || len(proofObservations) != 2 || proofObservations[0].status != "current" {
				t.Fatalf("real current proof control: %v", err)
			}
			key, facts, annotation := sourceBindingREST118A(t, "[]", "{}")
			check := facts.bindings.Bundles[key.Connector].HTTP.Check
			if check == nil || check.Path != "/check" || len(check.SuccessStatuses) != 0 {
				t.Fatal("fixture no longer isolates default /check from selected /widgets operation")
			}
			want := 1
			if local {
				delete(facts.bindings.Artifacts, annotation.IntendedBindings[0].Artifact)
				want = 0
			}
			cell := sourceBindingOutcome099F(t, key, facts, annotation, want)
			assessment := document.Assessments[0]
			assessment.Key, assessment.Lane = key, cell.Lane
			req := &assessment.Requirements[0]
			req.Statement = proofs.Records[1].Assertion.Statement
			req.ProofIDs = []string{proofs.Records[1].ID}
			req.AtlasLookup.Candidates[0].AtlasID = proofs.Records[1].AtlasID
			req.AtlasLookup.Candidates[0].Contract = proofs.Records[1].Contract
			req.Assessment = "existing_shared_capability"
			if local {
				req.Assessment = "connector_local_configuration"
			}
			req.SourceRefs = []sourceFactRef{facts.Refs["responses"]}
			req.FitBindings = annotation.IntendedBindings
			req.AffectedArtifacts = []string{annotation.IntendedBindings[0].Artifact}
			universe.manifest.SourceOperations = []sourceLaneManifestRow{{Source: retainedSourceOperation{Key: key}, Facts: facts, Lanes: []sourceLaneCell{cell}}}
			observed.universe = universe
			observed.authored = []sourceFoundationCellAssessment{assessment}
			got, err := buildSourceFoundationRequirements(t.Context(), repo, observed)
			if err == nil && len(got) == 1 && got[0].Status == req.Assessment {
				t.Fatalf("resolved %s for operation selector %s while HTTP.Check is unrelated /check with default statuses, using Check-only proof %s", got[0].Status, req.FitBindings[0].CanonicalID, proofs.Records[1].ID)
			}
		})
	}
}
