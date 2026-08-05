package vault_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"polymetrics.ai/internal/vault"
)

// fakeProtector is an in-memory KeyProtector so vault v2 tests never touch a
// real OS keychain — deterministic across every CI platform and safe in a
// sandboxed, non-interactive environment where a real keychain prompt could
// hang.
type fakeProtector struct {
	name string
	mu   sync.Mutex
	keys map[string][]byte
	fail bool
}

func newFakeProtector(name string) *fakeProtector {
	return &fakeProtector{name: name, keys: map[string][]byte{}}
}

func (p *fakeProtector) Name() string { return p.name }

func (p *fakeProtector) LoadOrCreateKey(dir string) ([]byte, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.fail {
		return nil, errors.New("fake protector unavailable")
	}
	if key, ok := p.keys[dir]; ok {
		return key, nil
	}
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i + 1)
	}
	p.keys[dir] = key
	return key, nil
}

func (p *fakeProtector) StoreKey(dir string, key []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.fail {
		return errors.New("fake protector unavailable")
	}
	cp := make([]byte, len(key))
	copy(cp, key)
	p.keys[dir] = cp
	return nil
}

func testVault(t *testing.T) *vault.Vault {
	t.Helper()
	v, err := vault.InitWithProtector(t.TempDir(), newFakeProtector("fake"))
	if err != nil {
		t.Fatalf("InitWithProtector() error = %v", err)
	}
	return v
}

func TestVaultNamespacedCredentialRoundTrip(t *testing.T) {
	ctx := context.Background()
	v := testVault(t)
	ns := vault.Namespace{Connector: "linkedin-web", Profile: "default", Kind: "session"}

	if err := v.PutNamespaced(ctx, ns, map[string]string{"li_at": "secret-cookie"}); err != nil {
		t.Fatalf("PutNamespaced() error = %v", err)
	}
	got, err := v.GetNamespaced(ctx, ns)
	if err != nil {
		t.Fatalf("GetNamespaced() error = %v", err)
	}
	if got["li_at"] != "secret-cookie" {
		t.Fatalf("li_at = %q", got["li_at"])
	}

	if err := v.DeleteNamespaced(ctx, ns); err != nil {
		t.Fatalf("DeleteNamespaced() error = %v", err)
	}
	if _, err := v.GetNamespaced(ctx, ns); err == nil {
		t.Fatalf("GetNamespaced() after delete: want error, got nil")
	}
}

func TestVaultBlobRoundTrip(t *testing.T) {
	ctx := context.Background()
	v := testVault(t)
	ns := vault.Namespace{Connector: "whatsapp-web", Profile: "default", Kind: "device_store"}
	blob := []byte{0x00, 0x01, 0xff, 0xfe, 'h', 'i'}

	if err := v.PutBlob(ctx, ns, blob); err != nil {
		t.Fatalf("PutBlob() error = %v", err)
	}
	got, err := v.GetBlob(ctx, ns)
	if err != nil {
		t.Fatalf("GetBlob() error = %v", err)
	}
	if string(got) != string(blob) {
		t.Fatalf("GetBlob() = %v, want %v", got, blob)
	}

	if err := v.DeleteBlob(ctx, ns); err != nil {
		t.Fatalf("DeleteBlob() error = %v", err)
	}
	if _, err := v.GetBlob(ctx, ns); err == nil {
		t.Fatalf("GetBlob() after delete: want error, got nil")
	}
}

func TestVaultBlobAtRestIsEncrypted(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	v, err := vault.InitWithProtector(root, newFakeProtector("fake"))
	if err != nil {
		t.Fatalf("InitWithProtector() error = %v", err)
	}
	ns := vault.Namespace{Connector: "twitter-web", Profile: "default", Kind: "session"}
	if err := v.PutBlob(ctx, ns, []byte("super-secret-cookie-jar")); err != nil {
		t.Fatalf("PutBlob() error = %v", err)
	}

	path := filepath.Join(root, "vault", "ns", "twitter-web", "default", "session.blob")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if strings.Contains(string(raw), "super-secret-cookie-jar") {
		t.Fatalf("blob file contains plaintext secret")
	}
}

