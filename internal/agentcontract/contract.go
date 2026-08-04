// Package agentcontract validates and projects the canonical agent delivery contract.
package agentcontract

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"unicode"
)

const SourcePath = ".agents/agentic-delivery/canonical/delivery-contract.json"

const requiredChildCommand = "no-mistakes axi run --intent '<issue-intent>' --skip=push,pr,ci"

type Contract struct {
	SchemaVersion    int                `json:"schema_version"`
	ContractVersion  string             `json:"contract_version"`
	SourceID         string             `json:"source_id"`
	Ownership        Ownership          `json:"ownership"`
	BaseRole         Role               `json:"base_role"`
	StateMachine     StateMachine       `json:"state_machine"`
	ConnectorOverlay ConnectorOverlay   `json:"connector_overlay"`
	Tracker          TrackerContract    `json:"tracker"`
	GSD              GSDContract        `json:"gsd"`
	NoMistakes       NoMistakesContract `json:"no_mistakes"`
	Authority        AuthorityContract  `json:"authority"`
	Wayfinder        WayfinderDecision  `json:"wayfinder"`
	Projections      []ProjectionTarget `json:"projections"`
}

type Ownership struct {
	SourcePath string `json:"source_path"`
	Owner      string `json:"owner"`
	UpdateRule string `json:"update_rule"`
}

type Role struct {
	Name             string   `json:"name"`
	Summary          string   `json:"summary"`
	MaxActiveWorkers int      `json:"max_active_workers"`
	Delegation       string   `json:"delegation"`
	ForbiddenRoles   []string `json:"forbidden_roles"`
	DurableHandoff   []string `json:"durable_handoff"`
}

type StateMachine struct {
	Name         string `json:"name"`
	InitialState string `json:"initial_state"`
	Steps        []Step `json:"steps"`
}

type Step struct {
	ID          string `json:"id"`
	Instruction string `json:"instruction"`
}

type ConnectorOverlay struct {
	Name           string `json:"name"`
	Inherits       string `json:"inherits"`
	Summary        string `json:"summary"`
	CompletionWave int    `json:"completion_wave"`
	WrapsState     string `json:"wraps_state"`
	Steps          []Step `json:"steps"`
}

type TrackerContract struct {
	DraftParentBeforeProduction bool     `json:"draft_parent_before_production"`
	ParentSeed                  string   `json:"parent_seed"`
	SubPRBase                   string   `json:"sub_pr_base"`
	IntegrateWhen               []string `json:"integrate_when"`
	ReadyWhen                   []string `json:"ready_when"`
	FinalMerge                  string   `json:"final_merge"`
}

type GSDContract struct {
	Commands      []string        `json:"commands"`
	Sequence      []GSDInvocation `json:"sequence"`
	ShipUsed      bool            `json:"ship_used"`
	ShipExclusion string          `json:"ship_exclusion"`
}

type GSDInvocation struct {
	Command string `json:"command"`
	Args    string `json:"args"`
	Purpose string `json:"purpose"`
}

type NoMistakesContract struct {
	VerifiedVersion string   `json:"verified_version"`
	ForbiddenFlags  []string `json:"forbidden_flags"`
	ChildCommand    string   `json:"child_command"`
	GateResponse    string   `json:"gate_response"`
	SubPROpen       string   `json:"sub_pr_open"`
	ParentCommand   string   `json:"parent_command"`
}

type AuthorityContract struct {
	Principle      string `json:"principle"`
	SelfAnswerWhen []Rule `json:"self_answer_when"`
	AutoFixWhen    []Rule `json:"auto_fix_when"`
	PauseWhen      []Rule `json:"pause_when"`
	Invariants     []Rule `json:"invariants"`
}

type Rule struct {
	ID          string `json:"id"`
	Instruction string `json:"instruction"`
}

type WayfinderDecision struct {
	Disposition string   `json:"disposition"`
	Dependency  bool     `json:"dependency"`
	Borrowed    []string `json:"borrowed"`
	Rationale   []string `json:"rationale"`
}

type ProjectionTarget struct {
	Harness  string `json:"harness"`
	Role     string `json:"role"`
	Path     string `json:"path"`
	Required bool   `json:"required"`
}

func Load(path string) (*Contract, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open canonical contract: %w", err)
	}
	defer func() { _ = file.Close() }()

	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	var contract Contract
	if err := decoder.Decode(&contract); err != nil {
		return nil, fmt.Errorf("decode canonical contract: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return nil, fmt.Errorf("decode canonical contract: trailing JSON value")
	}
	if err := contract.Validate(); err != nil {
		return nil, err
	}
	return &contract, nil
}

