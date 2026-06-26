package getlago_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/getlago"
)

// TestReadPaginatesAndAuthenticates is the red-first test: it asserts Bearer auth
// using the api_key secret, Lago page-number pagination via meta.next_page across
// two pages, and that records carry the lago_id primary key after mapping.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/customers" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page") {
		case "", "1":
			_, _ = w.Write([]byte(`{"customers":[{"lago_id":"cust_1","external_id":"ext_1","created_at":"2026-01-01T00:00:00Z"},{"lago_id":"cust_2","external_id":"ext_2","created_at":"2026-01-02T00:00:00Z"}],"meta":{"current_page":1,"next_page":2,"total_pages":2}}`))
		case "2":
			_, _ = w.Write([]byte(`{"customers":[{"lago_id":"cust_3","external_id":"ext_3","created_at":"2026-01-03T00:00:00Z"}],"meta":{"current_page":2,"next_page":null,"total_pages":2}}`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
			_, _ = w.Write([]byte(`{"customers":[],"meta":{"next_page":null}}`))
		}
	}))
	defer srv.Close()

	c := getlago.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"api_url": srv.URL},
		Secrets: map[string]string{"api_key": "lago_key_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "customers", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer lago_key_123" {
		t.Fatalf("Authorization = %q, want Bearer lago_key_123", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["lago_id"] == nil || rec["created_at"] == nil {
			t.Fatalf("record missing lago_id/created_at: %+v", rec)
		}
	}
}

// TestFixtureModeReadsWithoutNetwork confirms the credential-free fixture path
// used by conformance: no base_url, mode=fixture, deterministic records.
func TestFixtureModeReadsWithoutNetwork(t *testing.T) {
	c := getlago.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "invoices", Config: cfg}, func(rec connectors.Record) error {
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
		if rec["lago_id"] == nil {
			t.Fatalf("fixture record missing lago_id: %+v", rec)
		}
	}
}

// TestCheckFixtureMode confirms Check short-circuits in fixture mode with no creds.
func TestCheckFixtureMode(t *testing.T) {
	c := getlago.New()
	if err := c.Check(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

// TestCatalogStreams confirms the published catalog has the core streams with
// lago_id primary keys.
func TestCatalogStreams(t *testing.T) {
	c := getlago.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("streams = %d, want >= 3", len(cat.Streams))
	}
	for _, s := range cat.Streams {
		if len(s.PrimaryKey) == 0 || s.PrimaryKey[0] != "lago_id" {
			t.Fatalf("stream %q primary key = %v, want [lago_id]", s.Name, s.PrimaryKey)
		}
	}
}

// TestReadOnlyCapability confirms write is disabled (Lago source is read-only).
func TestReadOnlyCapability(t *testing.T) {
	c := getlago.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Check || !caps.Catalog {
		t.Fatalf("capabilities = %+v, want Read && Check && Catalog", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false", caps)
	}
}

// TestRegistryResolution confirms self-registration via init() resolves through
// the connectors registry.
func TestRegistryResolution(t *testing.T) {
	_ = getlago.New() // ensure package init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("getlago"); !ok {
		t.Fatal("registry did not resolve getlago (self-registration)")
	}
}
