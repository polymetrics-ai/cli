package discovery

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strconv"
	"sync"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors"
)

func TestDriverRetriesRateLimitedDescriptionsAndBuildsSchema(t *testing.T) {
	provider := &fakeProvider{
		objects: []Object{{Name: "contacts", Description: "Contact records", CursorField: "created_at"}},
		fields: map[string][]Field{
			"contacts": {
				{Name: "id", Unique: true, Nullable: false, Raw: "integer"},
				{Name: "created_at", Raw: "timestamp"},
			},
		},
		failures: map[string][]error{"contacts": {rateLimitedError{}}},
	}
	var sleeps []time.Duration
	driver, err := New(Spec{
		Connector:   "acme",
		Fallback:    []Object{{Name: "contacts", Description: "Contact records"}},
		Converter:   typeConverter,
		BaseBackoff: time.Second,
		MaxAttempts: 3,
		Jitter:      func(time.Duration) time.Duration { return 0 },
		Sleep: func(_ context.Context, d time.Duration) error {
			sleeps = append(sleeps, d)
			return nil
		},
		Now: func() time.Time { return time.Date(2026, 8, 6, 0, 0, 0, 0, time.UTC) },
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	catalog, err := driver.Catalog(context.Background(), connectors.RuntimeConfig{}, provider)
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if provider.describeCalls("contacts") != 2 {
		t.Fatalf("describe calls = %d, want 2", provider.describeCalls("contacts"))
	}
	if len(sleeps) != 1 || sleeps[0] != time.Second {
		t.Fatalf("sleeps = %v, want [1s]", sleeps)
	}
	if catalog.Discovery == nil || !catalog.Discovery.Complete {
		t.Fatalf("discovery = %#v, want complete", catalog.Discovery)
	}
	if len(catalog.Streams) != 1 {
		t.Fatalf("streams = %#v, want one", catalog.Streams)
	}
	stream := catalog.Streams[0]
	if got, want := stream.PrimaryKey, []string{"id"}; !sameStrings(got, want) {
		t.Fatalf("primary key = %v, want %v", got, want)
	}
	if got, want := stream.CursorFields, []string{"created_at"}; !sameStrings(got, want) {
		t.Fatalf("cursor fields = %v, want %v", got, want)
	}
	var schema map[string]any
	if err := json.Unmarshal(stream.Schema, &schema); err != nil {
		t.Fatalf("schema invalid: %v", err)
	}
	if schema["x-primary-key"] == nil || schema["x-cursor-field"] != "created_at" || schema["x-stream_name"] != "contacts" || schema["x-default_sync_mode"] != "incremental_append" {
		t.Fatalf("schema sync contract = %#v", schema)
	}
}

func TestDefaultSyncModeNeverSelectsTypedOnlyCompatibilityName(t *testing.T) {
	tests := []struct {
		name          string
		hasPrimaryKey bool
		hasCursor     bool
		want          string
	}{
		{name: "neither", want: "full_refresh_append"},
		{name: "primary key only", hasPrimaryKey: true, want: "full_refresh_overwrite"},
		{name: "cursor only", hasCursor: true, want: "incremental_append"},
		{name: "both", hasPrimaryKey: true, hasCursor: true, want: "incremental_append"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := defaultSyncMode(tt.hasPrimaryKey, tt.hasCursor); got != tt.want {
				t.Fatalf("defaultSyncMode(%t, %t) = %q, want %q", tt.hasPrimaryKey, tt.hasCursor, got, tt.want)
			}
		})
	}
}

func TestDriverUsesDeclaredFallbackAndReportsPartialDiscovery(t *testing.T) {
	provider := &fakeProvider{
		listErr: errors.New("provider error body must not escape"),
		fields: map[string][]Field{
			"contacts":  {{Name: "id", Unique: true, Raw: "string"}},
			"companies": {{Name: "id", Unique: true, Raw: "string"}},
		},
		failures: map[string][]error{"companies": {errors.New("provider error body must not escape")}},
	}
	driver := mustDriver(t, Spec{
		Connector:   "acme",
		Fallback:    []Object{{Name: "contacts"}, {Name: "companies"}},
		Converter:   typeConverter,
		MaxAttempts: 1,
	})

	catalog, err := driver.Catalog(context.Background(), connectors.RuntimeConfig{}, provider)
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(catalog.Streams) != 1 || catalog.Streams[0].Name != "contacts" {
		t.Fatalf("streams = %#v, want only contacts", catalog.Streams)
	}
	status := catalog.Discovery
	if status == nil || status.Complete || !status.UsedFallback || len(status.Failures) != 2 {
		t.Fatalf("status = %#v, want declared incomplete fallback", status)
	}
	for _, failure := range status.Failures {
		if failure.Stage == "" {
			t.Fatalf("failure is missing its safe stage marker: %#v", failure)
		}
	}
	raw, err := json.Marshal(catalog)
	if err != nil {
		t.Fatalf("marshal safe catalog status: %v", err)
	}
	if bytes.Contains(raw, []byte("provider error body must not escape")) {
		t.Fatalf("catalog retained an upstream error body: %s", raw)
	}
}

func TestDriverCachesOnlyByCoordinationIdentityAndForceRefreshes(t *testing.T) {
	identity, err := connectors.NewCoordinationIdentity([]byte("dynamic-discovery-test-salt"), connectors.CredentialBinding{
		BindingID: "binding-test", ProviderFamily: "acme", AuthProfile: "api_token",
	})
	if err != nil {
		t.Fatalf("NewCoordinationIdentity: %v", err)
	}
	provider := &fakeProvider{
		objects: []Object{{Name: "contacts"}},
		fields:  map[string][]Field{"contacts": {{Name: "id", Unique: true, Raw: "string"}}},
	}
	now := time.Date(2026, 8, 6, 0, 0, 0, 0, time.UTC)
	driver := mustDriver(t, Spec{Connector: "acme", Fallback: []Object{{Name: "contacts"}}, Converter: typeConverter, CacheTTL: time.Hour, Now: func() time.Time { return now }})
	cfg := connectors.RuntimeConfig{CoordinationIdentity: identity}

	first, err := driver.Catalog(context.Background(), cfg, provider)
	if err != nil {
		t.Fatalf("first Catalog: %v", err)
	}
	second, err := driver.Catalog(context.Background(), cfg, provider)
	if err != nil {
		t.Fatalf("second Catalog: %v", err)
	}
	if provider.listCalls() != 1 || !second.Discovery.Cached {
		t.Fatalf("list calls = %d, cached = %#v", provider.listCalls(), second.Discovery)
	}
	if first.Discovery.ExpiresAt.IsZero() {
		t.Fatal("fresh discovery has no expiration for stale-cache detection")
	}
	_, err = driver.Catalog(context.Background(), connectors.RuntimeConfig{CoordinationIdentity: identity, ForceCatalogRefresh: true}, provider)
	if err != nil {
		t.Fatalf("forced Catalog: %v", err)
	}
	if provider.listCalls() != 2 {
		t.Fatalf("forced refresh list calls = %d, want 2", provider.listCalls())
	}
	_, err = driver.Catalog(context.Background(), connectors.RuntimeConfig{CoordinationIdentity: identity, Config: map[string]string{"endpoint_label": "second-connection"}}, provider)
	if err != nil {
		t.Fatalf("same-account Catalog: %v", err)
	}
	if provider.listCalls() != 2 {
		t.Fatalf("same-account cache was not reused: list calls = %d, want 2", provider.listCalls())
	}
}

func TestDriverBoundsConcurrencyToTen(t *testing.T) {
	objects := make([]Object, 11)
	for index := range objects {
		objects[index] = Object{Name: "object-" + string(rune('a'+index))}
	}
	provider := &blockingProvider{objects: objects, started: make(chan struct{}, len(objects)), release: make(chan struct{})}
	driver := mustDriver(t, Spec{Connector: "acme", Fallback: objects, Converter: typeConverter, Concurrency: 100})
	done := make(chan error, 1)
	go func() {
		_, err := driver.Catalog(context.Background(), connectors.RuntimeConfig{}, provider)
		done <- err
	}()
	for range 10 {
		select {
		case <-provider.started:
		case <-time.After(time.Second):
			t.Fatal("did not start ten bounded describe workers")
		}
	}
	select {
	case <-provider.started:
		t.Fatal("started an eleventh describe before a worker completed")
	case <-time.After(25 * time.Millisecond):
	}
	close(provider.release)
	if err := <-done; err != nil {
		t.Fatalf("Catalog: %v", err)
	}
}

func TestDriverCancellationLeavesNoBlockedWorkers(t *testing.T) {
	provider := &blockingProvider{
		objects:        []Object{{Name: "contacts"}},
		started:        make(chan struct{}, 1),
		release:        nil,
		waitForContext: true,
	}
	driver := mustDriver(t, Spec{Connector: "acme", Fallback: provider.objects, Converter: typeConverter})
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		_, err := driver.Catalog(ctx, connectors.RuntimeConfig{}, provider)
		done <- err
	}()
	select {
	case <-provider.started:
	case <-time.After(time.Second):
		t.Fatal("describe did not start")
	}
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Catalog error = %v, want context cancellation", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Catalog remained blocked after cancellation")
	}
}

