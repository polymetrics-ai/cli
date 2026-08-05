// Package vault is Polymetrics's encrypted local credential store: AES-256-GCM
// at rest, per-project directory, zero centralized custody. Vault v2 (dual-
// mechanism connector foundations, P0) extends the original flat
// map[string]string store with four things the browser-auth package and the
// -web connectors need: a pluggable key protector including an
// OS-keychain-backed one (protector.go, opt-in — see Init),
// opaque blob storage alongside the credential map (namespace.go), a
// connector/profile/kind namespace for per-account isolation (namespace.go),
// and key rotation (rotate.go). The original flat Put/Get/Delete/Redact API
// is unchanged and still backs `pm credentials add`.
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
)

// Vault is an encrypted local credential store rooted at
// <projectDir>/vault. The AES-256 data key is resolved by a KeyProtector
// (protector.go) — a plaintext on-disk key by default, or the OS keychain
// when a caller explicitly selects it via InitWithProtector.
type Vault struct {
	dir           string
	key           []byte
	protector     KeyProtector
	usingFallback bool
}

// KeyProtection names the KeyProtector that resolved this vault's data key
// ("keychain" or "filekey").
func (v *Vault) KeyProtection() string { return v.protector.Name() }

// UsingFallbackKeyProtection reports whether this vault fell back to a later
// protector in an explicit protector chain because an earlier one was
// unavailable when the vault was first created. Callers (e.g. `pm auth
// login`) should surface this as a one-time warning rather than silently
// degrading security. It is always false for the default auto-selected path,
// which uses the plaintext file key deliberately rather than as a fallback —
// see Init.
func (v *Vault) UsingFallbackKeyProtection() bool { return v.usingFallback }

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
	return v.writeFlat(id, plaintext)
}

func (v *Vault) Get(ctx context.Context, id string) (map[string]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := validateID(id); err != nil {
		return nil, err
	}
	plaintext, err := v.readFlat(id)
	if err != nil {
		return nil, err
	}
	var out map[string]string
	if err := json.Unmarshal(plaintext, &out); err != nil {
		return nil, fmt.Errorf("decode secret bundle: %w", err)
	}
	if out == nil {
		out = map[string]string{}
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

// Redact returns a copy of in with every value replaced by a fixed marker,
// for safe display (logs, `--json` output, error messages) of a secret
// bundle's key set without its values.
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

func (v *Vault) writeFlat(id string, plaintext []byte) error {
	sealed, err := v.seal([]byte(id), plaintext)
	if err != nil {
		return err
	}
	if err := os.WriteFile(v.path(id), sealed, 0o600); err != nil {
		return fmt.Errorf("write encrypted credential %s: %w", id, err)
	}
	return nil
}

func (v *Vault) readFlat(id string) ([]byte, error) {
	ciphertext, err := os.ReadFile(v.path(id))
	if err != nil {
		return nil, fmt.Errorf("read encrypted credential %s: %w", id, err)
	}
	plaintext, err := v.open([]byte(id), ciphertext)
	if err != nil {
		return nil, fmt.Errorf("decrypt credential %s: %w", id, err)
	}
	return plaintext, nil
}

func (v *Vault) listFlatIDs() ([]string, error) {
	entries, err := os.ReadDir(v.dir)
	if err != nil {
		return nil, fmt.Errorf("list vault: %w", err)
	}
	var ids []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if name, ok := strings.CutSuffix(e.Name(), ".enc"); ok {
			ids = append(ids, name)
		}
	}
	return ids, nil
}

// seal AES-256-GCM-encrypts plaintext, binding aad (additional authenticated
// data — a credential id or a namespace path) so ciphertext cannot be
// relocated to a different id/namespace and still decrypt.
func (v *Vault) seal(aad, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(v.key)
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create gcm: %w", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("generate nonce: %w", err)
	}
	return gcm.Seal(nonce, nonce, plaintext, aad), nil
}

func (v *Vault) open(aad, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(v.key)
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create gcm: %w", err)
	}
	if len(ciphertext) < gcm.NonceSize() {
		return nil, errors.New("encrypted payload is truncated")
	}
	nonce := ciphertext[:gcm.NonceSize()]
	payload := ciphertext[gcm.NonceSize():]
	return gcm.Open(nil, nonce, payload, aad)
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
