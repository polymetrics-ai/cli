package vault

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	pmlogging "polymetrics.ai/internal/logging"
)

type Vault struct {
	dir string
	key []byte
}

func Init(projectDir string) (*Vault, error) {
	dir := filepath.Join(projectDir, "vault")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("create vault directory: %w", err)
	}
	keyPath := filepath.Join(dir, "key")
	key, err := os.ReadFile(keyPath)
	if errors.Is(err, os.ErrNotExist) {
		key = make([]byte, 32)
		if _, err := rand.Read(key); err != nil {
			return nil, fmt.Errorf("generate vault key: %w", err)
		}
		if err := os.WriteFile(keyPath, key, 0o600); err != nil {
			return nil, fmt.Errorf("write vault key: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("read vault key: %w", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("vault key must be 32 bytes, got %d", len(key))
	}
	return &Vault{dir: dir, key: key}, nil
}

func Open(projectDir string) (*Vault, error) {
	return Init(projectDir)
}

func (v *Vault) Put(ctx context.Context, id string, secret map[string]string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := validateID(id); err != nil {
		return err
	}
	plaintext, err := json.Marshal(secret)
	if err != nil {
		return fmt.Errorf("marshal secret bundle: %w", err)
	}
	block, err := aes.NewCipher(v.key)
	if err != nil {
		return fmt.Errorf("create cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("create gcm: %w", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return fmt.Errorf("generate nonce: %w", err)
	}
	ciphertext := gcm.Seal(nonce, nonce, plaintext, []byte(id))
	if err := os.WriteFile(v.path(id), ciphertext, 0o600); err != nil {
		return fmt.Errorf("write encrypted credential %s: %w", id, err)
	}
	return nil
}

func (v *Vault) Get(ctx context.Context, id string) (map[string]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := validateID(id); err != nil {
		return nil, err
	}
	ciphertext, err := os.ReadFile(v.path(id))
	if err != nil {
		return nil, fmt.Errorf("read encrypted credential %s: %w", id, err)
	}
	block, err := aes.NewCipher(v.key)
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create gcm: %w", err)
	}
	if len(ciphertext) < gcm.NonceSize() {
		return nil, errors.New("encrypted credential is truncated")
	}
	nonce := ciphertext[:gcm.NonceSize()]
	payload := ciphertext[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, payload, []byte(id))
	if err != nil {
		return nil, fmt.Errorf("decrypt credential %s: %w", id, err)
	}
	var out map[string]string
	if err := json.Unmarshal(plaintext, &out); err != nil {
		return nil, fmt.Errorf("decode secret bundle: %w", err)
	}
	if out == nil {
		out = map[string]string{}
	}
	for _, value := range out {
		pmlogging.RegisterValueFromContext(ctx, value)
	}
	return out, nil
}

func (v *Vault) Delete(ctx context.Context, id string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := validateID(id); err != nil {
		return err
	}
	err := os.Remove(v.path(id))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func Redact(in map[string]string) map[string]string {
	out := make(map[string]string, len(in))
	for k := range in {
		out[k] = "***"
	}
	return out
}

func (v *Vault) path(id string) string {
	return filepath.Join(v.dir, id+".enc")
}

func validateID(id string) error {
	if id == "" {
		return errors.New("credential id is required")
	}
	if strings.Contains(id, "..") || strings.ContainsAny(id, `/\`) {
		return fmt.Errorf("invalid credential id %q", id)
	}
	for _, r := range id {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '_' || r == '-':
		default:
			return fmt.Errorf("invalid credential id %q", id)
		}
	}
	return nil
}
