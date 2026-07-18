package connectors

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestWarehouseWriteRejectsFinalFileSymlinkEscape(t *testing.T) {
	tests := []struct {
		name           string
		overwrite      bool
		externalExists bool
	}{
		{name: "append", overwrite: false, externalExists: true},
		{name: "truncate", overwrite: true, externalExists: true},
		{name: "create", overwrite: false, externalExists: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			dir := filepath.Join(root, "warehouse")
			if err := os.Mkdir(dir, 0o700); err != nil {
				t.Fatalf("create warehouse directory: %v", err)
			}
			external := filepath.Join(t.TempDir(), "outside.jsonl")
			before := []byte("outside-sentinel\n")
			if tt.externalExists {
				if err := os.WriteFile(external, before, 0o600); err != nil {
					t.Fatalf("create external target: %v", err)
				}
			}
			if err := os.Symlink(external, filepath.Join(dir, "records.jsonl")); err != nil {
				t.Skipf("symlinks unavailable on this platform: %v", err)
			}

			_, err := (Warehouse{}).Write(context.Background(), WriteRequest{
				Table:     "records",
				Overwrite: tt.overwrite,
				Config: RuntimeConfig{
					Config: map[string]string{"path": dir},
					LocalWritePolicy: &LocalWritePolicy{
						ProjectRoot: root,
					},
				},
			}, []Record{{"id": "synthetic"}})
			if err == nil {
				t.Fatal("Warehouse.Write() followed a final-file symlink outside the local-write root")
			}
			assertExternalTargetUnchanged(t, external, before, tt.externalExists)
		})
	}
}

func TestOutboxWriteRejectsFinalFileSymlinkEscape(t *testing.T) {
	for _, externalExists := range []bool{true, false} {
		name := "append"
		if !externalExists {
			name = "create"
		}
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			dir := filepath.Join(root, "outbox")
			if err := os.Mkdir(dir, 0o700); err != nil {
				t.Fatalf("create outbox directory: %v", err)
			}
			external := filepath.Join(t.TempDir(), "outside.jsonl")
			before := []byte("outside-sentinel\n")
			if externalExists {
				if err := os.WriteFile(external, before, 0o600); err != nil {
					t.Fatalf("create external target: %v", err)
				}
			}
			if err := os.Symlink(external, filepath.Join(dir, "records.jsonl")); err != nil {
				t.Skipf("symlinks unavailable on this platform: %v", err)
			}

			_, err := (Outbox{}).Write(context.Background(), WriteRequest{
				Table: "records",
				Config: RuntimeConfig{
					Config: map[string]string{"path": dir},
					LocalWritePolicy: &LocalWritePolicy{
						ProjectRoot: root,
					},
				},
			}, []Record{{"id": "synthetic"}})
			if err == nil {
				t.Fatal("Outbox.Write() followed a final-file symlink outside the local-write root")
			}
			assertExternalTargetUnchanged(t, external, before, externalExists)
		})
	}
}

func assertExternalTargetUnchanged(t *testing.T, path string, before []byte, existed bool) {
	t.Helper()
	after, err := os.ReadFile(path)
	if !existed {
		if !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("denied write created an external target: %v", err)
		}
		return
	}
	if err != nil {
		t.Fatalf("read external target after denied write: %v", err)
	}
	if !bytes.Equal(after, before) {
		t.Fatal("denied write changed the external target")
	}
}
