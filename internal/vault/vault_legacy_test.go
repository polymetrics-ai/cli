package vault

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"polymetrics.ai/internal/credential"
)

func TestGetRejectsLegacyEmptySecretValue(t *testing.T) {
	v, err := Init(filepath.Join(t.TempDir(), ".polymetrics"))
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	writeLegacySecret(t, v, "legacy_empty", map[string]string{"token": ""})

	_, err = v.Get(context.Background(), "legacy_empty")
	var empty *credential.EmptySecretError
	if !errors.As(err, &empty) {
		t.Fatalf("Get() error type = %T, want typed empty-secret classification", err)
	}
}

func TestGetAllowsLegacyOmittedOptionalSecret(t *testing.T) {
	const token = "legacy-optional-canary"
	v, err := Init(filepath.Join(t.TempDir(), ".polymetrics"))
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	writeLegacySecret(t, v, "legacy_omitted", map[string]string{"token": token})

	values, err := v.Get(context.Background(), "legacy_omitted")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	gotHash := sha256.Sum256([]byte(values["token"]))
	wantHash := sha256.Sum256([]byte(token))
	if len(values["token"]) != len(token) || gotHash != wantHash {
		t.Fatal("Get() did not preserve the non-empty legacy credential")
	}
	if _, present := values["optional"]; present {
		t.Fatal("Get() synthesized an omitted optional secret")
	}
}

func writeLegacySecret(t *testing.T, v *Vault, id string, values map[string]string) {
	t.Helper()
	plaintext, err := json.Marshal(values)
	if err != nil {
		t.Fatalf("marshal legacy secret: %v", err)
	}
	block, err := aes.NewCipher(v.key)
	if err != nil {
		t.Fatalf("create cipher: %v", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		t.Fatalf("create GCM: %v", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		t.Fatalf("read nonce: %v", err)
	}
	ciphertext := gcm.Seal(nonce, nonce, plaintext, []byte(id))
	if err := os.WriteFile(v.path(id), ciphertext, 0o600); err != nil {
		t.Fatalf("write legacy secret: %v", err)
	}
}
