package app

import (
	"fmt"

	"polymetrics.ai/internal/connectors"
)

// IncompleteReverseWriteAcknowledgement means a bulk connector returned from
// Write without proving that every staged record reached a permitted terminal
// disposition. It is typed so callers retain the sanitized provider receipt
// while treating the run as repairable failure rather than an executed plan.
type IncompleteReverseWriteAcknowledgement struct {
	Staged          int
	Written         int
	Unchanged       int
	Failed          int
	AllowsUnchanged bool
}

func (e *IncompleteReverseWriteAcknowledgement) Error() string {
	if e == nil {
		return "reverse write acknowledgement is incomplete"
	}
	return fmt.Sprintf(
		"reverse write acknowledgement is incomplete: staged=%d written=%d unchanged=%d failed=%d allows_unchanged=%t",
		e.Staged, e.Written, e.Unchanged, e.Failed, e.AllowsUnchanged,
	)
}

func (a *App) reverseWriteAllowsUnchanged(planID string) bool {
	if a == nil || a.registry == nil {
		return false
	}
	var connectorName, actionName string
	for _, plan := range a.state.ReversePlans {
		if plan.ID == planID {
			connectorName, actionName = plan.DestinationConnector, plan.Action
			break
		}
	}
	if connectorName == "" || actionName == "" {
		return false
	}
	connector, found := a.registry.Get(connectorName)
	if !found {
		return false
	}
	for _, action := range connectors.ManifestOf(connector).WriteActions {
		if action.Name == actionName {
			return action.AllowsUnchanged
		}
	}
	return false
}

func validateCompleteReverseWriteAcknowledgement(staged int, result connectors.WriteResult, allowsUnchanged bool) error {
	invalid := func() error {
		return &IncompleteReverseWriteAcknowledgement{
			Staged: staged, Written: result.RecordsWritten, Unchanged: result.RecordsUnchanged,
			Failed: result.RecordsFailed, AllowsUnchanged: allowsUnchanged,
		}
	}
	if staged < 0 || result.RecordsWritten < 0 || result.RecordsUnchanged < 0 || result.RecordsFailed < 0 {
		return invalid()
	}
	if result.RecordsFailed != 0 || result.RecordsWritten > staged {
		return invalid()
	}
	remaining := staged - result.RecordsWritten
	if result.RecordsUnchanged != 0 {
		if !allowsUnchanged || result.RecordsUnchanged > remaining {
			return invalid()
		}
		remaining -= result.RecordsUnchanged
	}
	if remaining != 0 {
		return invalid()
	}
	return nil
}
