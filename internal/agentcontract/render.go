package agentcontract

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

const (
	beginMarkerPrefix = "<!-- BEGIN POLYMETRICS CANONICAL AGENT CONTRACT"
	endMarker         = "<!-- END POLYMETRICS CANONICAL AGENT CONTRACT -->"
)

func RenderBlock(contract *Contract, role string) ([]byte, error) {
	if err := contract.Validate(); err != nil {
		return nil, err
	}
	if role != contract.BaseRole.Name && role != contract.ConnectorOverlay.Name {
		return nil, fmt.Errorf("canonical contract: unknown role %q", role)
	}

	var output bytes.Buffer
	fmt.Fprintf(&output, "%s role=%s version=%s; DO NOT EDIT -->\n", beginMarkerPrefix, role, contract.ContractVersion)
	fmt.Fprintf(&output, "# %s\n\n%s\n\n", role, roleSummary(contract, role))
	fmt.Fprintf(&output, "Canonical source: `%s` (schema %d, contract %s). %s\n\n", contract.Ownership.SourcePath, contract.SchemaVersion, contract.ContractVersion, contract.Ownership.UpdateRule)
	fmt.Fprintf(&output, "## Ownership and handoff\n\n%s Exactly %d worker is active; delegation is `%s`. Never spawn %s. Durable handoff state is carried only by %s.\n\n", contract.Ownership.Owner, contract.BaseRole.MaxActiveWorkers, contract.BaseRole.Delegation, joinNatural(contract.BaseRole.ForbiddenRoles), joinNatural(contract.BaseRole.DurableHandoff))
	fmt.Fprintf(&output, "## Canonical state machine: %s\n\n", contract.StateMachine.Name)
	for index, step := range contract.StateMachine.Steps {
		fmt.Fprintf(&output, "%d. `%s` — %s\n", index+1, step.ID, step.Instruction)
	}
	if role == contract.ConnectorOverlay.Name {
		fmt.Fprintf(&output, "\n## Connector overlay\n\n%s It inherits every base state and wraps `%s` with these ordered gates:\n\n", contract.ConnectorOverlay.Summary, contract.ConnectorOverlay.WrapsState)
		for index, step := range contract.ConnectorOverlay.Steps {
			fmt.Fprintf(&output, "%d. `%s` — %s\n", index+1, step.ID, step.Instruction)
		}
	}

	fmt.Fprintf(&output, "\n## Tracker and pull-request topology\n\n- Parent seed: %s\n- Sub-PR topology: %s\n- Integrate a child only when:\n", contract.Tracker.ParentSeed, contract.Tracker.SubPRBase)
	writeBullets(&output, contract.Tracker.IntegrateWhen)
	fmt.Fprintln(&output, "- Mark the parent ready only when:")
	writeBullets(&output, contract.Tracker.ReadyWhen)
	fmt.Fprintf(&output, "- Final merge: %s\n\n", contract.Tracker.FinalMerge)
	fmt.Fprintln(&output, "## Installed GSD lifecycle")
	fmt.Fprintln(&output)
	for _, invocation := range contract.GSD.Sequence {
		argv, err := marshalArgv(invocation.Argv)
		if err != nil {
			return nil, err
		}
		fmt.Fprintf(&output, "- argv `%s` — %s\n", argv, invocation.Purpose)
	}
	fmt.Fprintf(&output, "- GSD ship exclusion: %s\n\n", contract.GSD.ShipExclusion)

	fmt.Fprintln(&output, "## no-mistakes topology")
	fmt.Fprintln(&output)
	fmt.Fprintf(&output, "- Verified installed version: `%s`. Never use `%s`; it auto-resolves captain-owned ask-user gates.\n", contract.NoMistakes.VerifiedVersion, strings.Join(contract.NoMistakes.ForbiddenFlags, "` or `"))
	childArgv, err := marshalArgv(contract.NoMistakes.ChildCommand.Argv)
	if err != nil {
		return nil, err
	}
	subPRArgv, err := marshalArgv(contract.NoMistakes.SubPROpen.Argv)
	if err != nil {
		return nil, err
	}
	parentArgv, err := marshalArgv(contract.NoMistakes.ParentCommand.Argv)
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(&output, "- Child branch argv: `%s`. %s %s\n", childArgv, contract.NoMistakes.ChildCommand.Instruction, contract.NoMistakes.GateResponse)
	fmt.Fprintf(&output, "- Child PR argv: `%s`. %s\n", subPRArgv, contract.NoMistakes.SubPROpen.Instruction)
	fmt.Fprintf(&output, "- Integrated parent argv: `%s`. %s\n\n", parentArgv, contract.NoMistakes.ParentCommand.Instruction)

	fmt.Fprintf(&output, "## Away-mode authority\n\n%s\n\nSelf-answer only when:\n", contract.Authority.Principle)
	writeRules(&output, contract.Authority.SelfAnswerWhen)
	fmt.Fprintln(&output, "Auto-fix only when:")
	writeRules(&output, contract.Authority.AutoFixWhen)
	fmt.Fprintln(&output, "Pause and preserve state when:")
	writeRules(&output, contract.Authority.PauseWhen)
	fmt.Fprintln(&output, "Invariants:")
	writeRules(&output, contract.Authority.Invariants)

	fmt.Fprintln(&output, "## Wayfinder disposition")
	fmt.Fprintln(&output)
	fmt.Fprintf(&output, "Wayfinder is `%s` and is not a dependency. Borrow only %s. Rejection rationale: %s Do not install or rediscover it for this flow.\n\n", contract.Wayfinder.Disposition, joinNatural(contract.Wayfinder.Borrowed), strings.Join(contract.Wayfinder.Rationale, " "))
	fmt.Fprintln(&output, endMarker)
	return output.Bytes(), nil
}

