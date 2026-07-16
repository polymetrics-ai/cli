package events

import (
	"bytes"
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"
)

func TestContextEmitterDefaultsToNop(t *testing.T) {
	ctx := context.Background()
	if got := FromContext(ctx); got == nil {
		t.Fatal("FromContext returned nil, want Nop emitter")
	}
	FromContext(ctx).Emit(ctx, Event{Kind: KindStarted, Scope: ScopeFlow, RunID: "run-1"})

	collector := NewCollector()
	ctx = WithEmitter(ctx, collector)
	Emit(ctx, Event{Kind: KindStarted, Scope: ScopeFlow, RunID: "run-2"})

	events := collector.Events()
	if len(events) != 1 {
		t.Fatalf("collector events = %d, want 1", len(events))
	}
	if events[0].RunID != "run-2" || events[0].Kind != KindStarted {
		t.Fatalf("unexpected event: %+v", events[0])
	}
}

func TestNDJSONSanitizesAndRedacts(t *testing.T) {
	var buf bytes.Buffer
	sink := NewNDJSON(&buf)
	sink.Emit(context.Background(), Event{
		Kind:    KindFailed,
		Scope:   ScopeETL,
		RunID:   "run-1\x1b[31m",
		StepID:  "step-1",
		Status:  "failed",
		Message: "request failed token=abc123 http://example.com/path?api_key=secret",
		Attrs: map[string]string{
			"api_key": "secret-value",
			"note":    "hello\x1b[0m",
		},
	})

	var got Event
	if err := json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &got); err != nil {
		t.Fatalf("decode ndjson: %v\nraw=%q", err, buf.String())
	}
	if got.RunID != "run-1[31m" {
		t.Fatalf("RunID sanitized = %q, want escape byte removed", got.RunID)
	}
	if got.Message != "request failed token=[redacted] http://example.com/path" {
		t.Fatalf("Message = %q", got.Message)
	}
	if got.Attrs["api_key"] != "[redacted]" {
		t.Fatalf("api_key attr = %q, want redacted", got.Attrs["api_key"])
	}
	if got.Attrs["note"] != "hello[0m" {
		t.Fatalf("note attr = %q, want sanitized", got.Attrs["note"])
	}
}

func TestChanCoalescesProgressAndPreservesLifecycle(t *testing.T) {
	sink := NewChan(2)
	defer sink.Close()
	ctx := context.Background()

	sink.Emit(ctx, Event{Kind: KindStarted, Scope: ScopeETL, RunID: "run-1"})
	sink.Emit(ctx, Event{Kind: KindProgress, Scope: ScopeETL, RunID: "run-1", Counters: Counters{RecordsRead: 1}})
	sink.Emit(ctx, Event{Kind: KindProgress, Scope: ScopeETL, RunID: "run-1", Counters: Counters{RecordsRead: 2}})
	sink.Emit(ctx, Event{Kind: KindProgress, Scope: ScopeETL, RunID: "run-1", Counters: Counters{RecordsRead: 3}})
	sink.Emit(ctx, Event{Kind: KindCompleted, Scope: ScopeETL, RunID: "run-1"})

	got := drainEvents(t, sink.Events(), 3)
	wantKinds := []Kind{KindStarted, KindProgress, KindCompleted}
	for i, want := range wantKinds {
		if got[i].Kind != want {
			t.Fatalf("event[%d].Kind = %q, want %q; events=%+v", i, got[i].Kind, want, got)
		}
	}
	if got[1].Counters.RecordsRead != 3 {
		t.Fatalf("coalesced RecordsRead = %d, want latest 3", got[1].Counters.RecordsRead)
	}
	if sink.Dropped() == 0 {
		t.Fatal("Dropped() = 0, want progress drop accounting")
	}
}

func TestMultiAndCollectorAreRaceClean(t *testing.T) {
	collectorA := NewCollector()
	collectorB := NewCollector()
	emitter := NewMulti(collectorA, collectorB, Nop{})
	ctx := context.Background()

	const goroutines = 16
	const perGoroutine = 50
	var wg sync.WaitGroup
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < perGoroutine; i++ {
				emitter.Emit(ctx, Event{Kind: KindProgress, Scope: ScopeETL, RunID: "race"})
			}
		}()
	}
	wg.Wait()

	want := goroutines * perGoroutine
	if len(collectorA.Events()) != want || len(collectorB.Events()) != want {
		t.Fatalf("collector sizes = %d/%d, want %d", len(collectorA.Events()), len(collectorB.Events()), want)
	}
}

func TestThrottleForwardsLifecycleAndFlushesLatestProgress(t *testing.T) {
	collector := NewCollector()
	throttle := NewThrottle(time.Hour, collector)
	ctx := context.Background()

	throttle.Emit(ctx, Event{Kind: KindStarted, Scope: ScopeFlow, RunID: "run-1"})
	throttle.Emit(ctx, Event{Kind: KindProgress, Scope: ScopeFlow, RunID: "run-1", Counters: Counters{RecordsWritten: 1}})
	throttle.Emit(ctx, Event{Kind: KindProgress, Scope: ScopeFlow, RunID: "run-1", Counters: Counters{RecordsWritten: 2}})
	throttle.Emit(ctx, Event{Kind: KindCompleted, Scope: ScopeFlow, RunID: "run-1"})
	throttle.Flush(ctx)

	events := collector.Events()
	wantKinds := []Kind{KindStarted, KindProgress, KindCompleted, KindProgress}
	if len(events) != len(wantKinds) {
		t.Fatalf("len(events) = %d, want %d: %+v", len(events), len(wantKinds), events)
	}
	for i, want := range wantKinds {
		if events[i].Kind != want {
			t.Fatalf("event[%d].Kind = %q, want %q", i, events[i].Kind, want)
		}
	}
	if events[3].Counters.RecordsWritten != 2 {
		t.Fatalf("flushed RecordsWritten = %d, want latest 2", events[3].Counters.RecordsWritten)
	}
	if throttle.Dropped() == 0 {
		t.Fatal("Dropped() = 0, want throttle coalescing accounting")
	}
}

func drainEvents(t *testing.T, ch <-chan Event, n int) []Event {
	t.Helper()
	got := make([]Event, 0, n)
	timer := time.NewTimer(time.Second)
	defer timer.Stop()
	for len(got) < n {
		select {
		case ev := <-ch:
			got = append(got, ev)
		case <-timer.C:
			t.Fatalf("timed out waiting for %d events, got %d", n, len(got))
		}
	}
	return got
}
