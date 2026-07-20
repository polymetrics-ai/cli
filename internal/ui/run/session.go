package run

import (
	"context"
	"errors"
	"io"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"polymetrics.ai/internal/events"
)

const defaultSessionBuffer = 128

// SessionOptions configures one event-driven dashboard execution.
type SessionOptions struct {
	Config   Config
	Upstream events.Emitter
	Interval time.Duration
	Input    io.Reader
	Output   io.Writer
}

// Session owns the run context, event bridge, and dashboard model. Execute
// returns only after all emitted events have reached the model.
type Session struct {
	parent context.Context
	ctx    context.Context
	cancel context.CancelFunc
	model  *Model
	opts   SessionOptions
}

// NewSession creates a dashboard session whose cancellation is derived from parent.
func NewSession(parent context.Context, opts SessionOptions) *Session {
	if parent == nil {
		parent = context.Background()
	}
	ctx, cancel := context.WithCancel(parent)
	opts.Config.Cancel = cancel
	return &Session{
		parent: parent,
		ctx:    ctx,
		cancel: cancel,
		model:  NewModel(opts.Config),
		opts:   opts,
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

// Execute runs work as a Bubble Tea command and drains bridged events through
// Update before the inline program exits. Parent cancellation becomes a Tea
// message, allowing the runner to flush a truthful terminal lifecycle frame.
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
	watchCtx, stopWatch := context.WithCancel(context.Background())
	defer stopWatch()

	s.model.events = eventCh
	s.model.runCmd = func() tea.Msg {
		err := work(runCtx)
		bridge.Flush(runCtx)
		close(eventCh)
		return runResultMsg{Err: err}
	}
	s.model.cancelCmd = waitCancellation(s.parent.Done(), watchCtx.Done())

	output := s.opts.Output
	if output == nil {
		output = io.Discard
	}
	program := tea.NewProgram(
		s.model,
		tea.WithInput(s.opts.Input),
		tea.WithOutput(output),
		tea.WithWindowSize(s.model.cfg.Width, s.model.cfg.Height),
		tea.WithFPS(30),
		tea.WithoutSignalHandler(),
		tea.WithoutSignals(),
	)
	finalModel, programErr := program.Run()
	if model, ok := finalModel.(*Model); ok {
		s.model = model
	}
	return errors.Join(s.model.runErr, programErr)
}

type eventMsg struct {
	Event events.Event
}

type eventClosedMsg struct{}

type runResultMsg struct {
	Err error
}

type cancelMsg struct{}

func waitEvent(eventCh <-chan events.Event) tea.Cmd {
	if eventCh == nil {
		return nil
	}
	return func() tea.Msg {
		event, ok := <-eventCh
		if !ok {
			return eventClosedMsg{}
		}
		return eventMsg{Event: event}
	}
}

func waitCancellation(parentDone, stop <-chan struct{}) tea.Cmd {
	if parentDone == nil {
		return nil
	}
	return func() tea.Msg {
		select {
		case <-parentDone:
			return cancelMsg{}
		case <-stop:
			return nil
		}
	}
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
