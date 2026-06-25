package copper_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/copper"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Copper
// connector: it asserts the three Copper auth headers (X-PW-AccessToken,
// X-PW-Application, X-PW-UserEmail), POST /people/search body pagination over a
// top-level JSON array across two pages, and record mapping.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawToken, sawApp, sawEmail string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawToken = r.Header.Get("X-PW-AccessToken")
		sawApp = r.Header.Get("X-PW-Application")
		sawEmail = r.Header.Get("X-PW-UserEmail")
		if r.Method != http.MethodPost || r.URL.Path != "/people/search" {
			http.NotFound(w, r)
			return
		}
		body, _ := io.ReadAll(r.Body)
		var payload struct {
			PageNumber int `json:"page_number"`
			PageSize   int `json:"page_size"`
		}
		_ = json.Unmarshal(body, &payload)
		if payload.PageSize != 2 {
			t.Errorf("page_size = %d, want 2", payload.PageSize)
		}
		switch payload.PageNumber {
		case 1:
			// full page of 2 => paginator must request page 2
			_, _ = w.Write([]byte(`[{"id":1,"name":"Ada","date_modified":1700000000},{"id":2,"name":"Grace","date_modified":1700000100}]`))
		case 2:
			// short page => stop
			_, _ = w.Write([]byte(`[{"id":3,"name":"Katherine","date_modified":1700000200}]`))
		default:
			t.Errorf("unexpected page_number=%d", payload.PageNumber)
			_, _ = w.Write([]byte(`[]`))
		}
	}))
	defer srv.Close()

	c := copper.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "user_email": "me@example.com", "page_size": "2"},
		Secrets: map[string]string{"api_key": "copper_secret_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "people", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawToken != "copper_secret_123" {
		t.Fatalf("X-PW-AccessToken = %q, want copper_secret_123", sawToken)
	}
	if sawApp == "" {
		t.Fatalf("X-PW-Application header was not sent")
	}
	if sawEmail != "me@example.com" {
		t.Fatalf("X-PW-UserEmail = %q, want me@example.com", sawEmail)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["name"] == nil {
			t.Fatalf("record missing id/name: %+v", rec)
		}
	}
}

// TestFixtureModeNoNetwork confirms fixture mode emits deterministic records
// without any network access, so conformance runs credential-free.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := copper.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "companies", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read(fixture): %v", err)
	}
	if len(got) == 0 {
		t.Fatal("fixture mode emitted no records")
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("fixture record missing id: %+v", rec)
		}
	}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

// TestCatalogStreams verifies the published catalog exposes the core streams
// with primary keys.
func TestCatalogStreams(t *testing.T) {
	c := copper.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"people": false, "companies": false, "opportunities": false, "leads": false, "tasks": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q missing primary key", s.Name)
		}
	}
	for name, found := range want {
		if !found {
			t.Fatalf("catalog missing core stream %q", name)
		}
	}
}

// TestRegistryResolution confirms the connector self-registers and resolves via
// the shared registry.
func TestRegistryResolution(t *testing.T) {
	_ = copper.New() // ensure init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("copper"); !ok {
		t.Fatal("registry did not resolve copper (self-registration)")
	}
	caps := copper.New().Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
}
