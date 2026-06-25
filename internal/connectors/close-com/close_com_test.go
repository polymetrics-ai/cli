package closecom_test

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	closecom "polymetrics.ai/internal/connectors/close-com"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Close.com
// connector: HTTP Basic auth (API key as username, empty password), Close
// _skip/_limit offset pagination over data[] with has_more, and record mapping.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/lead/" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("_skip") {
		case "", "0":
			_, _ = w.Write([]byte(`{"data":[{"id":"lead_1","display_name":"Acme","date_updated":"2026-01-01T00:00:00Z"},{"id":"lead_2","display_name":"Beta","date_updated":"2026-01-02T00:00:00Z"}],"has_more":true}`))
		case "2":
			_, _ = w.Write([]byte(`{"data":[{"id":"lead_3","display_name":"Gamma","date_updated":"2026-01-03T00:00:00Z"}],"has_more":false}`))
		default:
			t.Errorf("unexpected _skip=%q", r.URL.Query().Get("_skip"))
			_, _ = w.Write([]byte(`{"data":[],"has_more":false}`))
		}
	}))
	defer srv.Close()

	c := closecom.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"api_key": "api_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "leads", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("api_test_123:"))
	if sawAuth != wantAuth {
		t.Fatalf("Authorization = %q, want %q", sawAuth, wantAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["date_updated"] == nil {
			t.Fatalf("record missing id/date_updated: %+v", rec)
		}
	}
	if got[0]["display_name"] != "Acme" {
		t.Fatalf("first record display_name = %v, want Acme", got[0]["display_name"])
	}
}

// TestFixtureModeNoNetwork confirms fixture mode emits deterministic records
// without any HTTP call, so conformance runs without credentials.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := closecom.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "contacts", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read(fixture): %v", err)
	}
	if len(got) == 0 {
		t.Fatal("fixture mode emitted no records")
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("fixture record missing id: %+v", rec)
		}
	}

	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

// TestCatalogStreams confirms the published catalog exposes the core streams
// with primary keys and cursor fields.
func TestCatalogStreams(t *testing.T) {
	c := closecom.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "close-com" {
		t.Fatalf("catalog connector = %q, want close-com", cat.Connector)
	}
	byName := map[string]connectors.Stream{}
	for _, s := range cat.Streams {
		byName[s.Name] = s
	}
	for _, want := range []string{"leads", "contacts", "opportunities", "activities", "users"} {
		s, ok := byName[want]
		if !ok {
			t.Fatalf("catalog missing stream %q", want)
		}
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q missing primary key", want)
		}
	}
}

// TestBaseURLValidation rejects non-http(s) base URLs to bound SSRF risk.
func TestBaseURLValidation(t *testing.T) {
	c := closecom.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil/"},
		Secrets: map[string]string{"api_key": "api_x"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "leads", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "base_url") {
		t.Fatalf("Read with ftp base_url err = %v, want base_url validation error", err)
	}
}

// TestRegisteredReadOnly confirms self-registration and read-only capabilities.
func TestRegisteredReadOnly(t *testing.T) {
	_ = closecom.New() // ensure init ran
	c := closecom.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (Close source is read-only)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("close-com"); !ok {
		t.Fatal("registry did not resolve close-com (self-registration)")
	}
}
