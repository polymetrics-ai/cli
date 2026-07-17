// Package telemetry provides opt-in OpenTelemetry tracing for the pm CLI.
package telemetry

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"polymetrics.ai/internal/safety"
)

// Exporter names the tracing exporter mode.
type Exporter string

const (
	// ExporterNone disables tracing. No SDK is constructed and no telemetry directory is created.
	ExporterNone Exporter = "none"
	// ExporterFile writes stdouttrace JSONL files under Config.Directory.
	ExporterFile Exporter = "file"
	// ExporterOTLP exports traces through OTLP HTTP/protobuf.
	ExporterOTLP Exporter = "otlp"
)

const (
	captureDefault = "default"
	captureMinimal = "minimal"
	defaultTimeout = 3 * time.Second
)

// WarningFunc receives sanitized telemetry warnings. Callers should write them to stderr.
type WarningFunc func(msg string)

// Config controls one invocation-scoped telemetry initialization.
type Config struct {
	Exporter        Exporter
	Endpoint        string
	Directory       string
	Capture         string
	RunID           string
	ServiceName     string
	ShutdownTimeout time.Duration
}

// Attr is a telemetry attribute candidate. Only an internal allowlist reaches exporters.
type Attr struct {
	Key   string
	Value any
}

// StringAttr returns a string attribute candidate.
func StringAttr(key, value string) Attr { return Attr{Key: key, Value: value} }

// IntAttr returns an integer attribute candidate.
func IntAttr(key string, value int) Attr { return Attr{Key: key, Value: value} }

// Int64Attr returns an int64 attribute candidate.
func Int64Attr(key string, value int64) Attr { return Attr{Key: key, Value: value} }

// BoolAttr returns a bool attribute candidate.
func BoolAttr(key string, value bool) Attr { return Attr{Key: key, Value: value} }

// Handle owns the invocation-scoped tracer provider and exporter resources.
type Handle struct {
	enabled  bool
	provider *sdktrace.TracerProvider
	file     io.Closer
	shutdown func(context.Context) error
	timeout  time.Duration
}

// Enabled reports whether tracing is active for this invocation.
func (h *Handle) Enabled() bool { return h != nil && h.enabled }

type tracerContextKey struct{}
type captureContextKey struct{}
type startFunc func(context.Context, string, []attribute.KeyValue) (context.Context, Span)

type noopSpan struct{}

func (noopSpan) End()                     {}
func (noopSpan) AddEvent(string, ...Attr) {}
func (noopSpan) SetAttributes(...Attr)    {}
func (noopSpan) RecordError(error)        {}
func (noopSpan) SetStatus(string)         {}
func (noopSpan) IsRecording() bool        { return false }

type Span interface {
	End()
	AddEvent(name string, attrs ...Attr)
	SetAttributes(attrs ...Attr)
	RecordError(err error)
	SetStatus(status string)
	IsRecording() bool
}

type otelSpan struct {
	ctx           context.Context
	end           func()
	addEvent      func(string)
	setAttributes func(...attribute.KeyValue)
	recordError   func(error)
	setStatus     func(codes.Code, string)
	isRecording   func() bool
}

func (s otelSpan) End() { s.end() }

func (s otelSpan) AddEvent(name string, attrs ...Attr) {
	if name == "" {
		return
	}
	if filtered := filterAttrs(s.ctx, attrs); len(filtered) > 0 {
		s.setAttributes(filtered...)
	}
	s.addEvent(sanitizeString(name))
}

func (s otelSpan) SetAttributes(attrs ...Attr) {
	s.setAttributes(filterAttrs(s.ctx, attrs)...)
}

func (s otelSpan) RecordError(err error) {
	if err == nil {
		return
	}
	s.recordError(errors.New(sanitizeString(err.Error())))
	s.setStatus(codes.Error, "error")
}

func (s otelSpan) SetStatus(status string) {
	if strings.EqualFold(status, "error") || strings.EqualFold(status, "failed") {
		s.setStatus(codes.Error, "error")
		return
	}
	s.setStatus(codes.Ok, sanitizeString(status))
}

func (s otelSpan) IsRecording() bool { return s.isRecording() }

