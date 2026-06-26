package basespace_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	basespace "polymetrics.ai/internal/connectors/illumina-basespace"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Illumina
// BaseSpace connector: x-access-token header auth, Offset/Limit pagination over
// the Response.Items envelope across two pages, and record mapping. Red until
// internal/connectors/illumina-basespace is implemented.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawToken string
	var sawPaths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawToken = r.Header.Get("x-access-token")
		sawPaths = append(sawPaths, r.URL.Path)
		if r.URL.Path != "/v1pre3/users/current/projects" {
			http.NotFound(w, r)
			return
		}
		limit := r.URL.Query().Get("Limit")
		offset := r.URL.Query().Get("Offset")
		if limit == "" {
			t.Errorf("expected Limit query param")
		}
		switch offset {
		case "", "0":
			// First page is full (2 == limit when limit is 2), so a second page is fetched.
			_, _ = w.Write([]byte(`{"Response":{"Items":[` +
				`{"Id":"p1","Name":"Project One","DateCreated":"2026-01-01T00:00:00Z","TotalSize":100},` +
				`{"Id":"p2","Name":"Project Two","DateCreated":"2026-01-02T00:00:00Z","TotalSize":200}` +
				`],"Offset":0,"Limit":2,"TotalCount":3}}`))
		case "2":
			_, _ = w.Write([]byte(`{"Response":{"Items":[` +
				`{"Id":"p3","Name":"Project Three","DateCreated":"2026-01-03T00:00:00Z","TotalSize":300}` +
				`],"Offset":2,"Limit":2,"TotalCount":3}}`))
		default:
			t.Errorf("unexpected Offset=%q", offset)
			_, _ = w.Write([]byte(`{"Response":{"Items":[],"Offset":0,"Limit":2,"TotalCount":3}}`))
		}
	}))
	defer srv.Close()

	c := basespace.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":  srv.URL,
			"user":      "current",
			"page_size": "2",
		},
		Secrets: map[string]string{"access_token": "tok_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "projects", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawToken != "tok_test_123" {
		t.Fatalf("x-access-token = %q, want tok_test_123", sawToken)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 across 2 pages; paths=%v", len(got), sawPaths)
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["name"] == nil {
			t.Fatalf("record missing mapped id/name: %+v", rec)
		}
	}
	// Confirm the second page was actually requested (pagination, not a single call).
	if len(sawPaths) < 2 {
		t.Fatalf("expected at least 2 page requests, got %d (%v)", len(sawPaths), sawPaths)
	}
}

// TestReadMapsRunFields verifies the runs stream mapper flattens PascalCase API
// fields onto the snake_case record shape.
func TestReadMapsRunFields(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1pre3/users/current/runs" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"Response":{"Items":[` +
			`{"Id":"r1","Name":"Run A","ExperimentName":"Exp1","Status":"Complete","DateCreated":"2026-02-01T00:00:00Z"}` +
			`],"Offset":0,"Limit":100,"TotalCount":1}}`))
	}))
	defer srv.Close()

	c := basespace.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "user": "current"},
		Secrets: map[string]string{"access_token": "tok"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "runs", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read runs: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("records = %d, want 1", len(got))
	}
	rec := got[0]
	if rec["id"] != "r1" || rec["name"] != "Run A" || rec["status"] != "Complete" || rec["experiment_name"] != "Exp1" {
		t.Fatalf("run record mismatch: %+v", rec)
	}
}

// TestFixtureModeNoNetwork verifies fixture mode emits deterministic records
// without any network access (mandatory for credential-free conformance).
func TestFixtureModeNoNetwork(t *testing.T) {
	c := basespace.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{"mode": "fixture"},
	}
	for _, stream := range []string{"projects", "runs", "samples", "appsessions", "datasets"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if len(got) != 2 {
			t.Fatalf("fixture stream %s: records = %d, want 2", stream, len(got))
		}
		for i, rec := range got {
			if rec["id"] == nil {
				t.Fatalf("fixture stream %s record %d missing id: %+v", stream, i, rec)
			}
		}
	}
	// Check must also short-circuit in fixture mode (no creds).
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

func TestCheckRequiresToken(t *testing.T) {
	c := basespace.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "https://example.basespace.illumina.com"},
		Secrets: map[string]string{},
	}
	if err := c.Check(context.Background(), cfg); err == nil {
		t.Fatal("Check should fail without access_token")
	}
}

func TestBaseURLSSRFGuard(t *testing.T) {
	c := basespace.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil"},
		Secrets: map[string]string{"access_token": "tok"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "projects", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read should reject a non-http(s) base_url")
	}
}

func TestCatalogStreams(t *testing.T) {
	c := basespace.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"projects": true, "runs": true, "samples": true, "appsessions": true, "datasets": true}
	for _, s := range cat.Streams {
		delete(want, s.Name)
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %s missing primary key", s.Name)
		}
	}
	if len(want) != 0 {
		t.Fatalf("catalog missing streams: %v", want)
	}
}

func TestRegisteredReadOnly(t *testing.T) {
	_ = basespace.New() // ensure init ran
	c := basespace.New()
	if c.Name() != "illumina-basespace" {
		t.Fatalf("Name = %q, want illumina-basespace", c.Name())
	}
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || caps.Write {
		t.Fatalf("capabilities = %+v, want Read+Catalog, no Write", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("illumina-basespace"); !ok {
		t.Fatal("registry did not resolve illumina-basespace (self-registration)")
	}
}
