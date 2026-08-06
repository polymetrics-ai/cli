package app

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestCredentialCoordination_IdentityDerivationDoesNotReadVault(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	if err := InitProject(root); err != nil {
		t.Fatalf("InitProject() error = %v", err)
	}
	instance, err := Open(root)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	credential, err := instance.AddCredential(ctx, AddCredentialRequest{
		Name:      "sample-isolated",
		Connector: "sample",
	})
	if err != nil {
		t.Fatalf("AddCredential() error = %v", err)
	}

	if err := os.Remove(filepath.Join(root, ".polymetrics", "vault", credential.ID+".enc")); err != nil {
		t.Fatalf("remove temporary encrypted credential: %v", err)
	}
	identity, err := instance.coordinationIdentityForCredential(credential)
	if err != nil {
		t.Fatalf("coordinationIdentityForCredential() error = %v", err)
	}
	if identity.AuthCohortKey() == "" {
		t.Fatal("identity derivation returned no opaque auth cohort key")
	}
}