// Init creates invocation-scoped tracing state. Disabled modes return the input
// context and a disabled handle without constructing an SDK provider.
func Init(ctx context.Context, cfg Config, warn WarningFunc) (context.Context, *Handle) {
	if ctx == nil {
		ctx = context.Background()
	}
	if sdkDisabled() || !exporterEnabled(cfg.Exporter) {
		return ctx, &Handle{timeout: timeoutOrDefault(cfg.ShutdownTimeout)}
	}

	exporter, file, err := newExporter(ctx, cfg)
	if err != nil {
		warnf(warn, "initialize telemetry exporter: %v", err)
		return ctx, &Handle{timeout: timeoutOrDefault(cfg.ShutdownTimeout)}
	}

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exporter),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
	handle := &Handle{
		enabled:  true,
		provider: provider,
		file:     file,
		timeout:  timeoutOrDefault(cfg.ShutdownTimeout),
	}
	handle.shutdown = func(shutdownCtx context.Context) error {
		var errs []error
		if err := provider.Shutdown(shutdownCtx); err != nil {
			errs = append(errs, err)
		}
		if file != nil {
			if err := file.Close(); err != nil {
				errs = append(errs, err)
			}
		}
		return errors.Join(errs...)
	}

	tracer := provider.Tracer("polymetrics.ai/pm")
	starter := startFunc(func(parent context.Context, name string, attrs []attribute.KeyValue) (context.Context, Span) {
		spanCtx, span := tracer.Start(parent, name)
		if len(attrs) > 0 {
			span.SetAttributes(attrs...)
		}
		return spanCtx, otelSpan{
			ctx: spanCtx,
			end: func() {
				span.End()
			},
			addEvent: func(name string) {
				span.AddEvent(name)
			},
			setAttributes: span.SetAttributes,
			recordError: func(err error) {
				span.RecordError(err)
			},
			setStatus:   span.SetStatus,
			isRecording: span.IsRecording,
		}
	})
	ctx = context.WithValue(ctx, tracerContextKey{}, starter)
	ctx = context.WithValue(ctx, captureContextKey{}, normalizeCapture(cfg.Capture))
	return ctx, handle
}

// Shutdown flushes and closes tracing resources, warning without returning an error.
func Shutdown(ctx context.Context, handle *Handle, warn WarningFunc) {
	if handle == nil || handle.shutdown == nil {
		return
	}
	timeout := timeoutOrDefault(handle.timeout)
	if ctx == nil {
		ctx = context.Background()
	}
	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), timeout)
	defer cancel()
	if err := handle.shutdown(shutdownCtx); err != nil {
		warnf(warn, "telemetry shutdown: %v", err)
	}
}

// StartSpan starts a span when tracing is enabled in ctx, otherwise returns a no-op span.
func StartSpan(ctx context.Context, name string, attrs ...Attr) (context.Context, Span) {
	if ctx == nil {
		ctx = context.Background()
	}
	starter, ok := ctx.Value(tracerContextKey{}).(startFunc)
	if !ok || starter == nil || name == "" {
		return ctx, noopSpan{}
	}
	return starter(ctx, name, filterAttrs(ctx, attrs))
}

// HTTPAttrs returns allowlisted HTTP metadata with query strings, userinfo,
// fragments, bodies, and headers deliberately omitted.
func HTTPAttrs(method, rawURL string) []Attr {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return []Attr{StringAttr("pm.http.method", strings.ToUpper(method))}
	}
	return []Attr{
		StringAttr("pm.http.method", strings.ToUpper(method)),
		StringAttr("pm.http.scheme", parsed.Scheme),
		StringAttr("pm.http.host", parsed.Host),
		StringAttr("pm.http.path", parsed.EscapedPath()),
	}
}

func newExporter(ctx context.Context, cfg Config) (sdktrace.SpanExporter, io.Closer, error) {
	switch normalizeExporter(cfg.Exporter) {
	case ExporterFile:
		return newFileExporter(cfg)
	case ExporterOTLP:
		opts := []otlptracehttp.Option{}
		if endpoint := strings.TrimSpace(cfg.Endpoint); endpoint != "" {
			if strings.Contains(endpoint, "://") {
				opts = append(opts, otlptracehttp.WithEndpointURL(endpoint))
			} else {
				opts = append(opts, otlptracehttp.WithEndpoint(endpoint))
			}
		}
		exporter, err := otlptracehttp.New(ctx, opts...)
		return exporter, nil, err
	default:
		return nil, nil, fmt.Errorf("unsupported exporter %q", cfg.Exporter)
	}
}