func TestVaultPerAccountIsolation(t *testing.T) {
	ctx := context.Background()
	v := testVault(t)

	nsA := vault.Namespace{Connector: "reddit", Profile: "personal", Kind: "oauth"}
	nsB := vault.Namespace{Connector: "reddit", Profile: "work", Kind: "oauth"}
	if err := v.PutNamespaced(ctx, nsA, map[string]string{"access_token": "token-a"}); err != nil {
		t.Fatalf("PutNamespaced(a) error = %v", err)
	}
	if err := v.PutNamespaced(ctx, nsB, map[string]string{"access_token": "token-b"}); err != nil {
		t.Fatalf("PutNamespaced(b) error = %v", err)
	}

	list, err := v.List(ctx, "reddit")
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("List() returned %d entries, want 2: %+v", len(list), list)
	}

	removed, err := v.DeleteAll(ctx, "reddit", "personal")
	if err != nil {
		t.Fatalf("DeleteAll() error = %v", err)
	}
	if removed != 1 {
		t.Fatalf("DeleteAll() removed = %d, want 1", removed)
	}

	if _, err := v.GetNamespaced(ctx, nsA); err == nil {
		t.Fatalf("profile %q credential survived DeleteAll", nsA.Profile)
	}
	got, err := v.GetNamespaced(ctx, nsB)
	if err != nil {
		t.Fatalf("GetNamespaced(b) error = %v", err)
	}
	if got["access_token"] != "token-b" {
		t.Fatalf("unrelated profile %q was affected by DeleteAll(%q)", nsB.Profile, nsA.Profile)
	}
}

func TestVaultRejectsPathTraversalInNamespace(t *testing.T) {
	ctx := context.Background()
	v := testVault(t)
	bad := []vault.Namespace{
		{Connector: "../../etc", Profile: "default", Kind: "session"},
		{Connector: "reddit", Profile: "../escape", Kind: "session"},
		{Connector: "reddit", Profile: "default", Kind: "../../escape"},
		{Connector: "reddit/x", Profile: "default", Kind: "session"},
	}
	for _, ns := range bad {
		if err := v.PutNamespaced(ctx, ns, map[string]string{"x": "y"}); err == nil {
			t.Fatalf("PutNamespaced(%+v): want error, got nil", ns)
		}
	}
}

