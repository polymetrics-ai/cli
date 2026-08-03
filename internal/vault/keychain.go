package vault

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/zalando/go-keyring"
)

// keychainService namespaces every Polymetrics vault key in the OS
// credential store, so uninstalling/reinstalling other tools never collides
// with it.
const keychainService = "polymetrics-cli-vault"

// vaultIDFile holds the vault's stable identity: a random hex string minted
// once, inside the vault directory, and never rewritten. The OS-keychain
// account is derived from it rather than from the directory's path so that
// moving, restoring, or symlinking the vault directory still resolves the
// same key instead of missing and minting a fresh one over live ciphertext.
const vaultIDFile = "id"

// KeychainProtector stores the vault's data key in the host OS's native
// credential store — macOS Keychain, Windows Credential Manager, or Linux
// Secret Service (D-Bus) — via github.com/zalando/go-keyring, instead of on
// disk. This closes vault v2 gap #1 (F3', "key stored in plaintext beside
// the ciphertext"): filesystem read access alone no longer yields every
// stored credential.
//
// Library choice: go-keyring wraps all three platforms behind one small
// interface, and its macOS backend shells out to the fixed system
// /usr/bin/security binary with fixed arguments (not a general shell
// escape — see AGENTS.md's ban on generic shell tools, which is about not
// exposing an arbitrary-command tool to the agent, not about a pinned
// system binary called with fixed flags) rather than linking cgo, keeping
// `pm`'s build simple and cross-compilable. Its Linux backend needs a
// reachable D-Bus Secret Service (gnome-keyring, kwallet, ...); on a
// headless host with none, LoadOrCreateKey returns an error and
// InitWithProtector's auto-selection falls back to filekeyProtector — that
// fallback is recorded via Vault.UsingFallbackKeyProtection(), never
// silent.
//
// The account name is derived from the vault's own persisted id (not from
// any filesystem path) so distinct project vaults never collide in one
// shared OS keychain, no filesystem path is stored as OS-keychain metadata,
// and relocating a vault directory never orphans its key.
type KeychainProtector struct{}

func (KeychainProtector) Name() string { return "keychain" }

func (KeychainProtector) LoadOrCreateKey(dir string) ([]byte, error) {
	account, err := keychainAccount(dir)
	if err != nil {
		return nil, err
	}
	existing, err := keyring.Get(keychainService, account)
	switch {
	case err == nil:
		return decodeKeychainKey(existing)
	case errors.Is(err, keyring.ErrNotFound):
		key, migrated, err := migrateLegacyPathKeyedEntry(dir, account)
		if err != nil {
			return nil, err
		}
		if migrated {
			return key, nil
		}
		fresh := make([]byte, 32)
		if _, err := rand.Read(fresh); err != nil {
			return nil, fmt.Errorf("generate vault key: %w", err)
		}
		if err := (KeychainProtector{}).StoreKey(dir, fresh); err != nil {
			return nil, err
		}
		return fresh, nil
	default:
		return nil, fmt.Errorf("OS keychain unavailable: %w", err)
	}
}

func (KeychainProtector) StoreKey(dir string, key []byte) error {
	account, err := keychainAccount(dir)
	if err != nil {
		return err
	}
	if err := keyring.Set(keychainService, account, base64.StdEncoding.EncodeToString(key)); err != nil {
		return fmt.Errorf("store vault key in OS keychain: %w", err)
	}
	return nil
}

// migrateLegacyPathKeyedEntry moves a key stored under the pre-existing
// path-derived account onto the id-derived one, so vaults created before the
// id file existed keep decrypting instead of being stranded behind a missing
// lookup. Reports whether such an entry was found.
func migrateLegacyPathKeyedEntry(dir, account string) ([]byte, bool, error) {
	legacy, err := legacyPathKeychainAccount(dir)
	if err != nil {
		return nil, false, err
	}
	if legacy == account {
		return nil, false, nil
	}
	stored, err := keyring.Get(keychainService, legacy)
	if errors.Is(err, keyring.ErrNotFound) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("OS keychain unavailable: %w", err)
	}
	key, err := decodeKeychainKey(stored)
	if err != nil {
		return nil, false, err
	}
	if err := keyring.Set(keychainService, account, stored); err != nil {
		return nil, false, fmt.Errorf("store vault key in OS keychain: %w", err)
	}
	if err := keyring.Delete(keychainService, legacy); err != nil && !errors.Is(err, keyring.ErrNotFound) {
		return nil, false, fmt.Errorf("remove legacy vault keychain entry: %w", err)
	}
	return key, true, nil
}

func decodeKeychainKey(stored string) ([]byte, error) {
	key, err := base64.StdEncoding.DecodeString(stored)
	if err != nil {
		return nil, fmt.Errorf("decode keychain-protected vault key: %w", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("keychain-protected vault key must be 32 bytes, got %d", len(key))
	}
	return key, nil
}

func keychainAccount(dir string) (string, error) {
	id, err := vaultIdentity(dir)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256([]byte("polymetrics-vault-id:" + id))
	return hex.EncodeToString(sum[:]), nil
}

// legacyPathKeychainAccount reproduces the original path-derived account so
// entries minted before vaultIDFile existed can still be found and migrated.
func legacyPathKeychainAccount(dir string) (string, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("resolve vault directory: %w", err)
	}
	sum := sha256.Sum256([]byte(abs))
	return hex.EncodeToString(sum[:]), nil
}

// vaultIdentity reads the vault's persisted id, minting and writing one on
// first use.
func vaultIdentity(dir string) (string, error) {
	path := filepath.Join(dir, vaultIDFile)
	raw, err := os.ReadFile(path)
	if err == nil {
		id := strings.TrimSpace(string(raw))
		if id == "" {
			return "", fmt.Errorf("vault id file %s is empty", path)
		}
		return id, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("read vault id: %w", err)
	}
	seed := make([]byte, 32)
	if _, err := rand.Read(seed); err != nil {
		return "", fmt.Errorf("generate vault id: %w", err)
	}
	id := hex.EncodeToString(seed)
	if err := os.WriteFile(path, []byte(id), 0o600); err != nil {
		return "", fmt.Errorf("write vault id: %w", err)
	}
	return id, nil
}
