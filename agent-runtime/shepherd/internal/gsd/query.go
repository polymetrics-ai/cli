package gsd

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/sensitive"
)

type WorkflowSnapshot struct {
	MilestoneID string
	SliceID     string
	TaskID      string
	Phase       string
	NextAction  string
	Blockers    []string
	Next        NextDispatch
}

type NextDispatch struct {
	Action   string
	UnitType string
	UnitID   string
}

func DecodeQuery(raw []byte) (WorkflowSnapshot, error) {
	if len(raw) == 0 || len(raw) > 1024*1024 {
		return WorkflowSnapshot{}, errors.New("query snapshot is empty or oversized")
	}
	var value struct {
		State struct {
			ActiveMilestone *struct {
				ID string `json:"id"`
			} `json:"activeMilestone"`
			ActiveSlice *struct {
				ID string `json:"id"`
			} `json:"activeSlice"`
			ActiveTask *struct {
				ID string `json:"id"`
			} `json:"activeTask"`
			Phase      string   `json:"phase"`
			NextAction string   `json:"nextAction"`
			Blockers   []string `json:"blockers"`
		} `json:"state"`
		Next struct {
			Action   string `json:"action"`
			UnitType string `json:"unitType"`
			UnitID   string `json:"unitId"`
		} `json:"next"`
	}
	if err := json.Unmarshal(raw, &value); err != nil {
		return WorkflowSnapshot{}, fmt.Errorf("decode headless query: %w", err)
	}
	if value.State.Phase == "" || value.Next.Action == "" {
		return WorkflowSnapshot{}, errors.New("query snapshot lacks canonical state or next action")
	}
	allowedActions := map[string]struct{}{"dispatch": {}, "skip": {}, "stop": {}}
	if _, ok := allowedActions[value.Next.Action]; !ok {
		return WorkflowSnapshot{}, fmt.Errorf("unknown next action %q", value.Next.Action)
	}
	if value.Next.UnitID != "" {
		if err := sensitive.ValidatePublicIdentifier(value.Next.UnitID); err != nil {
			return WorkflowSnapshot{}, fmt.Errorf("unsafe next unit identity: %w", err)
		}
	}
	snapshot := WorkflowSnapshot{Phase: value.State.Phase, NextAction: value.State.NextAction, Blockers: value.State.Blockers}
	if value.State.ActiveMilestone != nil {
		snapshot.MilestoneID = value.State.ActiveMilestone.ID
	}
	if value.State.ActiveSlice != nil {
		snapshot.SliceID = value.State.ActiveSlice.ID
	}
	if value.State.ActiveTask != nil {
		snapshot.TaskID = value.State.ActiveTask.ID
	}
	snapshot.Next = NextDispatch{Action: value.Next.Action, UnitType: value.Next.UnitType, UnitID: value.Next.UnitID}
	return snapshot, nil
}
