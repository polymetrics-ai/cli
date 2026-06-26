package datadog_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/datadog"
)

// TestReadMonitorsPaginatesAndAuthenticates is the red-first test: it asserts the
// two Datadog auth headers (DD-API-KEY + DD-APPLICATION-KEY), page/page_size
// pagination across two pages of the top-level JSON array returned by
// GET /api/v1/monitor, and record mapping.
func TestReadMonitorsPaginatesAndAuthenticates(t *testing.T) {
	var sawAPIKey, sawAppKey string
	var pagesSeen []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAPIKey = r.Header.Get("DD-API-KEY")
		sawAppKey = r.Header.Get("DD-APPLICATION-KEY")
		if r.URL.Path != "/api/v1/monitor" {
			http.NotFound(w, r)
			return
		}
		page := r.URL.Query().Get("page")
		pagesSeen = append(pagesSeen, page)
		switch page {
		case "0":
			// A full page (page_size 2) signals there may be more.
			_, _ = w.Write([]byte(`[{"id":101,"name":"CPU high","type":"metric alert","overall_state":"OK"},{"id":102,"name":"Disk full","type":"metric alert","overall_state":"Alert"}]`))
		case "1":
			// A short page (fewer than page_size) signals the end.
			_, _ = w.Write([]byte(`[{"id":103,"name":"Latency","type":"metric alert","overall_state":"OK"}]`))
		default:
			t.Errorf("unexpected page=%q", page)
			_, _ = w.Write([]byte(`[]`))
		}
	}))
	defer srv.Close()

	c := datadog.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"api_key": "dd_api_123", "application_key": "dd_app_456"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "monitors", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAPIKey != "dd_api_123" {
		t.Fatalf("DD-API-KEY = %q, want dd_api_123", sawAPIKey)
	}
	if sawAppKey != "dd_app_456" {
		t.Fatalf("DD-APPLICATION-KEY = %q, want dd_app_456", sawAppKey)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 across 2 pages; pages seen: %v", len(got), pagesSeen)
	}
	if len(pagesSeen) != 2 {
		t.Fatalf("pages requested = %v, want exactly [0 1]", pagesSeen)
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["name"] == nil {
			t.Fatalf("monitor record missing id/name: %+v", rec)
		}
	}
}

// TestReadUsersPaginatesDataEnvelope exercises the v2 {data:[...]} envelope with
// page[number]/page[size] pagination so both pagination shapes are covered.
func TestReadUsersPaginatesDataEnvelope(t *testing.T) {
	var pages []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/users" {
			http.NotFound(w, r)
			return
		}
		page := r.URL.Query().Get("page[number]")
		pages = append(pages, page)
		switch page {
		case "0":
			_, _ = w.Write([]byte(`{"data":[{"id":"u1","type":"users","attributes":{"name":"Ada","email":"ada@example.com"}},{"id":"u2","type":"users","attributes":{"name":"Grace","email":"grace@example.com"}}]}`))
		case "1":
			_, _ = w.Write([]byte(`{"data":[{"id":"u3","type":"users","attributes":{"name":"Kay","email":"kay@example.com"}}]}`))
		default:
			t.Errorf("unexpected page[number]=%q", page)
			_, _ = w.Write([]byte(`{"data":[]}`))
		}
	}))
	defer srv.Close()

	c := datadog.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"api_key": "k", "application_key": "a"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "users", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read users: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("user records = %d, want 3; pages: %v", len(got), pages)
	}
	if got[0]["name"] != "Ada" || got[0]["email"] != "ada@example.com" {
		t.Fatalf("user attributes not flattened: %+v", got[0])
	}
}

// TestFixtureModeNoNetwork confirms credential-free conformance: fixture mode
// emits deterministic records with no HTTP calls.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := datadog.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"monitors", "dashboards", "users", "slo", "downtimes"} {
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
	}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

func TestCheckRequiresBothSecrets(t *testing.T) {
	c := datadog.New()
	// Missing application_key should fail (non-fixture).
	err := c.Check(context.Background(), connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "https://api.datadoghq.com"},
		Secrets: map[string]string{"api_key": "only_api"},
	})
	if err == nil {
		t.Fatal("Check should fail when application_key is missing")
	}
}

func TestCatalogStreams(t *testing.T) {
	c := datadog.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"monitors": false, "dashboards": false, "users": false, "slo": false, "downtimes": false}
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

func TestRegisteredReadOnly(t *testing.T) {
	_ = datadog.New() // ensure init ran
	c := datadog.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("datadog is read-only; Write capability should be false")
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("datadog"); !ok {
		t.Fatal("registry did not resolve datadog (self-registration)")
	}
}
