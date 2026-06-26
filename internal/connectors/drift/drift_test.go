package drift_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/drift"
)

// TestReadConversationsCursorPaginatesAndAuthenticates is the red-first test for
// the Drift connector: Bearer auth, the conversations cursor pagination
// (pagination.more / pagination.next over the data[] array), and record mapping.
func TestReadConversationsCursorPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/conversations/list" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("next") {
		case "":
			_, _ = w.Write([]byte(`{"data":[{"id":1,"status":"open","contactId":10,"updatedAt":1700000100},{"id":2,"status":"closed","contactId":11,"updatedAt":1700000200}],"pagination":{"more":true,"next":"cursor_page_2"}}`))
		case "cursor_page_2":
			_, _ = w.Write([]byte(`{"data":[{"id":3,"status":"pending","contactId":12,"updatedAt":1700000300}],"pagination":{"more":false}}`))
		default:
			t.Errorf("unexpected next=%q", r.URL.Query().Get("next"))
			_, _ = w.Write([]byte(`{"data":[],"pagination":{"more":false}}`))
		}
	}))
	defer srv.Close()

	c := drift.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"credentials.access_token": "tok_abc"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "conversations", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer tok_abc" {
		t.Fatalf("Authorization = %q, want Bearer tok_abc", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["status"] == nil {
			t.Fatalf("record missing id/status: %+v", rec)
		}
	}
}

// TestReadAccountsNextURLPagination exercises the accounts stream, whose records
// live at data.accounts and whose pages are followed via the absolute "next"
// URL returned in the body.
func TestReadAccountsNextURLPagination(t *testing.T) {
	var srvURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/accounts" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("index") {
		case "", "0":
			_, _ = w.Write([]byte(`{"data":{"accounts":[{"accountId":"a1","name":"Acme","updateDateTime":1700000100},{"accountId":"a2","name":"Globex","updateDateTime":1700000200}],"next":"` + srvURL + `/accounts?index=2"}}`))
		case "2":
			_, _ = w.Write([]byte(`{"data":{"accounts":[{"accountId":"a3","name":"Initech","updateDateTime":1700000300}]}}`))
		default:
			t.Errorf("unexpected index=%q", r.URL.Query().Get("index"))
			_, _ = w.Write([]byte(`{"data":{"accounts":[]}}`))
		}
	}))
	defer srv.Close()
	srvURL = srv.URL

	c := drift.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"credentials.access_token": "tok_abc"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "accounts", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	if got[0]["account_id"] != "a1" {
		t.Fatalf("first account_id = %v, want a1", got[0]["account_id"])
	}
}

// TestReadUsersNoPagination exercises the users stream (records at data, no
// pagination).
func TestReadUsersNoPagination(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.URL.Path != "/users/list" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"data":[{"id":1,"email":"a@x.com","role":"admin"},{"id":2,"email":"b@x.com","role":"member"}]}`))
	}))
	defer srv.Close()

	c := drift.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"credentials.access_token": "tok_abc"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "users", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2", len(got))
	}
	if calls != 1 {
		t.Fatalf("server calls = %d, want 1 (no pagination)", calls)
	}
	if got[0]["email"] != "a@x.com" {
		t.Fatalf("first email = %v, want a@x.com", got[0]["email"])
	}
}

// TestFixtureModeNoNetwork ensures fixture mode emits deterministic records with
// no network access and no credentials, so conformance can run credential-free.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := drift.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"users", "accounts", "conversations", "contacts"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("Read fixture %s: %v", stream, err)
		}
		if len(got) != 2 {
			t.Fatalf("fixture %s records = %d, want 2", stream, len(got))
		}
	}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
}

// TestBaseURLRejectsBadScheme guards the SSRF validation on base_url overrides.
func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := drift.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example.com"},
		Secrets: map[string]string{"credentials.access_token": "tok_abc"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "users", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "base_url") {
		t.Fatalf("Read with bad scheme err = %v, want base_url error", err)
	}
}

func TestCatalogAndMetadata(t *testing.T) {
	c := drift.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read/Catalog/Check", caps)
	}
	if caps.Write {
		t.Fatalf("drift is read-only, Write should be false")
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("streams = %d, want >= 3", len(cat.Streams))
	}
}

func TestRegistryResolvesDrift(t *testing.T) {
	_ = drift.New() // ensure init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("drift"); !ok {
		t.Fatal("registry did not resolve drift (self-registration)")
	}
}
