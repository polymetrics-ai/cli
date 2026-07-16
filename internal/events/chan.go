package events

import (
	"context"
	"sync"
)

// Chan is a bounded channel-backed emitter for TUI and progress consumers.
// Lifecycle events are queued or block until context cancellation; progress
// events may be coalesced with drop accounting when the queue is full.
type Chan struct {
	capacity int

	mu      sync.Mutex
	queue   []Event
	dropped uint64
	closed  bool

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
		capacity: capacity,
		out:      make(chan Event),
		notify:   make(chan struct{}, 1),
		done:     make(chan struct{}),
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
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.dropped
}

// Close stops the sink and closes Events().
func (c *Chan) Close() {
	if c == nil {
		return
	}
	c.closeOnce.Do(func() {
		c.mu.Lock()
		c.closed = true
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
	for len(c.queue) >= c.capacity && !c.closed {
		progressCount := c.progressCountLocked()
		if progressCount > 1 && c.dropOldestProgressLocked() {
			break
		}
		if progressCount == 1 {
			// Keep the latest coalesced progress visible; lifecycle may exceed the
			// progress queue bound instead of silently erasing the only progress state.
			break
		}
		c.mu.Unlock()
		select {
		case <-ctx.Done():
			c.mu.Lock()
			return
		case <-c.notify:
			c.mu.Lock()
		case <-c.done:
			c.mu.Lock()
			return
		}
	}
	if c.closed {
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
			c.dropped++
			return
		}
	}
	for i := len(c.queue) - 1; i >= 0; i-- {
		if !c.queue[i].Lifecycle() {
			c.queue[i] = event
			c.dropped++
			return
		}
	}
	c.dropped++
}

func (c *Chan) dropOldestProgressLocked() bool {
	for i, event := range c.queue {
		if event.Lifecycle() {
			continue
		}
		copy(c.queue[i:], c.queue[i+1:])
		c.queue = c.queue[:len(c.queue)-1]
		c.dropped++
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
			return
		}
	}
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
