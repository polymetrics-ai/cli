package connectors_test

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"strings"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/vault"
)

func TestWriteApprovalGrantAuthenticatesTargetExpiryAndCredentialRevision(t *testing.T) {
	authority, err := connectors.NewUntrustedWriteApprovalAuthority(bytes.Repeat([]byte{0x71}, sha256.Size))
	if err != nil {
		t.Fatalf("NewUntrustedWriteApprovalAuthority() error = %v", err)
	}
	revisionOne, err := authority.CredentialRevision("cred_fixture", map[string]string{"token": "secret-one"})
	if err != nil {
		t.Fatalf("CredentialRevision(one) error = %v", err)
	}
	revisionTwo, err := authority.CredentialRevision("cred_fixture", map[string]string{"token": "secret-two"})
	if err != nil {
		t.Fatalf("CredentialRevision(two) error = %v", err)
	}
	if revisionOne == revisionTwo {
		t.Fatal("credential revision did not change with credential material")
	}
	if strings.Contains(revisionOne, "secret-one") || strings.Contains(revisionTwo, "secret-two") {
		t.Fatal("credential revision exposed secret material")
	}

	target := connectors.WriteApprovalTarget{
		Connector: "acme", Operation: "delete_widget", Method: "DELETE", MutationClass: "delete",
		TargetDigest: strings.Repeat("b", 64), CredentialRevision: revisionOne,
		ConfigurationDigest: strings.Repeat("e", 64), Scope: connectors.WriteApprovalScopeProject,
		Confirmation: connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive},
	}
	grant, err := authority.IssueWriteGrant(connectors.WriteApprovalGrantRequest{
		PlanID: "rplan_fixture", PlanHash: strings.Repeat("a", 64), PreviewDigest: strings.Repeat("c", 64),
		ApprovalToken: "fixture-token", Target: target,
		Confirmation: connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive},
	})
	if err != nil {
		t.Fatalf("IssueWriteGrant() error = %v", err)
	}

	tampered := grant
	tampered.Target.TargetDigest = strings.Repeat("d", 64)
	if _, err := authority.VerifyWriteGrant(tampered, approvalExpectation(grant, target)); err == nil {
		t.Fatal("VerifyWriteGrant() accepted a tampered target")
	}
	expired := grant
	expired.ExpiresAt = expired.IssuedAt
	if _, err := authority.VerifyWriteGrant(expired, approvalExpectation(grant, target)); err == nil {
		t.Fatal("VerifyWriteGrant() accepted an expired grant")
	}
	changedTarget := target
	changedTarget.CredentialRevision = revisionTwo
	if _, err := authority.VerifyWriteGrant(grant, approvalExpectation(grant, changedTarget)); err == nil {
		t.Fatal("VerifyWriteGrant() accepted a changed credential revision")
	}
	changedTarget = target
	changedTarget.ConfigurationDigest = strings.Repeat("f", 64)
	if _, err := authority.VerifyWriteGrant(grant, approvalExpectation(grant, changedTarget)); err == nil {
		t.Fatal("VerifyWriteGrant() accepted changed configuration semantics")
	}
	changedTarget = target
	changedTarget.Batchable = !target.Batchable
	if _, err := authority.VerifyWriteGrant(grant, approvalExpectation(grant, changedTarget)); err == nil {
		t.Fatal("VerifyWriteGrant() accepted changed batchability")
	}
}

func TestProcessWriteApprovalRequiresSealedPlanAndPersistentConsumption(t *testing.T) {
	v, err := vault.Init(t.TempDir())
	if err != nil {
		t.Fatalf("vault.Init() error = %v", err)
	}
	authority, err := connectors.NewProcessWriteApprovalAuthority(v.WriteApprovalRoot())
	if err != nil {
		t.Fatalf("NewProcessWriteApprovalAuthority() error = %v", err)
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
	if _, err := authority.VerifyWriteGrant(grant, expected); err != nil {
		t.Fatalf("VerifyWriteGrant(first) error = %v", err)
	}
	if _, err := authority.VerifyWriteGrant(grant, expected); !errors.Is(err, vault.ErrWriteApprovalConsumed) {
		t.Fatalf("VerifyWriteGrant(replay) error = %v, want consumed marker rejection", err)
	}
}

func approvalExpectation(grant connectors.WriteApprovalGrant, target connectors.WriteApprovalTarget) connectors.WriteApprovalExpectation {
	return connectors.WriteApprovalExpectation{
		PlanID: grant.PlanID, PlanHash: grant.PlanHash, PreviewDigest: grant.PreviewDigest,
		ApprovalToken: "fixture-token", Target: target,
		Confirmation: grant.Confirmation,
	}
}
