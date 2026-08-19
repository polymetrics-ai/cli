package connectors_test

import (
	"bytes"
	"crypto/sha256"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors"
)

type callerProjectWriteApprovalEvidence struct{}

func (callerProjectWriteApprovalEvidence) ValidateProjectWrite(connectors.WriteApprovalTarget, string, time.Time) error {
	return nil
}

func (callerProjectWriteApprovalEvidence) AuthorizeProjectWrite(connectors.WriteApprovalTarget, string, time.Time) error {
	return nil
}

func TestProductionApprovalAuthorityIsNotPubliclyConstructible(t *testing.T) {
	// ParseDir intentionally scans every Go source file, including build-tagged
	// approval constructors that a package loader could omit for this platform.
	packages, err := parser.ParseDir(token.NewFileSet(), ".", nil, 0) //nolint:staticcheck
	if err != nil {
		t.Fatalf("ParseDir() error = %v", err)
	}
	for _, pkg := range packages {
		for _, file := range pkg.Files {
			ast.Inspect(file, func(node ast.Node) bool {
				declaration, ok := node.(*ast.FuncDecl)
				if ok && declaration.Recv == nil && declaration.Name.Name == "NewProcessWriteApprovalAuthority" {
					t.Fatal("production write approval authority remains publicly constructible")
				}
				return true
			})
		}
	}
}

func TestFixtureWriteApprovalGrantCannotBeVerifiedTwice(t *testing.T) {
	authority, err := connectors.NewFixtureWriteApprovalAuthority()
	if err != nil {
		t.Fatalf("NewFixtureWriteApprovalAuthority() error = %v", err)
	}
	target := connectors.WriteApprovalTarget{
		Connector: "acme", Operation: "delete_widget", Method: "DELETE", MutationClass: "delete",
		TargetDigest: strings.Repeat("b", 64), CredentialRevision: strings.Repeat("c", 64),
		ConfigurationDigest: strings.Repeat("d", 64), Scope: connectors.WriteApprovalScopeFixture,
		Confirmation: connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive},
	}
	grant, err := authority.IssueWriteGrant(connectors.WriteApprovalGrantRequest{
		PlanID: "rplan_fixture", PlanHash: strings.Repeat("a", 64), PreviewDigest: strings.Repeat("e", 64),
		ApprovalToken: "fixture-token", Target: target,
		Confirmation: connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive},
	})
	if err != nil {
		t.Fatalf("IssueWriteGrant() error = %v", err)
	}
	expected := approvalExpectation(grant, target)
	copiedAuthority := *authority
	if err := authority.ValidateWriteGrant(grant, expected); err != nil {
		t.Fatalf("ValidateWriteGrant(first) error = %v", err)
	}
	if err := authority.ValidateWriteGrant(grant, expected); err != nil {
		t.Fatalf("ValidateWriteGrant(second) error = %v", err)
	}
	if _, err := authority.VerifyWriteGrant(grant, expected); err != nil {
		t.Fatalf("VerifyWriteGrant(first) error = %v", err)
	}
	if _, err := copiedAuthority.VerifyWriteGrant(grant, expected); err == nil {
		t.Fatal("VerifyWriteGrant(replay) returned fresh fixture evidence")
	}
}

func TestProjectWriteApprovalEvidenceRejectsCallerImplementation(t *testing.T) {
	if _, err := connectors.BindProjectWriteApprovalEvidence(callerProjectWriteApprovalEvidence{}); err == nil {
		t.Fatal("BindProjectWriteApprovalEvidence() accepted caller implementation")
	}
}

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

func approvalExpectation(grant connectors.WriteApprovalGrant, target connectors.WriteApprovalTarget) connectors.WriteApprovalExpectation {
	return connectors.WriteApprovalExpectation{
		PlanID: grant.PlanID, PlanHash: grant.PlanHash, PreviewDigest: grant.PreviewDigest,
		ApprovalToken: "fixture-token", Target: target,
		Confirmation: grant.Confirmation,
	}
}
