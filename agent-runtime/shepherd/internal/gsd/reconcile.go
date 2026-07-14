package gsd

import (
	"errors"
	"fmt"
)

var ErrMutatingSkip = errors.New("GSD query reconciled a mutating skip")

func Reconcile(command string, result Result, before, snapshot WorkflowSnapshot) (Terminal, error) {
	if result.Terminal != TerminalSuccess && result.Terminal != TerminalBlocked {
		return result.Terminal, result.Err
	}
	if len(snapshot.Blockers) > 0 {
		return TerminalBlocked, nil
	}
	if result.Terminal == TerminalBlocked {
		return TerminalBlocked, nil
	}
	if (command == "next" || IsCanonicalUnitCommand(command)) && before.Phase == snapshot.Phase && before.Next == snapshot.Next {
		return TerminalError, fmt.Errorf("%s exited successfully without advancing canonical GSD state", command)
	}
	switch snapshot.Next.Action {
	case "dispatch":
		if command == "next" || command == "new-milestone" || command == "status" || command == "discuss" || IsCanonicalUnitCommand(command) {
			return TerminalSuccess, nil
		}
		return TerminalBlocked, fmt.Errorf("%s process exited before pending unit %s/%s was settled", command, snapshot.Next.UnitType, snapshot.Next.UnitID)
	case "skip":
		return TerminalBlocked, fmt.Errorf("%w; run the next fenced unit explicitly", ErrMutatingSkip)
	case "stop":
		if snapshot.Phase == "complete" {
			return TerminalSuccess, nil
		}
		return TerminalBlocked, nil
	default:
		return TerminalError, errors.New("event and query state cannot be reconciled")
	}
}
