package missive_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/missive"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Missive
// connector: Bearer auth via api_key, offset-based pagination (limit/offset)
// over the stream-named record array, and record mapping. Missive returns
// {"contacts":[...]} where the JSON key equals the stream name; pages advance by
// offset until a short page is returned.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var sawLimits []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/contacts" {
			http.NotFound(w, r)
			return
		}
		sawLimits = append(sawLimits, r.URL.Query().Get("limit"))
		switch r.URL.Query().Get("offset") {
		case "", "0":
			_, _ = w.Write([]byte(`{"contacts":[{"id":"c_1","first_name":"Ada"},{"id":"c_2","first_name":"Grace"}]}`))
		case "2":
			_, _ = w.Write([]byte(`{"contacts":[{"id":"c_3","first_name":"Kat"}]}`))
		default:
			t.Errorf("unexpected offset=%q", r.URL.Query().Get("offset"))
			_, _ = w.Write([]byte(`{"contacts":[]}`))
		}
	}))
	defer srv.Close()

	c := missive.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "limit": "2"},
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
	if sawAuth != "Bearer key_test_123" {
		t.Fatalf("Authorization = %q, want Bearer key_test_123", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	if len(sawLimits) != 2 {
		t.Fatalf("requests = %d, want 2 pages", len(sawLimits))
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("record missing id: %+v", rec)
		}
	}
}

// TestFixtureModeNoNetwork verifies that mode=fixture emits deterministic
// records without any network call, so conformance runs without credentials.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := missive.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "teams", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read(fixture): %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("fixture records = %d, want 2", len(got))
	}
	if got[0]["id"] == nil {
		t.Fatalf("fixture record missing id: %+v", got[0])
	}
}

// TestCheckFixtureMode confirms Check short-circuits without network in fixture
// mode and that the connector advertises a read-only catalog.
func TestCheckFixtureMode(t *testing.T) {
	c := missive.New()
	if err := c.Check(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("streams = %d, want >= 3", len(cat.Streams))
	}
}

// TestRegisteredReadOnly confirms self-registration via init() and that the
// connector resolves through the registry as a read-only source.
func TestRegisteredReadOnly(t *testing.T) {
	_ = missive.New() // ensure init ran
	caps := missive.New().Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (Missive source is read-only)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("missive"); !ok {
		t.Fatal("registry did not resolve missive (self-registration)")
	}
}
