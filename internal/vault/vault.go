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
		candidate := make([]byte, 32)
		if _, err := rand.Read(candidate); err != nil {
			return nil, fmt.Errorf("generate vault key: %w", err)
		}
		created, err := writeDurableVaultKey(dir, keyPath, candidate)
		if err != nil {
			return nil, err
		}
		if created {
			key = candidate
		} else {
			key, err = os.ReadFile(keyPath)
			if err != nil {
				return nil, fmt.Errorf("read vault key: %w", err)
			}
		}
	} else if err != nil {
		return nil, fmt.Errorf("read vault key: %w", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("vault key must be 32 bytes, got %d", len(key))
	}
	return &Vault{dir: dir, key: key}, nil
}

func writeDurableVaultKey(dir, keyPath string, key []byte) (created bool, err error) {
	tmp, err := os.CreateTemp(dir, ".key.tmp-*")
	if err != nil {
		return false, fmt.Errorf("create temporary vault key: %w", err)
	}
	tmpPath := tmp.Name()
	defer func() {
		if tmp != nil {
			_ = tmp.Close()
		}
		if err != nil {
			_ = os.Remove(tmpPath)
		}
	}()
	if _, err := tmp.Write(key); err != nil {
		return false, fmt.Errorf("write temporary vault key: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		return false, fmt.Errorf("sync temporary vault key: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return false, fmt.Errorf("close temporary vault key: %w", err)
	}
	tmp = nil
	if err := os.Link(tmpPath, keyPath); err != nil {
		if !errors.Is(err, os.ErrExist) {
			return false, fmt.Errorf("link vault key: %w", err)
		}
		if err := os.Remove(tmpPath); err != nil && !errors.Is(err, os.ErrNotExist) {
			return false, fmt.Errorf("remove temporary vault key: %w", err)
		}
		if err := syncDirectory(dir); err != nil {
			return false, fmt.Errorf("sync vault directory: %w", err)
		}
		return false, nil
	}
	if err := os.Remove(tmpPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return false, fmt.Errorf("remove temporary vault key: %w", err)
	}
	if err := syncDirectory(dir); err != nil {
		return false, fmt.Errorf("sync vault directory: %w", err)
	}
	return true, nil
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
	ciphertext, err := v.encrypt(id, secret)
	if err != nil {
		return err
	}
	tmpPath, err := v.writeTemporaryCiphertext(id, ciphertext)
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(tmpPath) }()
	if err := os.Rename(tmpPath, v.path(id)); err != nil {
		return fmt.Errorf("replace encrypted credential %s: %w", id, err)
	}
	if err := v.syncDirectory(); err != nil {
		return fmt.Errorf("sync vault directory: %w", err)
	}
	return nil
}

// PutDurableIfAbsent writes an encrypted value exactly once and makes its
// directory entry durable before reporting whether this call created it.
func (v *Vault) PutDurableIfAbsent(ctx context.Context, id string, secret map[string]string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	if err := validateID(id); err != nil {
		return false, err
	}
	ciphertext, err := v.encrypt(id, secret)
	if err != nil {
		return false, err
	}
	tmpPath, err := v.writeTemporaryCiphertext(id, ciphertext)
	if err != nil {
		return false, err
	}
	defer func() { _ = os.Remove(tmpPath) }()
	if err := os.Link(tmpPath, v.path(id)); err != nil {
		if !errors.Is(err, os.ErrExist) {
			return false, fmt.Errorf("link encrypted credential %s: %w", id, err)
		}
		if err := os.Remove(tmpPath); err != nil && !errors.Is(err, os.ErrNotExist) {
			return false, fmt.Errorf("remove temporary encrypted credential %s: %w", id, err)
		}
		if err := v.syncDirectory(); err != nil {
			return false, fmt.Errorf("sync vault directory: %w", err)
		}
		return false, nil
	}
	if err := os.Remove(tmpPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return false, fmt.Errorf("remove temporary encrypted credential %s: %w", id, err)
	}
	if err := v.syncDirectory(); err != nil {
		return false, fmt.Errorf("sync vault directory: %w", err)
	}
	return true, nil
}

func (v *Vault) encrypt(id string, secret map[string]string) ([]byte, error) {
	plaintext, err := json.Marshal(secret)
	if err != nil {
		return nil, fmt.Errorf("marshal secret bundle: %w", err)
	}
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
	return gcm.Seal(nonce, nonce, plaintext, []byte(id)), nil
}

func (v *Vault) writeTemporaryCiphertext(id string, ciphertext []byte) (path string, err error) {
	file, err := os.CreateTemp(v.dir, "."+id+".tmp-*")
	if err != nil {
		return "", fmt.Errorf("create temporary encrypted credential %s: %w", id, err)
	}
	path = file.Name()
	defer func() {
		if file != nil {
			_ = file.Close()
		}
		if err != nil {
			_ = os.Remove(path)
		}
	}()
	if _, err := file.Write(ciphertext); err != nil {
		return "", fmt.Errorf("write temporary encrypted credential %s: %w", id, err)
	}
	if err := file.Sync(); err != nil {
		return "", fmt.Errorf("sync temporary encrypted credential %s: %w", id, err)
	}
	if err := file.Close(); err != nil {
		return "", fmt.Errorf("close temporary encrypted credential %s: %w", id, err)
	}
	file = nil
	return path, nil
}

func (v *Vault) syncDirectory() error {
	return syncDirectory(v.dir)
}

func syncDirectory(path string) error {
	dir, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = dir.Close() }()
	return dir.Sync()
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
