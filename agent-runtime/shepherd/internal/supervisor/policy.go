package supervisor

import (
	"errors"
	"fmt"

	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/gsd"
)

type DecisionKind string

const (
	DecisionDispatch  DecisionKind = "dispatch"
	DecisionFinalGate DecisionKind = "final_human_gate"
	DecisionBlocked   DecisionKind = "blocked"
)

type Decision struct {
	Kind     DecisionKind
	Command  string
	Unit     string
	Reason   string
	Snapshot gsd.WorkflowSnapshot
}

func Decide(snapshot gsd.WorkflowSnapshot) (Decision, error) {
	return DecideWithRegistry(snapshot, gsd.BuiltinUnitRegistry())
}

func DecideWithRegistry(snapshot gsd.WorkflowSnapshot, registry gsd.UnitRegistry) (Decision, error) {
	decision := Decision{Snapshot: snapshot}
	switch snapshot.Next.Action {
	case "dispatch":
		if snapshot.Next.UnitType == "" {
			return Decision{}, errors.New("canonical dispatch is missing a unit type")
		}
		command, err := registry.CommandForUnit(snapshot.Next.UnitType)
		if err != nil {
			return Decision{}, err
		}
		unit := snapshot.Next.UnitType
		if snapshot.Next.UnitID != "" {
			unit += "/" + snapshot.Next.UnitID
		}
		decision.Kind = DecisionDispatch
		decision.Command = command
		decision.Unit = unit
		return decision, nil
	case "stop":
		if snapshot.Phase == "complete" {
			decision.Kind = DecisionFinalGate
			decision.Reason = "final parent PR merge remains human-gated"
			return decision, nil
		}
		decision.Kind = DecisionBlocked
		decision.Reason = "canonical GSD query requested stop before milestone completion"
		return decision, nil
	case "skip":
		decision.Kind = DecisionBlocked
		decision.Reason = "canonical GSD query requested skip; supervisor requires an explicit human decision"
		return decision, nil
	default:
		return Decision{}, fmt.Errorf("unknown canonical action %q", snapshot.Next.Action)
	}
}