func (contract *Contract) Validate() error {
	if contract.SchemaVersion != 1 {
		return fmt.Errorf("canonical contract: schema_version must be 1")
	}
	if strings.TrimSpace(contract.ContractVersion) == "" || strings.TrimSpace(contract.SourceID) == "" {
		return fmt.Errorf("canonical contract: contract_version and source_id are required")
	}
	if contract.Ownership.SourcePath != SourcePath || strings.TrimSpace(contract.Ownership.Owner) == "" || strings.TrimSpace(contract.Ownership.UpdateRule) == "" {
		return fmt.Errorf("canonical contract: ownership source_path, owner, and update_rule are required")
	}
	if contract.BaseRole.Name != "pm-delivery-worker" || contract.BaseRole.MaxActiveWorkers != 1 || contract.BaseRole.Delegation != "none" {
		return fmt.Errorf("canonical contract: base role must be pm-delivery-worker with one active worker and no delegation")
	}
	if strings.TrimSpace(contract.BaseRole.Summary) == "" || len(contract.BaseRole.DurableHandoff) != 4 || !allNonEmpty(contract.BaseRole.DurableHandoff) {
		return fmt.Errorf("canonical contract: base role summary and durable handoff are required")
	}
	for _, forbidden := range []string{"orchestrator", "shepherd", "planner", "reviewer", "verifier", "GSD role"} {
		if !slices.Contains(contract.BaseRole.ForbiddenRoles, forbidden) {
			return fmt.Errorf("canonical contract: base role must forbid %q", forbidden)
		}
	}
	stateIDs := []string{"job_received", "issue_map", "parent_draft_pr", "map_wave_phase", "discuss_decisions", "plan_tdd", "execute_tdd", "verify_gaps", "review_no_mistakes", "open_sub_pr", "integrate_sub_pr", "integrated_parent_gates", "ready_parent", "captain_merge"}
	if contract.StateMachine.Name != "issue-first-delivery" || contract.StateMachine.InitialState != stateIDs[0] || !equalStepIDs(contract.StateMachine.Steps, stateIDs) {
		return fmt.Errorf("canonical contract: base state machine is incomplete or out of order")
	}
	connectorIDs := []string{"source_policy_map", "bundle_operation_plan", "replay_conformance", "implementation_slices", "runtime_surface_gates", "website_data_refresh"}
	if contract.ConnectorOverlay.Name != "pm-connector-worker" || contract.ConnectorOverlay.Inherits != contract.BaseRole.Name || strings.TrimSpace(contract.ConnectorOverlay.Summary) == "" || contract.ConnectorOverlay.CompletionWave != 5 || contract.ConnectorOverlay.WrapsState != "execute_tdd" || !equalStepIDs(contract.ConnectorOverlay.Steps, connectorIDs) {
		return fmt.Errorf("canonical contract: connector role must inherit the base and wrap execute_tdd with the required ordered gates")
	}
	commands := []string{"discuss-phase", "plan-phase", "execute-phase", "verify-work", "code-review", "ship"}
	if !slices.Equal(contract.GSD.Commands, commands) || contract.GSD.ShipUsed || strings.TrimSpace(contract.GSD.ShipExclusion) == "" {
		return fmt.Errorf("canonical contract: GSD lifecycle or ship exclusion is invalid")
	}
	requiredGSDSequence := [...]GSDInvocation{
		{Command: "discuss-phase", Args: "<phase>"},
		{Command: "plan-phase", Args: "<phase> --tdd"},
		{Command: "execute-phase", Args: "<phase>"},
		{Command: "verify-work", Args: "<phase>"},
		{Command: "plan-phase", Args: "<phase> --gaps"},
		{Command: "execute-phase", Args: "<phase> --gaps-only"},
		{Command: "code-review", Args: "<phase>"},
	}
	if len(contract.GSD.Sequence) != len(requiredGSDSequence) {
		return fmt.Errorf("canonical contract: GSD sequence is incomplete or out of order")
	}
	for index, want := range requiredGSDSequence {
		invocation := contract.GSD.Sequence[index]
		if invocation.Command != want.Command || invocation.Args != want.Args || strings.TrimSpace(invocation.Purpose) == "" {
			return fmt.Errorf("canonical contract: GSD sequence is incomplete or out of order at step %d", index+1)
		}
	}
	if !contract.Tracker.DraftParentBeforeProduction || strings.TrimSpace(contract.Tracker.ParentSeed) == "" || strings.TrimSpace(contract.Tracker.SubPRBase) == "" || !allNonEmpty(contract.Tracker.IntegrateWhen) || !allNonEmpty(contract.Tracker.ReadyWhen) || strings.TrimSpace(contract.Tracker.FinalMerge) == "" {
		return fmt.Errorf("canonical contract: tracker draft, integration, readiness, and merge gates are required")
	}
	if !slices.Contains(contract.Tracker.IntegrateWhen, "CI checks pass") {
		return fmt.Errorf("canonical contract: child integration requires passing CI checks")
	}
	for _, criterion := range contract.Tracker.IntegrateWhen {
		if strings.Contains(strings.ToLower(criterion), "infrastructure blocker") {
			return fmt.Errorf("canonical contract: infrastructure blockers must pause child integration")
		}
	}
	if contract.NoMistakes.VerifiedVersion != "v1.41.2" || !slices.Contains(contract.NoMistakes.ForbiddenFlags, "--yes") || contract.NoMistakes.ChildCommand != requiredChildCommand || strings.TrimSpace(contract.NoMistakes.GateResponse) == "" || strings.Contains(contract.NoMistakes.ParentCommand, "--skip=") || strings.TrimSpace(contract.NoMistakes.ParentCommand) == "" || !strings.Contains(contract.NoMistakes.SubPROpen, "gh-axi") {
		return fmt.Errorf("canonical contract: no-mistakes child workaround, full parent pipeline, gh-axi, and --yes prohibition are required")
	}
	for _, forbidden := range contract.NoMistakes.ForbiddenFlags {
		if strings.TrimSpace(forbidden) == "" {
			return fmt.Errorf("canonical contract: no-mistakes forbidden flags must be non-empty")
		}
		if containsCommandToken(contract.NoMistakes.ChildCommand, forbidden) || containsCommandToken(contract.NoMistakes.ParentCommand, forbidden) {
			return fmt.Errorf("canonical contract: no-mistakes command uses forbidden flag %q", forbidden)
		}
	}
	pauseIDs := []string{"product_ambiguity", "destructive_irreversible", "secrets_auth_security", "dependency_production", "generic_write", "reverse_etl_execute", "quality_gate_weakening", "final_merge"}
	if contract.Authority.Principle != "Absence never expands authority." || !equalRuleIDs(contract.Authority.PauseWhen, pauseIDs) {
		return fmt.Errorf("canonical contract: away-mode authority principle or pause boundary is incomplete")
	}
	if !equalRuleIDs(contract.Authority.SelfAnswerWhen, []string{"contract_fixed"}) || !equalRuleIDs(contract.Authority.AutoFixWhen, []string{"bounded_finding"}) {
		return fmt.Errorf("canonical contract: away-mode self-answer or bounded auto-fix rule is incomplete")
	}
	if !equalRuleIDs(contract.Authority.Invariants, []string{"absence_no_authority", "never_merge_red", "parent_merge_captain"}) {
		return fmt.Errorf("canonical contract: merge and absence invariants are incomplete")
	}
	if contract.Wayfinder.Disposition != "rejected" || contract.Wayfinder.Dependency || len(contract.Wayfinder.Borrowed) != 3 || !allNonEmpty(contract.Wayfinder.Borrowed) || !allNonEmpty(contract.Wayfinder.Rationale) {
		return fmt.Errorf("canonical contract: Wayfinder rejection and borrowed ideas are required")
	}
	return validateProjections(contract.Projections)
}

