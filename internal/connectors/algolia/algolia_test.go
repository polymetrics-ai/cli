package algolia_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/algolia"
)

// TestReadIndicesPaginatesAndAuthenticates is the red-first test for the Algolia
// connector. It asserts the two Algolia auth headers, page-number pagination
// over the indices endpoint (which returns {"items":[...],"nbPages":N}), and
// record mapping.
func TestReadIndicesPaginatesAndAuthenticates(t *testing.T) {
	var sawAPIKey, sawAppID string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAPIKey = r.Header.Get("X-Algolia-API-Key")
		sawAppID = r.Header.Get("X-Algolia-Application-Id")
		if r.URL.Path != "/1/indexes" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page") {
		case "", "0":
			_, _ = w.Write([]byte(`{"items":[{"name":"products","entries":10},{"name":"users","entries":5}],"nbPages":2}`))
		case "1":
			_, _ = w.Write([]byte(`{"items":[{"name":"orders","entries":3}],"nbPages":2}`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
			_, _ = w.Write([]byte(`{"items":[],"nbPages":2}`))
		}
	}))
	defer srv.Close()

	c := algolia.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "application_id": "APP123"},
		Secrets: map[string]string{"api_key": "secret_key_abc"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "indices", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAPIKey != "secret_key_abc" {
		t.Fatalf("X-Algolia-API-Key = %q, want secret_key_abc", sawAPIKey)
	}
	if sawAppID != "APP123" {
		t.Fatalf("X-Algolia-Application-Id = %q, want APP123", sawAppID)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	if got[0]["name"] != "products" {
		t.Fatalf("first record name = %v, want products", got[0]["name"])
	}
	for _, rec := range got {
		if rec["name"] == nil {
			t.Fatalf("record missing name: %+v", rec)
		}
	}
}

// TestReadKeys exercises a second stream (api_keys) that reads the "keys" array.
func TestReadKeys(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/1/keys" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"keys":[{"value":"abc","description":"search-only","acl":["search"]}]}`))
	}))
	defer srv.Close()

	c := algolia.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "application_id": "APP123"},
		Secrets: map[string]string{"api_key": "secret_key_abc"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "api_keys", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("records = %d, want 1", len(got))
	}
	if got[0]["value"] != "abc" {
		t.Fatalf("value = %v, want abc", got[0]["value"])
	}
}

// TestFixtureModeNoNetwork confirms fixture mode emits deterministic records
// with no network access, so conformance can run credential-free.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := algolia.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"indices", "api_keys", "index_settings"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("Read(%s) fixture: %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture stream %s emitted no records", stream)
		}
	}
}

// TestCheckFixtureAndCatalog confirms Check short-circuits in fixture mode and
// that the catalog exposes the expected streams.
func TestCheckFixtureAndCatalog(t *testing.T) {
	c := algolia.New()
	if err := c.Check(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "algolia" {
		t.Fatalf("catalog connector = %q, want algolia", cat.Connector)
	}
	want := map[string]bool{"indices": false, "api_keys": false, "index_settings": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
	}
	for name, found := range want {
		if !found {
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

// TestReadValidatesBaseURL rejects a malformed base_url override (SSRF guard).
func TestReadValidatesBaseURL(t *testing.T) {
	c := algolia.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil", "application_id": "APP"},
		Secrets: map[string]string{"api_key": "k"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "indices", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("expected error for non-http base_url scheme")
	}
}

// TestRegisteredReadOnly confirms self-registration and read-only capabilities.
func TestRegisteredReadOnly(t *testing.T) {
	_ = algolia.New() // ensure init ran
	caps := algolia.New().Metadata().Capabilities
	if !caps.Read || caps.Write {
		t.Fatalf("capabilities = %+v, want Read && !Write", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("algolia"); !ok {
		t.Fatal("registry did not resolve algolia (self-registration)")
	}
}
