package agentcontract

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"testing"
)

func TestCanonicalContractRequiredInvariants(t *testing.T) {
	root := repositoryRoot(t)

	tests := []struct {
		name   string
		mutate func(*Contract)
	}{
		{
			name: "more than one worker",
			mutate: func(value *Contract) {
				value.BaseRole.MaxActiveWorkers = 2
			},
		},
		{
			name: "delegation allowed",
			mutate: func(value *Contract) {
				value.BaseRole.Delegation = "allowed"
			},
		},
		{
			name: "Claude tools include Agent",
			mutate: func(value *Contract) {
				value.HarnessPolicies[0].Tools = append(value.HarnessPolicies[0].Tools, "Agent")
			},
		},
		{
			name: "Claude tools include bare Skill",
			mutate: func(value *Contract) {
				value.HarnessPolicies[0].Tools = append(value.HarnessPolicies[0].Tools, "Skill")
			},
		},
		{
			name: "Claude trusted preload removed",
			mutate: func(value *Contract) {
				policy := &value.HarnessPolicies[0]
				policy.PreloadedSkills = policy.PreloadedSkills[:len(policy.PreloadedSkills)-1]
			},
		},
		{
			name: "Claude unqualified skill added",
			mutate: func(value *Contract) {
				policy := &value.HarnessPolicies[0]
				policy.PreloadedSkills = append(policy.PreloadedSkills, "gsd-programming-loop")
			},
		},
		{
			name: "Claude Task alias not denied",
			mutate: func(value *Contract) {
				value.HarnessPolicies[0].DisallowedTools = []string{"Agent", "Skill"}
			},
		},
		{
			name: "Claude Skill tool not denied",
			mutate: func(value *Contract) {
				value.HarnessPolicies[0].DisallowedTools = []string{"Agent", "Task"}
			},
		},
		{
			name: "Claude unavailable skill cost removed",
			mutate: func(value *Contract) {
				value.HarnessPolicies[0].UnavailableSkillCost = ""
			},
		},
		{
			name: "Claude precedence skips project",
			mutate: func(value *Contract) {
				value.HarnessPolicies[0].Precedence[2] = "user ~/.claude/agents"
			},
		},
		{
			name: "Claude project projection optional",
			mutate: func(value *Contract) {
				value.Projections[0].Required = false
			},
		},
		{
			name: "connector does not inherit base",
			mutate: func(value *Contract) {
				value.ConnectorOverlay.Inherits = "other-role"
			},
		},
		{
			name: "Codex delegation enabled",
			mutate: func(value *Contract) {
				enabled := true
				value.Codex.ToolAccess.Value = &enabled
			},
		},
		{
			name: "Codex required field missing",
			mutate: func(value *Contract) {
				value.Codex.RequiredFields = value.Codex.RequiredFields[:2]
			},
		},
		{
			name: "Codex collision guidance missing",
			mutate: func(value *Contract) {
				value.Codex.CollisionBehavior = ""
			},
		},
		{
			name: "Codex render mode removed",
			mutate: func(value *Contract) {
				for index := range value.Projections {
					if value.Projections[index].Harness == "codex" {
						value.Projections[index].RenderMode = "markdown_block"
					}
				}
			},
		},
		{
			name: "yes allowed",
			mutate: func(value *Contract) {
				value.NoMistakes.ForbiddenFlags = nil
			},
		},
		{
			name: "GSD ship used",
			mutate: func(value *Contract) {
				value.GSD.ShipUsed = true
			},
		},
		{
			name: "GSD sequence reordered",
			mutate: func(value *Contract) {
				value.GSD.Sequence[0], value.GSD.Sequence[1] = value.GSD.Sequence[1], value.GSD.Sequence[0]
			},
		},
		{
			name: "GSD TDD flag removed",
			mutate: func(value *Contract) {
				value.GSD.Sequence[1].Argv = []string{"scripts/gsd", "prompt", "plan-phase", "<phase>"}
			},
		},
		{
			name: "GSD ship executable",
			mutate: func(value *Contract) {
				value.GSD.Sequence[len(value.GSD.Sequence)-1].Argv[2] = "ship"
			},
		},
		{
			name: "GSD ship executable through parent field",
			mutate: func(value *Contract) {
				value.NoMistakes.ParentCommand.Argv = []string{"scripts/gsd", "prompt", "ship", "<phase>"}
			},
		},
		{
			name: "yes used by GSD command",
			mutate: func(value *Contract) {
				value.GSD.Sequence[0].Argv = append(value.GSD.Sequence[0].Argv, "--yes")
			},
		},
		{
			name: "yes used by child command",
			mutate: func(value *Contract) {
				value.NoMistakes.ChildCommand.Argv = append(value.NoMistakes.ChildCommand.Argv, "--yes")
			},
		},
		{
			name: "yes used by sub-PR command",
			mutate: func(value *Contract) {
				value.NoMistakes.SubPROpen.Argv = append(value.NoMistakes.SubPROpen.Argv, "--yes")
			},
		},
		{
			name: "yes used by parent command",
			mutate: func(value *Contract) {
				value.NoMistakes.ParentCommand.Argv = append(value.NoMistakes.ParentCommand.Argv, "--yes")
			},
		},
		{
			name: "shell-quoted yes used by GSD command",
			mutate: func(value *Contract) {
				value.GSD.Sequence[0].Argv = append(value.GSD.Sequence[0].Argv, "--'yes'")
			},
		},
		{
			name: "shell-quoted yes used by child command",
			mutate: func(value *Contract) {
				value.NoMistakes.ChildCommand.Argv = append(value.NoMistakes.ChildCommand.Argv, "--'yes'")
			},
		},
		{
			name: "shell-quoted yes used by sub-PR command",
			mutate: func(value *Contract) {
				value.NoMistakes.SubPROpen.Argv = append(value.NoMistakes.SubPROpen.Argv, "--'yes'")
			},
		},
		{
			name: "shell-quoted yes used by parent command",
			mutate: func(value *Contract) {
				value.NoMistakes.ParentCommand.Argv = append(value.NoMistakes.ParentCommand.Argv, "--'yes'")
			},
		},
		{
			name: "child intent split across argv",
			mutate: func(value *Contract) {
				value.NoMistakes.ChildCommand.Argv = []string{"no-mistakes", "axi", "run", "--intent", "<issue", "intent>", "--skip=push,pr,ci"}
			},
		},
		{
			name: "infrastructure blocker substitutes for CI",
			mutate: func(value *Contract) {
				value.Tracker.IntegrateWhen[1] = "CI checks pass or an infrastructure blocker is recorded"
			},
		},
		{
			name: "missing parent pipeline",
			mutate: func(value *Contract) {
				value.NoMistakes.ParentCommand.Argv = nil
			},
		},
		{
			name: "missing durable handoff",
			mutate: func(value *Contract) {
				value.BaseRole.DurableHandoff = nil
			},
		},
		{
			name: "authority pause removed",
			mutate: func(value *Contract) {
				value.Authority.PauseWhen = value.Authority.PauseWhen[:len(value.Authority.PauseWhen)-1]
			},
		},
		{
			name: "Pi clean project scope changed",
			mutate: func(value *Contract) {
				value.PiHarness.CleanProjectScope = "project"
			},
		},
		{
			name: "Pi additional role allowed",
			mutate: func(value *Contract) {
				value.PiHarness.Roles = append(value.PiHarness.Roles, "ambient-worker")
			},
		},
		{
			name: "Pi child subagent tool allowed",
			mutate: func(value *Contract) {
				value.PiHarness.ChildTools = append(value.PiHarness.ChildTools, "subagent")
			},
		},
		{
			name: "OpenCode task permission allowed",
			mutate: func(value *Contract) {
				for index := range value.OpenCode.Permissions {
					if value.OpenCode.Permissions[index].Tool == "task" {
						value.OpenCode.Permissions[index].Access = "allow"
					}
				}
			},
		},
		{
			name: "OpenCode projection missing",
			mutate: func(value *Contract) {
				value.Projections = value.Projections[:len(value.Projections)-1]
			},
		},
		{
			name: "certification gate does not protect accepted transition",
			mutate: func(value *Contract) {
				value.CertificationGate.EnforcedTransitions = []string{"integrate_sub_pr", "ready_parent", "human_ready"}
			},
		},
		{
			name: "certification gate executable argv changed",
			mutate: func(value *Contract) {
				value.CertificationGate.Command.Argv = []string{"go", "run", "./cmd/agentcontractgen", "check"}
			},
		},
		{
			name: "tracker omits certification gate",
			mutate: func(value *Contract) {
				value.Tracker.IntegrateWhen = value.Tracker.IntegrateWhen[:len(value.Tracker.IntegrateWhen)-1]
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			contract := loadRepositoryContract(t, root)
			test.mutate(contract)
			if err := contract.Validate(); err == nil {
				t.Fatal("Validate accepted a contract missing a required invariant")
			}
		})
	}
}

