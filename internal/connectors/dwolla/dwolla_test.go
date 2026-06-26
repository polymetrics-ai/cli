package dwolla_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/dwolla"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Dwolla
// connector. It exercises the full live path against an httptest server:
//   - OAuth2 client-credentials token exchange (form grant_type=client_credentials).
//   - Bearer auth carrying the minted token on the data request.
//   - The Dwolla HAL Accept header.
//   - HAL pagination across two pages via _links.next.href (absolute URLs).
//   - Record extraction from _embedded.customers and field mapping.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var (
		sawTokenGrant string
		sawDataAuth   string
		sawAccept     string
		tokenCalls    int
	)

	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	defer srv.Close()

	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		tokenCalls++
		_ = r.ParseForm()
		sawTokenGrant = r.Form.Get("grant_type")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"tok_abc","token_type":"bearer","expires_in":3600}`))
	})

	mux.HandleFunc("/customers", func(w http.ResponseWriter, r *http.Request) {
		sawDataAuth = r.Header.Get("Authorization")
		sawAccept = r.Header.Get("Accept")
		w.Header().Set("Content-Type", "application/vnd.dwolla.v1.hal+json")
		if r.URL.Query().Get("offset") == "25" {
			_, _ = w.Write([]byte(`{"_embedded":{"customers":[{"id":"cus_3","firstName":"Cee","status":"verified","created":"2026-01-03T00:00:00.000Z"}]}}`))
			return
		}
		// First page advertises a next link (absolute URL back to this server).
		next := srv.URL + "/customers?limit=25&offset=25"
		_, _ = w.Write([]byte(`{"_embedded":{"customers":[` +
			`{"id":"cus_1","firstName":"Ay","status":"verified","created":"2026-01-01T00:00:00.000Z"},` +
			`{"id":"cus_2","firstName":"Bee","status":"unverified","created":"2026-01-02T00:00:00.000Z"}` +
			`]},"_links":{"next":{"href":"` + next + `"}}}`))
	})

	c := dwolla.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "start_date": "2026-01-01T00:00:00Z"},
		Secrets: map[string]string{"client_id": "id_123", "client_secret": "sec_456"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "customers", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if sawTokenGrant != "client_credentials" {
		t.Fatalf("token grant_type = %q, want client_credentials", sawTokenGrant)
	}
	if tokenCalls == 0 {
		t.Fatal("expected the connector to call the OAuth token endpoint")
	}
	if sawDataAuth != "Bearer tok_abc" {
		t.Fatalf("data Authorization = %q, want Bearer tok_abc", sawDataAuth)
	}
	if !strings.Contains(sawAccept, "hal+json") {
		t.Fatalf("Accept = %q, want a Dwolla HAL media type", sawAccept)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 across 2 pages", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["created"] == nil {
			t.Fatalf("record missing id/created: %+v", rec)
		}
	}
	if got[0]["id"] != "cus_1" || got[2]["id"] != "cus_3" {
		t.Fatalf("unexpected record order: %v", []any{got[0]["id"], got[2]["id"]})
	}
}

// TestFixtureModeNoNetwork confirms fixture mode emits deterministic records
// without any network call, so credential-free conformance can run.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := dwolla.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	for _, stream := range []string{"customers", "events", "exchange_partners", "business_classifications"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("Read(%s) fixture: %v", stream, err)
		}
		if len(got) != 2 {
			t.Fatalf("Read(%s) fixture records = %d, want 2", stream, len(got))
		}
		for _, rec := range got {
			if rec["id"] == nil {
				t.Fatalf("fixture %s record missing id: %+v", stream, rec)
			}
		}
	}

	// Check in fixture mode must not require secrets or network.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

// TestCatalogAndMetadata checks the published catalog and read-only capabilities.
func TestCatalogAndMetadata(t *testing.T) {
	c := dwolla.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatal("dwolla is read-only; Write capability must be false")
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 4 {
		t.Fatalf("catalog streams = %d, want at least 4", len(cat.Streams))
	}
	for _, s := range cat.Streams {
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q missing primary key", s.Name)
		}
	}
}

// TestRegistryResolution confirms self-registration via init().
func TestRegistryResolution(t *testing.T) {
	_ = dwolla.New() // ensure the package init() ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("dwolla"); !ok {
		t.Fatal("registry did not resolve dwolla (self-registration)")
	}
}

// TestBaseURLRejectsBadScheme guards the SSRF check on base_url overrides.
func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := dwolla.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example.com"},
		Secrets: map[string]string{"client_id": "x", "client_secret": "y"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "customers", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("expected base_url scheme validation to reject ftp://")
	}
}
