package app

import (
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
)

// The real database binary regression reaches this shared grant consumer in a
// fresh process. This bounded companion isolates its persisted policy binding.
func TestPostgresTransportPersistedConfirmation173(t *testing.T) {
	for _, mutation := range []string{"healthy", "missing_seal", "tampered_seal", "changed_target"} {
		t.Run(mutation, func(t *testing.T) {
			root := t.TempDir()
			if err := InitProject(root); err != nil {
				t.Fatal(err)
			}
			a, err := Open(root)
			if err != nil {
				t.Fatal(err)
			}
			confirmation := connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive}
			seal, err := a.approval.IssueWritePlanSeal(connectors.WritePlanSealRequest{
				PlanID: "rplan_cp16_confirmation", PlanHash: strings.Repeat("a", 64), Mode: reversePlanModePostgresManagedTarget,
				Connector: "postgres", Operation: "managed_incremental_upsert", CredentialRevision: "fixture-revision",
				ConfigurationDigest: "fixture-config", Batchable: true, Scope: connectors.WriteApprovalScopeProject, Confirmation: confirmation,
			})
			if err != nil {
				t.Fatal(err)
			}
			plan := ReversePlan{ID: seal.PlanID, PlanHash: seal.PlanHash, Mode: seal.Mode, Status: "planned",
				DestinationConnector: "postgres", Action: seal.Operation, PlanSeal: &seal, ConfirmationPolicy: confirmation,
				CreatedAt: seal.IssuedAt, ExpiresAt: seal.ExpiresAt}
			target := connectors.WriteApprovalTarget{Connector: "postgres", Operation: seal.Operation, Method: "POSTGRESQL",
				MutationClass: "incremental_upsert", TargetDigest: strings.Repeat("b", 64), CredentialRevision: "fixture-revision",
				ConfigurationDigest: "fixture-config", Batchable: true, Scope: connectors.WriteApprovalScopeProject, Confirmation: confirmation}
			if err := a.verifyPlanSealForTarget(plan, target); err != nil {
				t.Fatalf("healthy seal setup: %v", err)
			}
			switch mutation {
			case "missing_seal":
				plan.PlanSeal = nil
			case "tampered_seal":
				plan.PlanSeal.MAC = "tampered"
			case "changed_target":
				target.ConfigurationDigest = "different-config"
			}
			a.state.ReversePlans = append(a.state.ReversePlans, plan)
			if err := a.save(); err != nil {
				t.Fatal(err)
			}
			reopened, err := Open(root)
			if err != nil {
				t.Fatal(err)
			}
			stored, err := reopened.GetReversePlan(plan.ID)
			if err != nil {
				t.Fatal(err)
			}
			issued, err := reopened.persistDestructivePreview(stored, connectors.WritePreview{Digest: strings.Repeat("c", 64), ApprovalTarget: target})
			if mutation != "healthy" {
				if err == nil {
					t.Fatal("invalid seal/target minted a grant")
				}
				after, loadErr := Open(root)
				if loadErr != nil {
					t.Fatal(loadErr)
				}
				refused, loadErr := after.GetReversePlan(plan.ID)
				if loadErr != nil || refused.Status != "planned" || refused.ApprovalGrant != nil || refused.ApprovalTokenHash != "" {
					t.Fatal("refusal changed approval state")
				}
				return
			}
			if err != nil {
				t.Fatalf("valid closed transport preview grant: %v", err)
			}
			if issued.Status != "previewed" || issued.ApprovalToken == "" || issued.ApprovalGrant == nil {
				t.Fatal("preview did not issue bounded grant")
			}
			if err := reopened.validatePlanConfirmation(issued, connectors.WriteConfirmation{}); err == nil {
				t.Fatal("missing destructive confirmation accepted")
			}
			if err := reopened.validatePlanConfirmation(issued, confirmation); err != nil {
				t.Fatalf("correct confirmation refused: %v", err)
			}
			ordinary := ReversePlan{DestinationConnector: "postgres", Action: "managed_incremental_upsert"}
			if reopened.confirmationPolicyForPlan(ordinary).Kind != "" {
				t.Fatal("transport policy changed standalone action")
			}
		})
	}
}
