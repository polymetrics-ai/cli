package agentcontract

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"testing"
)

func TestCanonicalContractRequiredInvariants(t *testing.T) {
	contract := loadRepositoryContract(t, repositoryRoot(t))

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
			copy := *contract
			test.mutate(&copy)
			if err := copy.Validate(); err == nil {
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

	const expectedSHA256 = "4c19f4942a2aa7fab078abc491e1df89964e383891aecfe584cc11bb871c4fe2"
	gotSHA256 := fmt.Sprintf("%x", sha256.Sum256(base))
	if gotSHA256 != expectedSHA256 {
		t.Fatalf("base rendering hash = %s, update expected hash after intentional canonical change", gotSHA256)
	}
	const expectedConnectorSHA256 = "6b6d433476c140f089d2b9232c6ff05ea5018e1b8a797ba3311eef13b311f1c3"
	gotConnectorSHA256 := fmt.Sprintf("%x", sha256.Sum256(connector))
	if gotConnectorSHA256 != expectedConnectorSHA256 {
		t.Fatalf("connector rendering hash = %s, update expected hash after intentional canonical change", gotConnectorSHA256)
	}
}
