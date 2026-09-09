package agentcontract

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/spf13/viper"
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
		t.Fatalf("SyncProjections creates required projections: %v", err)
	}
	if updated != 8 {
		t.Fatalf("SyncProjections created %d projections, want 8", updated)
	}
	if err := CheckProjections(root, contract); err != nil {
		t.Fatalf("matching projections failed: %v", err)
	}

	var claude ProjectionTarget
	for _, target := range contract.Projections {
		if target.Harness == "claude" {
			claude = target
			break
		}
	}
	if claude.Path == "" {
		t.Fatal("canonical contract does not register a Claude projection")
	}
	path := filepath.Join(root, filepath.FromSlash(claude.Path))
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

func TestClaudeProjectionCRLFIsNotDrift(t *testing.T) {
	contract := loadRepositoryContract(t, repositoryRoot(t))
	root := t.TempDir()
	if _, err := SyncProjections(root, contract); err != nil {
		t.Fatalf("SyncProjections creates required Claude projections: %v", err)
	}

	for _, target := range contract.Projections {
		if target.Harness != "claude" {
			continue
		}
		projectionPath := filepath.Join(root, filepath.FromSlash(target.Path))
		content, err := os.ReadFile(projectionPath)
		if err != nil {
			t.Fatal(err)
		}
		content = bytes.ReplaceAll(content, []byte("\n"), []byte("\r\n"))
		if err := os.WriteFile(projectionPath, content, 0o644); err != nil {
			t.Fatal(err)
		}
	}

	if err := CheckProjections(root, contract); err != nil {
		t.Fatalf("CheckProjections rejected canonical CRLF projections: %v", err)
	}
	updated, err := SyncProjections(root, contract)
	if err != nil {
		t.Fatalf("SyncProjections rejected canonical CRLF projections: %v", err)
	}
	if updated != 0 {
		t.Fatalf("SyncProjections updated %d canonical CRLF projections, want 0", updated)
	}
}

func TestPiProjectionsAreRequired(t *testing.T) {
	contract := loadRepositoryContract(t, repositoryRoot(t))
	count := 0
	for _, target := range contract.Projections {
		if target.Harness != "pi" {
			continue
		}
		count++
		if !target.Required {
			t.Fatalf("Pi projection %s must be required once its wave owns the generated file", target.Path)
		}
		if target.RenderMode != "full" {
			t.Fatalf("Pi projection %s must be a complete generated file", target.Path)
		}
	}
	if count != 2 {
		t.Fatalf("canonical contract registers %d Pi projections, want 2", count)
	}
}

func TestSyncCreatesRequiredPiProjections(t *testing.T) {
	contract := loadRepositoryContract(t, repositoryRoot(t))
	root := t.TempDir()
	updated, err := SyncProjections(root, contract)
	if err != nil {
		t.Fatalf("SyncProjections must create required Pi projections: %v", err)
	}
	if updated != 8 {
		t.Fatalf("SyncProjections created %d required projections, want 8", updated)
	}
	if err := CheckProjections(root, contract); err != nil {
		t.Fatalf("created Pi projections must pass drift check: %v", err)
	}
	for _, target := range contract.Projections {
		if target.Harness != "pi" {
			continue
		}
		content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(target.Path)))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.HasPrefix(content, []byte("---\nname: ")) || !bytes.Contains(content, []byte("tools:\n")) {
			t.Fatalf("Pi projection %s is not a complete generated Pi agent file", target.Path)
		}
	}
}

func TestPiProjectionRejectsWholeFileDrift(t *testing.T) {
	contract := loadRepositoryContract(t, repositoryRoot(t))
	root := t.TempDir()
	if updated, err := SyncProjections(root, contract); err != nil || updated != 8 {
		t.Fatalf("create required projections: updated=%d err=%v", updated, err)
	}

	var target ProjectionTarget
	for _, candidate := range contract.Projections {
		if candidate.Harness == "pi" && candidate.Role == contract.BaseRole.Name {
			target = candidate
			break
		}
	}
	if target.Path == "" {
		t.Fatal("missing delivery-worker Pi projection")
	}
	path := filepath.Join(root, filepath.FromSlash(target.Path))
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(content, []byte("\nhand-written footer\n")...), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := CheckProjections(root, contract); err == nil {
		t.Fatal("CheckProjections accepted hand-written content in a complete Pi projection")
	}
	updated, err := SyncProjections(root, contract)
	if err != nil {
		t.Fatal(err)
	}
	if updated != 1 {
		t.Fatalf("SyncProjections updated %d drifted Pi projection(s), want 1", updated)
	}
	if err := CheckProjections(root, contract); err != nil {
		t.Fatalf("Pi projection did not pass after sync: %v", err)
	}
}

