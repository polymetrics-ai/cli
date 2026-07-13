package gsd

import (
	"strings"
	"testing"
)

func TestProjectorAllowlistDropsPayloads(t *testing.T) {
	t.Parallel()

	raw := `{"type":"tool_execution_end","runId":"r1","toolName":"bash","toolCallId":"call-1","isError":true,"result":"secret output","thinkingSignature":{"encrypted_content":"huge"}}`
	event, err := ProjectEvent([]byte(raw), 1024)
	if err != nil {
		t.Fatalf("project event: %v", err)
	}
	if event.Kind != EventToolEnd || event.RunID != "r1" || event.Tool != "bash" || event.ToolCallID != "call-1" {
		t.Fatalf("unexpected projection: %+v", event)
	}
	if event.Status != "error" {
		t.Fatalf("tool error status=%q", event.Status)
	}
	if strings.Contains(event.String(), "secret") || strings.Contains(event.String(), "encrypted") {
		t.Fatalf("projection retained forbidden payload: %s", event.String())
	}
}

func TestProjectorRequiresCorrelationIDForToolLifecycle(t *testing.T) {
	t.Parallel()

	for _, kind := range []string{"tool_execution_start", "tool_execution_end"} {
		if _, err := ProjectEvent([]byte(`{"type":"`+kind+`","toolName":"bash"}`), 1024); err == nil {
			t.Fatalf("expected %s without toolCallId to fail closed", kind)
		}
	}
}

func TestProjectorCapturesCompactEffectiveRuntimeIdentity(t *testing.T) {
	t.Parallel()

	model, err := ProjectEvent([]byte(`{"type":"model_select","model":{"provider":"openai-codex","id":"gpt-5.6-sol"},"source":"restore"}`), 1024)
	if err != nil || model.Model != "openai-codex/gpt-5.6-sol" {
		t.Fatalf("model event=%+v err=%v", model, err)
	}
	thinking, err := ProjectEvent([]byte(`{"type":"thinking_level_select","level":"high","previousLevel":"off"}`), 1024)
	if err != nil || thinking.Thinking != "high" {
		t.Fatalf("thinking event=%+v err=%v", thinking, err)
	}
}

func TestProjectorAcceptsRequestedAgentEndWithoutPayload(t *testing.T) {
	t.Parallel()

	event, err := ProjectEvent([]byte(`{"type":"agent_end","messages":[{"role":"assistant","provider":"openai-codex","model":"gpt-5.6-sol","content":"must not persist"}]}`), 1024)
	if err != nil {
		t.Fatalf("project agent end: %v", err)
	}
	if event.Kind != EventAgentEnd || event.Model != "openai-codex/gpt-5.6-sol" {
		t.Fatalf("agent end=%+v", event)
	}
	if strings.Contains(event.String(), "must not persist") {
		t.Fatalf("projection retained agent payload: %s", event.String())
	}
}

func TestProjectorRejectsUnknownAndOversizedEvents(t *testing.T) {
	t.Parallel()

	if _, err := ProjectEvent([]byte(`{"type":"message_update","content":"raw"}`), 1024); err == nil {
		t.Fatal("expected unknown event type to fail")
	}
	if _, err := ProjectEvent([]byte(strings.Repeat("x", 1025)), 1024); err == nil {
		t.Fatal("expected oversized event to fail")
	}
}
