package flexport_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/flexport"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Flexport
// connector: Bearer auth, Flexport cursor pagination (records at data.data with a
// data.next absolute URL), and record mapping. Two pages are served and three
// records expected.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/products" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Query().Get("page") {
		case "", "1":
			// First page advertises an absolute next URL on the same test server.
			_, _ = w.Write([]byte(`{"_object":"list","data":{"_object":"list","prev":null,"next":"` + srv.URL + `/products?page=2","data":[{"id":"1","name":"Widget A","sku":"SKU-A"},{"id":"2","name":"Widget B","sku":"SKU-B"}]}}`))
		case "2":
			_, _ = w.Write([]byte(`{"_object":"list","data":{"_object":"list","prev":null,"next":null,"data":[{"id":"3","name":"Widget C","sku":"SKU-C"}]}}`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
			_, _ = w.Write([]byte(`{"data":{"next":null,"data":[]}}`))
		}
	}))
	defer srv.Close()

	c := flexport.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "test_token_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "products", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer test_token_123" {
		t.Fatalf("Authorization = %q, want Bearer test_token_123", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["name"] == nil {
			t.Fatalf("record missing id/name: %+v", rec)
		}
	}
	if got[0]["sku"] != "SKU-A" {
		t.Fatalf("record mapping wrong: sku = %v, want SKU-A", got[0]["sku"])
	}
}

// TestFixtureModeNoNetwork confirms the credential-free fixture path emits
// deterministic records without any HTTP call.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := flexport.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	for _, stream := range []string{"companies", "locations", "products", "invoices", "shipments"} {
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

// TestCheckFixtureMode short-circuits without network access.
func TestCheckFixtureMode(t *testing.T) {
	c := flexport.New()
	if err := c.Check(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}); err != nil {
		t.Fatalf("Check(fixture): %v", err)
	}
}

// TestBaseURLValidation rejects non-http(s) base URLs (SSRF guard).
func TestBaseURLValidation(t *testing.T) {
	c := flexport.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "file:///etc/passwd"},
		Secrets: map[string]string{"api_key": "x"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "products", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "base_url") {
		t.Fatalf("Read with bad base_url err = %v, want base_url validation error", err)
	}
}

// TestCatalogStreams confirms the published catalog exposes the core streams.
func TestCatalogStreams(t *testing.T) {
	c := flexport.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"companies": false, "locations": false, "products": false, "invoices": false, "shipments": false}
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

// TestRegisteredReadOnly confirms self-registration and read-only capabilities.
func TestRegisteredReadOnly(t *testing.T) {
	_ = flexport.New() // ensure init ran
	c := flexport.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog {
		t.Fatalf("capabilities = %+v, want Read && Catalog", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("flexport"); !ok {
		t.Fatal("registry did not resolve flexport (self-registration)")
	}
}
