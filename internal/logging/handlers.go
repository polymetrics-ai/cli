package logging

import (
	"context"
	"errors"
	"io"
	"log/slog"
)

type discardHandler struct{}

func (discardHandler) Enabled(context.Context, slog.Level) bool  { return false }
func (discardHandler) Handle(context.Context, slog.Record) error { return nil }
func (discardHandler) WithAttrs([]slog.Attr) slog.Handler        { return discardHandler{} }
func (discardHandler) WithGroup(string) slog.Handler             { return discardHandler{} }

// MultiHandler fans a slog record out to multiple handlers.
type MultiHandler struct {
	handlers []slog.Handler
}

// NewMultiHandler returns a handler that synchronously writes to handlers in order.
func NewMultiHandler(handlers ...slog.Handler) slog.Handler {
	out := make([]slog.Handler, 0, len(handlers))
	for _, handler := range handlers {
		if handler == nil {
			continue
		}
		out = append(out, handler)
	}
	if len(out) == 0 {
		return discardHandler{}
	}
	return &MultiHandler{handlers: out}
}

func (h *MultiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (h *MultiHandler) Handle(ctx context.Context, record slog.Record) error {
	var errs []error
	for _, handler := range h.handlers {
		if !handler.Enabled(ctx, record.Level) {
			continue
		}
		if err := handler.Handle(ctx, record.Clone()); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (h *MultiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	out := make([]slog.Handler, 0, len(h.handlers))
	for _, handler := range h.handlers {
		out = append(out, handler.WithAttrs(attrs))
	}
	return &MultiHandler{handlers: out}
}

func (h *MultiHandler) WithGroup(name string) slog.Handler {
	out := make([]slog.Handler, 0, len(h.handlers))
	for _, handler := range h.handlers {
		out = append(out, handler.WithGroup(name))
	}
	return &MultiHandler{handlers: out}
}

// LevelHandler forwards records at or above min to the wrapped handler.
type LevelHandler struct {
	min  slog.Level
	next slog.Handler
}

// NewLevelHandler returns a level-filtering handler.
func NewLevelHandler(min slog.Level, next slog.Handler) slog.Handler {
	if next == nil {
		next = discardHandler{}
	}
	return &LevelHandler{min: min, next: next}
}

func (h *LevelHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.min && h.next.Enabled(ctx, level)
}

func (h *LevelHandler) Handle(ctx context.Context, record slog.Record) error {
	if record.Level < h.min {
		return nil
	}
	return h.next.Handle(ctx, record)
}

func (h *LevelHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &LevelHandler{min: h.min, next: h.next.WithAttrs(attrs)}
}

func (h *LevelHandler) WithGroup(name string) slog.Handler {
	return &LevelHandler{min: h.min, next: h.next.WithGroup(name)}
}

// LoggerOptions configures NewLogger.
type LoggerOptions struct {
	SensitiveKeys []string
	Registry      *ValueRegistry
	MaxLogFiles   int
}

// NewLogger builds the redacting run logger used by the CLI.
func NewLogger(projectDir string, stderr io.Writer, opts LoggerOptions) (*slog.Logger, func() error) {
	if stderr == nil {
		stderr = io.Discard
	}
	runFiles := NewRunFileHandler(projectDir, RunFileOptions{MaxFiles: opts.MaxLogFiles})
	stderrHandler := slog.NewJSONHandler(stderr, &slog.HandlerOptions{Level: slog.LevelWarn})
	fanout := NewMultiHandler(runFiles, NewLevelHandler(slog.LevelWarn, stderrHandler))
	redactor := NewRedactingHandler(fanout, RedactionOptions{SensitiveKeys: opts.SensitiveKeys, Registry: opts.Registry})
	return slog.New(redactor), runFiles.Close
}
