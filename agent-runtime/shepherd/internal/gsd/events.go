package gsd

import (
	"encoding/json"
	"errors"
	"fmt"
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
	EventAgentStart: {}, EventTurnStart: {}, EventToolStart: {}, EventToolEnd: {},
	EventModelSelect: {}, EventThinkingSelect: {},
}

type Event struct {
	Kind     EventKind
	RunID    string
	UnitID   string
	Tool     string
	Status   string
	Model    string
	Thinking string
	At       time.Time
}

func (e Event) String() string {
	return fmt.Sprintf("kind=%s run=%s unit=%s tool=%s status=%s model=%s thinking=%s", e.Kind, e.RunID, e.UnitID, e.Tool, e.Status, e.Model, e.Thinking)
}

func ProjectEvent(raw []byte, maxBytes int) (Event, error) {
	if maxBytes <= 0 || len(raw) > maxBytes {
		return Event{}, errors.New("event exceeds configured size")
	}
	var envelope struct {
		Type          EventKind `json:"type"`
		RunID         string    `json:"runId"`
		UnitID        string    `json:"unitId"`
		Tool          string    `json:"toolName"`
		Status        string    `json:"status"`
		IsError       bool      `json:"isError"`
		Level         string    `json:"level"`
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
		return Event{}, fmt.Errorf("decode event envelope: %w", err)
	}
	if _, ok := allowedEvents[envelope.Type]; !ok {
		return Event{}, fmt.Errorf("event type %q is not allowlisted", envelope.Type)
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
	return Event{
		Kind: envelope.Type, RunID: envelope.RunID, UnitID: envelope.UnitID,
		Tool: envelope.Tool, Status: status, Model: model, Thinking: thinking, At: time.Now().UTC(),
	}, nil
}
