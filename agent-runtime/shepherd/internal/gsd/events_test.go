package gsd

import (
	"strings"
	"testing"
)

func TestProjectorAllowlistDropsPayloads(t *testing.T) {
	t.Parallel()

	raw := `{"type":"tool_execution_end","runId":"r1","toolName":"bash","result":"secret output","thinkingSignature":{"encrypted_content":"huge"}}`
	event, err := ProjectEvent([]byte(raw), 1024)
	if err != nil {
		t.Fatalf("project event: %v", err)
	}
	if event.Kind != EventToolEnd || event.RunID != "r1" || event.Tool != "bash" {
		t.Fatalf("unexpected projection: %+v", event)
	}
	if strings.Contains(event.String(), "secret") || strings.Contains(event.String(), "encrypted") {
		t.Fatalf("projection retained forbidden payload: %s", event.String())
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
