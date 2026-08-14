// Package discovery turns tenant-defined provider objects into the same
// catalog stream and draft-07 schema contract used by static bundles.
package discovery

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"sort"
	"strings"
	"sync"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/synccontract"
)

const (
	defaultConcurrency = 10
	defaultAttempts    = 5
	defaultBackoff     = time.Second
	defaultMaxBackoff  = 30 * time.Second
	defaultCacheTTL    = 15 * time.Minute
)

// Provider contains the connector-specific network declaration. The driver
// owns scheduling, retrying, fallback assembly, schemas, progress and cache
// lifetime; a provider only lists tenant objects and describes one object.
type Provider interface {
	List(context.Context) ([]Object, error)
	Describe(context.Context, Object) ([]Field, error)
}

// Object is the safe, provider-described identity of one stream. PrimaryKey
// and CursorField are optional explicit provider contracts; when omitted the
// driver derives a key from described unique non-nullable fields.
type Object struct {
	Name        string
	Description string
	PrimaryKey  []string
	CursorField string
}

// Field preserves the provider facts needed for schema assembly. Raw is
// opaque to the driver and is only interpreted by the connector's Converter.
type Field struct {
	Name        string
	Description string
	Required    bool
	Nullable    bool
	Unique      bool
	Raw         any
}

// Converter translates exactly one provider-described field to a draft-07
// property schema. It cannot create fields: the driver always uses Field.Name.
type Converter func(Field) (json.RawMessage, error)

// Progress is emitted at bounded heartbeats, not once per object, so a large
// catalog remains observable without turning progress into log noise.
type Progress struct {
	Completed int
	Total     int
}

// Spec is a declarative driver configuration. Fallback must be a provider
// documented object list: it is used only when the provider-wide List request
// fails, and the resulting catalog is always marked incomplete.
type Spec struct {
	Connector          string
	Fallback           []Object
	FallbackPrimaryKey []string
	FallbackCursor     string
	Converter          Converter
	Concurrency        int
	MaxAttempts        int
	BaseBackoff        time.Duration
	MaxBackoff         time.Duration
	CacheTTL           time.Duration
	Now                func() time.Time
	Sleep              func(context.Context, time.Duration) error
	// Jitter receives the exponential backoff cap and returns an additional,
	// bounded delay. A nil function uses no additional delay.
	Jitter   func(time.Duration) time.Duration
	Progress func(Progress)
}

// Driver is safely reusable by a connector instance. Its cache key comprises
// only connector name and CoordinationIdentity.AuthCohortKey(), never config
// values, secret values, or a credential-derived value.
type Driver struct {
	spec  Spec
	mu    sync.Mutex
	cache map[cacheKey]cachedCatalog
}

type cacheKey struct {
	connector string
	cohort    connectors.AuthCohortKey
}

type cachedCatalog struct {
	catalog connectors.Catalog
}

// New validates a reusable discovery declaration.
func New(spec Spec) (*Driver, error) {
	if strings.TrimSpace(spec.Connector) == "" {
		return nil, errors.New("discovery connector is required")
	}
	if len(spec.Fallback) == 0 {
		return nil, errors.New("discovery fallback objects are required")
	}
	if spec.Converter == nil {
		return nil, errors.New("discovery field converter is required")
	}
	if spec.Concurrency <= 0 || spec.Concurrency > defaultConcurrency {
		spec.Concurrency = defaultConcurrency
	}
	if spec.MaxAttempts <= 0 {
		spec.MaxAttempts = defaultAttempts
	}
	if spec.BaseBackoff <= 0 {
		spec.BaseBackoff = defaultBackoff
	}
	if spec.MaxBackoff <= 0 {
		spec.MaxBackoff = defaultMaxBackoff
	}
	if spec.CacheTTL <= 0 {
		spec.CacheTTL = defaultCacheTTL
	}
	if spec.Now == nil {
		spec.Now = func() time.Time { return time.Now().UTC() }
	}
	if spec.Sleep == nil {
		spec.Sleep = sleepContext
	}
	if spec.Jitter == nil {
		spec.Jitter = func(cap time.Duration) time.Duration {
			if cap <= 0 {
				return 0
			}
			return time.Duration(rand.Int64N(int64(cap) + 1))
		}
	}
	return &Driver{spec: spec, cache: make(map[cacheKey]cachedCatalog)}, nil
}

