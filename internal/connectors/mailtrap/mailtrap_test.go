package mailtrap_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/mailtrap"
)

// TestReadAccountsAuthenticates verifies Bearer auth and root-array record
// mapping for the top-level accounts stream. Red until the package exists.
func TestReadAccountsAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/api/accounts" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`[{"id":1,"name":"Acme","access_levels":[1000]},{"id":2,"name":"Beta","access_levels":[100]}]`))
	}))
	defer srv.Close()

	c := mailtrap.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL + "/api"},
		Secrets: map[string]string{"api_token": "tok_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "accounts", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer tok_123" {
		t.Fatalf("Authorization = %q, want Bearer tok_123", sawAuth)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2", len(got))
	}
	if got[0]["id"] == nil || got[0]["name"] != "Acme" {
		t.Fatalf("record 0 mismatch: %+v", got[0])
	}
}

// TestReadInboxesPaginates verifies that an account-scoped stream issues the
// account-scoped path and pages across 2 pages, stopping when a page yields no
// new records.
func TestReadInboxesPaginates(t *testing.T) {
	var paths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		if r.URL.Path != "/api/accounts/42/inboxes" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page") {
		case "", "1":
			_, _ = w.Write([]byte(`[{"id":10,"name":"Inbox A","email_username":"a"},{"id":11,"name":"Inbox B","email_username":"b"}]`))
		case "2":
			_, _ = w.Write([]byte(`[{"id":12,"name":"Inbox C","email_username":"c"}]`))
		default:
			_, _ = w.Write([]byte(`[]`))
		}
	}))
	defer srv.Close()

	c := mailtrap.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":   srv.URL + "/api",
			"account_id": "42",
			"page_size":  "2",
		},
		Secrets: map[string]string{"api_token": "tok_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "inboxes", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 across 2 pages; paths=%v", len(got), paths)
	}
	if len(paths) < 2 {
		t.Fatalf("expected at least 2 page requests, got %v", paths)
	}
	ids := map[string]bool{}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("record missing id: %+v", rec)
		}
		ids[rec["account_id"].(string)] = true
	}
	if !ids["42"] {
		t.Fatalf("expected account_id stamped on inbox records, got %+v", got)
	}
}

// TestReadSendingDomainsDataPath verifies records under the {"data":[...]}
// envelope are extracted for the sending_domains stream.
func TestReadSendingDomainsDataPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/accounts/7/sending_domains" {
			http.NotFound(w, r)
			return
		}
		if r.URL.Query().Get("page") == "" || r.URL.Query().Get("page") == "1" {
			_, _ = w.Write([]byte(`{"data":[{"id":99,"domain_name":"example.com","status":"verified"}]}`))
			return
		}
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	defer srv.Close()

	c := mailtrap.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL + "/api", "account_id": "7"},
		Secrets: map[string]string{"api_token": "tok_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "sending_domains", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("records = %d, want 1", len(got))
	}
	if got[0]["domain_name"] != "example.com" {
		t.Fatalf("record mismatch: %+v", got[0])
	}
}

// TestFixtureModeNoNetwork ensures fixture mode emits deterministic records with
// no credentials or network access, so conformance passes credential-free.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := mailtrap.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	for _, stream := range []string{"accounts", "inboxes", "projects", "sending_domains"} {
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
		for _, rec := range got {
			if rec["id"] == nil {
				t.Fatalf("fixture %s record missing id: %+v", stream, rec)
			}
		}
	}

	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

// TestRegistryResolves confirms self-registration and read-only capabilities.
func TestRegistryResolves(t *testing.T) {
	_ = mailtrap.New() // ensure init ran
	r := connectors.NewRegistry()
	c, ok := r.Get("mailtrap")
	if !ok {
		t.Fatal("registry did not resolve mailtrap (self-registration)")
	}
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("mailtrap should be read-only, got Write=true")
	}
}

// TestCatalogStreams verifies the published catalog exposes the core streams.
func TestCatalogStreams(t *testing.T) {
	c := mailtrap.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"accounts": false, "inboxes": false, "projects": false, "sending_domains": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q missing primary key", s.Name)
		}
	}
	for name, found := range want {
		if !found {
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}