func newFileExporter(cfg Config) (sdktrace.SpanExporter, io.Closer, error) {
	dir := strings.TrimSpace(cfg.Directory)
	if dir == "" {
		dir = filepath.Join(".polymetrics", "telemetry")
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, nil, err
	}
	name := safeFileName(cfg.RunID)
	if name == "" {
		name = time.Now().UTC().Format("20060102T150405.000000000Z")
	}
	file, err := os.OpenFile(filepath.Join(dir, name+".jsonl"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, nil, err
	}
	exporter, err := stdouttrace.New(stdouttrace.WithWriter(file), stdouttrace.WithoutTimestamps())
	if err != nil {
		_ = file.Close()
		return nil, nil, err
	}
	return exporter, file, nil
}

var safeNamePattern = regexp.MustCompile(`[^A-Za-z0-9._-]+`)

func safeFileName(value string) string {
	value = strings.TrimSpace(value)
	value = safeNamePattern.ReplaceAllString(value, "-")
	value = strings.Trim(value, ".-")
	if len(value) > 80 {
		value = value[:80]
	}
	return value
}

func filterAttrs(ctx context.Context, attrs []Attr) []attribute.KeyValue {
	if len(attrs) == 0 || captureMode(ctx) == captureMinimal {
		return nil
	}
	out := make([]attribute.KeyValue, 0, len(attrs))
	for _, attr := range attrs {
		if _, ok := allowedAttributeKeys[attr.Key]; !ok {
			continue
		}
		key := attribute.Key(attr.Key)
		switch value := attr.Value.(type) {
		case string:
			out = append(out, key.String(sanitizeString(value)))
		case int:
			out = append(out, key.Int(value))
		case int64:
			out = append(out, key.Int64(value))
		case bool:
			out = append(out, key.Bool(value))
		default:
			out = append(out, key.String(sanitizeString(fmt.Sprint(value))))
		}
	}
	return out
}

var allowedAttributeKeys = map[string]struct{}{
	"pm.command.name":       {},
	"pm.command.status":     {},
	"pm.command.exit_code":  {},
	"pm.etl.connection":     {},
	"pm.etl.stream":         {},
	"pm.etl.status":         {},
	"pm.etl.records_read":   {},
	"pm.etl.records_loaded": {},
	"pm.flow.name":          {},
	"pm.flow.step_id":       {},
	"pm.flow.step_kind":     {},
	"pm.flow.status":        {},
	"pm.certify.connector":  {},
	"pm.certify.mode":       {},
	"pm.certify.status":     {},
	"pm.http.method":        {},
	"pm.http.scheme":        {},
	"pm.http.host":          {},
	"pm.http.path":          {},
	"pm.http.status_code":   {},
	"pm.http.attempt":       {},
	"pm.http.max_attempts":  {},
	"pm.http.retry":         {},
}

func captureMode(ctx context.Context) string {
	if ctx == nil {
		return captureDefault
	}
	mode, _ := ctx.Value(captureContextKey{}).(string)
	return normalizeCapture(mode)
}

func normalizeCapture(mode string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case captureMinimal:
		return captureMinimal
	default:
		return captureDefault
	}
}

func exporterEnabled(exporter Exporter) bool {
	switch normalizeExporter(exporter) {
	case ExporterFile, ExporterOTLP:
		return true
	default:
		return false
	}
}

func normalizeExporter(exporter Exporter) Exporter {
	switch strings.ToLower(strings.TrimSpace(string(exporter))) {
	case "file":
		return ExporterFile
	case "otlp":
		return ExporterOTLP
	default:
		return ExporterNone
	}
}

func sdkDisabled() bool {
	value, ok := os.LookupEnv("OTEL_SDK_DISABLED")
	return ok && strings.EqualFold(strings.TrimSpace(value), "true")
}

func timeoutOrDefault(timeout time.Duration) time.Duration {
	if timeout <= 0 {
		return defaultTimeout
	}
	return timeout
}

func warnf(warn WarningFunc, format string, args ...any) {
	if warn == nil {
		return
	}
	warn(sanitizeString(fmt.Sprintf(format, args...)))
}

func sanitizeString(value string) string {
	return safety.SanitizeTerminalLine(safety.RedactErrorText(value))
}
