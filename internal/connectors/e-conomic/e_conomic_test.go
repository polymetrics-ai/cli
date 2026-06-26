package economic_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	economic "polymetrics.ai/internal/connectors/e-conomic"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the e-conomic
// connector: dual-token header auth (X-AppSecretToken + X-AgreementGrantToken),
// pagination across two pages following the pagination.nextPage absolute URL,
// records extracted from the "collection" array, and record mapping.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAppSecret, sawGrant, sawContentType string
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAppSecret = r.Header.Get("X-AppSecretToken")
		sawGrant = r.Header.Get("X-AgreementGrantToken")
		if ct := r.Header.Get("Content-Type"); ct != "" {
			sawContentType = ct
		}
		if r.URL.Path != "/customers" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Query().Get("skippages") {
		case "", "0":
			// First page carries a nextPage absolute link back to this server.
			next := srv.URL + "/customers?skippages=1&pagesize=2"
			_, _ = w.Write([]byte(`{"collection":[` +
				`{"customerNumber":1,"name":"Acme A/S","currency":"DKK","email":"a@acme.dk"},` +
				`{"customerNumber":2,"name":"Beta ApS","currency":"DKK","email":"b@beta.dk"}` +
				`],"pagination":{"skipPages":0,"pageSize":2,"results":3,"nextPage":"` + next + `"}}`))
		case "1":
			_, _ = w.Write([]byte(`{"collection":[` +
				`{"customerNumber":3,"name":"Gamma GmbH","currency":"EUR","email":"g@gamma.de"}` +
				`],"pagination":{"skipPages":1,"pageSize":2,"results":3}}`))
		default:
			t.Errorf("unexpected skippages=%q", r.URL.Query().Get("skippages"))
			_, _ = w.Write([]byte(`{"collection":[],"pagination":{}}`))
		}
	}))
	defer srv.Close()

	c := economic.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{
			"app_secret_token":      "app-secret-123",
			"agreement_grant_token": "grant-456",
		},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "customers", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAppSecret != "app-secret-123" {
		t.Fatalf("X-AppSecretToken = %q, want app-secret-123", sawAppSecret)
	}
	if sawGrant != "grant-456" {
		t.Fatalf("X-AgreementGrantToken = %q, want grant-456", sawGrant)
	}
	if sawContentType != "" && !strings.HasPrefix(sawContentType, "application/json") {
		t.Fatalf("Content-Type = %q, want application/json", sawContentType)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["customer_number"] == nil {
			t.Fatalf("record missing customer_number: %+v", rec)
		}
		if rec["name"] == nil {
			t.Fatalf("record missing name: %+v", rec)
		}
	}
	if got[2]["name"] != "Gamma GmbH" {
		t.Fatalf("third record name = %v, want Gamma GmbH", got[2]["name"])
	}
}

// TestFixtureModeReadsWithoutNetwork ensures fixture mode emits deterministic
// records with no live credentials, so conformance passes credential-free.
func TestFixtureModeReadsWithoutNetwork(t *testing.T) {
	c := economic.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"customers", "products", "suppliers", "accounts", "invoices"} {
		var n int
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			n++
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if n == 0 {
			t.Fatalf("fixture Read(%s) emitted no records", stream)
		}
	}
	// Check should short-circuit in fixture mode (no creds).
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

// TestCatalogStreams confirms the catalog exposes the core streams with primary
// keys.
func TestCatalogStreams(t *testing.T) {
	c := economic.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("streams = %d, want >= 3", len(cat.Streams))
	}
	for _, s := range cat.Streams {
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %s missing primary key", s.Name)
		}
		if len(s.Fields) == 0 {
			t.Fatalf("stream %s missing fields", s.Name)
		}
	}
}

// TestBaseURLValidation rejects non-http(s) base_url overrides (SSRF guard).
func TestBaseURLValidation(t *testing.T) {
	c := economic.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{"base_url": "file:///etc/passwd"},
		Secrets: map[string]string{
			"app_secret_token":      "x",
			"agreement_grant_token": "y",
		},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "customers", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read with file:// base_url should fail SSRF validation")
	}
}

// TestRegisteredReadOnly confirms self-registration and read-only capabilities.
func TestRegisteredReadOnly(t *testing.T) {
	_ = economic.New() // ensure init ran
	c := economic.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write false (read-only source)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("e-conomic"); !ok {
		t.Fatal("registry did not resolve e-conomic (self-registration)")
	}
}
