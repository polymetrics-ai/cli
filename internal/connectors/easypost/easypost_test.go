package easypost_test

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/easypost"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the EasyPost
// connector: HTTP Basic auth (API key as username, empty password), EasyPost
// has_more/before_id pagination over the resource-named array, and record
// mapping. Red until internal/connectors/easypost exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var pages int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/shipments" {
			http.NotFound(w, r)
			return
		}
		pages++
		switch r.URL.Query().Get("before_id") {
		case "":
			_, _ = w.Write([]byte(`{"shipments":[{"id":"shp_1","object":"Shipment","created_at":"2026-01-01T00:00:00Z","status":"delivered"},{"id":"shp_2","object":"Shipment","created_at":"2026-01-02T00:00:00Z","status":"in_transit"}],"has_more":true}`))
		case "shp_2":
			_, _ = w.Write([]byte(`{"shipments":[{"id":"shp_3","object":"Shipment","created_at":"2026-01-03T00:00:00Z","status":"unknown"}],"has_more":false}`))
		default:
			t.Errorf("unexpected before_id=%q", r.URL.Query().Get("before_id"))
			_, _ = w.Write([]byte(`{"shipments":[],"has_more":false}`))
		}
	}))
	defer srv.Close()

	c := easypost.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"username": "EZAK_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "shipments", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("EZAK_test_123:"))
	if sawAuth != wantAuth {
		t.Fatalf("Authorization = %q, want %q", sawAuth, wantAuth)
	}
	if pages != 2 {
		t.Fatalf("server pages = %d, want 2", pages)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["created_at"] == nil {
			t.Fatalf("record missing id/created_at: %+v", rec)
		}
		if rec["status"] == nil {
			t.Fatalf("record missing mapped status field: %+v", rec)
		}
	}
}

// TestReadFixtureModeNoNetwork confirms fixture mode emits deterministic records
// without any network access so credential-free conformance passes.
func TestReadFixtureModeNoNetwork(t *testing.T) {
	c := easypost.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "trackers", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read fixture: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("fixture records = %d, want 2", len(got))
	}
	for _, rec := range got {
		if id, _ := rec["id"].(string); id == "" || !strings.Contains(id, "trackers") {
			t.Fatalf("fixture record missing/odd id: %+v", rec)
		}
	}
}

// TestCheckFixtureNoSecret confirms fixture-mode Check short-circuits without
// requiring a secret or a network call.
func TestCheckFixtureNoSecret(t *testing.T) {
	c := easypost.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

// TestCatalogStreams confirms the published catalog exposes the core streams with
// a primary key.
func TestCatalogStreams(t *testing.T) {
	c := easypost.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"shipments": false, "trackers": false, "addresses": false, "parcels": false, "insurances": false}
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

// TestRegistryResolves confirms self-registration via init() resolves through the
// shared registry.
func TestRegistryResolves(t *testing.T) {
	_ = easypost.New() // ensure init ran
	r := connectors.NewRegistry()
	c, ok := r.Get("easypost")
	if !ok {
		t.Fatal("registry did not resolve easypost (self-registration)")
	}
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("easypost is read-only; Write should be false, got %+v", caps)
	}
}

// TestUnknownStreamRejected confirms an unknown stream is an error, not a silent
// empty read.
func TestUnknownStreamRejected(t *testing.T) {
	c := easypost.New()
	err := c.Read(context.Background(), connectors.ReadRequest{
		Stream: "not_a_stream",
		Config: connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}},
	}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read(unknown stream) should error")
	}
}