func TestDriverReportsProgressEveryHundredObjects(t *testing.T) {
	objects := make([]Object, 101)
	fields := make(map[string][]Field, len(objects))
	for index := range objects {
		name := "object-" + strconv.Itoa(index)
		objects[index] = Object{Name: name}
		fields[name] = []Field{{Name: "id", Unique: true, Raw: "string"}}
	}
	var progress []Progress
	driver := mustDriver(t, Spec{
		Connector: "acme", Fallback: objects, Converter: typeConverter,
		Progress: func(heartbeat Progress) { progress = append(progress, heartbeat) },
	})
	_, err := driver.Catalog(context.Background(), connectors.RuntimeConfig{}, &fakeProvider{objects: objects, fields: fields})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if got, want := progress, []Progress{{Completed: 100, Total: 101}, {Completed: 101, Total: 101}}; !reflect.DeepEqual(got, want) {
		t.Fatalf("progress = %#v, want %#v", got, want)
	}
}

type fakeProvider struct {
	mu       sync.Mutex
	objects  []Object
	listErr  error
	fields   map[string][]Field
	failures map[string][]error
	listN    int
	describe map[string]int
}

func (p *fakeProvider) List(context.Context) ([]Object, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.listN++
	if p.listErr != nil {
		return nil, p.listErr
	}
	return append([]Object(nil), p.objects...), nil
}

