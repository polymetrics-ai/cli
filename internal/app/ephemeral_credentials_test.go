package app_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"polymetrics.ai/internal/app"
)

func TestCertificationEphemeralCredentialsResolveWithoutVaultOrProfile(t *testing.T) {
	root := t.TempDir()
	session, err := app.BeginCertificationEphemeralCredentials(root, app.EphemeralCredential{
		Name:      "cert-source",
		Connector: "sample",
		Config:    map[string]string{"region": "test"},
		Secrets:   map[string]string{"token": "cert-canary-ephemeral-only"},
	})
	if err != nil {
		t.Fatalf("begin ephemeral session: %v", err)
	}
	t.Cleanup(session.Close)

	if err := app.InitProject(root); err != nil {
		t.Fatalf("init ephemeral project: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".polymetrics", "vault")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("ephemeral init created vault material: stat error = %v, want not exist", err)
	}

	instance, err := app.Open(root)
	if err != nil {
		t.Fatalf("open ephemeral project: %v", err)
	}
	_, runtime, err := instance.ResolveConnectorCredential(context.Background(), "sample", "cert-source", map[string]string{"region": "override"})
	if err != nil {
		t.Fatalf("resolve ephemeral credential: %v", err)
	}
	if got := runtime.Secrets["token"]; got != "cert-canary-ephemeral-only" {
		t.Fatalf("resolved token = %q, want in-memory canary", got)
	}
	if got := runtime.Config["region"]; got != "override" {
		t.Fatalf("resolved config override = %q, want override", got)
	}
	if got := instance.ListCredentials(); len(got) != 0 {
		t.Fatalf("persisted credential metadata = %#v, want none", got)
	}
	if _, err := os.Stat(filepath.Join(root, ".polymetrics", "vault", "key")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("ephemeral run created vault key: stat error = %v, want not exist", err)
	}
}
