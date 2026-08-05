package vault

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type rotationTestProtector struct {
	key      []byte
	storeErr error
}

func (p *rotationTestProtector) Name() string { return "rotation-test" }

func (p *rotationTestProtector) LoadOrCreateKey(string) ([]byte, error) {
	if p.key == nil {
		p.key = make([]byte, 32)
		for i := range p.key {
			p.key[i] = byte(i + 1)
		}
	}
	return append([]byte(nil), p.key...), nil
}

func (p *rotationTestProtector) StoreKey(_ string, key []byte) error {
	if p.storeErr != nil {
		return p.storeErr
	}
	p.key = append([]byte(nil), key...)
	return nil
}

func TestRotateKeyRestoresOldCiphertextWhenCommitRenameFails(t *testing.T) {
	root := t.TempDir()
	protector := &rotationTestProtector{}
	v, err := InitWithProtector(root, protector)
	if err != nil {
		t.Fatalf("InitWithProtector() error = %v", err)
	}
	ctx := context.Background()
	for id, token := range map[string]string{"alpha": "first", "beta": "second"} {
		if err := v.Put(ctx, id, map[string]string{"token": token}); err != nil {
			t.Fatalf("Put(%q) error = %v", id, err)
		}
	}
	oldKey := append([]byte(nil), protector.key...)

	originalRename := rotateRename
	calls := 0
	rotateRename = func(oldPath, newPath string) error {
		calls++
		if calls == 2 {
			return errors.New("injected rename failure")
		}
		return originalRename(oldPath, newPath)
	}
	t.Cleanup(func() { rotateRename = originalRename })

	if err := v.RotateKey(ctx); err == nil {
		t.Fatal("RotateKey() error = nil, want commit failure")
	}
	if string(protector.key) != string(oldKey) {
		t.Fatal("RotateKey() persisted a new key after a commit failure")
	}
	for id, want := range map[string]string{"alpha": "first", "beta": "second"} {
		got, err := v.Get(ctx, id)
		if err != nil {
			t.Fatalf("Get(%q) after failed rotation: %v", id, err)
		}
		if got["token"] != want {
			t.Fatalf("Get(%q)[token] = %q, want %q", id, got["token"], want)
		}
	}

	reopened, err := InitWithProtector(root, protector)
	if err != nil {
		t.Fatalf("reopen vault: %v", err)
	}
	if _, err := reopened.Get(ctx, "alpha"); err != nil {
		t.Fatalf("reopened vault cannot read restored ciphertext: %v", err)
	}
	err = filepath.WalkDir(filepath.Join(root, "vault"), func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if strings.HasSuffix(path, ".rotate-tmp") {
			t.Fatalf("leftover rotation staging file: %s", path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("WalkDir() error = %v", err)
	}
}

func TestRotateKeyRestoresOldCiphertextWhenStoreFails(t *testing.T) {
	root := t.TempDir()
	protector := &rotationTestProtector{}
	v, err := InitWithProtector(root, protector)
	if err != nil {
		t.Fatalf("InitWithProtector() error = %v", err)
	}
	ctx := context.Background()
	if err := v.Put(ctx, "credential", map[string]string{"token": "value"}); err != nil {
		t.Fatalf("Put() error = %v", err)
	}
	oldKey := append([]byte(nil), protector.key...)
	protector.storeErr = errors.New("injected key storage failure")

	if err := v.RotateKey(ctx); err == nil {
		t.Fatal("RotateKey() error = nil, want key storage failure")
	}
	if string(protector.key) != string(oldKey) {
		t.Fatal("RotateKey() changed the persisted key after storage failure")
	}
	got, err := v.Get(ctx, "credential")
	if err != nil {
		t.Fatalf("Get() after failed rotation: %v", err)
	}
	if got["token"] != "value" {
		t.Fatalf("Get()[token] = %q, want value", got["token"])
	}
}