func allNonEmpty(values []string) bool {
	if len(values) == 0 {
		return false
	}
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			return false
		}
	}
	return true
}

func containsCommandToken(command, target string) bool {
	tokens := strings.FieldsFunc(command, func(character rune) bool {
		return unicode.IsSpace(character) || strings.ContainsRune("'\"`;|&()<>\\", character)
	})
	return slices.Contains(tokens, target)
}

func equalStepIDs(steps []Step, want []string) bool {
	if len(steps) != len(want) {
		return false
	}
	for index, step := range steps {
		if step.ID != want[index] || strings.TrimSpace(step.Instruction) == "" {
			return false
		}
	}
	return true
}

func equalRuleIDs(rules []Rule, want []string) bool {
	if len(rules) != len(want) {
		return false
	}
	for index, rule := range rules {
		if rule.ID != want[index] || strings.TrimSpace(rule.Instruction) == "" {
			return false
		}
	}
	return true
}

func validateProjections(targets []ProjectionTarget) error {
	expected := map[string]string{
		"claude/pm-delivery-worker":  ".claude/agents/pm-delivery-worker.md",
		"claude/pm-connector-worker": ".claude/agents/pm-connector-worker.md",
		"codex/pm-delivery-worker":   ".codex/agents/pm-delivery-worker.toml",
		"codex/pm-connector-worker":  ".codex/agents/pm-connector-worker.toml",
		"pi/pm-delivery-worker":      ".pi/agents/pm-delivery-worker.md",
		"pi/pm-connector-worker":     ".pi/agents/pm-connector-worker.md",
	}
	if len(targets) != len(expected) {
		return fmt.Errorf("canonical contract: exactly six harness projection targets are required")
	}
	seen := make(map[string]bool, len(targets))
	for _, target := range targets {
		key := target.Harness + "/" + target.Role
		wantPath, ok := expected[key]
		if !ok || target.Path != wantPath || filepath.Clean(target.Path) != target.Path || filepath.IsAbs(target.Path) || seen[key] {
			return fmt.Errorf("canonical contract: invalid projection target %q at %q", key, target.Path)
		}
		seen[key] = true
	}
	return nil
}
