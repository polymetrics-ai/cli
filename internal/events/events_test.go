package events

import (
	"bytes"
	"context"
	"encoding/json"
	"runtime"
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
	started := drainEvents(t, sink.Events(), 1)
	if started[0].Kind != KindStarted {
		t.Fatalf("first event kind = %q, want started", started[0].Kind)
	}

	sink.Emit(ctx, Event{Kind: KindProgress, Scope: ScopeETL, RunID: "run-1", Counters: Counters{RecordsRead: 1}})
	sink.Emit(ctx, Event{Kind: KindProgress, Scope: ScopeETL, RunID: "run-1", Counters: Counters{RecordsRead: 2}})
	sink.Emit(ctx, Event{Kind: KindProgress, Scope: ScopeETL, RunID: "run-1", Counters: Counters{RecordsRead: 3}})
	sink.Emit(ctx, Event{Kind: KindCompleted, Scope: ScopeETL, RunID: "run-1"})

	got := drainUntilKind(t, sink.Events(), KindCompleted)
	if got[len(got)-1].Kind != KindCompleted {
		t.Fatalf("last event kind = %q, want completed; events=%+v", got[len(got)-1].Kind, got)
	}
	latestProgress := int64(0)
	for _, event := range got {
		if event.Kind == KindProgress {
			latestProgress = event.Counters.RecordsRead
		}
	}
	if latestProgress != 3 {
		t.Fatalf("latest progress RecordsRead = %d, want 3; events=%+v", latestProgress, got)
	}
	if sink.Dropped() == 0 {
		t.Fatal("Dropped() = 0, want progress drop accounting")
	}
}

func TestChanLifecycleInsertionEvictsProgressWithinCapacity(t *testing.T) {
	sink := NewChan(2)
	defer sink.Close()
	ctx := context.Background()

	sink.Emit(ctx, Event{Kind: KindStarted, Scope: ScopeETL, RunID: "run-capacity"})
	sink.Emit(ctx, Event{Kind: KindProgress, Scope: ScopeETL, RunID: "run-capacity", Counters: Counters{RecordsRead: 1}})
	sink.Emit(ctx, Event{Kind: KindCompleted, Scope: ScopeETL, RunID: "run-capacity"})
	sink.Emit(ctx, Event{Kind: KindFailed, Scope: ScopeETL, RunID: "run-capacity"})

	sink.mu.Lock()
	queueLen := len(sink.queue)
	sink.mu.Unlock()
	if queueLen > 2 {
		t.Fatalf("queue length = %d, want <= capacity 2", queueLen)
	}
	if stats := sink.DropStats(); stats.Progress == 0 {
		t.Fatalf("DropStats().Progress = 0, want progress eviction accounted (stats=%+v)", stats)
	}
}

func TestChanFullLifecycleQueueTimesOutAndAccountsLifecycle(t *testing.T) {
	sink := NewChan(1)
	defer sink.Close()
	ctx := context.Background()

	sink.Emit(ctx, Event{Kind: KindStarted, Scope: ScopeFlow, RunID: "run-lifecycle"})
	sink.Emit(ctx, Event{Kind: KindCompleted, Scope: ScopeFlow, RunID: "run-lifecycle"})

	done := make(chan struct{})
	go func() {
		sink.Emit(context.Background(), Event{Kind: KindFailed, Scope: ScopeFlow, RunID: "run-lifecycle"})
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(250 * time.Millisecond):
		t.Fatal("lifecycle Emit blocked on a full lifecycle queue; want bounded wait")
	}
	if stats := sink.DropStats(); stats.Lifecycle == 0 {
		t.Fatalf("DropStats().Lifecycle = 0, want lifecycle timeout/drop accounted (stats=%+v)", stats)
	}
}

func TestChanCloseAccountsQueuedEvents(t *testing.T) {
	sink := NewChan(2)
	ctx := context.Background()

	sink.Emit(ctx, Event{Kind: KindStarted, Scope: ScopeETL, RunID: "run-close"})
	sink.Emit(ctx, Event{Kind: KindProgress, Scope: ScopeETL, RunID: "run-close"})
	sink.Emit(ctx, Event{Kind: KindCompleted, Scope: ScopeETL, RunID: "run-close"})
	sink.Close()

	stats := sink.DropStats()
	if stats.Progress == 0 || stats.Lifecycle == 0 {
		t.Fatalf("DropStats() = %+v, want queued progress and lifecycle close drops accounted", stats)
	}
	select {
	case _, ok := <-sink.Events():
		if ok {
			t.Fatal("Events() yielded an event after Close accounted queued drops; want closed channel")
		}
	case <-time.After(250 * time.Millisecond):
		t.Fatal("Events() did not close after Close")
	}
}

