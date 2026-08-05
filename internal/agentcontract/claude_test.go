package agentcontract

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

type claudeTestFrontmatter struct {
	Name           string   `yaml:"name"`
	Description    string   `yaml:"description"`
	Tools          []string `yaml:"tools"`
	PermissionMode string   `yaml:"permissionMode"`
}

func TestClaudeProjectWorkersBlockAmbientAgentDelegation(t *testing.T) {
	root := repositoryRoot(t)
	contract := loadRepositoryContract(t, root)
	wantTools := []string{"Bash", "Edit", "Glob", "Grep", "Read", "Write"}

	for _, target := range contract.Projections {
		if target.Harness != "claude" {
			continue
		}
		t.Run(target.Role, func(t *testing.T) {
			content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(target.Path)))
			if os.IsNotExist(err) {
				t.Fatalf("%s can resolve to an ambient same-name agent: missing project projection %s leaves no project-local tools allowlist that omits Agent", target.Role, target.Path)
			}
			if err != nil {
				t.Fatal(err)
			}

			frontmatter, err := parseClaudeTestFrontmatter(content)
			if err != nil {
				t.Fatal(err)
			}
			if frontmatter.Name != target.Role || strings.TrimSpace(frontmatter.Description) == "" {
				t.Fatalf("frontmatter = %#v, want required name %q and a description", frontmatter, target.Role)
			}
			if !slices.Equal(frontmatter.Tools, wantTools) {
				t.Fatalf("%s tools = %#v, want minimum allowlist %#v", target.Role, frontmatter.Tools, wantTools)
			}
			if frontmatter.PermissionMode != "default" {
				t.Fatalf("%s permissionMode = %q, want default", target.Role, frontmatter.PermissionMode)
			}
			if slices.Contains(frontmatter.Tools, "Agent") {
				t.Fatalf("%s can delegate to an ambient agent because its tools allowlist grants Agent", target.Role)
			}
		})
	}
}

func parseClaudeTestFrontmatter(content []byte) (claudeTestFrontmatter, error) {
	var frontmatter claudeTestFrontmatter
	if !bytes.HasPrefix(content, []byte("---\n")) {
		return frontmatter, os.ErrInvalid
	}
	end := bytes.Index(content[len("---\n"):], []byte("\n---\n"))
	if end < 0 {
		return frontmatter, os.ErrInvalid
	}
	end += len("---\n")
	if err := yaml.Unmarshal(content[len("---\n"):end], &frontmatter); err != nil {
		return frontmatter, err
	}
	return frontmatter, nil
}
