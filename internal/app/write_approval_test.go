package app

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors"
)

func TestProjectWriteApprovalRequiresSealedPlanAndPersistentConsumption(t *testing.T) {
	root := t.TempDir()
	if err := InitProject(root); err != nil {
		t.Fatalf("InitProject() error = %v", err)
	}
	projectDir := filepath.Join(root, ".polymetrics")
	authority, err := newProjectWriteApprovalAuthority(projectDir)
	if err != nil {
		t.Fatalf("newProjectWriteApprovalAuthority() error = %v", err)
	}
	revision, err := authority.CredentialRevision("cred_fixture", map[string]string{"token": "secret"})
	if err != nil {
		t.Fatalf("CredentialRevision() error = %v", err)
	}
	configuration, err := authority.ConfigurationDigest("cred_fixture", map[string]string{"base_url": "https://api.example.test"})
	if err != nil {
		t.Fatalf("ConfigurationDigest() error = %v", err)
	}
	confirmation := connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive}
	seal, err := authority.IssueWritePlanSeal(connectors.WritePlanSealRequest{
		PlanID: "rplan_fixture", PlanHash: strings.Repeat("a", 64), Connector: "acme", Operation: "delete_widget",
		CredentialRevision: revision, ConfigurationDigest: configuration, Batchable: false,
		Scope: connectors.WriteApprovalScopeProject, Confirmation: confirmation,
	})
	if err != nil {
		t.Fatalf("IssueWritePlanSeal() error = %v", err)
	}
	target := connectors.WriteApprovalTarget{
		Connector: "acme", Operation: "delete_widget", Method: "DELETE", MutationClass: "delete",
		TargetDigest: strings.Repeat("b", 64), CredentialRevision: revision, ConfigurationDigest: configuration,
		Batchable: false, Scope: connectors.WriteApprovalScopeProject, Confirmation: confirmation,
	}
	grant, err := authority.IssueWriteGrant(connectors.WriteApprovalGrantRequest{
		PlanID: seal.PlanID, PlanHash: seal.PlanHash, PlanSeal: &seal,
		PreviewDigest: strings.Repeat("c", 64), ApprovalToken: "fixture-token", Target: target, Confirmation: confirmation,
	})
	if err != nil {
		t.Fatalf("IssueWriteGrant() error = %v", err)
	}
	if grant.ExpiresAt.After(time.Now().UTC().Add(16 * time.Minute)) {
		t.Fatalf("grant expiry = %s, want trusted short-lived deadline", grant.ExpiresAt)
	}
	expected := connectors.WriteApprovalExpectation{
		PlanID: grant.PlanID, PlanHash: grant.PlanHash, PreviewDigest: grant.PreviewDigest,
		ApprovalToken: "fixture-token", Target: target, Confirmation: confirmation,
	}
	if err := authority.ValidateWriteGrant(grant, expected, &seal); err != nil {
		t.Fatalf("ValidateWriteGrant(first) error = %v", err)
	}
	if err := authority.ValidateWriteGrant(grant, expected, &seal); err != nil {
		t.Fatalf("ValidateWriteGrant(second) error = %v", err)
	}
	evidence, err := authority.VerifyWriteGrant(grant, expected, &seal)
	if err != nil {
		t.Fatalf("VerifyWriteGrant(first) error = %v", err)
	}
	if err := evidence.Authorize(target, grant.PreviewDigest, time.Now().UTC()); err != nil {
		t.Fatalf("Authorize() error = %v", err)
	}
	if _, err := authority.VerifyWriteGrant(grant, expected, &seal); !errors.Is(err, connectors.ErrWriteApprovalConsumed) {
		t.Fatalf("VerifyWriteGrant(replay) error = %v, want consumed marker rejection", err)
	}
	reopened, err := newProjectWriteApprovalAuthority(projectDir)
	if err != nil {
		t.Fatalf("newProjectWriteApprovalAuthority(reopen) error = %v", err)
	}
	if _, err := reopened.IssueWriteGrant(connectors.WriteApprovalGrantRequest{
		PlanID: seal.PlanID, PlanHash: seal.PlanHash, PlanSeal: &seal,
		PreviewDigest: grant.PreviewDigest, ApprovalToken: "another-token", Target: target, Confirmation: confirmation,
	}); !errors.Is(err, connectors.ErrWriteApprovalConsumed) {
		t.Fatalf("IssueWriteGrant(consumed plan) error = %v, want persistent rejection", err)
	}
}
