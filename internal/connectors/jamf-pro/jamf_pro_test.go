package jamfpro_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	jamfpro "polymetrics.ai/internal/connectors/jamf-pro"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Jamf Pro
// connector. It verifies the two-step auth (Basic credentials POST to the token
// endpoint, then Bearer token on data requests), page-based pagination over the
// {totalCount, results} shape across two pages, and record mapping.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawTokenBasicAuth string
	var sawBearer string
	tokenCalls := 0

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/auth/token":
			if r.Method != http.MethodPost {
				t.Errorf("token request method = %q, want POST", r.Method)
			}
			sawTokenBasicAuth = r.Header.Get("Authorization")
			tokenCalls++
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"token":"jamf-token-xyz","expires":"2099-01-01T00:00:00.000Z"}`))
		case "/v1/buildings":
			sawBearer = r.Header.Get("Authorization")
			page := r.URL.Query().Get("page")
			switch page {
			case "0":
				_, _ = w.Write([]byte(`{"totalCount":3,"results":[{"id":"1","name":"HQ"},{"id":"2","name":"Annex"}]}`))
			case "1":
				_, _ = w.Write([]byte(`{"totalCount":3,"results":[{"id":"3","name":"Warehouse"}]}`))
			default:
				t.Errorf("unexpected page=%q", page)
				_, _ = w.Write([]byte(`{"totalCount":3,"results":[]}`))
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := jamfpro.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":  srv.URL,
			"username":  "apiuser",
			"page_size": "2",
		},
		Secrets: map[string]string{"password": "s3cret"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "buildings", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	// Basic base64("apiuser:s3cret") == "YXBpdXNlcjpzM2NyZXQ="
	if sawTokenBasicAuth != "Basic YXBpdXNlcjpzM2NyZXQ=" {
		t.Fatalf("token Authorization = %q, want Basic credentials", sawTokenBasicAuth)
	}
	if sawBearer != "Bearer jamf-token-xyz" {
		t.Fatalf("data Authorization = %q, want Bearer jamf-token-xyz", sawBearer)
	}
	if tokenCalls != 1 {
		t.Fatalf("token endpoint called %d times, want 1 (token reused across pages)", tokenCalls)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["name"] == nil {
			t.Fatalf("record missing id/name: %+v", rec)
		}
	}
}

// TestFixtureModeNoNetwork verifies fixture mode emits deterministic records
// without any network access (used by credential-free conformance).
func TestFixtureModeNoNetwork(t *testing.T) {
	c := jamfpro.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "departments", Config: cfg}, func(rec connectors.Record) error {
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
	// Check must also short-circuit without network in fixture mode.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

func TestCatalogStreams(t *testing.T) {
	c := jamfpro.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "jamf-pro" {
		t.Fatalf("catalog connector = %q, want jamf-pro", cat.Connector)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("streams = %d, want at least 3", len(cat.Streams))
	}
	for _, s := range cat.Streams {
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q missing primary key", s.Name)
		}
	}
}

func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := jamfpro.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example.com", "username": "u"},
		Secrets: map[string]string{"password": "p"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "buildings", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("expected error for non-http(s) base_url scheme")
	}
}

func TestRegisteredReadOnly(t *testing.T) {
	_ = jamfpro.New() // ensure init ran
	c := jamfpro.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only source)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("jamf-pro"); !ok {
		t.Fatal("registry did not resolve jamf-pro (self-registration)")
	}
}