// Catalog discovers a current tenant surface. A valid cache is reused only for
// the same opaque coordination identity. Explicit refresh bypasses it; expired
// data is never returned by this in-process cache. The persisted app snapshot
// separately exposes its expiry through DiscoveryStatus.Stale.
func (d *Driver) Catalog(ctx context.Context, cfg connectors.RuntimeConfig, provider Provider) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	if provider == nil {
		return connectors.Catalog{}, errors.New("discovery provider is required")
	}
	now := d.spec.Now().UTC()
	key, cacheable := d.cacheKey(cfg)
	if cached, ok := d.cached(key, cacheable); ok && !cfg.ForceCatalogRefresh {
		if cached.Discovery != nil && now.Before(cached.Discovery.ExpiresAt) {
			cached.Discovery.Cached = true
			return cached, nil
		}
	}

	catalog, err := d.discover(ctx, provider, now)
	if err != nil {
		return connectors.Catalog{}, err
	}
	if catalog.Discovery.Complete || len(catalog.Streams) > 0 {
		d.store(key, cacheable, catalog)
	}
	return catalog, nil
}

func (d *Driver) discover(ctx context.Context, provider Provider, now time.Time) (connectors.Catalog, error) {
	objects, listAttempts, err := d.list(ctx, provider)
	usedFallback := err != nil
	failures := make([]connectors.DiscoveryFailure, 0)
	if err != nil {
		failures = append(failures, connectors.DiscoveryFailure{Stage: "list", Attempts: listAttempts})
		objects = append([]Object(nil), d.spec.Fallback...)
	}
	objects = normalizeObjects(objects)

	streams, describeFailures, err := d.describeAll(ctx, provider, objects)
	if err != nil {
		return connectors.Catalog{}, err
	}
	failures = append(failures, describeFailures...)
	status := &connectors.DiscoveryStatus{
		Complete:     !usedFallback && len(failures) == 0,
		UsedFallback: usedFallback,
		RefreshedAt:  now,
		ExpiresAt:    now.Add(d.spec.CacheTTL),
		Failures:     failures,
	}
	return connectors.Catalog{Connector: d.spec.Connector, Streams: streams, Discovery: status}, nil
}

func (d *Driver) list(ctx context.Context, provider Provider) ([]Object, int, error) {
	var result []Object
	attempts, err := d.retry(ctx, func() error {
		var listErr error
		result, listErr = provider.List(ctx)
		return listErr
	})
	return result, attempts, err
}

func (d *Driver) describeAll(ctx context.Context, provider Provider, objects []Object) ([]connectors.Stream, []connectors.DiscoveryFailure, error) {
	type result struct {
		index    int
		stream   connectors.Stream
		failure  *connectors.DiscoveryFailure
		terminal error
	}
	results := make(chan result, len(objects))
	jobs := make(chan int)
	workers := d.spec.Concurrency
	if workers > len(objects) {
		workers = len(objects)
	}
	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for index := range jobs {
				object := objects[index]
				fields, attempts, err := d.describe(ctx, provider, object)
				if err != nil {
					if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
						results <- result{terminal: err}
						continue
					}
					results <- result{index: index, failure: &connectors.DiscoveryFailure{Object: object.Name, Stage: "describe", Attempts: attempts}}
					continue
				}
				stream, err := d.stream(object, fields)
				if err != nil {
					results <- result{index: index, failure: &connectors.DiscoveryFailure{Object: object.Name, Stage: "schema", Attempts: attempts}}
					continue
				}
				results <- result{index: index, stream: stream}
			}
		}()
	}
	go func() {
		for index := range objects {
			select {
			case jobs <- index:
			case <-ctx.Done():
				close(jobs)
				wg.Wait()
				close(results)
				return
			}
		}
		close(jobs)
		wg.Wait()
		close(results)
	}()

	ordered := make([]connectors.Stream, len(objects))
	present := make([]bool, len(objects))
	failures := make([]connectors.DiscoveryFailure, 0)
	completed := 0
	for result := range results {
		if result.terminal != nil {
			return nil, nil, result.terminal
		}
		completed++
		if result.failure != nil {
			failures = append(failures, *result.failure)
		} else {
			ordered[result.index] = result.stream
			present[result.index] = true
		}
		if d.spec.Progress != nil && (completed%100 == 0 || completed == len(objects)) {
			d.spec.Progress(Progress{Completed: completed, Total: len(objects)})
		}
	}
	streams := make([]connectors.Stream, 0, len(objects)-len(failures))
	for index, stream := range ordered {
		if present[index] {
			streams = append(streams, stream)
		}
	}
	return streams, failures, nil
}

