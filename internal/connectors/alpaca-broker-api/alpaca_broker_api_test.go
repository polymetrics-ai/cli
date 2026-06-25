package alpacabrokerapi_test

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	alpacabrokerapi "polymetrics.ai/internal/connectors/alpaca-broker-api"
)

// TestReadPaginatesAndAuthenticates is the red-first test: HTTP Basic auth
// (API key id as username, secret key as password), page-based pagination over
// the top-level accounts array across 2 pages, and record mapping.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/accounts" {
			http.NotFound(w, r)
			return
		}
		// Alpaca account listings return a top-level JSON array. With a limit of
		// 2, a full page implies another page may follow; a short page ends it.
		switch r.URL.Query().Get("page_token") {
		case "":
			_, _ = w.Write([]byte(`[{"id":"acct_1","account_number":"100","status":"ACTIVE","created_at":"2026-01-01T00:00:00Z"},{"id":"acct_2","account_number":"101","status":"ACTIVE","created_at":"2026-01-02T00:00:00Z"}]`))
		case "acct_2":
			_, _ = w.Write([]byte(`[{"id":"acct_3","account_number":"102","status":"ACTIVE","created_at":"2026-01-03T00:00:00Z"}]`))
		default:
			t.Errorf("unexpected page_token=%q", r.URL.Query().Get("page_token"))
			_, _ = w.Write([]byte(`[]`))
		}
	}))
	defer srv.Close()

	c := alpacabrokerapi.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "limit": "2"},
		Secrets: map[string]string{"password": "secret_key_123"},
	}
	cfg.Config["username"] = "key_id_abc"

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "accounts", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("key_id_abc:secret_key_123"))
	if sawAuth != wantAuth {
		t.Fatalf("Authorization = %q, want %q", sawAuth, wantAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["account_number"] == nil {
			t.Fatalf("record missing id/account_number: %+v", rec)
		}
	}
}

// TestReadSingleObjectStream covers the clock stream, whose endpoint returns a
// single object rather than an array.
func TestReadSingleObjectStream(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/clock" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"timestamp":"2026-06-25T12:00:00Z","is_open":true,"next_open":"2026-06-26T13:30:00Z","next_close":"2026-06-25T20:00:00Z"}`))
	}))
	defer srv.Close()

	c := alpacabrokerapi.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "username": "key_id_abc"},
		Secrets: map[string]string{"password": "secret_key_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "clock", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("records = %d, want 1", len(got))
	}
	if got[0]["is_open"] != true {
		t.Fatalf("is_open = %v, want true", got[0]["is_open"])
	}
}

// TestFixtureModeNoNetwork verifies the credential-free conformance path emits
// deterministic records without any network access.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := alpacabrokerapi.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	for _, stream := range []string{"accounts", "assets", "calendar", "clock", "country_info"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture Read(%s) emitted no records", stream)
		}
	}

	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

func TestCatalogAndMetadata(t *testing.T) {
	c := alpacabrokerapi.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only API)", caps)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("streams = %d, want >= 3", len(cat.Streams))
	}
}

func TestRegistryResolution(t *testing.T) {
	_ = alpacabrokerapi.New() // ensure init() ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("alpaca-broker-api"); !ok {
		t.Fatal("registry did not resolve alpaca-broker-api (self-registration)")
	}
}

func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := alpacabrokerapi.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example.com", "username": "u"},
		Secrets: map[string]string{"password": "p"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "accounts", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read with ftp base_url should be rejected (SSRF guard)")
	}
}
