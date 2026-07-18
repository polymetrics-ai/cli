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
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	pmlogging "polymetrics.ai/internal/logging"
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
	captureDefault         = "default"
	captureMinimal         = "minimal"
	defaultTimeout         = 3 * time.Second
	defaultOTLPEndpointURL = "http://localhost:4318/v1/traces"
)

// WarningFunc receives sanitized telemetry warnings. Callers should write them to stderr.
type WarningFunc func(msg string)

// Config controls one invocation-scoped telemetry initialization.
type Config struct {
	Exporter        Exporter
	Endpoint        string
	Directory       string
	ProjectRoot     string
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
	addEvent      func(string, []attribute.KeyValue)
	setAttributes func(...attribute.KeyValue)
	setStatus     func(codes.Code, string)
	isRecording   func() bool
}

func (s otelSpan) End() { s.end() }

func (s otelSpan) AddEvent(name string, attrs ...Attr) {
	if name == "" {
		return
	}
	s.addEvent(sanitizeLine(s.ctx, name), filterAttrs(s.ctx, attrs))
}

func (s otelSpan) SetAttributes(attrs ...Attr) {
	s.setAttributes(filterAttrs(s.ctx, attrs)...)
}

func (s otelSpan) RecordError(err error) {
	if err == nil {
		return
	}
	attrs := errorAttrs(s.ctx, err)
	if len(attrs) > 0 {
		s.setAttributes(attrs...)
		s.addEvent("pm.error", attrs)
	}
	s.setStatus(codes.Error, "error")
}

