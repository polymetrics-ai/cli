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
			name: "Claude required skill removed",
			mutate: func(value *Contract) {
				policy := &value.HarnessPolicies[0]
				policy.ReachableSkills = policy.ReachableSkills[:len(policy.ReachableSkills)-1]
			},
		},
		{
			name: "Claude fork-capable skill added",
			mutate: func(value *Contract) {
				policy := &value.HarnessPolicies[0]
				policy.ReachableSkills = append(policy.ReachableSkills, "gsd-programming-loop")
			},
		},
		{
			name: "Claude Task alias not denied",
			mutate: func(value *Contract) {
				value.HarnessPolicies[0].DisallowedTools = []string{"Agent"}
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

	const expectedSHA256 = "cd548735a93c2c0fa5113082255ad07f5fa39af27a033b07797f2f370a2999b1"
	gotSHA256 := fmt.Sprintf("%x", sha256.Sum256(base))
	if gotSHA256 != expectedSHA256 {
		t.Fatalf("base rendering hash = %s, update expected hash after intentional canonical change", gotSHA256)
	}
	const expectedConnectorSHA256 = "a0b5986925a716685b028b3617ba21b59216980e3f5f1fd5a323be75fd9b54c8"
	gotConnectorSHA256 := fmt.Sprintf("%x", sha256.Sum256(connector))
	if gotConnectorSHA256 != expectedConnectorSHA256 {
		t.Fatalf("connector rendering hash = %s, update expected hash after intentional canonical change", gotConnectorSHA256)
	}
}
