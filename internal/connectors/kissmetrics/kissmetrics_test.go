package kissmetrics_test

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/kissmetrics"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Kissmetrics
// connector: HTTP Basic auth (username + secret password), OffsetIncrement
// pagination over data[] with limit/offset, and record mapping. Red until
// internal/connectors/kissmetrics exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	const wantUser = "acct@example.com"
	const wantPass = "km_secret_123"
	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(wantUser+":"+wantPass))

	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/products" {
			http.NotFound(w, r)
			return
		}
		offset := r.URL.Query().Get("offset")
		if r.URL.Query().Get("limit") == "" {
			t.Errorf("missing limit param")
		}
		switch offset {
		case "", "0":
			// Full page (50 records) signals there is a next page.
			w.Write([]byte(productPage(0, 50)))
		case "50":
			// Short page (1 record) signals the end.
			w.Write([]byte(productPage(50, 1)))
		default:
			t.Errorf("unexpected offset=%q", offset)
			w.Write([]byte(`{"data":[]}`))
		}
	}))
	defer srv.Close()

	c := kissmetrics.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "username": wantUser},
		Secrets: map[string]string{"password": wantPass},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "products", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != wantAuth {
		t.Fatalf("Authorization = %q, want %q", sawAuth, wantAuth)
	}
	if len(got) != 51 {
		t.Fatalf("records = %d, want 51 (2 pages: 50 + 1)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("record missing id: %+v", rec)
		}
	}
}

// productPage renders a data[] page of n product records starting at the given
// id offset.
func productPage(start, n int) string {
	out := `{"data":[`
	for i := 0; i < n; i++ {
		if i > 0 {
			out += ","
		}
		id := strconv.Itoa(start + i + 1)
		out += `{"id":"` + id + `","name":"Product ` + id + `"}`
	}
	out += `]}`
	return out
}

// TestNestedStreamUsesProductPartition verifies that reports/events/properties
// streams read from products/{product_id}/<resource> using the product_id config.
func TestNestedStreamUsesProductPartition(t *testing.T) {
	var sawPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawPath = r.URL.Path
		w.Write([]byte(`{"data":[{"id":"e1","name":"Signed Up"}]}`))
	}))
	defer srv.Close()

	c := kissmetrics.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "username": "u", "product_id": "prod_42"},
		Secrets: map[string]string{"password": "p"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "events", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read events: %v", err)
	}
	if sawPath != "/products/prod_42/events" {
		t.Fatalf("path = %q, want /products/prod_42/events", sawPath)
	}
	if len(got) != 1 || got[0]["id"] != "e1" {
		t.Fatalf("records = %+v, want one event e1", got)
	}
}

// TestNestedStreamRequiresProductID asserts a clear error when a nested stream
// is read without a product_id.
func TestNestedStreamRequiresProductID(t *testing.T) {
	c := kissmetrics.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"username": "u"},
		Secrets: map[string]string{"password": "p"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "events", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read events without product_id should error")
	}
}

// TestFixtureModeNoNetwork ensures fixture mode emits deterministic records
// without any network access, so conformance passes without live creds.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := kissmetrics.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"products", "reports", "events", "properties"} {
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
		for _, rec := range got {
			if rec["id"] == nil {
				t.Fatalf("fixture %s record missing id: %+v", stream, rec)
			}
		}
	}
	// Check in fixture mode must not require creds or network.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
}

// TestCatalogAndMetadata verifies read-only capabilities and the published
// stream catalog.
func TestCatalogAndMetadata(t *testing.T) {
	c := kissmetrics.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("kissmetrics is read-only; Write must be false")
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"products": true, "reports": true, "events": true, "properties": true}
	if len(cat.Streams) != len(want) {
		t.Fatalf("streams = %d, want %d", len(cat.Streams), len(want))
	}
	for _, s := range cat.Streams {
		if !want[s.Name] {
			t.Fatalf("unexpected stream %q", s.Name)
		}
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q missing primary key", s.Name)
		}
	}
}

// TestRegistryResolution confirms the connector self-registers and resolves via
// the shared registry.
func TestRegistryResolution(t *testing.T) {
	_ = kissmetrics.New() // ensure init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("kissmetrics"); !ok {
		t.Fatal("registry did not resolve kissmetrics (self-registration)")
	}
}
