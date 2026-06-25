package castoredc_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	castoredc "polymetrics.ai/internal/connectors/castor-edc"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Castor EDC
// connector. It asserts that:
//   - the connector exchanges client credentials for a bearer token at the
//     OAuth2 token endpoint and sends "Authorization: Bearer <token>",
//   - it follows HAL page pagination across two pages, and
//   - it maps records out of _embedded.<collection>.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var tokenCalls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/oauth/token":
			tokenCalls++
			if r.Method != http.MethodPost {
				t.Errorf("token method = %s, want POST", r.Method)
			}
			_, _ = w.Write([]byte(`{"access_token":"tok_abc","token_type":"Bearer","expires_in":3600}`))
		case r.URL.Path == "/api/study":
			sawAuth = r.Header.Get("Authorization")
			switch r.URL.Query().Get("page") {
			case "", "1":
				_, _ = w.Write([]byte(`{"_embedded":{"study":[{"study_id":"S1","name":"Trial One"},{"study_id":"S2","name":"Trial Two"}]},"page":1,"page_count":2,"total_items":3}`))
			case "2":
				_, _ = w.Write([]byte(`{"_embedded":{"study":[{"study_id":"S3","name":"Trial Three"}]},"page":2,"page_count":2,"total_items":3}`))
			default:
				t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
				_, _ = w.Write([]byte(`{"_embedded":{"study":[]},"page":99,"page_count":2}`))
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := castoredc.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":  srv.URL + "/api",
			"token_url": srv.URL + "/oauth/token",
		},
		Secrets: map[string]string{
			"client_id":     "cid_123",
			"client_secret": "csecret_456",
		},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "study", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer tok_abc" {
		t.Fatalf("Authorization = %q, want Bearer tok_abc", sawAuth)
	}
	if tokenCalls == 0 {
		t.Fatal("expected at least one OAuth2 token exchange")
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["study_id"] == nil || rec["name"] == nil {
			t.Fatalf("record missing study_id/name: %+v", rec)
		}
	}
}

// TestReadNonPaginatedStream verifies a stream whose endpoint returns a single
// HAL page (e.g. country) is mapped correctly.
func TestReadNonPaginatedStream(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth/token":
			_, _ = w.Write([]byte(`{"access_token":"tok_xyz","expires_in":3600}`))
		case "/api/country":
			_, _ = w.Write([]byte(`{"_embedded":{"countries":[{"id":1,"country_name":"Netherlands"},{"id":2,"country_name":"United Kingdom"}]},"page":1,"page_count":1}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := castoredc.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":  srv.URL + "/api",
			"token_url": srv.URL + "/oauth/token",
		},
		Secrets: map[string]string{"client_id": "cid", "client_secret": "csec"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "country", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2", len(got))
	}
	if got[0]["country_name"] == nil {
		t.Fatalf("record missing country_name: %+v", got[0])
	}
}

// TestFixtureModeNoNetwork verifies fixture mode emits deterministic records
// without any network call, so conformance passes without live credentials.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := castoredc.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "study", Config: cfg}, func(rec connectors.Record) error {
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
		if rec["fixture"] != true {
			t.Fatalf("fixture record missing fixture marker: %+v", rec)
		}
	}
	// Check must also short-circuit in fixture mode.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

// TestCatalogStreams verifies the published catalog has the core streams with
// primary keys.
func TestCatalogStreams(t *testing.T) {
	c := castoredc.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("streams = %d, want >= 3", len(cat.Streams))
	}
	for _, s := range cat.Streams {
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q missing primary key", s.Name)
		}
		if len(s.Fields) == 0 {
			t.Fatalf("stream %q missing fields", s.Name)
		}
	}
}

// TestBaseURLValidation rejects non-http(s) base URLs (SSRF guard).
func TestBaseURLValidation(t *testing.T) {
	c := castoredc.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "file:///etc/passwd"},
		Secrets: map[string]string{"client_id": "x", "client_secret": "y"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "study", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "base_url") {
		t.Fatalf("expected base_url validation error, got %v", err)
	}
}

// TestRegistryResolution verifies self-registration via init() resolves through
// the public registry under the bare system name.
func TestRegistryResolution(t *testing.T) {
	_ = castoredc.New() // ensure init ran
	r := connectors.NewRegistry()
	conn, ok := r.Get("castor-edc")
	if !ok {
		t.Fatal("registry did not resolve castor-edc (self-registration)")
	}
	caps := conn.Metadata().Capabilities
	if !caps.Read || !caps.Catalog {
		t.Fatalf("capabilities = %+v, want Read && Catalog", caps)
	}
}
