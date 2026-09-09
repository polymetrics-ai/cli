package main

import (
	"bytes"
	"encoding/json"
	"testing"
)

// Replaces the historical147 Check/GET fixture, which did not prove body or
// auth and retained a stale bundle after deleting a map entry. Original reports
// and that false-positive history remain; ExactMechanism153 preserves its
// permanent negative. This control uses physical files and independent facts.
func TestSourceFoundationRequirementSharedAndLocalAspects(t *testing.T) {
	repo, _, inputs, _ := sourceFoundationCombinedFixture153(t, true)
	before, err := json.Marshal(inputs.assessments.universe.manifest)
	if err != nil {
		t.Fatal(err)
	}
	got, err := buildSourceFoundationRequirements(t.Context(), repo, inputs.assessments)
	if err != nil {
		t.Fatal(err)
	}
	expected := map[string]string{"shared-structured-body": "existing_shared_capability", "shared-static-header-auth": "existing_shared_capability", "local-command-artifact": "connector_local_configuration"}
	if len(got) != 3 {
		t.Fatal("three independent body/auth/local aspects required")
	}
	for _, r := range got {
		if r.Status != expected[r.ID] || r.Identity.Key.ID != "source.a" || r.Identity.Lane != "direct_write" || len(r.Proofs) != 1 || len(r.MechanismFits) != 1 || r.MechanismFits[0].DeclarationState != "canonical_intended" || r.SourceState != "mapped_unproven" {
			t.Fatal("independent requirement meaning or source state changed")
		}
		if r.ID == "shared-static-header-auth" && r.Proofs[0].ID != "runtime.direct-execution.v1.static-api-key-header" {
			t.Fatal("auth proof borrowed")
		}
		if r.ID != "shared-static-header-auth" && r.Proofs[0].ID != "runtime.direct-execution.v1.structured-rest-body" {
			t.Fatal("body/local proof borrowed")
		}
		delete(expected, r.ID)
	}
	if len(expected) != 0 {
		t.Fatal("an independent aspect disappeared")
	}
	after, err := json.Marshal(inputs.assessments.universe.manifest)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("requirements promoted or rewrote current execution")
	}
}
