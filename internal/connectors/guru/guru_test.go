package guru_test

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/guru"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Guru
// connector: HTTP Basic auth (email:token), RFC 5988 Link-header pagination
// across two pages, top-level-array record mapping. Red until
// internal/connectors/guru is implemented.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/collections" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page") {
		case "", "1":
			// Advertise a next page via the Link header (absolute URL back to us).
			w.Header().Set("Link", fmt.Sprintf(`<%s/collections?page=2>; rel="next"`, srv.URL))
			_, _ = w.Write([]byte(`[{"id":"col_1","name":"Engineering","slug":"abc/engineering"},{"id":"col_2","name":"Sales","slug":"def/sales"}]`))
		case "2":
			// No Link header => last page.
			_, _ = w.Write([]byte(`[{"id":"col_3","name":"Support","slug":"ghi/support"}]`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
			_, _ = w.Write([]byte(`[]`))
		}
	}))
	defer srv.Close()

	c := guru.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "username": "agent@example.com"},
		Secrets: map[string]string{"password": "tok_secret_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "collections", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("agent@example.com:tok_secret_123"))
	if sawAuth != wantAuth {
		t.Fatalf("Authorization = %q, want %q", sawAuth, wantAuth)
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

// TestFixtureModeNoNetwork confirms fixture mode emits deterministic records
// without any network call so conformance passes without live creds.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := guru.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "members", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read(fixture): %v", err)
	}
	if len(got) == 0 {
		t.Fatal("fixture mode produced no records")
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("fixture record missing id: %+v", rec)
		}
	}
}

// TestCheckFixtureMode short-circuits without network and without creds.
func TestCheckFixtureMode(t *testing.T) {
	c := guru.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

// TestCatalogStreams asserts the published catalog exposes the core streams.
func TestCatalogStreams(t *testing.T) {
	c := guru.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"collections": false, "groups": false, "members": false, "teams": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
	}
	for name, found := range want {
		if !found {
			t.Fatalf("catalog missing core stream %q", name)
		}
	}
}

// TestBaseURLSSRFValidation rejects a base_url with a non-http(s) scheme.
func TestBaseURLSSRFValidation(t *testing.T) {
	c := guru.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "file:///etc/passwd", "username": "a@b.com"},
		Secrets: map[string]string{"password": "tok"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "collections", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read with file:// base_url should be rejected (SSRF guard)")
	}
}

// TestRegisteredReadOnly confirms self-registration and read-only capabilities.
func TestRegisteredReadOnly(t *testing.T) {
	_ = guru.New() // ensure init ran
	c := guru.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("guru"); !ok {
		t.Fatal("registry did not resolve guru (self-registration)")
	}
}
