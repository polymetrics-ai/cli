package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGovernedPathRejectsEscape(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	inside := filepath.Join(root, "context.md")
	if err := os.WriteFile(inside, []byte("context"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := governedPath(root, inside); err != nil {
		t.Fatalf("inside path rejected: %v", err)
	}
	outside := filepath.Join(t.TempDir(), "context.md")
	if err := os.WriteFile(outside, []byte("context"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := governedPath(root, outside); err == nil {
		t.Fatal("expected path escape to fail")
	}
}

func TestLoadConfigRejectsUnknownFields(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "config.json")
	raw := `{"gsd_command":["gsd"],"work_dir":"/tmp/work","gsd_home":"/tmp/home","state_dir":"/tmp/state","unexpected":true}`
	if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := loadConfig(path); err == nil {
		t.Fatal("expected unknown config field to fail")
	}
}
