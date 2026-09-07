package main

import "testing"

func TestSourceFoundationRequirementSharedAndLocalAspects(t *testing.T) {
	repo, universe, document, proofs := sourceFoundationRequirementsFixture(t)
	observed, err := observeSourceFoundationAssessments(t.Context(), repo, sourceFoundationAssessmentsPath, universe)
	if err != nil {
		t.Fatal(err)
	}
	key, facts, annotation := sourceBindingREST118A(t, "[]", "{}")
	operation := annotation.IntendedBindings[0]
	path := "internal/connectors/defs/acme/cli_surface.json"
	command := sourceLaneTargetRef{Kind: "command", Connector: "acme", ID: "widgets get", Lane: "direct_read",
		Artifact: path, Pointer: "/commands/0", ArtifactSHA256: sourceBytesHash(facts.bindings.Artifacts[path]),
		CanonicalID: operation.CanonicalID, CanonicalPointer: "/operations/0/commands/0", Generation: operation.Generation}
	annotation.IntendedBindings = append(annotation.IntendedBindings, command)
	// Both bindings first pass the real canonical/source reconciliation.
	complete := sourceBindingOutcome099F(t, key, facts, annotation, 2)
	if len(complete.References) != 2 {
		t.Fatal("positive operation/command binding frontier not reached")
	}
	delete(facts.bindings.Artifacts, path)
	cell := sourceBindingOutcome099F(t, key, facts, annotation, 1)
	if !sourceLaneTargetRefEqual(cell.References[0], operation) {
		t.Fatal("missing CLI declaration erased independent operation binding")
	}
	assessment := document.Assessments[0]
	assessment.Key, assessment.Lane = key, cell.Lane
	shared := assessment.Requirements[0]
	shared.ID, shared.Statement = "shared-status-contract", proofs.Records[1].Assertion.Statement
	shared.Assessment, shared.ProofIDs = "existing_shared_capability", []string{proofs.Records[1].ID}
	shared.SourceRefs = []sourceFactRef{facts.Refs["responses"]}
	shared.AtlasLookup.Candidates = []sourceFoundationLookupCandidate{{AtlasID: proofs.Records[1].AtlasID,
		Contract: proofs.Records[1].Contract, Disposition: "existing_shared_capability",
		Rationale: "Retain only the reviewed conditional status assertion and exact direct-execution declaration."}}
	shared.FitBindings, shared.AffectedArtifacts = []sourceLaneTargetRef{operation}, []string{operation.Artifact}
	local := shared
	local.ID, local.Assessment = "local-command-declaration", "connector_local_configuration"
	local.FitBindings, local.AffectedArtifacts = []sourceLaneTargetRef{command}, []string{path}
	assessment.Requirements = []sourceFoundationRequirement{shared, local}
	observed.authored = []sourceFoundationCellAssessment{assessment}
	observed.universe.manifest.SourceOperations = []sourceLaneManifestRow{{Source: retainedSourceOperation{Key: key}, Facts: facts, Lanes: []sourceLaneCell{cell}}}
	got, err := buildSourceFoundationRequirements(t.Context(), repo, observed)
	if err != nil || len(got) != 2 {
		t.Fatalf("independent shared/local requirements: %v", err)
	}
	if got[0].ID != local.ID || got[0].Status != "connector_local_configuration" || len(got[0].AffectedArtifacts) != 1 || got[0].AffectedArtifacts[0] != path ||
		got[1].ID != shared.ID || got[1].Status != "existing_shared_capability" || got[0].Identity != got[1].Identity ||
		got[0].Proofs[0].Assertion != shared.Statement || got[1].Proofs[0].Assertion != shared.Statement {
		t.Fatalf("single cell collapsed shared/local outcomes or widened assertion: %+v", got)
	}
	if cell.State != "mapped_unproven" || len(cell.ProofRefs) != 0 {
		t.Fatal("authoring requirements promoted source execution")
	}
}
