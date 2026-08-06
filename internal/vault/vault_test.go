package vault_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
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

func TestVaultInitConcurrentlySharesOneKey(t *testing.T) {
	root := filepath.Join(t.TempDir(), ".polymetrics")
	start := make(chan struct{})
	results := make(chan *vault.Vault, 2)
	errors := make(chan error, 2)
	var group sync.WaitGroup
	for range 2 {
		group.Add(1)
		go func() {
			defer group.Done()
			<-start
			instance, err := vault.Init(root)
			if err != nil {
				errors <- err
				return
			}
			results <- instance
		}()
	}
	close(start)
	group.Wait()
	close(results)
	close(errors)
	for err := range errors {
		t.Fatal(err)
	}
	instances := make([]*vault.Vault, 0, 2)
	for instance := range results {
		instances = append(instances, instance)
	}
	if len(instances) != 2 {
		t.Fatalf("vault instances = %d, want 2", len(instances))
	}
	ctx := context.Background()
	if err := instances[0].Put(ctx, "shared", map[string]string{"token": "fixture"}); err != nil {
		t.Fatal(err)
	}
	secret, err := instances[1].Get(ctx, "shared")
	if err != nil {
		t.Fatal(err)
	}
	if secret["token"] != "fixture" {
		t.Fatalf("shared key decrypted token = %q", secret["token"])
	}
}
