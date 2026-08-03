package vault

import (
	"crypto/rand"
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"

	"github.com/zalando/go-keyring"
)

// newVaultDirWithID creates a vault directory holding a fixed id file, so a
// test can build two directories at different paths that are nonetheless the
// same logical vault (the shape "the user moved/restored the project").
func newVaultDirWithID(t *testing.T, id string) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "vault")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("create vault dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, vaultIDFile), []byte(id), 0o600); err != nil {
		t.Fatalf("write vault id: %v", err)
	}
	return dir
}

// TestKeychainKeySurvivesVaultRelocation is the regression guard for silent,
// unrecoverable credential loss: the OS-keychain account must follow the
// vault's own persisted id, not its filesystem path, so relocating the
// directory resolves the SAME key instead of missing and minting a fresh one
// over live ciphertext.
func TestKeychainKeySurvivesVaultRelocation(t *testing.T) {
	keyring.MockInit()

	original := newVaultDirWithID(t, "fixed-vault-identity-for-test")
	first, err := KeychainProtector{}.LoadOrCreateKey(original)
	if err != nil {
		t.Fatalf("LoadOrCreateKey(original): %v", err)
	}

	relocated := newVaultDirWithID(t, "fixed-vault-identity-for-test")
	if relocated == original {
		t.Fatal("relocated vault dir must differ from the original path")
	}
	second, err := KeychainProtector{}.LoadOrCreateKey(relocated)
	if err != nil {
		t.Fatalf("LoadOrCreateKey(relocated): %v", err)
	}

	if string(first) != string(second) {
		t.Fatal("relocating the vault directory minted a different key; every existing credential would be orphaned")
	}
}

// TestKeychainMigratesLegacyPathDerivedEntry proves a vault whose key was
// minted under the original path-derived account is migrated onto the
// id-derived account on first access, rather than stranded behind a lookup
// that now misses.
func TestKeychainMigratesLegacyPathDerivedEntry(t *testing.T) {
	keyring.MockInit()

	dir := filepath.Join(t.TempDir(), "vault")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("create vault dir: %v", err)
	}

	legacyKey := make([]byte, 32)
	if _, err := rand.Read(legacyKey); err != nil {
		t.Fatalf("generate legacy key: %v", err)
	}
	legacyAccount, err := legacyPathKeychainAccount(dir)
	if err != nil {
		t.Fatalf("legacyPathKeychainAccount: %v", err)
	}
	if err := keyring.Set(keychainService, legacyAccount, base64.StdEncoding.EncodeToString(legacyKey)); err != nil {
		t.Fatalf("seed legacy keychain entry: %v", err)
	}

	got, err := KeychainProtector{}.LoadOrCreateKey(dir)
	if err != nil {
		t.Fatalf("LoadOrCreateKey: %v", err)
	}
	if string(got) != string(legacyKey) {
		t.Fatal("legacy path-derived key was not adopted; existing credentials would be orphaned")
	}

	account, err := keychainAccount(dir)
	if err != nil {
		t.Fatalf("keychainAccount: %v", err)
	}
	stored, err := keyring.Get(keychainService, account)
	if err != nil {
		t.Fatalf("migrated entry not readable under the id-derived account: %v", err)
	}
	if stored != base64.StdEncoding.EncodeToString(legacyKey) {
		t.Fatal("migrated entry holds a different key than the legacy one")
	}
	if _, err := keyring.Get(keychainService, legacyAccount); err == nil {
		t.Fatal("legacy path-derived entry still present; migration must remove it")
	}
}

// TestInitNeverReachesRealKeychainUnderGoTest proves vault.Init's
// auto-selection short-circuits to the file key while running as a test
// binary. Without this, every t.TempDir() project root any test in this repo
// creates would leave a fresh, never-deleted secret in the developer's own
// OS login keychain.
func TestInitNeverReachesRealKeychainUnderGoTest(t *testing.T) {
	if !testing.Testing() {
		t.Fatal("testing.Testing() is false inside a test binary")
	}

	root := t.TempDir()
	v, err := Init(root)
	if err != nil {
		t.Fatalf("Init: %v", err)
	}
	if got := v.KeyProtection(); got != (filekeyProtector{}).Name() {
		t.Fatalf("KeyProtection() = %q, want %q under go test", got, (filekeyProtector{}).Name())
	}

	marker, err := os.ReadFile(filepath.Join(root, "vault", protectorMarkerFile))
	if err != nil {
		t.Fatalf("read protector marker: %v", err)
	}
	if string(marker) != (filekeyProtector{}).Name() {
		t.Fatalf("protector marker = %q, want %q; auto-selection reached the real OS keychain", marker, (filekeyProtector{}).Name())
	}
}
