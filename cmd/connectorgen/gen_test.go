package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenWritesDeterministicManifestIndex(t *testing.T) {
	root := t.TempDir()
	hooksRoot := filepath.Join(root, "internal", "connectors", "hooks")
	if err := os.MkdirAll(hooksRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	defsRoot := filepath.Join(root, "internal", "connectors", "defs", "alpha")
	if err := os.MkdirAll(defsRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(defsRoot, "metadata.json"), []byte(`{"name":"alpha"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	if got := runGenAt(nil, io.Discard, io.Discard, hooksRoot); got != 0 {
		t.Fatalf("runGenAt() = %d, want 0", got)
	}
	indexPath := filepath.Join(root, "internal", "connectors", "manifestindex", "index_gen.go")
	first, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("read generated index: %v", err)
	}
	if !strings.Contains(string(first), `Connector: "alpha"`) || !strings.Contains(string(first), `Executor: "api_engine.v1"`) {
		t.Fatalf("generated index = %s, want alpha api engine entry", first)
	}
	if got := runGenAt(nil, io.Discard, io.Discard, hooksRoot); got != 0 {
		t.Fatalf("second runGenAt() = %d, want 0", got)
	}
	second, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("read second generated index: %v", err)
	}
	if string(first) != string(second) {
		t.Fatalf("generated index changed between identical inputs:\nfirst=%s\nsecond=%s", first, second)
	}
}

func TestGeneratedManifestIndexDigestIncludesRateLimits(t *testing.T) {
	root := t.TempDir()
	hooksRoot := filepath.Join(root, "internal", "connectors", "hooks")
	defsRoot := filepath.Join(root, "internal", "connectors", "defs", "alpha")
	if err := os.MkdirAll(hooksRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(defsRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(defsRoot, "metadata.json"), []byte(`{"name":"alpha","display_name":"Alpha","integration_type":"api","capabilities":{"check":true}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	indexPath := filepath.Join(root, "internal", "connectors", "manifestindex", "index_gen.go")
	if got := runGenAt(nil, io.Discard, io.Discard, hooksRoot); got != 0 {
		t.Fatalf("baseline runGenAt() = %d, want 0", got)
	}
	withoutRate, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(defsRoot, "rate_limits.json"), []byte(`{"state":"not_applicable","policies":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := runGenAt(nil, io.Discard, io.Discard, hooksRoot); got != 0 {
		t.Fatalf("rate-file runGenAt() = %d, want 0", got)
	}
	withRate, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(withoutRate) == string(withRate) {
		t.Fatal("rate_limits.json did not change generated manifest digest or byte charge")
	}
	if err := os.Remove(filepath.Join(defsRoot, "rate_limits.json")); err != nil {
		t.Fatal(err)
	}
	if got := runGenAt(nil, io.Discard, io.Discard, hooksRoot); got != 0 {
		t.Fatalf("removed-rate runGenAt() = %d, want 0", got)
	}
	restored, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(restored) != string(withoutRate) {
		t.Fatal("removing rate_limits.json did not restore the closed generated index")
	}
}

func TestHookPackagesDoNotRegisterGlobalHooks(t *testing.T) {
	root, err := repoRoot()
	if err != nil {
		t.Fatal(err)
	}
	hooksRoot := filepath.Join(root, "internal", "connectors", "hooks")
	err = filepath.Walk(hooksRoot, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		source, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if strings.Contains(string(source), "engine.RegisterHooks(") {
			t.Fatalf("legacy global hook registration remains in %s", path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
