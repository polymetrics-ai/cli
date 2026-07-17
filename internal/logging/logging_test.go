package logging

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tlog "go.temporal.io/sdk/log"
)

const syntheticCanary = "pm-test-canary-redaction-value-404"

func TestRedactingHandlerScrubsMessagesAttrsGroupsErrorsAndURLs(t *testing.T) {
	registry := NewValueRegistry()
	registry.Register(syntheticCanary)

	var buf bytes.Buffer
	logger := slog.New(NewRedactingHandler(slog.NewJSONHandler(&buf, nil), RedactionOptions{
		SensitiveKeys: []string{"workspace_token"},
		Registry:      registry,
	}))

	err := errors.New("request failed with " + syntheticCanary + " at https://api.example.test/path?token=" + syntheticCanary)
	logger.InfoContext(context.Background(),
		"message contains "+syntheticCanary+" and https://example.test/a?api_key="+syntheticCanary,
		"token", syntheticCanary,
		"workspace_token", "connector-specific-field",
		"url", "https://example.test/a?api_key="+syntheticCanary+"&ok=1",
		"err", err,
		slog.Group("nested",
			"api_key", syntheticCanary,
			"note", "nested "+syntheticCanary,
			"safe_url", "https://nested.example.test/path?password="+syntheticCanary,
		),
		"headers", map[string]string{"Authorization": "Bearer " + syntheticCanary},
	)

	out := buf.String()
	assertDoesNotContainCanary(t, out)
	for _, forbidden := range []string{"api_key=", "token=", "password=", "Authorization"} {
		if strings.Contains(out, forbidden) {
			t.Fatalf("redacted log retained forbidden marker %q", forbidden)
		}
	}
	if !strings.Contains(out, "[redacted]") {
		t.Fatalf("redacted log missing replacement marker: %s", out)
	}
}

func TestRegisterSensitiveKeyRedactsDynamicConnectorFields(t *testing.T) {
	RegisterSensitiveKey("connector_dynamic_secret")
	var buf bytes.Buffer
	logger := slog.New(NewRedactingHandler(slog.NewJSONHandler(&buf, nil), RedactionOptions{}))
	logger.Info("dynamic key", "connector_dynamic_secret", "field-value")
	if strings.Contains(buf.String(), "field-value") {
		t.Fatalf("dynamic sensitive key was not redacted")
	}
}

func TestRedactingHandlerRedactsInlineEmptyAndSensitiveGroupsAtHandle(t *testing.T) {
	registry := NewValueRegistry()
	var buf bytes.Buffer
	logger := slog.New(NewRedactingHandler(slog.NewJSONHandler(&buf, nil), RedactionOptions{Registry: registry}))

	bound := logger.With(
		slog.Group("", slog.String("inline_note", "inline "+syntheticCanary)),
		slog.Group("\u202e", slog.String("sanitized_empty_note", "hidden "+syntheticCanary)),
	).WithGroup("token")
	registry.Register(syntheticCanary)
	bound.Info("late-bound attrs", slog.String("child", "sensitive-group-child-value"))

	out := buf.String()
	assertDoesNotContainCanary(t, out)
	if strings.Contains(out, "sensitive-group-child-value") {
		t.Fatalf("sensitive group child value was not redacted")
	}
}

func TestRedactingHandlerScrubsTypedURLsAndEncodedValues(t *testing.T) {
	registry := NewValueRegistry()
	registry.Register(syntheticCanary)

	escapedPath := url.PathEscape(syntheticCanary)
	escapedQuery := url.QueryEscape(syntheticCanary)
	typed, err := url.Parse("https://user:" + syntheticCanary + "@" + syntheticCanary + ".example.test/path/" + escapedPath + "?token=" + escapedQuery + "#frag")
	if err != nil {
		t.Fatalf("parse typed url: %v", err)
	}

	var buf bytes.Buffer
	logger := slog.New(NewRedactingHandler(slog.NewJSONHandler(&buf, nil), RedactionOptions{Registry: registry}))
	logger.Info("typed urls", "typed", *typed, "ptr", typed)

	out := buf.String()
	assertDoesNotContainCanary(t, out)
	if strings.Contains(out, escapedPath) || strings.Contains(out, escapedQuery) {
		t.Fatalf("encoded registered value was not redacted")
	}
	for _, forbidden := range []string{"user:", "?token=", "#frag"} {
		if strings.Contains(out, forbidden) {
			t.Fatalf("typed url retained forbidden component %q", forbidden)
		}
	}
}

