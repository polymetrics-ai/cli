package klaviyo_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/klaviyo"
)

// TestReadPaginatesAndAuthenticates is the red-first test: it asserts the
// Klaviyo-API-Key auth header, the required revision header, cursor pagination
// across two pages via the JSON:API links.next URL, and id+attributes flattening
// for record mapping.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth, sawRevision string
	var srvURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		sawRevision = r.Header.Get("revision")
		if r.URL.Path != "/api/profiles" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Query().Get("page[cursor]") {
		case "":
			// First page: a links.next pointing back at this server with a cursor.
			next := srvURL + "/api/profiles?page%5Bcursor%5D=CURSOR2"
			_, _ = w.Write([]byte(`{"data":[` +
				`{"type":"profile","id":"p1","attributes":{"email":"a@example.com","created":"2024-01-01T00:00:00Z","updated":"2024-01-02T00:00:00Z","first_name":"Ada"}},` +
				`{"type":"profile","id":"p2","attributes":{"email":"b@example.com","created":"2024-01-03T00:00:00Z","updated":"2024-01-04T00:00:00Z","first_name":"Bea"}}` +
				`],"links":{"self":"` + srvURL + `/api/profiles","next":"` + next + `","prev":null}}`))
		case "CURSOR2":
			_, _ = w.Write([]byte(`{"data":[` +
				`{"type":"profile","id":"p3","attributes":{"email":"c@example.com","created":"2024-01-05T00:00:00Z","updated":"2024-01-06T00:00:00Z","first_name":"Cy"}}` +
				`],"links":{"self":"","next":null,"prev":null}}`))
		default:
			t.Errorf("unexpected page[cursor]=%q", r.URL.Query().Get("page[cursor]"))
			_, _ = w.Write([]byte(`{"data":[],"links":{"next":null}}`))
		}
	}))
	defer srv.Close()
	srvURL = srv.URL

	c := klaviyo.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL + "/api"},
		Secrets: map[string]string{"api_key": "pk_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "profiles", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Klaviyo-API-Key pk_test_123" {
		t.Fatalf("Authorization = %q, want Klaviyo-API-Key pk_test_123", sawAuth)
	}
	if sawRevision == "" {
		t.Fatalf("revision header missing, want a date version")
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	// Record mapping: id is promoted to top level, attributes flattened.
	first := got[0]
	if first["id"] != "p1" {
		t.Fatalf("record[0].id = %v, want p1", first["id"])
	}
	if first["email"] != "a@example.com" {
		t.Fatalf("record[0].email = %v, want a@example.com (flattened from attributes)", first["email"])
	}
	if first["created"] != "2024-01-01T00:00:00Z" {
		t.Fatalf("record[0].created = %v, want flattened attributes.created", first["created"])
	}
}

// TestFixtureModeNoNetwork verifies fixture mode emits deterministic records
// without any network call (required for credential-free conformance).
func TestFixtureModeNoNetwork(t *testing.T) {
	c := klaviyo.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "events", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) == 0 {
		t.Fatal("fixture Read emitted no records")
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("fixture record missing id: %+v", rec)
		}
	}
	// Check must also short-circuit in fixture mode (no creds).
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

func TestBaseURLValidationRejectsBadScheme(t *testing.T) {
	c := klaviyo.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example.com"},
		Secrets: map[string]string{"api_key": "pk_test_123"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "profiles", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "base_url") {
		t.Fatalf("expected base_url scheme validation error, got %v", err)
	}
}

func TestCatalogAndRegistry(t *testing.T) {
	c := klaviyo.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("catalog streams = %d, want >= 3", len(cat.Streams))
	}
	hasProfiles := false
	for _, s := range cat.Streams {
		if s.Name == "profiles" {
			hasProfiles = true
			if len(s.PrimaryKey) == 0 {
				t.Fatalf("profiles stream missing primary key")
			}
		}
	}
	if !hasProfiles {
		t.Fatal("catalog missing profiles stream")
	}

	r := connectors.NewRegistry()
	if _, ok := r.Get("klaviyo"); !ok {
		t.Fatal("registry did not resolve klaviyo (self-registration)")
	}
}
