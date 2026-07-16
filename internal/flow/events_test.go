package flow

import (
	"context"
	"errors"
	"testing"

	"polymetrics.ai/internal/events"
)

func TestEngineEmitsDeterministicProgressEvents(t *testing.T) {
	m := twoStepManifest()
	tracker := &stepCallTracker{}
	e := newEngineForTest(t, m, tracker, &stubLedger{}, &FileCheckpointStore{Dir: t.TempDir()})
	collector := events.NewCollector()
	ctx := events.WithEmitter(context.Background(), collector)

	result, err := e.Run(ctx, RunOptions{})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.Status != "ok" {
		t.Fatalf("result.Status = %q, want ok", result.Status)
	}

	got := eventSequence(collector.Events())
	want := []string{
		"flow::started:running",
		"flow:sync-hubspot:started:running",
		"flow:sync-hubspot:completed:success",
		"flow:score-contacts:started:running",
		"flow:score-contacts:completed:success",
		"flow::completed:success",
	}
	assertStringSlice(t, got, want)
}

func TestEngineEmitsFailureProgressEvents(t *testing.T) {
	m := twoStepManifest()
	tracker := &stepCallTracker{fail: map[string]error{"SELECT * FROM contacts WHERE email IS NOT NULL": errors.New("query failed token=abc123")}}
	e := newEngineForTest(t, m, tracker, &stubLedger{}, &FileCheckpointStore{Dir: t.TempDir()})
	collector := events.NewCollector()
	ctx := events.WithEmitter(context.Background(), collector)

	_, err := e.Run(ctx, RunOptions{})
	if err == nil {
		t.Fatal("Run() error = nil, want failure")
	}

	got := eventSequence(collector.Events())
	want := []string{
		"flow::started:running",
		"flow:sync-hubspot:started:running",
		"flow:sync-hubspot:completed:success",
		"flow:score-contacts:started:running",
		"flow:score-contacts:failed:failed",
		"flow::failed:failed",
	}
	assertStringSlice(t, got, want)
}

func TestEngineEmitsSkippedProgressEvents(t *testing.T) {
	m := twoStepManifest()
	tracker := &stepCallTracker{}
	checkpoint := &FileCheckpointStore{Dir: t.TempDir()}
	if err := checkpoint.Set(m.Name, "sync-hubspot", "success"); err != nil {
		t.Fatal(err)
	}
	e := newEngineForTest(t, m, tracker, &stubLedger{}, checkpoint)
	collector := events.NewCollector()
	ctx := events.WithEmitter(context.Background(), collector)

	result, err := e.Run(ctx, RunOptions{})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.Status != "ok" {
		t.Fatalf("result.Status = %q, want ok", result.Status)
	}

	got := eventSequence(collector.Events())
	want := []string{
		"flow::started:running",
		"flow:sync-hubspot:skipped:skipped",
		"flow:score-contacts:started:running",
		"flow:score-contacts:completed:success",
		"flow::completed:success",
	}
	assertStringSlice(t, got, want)
}

func eventSequence(in []events.Event) []string {
	out := make([]string, 0, len(in))
	for _, ev := range in {
		out = append(out, string(ev.Scope)+":"+ev.StepID+":"+string(ev.Kind)+":"+ev.Status)
	}
	return out
}

func assertStringSlice(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("sequence length = %d, want %d\ngot  %#v\nwant %#v", len(got), len(want), got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("sequence[%d] = %q, want %q\ngot  %#v\nwant %#v", i, got[i], want[i], got, want)
		}
	}
}
