package agentcontract

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestRenderClaudeProjectionsIsStableAndSelfContained(t *testing.T) {
	contract := loadRepositoryContract(t, repositoryRoot(t))
	policy, ok := contract.ProjectionFor("claude")
	if !ok {
		t.Fatal("canonical contract does not define a Claude harness policy")
	}

	for _, target := range contract.Projections {
		if target.Harness != "claude" {
			continue
		}
		t.Run(target.Role, func(t *testing.T) {
			first, err := RenderProjection(contract, target)
			if err != nil {
				t.Fatal(err)
			}
			second, err := RenderProjection(contract, target)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(first, second) {
				t.Fatalf("%s projection rendering is not deterministic", target.Role)
			}
			frontmatter, err := parseClaudeFrontmatter(first)
			if err != nil {
				t.Fatal(err)
			}
			if err := validateClaudeFrontmatter(frontmatter, target, policy); err != nil {
				t.Fatal(err)
			}
			smoke := strings.ReplaceAll(policy.SmokeProcedure, "<role>", target.Role)
			if !strings.Contains(string(first), policy.DelegationGuarantee) ||
				!strings.Contains(string(first), policy.SkillBoundary) ||
				!strings.Contains(string(first), policy.UnavailableSkillCost) ||
				!strings.Contains(string(first), policy.SkillsDocumentationURL) ||
				!strings.Contains(string(first), smoke) {
				t.Fatalf("%s projection omitted canonical Claude skill, isolation, or smoke guidance", target.Role)
			}
		})
	}
}

func TestParseClaudeFrontmatterAcceptsCRLF(t *testing.T) {
	contract := loadRepositoryContract(t, repositoryRoot(t))
	target := contract.Projections[0]
	policy, ok := contract.ProjectionFor(target.Harness)
	if !ok {
		t.Fatalf("canonical contract does not define a %s harness policy", target.Harness)
	}
	projection, err := RenderProjection(contract, target)
	if err != nil {
		t.Fatal(err)
	}
	projection = bytes.ReplaceAll(projection, []byte("\n"), []byte("\r\n"))

	frontmatter, err := parseClaudeFrontmatter(projection)
	if err != nil {
		t.Fatalf("parseClaudeFrontmatter rejected CRLF projection: %v", err)
	}
	if err := validateClaudeFrontmatter(frontmatter, target, policy); err != nil {
		t.Fatalf("validateClaudeFrontmatter rejected CRLF projection: %v", err)
	}
}

func TestNormalizeClaudeProjectionIsIdempotent(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "LF unchanged", input: "first\nsecond\n", want: "first\nsecond\n"},
		{name: "CRLF", input: "first\r\nsecond\r\n", want: "first\nsecond\n"},
		{name: "repeated carriage returns", input: "first\r\r\r\nsecond", want: "first\nsecond"},
		{name: "multiple runs", input: "\r\r\nfirst\r\nsecond\r\r\r\n", want: "\nfirst\nsecond\n"},
		{name: "bare carriage returns preserved", input: "first\rsecond\r\rthird", want: "first\rsecond\r\rthird"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			once := normalizeClaudeProjection([]byte(test.input))
			if string(once) != test.want {
				t.Fatalf("normalizeClaudeProjection(%q) = %q, want %q", test.input, once, test.want)
			}
			twice := normalizeClaudeProjection(once)
			if !bytes.Equal(twice, once) {
				t.Fatalf("second normalization = %q, want fixed point %q", twice, once)
			}
		})
	}
}

func TestClaudeProjectWorkersBlockAmbientAgentDelegation(t *testing.T) {
	root := repositoryRoot(t)
	contract := loadRepositoryContract(t, root)
	policy, ok := contract.ProjectionFor("claude")
	if !ok {
		t.Fatal("canonical contract does not define a Claude harness policy")
	}

	for _, target := range contract.Projections {
		if target.Harness != "claude" {
			continue
		}
		t.Run(target.Role, func(t *testing.T) {
			if !target.Required {
				t.Fatalf("%s must be a required project projection", target.Path)
			}
			content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(target.Path)))
			if os.IsNotExist(err) {
				t.Fatalf("%s can resolve to an ambient same-name agent: missing project projection %s leaves no project-local tools allowlist that omits Agent", target.Role, target.Path)
			}
			if err != nil {
				t.Fatal(err)
			}

			frontmatter, err := parseClaudeFrontmatter(content)
			if err != nil {
				t.Fatal(err)
			}
			if frontmatter.Name != target.Role || strings.TrimSpace(frontmatter.Description) == "" {
				t.Fatalf("frontmatter = %#v, want required name %q and a description", frontmatter, target.Role)
			}
			wantTools := policy.Tools
			if !slices.Equal(frontmatter.Tools, wantTools) {
				t.Fatalf("%s tools = %#v, want base allowlist %#v", target.Role, frontmatter.Tools, wantTools)
			}
			if !slices.Equal(frontmatter.Skills, policy.PreloadedSkills) {
				t.Fatalf("%s skills = %#v, want trusted preloads %#v", target.Role, frontmatter.Skills, policy.PreloadedSkills)
			}
			if !slices.Equal(frontmatter.DisallowedTools, policy.DisallowedTools) {
				t.Fatalf("%s disallowedTools = %#v, want %#v", target.Role, frontmatter.DisallowedTools, policy.DisallowedTools)
			}
			if frontmatter.PermissionMode != policy.PermissionMode {
				t.Fatalf("%s permissionMode = %q, want %q", target.Role, frontmatter.PermissionMode, policy.PermissionMode)
			}
			if slices.Contains(frontmatter.Tools, policy.DelegationTool) {
				t.Fatalf("%s can delegate to an ambient agent because its tools allowlist grants Agent", target.Role)
			}
			if slices.Contains(frontmatter.Tools, policy.SkillTool) {
				t.Fatalf("%s grants runtime Skill access instead of trusted preloads", target.Role)
			}
			for _, tool := range frontmatter.Tools {
				if strings.HasPrefix(tool, policy.SkillTool+"(") {
					t.Fatalf("%s uses undocumented scoped Skill tool entry %q", target.Role, tool)
				}
			}
			for _, skill := range policy.PreloadedSkills {
				if !strings.Contains(skill, ":") {
					t.Fatalf("%s preloads collision-prone unqualified skill %q", target.Role, skill)
				}
			}
			for _, unavailable := range policy.UnavailableSkills {
				if slices.Contains(frontmatter.Skills, unavailable) {
					t.Fatalf("%s preloads unqualified unavailable skill %q", target.Role, unavailable)
				}
				if !strings.Contains(string(content), "`"+unavailable+"`") {
					t.Fatalf("%s does not document unavailable skill %q", target.Role, unavailable)
				}
			}
			if !strings.Contains(string(content), policy.DocumentationURL) ||
				!strings.Contains(string(content), policy.SkillsDocumentationURL) {
				t.Fatalf("%s does not document the official Claude agent and skill sources", target.Path)
			}
		})
	}
}
