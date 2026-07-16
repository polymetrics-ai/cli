package events

import (
	"context"
	"sync"
	"time"
)

const defaultChanLifecycleWait = 50 * time.Millisecond

// DropStats reports events that a Chan could not deliver. Progress drops include
// coalesced progress updates; lifecycle drops include terminal and other
// lifecycle events rejected by bounded backpressure or explicit Close drop
// semantics.
type DropStats struct {
	Progress  uint64
	Lifecycle uint64
}

// Total returns all accounted drops.
func (s DropStats) Total() uint64 {
	return s.Progress + s.Lifecycle
}

// Chan is a bounded channel-backed emitter for TUI and progress consumers.
// The internal queue never exceeds the configured capacity. Progress events may
// be coalesced or evicted first; lifecycle events use a bounded wait when the
// queue is full and are explicitly counted in DropStats when not delivered.
// Close is finite and uses accounted close-drop semantics for queued events.
type Chan struct {
	capacity int

	mu               sync.Mutex
	queue            []Event
	droppedProgress  uint64
	droppedLifecycle uint64
	closed           bool
	lifecycleWait    time.Duration

	out       chan Event
	notify    chan struct{}
	done      chan struct{}
	closeOnce sync.Once
}

// NewChan returns a channel sink with a bounded internal queue.
func NewChan(capacity int) *Chan {
	if capacity < 1 {
		capacity = 1
	}
	c := &Chan{
		capacity:      capacity,
		lifecycleWait: defaultChanLifecycleWait,
		out:           make(chan Event),
		notify:        make(chan struct{}, 1),
		done:          make(chan struct{}),
	}
	go c.run()
	return c
}

// Events returns the receive side of the sink.
func (c *Chan) Events() <-chan Event {
	if c == nil {
		ch := make(chan Event)
		close(ch)
		return ch
	}
	return c.out
}

// Dropped returns the number of coalesced or dropped progress events.
func (c *Chan) Dropped() uint64 {
	if c == nil {
		return 0
	}
	return c.DropStats().Progress
}

// DropStats returns progress and lifecycle drops accounted by backpressure,
// coalescing, or Close. Close is explicitly drop-accounting: queued events that
// have not already been delivered on Events are counted and discarded so Close
// remains finite even when consumers stall.
func (c *Chan) DropStats() DropStats {
	if c == nil {
		return DropStats{}
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return DropStats{Progress: c.droppedProgress, Lifecycle: c.droppedLifecycle}
}

// Close stops the sink and closes Events().
func (c *Chan) Close() {
	if c == nil {
		return
	}
	c.closeOnce.Do(func() {
		c.mu.Lock()
		c.closed = true
		for _, event := range c.queue {
			c.accountDropLocked(event)
		}
		c.queue = nil
		c.mu.Unlock()
		close(c.done)
		c.signal()
	})
}

// Emit implements Emitter.
func (c *Chan) Emit(ctx context.Context, event Event) {
	if c == nil {
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}
	event = event.Clone()
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return
	}
	if !event.Lifecycle() {
		c.enqueueProgressLocked(event)
		return
	}
	var timer *time.Timer
	for len(c.queue) >= c.capacity && !c.closed {
		if c.dropOldestProgressLocked() {
			continue
		}
		if timer == nil {
			timer = time.NewTimer(c.lifecycleWait)
			defer timer.Stop()
		}
		c.mu.Unlock()
		select {
		case <-ctx.Done():
			c.mu.Lock()
			c.accountDropLocked(event)
			return
		case <-timer.C:
			c.mu.Lock()
			c.accountDropLocked(event)
			return
		case <-c.notify:
			c.mu.Lock()
		case <-c.done:
			c.mu.Lock()
			c.accountDropLocked(event)
			return
		}
	}
	if c.closed {
		c.accountDropLocked(event)
		return
	}
	if len(c.queue) >= c.capacity {
		c.accountDropLocked(event)
		return
	}
	c.queue = append(c.queue, event)
	c.signal()
}

func (c *Chan) enqueueProgressLocked(event Event) {
	if len(c.queue) < c.capacity {
		c.queue = append(c.queue, event)
		c.signal()
		return
	}
	for i := len(c.queue) - 1; i >= 0; i-- {
		if !c.queue[i].Lifecycle() && sameProgressSlot(c.queue[i], event) {
			c.queue[i] = event
			c.droppedProgress++
			return
		}
	}
	for i := len(c.queue) - 1; i >= 0; i-- {
		if !c.queue[i].Lifecycle() {
			c.queue[i] = event
			c.droppedProgress++
			return
		}
	}
	c.droppedProgress++
}

func (c *Chan) dropOldestProgressLocked() bool {
	for i, event := range c.queue {
		if event.Lifecycle() {
			continue
		}
		copy(c.queue[i:], c.queue[i+1:])
		c.queue = c.queue[:len(c.queue)-1]
		c.droppedProgress++
		c.signal()
		return true
	}
	return false
}

func (c *Chan) progressCountLocked() int {
	count := 0
	for _, event := range c.queue {
		if !event.Lifecycle() {
			count++
		}
	}
	return count
}

func (c *Chan) run() {
	defer close(c.out)
	for {
		c.mu.Lock()
		for len(c.queue) == 0 && !c.closed {
			c.mu.Unlock()
			select {
			case <-c.notify:
			case <-c.done:
				return
			}
			c.mu.Lock()
		}
		if len(c.queue) == 0 && c.closed {
			c.mu.Unlock()
			return
		}
		event := c.queue[0]
		copy(c.queue[0:], c.queue[1:])
		c.queue = c.queue[:len(c.queue)-1]
		c.signal()
		c.mu.Unlock()

		select {
		case c.out <- event:
		case <-c.done:
			c.mu.Lock()
			c.accountDropLocked(event)
			c.mu.Unlock()
			return
		}
	}
}

func (c *Chan) accountDropLocked(event Event) {
	if event.Lifecycle() {
		c.droppedLifecycle++
		return
	}
	c.droppedProgress++
}

func (c *Chan) signal() {
	select {
	case c.notify <- struct{}{}:
	default:
	}
}

func sameProgressSlot(a, b Event) bool {
	return a.Scope == b.Scope && a.RunID == b.RunID && a.StepID == b.StepID
}

var _ Emitter = (*Chan)(nil)
