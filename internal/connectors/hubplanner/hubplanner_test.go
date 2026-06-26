package hubplanner_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/hubplanner"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Hubplanner
// connector: it asserts the raw API-key Authorization header (no "Bearer"
// prefix), page/limit pagination over a top-level JSON array stopping on a short
// page, and _id-based record mapping. Red until internal/connectors/hubplanner
// is implemented.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var sawPaths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/resource" {
			http.NotFound(w, r)
			return
		}
		page := r.URL.Query().Get("page")
		sawPaths = append(sawPaths, page)
		if r.URL.Query().Get("limit") == "" {
			t.Errorf("expected limit query param, got none")
		}
		w.Header().Set("Content-Type", "application/json")
		switch page {
		case "0":
			// Full page of 2 (page_size=2) -> paginator must fetch page 1.
			_, _ = w.Write([]byte(`[{"_id":"r1","firstName":"Ada","email":"ada@example.com","status":"STATUS_ACTIVE"},{"_id":"r2","firstName":"Grace","email":"grace@example.com","status":"STATUS_ACTIVE"}]`))
		case "1":
			// Short page (1 < page_size) -> stop.
			_, _ = w.Write([]byte(`[{"_id":"r3","firstName":"Katherine","email":"kat@example.com","status":"STATUS_ACTIVE"}]`))
		default:
			t.Errorf("unexpected page=%q", page)
			_, _ = w.Write([]byte(`[]`))
		}
	}))
	defer srv.Close()

	c := hubplanner.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"api_key": "hp_secret_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "resources", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "hp_secret_123" {
		t.Fatalf("Authorization = %q, want raw api key hp_secret_123 (no Bearer prefix)", sawAuth)
	}
	if len(sawPaths) != 2 {
		t.Fatalf("requested pages = %v, want exactly 2 (pages 0 and 1)", sawPaths)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 across 2 pages", len(got))
	}
	for _, rec := range got {
		if rec["_id"] == nil {
			t.Fatalf("record missing _id: %+v", rec)
		}
	}
}

// TestFixtureModeNeedsNoNetwork ensures conformance can run credential-free.
func TestFixtureModeNeedsNoNetwork(t *testing.T) {
	c := hubplanner.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "projects", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read(fixture) = %v", err)
	}
	if len(got) == 0 {
		t.Fatal("fixture mode emitted no records")
	}
	for _, rec := range got {
		if rec["_id"] == nil {
			t.Fatalf("fixture record missing _id: %+v", rec)
		}
	}
}

func TestCatalogStreams(t *testing.T) {
	c := hubplanner.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "hubplanner" {
		t.Fatalf("catalog connector = %q, want hubplanner", cat.Connector)
	}
	want := map[string]bool{"resources": false, "projects": false, "clients": false, "events": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q has no primary key", s.Name)
		}
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("catalog missing expected stream %q", name)
		}
	}
}

func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := hubplanner.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example.com"},
		Secrets: map[string]string{"api_key": "x"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "resources", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("expected error for non-http(s) base_url, got nil")
	}
}

func TestRegistryResolution(t *testing.T) {
	_ = hubplanner.New() // ensure package init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("hubplanner"); !ok {
		t.Fatal("registry did not resolve hubplanner (self-registration)")
	}
	caps := hubplanner.New().Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only)", caps)
	}
}
