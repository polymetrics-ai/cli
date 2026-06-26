package invoiceninja_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/invoiceninja"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Invoice Ninja
// connector: it asserts the X-API-TOKEN header auth, PageIncrement pagination
// over page/per_page until a short page, records extracted from data[], and the
// record mapper surfacing id. Red until internal/connectors/invoiceninja exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawToken string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawToken = r.Header.Get("X-API-TOKEN")
		if r.URL.Path != "/clients" {
			http.NotFound(w, r)
			return
		}
		// per_page must be passed through so a short final page is detectable.
		perPage := r.URL.Query().Get("per_page")
		if perPage != "2" {
			t.Errorf("per_page = %q, want 2", perPage)
		}
		switch r.URL.Query().Get("page") {
		case "", "1":
			_, _ = w.Write([]byte(`{"data":[{"id":"c1","name":"Acme","balance":10},{"id":"c2","name":"Globex","balance":20}],"meta":{"pagination":{"total":3}}}`))
		case "2":
			_, _ = w.Write([]byte(`{"data":[{"id":"c3","name":"Initech","balance":30}],"meta":{"pagination":{"total":3}}}`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
			_, _ = w.Write([]byte(`{"data":[]}`))
		}
	}))
	defer srv.Close()

	c := invoiceninja.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"api_key": "tok_abc"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "clients", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawToken != "tok_abc" {
		t.Fatalf("X-API-TOKEN = %q, want tok_abc", sawToken)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for i, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("record %d missing id: %+v", i, rec)
		}
		if rec["name"] == nil {
			t.Fatalf("record %d missing mapped name: %+v", i, rec)
		}
	}
}

// TestReadInvoicesMapsFields confirms the invoices stream routes to the invoices
// path and the mapper surfaces invoice-specific fields.
func TestReadInvoicesMapsFields(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/invoices" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"data":[{"id":"inv1","number":"0001","status_id":"2","amount":100,"balance":50,"client_id":"c1"}]}`))
	}))
	defer srv.Close()

	c := invoiceninja.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "tok_abc"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "invoices", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read invoices: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("records = %d, want 1", len(got))
	}
	rec := got[0]
	if rec["id"] != "inv1" || rec["number"] != "0001" || rec["client_id"] != "c1" {
		t.Fatalf("invoice not mapped: %+v", rec)
	}
}

// TestFixtureModeNoNetwork ensures the fixture path emits deterministic records
// with no live credentials and no network access.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := invoiceninja.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "clients", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) == 0 {
		t.Fatal("fixture mode emitted no records")
	}
	for i, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("fixture record %d missing id: %+v", i, rec)
		}
	}
	// Check must also short-circuit in fixture mode.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

func TestCatalogStreams(t *testing.T) {
	c := invoiceninja.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "invoiceninja" {
		t.Fatalf("connector = %q, want invoiceninja", cat.Connector)
	}
	want := map[string]bool{"clients": true, "invoices": true, "products": true, "payments": true, "quotes": true}
	seen := map[string]bool{}
	for _, s := range cat.Streams {
		seen[s.Name] = true
		if len(s.PrimaryKey) == 0 || s.PrimaryKey[0] != "id" {
			t.Fatalf("stream %q primary key = %v, want [id]", s.Name, s.PrimaryKey)
		}
		if len(s.Fields) == 0 {
			t.Fatalf("stream %q has no fields", s.Name)
		}
	}
	for name := range want {
		if !seen[name] {
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

func TestRegisteredReadOnly(t *testing.T) {
	_ = invoiceninja.New() // ensure init ran
	c := invoiceninja.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("invoiceninja"); !ok {
		t.Fatal("registry did not resolve invoiceninja (self-registration)")
	}
}

func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := invoiceninja.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example.com"},
		Secrets: map[string]string{"api_key": "tok_abc"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "clients", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("expected error for non-http(s) base_url scheme")
	}
}