// RenderProjection renders either a harness-owned canonical block or a complete generated file,
// according to the registered projection target.
func RenderProjection(contract *Contract, target ProjectionTarget) ([]byte, error) {
	if err := contract.Validate(); err != nil {
		return nil, err
	}

	switch target.RenderMode {
	case "markdown_block":
		return RenderBlock(contract, target.Role)
	case "standalone_toml":
		if target.Harness != "codex" {
			return nil, fmt.Errorf("canonical contract: standalone_toml render mode requires the Codex harness")
		}
		return renderCodexProjection(contract, target.Role)
	case claudeMarkdownYAMLFrontmatter:
		if target.Harness != "claude" {
			return nil, fmt.Errorf("canonical contract: markdown_yaml_frontmatter render mode requires the Claude harness")
		}
		policy, ok := contract.ProjectionFor(target.Harness)
		if !ok {
			return nil, fmt.Errorf("canonical contract: Claude harness policy is missing")
		}
		return renderClaudeProjection(contract, target, policy)
	case opencodeMarkdownYAMLFrontmatter:
		if target.Harness != "opencode" {
			return nil, fmt.Errorf("canonical contract: OpenCode frontmatter render mode requires the OpenCode harness")
		}
		return renderOpenCodeProjection(contract, target)
	case "full":
		if target.Harness != "pi" {
			return nil, fmt.Errorf("canonical contract: full render mode requires the Pi harness")
		}
		return renderPiProjection(contract, target)
	default:
		return nil, fmt.Errorf("canonical contract: unknown projection render mode %q", target.RenderMode)
	}
}

func renderCodexProjection(contract *Contract, role string) ([]byte, error) {
	instructions, err := renderCodexDeveloperInstructions(contract, role)
	if err != nil {
		return nil, err
	}
	fields, err := toml.Marshal(struct {
		Name                  string `toml:"name"`
		Description           string `toml:"description"`
		DeveloperInstructions string `toml:"developer_instructions,multiline"`
	}{
		Name:                  role,
		Description:           roleSummary(contract, role),
		DeveloperInstructions: instructions,
	})
	if err != nil {
		return nil, fmt.Errorf("canonical contract: encode Codex projection: %w", err)
	}

	var output bytes.Buffer
	fmt.Fprintln(&output, "# Code generated by agentcontractgen; DO NOT EDIT.")
	fmt.Fprintf(&output, "# Canonical source: %s\n\n", contract.Ownership.SourcePath)
	output.Write(fields)
	fmt.Fprintf(&output, "%s = %t\n", contract.Codex.ToolAccess.Setting, *contract.Codex.ToolAccess.Value)
	return output.Bytes(), nil
}

func renderCodexDeveloperInstructions(contract *Contract, role string) (string, error) {
	base, err := RenderBlock(contract, role)
	if err != nil {
		return "", err
	}

	var output strings.Builder
	output.Write(base)
	fmt.Fprintf(&output, "\n## Codex project-local configuration\n\n- Format and discovery: %s\n- Tool access: %s\n- Project trust: %s\n- General precedence: %s\n- Filename collisions: %s\n- Official sources: %s and %s\n", contract.Codex.Discovery, contract.Codex.ToolAccess.Effect, contract.Codex.ProjectTrustRequirement, contract.Codex.ConfigurationPrecedence, contract.Codex.CollisionBehavior, contract.Codex.Documentation.Subagents, contract.Codex.Documentation.ConfigBasics)
	return output.String(), nil
}

func renderPiProjection(contract *Contract, target ProjectionTarget) ([]byte, error) {
	block, err := RenderBlock(contract, target.Role)
	if err != nil {
		return nil, err
	}

	var output bytes.Buffer
	fmt.Fprintln(&output, "---")
	fmt.Fprintf(&output, "name: %s\n", yamlString(target.Role))
	fmt.Fprintf(&output, "description: %s\n", yamlString(roleSummary(contract, target.Role)))
	fmt.Fprintln(&output, "tools:")
	for _, tool := range contract.PiHarness.ChildTools {
		fmt.Fprintf(&output, "  - %s\n", yamlString(tool))
	}
	fmt.Fprintln(&output, "---")
	fmt.Fprintln(&output)
	output.Write(block)
	return output.Bytes(), nil
}

func yamlString(value string) string {
	return strconv.Quote(value)
}

func roleSummary(contract *Contract, role string) string {
	if role == contract.ConnectorOverlay.Name {
		return contract.ConnectorOverlay.Summary
	}
	return contract.BaseRole.Summary
}

func marshalArgv(argv []string) (string, error) {
	var output bytes.Buffer
	encoder := json.NewEncoder(&output)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(argv); err != nil {
		return "", fmt.Errorf("canonical contract: encode argv: %w", err)
	}
	return strings.TrimSuffix(output.String(), "\n"), nil
}

func writeBullets(output *bytes.Buffer, values []string) {
	for _, value := range values {
		fmt.Fprintf(output, "  - %s\n", value)
	}
}

func writeRules(output *bytes.Buffer, rules []Rule) {
	for _, rule := range rules {
		fmt.Fprintf(output, "- `%s` — %s\n", rule.ID, rule.Instruction)
	}
	fmt.Fprintln(output)
}

func joinNatural(values []string) string {
	switch len(values) {
	case 0:
		return "nothing"
	case 1:
		return values[0]
	case 2:
		return values[0] + " and " + values[1]
	default:
		return strings.Join(values[:len(values)-1], ", ") + ", and " + values[len(values)-1]
	}
}
