package coda_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/coda"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Coda
// connector: Bearer auth, Coda nextPageToken/pageToken pagination over items[],
// and record mapping across two pages.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/docs" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("pageToken") {
		case "":
			_, _ = w.Write([]byte(`{"items":[{"id":"d_1","type":"doc","name":"Doc One"},{"id":"d_2","type":"doc","name":"Doc Two"}],"nextPageToken":"tok_2"}`))
		case "tok_2":
			_, _ = w.Write([]byte(`{"items":[{"id":"d_3","type":"doc","name":"Doc Three"}]}`))
		default:
			t.Errorf("unexpected pageToken=%q", r.URL.Query().Get("pageToken"))
			_, _ = w.Write([]byte(`{"items":[]}`))
		}
	}))
	defer srv.Close()

	c := coda.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"auth_token": "tok_secret_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "docs", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer tok_secret_123" {
		t.Fatalf("Authorization = %q, want Bearer tok_secret_123", sawAuth)
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

// TestReadDocScopedStreamUsesDocID confirms a doc-scoped stream (tables) builds
// the /docs/{docId}/tables path from the doc_id config.
func TestReadDocScopedStreamUsesDocID(t *testing.T) {
	var sawPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawPath = r.URL.Path
		_, _ = w.Write([]byte(`{"items":[{"id":"grid_1","type":"table","name":"Tasks"}]}`))
	}))
	defer srv.Close()

	c := coda.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "doc_id": "AbCDeFGH"},
		Secrets: map[string]string{"auth_token": "tok_secret_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "tables", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawPath != "/docs/AbCDeFGH/tables" {
		t.Fatalf("path = %q, want /docs/AbCDeFGH/tables", sawPath)
	}
	if len(got) != 1 || got[0]["id"] != "grid_1" {
		t.Fatalf("records = %+v, want one table grid_1", got)
	}
}

// TestReadDocScopedRequiresDocID confirms a doc-scoped stream without doc_id is
// rejected rather than hitting a malformed URL.
func TestReadDocScopedRequiresDocID(t *testing.T) {
	c := coda.New()
	cfg := connectors.RuntimeConfig{
		Secrets: map[string]string{"auth_token": "tok_secret_123"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "tables", Config: cfg}, func(connectors.Record) error {
		return nil
	})
	if err == nil {
		t.Fatal("Read(tables) without doc_id should error")
	}
}

// TestFixtureModeNeedsNoNetwork confirms fixture mode emits deterministic
// records without credentials or a server.
func TestFixtureModeNeedsNoNetwork(t *testing.T) {
	c := coda.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture", "doc_id": "fixture_doc"}}

	for _, stream := range []string{"docs", "tables", "pages", "formulas", "controls"} {
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
}

func TestCheckFixtureMode(t *testing.T) {
	c := coda.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

func TestCatalogStreams(t *testing.T) {
	c := coda.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"docs": false, "tables": false, "pages": false, "formulas": false, "controls": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

func TestRegisteredReadOnly(t *testing.T) {
	_ = coda.New() // ensure init ran
	c := coda.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("coda is a read-only source; Write should be false, caps=%+v", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("coda"); !ok {
		t.Fatal("registry did not resolve coda (self-registration)")
	}
}
