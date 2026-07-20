package run

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"polymetrics.ai/internal/events"
)

const defaultSessionBuffer = 128

// SessionOptions configures one event-driven dashboard execution.
type SessionOptions struct {
	Config   Config
	Upstream events.Emitter
	Interval time.Duration
	Output   io.Writer
}

// Session owns the run context, event bridge, and dashboard model. Execute
// returns only after all emitted events have reached the model.
type Session struct {
	parent    context.Context
	ctx       context.Context
	cancel    context.CancelFunc
	model     *Model
	opts      SessionOptions
	renderer  inlineRenderer
	renderErr error
}

// NewSession creates a dashboard session whose cancellation is derived from parent.
func NewSession(parent context.Context, opts SessionOptions) *Session {
	if parent == nil {
		parent = context.Background()
	}
	ctx, cancel := context.WithCancel(parent)
	opts.Config.Cancel = cancel
	return &Session{
		parent:   parent,
		ctx:      ctx,
		cancel:   cancel,
		model:    NewModel(opts.Config),
		opts:     opts,
		renderer: inlineRenderer{output: opts.Output},
	}
}

// Model returns the session's dashboard model. Callers should inspect it after
// Execute returns; Execute owns model mutation while the run is active.
func (s *Session) Model() *Model {
	if s == nil {
		return nil
	}
	return s.model
}

// Execute runs work with the session emitter and drains every bridged event
// before returning. Parent cancellation requests runner cancellation but does
// not skip terminal event delivery or final-frame construction.
func (s *Session) Execute(work func(context.Context) error) error {
	if s == nil {
		return errors.New("dashboard session is nil")
	}
	if work == nil {
		return errors.New("dashboard work is nil")
	}
	defer s.cancel()

	eventCh := make(chan events.Event, defaultSessionBuffer)
	delivery := channelEmitter{events: eventCh}
	bridge := NewBridge(BridgeOptions{Interval: s.opts.Interval, Sink: delivery})
	runCtx := events.WithEmitter(s.ctx, events.NewMulti(s.opts.Upstream, bridge))
	resultCh := make(chan error, 1)

	go func() {
		err := work(runCtx)
		bridge.Flush(runCtx)
		close(eventCh)
		resultCh <- err
	}()

	parentDone := s.parent.Done()
	for eventCh != nil {
		select {
		case event, ok := <-eventCh:
			if !ok {
				eventCh = nil
				continue
			}
			s.model.Apply(event)
			s.render()
		case <-parentDone:
			s.model.HandleKey("ctrl+c")
			s.render()
			parentDone = nil
		}
	}

	err := <-resultCh
	if err != nil || !s.model.Done() {
		s.model.Apply(finalEvent(s.opts.Config, err))
		s.render()
	}
	return errors.Join(err, s.renderErr)
}

func (s *Session) render() {
	if s.renderErr != nil {
		return
	}
	s.renderErr = s.renderer.Render(s.model.View())
}

type inlineRenderer struct {
	output io.Writer
	lines  int
	drawn  bool
}

func (r *inlineRenderer) Render(frame string) error {
	if r == nil || r.output == nil {
		return nil
	}
	if r.drawn && r.lines > 0 {
		if _, err := fmt.Fprintf(r.output, "\x1b[%dA\r\x1b[J", r.lines); err != nil {
			return fmt.Errorf("refresh dashboard frame: %w", err)
		}
	}
	if _, err := io.WriteString(r.output, frame); err != nil {
		return fmt.Errorf("write dashboard frame: %w", err)
	}
	r.lines = strings.Count(frame, "\n")
	if !strings.HasSuffix(frame, "\n") {
		r.lines++
	}
	r.drawn = true
	return nil
}

type channelEmitter struct {
	events chan<- events.Event
}

func (e channelEmitter) Emit(_ context.Context, event events.Event) {
	e.events <- event.Clone()
}

func finalEvent(cfg Config, err error) events.Event {
	scope := events.Scope(strings.ToLower(strings.TrimSpace(cfg.Title)))
	event := events.Event{
		Kind:   events.KindCompleted,
		Scope:  scope,
		RunID:  cfg.Name,
		Status: "success",
	}
	if err == nil {
		return event
	}
	event.Kind = events.KindFailed
	event.Status = "failed"
	event.Message = err.Error()
	if errors.Is(err, context.Canceled) {
		event.Status = "cancelled"
	}
	return event
}

var _ events.Emitter = channelEmitter{}
