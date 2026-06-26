package ninjaonermm_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	ninjaonermm "polymetrics.ai/internal/connectors/ninjaone-rmm"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the NinjaOne RMM
// connector: Bearer auth, after-cursor pagination over a top-level JSON array,
// and record mapping. NinjaOne v2 list endpoints return a bare array and page
// forward with pageSize + after=<last id>.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/v2/organizations" {
			http.NotFound(w, r)
			return
		}
		calls++
		switch r.URL.Query().Get("after") {
		case "":
			_, _ = w.Write([]byte(`[{"id":1,"name":"Org One"},{"id":2,"name":"Org Two"}]`))
		case "2":
			_, _ = w.Write([]byte(`[{"id":3,"name":"Org Three"}]`))
		default:
			t.Errorf("unexpected after=%q", r.URL.Query().Get("after"))
			_, _ = w.Write([]byte(`[]`))
		}
	}))
	defer srv.Close()

	c := ninjaonermm.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"api_key": "tok_abc123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "organizations", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer tok_abc123" {
		t.Fatalf("Authorization = %q, want Bearer tok_abc123", sawAuth)
	}
	if calls != 2 {
		t.Fatalf("server calls = %d, want 2 (pagination across 2 pages)", calls)
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

// TestReadDevicesMapping verifies a second stream maps the documented device
// fields off a top-level array.
func TestReadDevicesMapping(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/devices" {
			http.NotFound(w, r)
			return
		}
		if r.URL.Query().Get("after") != "" {
			_, _ = w.Write([]byte(`[]`))
			return
		}
		_, _ = w.Write([]byte(`[{"id":10,"systemName":"host-10","organizationId":1,"nodeClass":"WINDOWS_WORKSTATION","offline":false}]`))
	}))
	defer srv.Close()

	c := ninjaonermm.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "tok"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "devices", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("records = %d, want 1", len(got))
	}
	rec := got[0]
	if rec["id"] == nil || rec["system_name"] != "host-10" || rec["node_class"] != "WINDOWS_WORKSTATION" {
		t.Fatalf("device mapping wrong: %+v", rec)
	}
}

// TestFixtureModeNoNetwork ensures fixture mode emits deterministic records with
// no network access so conformance runs credential-free.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := ninjaonermm.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "organizations", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read fixture: %v", err)
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
		t.Fatalf("Check fixture: %v", err)
	}
}

// TestCatalogStreams asserts the published catalog includes the core streams.
func TestCatalogStreams(t *testing.T) {
	c := ninjaonermm.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"organizations": false, "devices": false, "locations": false, "activities": false, "policies": false}
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
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

// TestRegistryResolves confirms self-registration via the global registry.
func TestRegistryResolves(t *testing.T) {
	_ = ninjaonermm.New() // ensure init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("ninjaone-rmm"); !ok {
		t.Fatal("registry did not resolve ninjaone-rmm (self-registration)")
	}
}

// TestReadUnknownStream rejects unknown streams.
func TestReadUnknownStream(t *testing.T) {
	c := ninjaonermm.New()
	cfg := connectors.RuntimeConfig{Secrets: map[string]string{"api_key": "tok"}}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "does_not_exist", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read of unknown stream should error")
	}
}
