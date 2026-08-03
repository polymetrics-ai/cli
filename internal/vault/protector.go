package vault

import (
	"crypto/rand"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// KeyProtector resolves the vault's 32-byte AES-256 data key and decides
// WHERE that key lives at rest. The Vault itself never knows or cares which
// protector produced its key.
type KeyProtector interface {
	// Name identifies the protector for the on-disk marker file
	// (vault/protector) and for Vault.KeyProtection() diagnostics. Stable
	// across releases — changing it orphans every vault already pinned to
	// the old name.
	Name() string
	// LoadOrCreateKey returns the existing 32-byte data key for dir,
	// generating and persisting a new random one on first use. An error
	// means this protector is unavailable (e.g. no OS keychain reachable),
	// not that the vault is corrupt.
	LoadOrCreateKey(dir string) ([]byte, error)
}

// keyStorer is implemented by protectors that support RotateKey
// (rotate.go): persisting a caller-supplied key rather than only
// generating one on first use.
type keyStorer interface {
	StoreKey(dir string, key []byte) error
}

// filekeyProtector is the pre-v2 behavior: a 32-byte key written in
// plaintext beside the ciphertext (vault/key, mode 0600). It is the
// explicit, warned fallback for headless hosts with no OS keychain —
// anyone with filesystem read access to the vault directory can decrypt
// every credential it holds, so InitWithProtector's auto-selection only
// reaches for it after KeychainProtector fails.
type filekeyProtector struct{}

func (filekeyProtector) Name() string { return "filekey" }

func (filekeyProtector) LoadOrCreateKey(dir string) ([]byte, error) {
	key, err := os.ReadFile(filekeyPath(dir))
	if errors.Is(err, os.ErrNotExist) {
		key := make([]byte, 32)
		if _, err := rand.Read(key); err != nil {
			return nil, fmt.Errorf("generate vault key: %w", err)
		}
		if err := (filekeyProtector{}).StoreKey(dir, key); err != nil {
			return nil, err
		}
		return key, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read vault key: %w", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("vault key must be 32 bytes, got %d", len(key))
	}
	return key, nil
}

func (filekeyProtector) StoreKey(dir string, key []byte) error {
	if err := os.WriteFile(filekeyPath(dir), key, 0o600); err != nil {
		return fmt.Errorf("write vault key: %w", err)
	}
	return nil
}

func filekeyPath(dir string) string { return filepath.Join(dir, "key") }

const protectorMarkerFile = "protector"

// Init opens or creates the vault under projectDir/vault, auto-selecting a
// KeyProtector: a vault that was already initialized (or a legacy pre-v2
// vault with a plaintext key file and no marker) reuses whatever protected
// it before; a brand-new vault tries the OS keychain first and falls back to
// the plaintext file key, recorded via UsingFallbackKeyProtection(), when no
// keychain is reachable (headless hosts, CI, sandboxes with no secret
// service).
func Init(projectDir string) (*Vault, error) {
	return InitWithProtector(projectDir, nil)
}

func Open(projectDir string) (*Vault, error) {
	return Init(projectDir)
}

// InitWithProtector is Init with an explicit KeyProtector, bypassing
// auto-selection. Passing a protector pins the vault to it unconditionally —
// tests and any caller that must not touch a real OS keychain should always
// use this with a protector they control.
func InitWithProtector(projectDir string, protector KeyProtector) (*Vault, error) {
	dir := filepath.Join(projectDir, "vault")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("create vault directory: %w", err)
	}

	if protector != nil {
		return newVaultWithProtector(dir, protector, false)
	}

	if pinned, ok, err := readProtectorMarker(dir); err != nil {
		return nil, err
	} else if ok {
		p, err := protectorByName(pinned)
		if err != nil {
			return nil, err
		}
		return newVaultWithProtector(dir, p, false)
	}
	return initNewVault(dir)
}

func initNewVault(dir string) (*Vault, error) {
	// Legacy pre-v2 vault: a plaintext key file with no marker yet. Preserve
	// it under filekeyProtector rather than reinterpreting it, so vaults
	// created before v2 keep decrypting exactly as before.
	if _, err := os.Stat(filekeyPath(dir)); err == nil {
		return newVaultWithProtector(dir, filekeyProtector{}, false)
	}

	// Brand-new vault: try the OS keychain first, fall back to the file key.
	return initWithProtectorChain(dir, KeychainProtector{}, filekeyProtector{})
}

// initWithProtectorChain tries each protector in order, using the first one
// that succeeds; every protector after the first is a recorded fallback
// (Vault.UsingFallbackKeyProtection()).
func initWithProtectorChain(dir string, chain ...KeyProtector) (*Vault, error) {
	var lastErr error
	for i, p := range chain {
		v, err := newVaultWithProtector(dir, p, i > 0)
		if err == nil {
			return v, nil
		}
		lastErr = err
	}
	return nil, fmt.Errorf("initialize vault: no key protector available: %w", lastErr)
}

func newVaultWithProtector(dir string, protector KeyProtector, fallback bool) (*Vault, error) {
	key, err := protector.LoadOrCreateKey(dir)
	if err != nil {
		return nil, fmt.Errorf("initialize vault key protection (%s): %w", protector.Name(), err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("vault key from protector %q must be 32 bytes, got %d", protector.Name(), len(key))
	}
	if err := writeProtectorMarker(dir, protector.Name()); err != nil {
		return nil, err
	}
	return &Vault{dir: dir, key: key, protector: protector, usingFallback: fallback}, nil
}

func protectorByName(name string) (KeyProtector, error) {
	switch name {
	case (filekeyProtector{}).Name():
		return filekeyProtector{}, nil
	case (KeychainProtector{}).Name():
		return KeychainProtector{}, nil
	default:
		return nil, fmt.Errorf("unknown vault key protector %q", name)
	}
}

func writeProtectorMarker(dir, name string) error {
	path := filepath.Join(dir, protectorMarkerFile)
	if existing, err := os.ReadFile(path); err == nil && string(existing) == name {
		return nil
	}
	if err := os.WriteFile(path, []byte(name), 0o600); err != nil {
		return fmt.Errorf("write vault protector marker: %w", err)
	}
	return nil
}

func readProtectorMarker(dir string) (string, bool, error) {
	b, err := os.ReadFile(filepath.Join(dir, protectorMarkerFile))
	if errors.Is(err, os.ErrNotExist) {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("read vault protector marker: %w", err)
	}
	return string(b), true, nil
}
