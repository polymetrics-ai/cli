package defs

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"slices"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors/engine"
)

const (
	githubSourceLockSHA256 = "79f6eaf203394aabe7d2558d0f87a8100a7a084b2e39bd264f7773f235acf2c8"
)

func TestProductionEmbedLoadsRuntimeBundles(t *testing.T) {
	bundles, err := engine.LoadAll(FS)
	if err != nil {
		t.Fatalf("LoadAll(FS): %v", err)
	}
	if len(bundles) == 0 {
		t.Fatal("LoadAll(FS) returned zero bundles")
	}

	var github *engine.Bundle
	for i := range bundles {
		if bundles[i].Name == "github" {
			github = &bundles[i]
			break
		}
	}
	if github == nil {
		t.Fatal("LoadAll(FS) missing github bundle")
	}
	if github.Metadata.Name != "github" {
		t.Fatalf("github metadata name = %q", github.Metadata.Name)
	}
	if len(github.Streams) == 0 {
		t.Fatal("github bundle has zero streams")
	}
	if github.Docs == "" {
		t.Fatal("github bundle docs are empty")
	}
	if github.Surface != nil {
		t.Fatal("production embed should not include api_surface.json")
	}
	if github.Fixtures != nil {
		t.Fatal("production embed should not include fixtures")
	}
}

func TestProductionEmbedExcludesConformanceArtifacts(t *testing.T) {
	for _, path := range []string{"github/api_surface.json", "github/fixtures"} {
		if _, err := fs.Stat(FS, path); !errors.Is(err, fs.ErrNotExist) {
			t.Fatalf("fs.Stat(%q) err = %v, want fs.ErrNotExist", path, err)
		}
	}
}

func TestProductionEmbedDeclaresOnlyGithubSourceLockException(t *testing.T) {
	defsSource, err := os.ReadFile("defs.go")
	if err != nil {
		t.Fatalf("ReadFile(defs.go): %v", err)
	}

	var directive string
	for _, line := range strings.Split(string(defsSource), "\n") {
		if strings.HasPrefix(line, "//go:embed ") {
			directive = strings.TrimPrefix(line, "//go:embed ")
			break
		}
	}
	if directive == "" {
		t.Fatal("defs.go has no //go:embed directive")
	}
	if strings.Contains(directive, "*/sources/*") {
		t.Fatalf("source locks must be explicitly allowlisted, directive = %q", directive)
	}
	if !strings.Contains(directive, githubGraphQLSourceLockPath) {
		t.Fatalf("directive = %q, want explicit %q exception", directive, githubGraphQLSourceLockPath)
	}

	report, err := Inventory()
	if err != nil {
		t.Fatalf("Inventory(): %v", err)
	}
	if len(report.Files) == 0 {
		t.Fatal("Inventory() returned zero embedded files")
	}
	if !slices.IsSortedFunc(report.Files, func(a, b EmbeddedInventoryFile) int {
		return strings.Compare(a.Path, b.Path)
	}) {
		t.Fatalf("inventory files are not sorted: %#v", report.Files)
	}
	if !slices.IsSortedFunc(report.Classes, func(a, b EmbeddedInventoryClass) int {
		return strings.Compare(a.Class, b.Class)
	}) {
		t.Fatalf("inventory classes are not sorted: %#v", report.Classes)
	}
}

func TestProductionEmbedPreservesGithubSourceLockBytes(t *testing.T) {
	diskBytes, err := os.ReadFile(githubGraphQLSourceLockPath)
	if err != nil {
		t.Fatalf("ReadFile(%q): %v", githubGraphQLSourceLockPath, err)
	}
	embeddedBytes, err := fs.ReadFile(FS, githubGraphQLSourceLockPath)
	if err != nil {
		t.Fatalf("fs.ReadFile(FS, %q): %v", githubGraphQLSourceLockPath, err)
	}
	if !bytes.Equal(embeddedBytes, diskBytes) {
		t.Fatalf("embedded %q differs from the committed source lock", githubGraphQLSourceLockPath)
	}
	digest := sha256.Sum256(embeddedBytes)
	if got := hex.EncodeToString(digest[:]); got != githubSourceLockSHA256 {
		t.Fatalf("SHA-256(%q) = %s, want %s", githubGraphQLSourceLockPath, got, githubSourceLockSHA256)
	}
}

func TestProductionEmbedInventoryIsDeterministicAndAttributed(t *testing.T) {
	first, err := Inventory()
	if err != nil {
		t.Fatalf("first Inventory(): %v", err)
	}
	second, err := Inventory()
	if err != nil {
		t.Fatalf("second Inventory(): %v", err)
	}
	firstJSON, err := json.Marshal(first)
	if err != nil {
		t.Fatalf("json.Marshal(first inventory): %v", err)
	}
	secondJSON, err := json.Marshal(second)
	if err != nil {
		t.Fatalf("json.Marshal(second inventory): %v", err)
	}
	if !bytes.Equal(firstJSON, secondJSON) {
		t.Fatalf("inventory JSON is not deterministic:\nfirst:  %s\nsecond: %s", firstJSON, secondJSON)
	}

	var filesTotal, sourceLockBytes int64
	for _, file := range first.Files {
		filesTotal += file.Bytes
		if file.Path == githubGraphQLSourceLockPath {
			if file.Class != "certification_source_lock_exception" {
				t.Fatalf("GitHub source lock class = %q", file.Class)
			}
			sourceLockBytes = file.Bytes
		}
	}
	if filesTotal != first.TotalBytes {
		t.Fatalf("file bytes = %d, report total = %d", filesTotal, first.TotalBytes)
	}
	if sourceLockBytes == 0 {
		t.Fatalf("inventory omits %q", githubGraphQLSourceLockPath)
	}

	var classesTotal int64
	for _, class := range first.Classes {
		classesTotal += class.Bytes
	}
	if classesTotal != first.TotalBytes {
		t.Fatalf("class bytes = %d, report total = %d", classesTotal, first.TotalBytes)
	}
}

func TestEmbeddedArtifactClassRejectsBuildTimeArtifacts(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		wantClass string
		wantErr   bool
	}{
		{name: "runtime metadata", path: "github/metadata.json", wantClass: "metadata"},
		{name: "runtime schema", path: "github/schemas/repository.json", wantClass: "schema"},
		{name: "runtime ledger", path: "operation_endpoint_ledger.json", wantClass: "runtime_endpoint_ledger"},
		{name: "GitHub certification exception", path: githubGraphQLSourceLockPath, wantClass: "certification_source_lock_exception"},
		{name: "API surface", path: "github/api_surface.json", wantErr: true},
		{name: "root fixture", path: "fixtures/page.json", wantErr: true},
		{name: "fixture nested below sources", path: "github/fixtures/streams/sources/page_1.json", wantErr: true},
		{name: "root source lock", path: "sources/operations.json", wantErr: true},
		{name: "non-exempt source lock", path: "hubspot/sources/operations.json", wantErr: true},
		{name: "unknown artifact", path: "github/generated.json", wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := embeddedArtifactClass(test.path)
			if test.wantErr {
				if err == nil {
					t.Fatalf("embeddedArtifactClass(%q) = %q, nil error", test.path, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("embeddedArtifactClass(%q): %v", test.path, err)
			}
			if got != test.wantClass {
				t.Fatalf("embeddedArtifactClass(%q) = %q, want %q", test.path, got, test.wantClass)
			}
		})
	}
}