func TestChanCloseWaitsForInFlightEventAccounting(t *testing.T) {
	previousProcs := runtime.GOMAXPROCS(1)
	defer runtime.GOMAXPROCS(previousProcs)

	sink := NewChan(2)
	ctx := context.Background()

	sink.Emit(ctx, Event{Kind: KindStarted, Scope: ScopeETL, RunID: "run-in-flight"})
	sink.Emit(ctx, Event{Kind: KindProgress, Scope: ScopeETL, RunID: "run-in-flight"})
	waitForQueueLen(t, sink, 1)

	closed := make(chan struct{})
	go func() {
		sink.Close()
		close(closed)
	}()

	select {
	case <-closed:
	case <-time.After(250 * time.Millisecond):
		t.Fatal("Close blocked with in-flight event and stalled consumer")
	}

	want := DropStats{Progress: 1, Lifecycle: 1}
	if got := sink.DropStats(); got != want {
		t.Fatalf("DropStats() after Close = %+v, want %+v", got, want)
	}
	select {
	case _, ok := <-sink.Events():
		if ok {
			t.Fatal("Events() yielded an event after Close; want closed channel")
		}
	default:
		t.Fatal("Events() not closed immediately after Close")
	}
}

func TestMultiWithBoundedChanDoesNotBlockIndefinitelyWhenChanLifecycleQueueStalls(t *testing.T) {
	sink := NewChan(1)
	defer sink.Close()
	collector := NewCollector()
	multi := NewMulti(sink, collector)
	ctx := context.Background()

	sink.Emit(ctx, Event{Kind: KindStarted, Scope: ScopeWorker, RunID: "run-multi"})
	sink.Emit(ctx, Event{Kind: KindCompleted, Scope: ScopeWorker, RunID: "run-multi"})

	done := make(chan struct{})
	go func() {
		multi.Emit(context.Background(), Event{Kind: KindFailed, Scope: ScopeWorker, RunID: "run-multi"})
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(250 * time.Millisecond):
		t.Fatal("Multi.Emit blocked behind a stalled Chan sink; want finite fanout")
	}
	if got := len(collector.Events()); got != 1 {
		t.Fatalf("collector events = %d, want 1 after Chan timeout", got)
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

func TestThrottleFlushesPendingProgressBeforeTerminal(t *testing.T) {
	collector := NewCollector()
	throttle := NewThrottle(time.Hour, collector)
	ctx := context.Background()

	throttle.Emit(ctx, Event{Kind: KindStarted, Scope: ScopeFlow, RunID: "run-1"})
	throttle.Emit(ctx, Event{Kind: KindProgress, Scope: ScopeFlow, RunID: "run-1", Counters: Counters{RecordsWritten: 1}})
	throttle.Emit(ctx, Event{Kind: KindProgress, Scope: ScopeFlow, RunID: "run-1", Counters: Counters{RecordsWritten: 2}})
	throttle.Emit(ctx, Event{Kind: KindCompleted, Scope: ScopeFlow, RunID: "run-1"})

	events := collector.Events()
	wantKinds := []Kind{KindStarted, KindProgress, KindProgress, KindCompleted}
	if len(events) != len(wantKinds) {
		t.Fatalf("len(events) = %d, want %d: %+v", len(events), len(wantKinds), events)
	}
	for i, want := range wantKinds {
		if events[i].Kind != want {
			t.Fatalf("event[%d].Kind = %q, want %q", i, events[i].Kind, want)
		}
	}
	if events[2].Counters.RecordsWritten != 2 {
		t.Fatalf("flushed RecordsWritten = %d, want latest 2", events[2].Counters.RecordsWritten)
	}
	if events[len(events)-1].Kind != KindCompleted {
		t.Fatalf("terminal event = %q, want completed last", events[len(events)-1].Kind)
	}
	if throttle.Dropped() == 0 {
		t.Fatal("Dropped() = 0, want throttle coalescing accounting")
	}
}

func waitForQueueLen(t *testing.T, sink *Chan, want int) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		sink.mu.Lock()
		got := len(sink.queue)
		sink.mu.Unlock()
		if got == want {
			return
		}
		time.Sleep(time.Millisecond)
	}
	sink.mu.Lock()
	got := len(sink.queue)
	sink.mu.Unlock()
	t.Fatalf("queue length = %d, want %d", got, want)
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

func drainUntilKind(t *testing.T, ch <-chan Event, kind Kind) []Event {
	t.Helper()
	got := make([]Event, 0, 4)
	timer := time.NewTimer(time.Second)
	defer timer.Stop()
	for {
		select {
		case ev := <-ch:
			got = append(got, ev)
			if ev.Kind == kind {
				return got
			}
		case <-timer.C:
			t.Fatalf("timed out waiting for %q, got %+v", kind, got)
		}
	}
}
