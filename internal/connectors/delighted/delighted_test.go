package delighted_test

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics/internal/connectors"
	"polymetrics/internal/connectors/delighted"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Delighted
// connector: HTTP Basic auth (API key as username, blank password), page-number
// pagination over a top-level JSON array, and record mapping. Delighted returns
// a short final page to signal the end.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/survey_responses.json" {
			http.NotFound(w, r)
			return
		}
		if r.URL.Query().Get("per_page") != "2" {
			t.Errorf("per_page = %q, want 2", r.URL.Query().Get("per_page"))
		}
		switch r.URL.Query().Get("page") {
		case "1":
			_, _ = w.Write([]byte(`[{"id":"1","score":10,"updated_at":1700000000},{"id":"2","score":9,"updated_at":1700000100}]`))
		case "2":
			// Short page (fewer than per_page) signals the last page.
			_, _ = w.Write([]byte(`[{"id":"3","score":8,"updated_at":1700000200}]`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
			_, _ = w.Write([]byte(`[]`))
		}
	}))
	defer srv.Close()

	c := delighted.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"api_key": "test_api_key_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "survey_responses", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("test_api_key_123:"))
	if sawAuth != wantAuth {
		t.Fatalf("Authorization = %q, want %q", sawAuth, wantAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("record missing id: %+v", rec)
		}
	}
}

func TestReadAppliesSinceFilter(t *testing.T) {
	var sawSince string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawSince = r.URL.Query().Get("since")
		_, _ = w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	c := delighted.New()
	cfg := connectors.RuntimeConfig{
		// 2022-05-30T00:00:00Z => 1653868800
		Config:  map[string]string{"base_url": srv.URL, "since": "2022-05-30T00:00:00Z"},
		Secrets: map[string]string{"api_key": "k"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "survey_responses", Config: cfg}, func(connectors.Record) error { return nil })
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawSince != "1653868800" {
		t.Fatalf("since = %q, want unix 1653868800", sawSince)
	}
}

func TestFixtureModeNeedsNoNetwork(t *testing.T) {
	c := delighted.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
	for _, stream := range []string{"survey_responses", "people", "bounces", "unsubscribes", "metrics"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("Read(fixture %s) = %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture stream %s produced no records", stream)
		}
	}
}

func TestMetricsStreamReadsSingleObject(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/metrics.json" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"nps":42,"promoter_count":10,"detractor_count":2,"response_count":20}`))
	}))
	defer srv.Close()

	c := delighted.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "k"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "metrics", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read(metrics): %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("metrics records = %d, want 1", len(got))
	}
	if got[0]["nps"] == nil {
		t.Fatalf("metrics record missing nps: %+v", got[0])
	}
}

func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := delighted.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example.com"},
		Secrets: map[string]string{"api_key": "k"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "people", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "base_url") {
		t.Fatalf("Read with bad base_url scheme = %v, want base_url error", err)
	}
}

func TestRegisteredReadOnly(t *testing.T) {
	_ = delighted.New() // ensure init ran
	c := delighted.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("delighted"); !ok {
		t.Fatal("registry did not resolve delighted (self-registration)")
	}
}

func TestCatalogListsCoreStreams(t *testing.T) {
	c := delighted.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"survey_responses": false, "people": false, "bounces": false, "unsubscribes": false, "metrics": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
	}
	for name, found := range want {
		if !found {
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}
