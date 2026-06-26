package incidentio_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	incidentio "polymetrics.ai/internal/connectors/incident-io"
)

// TestReadPaginatesAndAuthenticates is the red-first test: Bearer auth, the
// incident.io pagination_meta.after cursor pagination over the incidents[]
// array across two pages, and record mapping.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/v2/incidents" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("after") {
		case "":
			_, _ = w.Write([]byte(`{"incidents":[{"id":"01ABC","reference":"INC-1","name":"first","created_at":"2026-01-01T00:00:00Z"},{"id":"01ABD","reference":"INC-2","name":"second","created_at":"2026-01-02T00:00:00Z"}],"pagination_meta":{"after":"01ABD","page_size":2,"total_record_count":3}}`))
		case "01ABD":
			_, _ = w.Write([]byte(`{"incidents":[{"id":"01ABE","reference":"INC-3","name":"third","created_at":"2026-01-03T00:00:00Z"}],"pagination_meta":{"page_size":2,"total_record_count":3}}`))
		default:
			t.Errorf("unexpected after=%q", r.URL.Query().Get("after"))
			_, _ = w.Write([]byte(`{"incidents":[],"pagination_meta":{}}`))
		}
	}))
	defer srv.Close()

	c := incidentio.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "inc_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "incidents", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer inc_test_123" {
		t.Fatalf("Authorization = %q, want Bearer inc_test_123", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["reference"] == nil {
			t.Fatalf("record missing id/reference: %+v", rec)
		}
	}
}

// TestReadSeveritiesSinglePage exercises a v1 single-page stream that has no
// pagination_meta.after; the loop must stop after one page.
func TestReadSeveritiesSinglePage(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.URL.Path != "/v1/severities" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"severities":[{"id":"sev_1","name":"Major","rank":1},{"id":"sev_2","name":"Minor","rank":2}]}`))
	}))
	defer srv.Close()

	c := incidentio.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "inc_test_123"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "severities", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if calls != 1 {
		t.Fatalf("calls = %d, want 1 (no pagination_meta.after)", calls)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2", len(got))
	}
}

// TestFixtureMode confirms the no-network fixture path emits deterministic
// records so conformance runs without credentials.
func TestFixtureMode(t *testing.T) {
	c := incidentio.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "incidents", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read(fixture): %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("fixture records = %d, want 2", len(got))
	}
	if got[0]["id"] == nil {
		t.Fatalf("fixture record missing id: %+v", got[0])
	}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

func TestCatalogAndRegistry(t *testing.T) {
	c := incidentio.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write false (read-only connector)", caps)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("streams = %d, want >= 3", len(cat.Streams))
	}

	r := connectors.NewRegistry()
	if _, ok := r.Get("incident-io"); !ok {
		t.Fatal("registry did not resolve incident-io (self-registration)")
	}
}
