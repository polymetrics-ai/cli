package telemetry

import (
	"testing"
	"time"
)

func TestEvaluateNormalizedEvidence(t *testing.T) {
	t.Parallel()

	base := time.Unix(100, 0).UTC()
	summary := Evaluate([]Activity{
		{RunID: "r1", Kind: "heartbeat", At: base},
		{RunID: "r1", Kind: "tool.started", Tool: "read", At: base.Add(time.Second)},
		{RunID: "r1", Kind: "heartbeat", At: base.Add(15 * time.Second)},
		{RunID: "r1", Kind: "validation", Status: "RETRY", At: base.Add(16 * time.Second)},
		{RunID: "r1", Kind: "run.terminal", Status: "blocked", At: base.Add(17 * time.Second)},
	})
	if summary.Runs != 1 || summary.ToolCalls != 1 || summary.MaxHeartbeatGapMS != 15_000 {
		t.Fatalf("summary=%+v", summary)
	}
	if summary.ValidationCounts["RETRY"] != 1 || summary.TerminalCounts["blocked"] != 1 {
		t.Fatalf("summary=%+v", summary)
	}
}

func TestEvaluateCountsAHeartbeatFreeLongRunAsBreach(t *testing.T) {
	t.Parallel()
	base := time.Unix(100, 0).UTC()
	summary := Evaluate([]Activity{
		{RunID: "r1", Kind: "run.started", At: base},
		{RunID: "r1", Kind: "run.terminal", Status: "blocked", At: base.Add(16 * time.Second)},
	})
	if summary.HeartbeatSLOBreaches != 1 || summary.MaxHeartbeatGapMS != 16_000 {
		t.Fatalf("summary=%+v", summary)
	}
}
