package convertkit_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/convertkit"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the ConvertKit
// connector: api_secret query-param auth, page-based pagination over the
// subscribers[] array using total_pages, and record mapping. Red until
// internal/connectors/convertkit exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawSecret string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawSecret = r.URL.Query().Get("api_secret")
		if r.URL.Path != "/subscribers" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page") {
		case "", "1":
			_, _ = w.Write([]byte(`{"total_subscribers":3,"page":1,"total_pages":2,"subscribers":[{"id":1,"email_address":"a@example.com","state":"active","created_at":"2026-01-01T00:00:00Z"},{"id":2,"email_address":"b@example.com","state":"active","created_at":"2026-01-02T00:00:00Z"}]}`))
		case "2":
			_, _ = w.Write([]byte(`{"total_subscribers":3,"page":2,"total_pages":2,"subscribers":[{"id":3,"email_address":"c@example.com","state":"active","created_at":"2026-01-03T00:00:00Z"}]}`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
			_, _ = w.Write([]byte(`{"page":3,"total_pages":2,"subscribers":[]}`))
		}
	}))
	defer srv.Close()

	c := convertkit.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "secret_abc"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "subscribers", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawSecret != "secret_abc" {
		t.Fatalf("api_secret = %q, want secret_abc", sawSecret)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["email_address"] == nil {
			t.Fatalf("record missing id/email_address: %+v", rec)
		}
	}
}

// TestReadFormsSingleArrayPage verifies the unpaginated resource streams (forms,
// tags, sequences) that return a single array under their resource key.
func TestReadFormsSingleArrayPage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/forms" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"forms":[{"id":10,"name":"Newsletter","created_at":"2026-01-01T00:00:00Z"},{"id":11,"name":"Webinar","created_at":"2026-01-02T00:00:00Z"}]}`))
	}))
	defer srv.Close()

	c := convertkit.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "secret_abc"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "forms", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2", len(got))
	}
	if got[0]["name"] != "Newsletter" {
		t.Fatalf("form name = %v, want Newsletter", got[0]["name"])
	}
}

// TestFixtureModeNoNetwork confirms credential-free fixture conformance: no
// network, deterministic records, every published stream supported.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := convertkit.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"subscribers", "forms", "sequences", "tags", "broadcasts"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture Read(%s) returned no records", stream)
		}
		if got[0]["id"] == nil {
			t.Fatalf("fixture %s record missing id: %+v", stream, got[0])
		}
	}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

// TestCatalogAndMetadata asserts the published catalog and read-only capabilities.
func TestCatalogAndMetadata(t *testing.T) {
	c := convertkit.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("convertkit must be read-only; Write should be false")
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 5 {
		t.Fatalf("streams = %d, want >= 5", len(cat.Streams))
	}
}

// TestRegisteredResolvesFromRegistry asserts self-registration via init().
func TestRegisteredResolvesFromRegistry(t *testing.T) {
	_ = convertkit.New() // ensure package init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("convertkit"); !ok {
		t.Fatal("registry did not resolve convertkit (self-registration)")
	}
}

// TestBadBaseURLRejected guards the SSRF validation on base_url override.
func TestBadBaseURLRejected(t *testing.T) {
	c := convertkit.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil"},
		Secrets: map[string]string{"api_key": "secret_abc"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "forms", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read with invalid base_url scheme should fail")
	}
}
