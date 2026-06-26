package drip_test

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/drip"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Drip
// connector: HTTP Basic auth (api_key as username, blank password), Drip's
// page/meta.total_pages pagination over the subscribers[] array, and record
// mapping. Red until internal/connectors/drip exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var sawPages []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/v2/acct_123/subscribers" {
			http.NotFound(w, r)
			return
		}
		page := r.URL.Query().Get("page")
		if page == "" {
			page = "1"
		}
		sawPages = append(sawPages, page)
		switch page {
		case "1":
			_, _ = w.Write([]byte(`{"meta":{"page":1,"count":2,"total_pages":2,"total_count":3},"subscribers":[{"id":"sub_1","email":"a@example.com","status":"active","created_at":"2026-01-01T00:00:00Z"},{"id":"sub_2","email":"b@example.com","status":"active","created_at":"2026-01-02T00:00:00Z"}]}`))
		case "2":
			_, _ = w.Write([]byte(`{"meta":{"page":2,"count":1,"total_pages":2,"total_count":3},"subscribers":[{"id":"sub_3","email":"c@example.com","status":"unsubscribed","created_at":"2026-01-03T00:00:00Z"}]}`))
		default:
			t.Errorf("unexpected page=%q", page)
			_, _ = w.Write([]byte(`{"meta":{"page":3,"count":0,"total_pages":2,"total_count":3},"subscribers":[]}`))
		}
	}))
	defer srv.Close()

	c := drip.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL + "/v2", "account_id": "acct_123"},
		Secrets: map[string]string{"api_key": "key_abc"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "subscribers", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("key_abc:"))
	if sawAuth != wantAuth {
		t.Fatalf("Authorization = %q, want %q", sawAuth, wantAuth)
	}
	if len(sawPages) != 2 || sawPages[0] != "1" || sawPages[1] != "2" {
		t.Fatalf("pages requested = %v, want [1 2]", sawPages)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["email"] == nil || rec["created_at"] == nil {
			t.Fatalf("record missing id/email/created_at: %+v", rec)
		}
	}
	if got[0]["id"] != "sub_1" || got[2]["status"] != "unsubscribed" {
		t.Fatalf("unexpected mapped records: %+v", got)
	}
}

// TestReadAccountsNotAccountScoped verifies the accounts stream hits the
// account-agnostic /v2/accounts endpoint (no account_id path segment).
func TestReadAccountsNotAccountScoped(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/accounts" {
			t.Errorf("accounts path = %q, want /v2/accounts", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"accounts":[{"id":"acct_123","name":"Acme","created_at":"2026-01-01T00:00:00Z"}]}`))
	}))
	defer srv.Close()

	c := drip.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL + "/v2", "account_id": "acct_123"},
		Secrets: map[string]string{"api_key": "key_abc"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "accounts", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read accounts: %v", err)
	}
	if len(got) != 1 || got[0]["id"] != "acct_123" {
		t.Fatalf("accounts records = %+v, want one acct_123", got)
	}
}

// TestFixtureModeNoNetwork ensures fixture mode emits deterministic records
// without any network access so conformance passes without creds.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := drip.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"subscribers", "campaigns", "broadcasts", "accounts"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if len(got) != 2 {
			t.Fatalf("fixture %s records = %d, want 2", stream, len(got))
		}
		for _, rec := range got {
			if rec["id"] == nil {
				t.Fatalf("fixture %s record missing id: %+v", stream, rec)
			}
		}
	}
}

func TestCheckFixtureMode(t *testing.T) {
	c := drip.New()
	if err := c.Check(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

func TestCatalogStreams(t *testing.T) {
	c := drip.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"subscribers": true, "campaigns": true, "broadcasts": true, "accounts": true}
	for _, s := range cat.Streams {
		delete(want, s.Name)
		if len(s.PrimaryKey) == 0 {
			t.Errorf("stream %q has no primary key", s.Name)
		}
	}
	if len(want) != 0 {
		t.Fatalf("catalog missing streams: %v", want)
	}
}

func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := drip.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example", "account_id": "acct_123"},
		Secrets: map[string]string{"api_key": "key_abc"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "subscribers", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "base_url") {
		t.Fatalf("Read with ftp base_url err = %v, want base_url scheme error", err)
	}
}

func TestRegisteredReadOnly(t *testing.T) {
	_ = drip.New() // ensure init ran
	c := drip.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only connector)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("drip"); !ok {
		t.Fatal("registry did not resolve drip (self-registration)")
	}
}
