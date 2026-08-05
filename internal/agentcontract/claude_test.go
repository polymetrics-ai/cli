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
	policy, ok := contract.HarnessPolicyFor("claude")
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
			if !strings.Contains(string(first), policy.DelegationGuarantee) || !strings.Contains(string(first), smoke) {
				t.Fatalf("%s projection omitted canonical Claude isolation or smoke guidance", target.Role)
			}
		})
	}
}

func TestClaudeProjectWorkersBlockAmbientAgentDelegation(t *testing.T) {
	root := repositoryRoot(t)
	contract := loadRepositoryContract(t, root)
	policy, ok := contract.HarnessPolicyFor("claude")
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
			if !slices.Equal(frontmatter.Tools, policy.Tools) {
				t.Fatalf("%s tools = %#v, want minimum allowlist %#v", target.Role, frontmatter.Tools, policy.Tools)
			}
			if frontmatter.PermissionMode != policy.PermissionMode {
				t.Fatalf("%s permissionMode = %q, want %q", target.Role, frontmatter.PermissionMode, policy.PermissionMode)
			}
			if slices.Contains(frontmatter.Tools, policy.DelegationTool) {
				t.Fatalf("%s can delegate to an ambient agent because its tools allowlist grants Agent", target.Role)
			}
			if !strings.Contains(string(content), policy.DocumentationURL) {
				t.Fatalf("%s does not document the official Claude source", target.Path)
			}
		})
	}
}
