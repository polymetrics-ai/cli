package gsd

import (
	"errors"
	"fmt"
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
		if _, err := ProjectEvent([]byte(`{"type":"`+kind+`","toolName":"bash"}`), 1024); !errors.Is(err, ErrRuntimeContractMismatch) {
			t.Fatalf("error=%v, want runtime contract mismatch", err)
		}
	}
}

func TestProjectorRequiresUnambiguousToolCompletionOutcome(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{
		`{"type":"tool_execution_end","toolName":"read","toolCallId":"call-1"}`,
		`{"type":"tool_execution_end","toolName":"read","toolCallId":"call-1","isError":null}`,
		`{"type":"tool_execution_end","toolName":"read","toolCallId":"call-1","isError":false,"status":"error"}`,
		`{"type":"tool_execution_end","toolName":"read","toolCallId":"call-1","isError":true,"status":"success"}`,
	} {
		if _, err := ProjectEvent([]byte(raw), 1024); !errors.Is(err, ErrRuntimeContractMismatch) {
			t.Fatalf("ambiguous completion %s error=%v", raw, err)
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

func TestProjectorRejectsDuplicateLifecycleFields(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{
		`{"type":"agent_end","willRetry":true,"willRetry":false}`,
		`{"type":"turn_end","message":{"role":"assistant","stopReason":"error","stopReason":"stop"}}`,
		`{"type":"agent_start","type":"agent_end"}`,
		`{"type":"agent_end","willRetry":true,"WillRetry":false}`,
		`{"type":"turn_end","message":{"role":"assistant","stopReason":"error","StopReason":"stop"}}`,
		`{"type":"agent_end","Type":"agent_start"}`,
		`{"type":"agent_end","scope":"subagent","ſcope":"top"}`,
		`{"type":"agent_end","ſcope":"top"}`,
	} {
		if _, err := ProjectEvent([]byte(raw), 1024); !errors.Is(err, ErrRuntimeContractMismatch) {
			t.Fatalf("duplicate lifecycle event error=%v", err)
		}
	}
}

func TestDuplicateFieldValidationBoundsDistinctObjectKeys(t *testing.T) {
	t.Parallel()
	var raw strings.Builder
	raw.WriteByte('{')
	for index := 0; index < 1025; index++ {
		if index > 0 {
			raw.WriteByte(',')
		}
		fmt.Fprintf(&raw, "%q:true", fmt.Sprintf("field-%04d", index))
	}
	raw.WriteByte('}')
	if err := rejectDuplicateJSONFields([]byte(raw.String())); err == nil || !strings.Contains(err.Error(), "field-count bound") {
		t.Fatalf("oversized object fields error=%v", err)
	}
}

func TestProjectorRejectsUnknownAndOversizedEvents(t *testing.T) {
	t.Parallel()

	if _, err := ProjectEvent([]byte(`{"type":"message_update","content":"raw"}`), 1024); !errors.Is(err, ErrRuntimeContractMismatch) {
		t.Fatalf("unknown event error=%v, want runtime contract mismatch", err)
	}
	if _, err := ProjectEvent([]byte(strings.Repeat("x", 1025)), 1024); !errors.Is(err, ErrRuntimeContractMismatch) {
		t.Fatalf("oversized event error=%v, want runtime contract mismatch", err)
	}
}