func TestValueRegistryBoundsEntriesWithoutClearingPlaintext(t *testing.T) {
	registry := NewValueRegistryWithLimit(2)
	registry.Register("first-redaction-value")
	registry.Register("second-redaction-value")
	registry.Register("third-redaction-value")

	if strings.Contains(registry.redactString("first-redaction-value"), redactedValue) {
		t.Fatalf("oldest registry value was not evicted")
	}
	for _, value := range []string{"second-redaction-value", "third-redaction-value"} {
		if !strings.Contains(registry.redactString(value), redactedValue) {
			t.Fatalf("recent registry value was not redacted")
		}
	}
}

func TestRunLoggerRoutesByRunIDFansOutWarnsAndCloses(t *testing.T) {
	projectDir := t.TempDir()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	logger, closeLogs := NewLogger(projectDir, &stderr, LoggerOptions{MaxLogFiles: 5})
	ctx := WithRunID(context.Background(), "run_safe-1")

	logger.InfoContext(ctx, "info event")
	logger.WarnContext(ctx, "warn event")
	if err := closeLogs(); err != nil {
		t.Fatalf("close logs: %v", err)
	}

	if stdout.Len() != 0 {
		t.Fatalf("logger wrote to stdout")
	}
	if strings.Contains(stderr.String(), "info event") {
		t.Fatalf("stderr received info-level log")
	}
	if !strings.Contains(stderr.String(), "warn event") {
		t.Fatalf("stderr missing warn-level log")
	}

	logPath := filepath.Join(projectDir, "logs", "run_safe-1.jsonl")
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read run log: %v", err)
	}
	if !bytes.Contains(data, []byte("info event")) || !bytes.Contains(data, []byte("warn event")) {
		t.Fatalf("run log missing routed records: %s", string(data))
	}
}

func TestRunFileHandlerRejectsUnsafeRunIDsAndUsesSafePermissions(t *testing.T) {
	projectDir := t.TempDir()
	handler := NewRunFileHandler(projectDir, RunFileOptions{MaxFiles: 5})
	logger := slog.New(handler)

	logger.InfoContext(WithRunID(context.Background(), "run_safe-2"), "safe")
	logger.InfoContext(WithRunID(context.Background(), "../escape"), "unsafe")
	logger.InfoContext(WithRunID(context.Background(), "bad\nrun"), "unsafe")
	if err := handler.Close(); err != nil {
		t.Fatalf("close handler: %v", err)
	}

	logsDir := filepath.Join(projectDir, "logs")
	info, err := os.Stat(logsDir)
	if err != nil {
		t.Fatalf("stat logs dir: %v", err)
	}
	if got := info.Mode().Perm(); got != 0o700 {
		t.Fatalf("logs dir mode = %o, want 0700", got)
	}
	logPath := filepath.Join(logsDir, "run_safe-2.jsonl")
	info, err = os.Stat(logPath)
	if err != nil {
		t.Fatalf("stat log file: %v", err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("log file mode = %o, want 0600", got)
	}
	for _, rel := range []string{"escape.jsonl", "bad\nrun.jsonl"} {
		if _, err := os.Stat(filepath.Join(logsDir, rel)); err == nil {
			t.Fatalf("unsafe run id created log path %q", rel)
		}
	}
}

func TestRunFileHandlerPrunesRetention(t *testing.T) {
	projectDir := t.TempDir()
	logsDir := filepath.Join(projectDir, "logs")
	if err := os.MkdirAll(logsDir, 0o700); err != nil {
		t.Fatalf("mkdir logs: %v", err)
	}
	unrelated := filepath.Join(logsDir, "manual.jsonl")
	if err := os.WriteFile(unrelated, []byte("keep\n"), 0o600); err != nil {
		t.Fatalf("write unrelated log: %v", err)
	}

	handler := NewRunFileHandler(projectDir, RunFileOptions{MaxFiles: 2})
	logger := slog.New(handler)
	for _, id := range []string{"run_old", "run_mid", "run_new"} {
		logger.InfoContext(WithRunID(context.Background(), id), "event")
		time.Sleep(10 * time.Millisecond)
	}
	if err := handler.Close(); err != nil {
		t.Fatalf("close handler: %v", err)
	}

	if _, err := os.Stat(unrelated); err != nil {
		t.Fatalf("retention pruned unrelated jsonl: %v", err)
	}
	matches, err := filepath.Glob(filepath.Join(projectDir, "logs", "run_*.jsonl"))
	if err != nil {
		t.Fatalf("glob logs: %v", err)
	}
	if len(matches) != 2 {
		t.Fatalf("retention kept %d owned files, want 2 (%v)", len(matches), matches)
	}
	for _, path := range matches {
		if filepath.Base(path) == "run_old.jsonl" {
			t.Fatalf("retention kept oldest log")
		}
	}
}

func TestRunFileHandlerBlocksSymlinkEscape(t *testing.T) {
	projectDir := t.TempDir()
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(projectDir, "logs")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	handler := NewRunFileHandler(projectDir, RunFileOptions{MaxFiles: 5})
	logger := slog.New(handler)
	logger.InfoContext(WithRunID(context.Background(), "run_escape"), "should not escape")
	if err := handler.Close(); err != nil {
		t.Fatalf("close handler: %v", err)
	}

	if _, err := os.Stat(filepath.Join(outside, "run_escape.jsonl")); err == nil {
		t.Fatalf("log handler followed logs symlink outside project")
	} else if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("stat outside log: %v", err)
	}
}

