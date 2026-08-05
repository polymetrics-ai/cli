package agentcontract

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestReferencedGSDCommandsResolve(t *testing.T) {
	root := repositoryRoot(t)
	contract := loadRepositoryContract(t, root)

	if err := CheckGSDCommands(context.Background(), root, contract.GSD.Commands); err != nil {
		t.Fatalf("referenced GSD commands must resolve: %v", err)
	}
}

func TestCheckGSDCommandsRunsFromSelectedRoot(t *testing.T) {
	root := t.TempDir()
	script := filepath.Join(root, "scripts", "gsd")
	if err := os.MkdirAll(filepath.Dir(script), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".selected-root"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	contents := []byte("const fs = require(\"node:fs\");\nif (!fs.existsSync(\".selected-root\")) process.exit(1);\nif (process.argv[2] !== \"sources\" || process.argv[3] !== \"discuss-phase\") process.exit(1);\n")
	if err := os.WriteFile(script, contents, 0o644); err != nil {
		t.Fatal(err)
	}

	if err := CheckGSDCommands(context.Background(), root, []string{"discuss-phase"}); err != nil {
		t.Fatalf("GSD command did not run from selected root: %v", err)
	}
}

func TestCheckProjectionRejectsDivergence(t *testing.T) {
	want := []byte("canonical generated block\n")
	got := []byte("diverged generated block\n")

	if err := CheckProjection(want, got); err == nil {
		t.Fatal("CheckProjection accepted a diverged projection")
	}
}

func TestProjectionDriftCheckAndSync(t *testing.T) {
	repository := repositoryRoot(t)
	contract := loadRepositoryContract(t, repository)

	root := t.TempDir()
	updated, err := SyncProjections(root, contract)
	if err != nil {
		t.Fatalf("SyncProjections creates required Claude projections: %v", err)
	}
	if updated != 2 {
		t.Fatalf("SyncProjections created %d projections, want 2", updated)
	}
	if err := CheckProjections(root, contract); err != nil {
		t.Fatalf("matching projections failed: %v", err)
	}

	path := filepath.Join(root, filepath.FromSlash(contract.Projections[0].Path))
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	diverged := strings.Replace(string(content), "    - Bash", "    - Agent\n    - Bash", 1)
	if diverged == string(content) {
		t.Fatal("test fixture did not add Agent to the Claude tools allowlist")
	}
	if err := os.WriteFile(path, []byte(diverged), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := CheckProjections(root, contract); err == nil {
		t.Fatal("CheckProjections accepted a Claude worker that grants Agent and can delegate to an ambient agent")
	}

	updated, err = SyncProjections(root, contract)
	if err != nil {
		t.Fatalf("SyncProjections: %v", err)
	}
	if updated != 1 {
		t.Fatalf("SyncProjections updated %d files, want 1", updated)
	}
	if err := CheckProjections(root, contract); err != nil {
		t.Fatalf("projection did not pass after sync: %v", err)
	}
	restored, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	frontmatter, err := parseClaudeFrontmatter(restored)
	if err != nil {
		t.Fatal(err)
	}
	if slices.Contains(frontmatter.Tools, "Agent") {
		t.Fatalf("sync did not remove Agent from the canonical tools allowlist: %#v", frontmatter.Tools)
	}
	if slices.Contains(frontmatter.Tools, "Skill") {
		t.Fatalf("sync restored Skill instead of trusted preloads: %#v", frontmatter.Tools)
	}
	policy, ok := contract.ProjectionFor("claude")
	if !ok {
		t.Fatal("canonical contract does not define a Claude harness policy")
	}
	if !slices.Equal(frontmatter.Skills, policy.PreloadedSkills) {
		t.Fatalf("sync did not restore trusted skill preloads: %#v", frontmatter.Skills)
	}
	if !slices.Equal(frontmatter.DisallowedTools, []string{"Agent", "Task", "Skill"}) {
		t.Fatalf("sync did not restore the canonical Agent/Task/Skill denylist: %#v", frontmatter.DisallowedTools)
	}
}

func TestCheckProjectionsRejectsClaudeAgentInventoryDrift(t *testing.T) {
	contract := loadRepositoryContract(t, repositoryRoot(t))
	target := contract.Projections[0]
	projection, err := RenderProjection(contract, target)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name      string
		path      string
		content   []byte
		wantError string
	}{
		{
			name:      "duplicate canonical name",
			path:      ".claude/agents/shadow/pm-delivery-worker.md",
			content:   projection,
			wantError: "duplicate claude project agent name",
		},
		{
			name:      "unexpected definition",
			path:      ".claude/agents/shadow/unexpected-worker.md",
			content:   []byte(strings.Replace(string(projection), "name: pm-delivery-worker", "name: unexpected-worker", 1)),
			wantError: "unexpected definitions",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			if _, err := SyncProjections(root, contract); err != nil {
				t.Fatalf("SyncProjections creates required Claude projections: %v", err)
			}
			extraPath := filepath.Join(root, filepath.FromSlash(test.path))
			if err := os.MkdirAll(filepath.Dir(extraPath), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(extraPath, test.content, 0o644); err != nil {
				t.Fatal(err)
			}
			err := CheckProjections(root, contract)
			if err == nil || !strings.Contains(err.Error(), test.wantError) {
				t.Fatalf("CheckProjections error = %v, want substring %q", err, test.wantError)
			}
		})
	}
}

func TestOptionalWaveProjectionMayBeAbsent(t *testing.T) {
	contract := loadRepositoryContract(t, repositoryRoot(t))
	root := t.TempDir()
	if _, err := SyncProjections(root, contract); err != nil {
		t.Fatalf("SyncProjections creates the required Wave 2 projections: %v", err)
	}
	if err := CheckProjections(root, contract); err != nil {
		t.Fatalf("optional Wave 3-4 projections should be absent: %v", err)
	}
	for _, target := range contract.Projections {
		if target.Harness == "claude" {
			continue
		}
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(target.Path))); !os.IsNotExist(err) {
			t.Fatalf("optional projection %s exists after Claude-only sync: %v", target.Path, err)
		}
	}
}

func TestProjectionIORejectsSymlinkEscape(t *testing.T) {
	contract := loadRepositoryContract(t, repositoryRoot(t))
	target := contract.Projections[0]

	root := t.TempDir()
	outside := t.TempDir()
	escapedPath := filepath.Join(outside, "agents", filepath.Base(target.Path))
	if err := os.MkdirAll(filepath.Dir(escapedPath), 0o755); err != nil {
		t.Fatal(err)
	}
	original, err := RenderProjection(contract, target)
	if err != nil {
		t.Fatal(err)
	}
	original = []byte(strings.Replace(string(original), "Receive one assigned job", "Receive escaped work", 1))
	if err := os.WriteFile(escapedPath, original, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, ".claude")); err != nil {
		t.Skipf("cannot create projection ancestor symlink: %v", err)
	}

	if err := CheckProjections(root, contract); err == nil {
		t.Fatal("CheckProjections followed a projection ancestor outside the selected root")
	}
	if _, err := SyncProjections(root, contract); err == nil {
		t.Fatal("SyncProjections followed a projection ancestor outside the selected root")
	}
	after, err := os.ReadFile(escapedPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(after, original) {
		t.Fatal("SyncProjections modified a projection outside the selected root")
	}
}

func repositoryRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func loadRepositoryContract(t *testing.T, root string) *Contract {
	t.Helper()
	contract, err := Load(filepath.Join(root, SourcePath))
	if err != nil {
		t.Fatal(err)
	}
	return contract
}
