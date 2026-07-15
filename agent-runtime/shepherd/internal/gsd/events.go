package gsd

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type EventKind string

const (
	EventAgentStart     EventKind = "agent_start"
	EventTurnStart      EventKind = "turn_start"
	EventToolStart      EventKind = "tool_execution_start"
	EventToolEnd        EventKind = "tool_execution_end"
	EventAgentEnd       EventKind = "agent_end"
	EventModelSelect    EventKind = "model_select"
	EventThinkingSelect EventKind = "thinking_level_select"
)

var allowedEvents = map[EventKind]struct{}{
	EventAgentStart: {}, EventAgentEnd: {}, EventTurnStart: {}, EventToolStart: {}, EventToolEnd: {},
	EventModelSelect: {}, EventThinkingSelect: {},
}

type Event struct {
	Kind       EventKind
	RunID      string
	UnitID     string
	Tool       string
	ToolCallID string
	Status     string
	Model      string
	Thinking   string
	Nested     bool
	At         time.Time
}

func (e Event) String() string {
	return fmt.Sprintf("kind=%s run=%s unit=%s tool=%s tool_call_id=%s status=%s model=%s thinking=%s", e.Kind, e.RunID, e.UnitID, e.Tool, e.ToolCallID, e.Status, e.Model, e.Thinking)
}

// IsTopLevelIdentity excludes nested agent events from governed model evidence.
// Official top-level Pi model/thinking events carry neither subagent run nor
// unit correlation; delegated events carry at least one of those identifiers.
func (e Event) IsTopLevelIdentity() bool {
	return !e.Nested && e.RunID == "" && e.UnitID == ""
}

func ProjectEvent(raw []byte, maxBytes int) (Event, error) {
	if maxBytes <= 0 || len(raw) > maxBytes {
		return Event{}, fmt.Errorf("%w: event exceeds configured size", ErrRuntimeContractMismatch)
	}
	var envelope struct {
		Type          EventKind `json:"type"`
		RunID         string    `json:"runId"`
		UnitID        string    `json:"unitId"`
		Tool          string    `json:"toolName"`
		ToolCallID    string    `json:"toolCallId"`
		Status        string    `json:"status"`
		IsError       bool      `json:"isError"`
		Level         string    `json:"level"`
		ParentRunID   string    `json:"parentRunId"`
		ParentAgentID string    `json:"parentAgentId"`
		AgentID       string    `json:"agentId"`
		SubagentID    string    `json:"subagentId"`
		Scope         string    `json:"scope"`
		Source        string    `json:"source"`
		SelectedModel struct {
			Provider string `json:"provider"`
			ID       string `json:"id"`
		} `json:"model"`
		Messages []struct {
			Role     string `json:"role"`
			Provider string `json:"provider"`
			Model    string `json:"model"`
		} `json:"messages"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return Event{}, fmt.Errorf("%w: decode event envelope: %v", ErrRuntimeContractMismatch, err)
	}
	if _, ok := allowedEvents[envelope.Type]; !ok {
		return Event{}, fmt.Errorf("%w: event type %q is not allowlisted", ErrRuntimeContractMismatch, envelope.Type)
	}
	if (envelope.Type == EventToolStart || envelope.Type == EventToolEnd) &&
		(strings.TrimSpace(envelope.Tool) == "" || strings.TrimSpace(envelope.ToolCallID) == "") {
		return Event{}, fmt.Errorf("%w: tool lifecycle event requires toolName and toolCallId", ErrRuntimeContractMismatch)
	}
	model := ""
	if envelope.Type == EventModelSelect && envelope.SelectedModel.Provider != "" && envelope.SelectedModel.ID != "" {
		model = envelope.SelectedModel.Provider + "/" + envelope.SelectedModel.ID
	}
	for i := len(envelope.Messages) - 1; i >= 0; i-- {
		message := envelope.Messages[i]
		if message.Role == "assistant" && message.Model != "" {
			if message.Provider != "" {
				model = message.Provider + "/" + message.Model
			} else {
				model = message.Model
			}
			break
		}
	}
	status := envelope.Status
	if envelope.Type == EventToolEnd {
		if envelope.IsError {
			status = "error"
		} else {
			status = "success"
		}
	}
	thinking := ""
	if envelope.Type == EventThinkingSelect {
		thinking = envelope.Level
	}
	nested := envelope.ParentRunID != "" || envelope.ParentAgentID != "" || envelope.AgentID != "" || envelope.SubagentID != "" ||
		envelope.Scope == "subagent" || envelope.Scope == "nested" || envelope.Source == "subagent" || envelope.Source == "delegated"
	return Event{
		Kind: envelope.Type, RunID: envelope.RunID, UnitID: envelope.UnitID,
		Tool: envelope.Tool, ToolCallID: envelope.ToolCallID, Status: status, Model: model, Thinking: thinking, Nested: nested, At: time.Now().UTC(),
	}, nil
}
