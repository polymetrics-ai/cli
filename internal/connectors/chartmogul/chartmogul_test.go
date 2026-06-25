package chartmogul_test

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics/internal/connectors"
	"polymetrics/internal/connectors/chartmogul"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the ChartMogul
// connector: HTTP Basic auth (api_key as username, empty password), ChartMogul
// cursor/has_more pagination over entries[], and record mapping. Red until
// internal/connectors/chartmogul exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/customers" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("cursor") {
		case "":
			_, _ = w.Write([]byte(`{"entries":[{"uuid":"cus_1","external_id":"e1","status":"Active"},{"uuid":"cus_2","external_id":"e2","status":"Active"}],"has_more":true,"cursor":"PAGE2"}`))
		case "PAGE2":
			_, _ = w.Write([]byte(`{"entries":[{"uuid":"cus_3","external_id":"e3","status":"Lead"}],"has_more":false,"cursor":""}`))
		default:
			t.Errorf("unexpected cursor=%q", r.URL.Query().Get("cursor"))
			_, _ = w.Write([]byte(`{"entries":[],"has_more":false}`))
		}
	}))
	defer srv.Close()

	c := chartmogul.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "cm_test_key"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "customers", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("cm_test_key:"))
	if sawAuth != wantAuth {
		t.Fatalf("Authorization = %q, want %q", sawAuth, wantAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["uuid"] == nil {
			t.Fatalf("record missing uuid: %+v", rec)
		}
	}
	if got[0]["uuid"] != "cus_1" || got[2]["uuid"] != "cus_3" {
		t.Fatalf("unexpected record order: %v / %v", got[0]["uuid"], got[2]["uuid"])
	}
}

// TestReadActivitiesCursor confirms the activities stream also paginates over
// entries[] with cursor/has_more and maps its fields.
func TestReadActivitiesCursor(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/activities" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("cursor") {
		case "":
			_, _ = w.Write([]byte(`{"entries":[{"uuid":"act_1","type":"new_biz","date":"2020-01-01T00:00:00Z"}],"has_more":true,"cursor":"NEXT"}`))
		case "NEXT":
			_, _ = w.Write([]byte(`{"entries":[{"uuid":"act_2","type":"churn","date":"2020-02-01T00:00:00Z"}],"has_more":false}`))
		default:
			t.Errorf("unexpected cursor=%q", r.URL.Query().Get("cursor"))
			_, _ = w.Write([]byte(`{"entries":[],"has_more":false}`))
		}
	}))
	defer srv.Close()

	c := chartmogul.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "start_date": "2020-01-01T00:00:00Z"},
		Secrets: map[string]string{"api_key": "cm_test_key"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "activities", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read activities: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("activities = %d, want 2", len(got))
	}
	if got[0]["type"] != "new_biz" || got[1]["type"] != "churn" {
		t.Fatalf("unexpected activity mapping: %+v", got)
	}
}

// TestReadMetricsSinglePage confirms the customer_count metrics stream reads the
// single entries[] page without cursor pagination.
func TestReadMetricsSinglePage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/metrics/customer-count" {
			http.NotFound(w, r)
			return
		}
		if r.URL.Query().Get("start-date") == "" {
			t.Errorf("metrics request missing start-date")
		}
		_, _ = w.Write([]byte(`{"entries":[{"date":"2020-01-01","customers":10},{"date":"2020-02-01","customers":12}],"summary":{}}`))
	}))
	defer srv.Close()

	c := chartmogul.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "start_date": "2020-01-01T00:00:00Z"},
		Secrets: map[string]string{"api_key": "cm_test_key"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "customer_count", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read customer_count: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("customer_count entries = %d, want 2", len(got))
	}
	if got[0]["date"] != "2020-01-01" {
		t.Fatalf("unexpected metrics mapping: %+v", got)
	}
}

// TestFixtureModeNoNetwork ensures fixture mode emits deterministic records with
// no network access (credential-free conformance).
func TestFixtureModeNoNetwork(t *testing.T) {
	c := chartmogul.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"customers", "activities", "customer_count", "account"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read %s: %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture Read %s emitted no records", stream)
		}
	}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

// TestCatalogAndMetadata verifies the published streams and read-only caps.
func TestCatalogAndMetadata(t *testing.T) {
	c := chartmogul.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("chartmogul is read-only; Write should be false")
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("catalog streams = %d, want >= 3", len(cat.Streams))
	}
}

// TestRegistryResolves confirms self-registration via the connectors registry.
func TestRegistryResolves(t *testing.T) {
	_ = chartmogul.New() // ensure init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("chartmogul"); !ok {
		t.Fatal("registry did not resolve chartmogul (self-registration)")
	}
}

// TestBaseURLSSRFValidation rejects non-http(s) base_url overrides.
func TestBaseURLSSRFValidation(t *testing.T) {
	c := chartmogul.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "file:///etc/passwd"},
		Secrets: map[string]string{"api_key": "cm_test_key"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "customers", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read with file:// base_url should be rejected")
	}
}
