package bigmailer_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics/internal/connectors"
	"polymetrics/internal/connectors/bigmailer"
)

// TestReadBrandsPaginatesAndAuthenticates is the red-first test: it asserts the
// X-API-Key auth header, BigMailer cursor pagination over two pages of the
// top-level brands stream (data[] + has_more/cursor), and record mapping.
func TestReadBrandsPaginatesAndAuthenticates(t *testing.T) {
	var sawKey string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawKey = r.Header.Get("X-API-Key")
		if r.URL.Path != "/brands" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("cursor") {
		case "":
			_, _ = w.Write([]byte(`{"data":[{"id":"b_1","name":"Brand One","from_email":"a@x.io","num_contacts":2,"created":1700000000},{"id":"b_2","name":"Brand Two","from_email":"b@x.io","num_contacts":5,"created":1700000100}],"has_more":true,"cursor":"CUR2"}`))
		case "CUR2":
			_, _ = w.Write([]byte(`{"data":[{"id":"b_3","name":"Brand Three","from_email":"c@x.io","num_contacts":9,"created":1700000200}],"has_more":false,"cursor":""}`))
		default:
			t.Errorf("unexpected cursor=%q", r.URL.Query().Get("cursor"))
			_, _ = w.Write([]byte(`{"data":[],"has_more":false}`))
		}
	}))
	defer srv.Close()

	c := bigmailer.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "key_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "brands", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawKey != "key_test_123" {
		t.Fatalf("X-API-Key = %q, want key_test_123", sawKey)
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

// TestReadContactsSubstreamIteratesBrands asserts that the brand-substream
// contacts stream first lists brands, then paginates each brand's contacts
// endpoint, stamping brand_id onto every record.
func TestReadContactsSubstreamIteratesBrands(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-Key") == "" {
			t.Errorf("missing X-API-Key on %s", r.URL.Path)
		}
		switch r.URL.Path {
		case "/brands":
			_, _ = w.Write([]byte(`{"data":[{"id":"b_1","name":"Brand One"}],"has_more":false,"cursor":""}`))
		case "/brands/b_1/contacts":
			switch r.URL.Query().Get("cursor") {
			case "":
				_, _ = w.Write([]byte(`{"data":[{"id":"c_1","email":"one@x.io","created":1700000000}],"has_more":true,"cursor":"NEXT"}`))
			case "NEXT":
				_, _ = w.Write([]byte(`{"data":[{"id":"c_2","email":"two@x.io","created":1700000100}],"has_more":false,"cursor":""}`))
			default:
				t.Errorf("unexpected contacts cursor=%q", r.URL.Query().Get("cursor"))
				_, _ = w.Write([]byte(`{"data":[],"has_more":false}`))
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := bigmailer.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "key_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "contacts", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2 (2 pages across one brand)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["brand_id"] != "b_1" {
			t.Fatalf("contact record missing id or brand_id: %+v", rec)
		}
	}
}

// TestFixtureModeNoNetwork ensures fixture mode emits deterministic records with
// no network access, so conformance runs without live credentials.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := bigmailer.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"brands", "users", "contacts", "lists", "fields"} {
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
	// Check must short-circuit in fixture mode (no creds, no network).
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

func TestCatalogAndMetadata(t *testing.T) {
	c := bigmailer.New()
	if c.Name() != "bigmailer" {
		t.Fatalf("Name = %q, want bigmailer", c.Name())
	}
	caps := c.Metadata().Capabilities
	if !caps.Check || !caps.Catalog || !caps.Read {
		t.Fatalf("capabilities = %+v, want Check && Catalog && Read", caps)
	}
	if caps.Write {
		t.Fatalf("bigmailer is read-only; Write should be false")
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("catalog streams = %d, want >= 3", len(cat.Streams))
	}
}

func TestRegistryResolves(t *testing.T) {
	_ = bigmailer.New() // ensure init() ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("bigmailer"); !ok {
		t.Fatal("registry did not resolve bigmailer (self-registration)")
	}
}
