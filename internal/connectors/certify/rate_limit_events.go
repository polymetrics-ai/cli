package certify

import (
	"sync"

	"polymetrics.ai/internal/connectors/connsdk"
)

// rateLimitEventCollector is attached to the one ephemeral project used by a
// certification run. Requester events are synchronous today, but it is still
// synchronized so a future fan-out stage cannot corrupt the report.
type rateLimitEventCollector struct {
	mu           sync.Mutex
	currentStage string
	events       []RateLimitEvent
}

func (c *rateLimitEventCollector) setStage(stage string) {
	if c == nil {
		return
	}
	c.mu.Lock()
	c.currentStage = stage
	c.mu.Unlock()
}

func (c *rateLimitEventCollector) RecordRateLimitEvent(event connsdk.RateLimitEvent) {
	if c == nil {
		return
	}
	c.mu.Lock()
	c.events = append(c.events, RateLimitEvent{
		Type:       string(event.Type),
		Stage:      c.currentStage,
		Method:     event.Method,
		Attempt:    event.Attempt,
		DurationMS: event.DurationMS,
		ResetAt:    event.ResetAt,
		Reason:     event.Reason,
	})
	c.mu.Unlock()
}

func (c *rateLimitEventCollector) snapshot() []RateLimitEvent {
	if c == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]RateLimitEvent(nil), c.events...)
}