func (d *Driver) describe(ctx context.Context, provider Provider, object Object) ([]Field, int, error) {
	var fields []Field
	attempts, err := d.retry(ctx, func() error {
		var describeErr error
		fields, describeErr = provider.Describe(ctx, object)
		return describeErr
	})
	return fields, attempts, err
}

func (d *Driver) retry(ctx context.Context, call func() error) (int, error) {
	for attempt := 1; attempt <= d.spec.MaxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return attempt - 1, err
		}
		err := call()
		if err == nil {
			return attempt, nil
		}
		if !isRateLimited(err) || attempt == d.spec.MaxAttempts {
			return attempt, err
		}
		delay := d.backoff(attempt)
		if err := d.spec.Sleep(ctx, delay); err != nil {
			return attempt, err
		}
	}
	return d.spec.MaxAttempts, errors.New("discovery retry exhausted")
}

func (d *Driver) backoff(attempt int) time.Duration {
	delay := d.spec.BaseBackoff
	for index := 1; index < attempt && delay < d.spec.MaxBackoff; index++ {
		delay *= 2
		if delay > d.spec.MaxBackoff {
			delay = d.spec.MaxBackoff
		}
	}
	if d.spec.Jitter != nil {
		jitter := d.spec.Jitter(delay)
		if jitter > 0 {
			delay += jitter
			if delay > d.spec.MaxBackoff {
				delay = d.spec.MaxBackoff
			}
		}
	}
	return delay
}

func (d *Driver) stream(object Object, fields []Field) (connectors.Stream, error) {
	properties := make(map[string]json.RawMessage, len(fields))
	required := make([]string, 0)
	known := make(map[string]Field, len(fields))
	for _, field := range fields {
		if strings.TrimSpace(field.Name) == "" {
			return connectors.Stream{}, errors.New("provider field name is required")
		}
		if _, exists := known[field.Name]; exists {
			return connectors.Stream{}, fmt.Errorf("provider field %q is duplicated", field.Name)
		}
		property, err := d.spec.Converter(field)
		if err != nil {
			return connectors.Stream{}, errors.New("provider field conversion failed")
		}
		if !validPropertySchema(property) {
			return connectors.Stream{}, errors.New("provider field conversion produced an invalid schema")
		}
		properties[field.Name] = append(json.RawMessage(nil), property...)
		known[field.Name] = field
		if field.Required {
			required = append(required, field.Name)
		}
	}
	if len(properties) == 0 {
		return connectors.Stream{}, errors.New("provider object has no described fields")
	}
	sort.Strings(required)
	doc := map[string]any{
		"$schema":              "http://json-schema.org/draft-07/schema#",
		"type":                 "object",
		"additionalProperties": true,
		"properties":           properties,
		"title":                object.Name,
	}
	if object.Description != "" {
		doc["description"] = object.Description
	}
	if len(required) > 0 {
		doc["required"] = required
	}
	if primaryKey := d.primaryKey(object, known); len(primaryKey) > 0 {
		doc["x-primary-key"] = primaryKey
		doc["x-source_defined_primary_key"] = primaryKey
	}
	cursor := d.cursorField(object, known)
	if cursor != "" {
		doc["x-cursor-field"] = cursor
		doc["x-source_defined_cursor"] = true
		doc["x-default_cursor_field"] = cursor
	}
	doc["x-stream_name"] = object.Name
	doc["x-supported_sync_modes"] = supportedSyncModes(len(d.primaryKey(object, known)) > 0, cursor != "")
	doc["x-default_sync_mode"] = defaultSyncMode(len(d.primaryKey(object, known)) > 0, cursor != "")
	raw, err := json.Marshal(doc)
	if err != nil {
		return connectors.Stream{}, errors.New("assemble discovered schema")
	}
	return connectors.StreamFromSchema(object.Name, object.Description, raw)
}

func supportedSyncModes(hasPrimaryKey, hasCursor bool) []string {
	return synccontract.SupportedPublicModeNames(discoveredSyncModeCapabilities(hasPrimaryKey, hasCursor))
}

func defaultSyncMode(hasPrimaryKey, hasCursor bool) string {
	return synccontract.DefaultPublicModeName(discoveredSyncModeCapabilities(hasPrimaryKey, hasCursor))
}

