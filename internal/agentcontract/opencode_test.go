package agentcontract

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestOpenCodeProjectionsAreRequiredGeneratedGateInputs(t *testing.T) {
	contract := loadRepositoryContract(t, repositoryRoot(t))
	root := t.TempDir()

	updated, err := SyncProjections(root, contract)
	if err != nil {
		t.Fatalf("SyncProjections creates registered OpenCode projections: %v", err)
	}
	if updated != 8 {
		t.Fatalf("SyncProjections created %d registered projections, want 8", updated)
	}
	if err := CheckProjections(root, contract); err != nil {
		t.Fatalf("generated projections must pass drift check: %v", err)
	}

	count := 0
	for _, target := range contract.Projections {
		if target.Harness != "opencode" {
			continue
		}
		count++
		if !target.Required || target.RenderMode != "opencode_markdown_yaml_frontmatter" {
			t.Fatalf("OpenCode target %#v is not a required canonical projection", target)
		}
		content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(target.Path)))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Contains(content, []byte("## Connector certification Shepherd gate")) ||
			!bytes.Contains(content, []byte("inputs.certification_shards")) ||
			!bytes.Contains(content, []byte("decision")) {
			t.Fatalf("OpenCode projection %s omitted canonical certification gate inputs or verdict fields", target.Path)
		}
	}
	if count != 2 {
		t.Fatalf("canonical contract registers %d OpenCode projections, want 2", count)
	}

	missing := filepath.Join(root, ".opencode", "agents", "pm-delivery-worker.md")
	if err := os.Remove(missing); err != nil {
		t.Fatal(err)
	}
	if err := CheckProjections(root, contract); err == nil {
		t.Fatal("CheckProjections accepted a missing registered OpenCode projection")
	}
}

func TestRegisteredHarnessesEmbedByteEquivalentCertificationGateInputs(t *testing.T) {
	contract := loadRepositoryContract(t, repositoryRoot(t))
	var canonicalBlock []byte
	seenHarnesses := make(map[string]bool)
	for _, target := range contract.Projections {
		projection, err := RenderProjection(contract, target)
		if err != nil {
			t.Fatalf("render %s projection: %v", target.Path, err)
		}
		var block []byte
		if target.Harness == "opencode" {
			block, err = openCodeProjectionGateBlock(projection)
		} else {
			if target.Harness == "codex" {
				projection = []byte(parseCodexProjection(t, projection).GetString("developer_instructions"))
			}
			block, err = extractCertificationGateBlock(projection)
		}
		if err != nil {
			t.Fatalf("extract certification gate from %s projection: %v", target.Path, err)
		}
		if canonicalBlock == nil {
			canonicalBlock = block
		} else if !bytes.Equal(canonicalBlock, block) {
			t.Fatalf("%s projection has adapter-local certification gate content\nwant=%s\ngot=%s", target.Path, canonicalBlock, block)
		}
		seenHarnesses[target.Harness] = true
	}
	for _, harness := range []string{"claude", "codex", "pi", "opencode"} {
		if !seenHarnesses[harness] {
			t.Fatalf("canonical projection registry omitted %s", harness)
		}
	}
}

func TestProjectionCheckRejectsCertificationGateDriftForEveryHarness(t *testing.T) {
	contract := loadRepositoryContract(t, repositoryRoot(t))
	root := t.TempDir()
	if updated, err := SyncProjections(root, contract); err != nil || updated != 8 {
		t.Fatalf("create canonical projections: updated=%d err=%v", updated, err)
	}
	for _, harness := range []string{"claude", "codex", "pi", "opencode"} {
		t.Run(harness, func(t *testing.T) {
			var target ProjectionTarget
			for _, candidate := range contract.Projections {
				if candidate.Harness == harness && candidate.Role == contract.BaseRole.Name {
					target = candidate
					break
				}
			}
			if target.Path == "" {
				t.Fatalf("missing %s delivery-worker projection", harness)
			}
			path := filepath.Join(root, filepath.FromSlash(target.Path))
			content, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			content = bytes.Replace(content, []byte("inputs.certification_shards"), []byte("inputs.adapter_local_certification_shards"), 1)
			if err := os.WriteFile(path, content, 0o644); err != nil {
				t.Fatal(err)
			}
			if err := CheckProjections(root, contract); err == nil {
				t.Fatalf("CheckProjections accepted %s adapter-local gate drift", harness)
			}
			updated, err := SyncProjections(root, contract)
			if err != nil || updated != 1 {
				t.Fatalf("SyncProjections repair %s drift: updated=%d err=%v", harness, updated, err)
			}
			if err := CheckProjections(root, contract); err != nil {
				t.Fatalf("repaired %s projection did not pass check: %v", harness, err)
			}
		})
	}
}
