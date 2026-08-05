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

type Contract struct {
	SchemaVersion    int                `json:"schema_version"`
	ContractVersion  string             `json:"contract_version"`
	SourceID         string             `json:"source_id"`
	Ownership        Ownership          `json:"ownership"`
	BaseRole         Role               `json:"base_role"`
	HarnessPolicies  []HarnessPolicy    `json:"harness_policies"`
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

// HarnessPolicy contains the native settings required to render a registered harness projection.
type HarnessPolicy struct {
	Harness                string   `json:"harness"`
	Format                 string   `json:"format"`
	DocumentationURL       string   `json:"documentation_url"`
	ProjectDiscovery       string   `json:"project_discovery"`
	Precedence             []string `json:"precedence"`
	Tools                  []string `json:"tools"`
	SkillTool              string   `json:"skill_tool"`
	ReachableSkills        []string `json:"reachable_skills"`
	SkillsDocumentationURL string   `json:"skills_documentation_url"`
	SkillBoundary          string   `json:"skill_boundary"`
	DisallowedTools        []string `json:"disallowed_tools"`
	PermissionMode         string   `json:"permission_mode"`
	DelegationTool         string   `json:"delegation_tool"`
	DelegationGuarantee    string   `json:"delegation_guarantee"`
	SmokeProcedure         string   `json:"smoke_procedure"`
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
	Argv    []string `json:"argv"`
	Purpose string   `json:"purpose"`
}

type ExecutableAction struct {
	Argv        []string `json:"argv"`
	Instruction string   `json:"instruction"`
}

type NoMistakesContract struct {
	VerifiedVersion string           `json:"verified_version"`
	ForbiddenFlags  []string         `json:"forbidden_flags"`
	ChildCommand    ExecutableAction `json:"child_command"`
	GateResponse    string           `json:"gate_response"`
	SubPROpen       ExecutableAction `json:"sub_pr_open"`
	ParentCommand   ExecutableAction `json:"parent_command"`
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
	if err := validateHarnessPolicies(contract.HarnessPolicies); err != nil {
		return err
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
	if contract.NoMistakes.VerifiedVersion != "v1.41.2" || !slices.Contains(contract.NoMistakes.ForbiddenFlags, "--yes") || strings.TrimSpace(contract.NoMistakes.GateResponse) == "" {
		return fmt.Errorf("canonical contract: no-mistakes version, gate response, and forbidden flags are required")
	}
	for _, forbidden := range contract.NoMistakes.ForbiddenFlags {
		if strings.TrimSpace(forbidden) == "" {
			return fmt.Errorf("canonical contract: no-mistakes forbidden flags must be non-empty")
		}
	}
	if err := validateExecutableFields(contract.executableFields(), contract.NoMistakes.ForbiddenFlags); err != nil {
		return err
	}
	requiredGSDSequence := [...][]string{
		{"scripts/gsd", "prompt", "discuss-phase", "<phase>"},
		{"scripts/gsd", "prompt", "plan-phase", "<phase>", "--tdd"},
		{"scripts/gsd", "prompt", "execute-phase", "<phase>"},
		{"scripts/gsd", "prompt", "verify-work", "<phase>"},
		{"scripts/gsd", "prompt", "plan-phase", "<phase>", "--gaps"},
		{"scripts/gsd", "prompt", "execute-phase", "<phase>", "--gaps-only"},
		{"scripts/gsd", "prompt", "code-review", "<phase>"},
	}
	if len(contract.GSD.Sequence) != len(requiredGSDSequence) {
		return fmt.Errorf("canonical contract: GSD sequence is incomplete or out of order")
	}
	for index, want := range requiredGSDSequence {
		invocation := contract.GSD.Sequence[index]
		if !slices.Equal(invocation.Argv, want) || strings.TrimSpace(invocation.Purpose) == "" {
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
	requiredChildCommand := []string{"no-mistakes", "axi", "run", "--intent", "<issue-intent>", "--skip=push,pr,ci"}
	requiredSubPROpen := []string{"gh-axi", "pr", "create", "--base", "<parent-branch>", "--head", "<child-branch>", "--title", "<conventional-title>", "--body-file", "<pr-body-file>"}
	requiredParentCommand := []string{"no-mistakes", "axi", "run", "--intent", "<parent-intent>"}
	if !slices.Equal(contract.NoMistakes.ChildCommand.Argv, requiredChildCommand) ||
		!slices.Equal(contract.NoMistakes.SubPROpen.Argv, requiredSubPROpen) ||
		!slices.Equal(contract.NoMistakes.ParentCommand.Argv, requiredParentCommand) ||
		strings.TrimSpace(contract.NoMistakes.ChildCommand.Instruction) == "" ||
		strings.TrimSpace(contract.NoMistakes.SubPROpen.Instruction) == "" ||
		strings.TrimSpace(contract.NoMistakes.ParentCommand.Instruction) == "" {
		return fmt.Errorf("canonical contract: no-mistakes child workaround, full parent pipeline, and gh-axi sub-PR action are required")
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

// ProjectionFor returns the canonical native projection policy for a harness.
func (contract *Contract) ProjectionFor(harness string) (HarnessPolicy, bool) {
	for _, policy := range contract.HarnessPolicies {
		if policy.Harness == harness {
			return policy, true
		}
	}
	return HarnessPolicy{}, false
}

func validateHarnessPolicies(policies []HarnessPolicy) error {
	seen := make(map[string]bool, len(policies))
	var claude *HarnessPolicy
	for index := range policies {
		policy := &policies[index]
		if strings.TrimSpace(policy.Harness) == "" || strings.TrimSpace(policy.Format) == "" || seen[policy.Harness] {
			return fmt.Errorf("canonical contract: harness policies must have unique non-empty harness and format fields")
		}
		seen[policy.Harness] = true
		if policy.Harness == "claude" {
			claude = policy
		}
	}
	if claude == nil {
		return fmt.Errorf("canonical contract: Claude harness policy is required")
	}

	const claudeFormat = "markdown_yaml_frontmatter"
	const claudeDocumentationURL = "https://code.claude.com/docs/en/sub-agents"
	const claudeSkillsDocumentationURL = "https://code.claude.com/docs/en/slash-commands"
	if claude.Format != claudeFormat || claude.DocumentationURL != claudeDocumentationURL ||
		claude.SkillsDocumentationURL != claudeSkillsDocumentationURL ||
		strings.TrimSpace(claude.ProjectDiscovery) == "" || strings.TrimSpace(claude.DelegationGuarantee) == "" ||
		strings.TrimSpace(claude.SkillBoundary) == "" || !strings.Contains(claude.SkillBoundary, "context: fork") ||
		strings.TrimSpace(claude.SmokeProcedure) == "" || !strings.Contains(claude.SmokeProcedure, "<role>") ||
		claude.PermissionMode != "default" || claude.DelegationTool != "Agent" || claude.SkillTool != "Skill" {
		return fmt.Errorf("canonical contract: Claude format, documentation, discovery, delegation, permission, and smoke policy are required")
	}
	wantPrecedence := []string{"managed definitions", "CLI --agents", "project .claude/agents", "user ~/.claude/agents", "plugins"}
	if !slices.Equal(claude.Precedence, wantPrecedence) {
		return fmt.Errorf("canonical contract: Claude precedence must match the documented managed, CLI, project, user, plugin order")
	}
	wantTools := []string{"Bash", "Edit", "Glob", "Grep", "Read", "Write"}
	wantDisallowedTools := []string{"Agent", "Task"}
	if !slices.Equal(claude.Tools, wantTools) || !slices.Equal(claude.ReachableSkills, claudeReachableSkills()) ||
		!slices.Equal(claude.DisallowedTools, wantDisallowedTools) ||
		slices.Contains(claude.Tools, claude.DelegationTool) || slices.Contains(claude.Tools, claude.SkillTool) {
		return fmt.Errorf("canonical contract: Claude tools must use the minimal base allowlist, scoped required skills, and Agent/Task denylist")
	}
	return nil
}

func claudeReachableSkills() []string {
	return []string{
		"golang-cli",
		"golang-concurrency",
		"golang-context",
		"golang-database",
		"golang-design-patterns",
		"golang-documentation",
		"golang-error-handling",
		"golang-graphql",
		"golang-how-to",
		"golang-lint",
		"golang-safety",
		"golang-security",
		"golang-spf13-cobra",
		"golang-spf13-viper",
		"golang-structs-interfaces",
		"golang-testing",
		"cc-skills-golang:golang-cli",
		"cc-skills-golang:golang-concurrency",
		"cc-skills-golang:golang-context",
		"cc-skills-golang:golang-database",
		"cc-skills-golang:golang-design-patterns",
		"cc-skills-golang:golang-documentation",
		"cc-skills-golang:golang-error-handling",
		"cc-skills-golang:golang-graphql",
		"cc-skills-golang:golang-how-to",
		"cc-skills-golang:golang-lint",
		"cc-skills-golang:golang-safety",
		"cc-skills-golang:golang-security",
		"cc-skills-golang:golang-spf13-cobra",
		"cc-skills-golang:golang-spf13-viper",
		"cc-skills-golang:golang-structs-interfaces",
		"cc-skills-golang:golang-testing",
		"frontend-design",
		"frontend-design:frontend-design",
		"vercel-composition-patterns",
		"vercel-react-best-practices",
		"web-design-guidelines",
	}
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

type executableField struct {
	name string
	argv []string
}

func (contract *Contract) executableFields() []executableField {
	fields := make([]executableField, 0, len(contract.GSD.Sequence)+3)
	for index, invocation := range contract.GSD.Sequence {
		fields = append(fields, executableField{
			name: fmt.Sprintf("gsd.sequence[%d]", index),
			argv: invocation.Argv,
		})
	}
	return append(fields,
		executableField{name: "no_mistakes.child_command", argv: contract.NoMistakes.ChildCommand.Argv},
		executableField{name: "no_mistakes.sub_pr_open", argv: contract.NoMistakes.SubPROpen.Argv},
		executableField{name: "no_mistakes.parent_command", argv: contract.NoMistakes.ParentCommand.Argv},
	)
}

func validateExecutableFields(fields []executableField, forbiddenFlags []string) error {
	for _, field := range fields {
		if len(field.argv) == 0 {
			return fmt.Errorf("canonical contract: executable field %s requires argv", field.name)
		}
		for index, argument := range field.argv {
			if strings.TrimSpace(argument) == "" || strings.IndexFunc(argument, unicode.IsControl) >= 0 {
				return fmt.Errorf("canonical contract: executable field %s has an invalid argv[%d]", field.name, index)
			}
			if strings.ContainsAny(argument, "'\"\\`$;&|()") {
				return fmt.Errorf("canonical contract: executable field %s argv[%d] contains shell syntax", field.name, index)
			}
			if argument == "ship" {
				return fmt.Errorf("canonical contract: executable field %s uses prohibited GSD action ship", field.name)
			}
			for _, forbidden := range forbiddenFlags {
				if argument == forbidden || strings.HasPrefix(argument, forbidden+"=") {
					return fmt.Errorf("canonical contract: executable field %s uses forbidden flag %q", field.name, forbidden)
				}
			}
		}
	}
	return nil
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
	type expectedProjection struct {
		path     string
		required bool
	}
	expected := map[string]expectedProjection{
		"claude/pm-delivery-worker":  {path: ".claude/agents/pm-delivery-worker.md", required: true},
		"claude/pm-connector-worker": {path: ".claude/agents/pm-connector-worker.md", required: true},
		"codex/pm-delivery-worker":   {path: ".codex/agents/pm-delivery-worker.toml", required: false},
		"codex/pm-connector-worker":  {path: ".codex/agents/pm-connector-worker.toml", required: false},
		"pi/pm-delivery-worker":      {path: ".pi/agents/pm-delivery-worker.md", required: false},
		"pi/pm-connector-worker":     {path: ".pi/agents/pm-connector-worker.md", required: false},
	}
	if len(targets) != len(expected) {
		return fmt.Errorf("canonical contract: exactly six harness projection targets are required")
	}
	seen := make(map[string]bool, len(targets))
	for _, target := range targets {
		key := target.Harness + "/" + target.Role
		want, ok := expected[key]
		if !ok || target.Path != want.path || target.Required != want.required || filepath.Clean(target.Path) != target.Path || filepath.IsAbs(target.Path) || seen[key] {
			return fmt.Errorf("canonical contract: invalid projection target %q at %q", key, target.Path)
		}
		seen[key] = true
	}
	return nil
}
