package events

import (
	"context"
	"sync"
	"time"
)

// Throttle forwards lifecycle events immediately and coalesces progress events
// within an interval. Call Flush to deliver the latest coalesced progress event.
type Throttle struct {
	interval time.Duration
	sink     Emitter
	now      func() time.Time

	mu      sync.Mutex
	last    time.Time
	pending *Event
	dropped uint64
}

// NewThrottle returns a throttling emitter around sink.
func NewThrottle(interval time.Duration, sink Emitter) *Throttle {
	if sink == nil {
		sink = Nop{}
	}
	return &Throttle{interval: interval, sink: sink, now: time.Now}
}

// Emit implements Emitter.
func (t *Throttle) Emit(ctx context.Context, event Event) {
	if t == nil {
		return
	}
	if event.Lifecycle() {
		t.sink.Emit(ctx, event)
		return
	}

	now := t.now()
	var forward *Event
	t.mu.Lock()
	if t.last.IsZero() || t.interval <= 0 || now.Sub(t.last) >= t.interval {
		eventCopy := event.Clone()
		forward = &eventCopy
		t.last = now
	} else {
		eventCopy := event.Clone()
		t.pending = &eventCopy
		t.dropped++
	}
	t.mu.Unlock()
	if forward != nil {
		t.sink.Emit(ctx, *forward)
	}
}

// Flush forwards the latest coalesced progress event, when one exists.
func (t *Throttle) Flush(ctx context.Context) {
	if t == nil {
		return
	}
	var pending *Event
	t.mu.Lock()
	if t.pending != nil {
		eventCopy := t.pending.Clone()
		pending = &eventCopy
		t.pending = nil
		t.last = t.now()
	}
	t.mu.Unlock()
	if pending != nil {
		t.sink.Emit(ctx, *pending)
	}
}

// Dropped returns the number of coalesced progress events.
func (t *Throttle) Dropped() uint64 {
	if t == nil {
		return 0
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.dropped
}

var _ Emitter = (*Throttle)(nil)
