package agentcontract

import (
	"crypto/sha256"
	"fmt"
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
				value.GSD.Sequence[1].Args = "<phase>"
			},
		},
		{
			name: "GSD ship executable",
			mutate: func(value *Contract) {
				value.GSD.Sequence[len(value.GSD.Sequence)-1].Command = "ship"
			},
		},
		{
			name: "yes used by child command",
			mutate: func(value *Contract) {
				value.NoMistakes.ChildCommand += " --yes"
			},
		},
		{
			name: "yes used by parent command",
			mutate: func(value *Contract) {
				value.NoMistakes.ParentCommand += " --yes"
			},
		},
		{
			name: "child intent unquoted",
			mutate: func(value *Contract) {
				value.NoMistakes.ChildCommand = strings.Replace(value.NoMistakes.ChildCommand, "'<issue-intent>'", "<issue-intent>", 1)
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
				value.NoMistakes.ParentCommand = ""
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

	const expectedSHA256 = "da49efcdadddc85bb442a23ad75dc426ac23b9ee42cb3111ccbaef4e33d36bd1"
	gotSHA256 := fmt.Sprintf("%x", sha256.Sum256(base))
	if gotSHA256 != expectedSHA256 {
		t.Fatalf("base rendering hash = %s, update expected hash after intentional canonical change", gotSHA256)
	}
	const expectedConnectorSHA256 = "760c69bea2655c85c9e7ecee03e0fffcf61bf2b09fff8e74b255cfef669f1e2f"
	gotConnectorSHA256 := fmt.Sprintf("%x", sha256.Sum256(connector))
	if gotConnectorSHA256 != expectedConnectorSHA256 {
		t.Fatalf("connector rendering hash = %s, update expected hash after intentional canonical change", gotConnectorSHA256)
	}
}
