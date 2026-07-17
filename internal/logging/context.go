// Package logging provides secret-redacting slog handlers and per-run JSONL routing.
package logging

import (
	"context"
	"log/slog"
)

type loggerContextKey struct{}
type runIDContextKey struct{}

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
		return discardLogger
	}
	return logger
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
