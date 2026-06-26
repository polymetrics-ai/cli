package lessannoyingcrm_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	lessannoyingcrm "polymetrics.ai/internal/connectors/less-annoying-crm"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Less Annoying
// CRM connector. It asserts:
//   - the API key is sent as a raw Authorization header (no Bearer prefix),
//   - the request is a POST whose JSON body carries Function: GetContacts,
//   - PageIncrement pagination over the Results[] array advances Page until a
//     short page is returned,
//   - records are mapped (ContactId carried through).
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var sawMethod string
	var sawFunction string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		sawMethod = r.Method

		body, _ := io.ReadAll(r.Body)
		var payload map[string]any
		_ = json.Unmarshal(body, &payload)
		if fn, ok := payload["Function"].(string); ok {
			sawFunction = fn
		}

		page := r.URL.Query().Get("Page")
		if page == "" {
			if p, ok := payload["Page"]; ok {
				switch v := p.(type) {
				case string:
					page = v
				case float64:
					if v == 1 {
						page = "1"
					} else if v == 2 {
						page = "2"
					}
				}
			}
		}
		w.Header().Set("Content-Type", "application/json")
		switch page {
		case "", "1":
			// Full page (page_size handling: emit 2 records, server claims a
			// next page by returning a non-short page only when the connector
			// requested page 1).
			_, _ = w.Write([]byte(`{"Success":true,"Results":[{"ContactId":"c1","Name":"Ada"},{"ContactId":"c2","Name":"Grace"}]}`))
		case "2":
			_, _ = w.Write([]byte(`{"Success":true,"Results":[{"ContactId":"c3","Name":"Katherine"}]}`))
		default:
			_, _ = w.Write([]byte(`{"Success":true,"Results":[]}`))
		}
	}))
	defer srv.Close()

	c := lessannoyingcrm.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
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
	if sawAuth != "key_test_123" {
		t.Fatalf("Authorization = %q, want raw key (no Bearer)", sawAuth)
	}
	if sawMethod != http.MethodPost {
		t.Fatalf("method = %q, want POST", sawMethod)
	}
	if sawFunction != "GetContacts" {
		t.Fatalf("Function = %q, want GetContacts", sawFunction)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	if got[0]["ContactId"] == nil {
		t.Fatalf("record missing ContactId: %+v", got[0])
	}
}

// TestFixtureModeNoNetwork asserts fixture mode emits deterministic records
// without any network access.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := lessannoyingcrm.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "contacts", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) == 0 {
		t.Fatal("fixture mode emitted no records")
	}
	for _, rec := range got {
		if rec["ContactId"] == nil {
			t.Fatalf("fixture record missing ContactId: %+v", rec)
		}
	}
}

// TestCheckFixtureMode asserts Check short-circuits in fixture mode.
func TestCheckFixtureMode(t *testing.T) {
	c := lessannoyingcrm.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

// TestCatalogStreams asserts the catalog exposes the core streams with primary
// keys.
func TestCatalogStreams(t *testing.T) {
	c := lessannoyingcrm.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"users": false, "contacts": false, "tasks": false, "notes": false, "events": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
	}
	for name, found := range want {
		if !found {
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

// TestRegistryResolves asserts self-registration via the global registry.
func TestRegistryResolves(t *testing.T) {
	_ = lessannoyingcrm.New() // ensure init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("less-annoying-crm"); !ok {
		t.Fatal("registry did not resolve less-annoying-crm (self-registration)")
	}
}

// TestReadOnlyCapabilities asserts the connector advertises read but not write
// (Less Annoying CRM port is read-only).
func TestReadOnlyCapabilities(t *testing.T) {
	c := lessannoyingcrm.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read/Catalog/Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false", caps)
	}
}
