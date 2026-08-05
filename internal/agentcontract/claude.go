package agentcontract

import (
	"bytes"
	"fmt"
	"io"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

const claudeMarkdownYAMLFrontmatter = "markdown_yaml_frontmatter"

type claudeFrontmatter struct {
	Name           string   `yaml:"name"`
	Description    string   `yaml:"description"`
	Tools          []string `yaml:"tools"`
	PermissionMode string   `yaml:"permissionMode"`
}

// RenderProjection renders a registered harness-native projection. Harnesses without a native
// policy retain the generated-block rendering used by the optional future projections.
func RenderProjection(contract *Contract, target ProjectionTarget) ([]byte, error) {
	if err := contract.Validate(); err != nil {
		return nil, err
	}
	policy, ok := contract.HarnessPolicyFor(target.Harness)
	if !ok {
		return RenderBlock(contract, target.Role)
	}
	switch policy.Format {
	case claudeMarkdownYAMLFrontmatter:
		return renderClaudeProjection(contract, target, policy)
	default:
		return nil, fmt.Errorf("canonical contract: unsupported %s projection format %q", target.Harness, policy.Format)
	}
}

func projectionRendersWholeFile(contract *Contract, target ProjectionTarget) bool {
	policy, ok := contract.HarnessPolicyFor(target.Harness)
	return ok && policy.Format == claudeMarkdownYAMLFrontmatter
}

func renderClaudeProjection(contract *Contract, target ProjectionTarget, policy HarnessPolicy) ([]byte, error) {
	if target.Harness != "claude" {
		return nil, fmt.Errorf("canonical contract: %s cannot use Claude frontmatter", target.Harness)
	}
	frontmatter := claudeFrontmatter{
		Name:           target.Role,
		Description:    roleSummary(contract, target.Role),
		Tools:          policy.Tools,
		PermissionMode: policy.PermissionMode,
	}
	encodedFrontmatter, err := yaml.Marshal(frontmatter)
	if err != nil {
		return nil, fmt.Errorf("canonical contract: encode Claude frontmatter: %w", err)
	}
	block, err := RenderBlock(contract, target.Role)
	if err != nil {
		return nil, err
	}

	var output bytes.Buffer
	output.WriteString("---\n")
	output.Write(encodedFrontmatter)
	output.WriteString("---\n\n")
	fmt.Fprintln(&output, "## Claude Code projection")
	fmt.Fprintln(&output)
	fmt.Fprintf(&output, "Official behavior: %s\n\n", policy.DocumentationURL)
	fmt.Fprintf(&output, "Discovery: %s\n\n", policy.ProjectDiscovery)
	fmt.Fprintf(&output, "Precedence (highest first): %s. Managed definitions and CLI `--agents` remain higher-precedence caveats.\n\n", joinNatural(policy.Precedence))
	fmt.Fprintf(&output, "Isolation: %s\n\n", policy.DelegationGuarantee)
	fmt.Fprintf(&output, "Trusted-project smoke: %s\n\n", strings.ReplaceAll(policy.SmokeProcedure, "<role>", target.Role))
	output.Write(block)

	result := output.Bytes()
	parsed, err := parseClaudeFrontmatter(result)
	if err != nil {
		return nil, err
	}
	if err := validateClaudeFrontmatter(parsed, target, policy); err != nil {
		return nil, err
	}
	return result, nil
}

func parseClaudeFrontmatter(content []byte) (claudeFrontmatter, error) {
	var frontmatter claudeFrontmatter
	const delimiter = "---\n"
	if !bytes.HasPrefix(content, []byte(delimiter)) {
		return frontmatter, fmt.Errorf("claude projection frontmatter start marker is missing")
	}
	end := bytes.Index(content[len(delimiter):], []byte("\n---\n"))
	if end < 0 {
		return frontmatter, fmt.Errorf("claude projection frontmatter end marker is missing")
	}
	end += len(delimiter)
	decoder := yaml.NewDecoder(bytes.NewReader(content[len(delimiter):end]))
	decoder.KnownFields(true)
	if err := decoder.Decode(&frontmatter); err != nil {
		return frontmatter, fmt.Errorf("decode Claude projection frontmatter: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return frontmatter, fmt.Errorf("decode Claude projection frontmatter: multiple YAML documents")
		}
		return frontmatter, fmt.Errorf("decode Claude projection frontmatter: %w", err)
	}
	return frontmatter, nil
}

func validateClaudeFrontmatter(frontmatter claudeFrontmatter, target ProjectionTarget, policy HarnessPolicy) error {
	if frontmatter.Name != target.Role || strings.TrimSpace(frontmatter.Description) == "" ||
		!slices.Equal(frontmatter.Tools, policy.Tools) || frontmatter.PermissionMode != policy.PermissionMode {
		return fmt.Errorf("claude projection frontmatter does not match the canonical %s policy", target.Harness)
	}
	if slices.Contains(frontmatter.Tools, policy.DelegationTool) {
		return fmt.Errorf("claude projection frontmatter must omit %s from its tools allowlist", policy.DelegationTool)
	}
	return nil
}
