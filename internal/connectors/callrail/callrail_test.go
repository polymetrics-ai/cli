package callrail_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/callrail"
)

// TestReadPaginatesAndAuthenticates is the red-first test: it asserts the
// CallRail Token auth header, page-number pagination across two pages keyed on
// total_pages, the account-scoped path, and record mapping.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/a/ACC123/calls.json" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page") {
		case "", "1":
			_, _ = w.Write([]byte(`{"page":1,"per_page":1,"total_pages":2,"total_records":2,"calls":[{"id":"CAL_1","start_time":"2026-01-01T00:00:00Z","duration":42}]}`))
		case "2":
			_, _ = w.Write([]byte(`{"page":2,"per_page":1,"total_pages":2,"total_records":2,"calls":[{"id":"CAL_2","start_time":"2026-01-02T00:00:00Z","duration":7}]}`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
			_, _ = w.Write([]byte(`{"page":3,"total_pages":2,"calls":[]}`))
		}
	}))
	defer srv.Close()

	c := callrail.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "1"},
		Secrets: map[string]string{"api_key": "key_abc", "account_id": "ACC123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "calls", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if want := `Token token="key_abc"`; sawAuth != want {
		t.Fatalf("Authorization = %q, want %q", sawAuth, want)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2 (2 pages)", len(got))
	}
	if got[0]["id"] != "CAL_1" || got[1]["id"] != "CAL_2" {
		t.Fatalf("ids = %v / %v, want CAL_1 / CAL_2", got[0]["id"], got[1]["id"])
	}
	if got[0]["start_time"] == nil || got[0]["duration"] == nil {
		t.Fatalf("record missing mapped fields: %+v", got[0])
	}
}

// TestAccountIDFromConfig confirms account_id can also come from Config (not
// only Secrets) and that the path is account-scoped.
func TestAccountIDFromConfig(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/a/FROMCFG/companies.json" {
			t.Errorf("path = %q, want /a/FROMCFG/companies.json", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"page":1,"total_pages":1,"companies":[{"id":"CMP_1","name":"Acme","created_at":"2026-01-01T00:00:00Z"}]}`))
	}))
	defer srv.Close()

	c := callrail.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "account_id": "FROMCFG"},
		Secrets: map[string]string{"api_key": "key_abc"},
	}
	var n int
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "companies", Config: cfg}, func(connectors.Record) error {
		n++
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if n != 1 {
		t.Fatalf("records = %d, want 1", n)
	}
}

// TestFixtureMode exercises the credential-free fixture path the conformance
// harness relies on: deterministic records, no network.
func TestFixtureMode(t *testing.T) {
	c := callrail.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"calls", "companies", "users", "text_messages"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if len(got) != 2 {
			t.Fatalf("fixture %s records = %d, want 2", stream, len(got))
		}
		if got[0]["id"] == nil {
			t.Fatalf("fixture %s record missing id: %+v", stream, got[0])
		}
	}
	// Check and Catalog must also work credential-free in fixture mode.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	cat, err := c.Catalog(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("catalog streams = %d, want >= 3", len(cat.Streams))
	}
}

func TestMetadataReadOnly(t *testing.T) {
	caps := callrail.New().Metadata().Capabilities
	if !caps.Check || !caps.Catalog || !caps.Read {
		t.Fatalf("capabilities = %+v, want Check && Catalog && Read", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only)", caps)
	}
}

func TestRegistryResolution(t *testing.T) {
	_ = callrail.New() // ensure init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("callrail"); !ok {
		t.Fatal("registry did not resolve callrail (self-registration)")
	}
}
