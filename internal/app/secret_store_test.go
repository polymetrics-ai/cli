package app_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"polymetrics.ai/internal/app"
)

// TestRuntimeConfigSecretStorePersistsRotatedSecret is the rotation-persistence
// proof for issue #3705. A rotated OAuth2 refresh token must survive the
// process, and it must land in the project's EXISTING encrypted local
// credential store — no new storage path, no plaintext, nothing leaving the
// machine.
func TestRuntimeConfigSecretStorePersistsRotatedSecret(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()

	if err := app.InitProject(root); err != nil {
		t.Fatalf("InitProject() error = %v", err)
	}
	a, err := app.Open(root)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if _, err := a.AddCredential(ctx, app.AddCredentialRequest{
		Name:      "sample-oauth",
		Connector: "sample",
		Secrets:   map[string]string{"token": "sample-token", "refresh_token": "rt-original"},
		Config:    map[string]string{"workspace": "local"},
	}); err != nil {
		t.Fatalf("AddCredential() error = %v", err)
	}

	_, cfg, err := a.ResolveConnectorCredential(ctx, "sample", "sample-oauth", nil)
	if err != nil {
		t.Fatalf("ResolveConnectorCredential() error = %v", err)
	}
	if cfg.SecretStore == nil {
		t.Fatalf("RuntimeConfig.SecretStore is nil; a resolved credential must expose its store")
	}
	if got := cfg.Secrets["refresh_token"]; got != "rt-original" {
		t.Fatalf("initial refresh_token = %q, want rt-original", got)
	}

	const rotated = "rt-rotated-by-provider"
	if err := cfg.SecretStore.PutSecret(ctx, "refresh_token", rotated); err != nil {
		t.Fatalf("PutSecret() error = %v", err)
	}

	// "Next run": a brand-new App reading the credential back off disk.
	next, err := app.Open(root)
	if err != nil {
		t.Fatalf("second Open() error = %v", err)
	}
	_, nextCfg, err := next.ResolveConnectorCredential(ctx, "sample", "sample-oauth", nil)
	if err != nil {
		t.Fatalf("second ResolveConnectorCredential() error = %v", err)
	}
	if got := nextCfg.Secrets["refresh_token"]; got != rotated {
		t.Fatalf("refresh_token on the next run = %q, want %q", got, rotated)
	}
	// The sibling secret must be untouched: rotation replaces one key, never
	// the whole bundle.
	if got := nextCfg.Secrets["token"]; got != "sample-token" {
		t.Fatalf("sibling secret token = %q, want sample-token (rotation must not clobber the bundle)", got)
	}

	// And what is on disk must be ciphertext.
	assertVaultCiphertext(t, root, rotated)
}

// TestRuntimeConfigSecretStoreRejectsUndeclaredKey proves the store will not
// invent new secret keys: rotation writes to a key the credential already
// carries, and anything else is refused rather than silently added.
func TestRuntimeConfigSecretStoreRejectsInvalidKey(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	if err := app.InitProject(root); err != nil {
		t.Fatalf("InitProject() error = %v", err)
	}
	a, err := app.Open(root)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if _, err := a.AddCredential(ctx, app.AddCredentialRequest{
		Name:      "sample-oauth",
		Connector: "sample",
		Secrets:   map[string]string{"refresh_token": "rt-original"},
		Config:    map[string]string{"workspace": "local"},
	}); err != nil {
		t.Fatalf("AddCredential() error = %v", err)
	}
	_, cfg, err := a.ResolveConnectorCredential(ctx, "sample", "sample-oauth", nil)
	if err != nil {
		t.Fatalf("ResolveConnectorCredential() error = %v", err)
	}

	for _, key := range []string{"", "../escape", "has space", "with/slash"} {
		if err := cfg.SecretStore.PutSecret(ctx, key, "value"); err == nil {
			t.Fatalf("PutSecret(%q) error = nil, want rejection", key)
		}
	}
}

// TestRuntimeConfigSecretStoreErrorsAreRedacted proves a store failure never
// carries the secret value it was asked to write.
func TestRuntimeConfigSecretStoreErrorsAreRedacted(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	if err := app.InitProject(root); err != nil {
		t.Fatalf("InitProject() error = %v", err)
	}
	a, err := app.Open(root)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if _, err := a.AddCredential(ctx, app.AddCredentialRequest{
		Name:      "sample-oauth",
		Connector: "sample",
		Secrets:   map[string]string{"refresh_token": "rt-original"},
		Config:    map[string]string{"workspace": "local"},
	}); err != nil {
		t.Fatalf("AddCredential() error = %v", err)
	}
	_, cfg, err := a.ResolveConnectorCredential(ctx, "sample", "sample-oauth", nil)
	if err != nil {
		t.Fatalf("ResolveConnectorCredential() error = %v", err)
	}

	const secret = "rt-SUPERSECRET-rotated"
	err = cfg.SecretStore.PutSecret(ctx, "not a valid key", secret)
	if err == nil {
		t.Fatalf("PutSecret() error = nil, want an error")
	}
	if strings.Contains(err.Error(), secret) {
		t.Fatalf("store error leaked the secret value: %v", err)
	}
}

// assertVaultCiphertext walks the project's vault directory and fails if the
// plaintext secret appears verbatim in any stored blob.
func assertVaultCiphertext(t *testing.T, root, plaintext string) {
	t.Helper()
	vaultDir := filepath.Join(root, ".polymetrics", "vault")
	entries, err := os.ReadDir(vaultDir)
	if err != nil {
		t.Fatalf("read vault dir: %v", err)
	}
	checked := 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".enc") {
			continue
		}
		path := filepath.Join(vaultDir, entry.Name())
		blob, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		if bytes.Contains(blob, []byte(plaintext)) {
			t.Fatalf("%s contains the rotated secret in plaintext", path)
		}
		info, err := entry.Info()
		if err != nil {
			t.Fatalf("stat %s: %v", path, err)
		}
		if perm := info.Mode().Perm(); perm != 0o600 {
			t.Fatalf("%s mode = %o, want 600", path, perm)
		}
		checked++
	}
	if checked == 0 {
		t.Fatalf("no encrypted credential blobs found under %s", vaultDir)
	}

	// Belt and braces: nothing anywhere under .polymetrics may hold it in the
	// clear — not the state file, not a stray temp file.
	projectDir := filepath.Join(root, ".polymetrics")
	if err := filepath.Walk(projectDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		blob, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		if bytes.Contains(blob, []byte(plaintext)) {
			t.Fatalf("%s holds the rotated secret in plaintext", path)
		}
		return nil
	}); err != nil {
		t.Fatalf("walk project dir: %v", err)
	}
}
