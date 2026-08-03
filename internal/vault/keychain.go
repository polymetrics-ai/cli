package vault

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/zalando/go-keyring"
)

// keychainService namespaces every Polymetrics vault key in the OS
// credential store, so uninstalling/reinstalling other tools never collides
// with it.
const keychainService = "polymetrics-cli-vault"

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
// The account name is derived from the vault directory's absolute path (not
// the path itself) so distinct project vaults never collide in one shared
// OS keychain and no filesystem path is stored as OS-keychain metadata.
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
		key, decErr := base64.StdEncoding.DecodeString(existing)
		if decErr != nil {
			return nil, fmt.Errorf("decode keychain-protected vault key: %w", decErr)
		}
		if len(key) != 32 {
			return nil, fmt.Errorf("keychain-protected vault key must be 32 bytes, got %d", len(key))
		}
		return key, nil
	case errors.Is(err, keyring.ErrNotFound):
		key := make([]byte, 32)
		if _, err := rand.Read(key); err != nil {
			return nil, fmt.Errorf("generate vault key: %w", err)
		}
		if err := (KeychainProtector{}).StoreKey(dir, key); err != nil {
			return nil, err
		}
		return key, nil
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

func keychainAccount(dir string) (string, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("resolve vault directory: %w", err)
	}
	sum := sha256.Sum256([]byte(abs))
	return hex.EncodeToString(sum[:]), nil
}