func TestMarshalArgvPreservesIntentAsOneArgument(t *testing.T) {
	intent := "Complete Wave 5's overlay; $(unexpected)"
	argv := []string{"no-mistakes", "axi", "run", "--intent", intent, "--skip=push,pr,ci"}

	rendered, err := marshalArgv(argv)
	if err != nil {
		t.Fatal(err)
	}
	var decoded []string
	if err := json.Unmarshal([]byte(rendered), &decoded); err != nil {
		t.Fatalf("rendered argv is not valid JSON: %v", err)
	}
	if !slices.Equal(decoded, argv) || len(decoded) != 6 || decoded[4] != intent {
		t.Fatalf("rendered argv did not preserve the complete intent as one argument: %#v", decoded)
	}
}

func TestRenderIsStableAndConnectorInheritsBase(t *testing.T) {
	contract := loadRepositoryContract(t, repositoryRoot(t))
	base, err := RenderBlock(contract, contract.BaseRole.Name)
	if err != nil {
		t.Fatal(err)
	}
	baseAgain, err := RenderBlock(contract, contract.BaseRole.Name)
	if err != nil {
		t.Fatal(err)
	}
	if string(base) != string(baseAgain) {
		t.Fatal("RenderBlock output is not deterministic")
	}

	connector, err := RenderBlock(contract, contract.ConnectorOverlay.Name)
	if err != nil {
		t.Fatal(err)
	}
	for _, step := range contract.StateMachine.Steps {
		if !strings.Contains(string(base), "`"+step.ID+"`") || !strings.Contains(string(connector), "`"+step.ID+"`") {
			t.Fatalf("base state %q missing from one or both role renderings", step.ID)
		}
	}
	for _, step := range contract.ConnectorOverlay.Steps {
		if !strings.Contains(string(connector), "`"+step.ID+"`") {
			t.Fatalf("connector overlay state %q missing from connector rendering", step.ID)
		}
	}

	const expectedSHA256 = "6de2132cd2d935ef9cca0473b4577fb43ef419abc6f58d325b60beb0633ace85"
	gotSHA256 := fmt.Sprintf("%x", sha256.Sum256(base))
	if gotSHA256 != expectedSHA256 {
		t.Fatalf("base rendering hash = %s, update expected hash after intentional canonical change", gotSHA256)
	}
	const expectedConnectorSHA256 = "454f0437cdb516a22df2170bf5b30fa1c5e45596eba74e53aa9c51f16c563fb3"
	gotConnectorSHA256 := fmt.Sprintf("%x", sha256.Sum256(connector))
	if gotConnectorSHA256 != expectedConnectorSHA256 {
		t.Fatalf("connector rendering hash = %s, update expected hash after intentional canonical change", gotConnectorSHA256)
	}
}
