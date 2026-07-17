// Package logging provides secret-redacting slog handlers and per-run JSONL routing.
package logging

import (
	"context"
	"log/slog"

	"polymetrics.ai/internal/safety"
)

type loggerContextKey struct{}
type runIDContextKey struct{}
type registryContextKey struct{}

var discardLogger = slog.New(discardHandler{})

// WithLogger stores logger in ctx. A nil logger stores a no-op logger.
func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if logger == nil {
		logger = discardLogger
	}
	return context.WithValue(ctx, loggerContextKey{}, logger)
}

// FromContext returns the logger stored in ctx, or a no-op logger when none exists.
func FromContext(ctx context.Context) *slog.Logger {
	if ctx == nil {
		return discardLogger
	}
	logger, ok := ctx.Value(loggerContextKey{}).(*slog.Logger)
	if !ok || logger == nil {
		logger = discardLogger
	}
	if runID := RunIDFromContext(ctx); validRunID(runID) {
		return BindRunID(logger, runID)
	}
	return logger
}

// WithRegistry stores the invocation-scoped value registry in ctx.
func WithRegistry(ctx context.Context, registry *ValueRegistry) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if registry == nil {
		registry = NewValueRegistry()
	}
	return context.WithValue(ctx, registryContextKey{}, registry)
}

// RegistryFromContext returns the invocation-scoped registry in ctx, or the bounded package registry.
func RegistryFromContext(ctx context.Context) *ValueRegistry {
	if ctx == nil {
		return defaultRegistry
	}
	registry, ok := ctx.Value(registryContextKey{}).(*ValueRegistry)
	if !ok || registry == nil {
		return defaultRegistry
	}
	return registry
}

// RegisterValueFromContext registers value in the invocation-scoped registry when present.
func RegisterValueFromContext(ctx context.Context, value string) {
	RegistryFromContext(ctx).Register(value)
}

// RedactText applies the shared terminal, heuristic, and registered-value redaction primitive.
func RedactText(ctx context.Context, value string) string {
	return redactText(ctx, value, nil)
}

// RedactLine applies RedactText and collapses control/newline diagnostics to one safe line.
func RedactLine(ctx context.Context, value string) string {
	return safety.SanitizeTerminalLine(RedactText(ctx, value))
}

// WithRunID stores the validated run identifier used by run-file routing.
func WithRunID(ctx context.Context, runID string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, runIDContextKey{}, runID)
}

// RunIDFromContext returns the run identifier stored in ctx.
func RunIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	runID, _ := ctx.Value(runIDContextKey{}).(string)
	return runID
}

// BindRunID returns a logger whose handler injects runID into Handle contexts.
func BindRunID(logger *slog.Logger, runID string) *slog.Logger {
	if logger == nil {
		logger = discardLogger
	}
	if !validRunID(runID) {
		return logger
	}
	return slog.New(&boundRunIDHandler{runID: runID, next: logger.Handler()})
}

type boundRunIDHandler struct {
	runID string
	next  slog.Handler
}

func (h *boundRunIDHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(WithRunID(ctx, h.runID), level)
}

func (h *boundRunIDHandler) Handle(ctx context.Context, record slog.Record) error {
	ctx = WithRunID(ctx, h.runID)
	if !recordHasRunID(record) {
		record = record.Clone()
		record.AddAttrs(slog.String("run_id", h.runID))
	}
	return h.next.Handle(ctx, record)
}

func (h *boundRunIDHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &boundRunIDHandler{runID: h.runID, next: h.next.WithAttrs(attrs)}
}

func (h *boundRunIDHandler) WithGroup(name string) slog.Handler {
	return &boundRunIDHandler{runID: h.runID, next: h.next.WithGroup(name)}
}

func recordHasRunID(record slog.Record) bool {
	found := false
	record.Attrs(func(attr slog.Attr) bool {
		if attr.Key == "run_id" {
			found = true
			return false
		}
		return true
	})
	return found
}

func redactText(ctx context.Context, value string, fallback *ValueRegistry) string {
	value = safety.SanitizeTerminal(value)
	value = safety.RedactErrorText(value)
	if registry, ok := registryFromContext(ctx); ok {
		return registry.redactString(value)
	}
	if fallback != nil {
		return fallback.redactString(value)
	}
	return defaultRegistry.redactString(value)
}

func registryFromContext(ctx context.Context) (*ValueRegistry, bool) {
	if ctx == nil {
		return nil, false
	}
	registry, ok := ctx.Value(registryContextKey{}).(*ValueRegistry)
	if !ok || registry == nil {
		return nil, false
	}
	return registry, true
}