func TestSyncPiProjectionRejectsSymlinkParent(t *testing.T) {
	contract := loadRepositoryContract(t, repositoryRoot(t))
	root := t.TempDir()
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(root, ".pi")); err != nil {
		t.Skipf("cannot create projection ancestor symlink: %v", err)
	}

	if _, err := SyncProjections(root, contract); err == nil {
		t.Fatal("SyncProjections followed a required Pi projection ancestor outside the selected root")
	}
	entries, err := os.ReadDir(outside)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("SyncProjections created Pi projection content outside the selected root: %#v", entries)
	}
}

func TestClaudeProjectionCanonicalRepeatedCRIsStable(t *testing.T) {
	contract := loadRepositoryContract(t, repositoryRoot(t))
	foundPolicy := false
	for index := range contract.HarnessPolicies {
		if contract.HarnessPolicies[index].Harness != "claude" {
			continue
		}
		contract.HarnessPolicies[index].ProjectDiscovery += "\r\r\r\nCanonical discovery continuation."
		foundPolicy = true
	}
	if !foundPolicy {
		t.Fatal("canonical contract does not define a Claude harness policy")
	}

	projection, err := RenderProjection(contract, contract.Projections[0])
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(projection, []byte("\r")) {
		t.Fatal("RenderProjection retained a carriage return from canonical Claude policy")
	}
	if !bytes.Contains(projection, []byte("\nCanonical discovery continuation.")) {
		t.Fatal("RenderProjection did not preserve canonical discovery text with LF")
	}

	root := t.TempDir()
	updated, err := SyncProjections(root, contract)
	if err != nil {
		t.Fatalf("SyncProjections creates normalized Claude projections: %v", err)
	}
	if updated != 8 {
		t.Fatalf("SyncProjections created %d normalized harness projections, want 8", updated)
	}
	if err := CheckProjections(root, contract); err != nil {
		t.Fatalf("CheckProjections rejected normalized canonical CRLF: %v", err)
	}
	updated, err = SyncProjections(root, contract)
	if err != nil {
		t.Fatalf("second SyncProjections rejected normalized canonical CRLF: %v", err)
	}
	if updated != 0 {
		t.Fatalf("second SyncProjections updated %d projections, want 0", updated)
	}
}

