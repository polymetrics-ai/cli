package rlm

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestDeterministicRunRejectsWarehouseInputPathEscape(t *testing.T) {
	root := t.TempDir()
	warehouse := filepath.Join(root, "warehouse")
	if err := os.Mkdir(warehouse, 0o700); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(root, "outside.ndjson")
	if err := os.WriteFile(outside, []byte("{\"id\":\"outside\"}\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := (&DeterministicAnalyzer{}).Run(t.Context(), RunRequest{
		Spec:         &Spec{Name: "safe", Features: []Feature{{Name: "id", Weight: 1, ScoreIfSet: 1}}},
		InTable:      "../outside",
		OutTable:     "scores",
		WarehouseDir: warehouse,
	})
	if err == nil {
		t.Fatal("path-escaping input table succeeded")
	}
	if _, statErr := os.Stat(filepath.Join(warehouse, "scores.ndjson")); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("unsafe input created output: %v", statErr)
	}
}

func TestDeterministicRunRejectsExternalInputFinalLink(t *testing.T) {
	warehouse := t.TempDir()
	external := filepath.Join(t.TempDir(), "outside.ndjson")
	if err := os.WriteFile(external, []byte("{\"id\":\"outside\"}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(external, filepath.Join(warehouse, "contacts.ndjson")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	_, err := (&DeterministicAnalyzer{}).Run(t.Context(), RunRequest{
		Spec:         &Spec{Name: "safe", Features: []Feature{{Name: "id", Weight: 1, ScoreIfSet: 1}}},
		InTable:      "contacts",
		OutTable:     "scores",
		WarehouseDir: warehouse,
	})
	if err == nil {
		t.Fatal("external input final link succeeded")
	}
	if _, statErr := os.Stat(filepath.Join(warehouse, "scores.ndjson")); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("linked input created output: %v", statErr)
	}
}

func TestWriteOutTableRejectsExternalTemporaryFinalLink(t *testing.T) {
	warehouse := t.TempDir()
	external := filepath.Join(t.TempDir(), "outside.ndjson")
	sentinel := []byte("outside-sentinel\n")
	if err := os.WriteFile(external, sentinel, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(external, filepath.Join(warehouse, "scores.ndjson.tmp")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	err := writeOutTable(warehouse, "scores", nil, "deterministic", "safe", "now")
	if err == nil {
		t.Fatal("external temporary output final link succeeded")
	}
	after, readErr := os.ReadFile(external)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if !bytes.Equal(after, sentinel) {
		t.Fatal("output writer followed and changed external temporary target")
	}
}

func TestWriteOutTableReplacesExternalFinalLinkWithoutFollowingIt(t *testing.T) {
	warehouse := t.TempDir()
	external := filepath.Join(t.TempDir(), "outside.ndjson")
	sentinel := []byte("outside-sentinel\n")
	if err := os.WriteFile(external, sentinel, 0o600); err != nil {
		t.Fatal(err)
	}
	finalPath := filepath.Join(warehouse, "scores.ndjson")
	if err := os.Symlink(external, finalPath); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	if err := writeOutTable(warehouse, "scores", nil, "deterministic", "safe", "now"); err != nil {
		t.Fatalf("safe atomic replacement failed: %v", err)
	}
	after, err := os.ReadFile(external)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(after, sentinel) {
		t.Fatal("output writer followed and changed external final target")
	}
	info, err := os.Lstat(finalPath)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		t.Fatal("output writer left final path as a symlink")
	}
}
