package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestSourceFoundationMechanismClaims155(t *testing.T) {
	for _, fault := range []string{"body_response_only", "auth_response_only", "wrong_binding_hash", "wrong_generation", "wrong_canonical_id", "wrong_canonical_pointer", "borrow_check", "swap_body_auth", "unavailable_auth"} {
		t.Run(fault, func(t *testing.T) {
			repo, baseline, assessment := sourceFoundationCombinedRetained155(t, true)
			sourceFoundationAdmissionCommand155(t, repo, baseline)
			raw, err := os.ReadFile(filepath.Join(repo, sourceLaneManifestPath))
			if err != nil {
				t.Fatal(err)
			}
			var source sourceLaneManifest
			if json.Unmarshal(raw, &source) != nil {
				t.Fatal("source report")
			}
			body := &assessment.Assessments[0].Requirements[0]
			auth := &assessment.Assessments[0].Requirements[1]
			switch fault {
			case "body_response_only":
				body.SourceRefs = []sourceFactRef{source.SourceOperations[0].Facts.Refs["responses"]}
			case "auth_response_only":
				auth.SourceRefs = []sourceFactRef{source.SourceOperations[0].Facts.Refs["responses"]}
			case "wrong_binding_hash":
				body.FitBindings[0].ArtifactSHA256 = sourceBytesHash([]byte("wrong selected artifact"))
			case "wrong_generation":
				body.FitBindings[0].Generation = sourceBytesHash([]byte("another generation"))
			case "wrong_canonical_id":
				body.FitBindings[0].CanonicalID = "operation:widgets.other"
			case "wrong_canonical_pointer":
				body.FitBindings[0].CanonicalPointer = "/operations/1/operation"
			case "borrow_check":
				raw, err = os.ReadFile(filepath.Join(repo, sourceFoundationProofPath))
				if err != nil {
					t.Fatal(err)
				}
				var proofs sourceFoundationProofDocument
				if json.Unmarshal(raw, &proofs) != nil {
					t.Fatal("proof document")
				}
				for _, p := range proofs.Records {
					if p.ID == "runtime.direct-execution.v1.reject-undeclared-check-status" {
						body.ProofIDs = []string{p.ID}
						body.Statement = p.Assertion.Statement
						body.AtlasLookup.Candidates = []sourceFoundationLookupCandidate{{AtlasID: p.AtlasID, Contract: p.Contract, Disposition: "existing_shared_capability", Rationale: "test attempted same-file Check borrowing"}}
					}
				}
			case "swap_body_auth":
				body.ProofIDs = append([]string{}, auth.ProofIDs...)
				body.Statement = auth.Statement
				body.AtlasLookup = auth.AtlasLookup
			case "unavailable_auth":
				// Final150 §4.2 requires explicit unresolved accounting; an
				// authored resolved claim with missing proof is a hard refusal.
				auth.Assessment = "unresolved"
				if err := os.Remove(filepath.Join(repo, "data/connector-canon/proof-receipts/foundation/static-auth-proof-153-01/receipt.json")); err != nil {
					t.Fatal(err)
				}
			}
			writeSourceFoundationAssessmentFixture(t, repo, assessment)
			var out, diag bytes.Buffer
			code := runSourceDemandsPolicy(t.Context(), []string{"source-demands", "--repo", repo}, &out, &diag, baseline)
			if fault == "unavailable_auth" {
				var got sourceFoundationRegister
				if code != 0 || json.Unmarshal(out.Bytes(), &got) != nil || len(got.Requirements) != 3 || len(got.Adopters) != 2 {
					t.Fatalf("isolated optional auth proof erased independent demand/body/local fit: %s", diag.String())
				}
				for _, r := range got.Requirements {
					if r.ID == "shared-static-header-auth" && (r.Status != "unresolved" || len(r.MechanismFits) != 0 || len(r.Proofs) != 1 || r.Proofs[0].Status != "proof_unavailable") {
						t.Fatal("missing auth proof fabricated current shared fit")
					}
				}
			} else if code == 0 || out.Len() != 0 {
				t.Fatal("unrelated source/selector/assertion supplied exact mechanism authority")
			}
		})
	}
}

func TestSourceFoundationStaticAuthSelectors155(t *testing.T) {
	// Establish the same complete retained source/command/proof control once;
	// each sibling reaches real canonical/source admission with changed auth.
	control, baseline, _ := sourceFoundationCombinedRetained155(t, false)
	sourceFoundationAdmissionCommand155(t, control, baseline)
	for _, profile := range []string{"auth_query", "auth_header", "auth_scoped", "auth_conjunctive", "auth_oauth", "auth_prefix", "auth_first_none", "auth_conditional", "auth_template", "auth_nonsecret"} {
		t.Run(profile, func(t *testing.T) {
			repo, baseline, _ := sourceFoundationCombinedRetainedProfile155(t, false, profile)
			var out, diag bytes.Buffer
			if code := runSourceDemandsPolicy(t.Context(), []string{"source-demands", "--repo", repo}, &out, &diag, baseline); code == 0 || out.Len() != 0 {
				t.Fatal("uncovered static auth selector promoted to reviewed header mechanism")
			}
		})
	}
}

func TestSourceFoundationBodyMedia155(t *testing.T) {
	control, baseline, _ := sourceFoundationCombinedRetainedProfile155(t, false, "body_flat")
	sourceFoundationAdmissionCommand155(t, control, baseline)
	repo, baseline, _ := sourceFoundationCombinedRetainedProfile155(t, false, "body_form")
	var out, diag bytes.Buffer
	if code := runSourceDemandsPolicy(t.Context(), []string{"source-demands", "--repo", repo}, &out, &diag, baseline); code == 0 || out.Len() != 0 {
		t.Fatal("form encoding borrowed the structured JSON body proof")
	}
}