func TestFullProjectionIORejectsInRootSymlinks(t *testing.T) {
	contract := loadRepositoryContract(t, repositoryRoot(t))
	var target ProjectionTarget
	for _, candidate := range contract.Projections {
		if candidate.Harness == "pi" && candidate.Role == contract.BaseRole.Name {
			target = candidate
			break
		}
	}
	if target.Path == "" {
		t.Fatal("missing delivery-worker Pi projection")
	}

	tests := []struct {
		name   string
		mutate func(*testing.T, string, string)
	}{
		{
			name: "projection file",
			mutate: func(t *testing.T, _, targetPath string) {
				realPath := targetPath + ".real"
				if err := os.Rename(targetPath, realPath); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(filepath.Base(realPath), targetPath); err != nil {
					t.Skipf("cannot create projection file symlink: %v", err)
				}
			},
		},
		{
			name: "projection ancestor",
			mutate: func(t *testing.T, root, _ string) {
				piPath := filepath.Join(root, ".pi")
				realPath := filepath.Join(root, ".pi-real")
				if err := os.Rename(piPath, realPath); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(filepath.Base(realPath), piPath); err != nil {
					t.Skipf("cannot create projection ancestor symlink: %v", err)
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			if updated, err := SyncProjections(root, contract); err != nil || updated != 8 {
				t.Fatalf("create required harness projections: updated=%d err=%v", updated, err)
			}
			targetPath := filepath.Join(root, filepath.FromSlash(target.Path))
			test.mutate(t, root, targetPath)

			if err := CheckProjections(root, contract); err == nil || !strings.Contains(err.Error(), "symbolic link") {
				t.Fatalf("CheckProjections must reject the symlink, got %v", err)
			}
			if _, err := SyncProjections(root, contract); err == nil || !strings.Contains(err.Error(), "symbolic link") {
				t.Fatalf("SyncProjections must reject the symlink, got %v", err)
			}
		})
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

func TestCheckProjectionsIgnoresNestedCacheClaudeAgents(t *testing.T) {
	contract := loadRepositoryContract(t, repositoryRoot(t))
	root := t.TempDir()
	if _, err := SyncProjections(root, contract); err != nil {
		t.Fatalf("SyncProjections creates required Claude projections: %v", err)
	}
	projection, err := RenderProjection(contract, contract.Projections[0])
	if err != nil {
		t.Fatal(err)
	}
	cacheAgent := filepath.Join(root, ".cache", "preserved-baseline", ".claude", "agents", "pm-delivery-worker.md")
	if err := os.MkdirAll(filepath.Dir(cacheAgent), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cacheAgent, projection, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := CheckProjections(root, contract); err != nil {
		t.Fatalf("CheckProjections inventoried nested cache agent definitions: %v", err)
	}
}

func TestCheckProjectionsSkipsRootGitMetadata(t *testing.T) {
	contract := loadRepositoryContract(t, repositoryRoot(t))
	root := t.TempDir()
	if _, err := SyncProjections(root, contract); err != nil {
		t.Fatalf("SyncProjections creates required Claude projections: %v", err)
	}
	projection, err := RenderProjection(contract, contract.Projections[0])
	if err != nil {
		t.Fatal(err)
	}
	metadataAgent := filepath.Join(root, ".git", ".claude", "agents", "pm-delivery-worker.md")
	if err := os.MkdirAll(filepath.Dir(metadataAgent), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".git", "HEAD"), []byte("ref: refs/heads/main\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(metadataAgent, projection, 0o644); err != nil {
		t.Fatal(err)
	}

	if err := CheckProjections(root, contract); err != nil {
		t.Fatalf("CheckProjections inventoried root Git metadata: %v", err)
	}
}

func TestRequiredPiProjectionsCannotBeAbsent(t *testing.T) {
	contract := loadRepositoryContract(t, repositoryRoot(t))
	root := t.TempDir()
	updated, err := SyncProjections(root, contract)
	if err != nil {
		t.Fatal(err)
	}
	if updated != 8 {
		t.Fatalf("SyncProjections updated %d files, want eight required harness projections", updated)
	}
	for _, target := range contract.Projections {
		if target.Harness != "pi" {
			continue
		}
		if err := os.Remove(filepath.Join(root, filepath.FromSlash(target.Path))); err != nil {
			t.Fatal(err)
		}
	}
	if err := CheckProjections(root, contract); err == nil {
		t.Fatal("CheckProjections accepted a root with required Pi projections missing")
	}
}

func TestCodexWorkersCannotDelegateToAmbientAgents(t *testing.T) {
	contract := loadRepositoryContract(t, repositoryRoot(t))

	for _, target := range contract.Projections {
		if target.Harness != "codex" {
			continue
		}

		rendered, err := RenderProjection(contract, target)
		if err != nil {
			t.Fatal(err)
		}
		configuration := parseCodexProjection(t, rendered)
		for _, field := range contract.Codex.RequiredFields {
			if !configuration.IsSet(field) || strings.TrimSpace(configuration.GetString(field)) == "" {
				t.Fatalf("%s is missing required Codex field %q", target.Role, field)
			}
		}
		wantInstructions, err := renderCodexDeveloperInstructions(contract, target.Role)
		if err != nil {
			t.Fatal(err)
		}
		if got := configuration.GetString("developer_instructions"); got != wantInstructions {
			t.Fatalf("%s developer instructions changed during TOML serialization", target.Role)
		}
		for _, ambientAgent := range []string{"worker", "ambient-user-role"} {
			if codexCanDelegateToAmbientAgent(configuration, ambientAgent) {
				t.Fatalf("%s can delegate to ambient agent %q because agents.enabled is not false", target.Role, ambientAgent)
			}
		}
	}
}

func TestCodexProjectionPreservesTOMLSensitiveInstructions(t *testing.T) {
	contract := loadRepositoryContract(t, repositoryRoot(t))
	contract.Codex.CollisionBehavior = `Keep C:\work\worker and regex \d+\s.
Preserve a trailing backslash \
without TOML line folding and """ delimiters.`

	target := ProjectionTarget{
		Harness:    "codex",
		Role:       contract.BaseRole.Name,
		RenderMode: "standalone_toml",
	}
	rendered, err := RenderProjection(contract, target)
	if err != nil {
		t.Fatal(err)
	}
	want, err := renderCodexDeveloperInstructions(contract, target.Role)
	if err != nil {
		t.Fatal(err)
	}
	configuration := parseCodexProjection(t, rendered)
	if got := configuration.GetString("developer_instructions"); got != want {
		t.Fatalf("developer instructions changed during TOML serialization\ngot:  %q\nwant: %q", got, want)
	}
}

// Codex documents agents.enabled as true by default and false as disabling multi-agent tools.
// This models whether an otherwise reachable built-in or user-defined agent can be delegated to;
// it intentionally does not invoke a live model during a unit test.
func codexCanDelegateToAmbientAgent(configuration *viper.Viper, ambientAgent string) bool {
	return ambientAgent != "" && (!configuration.IsSet("agents.enabled") || configuration.GetBool("agents.enabled"))
}

func TestCodexProjectionDriftRejectsDelegationRegression(t *testing.T) {
	contract := loadRepositoryContract(t, repositoryRoot(t))
	root := t.TempDir()
	updated, err := SyncProjections(root, contract)
	if err != nil {
		t.Fatal(err)
	}
	if updated != 8 {
		t.Fatalf("SyncProjections created %d projections, want 8", updated)
	}

	var target ProjectionTarget
	for _, candidate := range contract.Projections {
		if candidate.Harness == "codex" {
			target = candidate
			break
		}
	}
	if target.Path == "" {
		t.Fatal("canonical contract does not register a Codex projection")
	}
	path := filepath.Join(root, filepath.FromSlash(target.Path))
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, bytes.Replace(content, []byte("agents.enabled = false"), []byte("agents.enabled = true"), 1), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := CheckProjections(root, contract); err == nil {
		t.Fatal("CheckProjections accepted a Codex worker with ambient delegation enabled")
	}

	updated, err = SyncProjections(root, contract)
	if err != nil {
		t.Fatal(err)
	}
	if updated != 1 {
		t.Fatalf("SyncProjections updated %d files, want 1", updated)
	}
	if err := CheckProjections(root, contract); err != nil {
		t.Fatalf("Codex worker did not pass after sync: %v", err)
	}
}

func parseCodexProjection(t *testing.T, content []byte) *viper.Viper {
	t.Helper()
	configuration := viper.New()
	configuration.SetConfigType("toml")
	if err := configuration.ReadConfig(bytes.NewReader(content)); err != nil {
		t.Fatalf("generated Codex projection is not valid TOML: %v", err)
	}
	return configuration
}

func TestProjectionIORejectsSymlinkEscape(t *testing.T) {
	contract := loadRepositoryContract(t, repositoryRoot(t))
	for _, harness := range []string{"claude", "codex", "pi"} {
		t.Run(harness, func(t *testing.T) {
			root := t.TempDir()
			outside := t.TempDir()
			original := make(map[string][]byte)

			for _, target := range contract.Projections {
				content, err := RenderProjection(contract, target)
				if err != nil {
					t.Fatal(err)
				}
				if target.Harness == harness {
					relative := strings.TrimPrefix(target.Path, "."+harness+"/")
					escapedPath := filepath.Join(outside, filepath.FromSlash(relative))
					if err := os.MkdirAll(filepath.Dir(escapedPath), 0o755); err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(escapedPath, content, 0o644); err != nil {
						t.Fatal(err)
					}
					original[escapedPath] = content
					continue
				}
				if !target.Required {
					continue
				}
				path := filepath.Join(root, filepath.FromSlash(target.Path))
				if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, content, 0o644); err != nil {
					t.Fatal(err)
				}
			}

			if err := os.Symlink(outside, filepath.Join(root, "."+harness)); err != nil {
				t.Skipf("cannot create projection ancestor symlink: %v", err)
			}
			if err := CheckProjections(root, contract); err == nil {
				t.Fatal("CheckProjections followed a projection ancestor outside the selected root")
			}
			if _, err := SyncProjections(root, contract); err == nil {
				t.Fatal("SyncProjections followed a projection ancestor outside the selected root")
			}
			for escapedPath, content := range original {
				after, err := os.ReadFile(escapedPath)
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(after, content) {
					t.Fatal("SyncProjections modified a projection outside the selected root")
				}
			}
		})
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
