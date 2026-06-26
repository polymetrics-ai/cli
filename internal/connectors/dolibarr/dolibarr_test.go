package dolibarr_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/dolibarr"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Dolibarr
// connector: it asserts the DOLAPIKEY auth header is sent, that page-based
// pagination walks across two pages (Dolibarr returns a top-level JSON array and
// signals end-of-data with a short final page), and that records are mapped with
// their primary key field present. Red until internal/connectors/dolibarr exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAPIKey string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAPIKey = r.Header.Get("DOLAPIKEY")
		if r.URL.Path != "/thirdparties" {
			http.NotFound(w, r)
			return
		}
		if r.URL.Query().Get("limit") != "2" {
			t.Errorf("limit = %q, want 2", r.URL.Query().Get("limit"))
		}
		page := r.URL.Query().Get("page")
		switch page {
		case "0":
			// Full page (2 == limit) -> the loop must request the next page.
			_, _ = w.Write([]byte(`[{"id":"1","name":"Acme"},{"id":"2","name":"Globex"}]`))
		case "1":
			// Short page (< limit) -> end of data, no further request.
			_, _ = w.Write([]byte(`[{"id":"3","name":"Initech"}]`))
		default:
			t.Errorf("unexpected page=%q", page)
			_, _ = w.Write([]byte(`[]`))
		}
	}))
	defer srv.Close()

	c := dolibarr.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"api_key": "doli_test_abc"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "thirdparties", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAPIKey != "doli_test_abc" {
		t.Fatalf("DOLAPIKEY = %q, want doli_test_abc", sawAPIKey)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("record missing id: %+v", rec)
		}
	}
}

// TestReadStopsOnEmptyTerminalPage confirms that when the final full page is
// followed by an empty page (Dolibarr's other end-of-data shape), the loop
// terminates without emitting phantom records.
func TestReadStopsOnEmptyTerminalPage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/products" {
			http.NotFound(w, r)
			return
		}
		page := r.URL.Query().Get("page")
		switch page {
		case "0":
			_, _ = w.Write([]byte(`[{"id":"10","ref":"P10"},{"id":"11","ref":"P11"}]`))
		case "1":
			_, _ = w.Write([]byte(`[]`))
		default:
			t.Errorf("unexpected page=%q", page)
			_, _ = w.Write([]byte(`[]`))
		}
	}))
	defer srv.Close()

	c := dolibarr.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"api_key": "doli_test_abc"},
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "products", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2", len(got))
	}
}

// TestFixtureModeReadsWithoutNetwork verifies the credential-free fixture path so
// the conformance harness can exercise the connector without live creds.
func TestFixtureModeReadsWithoutNetwork(t *testing.T) {
	c := dolibarr.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"thirdparties", "invoices", "products", "orders", "contacts"} {
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
	// Check must short-circuit in fixture mode with no creds.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

// TestCheckUsesAPIKeyHeader confirms Check performs a bounded authenticated read.
func TestCheckUsesAPIKeyHeader(t *testing.T) {
	var sawAPIKey string
	var sawLimit string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAPIKey = r.Header.Get("DOLAPIKEY")
		sawLimit = r.URL.Query().Get("limit")
		_, _ = w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	c := dolibarr.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "doli_test_abc"},
	}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check: %v", err)
	}
	if sawAPIKey != "doli_test_abc" {
		t.Fatalf("DOLAPIKEY = %q, want doli_test_abc", sawAPIKey)
	}
	if _, err := strconv.Atoi(sawLimit); err != nil {
		t.Fatalf("Check should bound the probe with a numeric limit, got %q", sawLimit)
	}
}

// TestCatalogStreams asserts the published catalog covers the core streams with
// primary keys.
func TestCatalogStreams(t *testing.T) {
	c := dolibarr.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"thirdparties": false, "invoices": false, "products": false, "orders": false, "contacts": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
			if len(s.PrimaryKey) == 0 {
				t.Fatalf("stream %s missing primary key", s.Name)
			}
		}
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("catalog missing core stream %q", name)
		}
	}
}

// TestRegistryResolution asserts self-registration via init() and that the
// connector is read-only (no Write capability).
func TestRegistryResolution(t *testing.T) {
	_ = dolibarr.New() // ensure package init ran
	r := connectors.NewRegistry()
	got, ok := r.Get("dolibarr")
	if !ok {
		t.Fatal("registry did not resolve dolibarr (self-registration)")
	}
	caps := got.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("dolibarr should be read-only, got Write=true")
	}
}
