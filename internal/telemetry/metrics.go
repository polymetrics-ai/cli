package telemetry

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	otelmetric "go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

const defaultOTLPMetricEndpointURL = "http://localhost:4318/v1/metrics"

type metricContextKey struct{}

type metricHandle struct {
	provider *sdkmetric.MeterProvider
	reader   *sdkmetric.ManualReader
	exporter sdkmetric.Exporter
	file     io.Closer
	meter    otelmetric.Meter
}

type metricState struct {
	meter otelmetric.Meter
}

// Enabled reports whether telemetry is active for ctx.
func Enabled(ctx context.Context) bool {
	_, ok := metricStateFromContext(ctx)
	return ok
}

// Meter returns the invocation-scoped OpenTelemetry meter when telemetry is enabled.
func Meter(ctx context.Context) (otelmetric.Meter, bool) {
	state, ok := metricStateFromContext(ctx)
	if !ok || state.meter == nil {
		return nil, false
	}
	return state.meter, true
}

func metricStateFromContext(ctx context.Context) (metricState, bool) {
	if ctx == nil {
		return metricState{}, false
	}
	state, ok := ctx.Value(metricContextKey{}).(metricState)
	return state, ok
}

func withMetricState(ctx context.Context, metrics *metricHandle) context.Context {
	if metrics == nil || metrics.meter == nil {
		return ctx
	}
	return context.WithValue(ctx, metricContextKey{}, metricState{meter: metrics.meter})
}

func newMetricHandle(ctx context.Context, cfg Config, warn WarningFunc) (*metricHandle, error) {
	exporter, file, err := newMetricExporter(ctx, cfg, warn)
	if err != nil {
		return nil, err
	}
	reader := sdkmetric.NewManualReader(
		sdkmetric.WithTemporalitySelector(func(sdkmetric.InstrumentKind) metricdata.Temporality {
			return metricdata.CumulativeTemporality
		}),
	)
	var provider *sdkmetric.MeterProvider
	var meter otelmetric.Meter
	err = withSanitizedSDKEnv(func() error {
		provider = sdkmetric.NewMeterProvider(
			sdkmetric.WithReader(reader),
			sdkmetric.WithResource(safeResource(cfg)),
		)
		meter = provider.Meter("polymetrics.ai/pm")
		return nil
	})
	if err != nil {
		_ = exporter.Shutdown(ctx)
		if file != nil {
			_ = file.Close()
		}
		return nil, err
	}
	return &metricHandle{
		provider: provider,
		reader:   reader,
		exporter: metricWarningExporter{ctx: ctx, warn: warn, next: exporter},
		file:     file,
		meter:    meter,
	}, nil
}

func (h *metricHandle) shutdown(ctx context.Context) error {
	if h == nil {
		return nil
	}
	var errs []error
	if h.reader != nil && h.exporter != nil {
		var rm metricdata.ResourceMetrics
		if err := h.reader.Collect(ctx, &rm); err != nil {
			errs = append(errs, fmt.Errorf("collect metrics: %w", err))
		} else if len(rm.ScopeMetrics) > 0 {
			if err := h.exporter.Export(ctx, &rm); err != nil {
				errs = append(errs, fmt.Errorf("export metrics: %w", err))
			}
		}
	}
	if h.provider != nil {
		if err := h.provider.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("shutdown metrics provider: %w", err))
		}
	}
	if h.exporter != nil {
		if err := h.exporter.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("shutdown metrics exporter: %w", err))
		}
	}
	if h.file != nil {
		if err := h.file.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close metrics file: %w", err))
		}
	}
	return errors.Join(errs...)
}

func newMetricExporter(ctx context.Context, cfg Config, warn WarningFunc) (sdkmetric.Exporter, io.Closer, error) {
	switch normalizeExporter(cfg.Exporter) {
	case ExporterFile:
		exporter, file, err := newMetricFileExporter(cfg)
		if err != nil {
			return nil, nil, err
		}
		return exporter, file, nil
	case ExporterOTLP:
		exporter, err := newMetricOTLPExporter(ctx, cfg, warn)
		if err != nil {
			return nil, nil, err
		}
		return exporter, nil, nil
	default:
		return nil, nil, fmt.Errorf("unsupported exporter %q", cfg.Exporter)
	}
}

