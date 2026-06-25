package churnkey_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics/internal/connectors"
	"polymetrics/internal/connectors/churnkey"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Churnkey
// connector: it asserts the two custom auth headers (x-ck-api-key from the
// secret, x-ck-app from config), Churnkey's limit/skip offset pagination over a
// top-level JSON array, and record mapping. The server hands back a full page
// (== limit) on the first request so the connector advances skip and fetches a
// second, short page that stops the loop.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAPIKey, sawApp string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAPIKey = r.Header.Get("x-ck-api-key")
		sawApp = r.Header.Get("x-ck-app")
		if r.URL.Path != "/sessions" {
			http.NotFound(w, r)
			return
		}
		// Force a page size of 2 so two records fill the first page and trigger
		// a second request at skip=2.
		switch r.URL.Query().Get("skip") {
		case "", "0":
			_, _ = w.Write([]byte(`[{"_id":"s_1","createdAt":"2026-01-01T00:00:00Z","canceled":true},{"_id":"s_2","createdAt":"2026-01-02T00:00:00Z","canceled":false}]`))
		case "2":
			_, _ = w.Write([]byte(`[{"_id":"s_3","createdAt":"2026-01-03T00:00:00Z","canceled":true}]`))
		default:
			t.Errorf("unexpected skip=%q", r.URL.Query().Get("skip"))
			_, _ = w.Write([]byte(`[]`))
		}
	}))
	defer srv.Close()

	c := churnkey.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "x-ck-app": "app_123", "page_size": "2"},
		Secrets: map[string]string{"api_key": "data_test_456"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "sessions", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAPIKey != "data_test_456" {
		t.Fatalf("x-ck-api-key = %q, want data_test_456", sawAPIKey)
	}
	if sawApp != "app_123" {
		t.Fatalf("x-ck-app = %q, want app_123", sawApp)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["_id"] == nil || rec["created_at"] == nil {
			t.Fatalf("record missing _id/created_at: %+v", rec)
		}
	}
	if got[0]["canceled"] != true {
		t.Fatalf("record canceled = %v, want true", got[0]["canceled"])
	}
}

// TestReadStopsAtMaxPages verifies that, when the server keeps returning full
// pages, the connector honours the configured page cap rather than looping
// forever.
func TestReadStopsAtMaxPages(t *testing.T) {
	var requests int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		// Always return a full page (size 1) so the loop never self-terminates.
		skip := r.URL.Query().Get("skip")
		_, _ = w.Write([]byte(`[{"_id":"s_` + skip + `","createdAt":"2026-01-01T00:00:00Z"}]`))
	}))
	defer srv.Close()

	c := churnkey.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "x-ck-app": "app_123", "page_size": "1", "max_pages": "3"},
		Secrets: map[string]string{"api_key": "data_test_456"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "sessions", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if requests != 3 {
		t.Fatalf("requests = %d, want 3 (max_pages cap)", requests)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3", len(got))
	}
}

// TestSessionAggregationStream exercises the second core stream, which is an
// unpaginated array of grouped counts.
func TestSessionAggregationStream(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/session-aggregation" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`[{"month":"2026-01","canceled":true,"count":42},{"month":"2026-02","canceled":false,"count":17}]`))
	}))
	defer srv.Close()

	c := churnkey.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "x-ck-app": "app_123"},
		Secrets: map[string]string{"api_key": "data_test_456"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "session_aggregation", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2", len(got))
	}
	if got[0]["count"] == nil {
		t.Fatalf("aggregation record missing count: %+v", got[0])
	}
}

// TestFixtureModeNoNetwork ensures fixture mode emits deterministic records
// without any network access, so conformance can run credential-free.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := churnkey.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"sessions", "session_aggregation"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("Read(%s) fixture: %v", stream, err)
		}
		if len(got) != 2 {
			t.Fatalf("fixture %s records = %d, want 2", stream, len(got))
		}
	}
	// Check should also short-circuit in fixture mode with no creds.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
}

// TestCheckRequiresCredentials asserts non-fixture Check fails fast without an
// api_key or app id rather than making a doomed network call.
func TestCheckRequiresCredentials(t *testing.T) {
	c := churnkey.New()
	if err := c.Check(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"x-ck-app": "app_123"}}); err == nil {
		t.Fatal("Check without api_key should fail")
	}
	if err := c.Check(context.Background(), connectors.RuntimeConfig{Secrets: map[string]string{"api_key": "data_x"}}); err == nil {
		t.Fatal("Check without x-ck-app should fail")
	}
}

// TestBaseURLValidation rejects non-http(s) base_url overrides (SSRF guard).
func TestBaseURLValidation(t *testing.T) {
	c := churnkey.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "file:///etc/passwd", "x-ck-app": "app_123"},
		Secrets: map[string]string{"api_key": "data_x"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "sessions", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read with file:// base_url should be rejected")
	}
}

// TestRegistryResolvesChurnkey confirms self-registration via init() and the
// read-only capability shape.
func TestRegistryResolvesChurnkey(t *testing.T) {
	_ = churnkey.New() // ensure init ran
	c := churnkey.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only connector)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("churnkey"); !ok {
		t.Fatal("registry did not resolve churnkey (self-registration)")
	}
}

// TestCatalogStreams asserts the published catalog exposes the core streams with
// the expected primary keys.
func TestCatalogStreams(t *testing.T) {
	c := churnkey.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "churnkey" {
		t.Fatalf("catalog connector = %q, want churnkey", cat.Connector)
	}
	byName := map[string]connectors.Stream{}
	for _, s := range cat.Streams {
		byName[s.Name] = s
	}
	sessions, ok := byName["sessions"]
	if !ok {
		t.Fatal("catalog missing sessions stream")
	}
	if len(sessions.PrimaryKey) != 1 || sessions.PrimaryKey[0] != "_id" {
		t.Fatalf("sessions primary key = %v, want [_id]", sessions.PrimaryKey)
	}
	if _, ok := byName["session_aggregation"]; !ok {
		t.Fatal("catalog missing session_aggregation stream")
	}
}
