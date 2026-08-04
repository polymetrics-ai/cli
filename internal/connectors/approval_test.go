package connectors_test

import (
	"bytes"
	"crypto/sha256"
	"strings"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors"
)

func TestWriteApprovalGrantAuthenticatesTargetExpiryAndCredentialRevision(t *testing.T) {
	authority, err := connectors.NewWriteApprovalAuthority(bytes.Repeat([]byte{0x71}, sha256.Size))
	if err != nil {
		t.Fatalf("NewWriteApprovalAuthority() error = %v", err)
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

	now := time.Now().UTC()
	target := connectors.WriteApprovalTarget{
		Connector: "acme", Operation: "delete_widget", Method: "DELETE", MutationClass: "delete",
		TargetDigest: strings.Repeat("b", 64), CredentialRevision: revisionOne,
	}
	grant, err := authority.IssueWriteGrant(connectors.WriteApprovalGrantRequest{
		PlanID: "rplan_fixture", PlanHash: strings.Repeat("a", 64), PreviewDigest: strings.Repeat("c", 64),
		ApprovalToken: "fixture-token", Target: target, IssuedAt: now, ExpiresAt: now.Add(time.Hour),
		Confirmation: connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive},
	})
	if err != nil {
		t.Fatalf("IssueWriteGrant() error = %v", err)
	}

	tampered := grant
	tampered.Target.TargetDigest = strings.Repeat("d", 64)
	if _, err := authority.VerifyWriteGrant(tampered, approvalExpectation(grant, target, now)); err == nil {
		t.Fatal("VerifyWriteGrant() accepted a tampered target")
	}
	if _, err := authority.VerifyWriteGrant(grant, approvalExpectation(grant, target, grant.ExpiresAt)); err == nil {
		t.Fatal("VerifyWriteGrant() accepted an expired grant")
	}
	changedTarget := target
	changedTarget.CredentialRevision = revisionTwo
	if _, err := authority.VerifyWriteGrant(grant, approvalExpectation(grant, changedTarget, now)); err == nil {
		t.Fatal("VerifyWriteGrant() accepted a changed credential revision")
	}
}

func approvalExpectation(grant connectors.WriteApprovalGrant, target connectors.WriteApprovalTarget, now time.Time) connectors.WriteApprovalExpectation {
	return connectors.WriteApprovalExpectation{
		PlanID: grant.PlanID, PlanHash: grant.PlanHash, PreviewDigest: grant.PreviewDigest,
		ApprovalToken: "fixture-token", Target: target, ExpiresAt: grant.ExpiresAt,
		Confirmation: grant.Confirmation, Now: now,
	}
}
