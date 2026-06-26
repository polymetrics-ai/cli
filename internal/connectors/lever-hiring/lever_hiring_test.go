package leverhiring_test

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	leverhiring "polymetrics.ai/internal/connectors/lever-hiring"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Lever Hiring
// connector: HTTP Basic auth (API key as username, blank password), Lever
// hasNext/next offset pagination over data[], and record mapping.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/postings" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("offset") {
		case "":
			_, _ = w.Write([]byte(`{"data":[{"id":"p_1","text":"Engineer","createdAt":1700000000},{"id":"p_2","text":"Designer","createdAt":1700000100}],"next":"cursor_2","hasNext":true}`))
		case "cursor_2":
			_, _ = w.Write([]byte(`{"data":[{"id":"p_3","text":"PM","createdAt":1700000200}],"hasNext":false}`))
		default:
			t.Errorf("unexpected offset=%q", r.URL.Query().Get("offset"))
			_, _ = w.Write([]byte(`{"data":[],"hasNext":false}`))
		}
	}))
	defer srv.Close()

	c := leverhiring.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "lever_key_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "postings", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("lever_key_123:"))
	if sawAuth != wantAuth {
		t.Fatalf("Authorization = %q, want %q", sawAuth, wantAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("record missing id: %+v", rec)
		}
	}
	if got[0]["text"] != "Engineer" {
		t.Fatalf("first record text = %v, want Engineer", got[0]["text"])
	}
}

// TestReadBearerAuth verifies the OAuth bearer path when an access token is
// supplied instead of an API key.
func TestReadBearerAuth(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		_, _ = w.Write([]byte(`{"data":[{"id":"u_1","name":"Ada","createdAt":1700000000}],"hasNext":false}`))
	}))
	defer srv.Close()

	c := leverhiring.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"access_token": "tok_abc"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "users", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer tok_abc" {
		t.Fatalf("Authorization = %q, want Bearer tok_abc", sawAuth)
	}
	if len(got) != 1 {
		t.Fatalf("records = %d, want 1", len(got))
	}
}

// TestFixtureModeNoNetwork verifies fixture mode emits deterministic records
// with no network access and no credentials.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := leverhiring.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "opportunities", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read fixture: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("fixture records = %d, want 2", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("fixture record missing id: %+v", rec)
		}
	}
}

// TestCheckFixture verifies Check short-circuits in fixture mode.
func TestCheckFixture(t *testing.T) {
	c := leverhiring.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check fixture = %v, want nil", err)
	}
}

// TestCatalogStreams verifies the published catalog includes the core streams.
func TestCatalogStreams(t *testing.T) {
	c := leverhiring.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"opportunities": false, "postings": false, "users": false, "requisitions": false, "stages": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q missing primary key", s.Name)
		}
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

// TestRejectsBadBaseURL verifies SSRF guard on base_url scheme.
func TestRejectsBadBaseURL(t *testing.T) {
	c := leverhiring.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example.com"},
		Secrets: map[string]string{"api_key": "k"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "postings", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "base_url") {
		t.Fatalf("Read with bad base_url err = %v, want base_url scheme error", err)
	}
}

// TestRegisteredReadOnly verifies self-registration and read-only capability.
func TestRegisteredReadOnly(t *testing.T) {
	_ = leverhiring.New() // ensure init ran
	c := leverhiring.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write false (read-only)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("lever-hiring"); !ok {
		t.Fatal("registry did not resolve lever-hiring (self-registration)")
	}
}