func (p *fakeProvider) Describe(_ context.Context, object Object) ([]Field, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.describe == nil {
		p.describe = make(map[string]int)
	}
	p.describe[object.Name]++
	if pending := p.failures[object.Name]; len(pending) > 0 {
		p.failures[object.Name] = pending[1:]
		return nil, pending[0]
	}
	return append([]Field(nil), p.fields[object.Name]...), nil
}

func (p *fakeProvider) listCalls() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.listN
}

func (p *fakeProvider) describeCalls(name string) int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.describe[name]
}

type rateLimitedError struct{}

func (rateLimitedError) Error() string   { return "rate limited" }
func (rateLimitedError) StatusCode() int { return 429 }

func typeConverter(field Field) (json.RawMessage, error) {
	typ, _ := field.Raw.(string)
	if typ == "timestamp" {
		return json.RawMessage(`{"type":"string","format":"date-time"}`), nil
	}
	return json.RawMessage(`{"type":"` + typ + `"}`), nil
}

type blockingProvider struct {
	objects        []Object
	started        chan struct{}
	release        chan struct{}
	waitForContext bool
}

func (p *blockingProvider) List(context.Context) ([]Object, error) {
	return append([]Object(nil), p.objects...), nil
}

func (p *blockingProvider) Describe(ctx context.Context, object Object) ([]Field, error) {
	p.started <- struct{}{}
	if p.waitForContext {
		<-ctx.Done()
		return nil, ctx.Err()
	}
	select {
	case <-p.release:
		return []Field{{Name: "id", Unique: true, Raw: "string"}}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func mustDriver(t *testing.T, spec Spec) *Driver {
	t.Helper()
	driver, err := New(spec)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return driver
}

func sameStrings(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
