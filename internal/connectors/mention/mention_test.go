package mention_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/mention"
)

// TestReadAlertsPaginatesAndAuthenticates is the red-first test: it asserts the
// Authorization header carries the raw api_key (Mention uses no Bearer prefix),
// that the connector follows _links.more.params.cursor across two pages, and that
// records are extracted from the "alerts" field path and mapped.
func TestReadAlertsPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/accounts/acc_1/alerts" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("cursor") {
		case "":
			_, _ = w.Write([]byte(`{"alerts":[{"id":1,"name":"Brand A"},{"id":2,"name":"Brand B"}],"_links":{"more":{"href":"x","params":{"cursor":"page2"}}}}`))
		case "page2":
			_, _ = w.Write([]byte(`{"alerts":[{"id":3,"name":"Brand C"}],"_links":{}}`))
		default:
			t.Errorf("unexpected cursor=%q", r.URL.Query().Get("cursor"))
			_, _ = w.Write([]byte(`{"alerts":[],"_links":{}}`))
		}
	}))
	defer srv.Close()

	c := mention.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "account_id": "acc_1"},
		Secrets: map[string]string{"api_key": "key_abc"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "alert", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "key_abc" {
		t.Fatalf("Authorization = %q, want raw key_abc", sawAuth)
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

// TestReadAccountMeResolvesAccount asserts that when account_id is not configured,
// the connector resolves the current account via /accounts/me before reading
// account-scoped streams, and maps the single account object at field path
// "account".
func TestReadAccountMeDiscovery(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/accounts/me":
			_, _ = w.Write([]byte(`{"account":{"id":"acc_9","name":"Acme"}}`))
		case "/accounts/acc_9/alerts":
			_, _ = w.Write([]byte(`{"alerts":[{"id":7,"name":"Alert Seven"}],"_links":{}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := mention.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "key_abc"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "alert", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 1 || got[0]["id"] == nil {
		t.Fatalf("alerts after account discovery = %+v, want one record with id", got)
	}
}

// TestFixtureModeNoNetwork asserts fixture mode emits deterministic records with
// no network access so conformance passes without credentials.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := mention.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "alert", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read fixture: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("fixture records = %d, want 2", len(got))
	}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
}

// TestRegisteredReadOnly asserts the connector self-registers and is resolvable,
// and that it advertises read (not write) capability.
func TestRegisteredReadOnly(t *testing.T) {
	_ = mention.New()
	caps := mention.New().Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("mention"); !ok {
		t.Fatal("registry did not resolve mention (self-registration)")
	}
}

func TestCatalogStreams(t *testing.T) {
	c := mention.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"account_me": false, "account": false, "alert": false, "mention": false, "alert_tag": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}