func discoveredSyncModeCapabilities(hasPrimaryKey, hasCursor bool) synccontract.PublicModeCapabilities {
	return synccontract.PublicModeCapabilities{
		HasPrimaryKey:          hasPrimaryKey,
		HasCursor:              hasCursor,
		HasIncrementalExecutor: hasCursor,
	}
}

func (d *Driver) primaryKey(object Object, known map[string]Field) []string {
	if fieldsPresent(object.PrimaryKey, known) {
		return append([]string(nil), object.PrimaryKey...)
	}
	candidates := make([]string, 0)
	for name, field := range known {
		if field.Unique && !field.Nullable {
			candidates = append(candidates, name)
		}
	}
	sort.Strings(candidates)
	if len(candidates) > 0 {
		return []string{candidates[0]}
	}
	if fieldsPresent(d.spec.FallbackPrimaryKey, known) {
		return append([]string(nil), d.spec.FallbackPrimaryKey...)
	}
	return nil
}

func (d *Driver) cursorField(object Object, known map[string]Field) string {
	if object.CursorField != "" {
		if _, ok := known[object.CursorField]; ok {
			return object.CursorField
		}
	}
	if d.spec.FallbackCursor != "" {
		if _, ok := known[d.spec.FallbackCursor]; ok {
			return d.spec.FallbackCursor
		}
	}
	return ""
}

func (d *Driver) cacheKey(cfg connectors.RuntimeConfig) (cacheKey, bool) {
	cohort := cfg.CoordinationIdentity.AuthCohortKey()
	if cohort == "" {
		return cacheKey{}, false
	}
	return cacheKey{connector: d.spec.Connector, cohort: cohort}, true
}

func (d *Driver) cached(key cacheKey, ok bool) (connectors.Catalog, bool) {
	if !ok {
		return connectors.Catalog{}, false
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	entry, exists := d.cache[key]
	if !exists {
		return connectors.Catalog{}, false
	}
	return cloneCatalog(entry.catalog), true
}

func (d *Driver) store(key cacheKey, ok bool, catalog connectors.Catalog) {
	if !ok {
		return
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	d.cache[key] = cachedCatalog{catalog: cloneCatalog(catalog)}
}

func normalizeObjects(objects []Object) []Object {
	seen := make(map[string]struct{}, len(objects))
	normalized := make([]Object, 0, len(objects))
	for _, object := range objects {
		object.Name = strings.TrimSpace(object.Name)
		if object.Name == "" {
			continue
		}
		if _, exists := seen[object.Name]; exists {
			continue
		}
		seen[object.Name] = struct{}{}
		normalized = append(normalized, object)
	}
	sort.Slice(normalized, func(i, j int) bool { return normalized[i].Name < normalized[j].Name })
	return normalized
}

func validPropertySchema(raw json.RawMessage) bool {
	var property map[string]json.RawMessage
	return len(raw) > 0 && json.Unmarshal(raw, &property) == nil && property != nil
}

func fieldsPresent(names []string, known map[string]Field) bool {
	if len(names) == 0 {
		return false
	}
	for _, name := range names {
		if _, ok := known[name]; !ok {
			return false
		}
	}
	return true
}

func isRateLimited(err error) bool {
	var httpError *connsdk.HTTPError
	if errors.As(err, &httpError) && httpError.Status == 429 {
		return true
	}
	var status interface{ StatusCode() int }
	return errors.As(err, &status) && status.StatusCode() == 429
}

func sleepContext(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func cloneCatalog(catalog connectors.Catalog) connectors.Catalog {
	copyCatalog := catalog
	copyCatalog.Streams = make([]connectors.Stream, len(catalog.Streams))
	for index, stream := range catalog.Streams {
		copyCatalog.Streams[index] = stream
		copyCatalog.Streams[index].Fields = append([]connectors.Field(nil), stream.Fields...)
		copyCatalog.Streams[index].PrimaryKey = append([]string(nil), stream.PrimaryKey...)
		copyCatalog.Streams[index].CursorFields = append([]string(nil), stream.CursorFields...)
		copyCatalog.Streams[index].Schema = append(json.RawMessage(nil), stream.Schema...)
	}
	if catalog.Discovery != nil {
		status := *catalog.Discovery
		status.Failures = append([]connectors.DiscoveryFailure(nil), catalog.Discovery.Failures...)
		copyCatalog.Discovery = &status
	}
	return copyCatalog
}