func (s otelSpan) SetStatus(status string) {
	if strings.EqualFold(status, "error") || strings.EqualFold(status, "failed") {
		s.setStatus(codes.Error, "error")
		return
	}
	s.setStatus(codes.Ok, sanitizeLine(s.ctx, status))
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

	exporter, file, err := newExporter(ctx, cfg, warn)
	if err != nil {
		warnf(ctx, warn, "initialize telemetry exporter: %v", err)
		return ctx, &Handle{timeout: timeoutOrDefault(cfg.ShutdownTimeout)}
	}

	provider, err := newTracerProvider(ctx, exporter, cfg, warn)
	if err != nil {
		_ = exporter.Shutdown(ctx)
		if file != nil {
			_ = file.Close()
		}
		warnf(ctx, warn, "initialize telemetry provider: %v", err)
		return ctx, &Handle{timeout: timeoutOrDefault(cfg.ShutdownTimeout)}
	}
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
			addEvent: func(name string, attrs []attribute.KeyValue) {
				if len(attrs) == 0 {
					span.AddEvent(name)
					return
				}
				span.AddEvent(name, trace.WithAttributes(attrs...))
			},
			setAttributes: span.SetAttributes,
			setStatus:     span.SetStatus,
			isRecording:   span.IsRecording,
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
		warnf(ctx, warn, "telemetry shutdown: %v", err)
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

func newTracerProvider(ctx context.Context, exporter sdktrace.SpanExporter, cfg Config, warn WarningFunc) (*sdktrace.TracerProvider, error) {
	warnUnsupportedSDKEnv(ctx, warn)
	var provider *sdktrace.TracerProvider
	err := withSanitizedSDKEnv(func() error {
		provider = sdktrace.NewTracerProvider(
			sdktrace.WithSyncer(exporter),
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
			sdktrace.WithResource(safeResource(cfg)),
			sdktrace.WithRawSpanLimits(defaultSpanLimits()),
		)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return provider, nil
}

func safeResource(cfg Config) *resource.Resource {
	serviceName := safeFileName(cfg.ServiceName)
	if serviceName == "" {
		serviceName = "pm"
	}
	return resource.NewSchemaless(attribute.String("service.name", serviceName))
}

func defaultSpanLimits() sdktrace.SpanLimits {
	return sdktrace.SpanLimits{
		AttributeValueLengthLimit:   sdktrace.DefaultAttributeValueLengthLimit,
		AttributeCountLimit:         sdktrace.DefaultAttributeCountLimit,
		EventCountLimit:             sdktrace.DefaultEventCountLimit,
		LinkCountLimit:              sdktrace.DefaultLinkCountLimit,
		AttributePerEventCountLimit: sdktrace.DefaultAttributePerEventCountLimit,
		AttributePerLinkCountLimit:  sdktrace.DefaultAttributePerLinkCountLimit,
	}
}

func newExporter(ctx context.Context, cfg Config, warn WarningFunc) (sdktrace.SpanExporter, io.Closer, error) {
	switch normalizeExporter(cfg.Exporter) {
	case ExporterFile:
		exporter, file, err := newFileExporter(cfg)
		if err != nil {
			return nil, nil, err
		}
		return warningExporter{ctx: ctx, warn: warn, next: exporter}, file, nil
	case ExporterOTLP:
		exporter, err := newOTLPExporter(ctx, cfg, warn)
		if err != nil {
			return nil, nil, err
		}
		return warningExporter{ctx: ctx, warn: warn, next: exporter}, nil, nil
	default:
		return nil, nil, fmt.Errorf("unsupported exporter %q", cfg.Exporter)
	}
}

func newOTLPExporter(ctx context.Context, cfg Config, warn WarningFunc) (sdktrace.SpanExporter, error) {
	endpoint := strings.TrimSpace(cfg.Endpoint)
	if endpoint == "" {
		endpoint = defaultOTLPEndpointURL
	}
	validated, err := validateOTLPEndpoint(endpoint)
	if err != nil {
		return nil, err
	}
	warnUnsupportedOTLPEnv(ctx, warn)
	opts := []otlptracehttp.Option{
		otlptracehttp.WithTimeout(timeoutOrDefault(cfg.ShutdownTimeout)),
		otlptracehttp.WithEndpointURL(validated),
		otlptracehttp.WithHeaders(map[string]string{}),
	}
	return withSanitizedOTLPEnv(func() (sdktrace.SpanExporter, error) {
		return otlptracehttp.New(ctx, opts...)
	})
}

func newFileExporter(cfg Config) (sdktrace.SpanExporter, io.Closer, error) {
	dir, err := resolveTelemetryDir(cfg.ProjectRoot, cfg.Directory)
	if err != nil {
		return nil, nil, err
	}
	if err := mkdirAllNoSymlink(dir.root, dir.path); err != nil {
		return nil, nil, err
	}
	name := safeFileName(cfg.RunID)
	if name == "" {
		name = time.Now().UTC().Format("20060102T150405.000000000Z")
	}
	filePath := filepath.Join(dir.path, name+".jsonl")
	if err := rejectSymlink(filePath); err != nil {
		return nil, nil, err
	}
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
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

type warningExporter struct {
	ctx  context.Context
	warn WarningFunc
	next sdktrace.SpanExporter
}

func (e warningExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	if e.next == nil {
		return nil
	}
	if err := e.next.ExportSpans(ctx, spans); err != nil {
		warnf(e.ctx, e.warn, "telemetry export failed")
	}
	return nil
}

func (e warningExporter) Shutdown(ctx context.Context) error {
	if e.next == nil {
		return nil
	}
	if err := e.next.Shutdown(ctx); err != nil {
		warnf(e.ctx, e.warn, "telemetry exporter shutdown failed")
	}
	return nil
}

type telemetryDir struct {
	root string
	path string
}

func resolveTelemetryDir(projectRoot, directory string) (telemetryDir, error) {
	root := strings.TrimSpace(projectRoot)
	if root == "" {
		root = "."
	}
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return telemetryDir{}, fmt.Errorf("resolve project root: %w", err)
	}
	dir := strings.TrimSpace(directory)
	if dir == "" {
		dir = filepath.Join(".polymetrics", "telemetry")
	}
	if filepath.IsAbs(dir) || !filepath.IsLocal(dir) {
		return telemetryDir{}, errors.New("telemetry directory must be relative to the project root")
	}
	clean := filepath.Clean(dir)
	pathAbs := filepath.Join(rootAbs, clean)
	rel, err := filepath.Rel(rootAbs, pathAbs)
	if err != nil {
		return telemetryDir{}, fmt.Errorf("compare telemetry directory to project root: %w", err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return telemetryDir{}, errors.New("telemetry directory must stay under the project root")
	}
	return telemetryDir{root: rootAbs, path: pathAbs}, nil
}

func mkdirAllNoSymlink(root, target string) error {
	info, err := os.Lstat(root)
	if err != nil {
		return fmt.Errorf("stat project root: %w", err)
	}
	if !info.IsDir() {
		return errors.New("project root is not a directory")
	}
	rel, err := filepath.Rel(root, target)
	if err != nil {
		return fmt.Errorf("compare telemetry directory to project root: %w", err)
	}
	current := root
	for _, part := range strings.Split(rel, string(filepath.Separator)) {
		if part == "" || part == "." {
			continue
		}
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if errors.Is(err, os.ErrNotExist) {
			if err := os.Mkdir(current, 0o700); err != nil {
				return fmt.Errorf("create telemetry directory: %w", err)
			}
			info, err = os.Lstat(current)
		}
		if err != nil {
			return fmt.Errorf("stat telemetry directory: %w", err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return errors.New("telemetry directory must not contain symlinks")
		}
		if !info.IsDir() {
			return errors.New("telemetry directory path contains a non-directory")
		}
	}
	return nil
}

func rejectSymlink(path string) error {
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("stat telemetry file: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return errors.New("telemetry file must not be a symlink")
	}
	return errors.New("telemetry file already exists")
}

var (
	otlpEnvMu = sync.Mutex{}

	supportedOTLPEndpointEnv = []string{
		"OTEL_EXPORTER_OTLP_ENDPOINT",
		"OTEL_EXPORTER_OTLP_TRACES_ENDPOINT",
	}

	unsupportedSDKEnv = []string{
		"OTEL_RESOURCE_ATTRIBUTES",
		"OTEL_SERVICE_NAME",
		"OTEL_TRACES_SAMPLER",
		"OTEL_TRACES_SAMPLER_ARG",
		"OTEL_ATTRIBUTE_VALUE_LENGTH_LIMIT",
		"OTEL_ATTRIBUTE_COUNT_LIMIT",
		"OTEL_SPAN_ATTRIBUTE_VALUE_LENGTH_LIMIT",
		"OTEL_SPAN_ATTRIBUTE_COUNT_LIMIT",
		"OTEL_SPAN_EVENT_COUNT_LIMIT",
		"OTEL_EVENT_ATTRIBUTE_COUNT_LIMIT",
		"OTEL_SPAN_LINK_COUNT_LIMIT",
		"OTEL_LINK_ATTRIBUTE_COUNT_LIMIT",
		"OTEL_BSP_SCHEDULE_DELAY",
		"OTEL_BSP_EXPORT_TIMEOUT",
		"OTEL_BSP_MAX_QUEUE_SIZE",
		"OTEL_BSP_MAX_EXPORT_BATCH_SIZE",
		"OTEL_GO_X_RESOURCE",
		"OTEL_GO_X_OBSERVABILITY",
	}

	unsupportedOTLPEnv = []string{
		"OTEL_EXPORTER_OTLP_HEADERS",
		"OTEL_EXPORTER_OTLP_TRACES_HEADERS",
		"OTEL_EXPORTER_OTLP_CERTIFICATE",
		"OTEL_EXPORTER_OTLP_TRACES_CERTIFICATE",
		"OTEL_EXPORTER_OTLP_CLIENT_CERTIFICATE",
		"OTEL_EXPORTER_OTLP_CLIENT_KEY",
		"OTEL_EXPORTER_OTLP_TRACES_CLIENT_CERTIFICATE",
		"OTEL_EXPORTER_OTLP_TRACES_CLIENT_KEY",
		"OTEL_EXPORTER_OTLP_INSECURE",
		"OTEL_EXPORTER_OTLP_TRACES_INSECURE",
		"OTEL_EXPORTER_OTLP_COMPRESSION",
		"OTEL_EXPORTER_OTLP_TRACES_COMPRESSION",
		"OTEL_EXPORTER_OTLP_PROTOCOL",
		"OTEL_EXPORTER_OTLP_TRACES_PROTOCOL",
		"OTEL_EXPORTER_OTLP_TIMEOUT",
		"OTEL_EXPORTER_OTLP_TRACES_TIMEOUT",
	}
)

type savedEnvValue struct {
	value string
	ok    bool
}

func warnUnsupportedOTLPEnv(ctx context.Context, warn WarningFunc) {
	for _, name := range unsupportedOTLPEnv {
		if _, ok := os.LookupEnv(name); ok {
			warnf(ctx, warn, "unsupported OTLP environment variable %s ignored; configure only telemetry exporter and endpoint through trusted env/flag", name)
		}
	}
}

func warnUnsupportedSDKEnv(ctx context.Context, warn WarningFunc) {
	for _, name := range unsupportedSDKEnv {
		if _, ok := os.LookupEnv(name); ok {
			warnf(ctx, warn, "unsupported OpenTelemetry SDK environment variable %s ignored; configure telemetry only through trusted pm env/flag", name)
		}
	}
}

func withSanitizedSDKEnv(fn func() error) error {
	return withSanitizedEnv(unsupportedSDKEnv, fn)
}

func withSanitizedOTLPEnv(fn func() (sdktrace.SpanExporter, error)) (sdktrace.SpanExporter, error) {
	all := append([]string{}, supportedOTLPEndpointEnv...)
	all = append(all, unsupportedOTLPEnv...)
	var exporter sdktrace.SpanExporter
	err := withSanitizedEnv(all, func() error {
		var err error
		exporter, err = fn()
		return err
	})
	return exporter, err
}

func withSanitizedEnv(names []string, fn func() error) error {
	otlpEnvMu.Lock()
	defer otlpEnvMu.Unlock()

	saved := make(map[string]savedEnvValue, len(names))
	defer restoreEnv(saved)
	for _, name := range names {
		value, ok := os.LookupEnv(name)
		saved[name] = savedEnvValue{value: value, ok: ok}
		if err := os.Unsetenv(name); err != nil {
			return fmt.Errorf("sanitize OpenTelemetry environment: %w", err)
		}
	}
	return fn()
}

func restoreEnv(saved map[string]savedEnvValue) {
	for name, item := range saved {
		if item.ok {
			_ = os.Setenv(name, item.value)
			continue
		}
		_ = os.Unsetenv(name)
	}
}

func validateOTLPEndpoint(endpoint string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(endpoint))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", errors.New("invalid OTLP endpoint: endpoint must be an http or https URL without credentials, query, or fragment")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", errors.New("invalid OTLP endpoint: endpoint must be an http or https URL without credentials, query, or fragment")
	}
	if parsed.User != nil || parsed.RawQuery != "" || parsed.ForceQuery || parsed.Fragment != "" {
		return "", errors.New("invalid OTLP endpoint: endpoint must be an http or https URL without credentials, query, or fragment")
	}
	return parsed.String(), nil
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
			out = append(out, key.String(sanitizeLine(ctx, value)))
		case int:
			out = append(out, key.Int(value))
		case int64:
			out = append(out, key.Int64(value))
		case bool:
			out = append(out, key.Bool(value))
		default:
			out = append(out, key.String(sanitizeLine(ctx, fmt.Sprint(value))))
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
	"pm.error.type":         {},
	"pm.error.code":         {},
	"pm.error.status_code":  {},
}

func errorAttrs(ctx context.Context, err error) []attribute.KeyValue {
	if err == nil {
		return nil
	}
	attrs := []Attr{StringAttr("pm.error.type", errorType(err))}
	if code := errorCode(err); code != "" {
		attrs = append(attrs, StringAttr("pm.error.code", code))
	}
	if status := errorStatusCode(err); status > 0 {
		attrs = append(attrs, IntAttr("pm.error.status_code", status))
	}
	return filterAttrs(ctx, attrs)
}

func errorType(err error) string {
	if err == nil {
		return "error"
	}
	var classifier errorClassifier
	if errors.As(err, &classifier) {
		if class := safeErrorToken(classifier.TelemetryErrorClass()); class != "" {
			return class
		}
	}
	if errors.Is(err, context.Canceled) {
		return "context.Canceled"
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "context.DeadlineExceeded"
	}
	if errorStatusCode(err) > 0 {
		return "http"
	}
	return "error"
}

func errorCode(err error) string {
	var coder errorCoder
	if errors.As(err, &coder) {
		if code := safeErrorToken(coder.TelemetryErrorCode()); code != "" {
			return code
		}
	}
	switch {
	case errors.Is(err, context.Canceled):
		return "context_canceled"
	case errors.Is(err, context.DeadlineExceeded):
		return "context_deadline_exceeded"
	case errorStatusCode(err) > 0:
		return "http_status"
	default:
		return "error"
	}
}

func safeErrorToken(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > 80 {
		return ""
	}
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '_' || r == '-' || r == '.' || r == ':':
		default:
			return ""
		}
	}
	return value
}

type errorClassifier interface {
	TelemetryErrorClass() string
}

type errorCoder interface {
	TelemetryErrorCode() string
}

type statusCoder interface {
	StatusCode() int
}

func errorStatusCode(err error) int {
	var status statusCoder
	if errors.As(err, &status) {
		return status.StatusCode()
	}
	return 0
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
	case "", "none", "off":
		return ExporterNone
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

func warnf(ctx context.Context, warn WarningFunc, format string, args ...any) {
	if warn == nil {
		return
	}
	warn(sanitizeLine(ctx, fmt.Sprintf(format, args...)))
}

func sanitizeLine(ctx context.Context, value string) string {
	return pmlogging.RedactLine(ctx, value)
}
