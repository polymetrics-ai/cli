package gsd

import (
	"errors"
	"fmt"
)

func Reconcile(command string, result Result, snapshot WorkflowSnapshot) (Terminal, error) {
	if result.Terminal != TerminalSuccess && result.Terminal != TerminalBlocked {
		return result.Terminal, result.Err
	}
	if len(snapshot.Blockers) > 0 || snapshot.Next.Action == "blocked" || snapshot.Next.Action == "wait" {
		return TerminalBlocked, nil
	}
	if result.Terminal == TerminalBlocked {
		return TerminalBlocked, nil
	}
	switch snapshot.Next.Action {
	case "complete":
		return TerminalSuccess, nil
	case "dispatch":
		return TerminalBlocked, fmt.Errorf("%s process exited before pending unit %s/%s was settled", command, snapshot.Next.UnitType, snapshot.Next.UnitID)
	case "stop":
		if snapshot.Phase == "complete" {
			return TerminalSuccess, nil
		}
		return TerminalBlocked, nil
	default:
		return TerminalError, errors.New("event and query state cannot be reconciled")
	}
}
