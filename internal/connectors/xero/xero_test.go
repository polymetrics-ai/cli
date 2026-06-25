package xero_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/xero"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Xero
// connector: Bearer auth + Xero-tenant-id header, page-based pagination over the
// resource-keyed array ({"Invoices":[...]}), and record mapping. Red until
// internal/connectors/xero exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth, sawTenant string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		sawTenant = r.Header.Get("Xero-tenant-id")
		if r.URL.Path != "/Invoices" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page") {
		case "1":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"Id":"x","Status":"OK","Invoices":[` + invoicePage(1, 100) + `]}`))
		case "2":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"Id":"x","Status":"OK","Invoices":[{"InvoiceID":"inv_201","Type":"ACCREC","Status":"PAID","Total":42.5,"UpdatedDateUTC":"2026-01-02T00:00:00Z"}]}`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
			_, _ = w.Write([]byte(`{"Invoices":[]}`))
		}
	}))
	defer srv.Close()

	c := xero.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"access_token": "at_test_123", "tenant_id": "tenant_abc"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "invoices", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer at_test_123" {
		t.Fatalf("Authorization = %q, want Bearer at_test_123", sawAuth)
	}
	if sawTenant != "tenant_abc" {
		t.Fatalf("Xero-tenant-id = %q, want tenant_abc", sawTenant)
	}
	if len(got) != 101 {
		t.Fatalf("records = %d, want 101 (2 pages: 100 + 1)", len(got))
	}
	last := got[len(got)-1]
	if last["id"] != "inv_201" {
		t.Fatalf("last record id = %v, want inv_201", last["id"])
	}
	if last["status"] != "PAID" || last["type"] != "ACCREC" {
		t.Fatalf("record mapping wrong: %+v", last)
	}
}

// invoicePage builds a JSON array fragment of n invoices with the given page tag.
func invoicePage(page, n int) string {
	out := ""
	for i := 0; i < n; i++ {
		if i > 0 {
			out += ","
		}
		out += `{"InvoiceID":"inv_` + itoa(page) + "_" + itoa(i) + `","Type":"ACCREC","Status":"DRAFT","Total":10,"UpdatedDateUTC":"2026-01-01T00:00:00Z"}`
	}
	return out
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	digits := ""
	for i > 0 {
		digits = string(rune('0'+i%10)) + digits
		i /= 10
	}
	return digits
}

func TestFixtureModeReadsWithoutNetwork(t *testing.T) {
	c := xero.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "contacts", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("fixture records = %d, want 2", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("fixture record missing id: %+v", rec)
		}
	}
}

func TestCheckFixtureModeNoNetwork(t *testing.T) {
	c := xero.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

func TestCatalogStreams(t *testing.T) {
	c := xero.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("streams = %d, want >= 3", len(cat.Streams))
	}
	wantPK := map[string]string{
		"invoices": "InvoiceID",
		"contacts": "ContactID",
		"accounts": "AccountID",
	}
	seen := map[string]bool{}
	for _, s := range cat.Streams {
		seen[s.Name] = true
		if pk, ok := wantPK[s.Name]; ok {
			if len(s.PrimaryKey) != 1 || s.PrimaryKey[0] != pk {
				t.Fatalf("stream %q primary key = %v, want [%s]", s.Name, s.PrimaryKey, pk)
			}
		}
	}
	for name := range wantPK {
		if !seen[name] {
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

func TestRegisteredReadOnly(t *testing.T) {
	_ = xero.New() // ensure init ran
	c := xero.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("xero is read-only; Write capability should be false")
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("xero"); !ok {
		t.Fatal("registry did not resolve xero (self-registration)")
	}
}

func TestReadRejectsMissingTenant(t *testing.T) {
	c := xero.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "https://api.xero.com/api.xro/2.0"},
		Secrets: map[string]string{"access_token": "at_test_123"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "invoices", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read without tenant_id should fail")
	}
}
