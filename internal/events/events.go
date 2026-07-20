// Package events provides a dependency-free progress event bus for long-running
// Polymetrics operations.
package events

import (
	"context"
	"encoding/json"
	"io"
	"strings"
	"sync"
	"time"

	pmlogging "polymetrics.ai/internal/logging"
	"polymetrics.ai/internal/safety"
)

// Kind identifies the event's lifecycle/progress role.
type Kind string

const (
	KindQueued    Kind = "queued"
	KindStarted   Kind = "started"
	KindProgress  Kind = "progress"
	KindSkipped   Kind = "skipped"
	KindResumed   Kind = "resumed"
	KindCompleted Kind = "completed"
	KindFailed    Kind = "failed"
)

// Scope identifies the subsystem that emitted an event.
type Scope string

const (
	ScopeFlow    Scope = "flow"
	ScopeETL     Scope = "etl"
	ScopeCertify Scope = "certify"
	ScopeWorker  Scope = "worker"
)

// Counters carries monotonically increasing operation counts when known.
type Counters struct {
	RecordsRead        int64 `json:"records_read,omitempty"`
	RecordsTransformed int64 `json:"records_transformed,omitempty"`
	RecordsWritten     int64 `json:"records_written,omitempty"`
	RecordsFailed      int64 `json:"records_failed,omitempty"`
	Batches            int64 `json:"batches,omitempty"`
	Completed          int64 `json:"completed,omitempty"`
	Total              int64 `json:"total,omitempty"`
	Dropped            int64 `json:"dropped,omitempty"`
}

// Event is the typed progress/lifecycle value passed through emitters.
type Event struct {
	Time     time.Time         `json:"time,omitempty"`
	Kind     Kind              `json:"kind"`
	Scope    Scope             `json:"scope,omitempty"`
	RunID    string            `json:"run_id,omitempty"`
	StepID   string            `json:"step_id,omitempty"`
	Status   string            `json:"status,omitempty"`
	Message  string            `json:"message,omitempty"`
	Counters Counters          `json:"counters,omitempty"`
	Attrs    map[string]string `json:"attrs,omitempty"`
}

// Lifecycle reports whether the event must not be coalesced as routine progress.
func (e Event) Lifecycle() bool {
	return e.Kind != KindProgress
}

// Clone returns a defensive copy of the event and its attributes.
func (e Event) Clone() Event {
	if len(e.Attrs) == 0 {
		return e
	}
	attrs := make(map[string]string, len(e.Attrs))
	for k, v := range e.Attrs {
		attrs[k] = v
	}
	e.Attrs = attrs
	return e
}

// Emitter is the minimal sink interface for progress events.
type Emitter interface {
	Emit(context.Context, Event)
}

type contextKey struct{}

// WithEmitter stores emitter in ctx. A nil emitter stores Nop.
func WithEmitter(ctx context.Context, emitter Emitter) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if emitter == nil {
		emitter = Nop{}
	}
	return context.WithValue(ctx, contextKey{}, emitter)
}

// FromContext returns the emitter in ctx or Nop when none is present.
func FromContext(ctx context.Context) Emitter {
	if ctx == nil {
		return Nop{}
	}
	emitter, ok := ctx.Value(contextKey{}).(Emitter)
	if !ok || emitter == nil {
		return Nop{}
	}
	return emitter
}

// Emit sends event to the emitter stored in ctx.
func Emit(ctx context.Context, event Event) {
	FromContext(ctx).Emit(ctx, sanitizeEvent(ctx, event))
}

// Nop is an emitter sink that discards every event.
type Nop struct{}

// Emit implements Emitter.
func (Nop) Emit(context.Context, Event) {}

var _ Emitter = Nop{}

// Collector records events in memory for tests and local adapters.
type Collector struct {
	mu     sync.Mutex
	events []Event
}

// NewCollector returns an in-memory emitter.
func NewCollector() *Collector {
	return &Collector{}
}

// Emit implements Emitter.
func (c *Collector) Emit(ctx context.Context, event Event) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.events = append(c.events, sanitizeEvent(ctx, event))
}

// Events returns a copy of collected events.
func (c *Collector) Events() []Event {
	if c == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]Event, len(c.events))
	for i, event := range c.events {
		out[i] = event.Clone()
	}
	return out
}

var _ Emitter = (*Collector)(nil)

// NDJSON writes sanitized events as newline-delimited JSON to the supplied writer.
type NDJSON struct {
	mu  sync.Mutex
	enc *json.Encoder
}

// NewNDJSON returns an NDJSON sink using w. Passing nil creates a Nop-like sink.
func NewNDJSON(w io.Writer) *NDJSON {
	if w == nil {
		return &NDJSON{}
	}
	return &NDJSON{enc: json.NewEncoder(w)}
}

// Emit implements Emitter.
func (n *NDJSON) Emit(ctx context.Context, event Event) {
	if n == nil || n.enc == nil {
		return
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	_ = n.enc.Encode(sanitizeEvent(ctx, event))
}

var _ Emitter = (*NDJSON)(nil)

// Multi emits every event to each configured sink synchronously in order.
// Multi does not create goroutines or make arbitrary sinks finite; sinks that
// can block must return on context cancellation or otherwise bound their work.
// The built-in bounded Chan sink observes this contract for backpressure and
// close paths.
type Multi struct {
	sinks []Emitter
}

// NewMulti returns a synchronous fanout sink.
func NewMulti(sinks ...Emitter) *Multi {
	out := make([]Emitter, 0, len(sinks))
	for _, sink := range sinks {
		if sink == nil {
			continue
		}
		out = append(out, sink)
	}
	return &Multi{sinks: out}
}

// Emit implements Emitter.
func (m *Multi) Emit(ctx context.Context, event Event) {
	if m == nil {
		return
	}
	event = sanitizeEvent(ctx, event)
	for _, sink := range m.sinks {
		sink.Emit(ctx, event.Clone())
	}
}

var _ Emitter = (*Multi)(nil)

func sanitizeEvent(ctx context.Context, event Event) Event {
	event = event.Clone()
	event.RunID = sanitizeString(ctx, event.RunID)
	event.StepID = sanitizeString(ctx, event.StepID)
	event.Status = sanitizeString(ctx, event.Status)
	event.Message = sanitizeString(ctx, event.Message)
	if len(event.Attrs) > 0 {
		attrs := make(map[string]string, len(event.Attrs))
		for k, v := range event.Attrs {
			key := sanitizeString(ctx, k)
			if secretKey(key) {
				attrs[key] = "[redacted]"
				continue
			}
			attrs[key] = sanitizeString(ctx, v)
		}
		event.Attrs = attrs
	}
	return event
}

func sanitizeString(ctx context.Context, value string) string {
	return pmlogging.RedactText(ctx, safety.SanitizeTerminal(value))
}

func secretKey(key string) bool {
	key = strings.ToLower(key)
	for _, token := range []string{"api_key", "apikey", "access_token", "token", "secret", "password"} {
		if strings.Contains(key, token) {
			return true
		}
	}
	return false
}
