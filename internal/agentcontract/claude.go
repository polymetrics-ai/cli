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
	Name            string   `yaml:"name"`
	Description     string   `yaml:"description"`
	Tools           []string `yaml:"tools"`
	Skills          []string `yaml:"skills"`
	DisallowedTools []string `yaml:"disallowedTools"`
	PermissionMode  string   `yaml:"permissionMode"`
}

func renderClaudeProjection(contract *Contract, target ProjectionTarget, policy HarnessPolicy) ([]byte, error) {
	if target.Harness != "claude" {
		return nil, fmt.Errorf("canonical contract: %s cannot use Claude frontmatter", target.Harness)
	}
	frontmatter := claudeFrontmatter{
		Name:            target.Role,
		Description:     roleSummary(contract, target.Role),
		Tools:           slices.Clone(policy.Tools),
		Skills:          slices.Clone(policy.PreloadedSkills),
		DisallowedTools: policy.DisallowedTools,
		PermissionMode:  policy.PermissionMode,
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
	fmt.Fprintf(&output, "Skill behavior: %s\n\n", policy.SkillsDocumentationURL)
	fmt.Fprintf(&output, "Skill boundary: %s\n\n", policy.SkillBoundary)
	fmt.Fprintf(&output, "Trusted preloaded skills: %s.\n\n", joinInlineCode(policy.PreloadedSkills))
	fmt.Fprintf(&output, "Unavailable repository-routed skills: %s. Cost: %s\n\n", joinInlineCode(policy.UnavailableSkills), policy.UnavailableSkillCost)
	fmt.Fprintf(&output, "Isolation: %s\n\n", policy.DelegationGuarantee)
	fmt.Fprintf(&output, "Required clean-home smoke (not generation evidence): %s\n\n", strings.ReplaceAll(policy.SmokeProcedure, "<role>", target.Role))
	output.Write(block)

	result := normalizeClaudeProjection(output.Bytes())
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
	content = normalizeClaudeProjection(content)
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

func normalizeClaudeProjection(content []byte) []byte {
	if !bytes.Contains(content, []byte("\r\n")) {
		return content
	}

	normalized := make([]byte, 0, len(content))
	for index := 0; index < len(content); {
		if content[index] != '\r' {
			normalized = append(normalized, content[index])
			index++
			continue
		}

		runStart := index
		for index < len(content) && content[index] == '\r' {
			index++
		}
		if index < len(content) && content[index] == '\n' {
			normalized = append(normalized, '\n')
			index++
			continue
		}
		normalized = append(normalized, content[runStart:index]...)
	}
	return normalized
}

func validateClaudeFrontmatter(frontmatter claudeFrontmatter, target ProjectionTarget, policy HarnessPolicy) error {
	if frontmatter.Name != target.Role || strings.TrimSpace(frontmatter.Description) == "" ||
		!slices.Equal(frontmatter.Tools, policy.Tools) ||
		!slices.Equal(frontmatter.Skills, policy.PreloadedSkills) ||
		!slices.Equal(frontmatter.DisallowedTools, policy.DisallowedTools) ||
		frontmatter.PermissionMode != policy.PermissionMode {
		return fmt.Errorf("claude projection frontmatter does not match the canonical %s policy", target.Harness)
	}
	if slices.Contains(frontmatter.Tools, policy.DelegationTool) || slices.Contains(frontmatter.Tools, policy.SkillTool) ||
		!slices.Contains(frontmatter.DisallowedTools, policy.DelegationTool) ||
		!slices.Contains(frontmatter.DisallowedTools, policy.SkillTool) {
		return fmt.Errorf("claude projection frontmatter must preload trusted skills and deny Agent, Task, and Skill")
	}
	return nil
}

func joinInlineCode(values []string) string {
	formatted := make([]string, len(values))
	for index, value := range values {
		formatted[index] = "`" + value + "`"
	}
	return joinNatural(formatted)
}
