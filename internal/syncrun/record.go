// Package syncrun defines additive durable run-transition records. It does not
// read or write checkpoint state.
package syncrun

import (
	"fmt"
	"strings"
)

const RecordVersion uint = 1

type Phase string

const (
	PhasePlanned                Phase = "planned"
	PhasePreviewed              Phase = "previewed"
	PhaseApproved               Phase = "approved"
	PhaseStaging                Phase = "staging"
	PhaseStaged                 Phase = "staged"
	PhaseApplying               Phase = "applying"
	PhaseApplied                Phase = "applied"
	PhaseReconciliationRequired Phase = "reconciliation_required"
	PhaseReadbackVerified       Phase = "readback_verified"
	PhaseAcknowledged           Phase = "acknowledged"
	PhaseCheckpointed           Phase = "checkpointed"
	PhaseCompleted              Phase = "completed"
)

// Transition is additive run metadata. Sequence supplies durable ordering
// without importing a clock or changing existing checkpoint bytes.
type Transition struct {
	Sequence uint  `json:"sequence"`
	From     Phase `json:"from"`
	To       Phase `json:"to"`
}

// Record is a versioned, append-only account of one run's state progression.
type Record struct {
	Version     uint         `json:"version"`
	RunID       string       `json:"run_id"`
	Transitions []Transition `json:"transitions"`
}

func (r Record) Validate() error {
	if r.Version != RecordVersion {
		return fmt.Errorf("unsupported sync run record version %d", r.Version)
	}
	if strings.TrimSpace(r.RunID) == "" {
		return fmt.Errorf("sync run_id is required")
	}
	if len(r.Transitions) == 0 {
		return fmt.Errorf("sync run transitions are required")
	}
	current := Phase("")
	for index, transition := range r.Transitions {
		if transition.Sequence != uint(index+1) || transition.From != current || !allowedTransition(transition.From, transition.To) {
			return fmt.Errorf("invalid sync run transition %d", index+1)
		}
		current = transition.To
	}
	return nil
}

func allowedTransition(from, to Phase) bool {
	switch from {
	case "":
		return to == PhasePlanned
	case PhasePlanned:
		return to == PhasePreviewed
	case PhasePreviewed:
		return to == PhaseApproved
	case PhaseApproved:
		return to == PhaseStaging
	case PhaseStaging:
		return to == PhaseStaged
	case PhaseStaged:
		return to == PhaseApplying
	case PhaseApplying:
		return to == PhaseApplied || to == PhaseReconciliationRequired
	case PhaseApplied:
		return to == PhaseReadbackVerified
	case PhaseReadbackVerified:
		return to == PhaseAcknowledged
	case PhaseAcknowledged:
		return to == PhaseCheckpointed
	case PhaseCheckpointed:
		return to == PhaseCompleted
	default:
		return false
	}
}