func TestRunFileHandlerRejectsPolymetricsRootSymlink(t *testing.T) {
	parent := t.TempDir()
	outside := t.TempDir()
	projectDir := filepath.Join(parent, ".polymetrics")
	if err := os.Symlink(outside, projectDir); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	handler := NewRunFileHandler(projectDir, RunFileOptions{MaxFiles: 5})
	logger := slog.New(handler)
	logger.InfoContext(WithRunID(context.Background(), "run_escape"), "should not escape")
	if err := handler.Close(); err != nil {
		t.Fatalf("close handler: %v", err)
	}

	if _, err := os.Stat(filepath.Join(outside, "logs", "run_escape.jsonl")); err == nil {
		t.Fatalf("log handler followed .polymetrics symlink outside project")
	} else if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("stat outside log: %v", err)
	}
}

func TestTemporalStructuredLoggerUsesContextRedactingLogger(t *testing.T) {
	registry := NewValueRegistry()
	registry.Register(syntheticCanary)

	var buf bytes.Buffer
	logger := slog.New(NewRedactingHandler(slog.NewJSONHandler(&buf, nil), RedactionOptions{Registry: registry}))
	temporal := tlog.NewStructuredLogger(FromContext(WithLogger(context.Background(), logger)))
	temporal.Warn("temporal warning", "token", syntheticCanary)

	out := buf.String()
	assertDoesNotContainCanary(t, out)
	if !strings.Contains(out, "temporal warning") {
		t.Fatalf("temporal logger did not write through slog: %s", out)
	}
}

func TestTemporalStructuredLoggerWithBoundRunIDRoutesWhenAdapterDropsContext(t *testing.T) {
	projectDir := t.TempDir()
	var stderr bytes.Buffer
	logger, closeLogs := NewLogger(projectDir, &stderr, LoggerOptions{MaxLogFiles: 5})
	ctx := WithLogger(context.Background(), logger)
	ctx = WithRunID(ctx, "run_temporal_bound")
	temporal := tlog.NewStructuredLogger(FromContext(ctx))
	temporal.Warn("temporal warning without slog context")
	if err := closeLogs(); err != nil {
		t.Fatalf("close logs: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(projectDir, "logs", "run_temporal_bound.jsonl"))
	if err != nil {
		t.Fatalf("read bound temporal log: %v", err)
	}
	if !bytes.Contains(data, []byte("temporal warning without slog context")) {
		t.Fatalf("bound temporal logger did not route to run log")
	}
}

func assertDoesNotContainCanary(t *testing.T, text string) {
	t.Helper()
	if strings.Contains(text, syntheticCanary) {
		t.Fatalf("output contained synthetic canary")
	}
}

var _ io.Writer = (*bytes.Buffer)(nil)