func newMetricFileExporter(cfg Config) (sdkmetric.Exporter, io.Closer, error) {
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
	filePath := filepath.Join(dir.path, "metrics-"+name+".jsonl")
	if err := rejectSymlink(filePath); err != nil {
		return nil, nil, err
	}
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, nil, err
	}
	var exporter sdkmetric.Exporter
	err = withSanitizedSDKEnv(func() error {
		var newErr error
		exporter, newErr = stdoutmetric.New(stdoutmetric.WithWriter(file), stdoutmetric.WithoutTimestamps())
		return newErr
	})
	if err != nil {
		_ = file.Close()
		return nil, nil, err
	}
	return exporter, file, nil
}

func newMetricOTLPExporter(ctx context.Context, cfg Config, warn WarningFunc) (sdkmetric.Exporter, error) {
	endpoint := strings.TrimSpace(metricEndpointFromEnv())
	if endpoint == "" {
		endpoint = metricEndpointFromTraceEndpoint(cfg.Endpoint)
	}
	if endpoint == "" {
		endpoint = defaultOTLPMetricEndpointURL
	}
	validated, err := validateOTLPEndpoint(endpoint)
	if err != nil {
		if name := metricEndpointEnvName(); name != "" {
			return nil, fmt.Errorf("invalid OTLP metrics endpoint from %s: endpoint must be an http or https URL without credentials, query, or fragment", name)
		}
		return nil, err
	}
	warnUnsupportedOTLPEnv(ctx, warn)
	opts := []otlpmetrichttp.Option{
		otlpmetrichttp.WithTimeout(timeoutOrDefault(cfg.ShutdownTimeout)),
		otlpmetrichttp.WithEndpointURL(validated),
		otlpmetrichttp.WithHeaders(map[string]string{}),
	}
	return withSanitizedOTLPMetricEnv(func() (sdkmetric.Exporter, error) {
		return otlpmetrichttp.New(ctx, opts...)
	})
}

