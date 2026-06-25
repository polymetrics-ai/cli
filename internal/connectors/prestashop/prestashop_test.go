package prestashop_test

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics/internal/connectors"
	"polymetrics/internal/connectors/prestashop"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the PrestaShop
// connector: HTTP Basic auth (access_key as username, empty password), JSON
// output via output_format=JSON, limit=offset,count pagination over two pages,
// and record mapping out of the resource-keyed envelope ({"customers":[...]}).
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var sawOutputFormat string
	var sawDisplay string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		q := r.URL.Query()
		sawOutputFormat = q.Get("output_format")
		sawDisplay = q.Get("display")
		if r.URL.Path != "/api/customers" {
			http.NotFound(w, r)
			return
		}
		// PrestaShop pagination: limit=<offset>,<count>
		switch q.Get("limit") {
		case "0,2":
			_, _ = w.Write([]byte(`{"customers":[{"id":1,"email":"a@example.com","date_upd":"2024-01-01 10:00:00"},{"id":2,"email":"b@example.com","date_upd":"2024-01-02 10:00:00"}]}`))
		case "2,2":
			_, _ = w.Write([]byte(`{"customers":[{"id":3,"email":"c@example.com","date_upd":"2024-01-03 10:00:00"}]}`))
		default:
			t.Errorf("unexpected limit=%q", q.Get("limit"))
			_, _ = w.Write([]byte(`{"customers":[]}`))
		}
	}))
	defer srv.Close()

	c := prestashop.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"access_key": "KEY123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "customers", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("KEY123:"))
	if sawAuth != wantAuth {
		t.Fatalf("Authorization = %q, want %q", sawAuth, wantAuth)
	}
	if !strings.EqualFold(sawOutputFormat, "JSON") {
		t.Fatalf("output_format = %q, want JSON", sawOutputFormat)
	}
	if sawDisplay != "full" {
		t.Fatalf("display = %q, want full", sawDisplay)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["date_upd"] == nil {
			t.Fatalf("record missing id/date_upd: %+v", rec)
		}
	}
}

// TestReadFixtureModeNoNetwork ensures fixture mode emits deterministic records
// without any network access, so credential-free conformance passes.
func TestReadFixtureModeNoNetwork(t *testing.T) {
	c := prestashop.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "orders", Config: cfg}, func(rec connectors.Record) error {
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
}

// TestCheckFixtureMode confirms Check short-circuits in fixture mode.
func TestCheckFixtureMode(t *testing.T) {
	c := prestashop.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
}

// TestCatalogHasCoreStreams verifies the published catalog includes the core set.
func TestCatalogHasCoreStreams(t *testing.T) {
	c := prestashop.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"customers": false, "orders": false, "products": false}
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

// TestUnknownStreamRejected ensures an unknown stream is an error.
func TestUnknownStreamRejected(t *testing.T) {
	c := prestashop.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "does_not_exist", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("expected error for unknown stream")
	}
}

// TestStartDateFilter confirms the start_date config becomes a date_upd filter.
func TestStartDateFilter(t *testing.T) {
	var sawFilter string
	var sawDate string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		sawFilter = q.Get("filter[date_upd]")
		sawDate = q.Get("date")
		_, _ = w.Write([]byte(`{"customers":[]}`))
	}))
	defer srv.Close()

	c := prestashop.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "start_date": "2024-01-01"},
		Secrets: map[string]string{"access_key": "KEY123"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "customers", Config: cfg}, func(connectors.Record) error { return nil })
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if !strings.Contains(sawFilter, "2024-01-01") {
		t.Fatalf("filter[date_upd] = %q, want it to contain 2024-01-01", sawFilter)
	}
	if sawDate != "1" {
		t.Fatalf("date = %q, want 1 (enables date filtering)", sawDate)
	}
}

// TestRegistryResolves verifies self-registration via RegisterFactory in init().
func TestRegistryResolves(t *testing.T) {
	_ = prestashop.New() // ensure package init ran
	r := connectors.NewRegistry()
	c, ok := r.Get("prestashop")
	if !ok {
		t.Fatal("registry did not resolve prestashop (self-registration)")
	}
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
}
