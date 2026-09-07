package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSourceFoundationExamplesCurrentReferences(t *testing.T) {
	repo, err := repoRoot()
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(repo, sourceDemandAtlasPath))
	if err != nil {
		t.Fatal(err)
	}
	rows, err := observeSourceFoundationExamples(t.Context(), repo, sourceArtifactPin{Path: sourceDemandAtlasPath, SHA256: sourceBytesHash(raw), Bytes: int64(len(raw))})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 35 {
		t.Fatalf("actual Atlas example membership=%d, want retained35", len(rows))
	}
	missing := map[string]bool{}
	seen := map[string]bool{}
	for _, row := range rows {
		identity := row.AtlasID + "#" + row.Pointer
		if seen[identity] || !sourceProofDigest(row.ValueSHA256) || row.Status != "example_unresolved" || row.Owner == "" || len(row.Files) == 0 {
			t.Fatal("example lost occurrence/owner or claimed unproved source adoption")
		}
		seen[identity] = true
		for _, file := range row.Files {
			switch file.State {
			case "missing":
				missing[file.Path] = true
				if file.SHA256 != "" || file.Bytes != 0 || row.Reason != "referenced_file_missing" {
					t.Fatal("missing reference invented file evidence")
				}
			case "present_regular":
				bytes, err := os.ReadFile(filepath.Join(repo, file.Path))
				if err != nil || sourceBytesHash(bytes) != file.SHA256 || int64(len(bytes)) != file.Bytes {
					t.Fatal("example file pin disagrees with real complete file")
				}
			default:
				t.Fatal("unknown example file state")
			}
		}
	}
	if len(missing) != 2 || !missing["internal/connectors/defs/asana/event_source_contract.json"] ||
		!missing["internal/connectors/defs/gitlab/enabled_connector_contract.json"] {
		t.Fatal("retained missing Asana/GitLab reference distinction changed")
	}
}

func TestSourceFoundationExamplesRejectSubstitutedFile(t *testing.T) {
	repo, _ := sourceFoundationUniverseFixture(t)
	raw, err := os.ReadFile(filepath.Join(repo, sourceDemandAtlasPath))
	if err != nil {
		t.Fatal(err)
	}
	pin := sourceArtifactPin{Path: sourceDemandAtlasPath, SHA256: sourceBytesHash(raw), Bytes: int64(len(raw))}
	if rows, err := observeSourceFoundationExamples(t.Context(), repo, pin); err != nil || len(rows) != 35 {
		t.Fatalf("real missing-reference observation control: %v", err)
	}
	name := filepath.Join(repo, "internal/connectors/defs/asana/event_source_contract.json")
	if err := os.MkdirAll(filepath.Dir(name), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(repo, sourceDemandAtlasPath), name); err != nil {
		t.Fatal(err)
	}
	if rows, err := observeSourceFoundationExamples(t.Context(), repo, pin); err == nil || len(rows) != 0 {
		t.Fatal("substituted reference followed or partial example result exposed")
	}
}
