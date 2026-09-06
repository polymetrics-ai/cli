package agentcontract

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"gopkg.in/yaml.v3"
)

const opencodeMarkdownYAMLFrontmatter = "opencode_markdown_yaml_frontmatter"

type openCodeFrontmatter struct {
	Description string            `yaml:"description"`
	Mode        string            `yaml:"mode"`
	Permission  map[string]string `yaml:"permission"`
}

func renderOpenCodeProjection(contract *Contract, target ProjectionTarget) ([]byte, error) {
	if target.Harness != "opencode" {
		return nil, fmt.Errorf("canonical contract: %s cannot use OpenCode frontmatter", target.Harness)
	}
	block, err := RenderBlock(contract, target.Role)
	if err != nil {
		return nil, err
	}

	var output bytes.Buffer
	fmt.Fprintln(&output, "---")
	fmt.Fprintf(&output, "description: %s\n", yamlString(roleSummary(contract, target.Role)))
	fmt.Fprintf(&output, "mode: %s\n", yamlString(contract.OpenCode.Mode))
	fmt.Fprintln(&output, "permission:")
	for _, permission := range contract.OpenCode.Permissions {
		fmt.Fprintf(&output, "  %s: %s\n", permission.Tool, permission.Access)
	}
	fmt.Fprintln(&output, "---")
	fmt.Fprintln(&output)
	fmt.Fprintln(&output, "## OpenCode projection")
	fmt.Fprintln(&output)
	fmt.Fprintf(&output, "Official behavior: %s\n\n", contract.OpenCode.DocumentationURL)
	fmt.Fprintf(&output, "Discovery: %s\n\n", contract.OpenCode.Discovery)
	fmt.Fprintf(&output, "Isolation: %s\n\n", contract.OpenCode.DelegationGuarantee)
	output.Write(block)

	result := normalizeClaudeProjection(output.Bytes())
	frontmatter, err := parseOpenCodeFrontmatter(result)
	if err != nil {
		return nil, err
	}
	if err := validateOpenCodeFrontmatter(frontmatter, target, contract.OpenCode); err != nil {
		return nil, err
	}
	return result, nil
}

func parseOpenCodeFrontmatter(content []byte) (openCodeFrontmatter, error) {
	var frontmatter openCodeFrontmatter
	content = normalizeClaudeProjection(content)
	const delimiter = "---\n"
	if !bytes.HasPrefix(content, []byte(delimiter)) {
		return frontmatter, fmt.Errorf("OpenCode projection frontmatter start marker is missing")
	}
	end := bytes.Index(content[len(delimiter):], []byte("\n---\n"))
	if end < 0 {
		return frontmatter, fmt.Errorf("OpenCode projection frontmatter end marker is missing")
	}
	end += len(delimiter)
	decoder := yaml.NewDecoder(bytes.NewReader(content[len(delimiter):end]))
	decoder.KnownFields(true)
	if err := decoder.Decode(&frontmatter); err != nil {
		return frontmatter, fmt.Errorf("decode OpenCode projection frontmatter: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return frontmatter, fmt.Errorf("decode OpenCode projection frontmatter: multiple YAML documents")
		}
		return frontmatter, fmt.Errorf("decode OpenCode projection frontmatter: %w", err)
	}
	return frontmatter, nil
}

func validateOpenCodeFrontmatter(frontmatter openCodeFrontmatter, target ProjectionTarget, policy OpenCodeContract) error {
	wantPermissions := make(map[string]string, len(policy.Permissions))
	for _, permission := range policy.Permissions {
		wantPermissions[permission.Tool] = permission.Access
	}
	if target.Harness != "opencode" || strings.TrimSpace(frontmatter.Description) == "" || frontmatter.Mode != policy.Mode ||
		!mapsEqual(frontmatter.Permission, wantPermissions) {
		return fmt.Errorf("OpenCode projection frontmatter does not match the canonical policy")
	}
	if frontmatter.Permission["task"] != "deny" || frontmatter.Permission["skill"] != "deny" {
		return fmt.Errorf("OpenCode projection frontmatter must deny task and skill")
	}
	return nil
}

func mapsEqual(left, right map[string]string) bool {
	if len(left) != len(right) {
		return false
	}
	for key, value := range left {
		if right[key] != value {
			return false
		}
	}
	return true
}
