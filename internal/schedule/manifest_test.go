package schedule

import (
	"errors"
	"os"
	"testing"
	"time"
)

// Group B — manifest CRUD.

func makeManifest(name string) Manifest {
	return Manifest{
		Name:                   name,
		Cron:                   "0 2 * * *",
		Flow:                   "test-flow",
		AuthorizationReference: "auth_0123456789abcdef",
		CreatedAt:              time.Date(2026, 6, 27, 0, 0, 0, 0, time.UTC),
		UpdatedAt:              time.Date(2026, 6, 27, 0, 0, 0, 0, time.UTC),
	}
}

func TestManifestSaveLoad(t *testing.T) {
	root := t.TempDir()
	m := makeManifest("nightly-leads")

	if err := Save(root, m, false); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := Load(root, "nightly-leads")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.Name != m.Name || got.Cron != m.Cron || got.Flow != m.Flow || got.AuthorizationReference != m.AuthorizationReference {
		t.Fatalf("round-trip mismatch: got %+v, want %+v", got, m)
	}
	if !got.CreatedAt.Equal(m.CreatedAt) || !got.UpdatedAt.Equal(m.UpdatedAt) {
		t.Fatalf("timestamp mismatch: got %v/%v", got.CreatedAt, got.UpdatedAt)
	}
}

func TestManifestRejectsMissingAuthorizationReference(t *testing.T) {
	root := t.TempDir()
	m := makeManifest("missing-authorization")
	m.AuthorizationReference = ""
	if err := Save(root, m, false); err == nil {
		t.Fatal("Save without an authorization reference succeeded")
	}
}

func TestManifestRejectsNonOpaqueAuthorizationReference(t *testing.T) {
	root := t.TempDir()
	m := makeManifest("token-like-authorization")
	m.AuthorizationReference = "approval-token-fixture"
	if err := Save(root, m, false); err == nil {
		t.Fatal("Save accepted token-like authorization material")
	}
}

func TestManifestLoadRejectsUnsafeAuthorizationReference(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(schedulesDir(root), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	path := manifestPath(root, "unsafe-reference")
	if err := os.WriteFile(path, []byte(`{"name":"unsafe-reference","cron":"0 2 * * *","flow":"flow","authorization_reference":"approval-token-fixture"}`), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if _, err := Load(root, "unsafe-reference"); err == nil {
		t.Fatal("Load accepted token-like authorization material")
	}
}

func TestManifestList(t *testing.T) {
	root := t.TempDir()
	names := []string{"alpha", "beta", "gamma"}
	for _, n := range names {
		if err := Save(root, makeManifest(n), false); err != nil {
			t.Fatalf("Save %q: %v", n, err)
		}
	}
	list, err := List(root)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != len(names) {
		t.Fatalf("List returned %d items, want %d", len(list), len(names))
	}
}

func TestManifestListIgnoresFireState(t *testing.T) {
	root := t.TempDir()
	manifest := makeManifest("fired")
	if err := Save(root, manifest, false); err != nil {
		t.Fatalf("Save: %v", err)
	}
	lease, err := BeginFire(root, manifest.Name)
	if err != nil {
		t.Fatalf("BeginFire: %v", err)
	}
	if err := lease.Complete(FireReceipt{ReceiptIDs: []string{"receipt-safe"}}); err != nil {
		t.Fatalf("Complete: %v", err)
	}

	list, err := List(root)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 1 || list[0].Name != manifest.Name {
		t.Fatalf("List returned %#v, want only %q", list, manifest.Name)
	}
}

func TestManifestDuplicateRejected(t *testing.T) {
	root := t.TempDir()
	m := makeManifest("dup-test")
	if err := Save(root, m, false); err != nil {
		t.Fatalf("first Save: %v", err)
	}
	err := Save(root, m, false)
	if err == nil {
		t.Fatal("expected error for duplicate name, got nil")
	}
	if !containsStr(err.Error(), "already exists") {
		t.Fatalf("error %q does not contain 'already exists'", err.Error())
	}
}

func TestManifestInvalidNameRejected(t *testing.T) {
	root := t.TempDir()
	bad := []string{"UPPER", "-leading", "has space", ""}
	for _, name := range bad {
		m := Manifest{Name: name, Cron: "0 2 * * *", Flow: "f"}
		if err := Save(root, m, false); err == nil {
			t.Fatalf("Save with name=%q expected error, got nil", name)
		}
	}
}

func TestManifestDelete(t *testing.T) {
	root := t.TempDir()
	m := makeManifest("delete-me")
	if err := Save(root, m, false); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if err := Delete(root, "delete-me"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	_, err := Load(root, "delete-me")
	if err == nil {
		t.Fatal("Load after Delete expected error, got nil")
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected ErrNotExist, got %v", err)
	}
}

func TestManifestLoadNonExistent(t *testing.T) {
	root := t.TempDir()
	_, err := Load(root, "ghost")
	if err == nil {
		t.Fatal("expected error loading non-existent manifest")
	}
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}