func TestVaultRotateKeyReencryptsEverything(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	protector := newFakeProtector("fake")
	v, err := vault.InitWithProtector(root, protector)
	if err != nil {
		t.Fatalf("InitWithProtector() error = %v", err)
	}

	if err := v.Put(ctx, "cred_flat", map[string]string{"token": "flat-secret"}); err != nil {
		t.Fatalf("Put() error = %v", err)
	}
	ns := vault.Namespace{Connector: "linkedin-web", Profile: "default", Kind: "session"}
	if err := v.PutNamespaced(ctx, ns, map[string]string{"li_at": "ns-secret"}); err != nil {
		t.Fatalf("PutNamespaced() error = %v", err)
	}
	blobNS := vault.Namespace{Connector: "whatsapp-web", Profile: "default", Kind: "device_store"}
	if err := v.PutBlob(ctx, blobNS, []byte("blob-secret")); err != nil {
		t.Fatalf("PutBlob() error = %v", err)
	}

	oldKeyCopy := append([]byte(nil), protector.keys[filepath.Join(root, "vault")]...)

	if err := v.RotateKey(ctx); err != nil {
		t.Fatalf("RotateKey() error = %v", err)
	}

	newKey := protector.keys[filepath.Join(root, "vault")]
	if string(newKey) == string(oldKeyCopy) {
		t.Fatalf("RotateKey() did not change the persisted key")
	}

	flat, err := v.Get(ctx, "cred_flat")
	if err != nil {
		t.Fatalf("Get() after rotation error = %v", err)
	}
	if flat["token"] != "flat-secret" {
		t.Fatalf("flat token after rotation = %q", flat["token"])
	}
	nsGot, err := v.GetNamespaced(ctx, ns)
	if err != nil {
		t.Fatalf("GetNamespaced() after rotation error = %v", err)
	}
	if nsGot["li_at"] != "ns-secret" {
		t.Fatalf("namespaced secret after rotation = %q", nsGot["li_at"])
	}
	blob, err := v.GetBlob(ctx, blobNS)
	if err != nil {
		t.Fatalf("GetBlob() after rotation error = %v", err)
	}
	if string(blob) != "blob-secret" {
		t.Fatalf("blob after rotation = %q", blob)
	}

	// No leftover staging files.
	err = filepath.WalkDir(filepath.Join(root, "vault"), func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, ".rotate-tmp") {
			t.Fatalf("leftover rotation staging file: %s", path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("WalkDir() error = %v", err)
	}

	// A fresh vault opened against the same protector must see the new key,
	// not decrypt with a stale in-memory copy of the old one.
	v2, err := vault.InitWithProtector(root, protector)
	if err != nil {
		t.Fatalf("re-InitWithProtector() error = %v", err)
	}
	got2, err := v2.Get(ctx, "cred_flat")
	if err != nil {
		t.Fatalf("Get() from fresh vault handle error = %v", err)
	}
	if got2["token"] != "flat-secret" {
		t.Fatalf("fresh vault handle token = %q", got2["token"])
	}
}

func TestInitFallsBackWhenKeychainUnavailable(t *testing.T) {
	// A brand-new vault (no marker, no legacy plaintext key) whose preferred
	// protector fails must fall back and record that it did, never silently
	// mint a vault whose protection strength the caller can't observe.
	root := t.TempDir()
	failing := &fakeProtector{name: "keychain-like", fail: true}
	fallback := newFakeProtector("fallback")

	v, err := vault.InitWithProtectorChainForTest(root, failing, fallback)
	if err != nil {
		t.Fatalf("init error = %v", err)
	}
	if !v.UsingFallbackKeyProtection() {
		t.Fatalf("UsingFallbackKeyProtection() = false, want true")
	}
	if v.KeyProtection() != "fallback" {
		t.Fatalf("KeyProtection() = %q, want %q", v.KeyProtection(), "fallback")
	}
}

func TestVaultProtectorPinnedAcrossReopen(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	first := newFakeProtector("first")
	v, err := vault.InitWithProtector(root, first)
	if err != nil {
		t.Fatalf("InitWithProtector() error = %v", err)
	}
	if err := v.Put(ctx, "cred_a", map[string]string{"x": "y"}); err != nil {
		t.Fatalf("Put() error = %v", err)
	}

	// Reopening with a DIFFERENT explicit protector still pins to whatever
	// that call passes (InitWithProtector always uses the protector given),
	// but Init's own auto-selection (no explicit protector) must reuse the
	// SAME protector recorded on first init, never silently re-select.
	pinned, ok, err := vault.ReadProtectorMarkerForTest(root)
	if err != nil {
		t.Fatalf("ReadProtectorMarkerForTest() error = %v", err)
	}
	if !ok || pinned != "first" {
		t.Fatalf("protector marker = (%q, %v), want (\"first\", true)", pinned, ok)
	}
}

func TestVaultRejectsWrongSizedProtectorKey(t *testing.T) {
	root := t.TempDir()
	bad := badSizedProtector{}
	if _, err := vault.InitWithProtector(root, bad); err == nil {
		t.Fatalf("InitWithProtector() with a bad-sized key: want error, got nil")
	}
}

type badSizedProtector struct{}

func (badSizedProtector) Name() string                           { return "bad" }
func (badSizedProtector) LoadOrCreateKey(string) ([]byte, error) { return []byte("too-short"), nil }

// TestListReturnsOneEntryPerNamespaceWithBothFiles pins List's contract that
// a logical namespace appears exactly once regardless of how many of its two
// backing files exist, and that RotateKey therefore stages each file once.
// A duplicated entry made RotateKey stage two items over one tmp path, commit
// the first rename, then fail the second with "no such file or directory" —
// after the new key had already been persisted, leaving the caller told
// rotation failed while it had in fact succeeded.
func TestListReturnsOneEntryPerNamespaceWithBothFiles(t *testing.T) {
	v := testVault(t)
	ctx := t.Context()
	ns := vault.Namespace{Connector: "whatsapp", Profile: "default", Kind: "session"}

	if err := v.PutNamespaced(ctx, ns, map[string]string{"token": "secret-value"}); err != nil {
		t.Fatalf("PutNamespaced: %v", err)
	}
	if err := v.PutBlob(ctx, ns, []byte("device-store-bytes")); err != nil {
		t.Fatalf("PutBlob: %v", err)
	}

	got, err := v.List(ctx, "whatsapp")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 1 || got[0] != ns {
		t.Fatalf("List() = %+v, want exactly one entry %+v", got, ns)
	}

	if err := v.RotateKey(ctx); err != nil {
		t.Fatalf("RotateKey with both a credential and a blob in one namespace: %v", err)
	}

	secret, err := v.GetNamespaced(ctx, ns)
	if err != nil {
		t.Fatalf("GetNamespaced after rotation: %v", err)
	}
	if secret["token"] != "secret-value" {
		t.Fatalf("credential after rotation = %v, want token=secret-value", secret)
	}
	blob, err := v.GetBlob(ctx, ns)
	if err != nil {
		t.Fatalf("GetBlob after rotation: %v", err)
	}
	if string(blob) != "device-store-bytes" {
		t.Fatalf("blob after rotation = %q, want device-store-bytes", blob)
	}
}
