package aviationstack_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/aviationstack"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the aviationstack
// connector: access_key query auth, limit/offset pagination over a {pagination,
// data[]} envelope, and record mapping. Red until the package exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawKey string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawKey = r.URL.Query().Get("access_key")
		if r.URL.Path != "/airlines" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("offset") {
		case "", "0":
			// First page: count == limit, so the connector must request a second page.
			_, _ = w.Write([]byte(`{"pagination":{"limit":2,"offset":0,"count":2,"total":3},"data":[{"id":"1","airline_name":"Alpha Air","iata_code":"AA","icao_code":"AAA","country_name":"Wonderland"},{"id":"2","airline_name":"Beta Air","iata_code":"BA","icao_code":"BAA","country_name":"Wonderland"}]}`))
		case "2":
			_, _ = w.Write([]byte(`{"pagination":{"limit":2,"offset":2,"count":1,"total":3},"data":[{"id":"3","airline_name":"Gamma Air","iata_code":"GA","icao_code":"GAA","country_name":"Wonderland"}]}`))
		default:
			t.Errorf("unexpected offset=%q", r.URL.Query().Get("offset"))
			_, _ = w.Write([]byte(`{"pagination":{"limit":2,"offset":99,"count":0,"total":3},"data":[]}`))
		}
	}))
	defer srv.Close()

	c := aviationstack.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"access_key": "key_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "airlines", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawKey != "key_test_123" {
		t.Fatalf("access_key = %q, want key_test_123", sawKey)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["airline_name"] == nil {
			t.Fatalf("record missing id/airline_name: %+v", rec)
		}
	}
}

// TestFixtureModeNeedsNoNetwork confirms mode=fixture emits deterministic records
// without any network access or credentials, as the conformance harness requires.
func TestFixtureModeNeedsNoNetwork(t *testing.T) {
	c := aviationstack.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "airports", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("fixture Read: %v", err)
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
		t.Fatalf("fixture Check: %v", err)
	}
}

// TestCatalogStreams confirms the published catalog exposes the core streams.
func TestCatalogStreams(t *testing.T) {
	c := aviationstack.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"flights": false, "airlines": false, "airports": false, "airplanes": false, "countries": false}
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

// TestBaseURLRejectsBadScheme guards the SSRF validation on base_url overrides.
func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := aviationstack.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example.com"},
		Secrets: map[string]string{"access_key": "k"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "airlines", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read with ftp base_url should be rejected")
	}
}

// TestRegisteredReadOnly verifies self-registration via the registry and that the
// connector advertises read-only capabilities.
func TestRegisteredReadOnly(t *testing.T) {
	c := aviationstack.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (aviationstack is read-only)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("aviationstack"); !ok {
		t.Fatal("registry did not resolve aviationstack (self-registration)")
	}
}
