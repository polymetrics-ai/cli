package synctransport

import (
	"context"
	"errors"
	"sync"
	"time"
)

var ErrByteCreditsInvalid = errors.New("transport byte credits are invalid")

// DefaultByteCreditCapacity is the global retained-data ceiling for the fast
// segment pipeline. Credit is charged before a reader retains an admitted
// range and released only after the segment is durable and handed off.
const DefaultByteCreditCapacity int64 = 512 << 20

// ByteCreditController is connector-neutral byte-weighted backpressure. It
// controls retained payload memory, not record count or a provider protocol,
// so an S3 reader and a database range reader obey the same bounded pipeline.
type ByteCreditController struct {
	mu        sync.Mutex
	capacity  int64
	inUse     int64
	peak      int64
	waitNanos int64
	notify    chan struct{}
}

type ByteCreditSnapshot struct {
	Capacity  int64
	InUse     int64
	Peak      int64
	WaitNanos int64
}

// ByteCreditLease releases exactly the bytes granted by Acquire. It is safe to
// call Release repeatedly, which keeps failure cleanup idempotent.
type ByteCreditLease struct {
	controller *ByteCreditController
	bytes      int64
	once       sync.Once
}

func NewByteCreditController(capacity int64) (*ByteCreditController, error) {
	if capacity <= 0 {
		return nil, ErrByteCreditsInvalid
	}
	return &ByteCreditController{capacity: capacity, notify: make(chan struct{})}, nil
}

// Acquire waits for enough retained-byte credit or returns the caller's
// cancellation. A request larger than the configured capacity is rejected
// before waiting or touching a source/destination operation.
func (c *ByteCreditController) Acquire(ctx context.Context, bytes int64) (*ByteCreditLease, error) {
	if c == nil || ctx == nil || bytes <= 0 || bytes > c.capacity {
		return nil, ErrByteCreditsInvalid
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	started := time.Now()
	for {
		c.mu.Lock()
		if c.inUse <= c.capacity-bytes {
			c.inUse += bytes
			if c.inUse > c.peak {
				c.peak = c.inUse
			}
			c.waitNanos += time.Since(started).Nanoseconds()
			c.mu.Unlock()
			return &ByteCreditLease{controller: c, bytes: bytes}, nil
		}
		notify := c.notify
		c.mu.Unlock()
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-notify:
		}
	}
}

func (l *ByteCreditLease) Release() {
	if l == nil || l.controller == nil || l.bytes <= 0 {
		return
	}
	l.once.Do(func() {
		c := l.controller
		c.mu.Lock()
		if c.inUse < l.bytes {
			// This can only follow a forged internal lease. Preserve the bounded
			// invariant and wake waiters rather than allowing negative credit.
			c.inUse = 0
		} else {
			c.inUse -= l.bytes
		}
		close(c.notify)
		c.notify = make(chan struct{})
		c.mu.Unlock()
	})
}

func (c *ByteCreditController) Snapshot() ByteCreditSnapshot {
	if c == nil {
		return ByteCreditSnapshot{}
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return ByteCreditSnapshot{Capacity: c.capacity, InUse: c.inUse, Peak: c.peak, WaitNanos: c.waitNanos}
}
