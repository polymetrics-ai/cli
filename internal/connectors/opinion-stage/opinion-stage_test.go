package opinionstage_test

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	opinionstage "polymetrics.ai/internal/connectors/opinion-stage"
)

// TestReadItemsPaginatesAndAuthenticates is the red-first test: Basic auth with
// the api_key as username and empty password, page[number]/page[size]
// pagination over the data[] array across two pages, and JSON:API attribute
// flattening.
func TestReadItemsPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/api/v2/items" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page[number]") {
		case "", "1":
			_, _ = w.Write([]byte(`{"data":[
				{"id":"i1","type":"item","attributes":{"title":"Poll A","status":"published","timestamps":{"created":"2026-01-01T00:00:00Z","modified":"2026-01-02T00:00:00Z"}}},
				{"id":"i2","type":"item","attributes":{"title":"Quiz B","status":"draft","timestamps":{"created":"2026-01-03T00:00:00Z","modified":"2026-01-04T00:00:00Z"}}}
			]}`))
		case "2":
			_, _ = w.Write([]byte(`{"data":[
				{"id":"i3","type":"item","attributes":{"title":"Form C","status":"published","timestamps":{"created":"2026-01-05T00:00:00Z","modified":"2026-01-06T00:00:00Z"}}}
			]}`))
		default:
			_, _ = w.Write([]byte(`{"data":[]}`))
		}
	}))
	defer srv.Close()

	c := opinionstage.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"api_key": "key_abc"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "items", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("key_abc:"))
	if sawAuth != wantAuth {
		t.Fatalf("Authorization = %q, want %q", sawAuth, wantAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	if got[0]["id"] != "i1" || got[0]["title"] != "Poll A" || got[0]["status"] != "published" {
		t.Fatalf("first record not flattened correctly: %+v", got[0])
	}
	if got[0]["created"] != "2026-01-01T00:00:00Z" {
		t.Fatalf("created not flattened from timestamps: %+v", got[0])
	}
}

// TestReadResponsesSubstream verifies the substream walks each item and reads
// its per-item responses endpoint.
func TestReadResponsesSubstream(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v2/items":
			_, _ = w.Write([]byte(`{"data":[{"id":"i1","type":"item","attributes":{"title":"A"}}]}`))
		case "/api/v2/items/i1/responses":
			if r.URL.Query().Get("page[number]") == "2" {
				_, _ = w.Write([]byte(`{"data":[]}`))
				return
			}
			_, _ = w.Write([]byte(`{"data":[
				{"id":"r1","type":"response","attributes":{"result":{"title":"Cat"},"timestamps":{"created":"2026-02-01T00:00:00Z"}}}
			]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := opinionstage.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "key_abc"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "responses", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("records = %d, want 1", len(got))
	}
	if got[0]["id"] != "r1" || got[0]["item_id"] != "i1" {
		t.Fatalf("response record not mapped with item_id: %+v", got[0])
	}
}

// TestFixtureModeNoNetwork verifies fixture mode emits deterministic records
// without any network access (required for credential-free conformance).
func TestFixtureModeNoNetwork(t *testing.T) {
	c := opinionstage.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	for _, stream := range []string{"items", "responses", "questions"} {
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
	c := opinionstage.New()
	if err := c.Check(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

func TestRegisteredReadOnly(t *testing.T) {
	_ = opinionstage.New() // ensure init ran
	caps := opinionstage.New().Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("opinion-stage"); !ok {
		t.Fatal("registry did not resolve opinion-stage (self-registration)")
	}
}