func metricEndpointFromEnv() string {
	for _, name := range supportedOTLPMetricEndpointEnv {
		if value, ok := os.LookupEnv(name); ok && strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func metricEndpointEnvName() string {
	for _, name := range supportedOTLPMetricEndpointEnv {
		if value, ok := os.LookupEnv(name); ok && strings.TrimSpace(value) != "" {
			return name
		}
	}
	return ""
}

func metricEndpointFromTraceEndpoint(endpoint string) string {
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		return ""
	}
	parsed, err := urlParse(endpoint)
	if err != nil {
		return endpoint
	}
	if strings.HasSuffix(parsed.Path, "/v1/traces") {
		parsed.Path = strings.TrimSuffix(parsed.Path, "/v1/traces") + "/v1/metrics"
	}
	return parsed.String()
}

func urlParse(endpoint string) (*url.URL, error) {
	return url.Parse(endpoint)
}

type metricWarningExporter struct {
	ctx  context.Context
	warn WarningFunc
	next sdkmetric.Exporter
}

func (e metricWarningExporter) Temporality(kind sdkmetric.InstrumentKind) metricdata.Temporality {
	return e.next.Temporality(kind)
}

func (e metricWarningExporter) Aggregation(kind sdkmetric.InstrumentKind) sdkmetric.Aggregation {
	return e.next.Aggregation(kind)
}

func (e metricWarningExporter) Export(ctx context.Context, rm *metricdata.ResourceMetrics) error {
	if e.next == nil {
		return nil
	}
	if err := e.next.Export(ctx, rm); err != nil {
		warnf(e.ctx, e.warn, "telemetry metrics export failed")
	}
	return nil
}

func (e metricWarningExporter) ForceFlush(ctx context.Context) error {
	if e.next == nil {
		return nil
	}
	if err := e.next.ForceFlush(ctx); err != nil {
		warnf(e.ctx, e.warn, "telemetry metrics flush failed")
	}
	return nil
}

func (e metricWarningExporter) Shutdown(ctx context.Context) error {
	if e.next == nil {
		return nil
	}
	if err := e.next.Shutdown(ctx); err != nil {
		warnf(e.ctx, e.warn, "telemetry metrics exporter shutdown failed")
	}
	return nil
}

func withSanitizedOTLPMetricEnv(fn func() (sdkmetric.Exporter, error)) (sdkmetric.Exporter, error) {
	all := append([]string{}, supportedOTLPEndpointEnv...)
	all = append(all, supportedOTLPMetricEndpointEnv...)
	all = append(all, unsupportedOTLPEnv...)
	all = append(all, unsupportedSDKEnv...)
	var exporter sdkmetric.Exporter
	err := withSanitizedEnv(all, func() error {
		var err error
		exporter, err = fn()
		return err
	})
	return exporter, err
}

type metricInstruments struct {
	recordsRead        otelmetric.Int64Counter
	recordsTransformed otelmetric.Int64Counter
	recordsLoaded      otelmetric.Int64Counter
	recordsFailed      otelmetric.Int64Counter
	batchesFlushed     otelmetric.Int64Counter
}

// RunCounters accumulates ETL counters locally and flushes OTel instruments at batch boundaries.
type RunCounters struct {
	enabled bool
	inst    metricInstruments

	recordsRead        int64
	recordsTransformed int64
	recordsLoaded      int64
	recordsFailed      int64
	batchesFlushed     int64

	flushedRead        int64
	flushedTransformed int64
	flushedLoaded      int64
	flushedFailed      int64
	flushedBatches     int64
}

// NewRunCounters creates local run counters. Disabled telemetry returns counters whose hot path only increments fields.
func NewRunCounters(ctx context.Context) *RunCounters {
	meter, ok := Meter(ctx)
	if !ok {
		return &RunCounters{}
	}
	inst, err := newMetricInstruments(meter)
	if err != nil {
		return &RunCounters{}
	}
	return &RunCounters{enabled: true, inst: inst}
}

func newMetricInstruments(meter otelmetric.Meter) (metricInstruments, error) {
	var inst metricInstruments
	var err error
	if inst.recordsRead, err = meter.Int64Counter("pm.records.read", otelmetric.WithDescription("Records read from source connectors.")); err != nil {
		return metricInstruments{}, err
	}
	if inst.recordsTransformed, err = meter.Int64Counter("pm.records.transformed", otelmetric.WithDescription("Records transformed for destination writes.")); err != nil {
		return metricInstruments{}, err
	}
	if inst.recordsLoaded, err = meter.Int64Counter("pm.records.loaded", otelmetric.WithDescription("Records loaded into destinations.")); err != nil {
		return metricInstruments{}, err
	}
	if inst.recordsFailed, err = meter.Int64Counter("pm.records.failed", otelmetric.WithDescription("Records failed during ETL.")); err != nil {
		return metricInstruments{}, err
	}
	if inst.batchesFlushed, err = meter.Int64Counter("pm.batches.flushed", otelmetric.WithDescription("ETL batches flushed to destination sinks.")); err != nil {
		return metricInstruments{}, err
	}
	return inst, nil
}

// RecordRead increments the local records-read counter without emitting OTel metrics.
func (c *RunCounters) RecordRead() {
	if c != nil {
		c.recordsRead++
	}
}

// RecordTransformed increments the local transformed counter without emitting OTel metrics.
func (c *RunCounters) RecordTransformed() {
	if c != nil {
		c.recordsTransformed++
	}
}

// RecordLoaded increments the local loaded counter without emitting OTel metrics.
func (c *RunCounters) RecordLoaded(n int) {
	if c != nil && n > 0 {
		c.recordsLoaded += int64(n)
	}
}

// RecordFailed increments the local failed counter without emitting OTel metrics.
func (c *RunCounters) RecordFailed(n int) {
	if c != nil && n > 0 {
		c.recordsFailed += int64(n)
	}
}

// RecordBatch increments the local batch counter without emitting OTel metrics.
func (c *RunCounters) RecordBatch() {
	if c != nil {
		c.batchesFlushed++
	}
}

// Flush emits deltas since the previous flush. Call from batch boundaries, not per record.
func (c *RunCounters) Flush(ctx context.Context) {
	if c == nil || !c.enabled {
		return
	}
	if delta := c.recordsRead - c.flushedRead; delta != 0 {
		c.inst.recordsRead.Add(ctx, delta)
		c.flushedRead = c.recordsRead
	}
	if delta := c.recordsTransformed - c.flushedTransformed; delta != 0 {
		c.inst.recordsTransformed.Add(ctx, delta)
		c.flushedTransformed = c.recordsTransformed
	}
	if delta := c.recordsLoaded - c.flushedLoaded; delta != 0 {
		c.inst.recordsLoaded.Add(ctx, delta)
		c.flushedLoaded = c.recordsLoaded
	}
	if delta := c.recordsFailed - c.flushedFailed; delta != 0 {
		c.inst.recordsFailed.Add(ctx, delta)
		c.flushedFailed = c.recordsFailed
	}
	if delta := c.batchesFlushed - c.flushedBatches; delta != 0 {
		c.inst.batchesFlushed.Add(ctx, delta)
		c.flushedBatches = c.batchesFlushed
	}
}
