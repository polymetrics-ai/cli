package gong_test

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/gong"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Gong
// connector: Basic auth (access_key:access_key_secret), Gong cursor pagination
// (records.cursor in body -> cursor query param across two pages), and record
// mapping for the users stream.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/users" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("cursor") {
		case "":
			_, _ = w.Write([]byte(`{"requestId":"a","records":{"totalRecords":3,"currentPageSize":2,"currentPageNumber":0,"cursor":"PAGE2"},"users":[{"id":"u1","emailAddress":"a@example.com"},{"id":"u2","emailAddress":"b@example.com"}]}`))
		case "PAGE2":
			// Final page: no "records" key signals stop.
			_, _ = w.Write([]byte(`{"requestId":"b","users":[{"id":"u3","emailAddress":"c@example.com"}]}`))
		default:
			t.Errorf("unexpected cursor=%q", r.URL.Query().Get("cursor"))
			_, _ = w.Write([]byte(`{"users":[]}`))
		}
	}))
	defer srv.Close()

	c := gong.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"credentials.access_key": "AK", "credentials.access_key_secret": "SK"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "users", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("AK:SK"))
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
		if rec["email_address"] == nil {
			t.Fatalf("record missing mapped email_address: %+v", rec)
		}
	}
}

// TestReadCallsMapsRecord covers the calls stream record mapper and its cursor
// field "started".
func TestReadCallsMapsRecord(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/calls" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"records":{"totalRecords":1,"cursor":null},"calls":[{"id":"c1","title":"Demo","started":"2026-01-02T10:00:00Z","duration":600,"isPrivate":false}]}`))
	}))
	defer srv.Close()

	c := gong.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"credentials.access_key": "AK", "credentials.access_key_secret": "SK"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "calls", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read calls: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("records = %d, want 1", len(got))
	}
	if got[0]["id"] != "c1" || got[0]["title"] != "Demo" || got[0]["started"] != "2026-01-02T10:00:00Z" {
		t.Fatalf("unexpected call record: %+v", got[0])
	}
}

// TestFixtureModeNoNetwork ensures fixture mode emits deterministic records with
// no network access so conformance works credential-free.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := gong.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	for _, stream := range []string{"users", "calls", "scorecards"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("Read fixture %s: %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture %s emitted no records", stream)
		}
	}

	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
}

func TestCatalogAndMetadata(t *testing.T) {
	c := gong.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("gong is read-only; Write should be false")
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("streams = %d, want >= 3", len(cat.Streams))
	}
}

func TestRegisteredViaRegistry(t *testing.T) {
	_ = gong.New() // ensure init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("gong"); !ok {
		t.Fatal("registry did not resolve gong (self-registration)")
	}
}

func TestBaseURLSSRFValidation(t *testing.T) {
	c := gong.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil"},
		Secrets: map[string]string{"credentials.access_key": "AK", "credentials.access_key_secret": "SK"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "users", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("expected error for non-http(s) base_url scheme")
	}
}
