package eventee_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/eventee"
)

// TestReadContentStreamAuthAndMapping is the red-first test for the Eventee
// connector: Bearer auth on the Authorization header, extraction of records from
// the nested field in the /content response, and record mapping. Most Eventee
// streams (lectures, speakers, days, halls, tracks, workshops, pauses) share the
// /content endpoint and select a nested array by their stream name.
func TestReadContentStreamAuthAndMapping(t *testing.T) {
	var sawAuth string
	var sawPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		sawPath = r.URL.Path
		if r.URL.Path != "/content" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{
			"lectures":[
				{"id":1,"name":"Opening Keynote","hall_id":10,"start":"2026-06-01T09:00:00Z"},
				{"id":2,"name":"Closing Panel","hall_id":11,"start":"2026-06-01T17:00:00Z"}
			],
			"speakers":[{"id":99,"name":"Ada"}]
		}`))
	}))
	defer srv.Close()

	c := eventee.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_token": "tok_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "lectures", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer tok_123" {
		t.Fatalf("Authorization = %q, want Bearer tok_123", sawAuth)
	}
	if sawPath != "/content" {
		t.Fatalf("path = %q, want /content", sawPath)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2 lectures", len(got))
	}
	if got[0]["id"] == nil || got[0]["name"] != "Opening Keynote" {
		t.Fatalf("first record not mapped: %+v", got[0])
	}
	if got[0]["hall_id"] == nil {
		t.Fatalf("first record missing hall_id: %+v", got[0])
	}
}

// TestReadRootArrayStream covers the streams whose endpoint returns a top-level
// JSON array (partners at /partners, participants at /participants). The
// connector must read across both the dedicated endpoint and the root array.
func TestReadRootArrayStream(t *testing.T) {
	var sawPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawPath = r.URL.Path
		switch r.URL.Path {
		case "/partners":
			_, _ = w.Write([]byte(`[{"id":1,"company":"Acme","email":"a@acme.test"},{"id":2,"company":"Globex","email":"b@globex.test"}]`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := eventee.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_token": "tok_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "partners", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawPath != "/partners" {
		t.Fatalf("path = %q, want /partners", sawPath)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2 partners", len(got))
	}
	if got[1]["company"] != "Globex" {
		t.Fatalf("second partner not mapped: %+v", got[1])
	}
}

// TestFixtureMode verifies the credential-free fixture path so conformance can
// run without a live token.
func TestFixtureMode(t *testing.T) {
	c := eventee.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "speakers", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read(fixture): %v", err)
	}
	if len(got) == 0 {
		t.Fatal("fixture mode produced no records")
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("fixture record missing id: %+v", rec)
		}
	}
}

// TestCatalogStreams verifies the published catalog exposes the core streams.
func TestCatalogStreams(t *testing.T) {
	c := eventee.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"lectures": false, "speakers": false, "days": false, "halls": false, "tracks": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q missing primary key", s.Name)
		}
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("catalog missing core stream %q", name)
		}
	}
}

// TestRegistryResolution confirms self-registration through the connectors
// registry.
func TestRegistryResolution(t *testing.T) {
	_ = eventee.New() // ensure init ran
	r := connectors.NewRegistry()
	got, ok := r.Get("eventee")
	if !ok {
		t.Fatal("registry did not resolve eventee (self-registration)")
	}
	caps := got.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("eventee should be read-only, got Write=true")
	}
}

// TestUnknownStream rejects a stream that is not in the routing table.
func TestUnknownStream(t *testing.T) {
	c := eventee.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "nope", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read(unknown stream) should error")
	}
}
