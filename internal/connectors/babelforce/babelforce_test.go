package babelforce_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/babelforce"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Babelforce
// connector: dual-header auth (X-Auth-Access-ID / X-Auth-Access-Token),
// pagination over two pages via response.pagination.current+1 -> page param,
// records extracted at "items", and record mapping. Red until
// internal/connectors/babelforce exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawID, sawToken string
	var sawMax string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawID = r.Header.Get("X-Auth-Access-ID")
		sawToken = r.Header.Get("X-Auth-Access-Token")
		sawMax = r.URL.Query().Get("max")
		if r.URL.Path != "/api/v2/calls/reporting/simple" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page") {
		case "", "0":
			// First page: pagination has current -> there is a next page.
			_, _ = w.Write([]byte(`{"items":[{"id":"c1","dateCreated":"1700000000","state":"finished"},{"id":"c2","dateCreated":"1700000100","state":"finished"}],"pagination":{"current":0,"max":1}}`))
		case "1":
			// Last page: no "current" key -> stop condition fires.
			_, _ = w.Write([]byte(`{"items":[{"id":"c3","dateCreated":"1700000200","state":"finished"}],"pagination":{"max":1}}`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
			_, _ = w.Write([]byte(`{"items":[],"pagination":{}}`))
		}
	}))
	defer srv.Close()

	c := babelforce.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL + "/api/v2"},
		Secrets: map[string]string{"access_key_id": "ak_123", "access_token": "tok_456"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "calls", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawID != "ak_123" {
		t.Fatalf("X-Auth-Access-ID = %q, want ak_123", sawID)
	}
	if sawToken != "tok_456" {
		t.Fatalf("X-Auth-Access-Token = %q, want tok_456", sawToken)
	}
	if sawMax == "" {
		t.Fatalf("page size param max not sent")
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["dateCreated"] == nil {
			t.Fatalf("record missing id/dateCreated: %+v", rec)
		}
	}
}

// TestFixtureModeNoNetwork ensures fixture mode emits deterministic records with
// no network access, so conformance can run credential-free.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := babelforce.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "calls", Config: cfg}, func(rec connectors.Record) error {
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
	// Check must also short-circuit in fixture mode (no creds, no network).
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

// TestCatalogStreams asserts the published catalog exposes the calls stream with
// the right primary key and cursor field.
func TestCatalogStreams(t *testing.T) {
	c := babelforce.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	var calls *connectors.Stream
	for i := range cat.Streams {
		if cat.Streams[i].Name == "calls" {
			calls = &cat.Streams[i]
		}
	}
	if calls == nil {
		t.Fatal("catalog missing calls stream")
	}
	if len(calls.PrimaryKey) != 1 || calls.PrimaryKey[0] != "id" {
		t.Fatalf("calls primary key = %v, want [id]", calls.PrimaryKey)
	}
	if len(calls.CursorFields) != 1 || calls.CursorFields[0] != "dateCreated" {
		t.Fatalf("calls cursor fields = %v, want [dateCreated]", calls.CursorFields)
	}
}

// TestRegisteredReadOnly verifies self-registration and read-only capabilities.
func TestRegisteredReadOnly(t *testing.T) {
	_ = babelforce.New() // ensure init ran
	c := babelforce.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("babelforce"); !ok {
		t.Fatal("registry did not resolve babelforce (self-registration)")
	}
}
