package mode_test

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/mode"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Mode
// connector: HTTP Basic auth (api_token:api_secret), HAL+JSON
// _links.next.href pagination over the _embedded.spaces array, and record
// mapping. Red until internal/connectors/mode exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/acme/spaces" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page") {
		case "":
			// First page, points to a relative next link with a query.
			_, _ = w.Write([]byte(`{
				"_embedded":{"spaces":[
					{"token":"sp_1","name":"Sales","space_type":"custom"},
					{"token":"sp_2","name":"Marketing","space_type":"custom"}
				]},
				"_links":{"next":{"href":"/acme/spaces?page=2"}}
			}`))
		case "2":
			_, _ = w.Write([]byte(`{
				"_embedded":{"spaces":[
					{"token":"sp_3","name":"Finance","space_type":"private"}
				]},
				"_links":{}
			}`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
		}
	}))
	defer srv.Close()

	c := mode.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "workspace": "acme"},
		Secrets: map[string]string{"api_token": "tok_123", "api_secret": "sec_456"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "spaces", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("tok_123:sec_456"))
	if sawAuth != wantAuth {
		t.Fatalf("Authorization = %q, want %q", sawAuth, wantAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["token"] == nil || rec["name"] == nil {
			t.Fatalf("record missing token/name: %+v", rec)
		}
	}
	if got[0]["token"] != "sp_1" || got[2]["token"] != "sp_3" {
		t.Fatalf("unexpected tokens: %v ... %v", got[0]["token"], got[2]["token"])
	}
}

// TestFixtureModeNeedsNoNetwork ensures the credential-free fixture path emits
// deterministic records for the conformance harness.
func TestFixtureModeNeedsNoNetwork(t *testing.T) {
	c := mode.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "reports", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("fixture records = %d, want 2", len(got))
	}
	for _, rec := range got {
		if rec["token"] == nil {
			t.Fatalf("fixture record missing token: %+v", rec)
		}
	}
}

// TestCheckFixtureMode confirms Check short-circuits without network in fixture
// mode and that Catalog lists the published streams.
func TestCheckFixtureMode(t *testing.T) {
	c := mode.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture): %v", err)
	}
	cat, err := c.Catalog(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("streams = %d, want >= 3", len(cat.Streams))
	}
}

// TestRegisteredReadOnly verifies self-registration and capabilities.
func TestRegisteredReadOnly(t *testing.T) {
	_ = mode.New() // ensure init ran
	c := mode.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, Mode is read-only", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("mode"); !ok {
		t.Fatal("registry did not resolve mode (self-registration)")
	}
}

// TestBaseURLValidation rejects non-http(s) base URLs (SSRF guard).
func TestBaseURLValidation(t *testing.T) {
	c := mode.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "file:///etc/passwd", "workspace": "acme"},
		Secrets: map[string]string{"api_token": "t", "api_secret": "s"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "spaces", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read with file:// base_url should be rejected")
	}
}
