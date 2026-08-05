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
				!strings.Contains(string(first), policy.SkillsDocumentationURL) ||
				!strings.Contains(string(first), smoke) {
				t.Fatalf("%s projection omitted canonical Claude skill, isolation, or smoke guidance", target.Role)
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
			wantTools := claudeProjectionTools(policy)
			if !slices.Equal(frontmatter.Tools, wantTools) {
				t.Fatalf("%s tools = %#v, want scoped allowlist %#v", target.Role, frontmatter.Tools, wantTools)
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
				t.Fatalf("%s grants bare Skill instead of scoped required skills", target.Role)
			}
			for _, skill := range policy.ReachableSkills {
				rule := policy.SkillTool + "(" + skill + ")"
				if !slices.Contains(frontmatter.Tools, rule) {
					t.Fatalf("%s cannot reach required skill %q", target.Role, skill)
				}
			}
			for _, forbidden := range []string{"Skill(gsd-programming-loop)", "Skill(batch)", "Skill(code-review)"} {
				if slices.Contains(frontmatter.Tools, forbidden) {
					t.Fatalf("%s can invoke agent-oriented skill rule %q", target.Role, forbidden)
				}
			}
			if !strings.Contains(string(content), policy.DocumentationURL) ||
				!strings.Contains(string(content), policy.SkillsDocumentationURL) {
				t.Fatalf("%s does not document the official Claude agent and skill sources", target.Path)
			}
		})
	}
}
