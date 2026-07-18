package app

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"polymetrics.ai/internal/connectors"
)

func TestWarehouseMaterializationRejectsFinalFileSymlinkEscape(t *testing.T) {
	tests := []struct {
		name           string
		mode           string
		externalExists bool
	}{
		{name: "append", mode: "full_refresh_append", externalExists: true},
		{name: "append_create", mode: "full_refresh_append", externalExists: false},
		{name: "truncate", mode: "full_refresh_overwrite", externalExists: true},
		{name: "truncate_create", mode: "full_refresh_overwrite", externalExists: false},
		{name: "deduped_truncate", mode: "full_refresh_overwrite_deduped", externalExists: true},
		{name: "deduped_create", mode: "full_refresh_overwrite_deduped", externalExists: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			root := t.TempDir()
			if err := InitProject(root); err != nil {
				t.Fatalf("InitProject() error = %v", err)
			}
			a, err := Open(root)
			if err != nil {
				t.Fatalf("Open() error = %v", err)
			}
			dir := filepath.Join(root, ".polymetrics", "warehouse")
			if err := os.MkdirAll(dir, 0o700); err != nil {
				t.Fatalf("create warehouse directory: %v", err)
			}
			mode, err := ParseSyncMode(tt.mode)
			if err != nil {
				t.Fatalf("ParseSyncMode() error = %v", err)
			}
			runID := "bounded-run"
			finalLink := filepath.Join(dir, "records.jsonl")
			if mode.IsOverwrite() || mode.IsDeduped() {
				finalLink += "." + runID + ".tmp"
			}
			external := filepath.Join(t.TempDir(), "outside.jsonl")
			before := []byte("outside-sentinel\n")
			if tt.externalExists {
				if err := os.WriteFile(external, before, 0o600); err != nil {
					t.Fatalf("create external target: %v", err)
				}
			}
			if err := os.Symlink(external, finalLink); err != nil {
				t.Skipf("symlinks unavailable on this platform: %v", err)
			}

			source := newScriptedSyncSource("symlink_source_"+tt.name, []connectors.Record{{
				"id":         "synthetic",
				"updated_at": "2026-01-01T00:00:00Z",
			}})
			_, err = a.runWarehouseETL(
				ctx,
				runID,
				Connection{Name: "synthetic_connection"},
				source,
				connectors.RuntimeConfig{},
				connectors.RuntimeConfig{
					Config: map[string]string{"path": dir},
					LocalWritePolicy: &connectors.LocalWritePolicy{
						ProjectRoot: root,
					},
				},
				"records",
				StreamConfig{
					DestinationTable: "records",
					CursorField:      "updated_at",
					PrimaryKey:       []string{"id"},
				},
				mode,
				1,
			)
			if err == nil {
				t.Fatal("warehouse materialization followed a final-file symlink outside the local-write root")
			}
			assertMaterializationExternalTargetUnchanged(t, external, before, tt.externalExists)
			if mode.IsOverwrite() {
				rawTemp := localRawPath(dir, "synthetic_connection", "records", "records") + "." + runID + ".tmp"
				if _, statErr := os.Stat(rawTemp); !errors.Is(statErr, os.ErrNotExist) {
					t.Fatalf("failed final-temp open left overwrite raw temp: %v", statErr)
				}
			}
		})
	}
}

func TestWarehouseOverwriteReplacesFinalSymlinkWithoutExternalEffect(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	if err := InitProject(root); err != nil {
		t.Fatalf("InitProject() error = %v", err)
	}
	a, err := Open(root)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	dir := filepath.Join(root, ".polymetrics", "warehouse")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("create warehouse directory: %v", err)
	}
	external := filepath.Join(t.TempDir(), "outside.jsonl")
	before := []byte("outside-sentinel\n")
	if err := os.WriteFile(external, before, 0o600); err != nil {
		t.Fatalf("create external target: %v", err)
	}
	finalPath := filepath.Join(dir, "records.jsonl")
	if err := os.Symlink(external, finalPath); err != nil {
		t.Skipf("symlinks unavailable on this platform: %v", err)
	}
	mode, err := ParseSyncMode("full_refresh_overwrite")
	if err != nil {
		t.Fatalf("ParseSyncMode() error = %v", err)
	}
	source := newScriptedSyncSource("symlink_rename_source", []connectors.Record{{"id": "synthetic"}})
	_, err = a.runWarehouseETL(
		ctx,
		"bounded-rename-run",
		Connection{Name: "synthetic_connection"},
		source,
		connectors.RuntimeConfig{},
		connectors.RuntimeConfig{
			Config: map[string]string{"path": dir},
			LocalWritePolicy: &connectors.LocalWritePolicy{
				ProjectRoot: root,
			},
		},
		"records",
		StreamConfig{DestinationTable: "records"},
		mode,
		1,
	)
	if err != nil {
		t.Fatalf("runWarehouseETL() error = %v", err)
	}
	assertMaterializationExternalTargetUnchanged(t, external, before, true)
	info, err := os.Lstat(finalPath)
	if err != nil {
		t.Fatalf("lstat final table: %v", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		t.Fatal("overwrite left the final table as a symlink")
	}
}

func assertMaterializationExternalTargetUnchanged(t *testing.T, path string, before []byte, existed bool) {
	t.Helper()
	after, err := os.ReadFile(path)
	if !existed {
		if !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("denied materialization created an external target: %v", err)
		}
		return
	}
	if err != nil {
		t.Fatalf("read external target after denied materialization: %v", err)
	}
	if !bytes.Equal(after, before) {
		t.Fatal("denied materialization changed the external target")
	}
}
