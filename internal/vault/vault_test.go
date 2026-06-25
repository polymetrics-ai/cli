package vault_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"polymetrics.ai/internal/vault"
)

func TestVaultEncryptsSecretsAtRest(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()

	v, err := vault.Init(filepath.Join(root, ".polymetrics"))
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	secret := map[string]string{"token": "super-secret-token"}
	if err := v.Put(ctx, "cred_test", secret); err != nil {
		t.Fatalf("Put() error = %v", err)
	}

	got, err := v.Get(ctx, "cred_test")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got["token"] != "super-secret-token" {
		t.Fatalf("decrypted token = %q", got["token"])
	}

	entries, err := os.ReadDir(filepath.Join(root, ".polymetrics", "vault"))
	if err != nil {
		t.Fatalf("ReadDir() error = %v", err)
	}
	var combined strings.Builder
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		b, err := os.ReadFile(filepath.Join(root, ".polymetrics", "vault", entry.Name()))
		if err != nil {
			t.Fatalf("ReadFile(%s) error = %v", entry.Name(), err)
		}
		combined.Write(b)
	}
	if strings.Contains(combined.String(), "super-secret-token") {
		t.Fatalf("vault files contain plaintext secret")
	}
}
