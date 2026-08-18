package certify

import (
	"testing"
	"time"

	"polymetrics.ai/internal/connectors/connsdk"
)

func TestRateLimitEventCollectorTagsStructuredEventsWithOwningStage(t *testing.T) {
	collector := &rateLimitEventCollector{}
	rc := &runContext{rateLimitEvents: collector}
	rep := &Report{}
	reset := time.Date(2026, time.August, 14, 13, 0, 0, 0, time.UTC)
	recordStage(rc, rep, "catalog", 2, func() (bool, CLIStageInfo, string) {
		collector.RecordRateLimitEvent(connsdk.RateLimitEvent{Type: connsdk.RateLimitEventAttempt, Method: "POST", Attempt: 2})
		collector.RecordRateLimitEvent(connsdk.RateLimitEvent{Type: connsdk.RateLimitEventReset, Attempt: 2, ResetAt: reset})
		collector.RecordRateLimitEvent(connsdk.RateLimitEvent{Type: connsdk.RateLimitEventNotSent, Method: "POST", Attempt: 3, Reason: "deadline_cutoff"})
		return true, CLIStageInfo{}, ""
	})

	rep.RateLimitEvents = collector.snapshot()
	if got, want := len(rep.RateLimitEvents), 3; got != want {
		t.Fatalf("RateLimitEvents = %+v, want %d events", rep.RateLimitEvents, want)
	}
	if got := rep.RateLimitEvents[0]; got.Type != "attempt" || got.Stage != "catalog" || got.Method != "POST" || got.Attempt != 2 {
		t.Fatalf("attempt event = %+v, want structured catalog attempt", got)
	}
	if got := rep.RateLimitEvents[1]; got.Type != "reset" || got.Stage != "catalog" || !got.ResetAt.Equal(reset) {
		t.Fatalf("reset event = %+v, want structured catalog reset", got)
	}
	if got := rep.RateLimitEvents[2]; got.Type != "not_sent" || got.Stage != "catalog" || got.Reason != "deadline_cutoff" {
		t.Fatalf("not-sent event = %+v, want catalog deadline cutoff", got)
	}
	if rc.currentStage != "" {
		t.Fatalf("current stage = %q after recordStage, want restored empty stage", rc.currentStage)
	}
}

func TestCertificationRateLimitAdmissionTimeoutUsesDefaultAndExplicitValue(t *testing.T) {
	if got := certificationRateLimitAdmissionTimeout(Options{}); got != defaultCertificationRateLimitAdmissionTimeout {
		t.Fatalf("default admission timeout = %s, want %s", got, defaultCertificationRateLimitAdmissionTimeout)
	}
	if got := certificationRateLimitAdmissionTimeout(Options{RateLimitAdmissionTimeout: 7 * time.Second}); got != 7*time.Second {
		t.Fatalf("explicit admission timeout = %s, want 7s", got)
	}
}
