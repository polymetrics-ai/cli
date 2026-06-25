package circa_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics/internal/connectors"
	"polymetrics/internal/connectors/circa"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Circa
// connector: Bearer auth, Circa's page-increment pagination over data[] (page
// starts at 1, advances while a full page of page_size is returned), and record
// mapping. Red until internal/connectors/circa is implemented.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var pagesSeen []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/events" {
			http.NotFound(w, r)
			return
		}
		page := r.URL.Query().Get("page")
		pagesSeen = append(pagesSeen, page)
		switch page {
		case "", "1":
			// A full page of page_size=2 means there is another page.
			_, _ = w.Write([]byte(`{"data":[{"id":"ev_1","name":"Launch","updated_at":"2026-01-01T00:00:00Z"},{"id":"ev_2","name":"Webinar","updated_at":"2026-01-02T00:00:00Z"}]}`))
		case "2":
			// A short page (1 < page_size) ends pagination.
			_, _ = w.Write([]byte(`{"data":[{"id":"ev_3","name":"Summit","updated_at":"2026-01-03T00:00:00Z"}]}`))
		default:
			t.Errorf("unexpected page=%q", page)
			_, _ = w.Write([]byte(`{"data":[]}`))
		}
	}))
	defer srv.Close()

	c := circa.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"api_key": "circa_secret_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "events", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer circa_secret_123" {
		t.Fatalf("Authorization = %q, want Bearer circa_secret_123", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages); pages seen=%v", len(got), pagesSeen)
	}
	if len(pagesSeen) != 2 {
		t.Fatalf("requests = %d, want 2 pages; pages seen=%v", len(pagesSeen), pagesSeen)
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["name"] == nil || rec["updated_at"] == nil {
			t.Fatalf("record missing id/name/updated_at: %+v", rec)
		}
	}
}

// TestReadIncrementalLowerBound asserts the incremental cursor / start_date is
// forwarded as updated_at[min].
func TestReadIncrementalLowerBound(t *testing.T) {
	var sawMin string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawMin = r.URL.Query().Get("updated_at[min]")
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	defer srv.Close()

	c := circa.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "k"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{
		Stream: "events",
		Config: cfg,
		State:  map[string]string{"cursor": "2026-03-01T00:00:00Z"},
	}, func(connectors.Record) error { return nil })
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawMin != "2026-03-01T00:00:00Z" {
		t.Fatalf("updated_at[min] = %q, want the cursor value", sawMin)
	}
}

// TestFixtureModeNoNetwork ensures fixture mode emits deterministic records with
// no live credentials and no network call, so conformance passes credential-free.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := circa.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"events", "contacts", "companies", "teams"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture Read(%s) emitted no records", stream)
		}
		for _, rec := range got {
			if rec["id"] == nil {
				t.Fatalf("fixture %s record missing id: %+v", stream, rec)
			}
		}
	}
	// Check short-circuits in fixture mode without creds.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

func TestCatalogAndMetadata(t *testing.T) {
	c := circa.New()
	caps := c.Metadata().Capabilities
	if !caps.Check || !caps.Catalog || !caps.Read {
		t.Fatalf("capabilities = %+v, want Check && Catalog && Read", caps)
	}
	if caps.Write {
		t.Fatalf("circa is read-only; Write should be false")
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"events": true, "contacts": true, "companies": true, "teams": true}
	seen := map[string]bool{}
	for _, s := range cat.Streams {
		seen[s.Name] = true
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q missing primary key", s.Name)
		}
	}
	for name := range want {
		if !seen[name] {
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

func TestBaseURLValidationRejectsBadScheme(t *testing.T) {
	c := circa.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example.com"},
		Secrets: map[string]string{"api_key": "k"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "events", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read with ftp base_url should be rejected (SSRF guard)")
	}
}

func TestRegisteredAndResolvable(t *testing.T) {
	_ = circa.New() // ensure init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("circa"); !ok {
		t.Fatal("registry did not resolve circa (self-registration)")
	}
}
