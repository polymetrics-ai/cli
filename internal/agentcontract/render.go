package agentcontract

import (
	"bytes"
	"fmt"
	"strings"
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
		fmt.Fprintf(&output, "- `scripts/gsd prompt %s %s` — %s\n", invocation.Command, invocation.Args, invocation.Purpose)
	}
	fmt.Fprintf(&output, "- GSD ship exclusion: %s\n\n", contract.GSD.ShipExclusion)

	fmt.Fprintln(&output, "## no-mistakes topology")
	fmt.Fprintln(&output)
	fmt.Fprintf(&output, "- Verified installed version: `%s`. Never use `%s`; it auto-resolves captain-owned ask-user gates.\n", contract.NoMistakes.VerifiedVersion, strings.Join(contract.NoMistakes.ForbiddenFlags, "` or `"))
	fmt.Fprintf(&output, "- Child branch: `%s`. %s\n", contract.NoMistakes.ChildCommand, contract.NoMistakes.GateResponse)
	fmt.Fprintf(&output, "- Child PR: %s\n", contract.NoMistakes.SubPROpen)
	fmt.Fprintf(&output, "- Integrated parent: %s\n\n", contract.NoMistakes.ParentCommand)

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

func roleSummary(contract *Contract, role string) string {
	if role == contract.ConnectorOverlay.Name {
		return contract.ConnectorOverlay.Summary
	}
	return contract.BaseRole.Summary
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
