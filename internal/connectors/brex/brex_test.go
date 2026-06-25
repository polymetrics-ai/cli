package brex_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/brex"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Brex
// connector: Bearer auth on the user_token secret, Brex cursor/next_cursor
// pagination over items[], and record mapping. Red until
// internal/connectors/brex exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/v2/users" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("cursor") {
		case "":
			_, _ = w.Write([]byte(`{"items":[{"id":"usr_1","email":"a@example.com","status":"ACTIVE"},{"id":"usr_2","email":"b@example.com","status":"ACTIVE"}],"next_cursor":"page2"}`))
		case "page2":
			_, _ = w.Write([]byte(`{"items":[{"id":"usr_3","email":"c@example.com","status":"ACTIVE"}],"next_cursor":null}`))
		default:
			t.Errorf("unexpected cursor=%q", r.URL.Query().Get("cursor"))
			_, _ = w.Write([]byte(`{"items":[],"next_cursor":null}`))
		}
	}))
	defer srv.Close()

	c := brex.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"user_token": "tok_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "users", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer tok_test_123" {
		t.Fatalf("Authorization = %q, want Bearer tok_test_123", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("record missing id: %+v", rec)
		}
		if rec["email"] == nil {
			t.Fatalf("record missing mapped email: %+v", rec)
		}
	}
}

// TestReadTransactionsPath confirms a non-default stream routes to its real
// Brex endpoint path (transactions live at /v2/transactions/card/primary).
func TestReadTransactionsPath(t *testing.T) {
	var sawPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawPath = r.URL.Path
		_, _ = w.Write([]byte(`{"items":[{"id":"txn_1","amount":{"amount":1000,"currency":"USD"},"posted_at_date":"2026-01-02"}],"next_cursor":null}`))
	}))
	defer srv.Close()

	c := brex.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"user_token": "tok_test_123"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "transactions", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read transactions: %v", err)
	}
	if sawPath != "/v2/transactions/card/primary" {
		t.Fatalf("transactions path = %q, want /v2/transactions/card/primary", sawPath)
	}
	if len(got) != 1 || got[0]["id"] != "txn_1" {
		t.Fatalf("unexpected transactions records: %+v", got)
	}
}

// TestFixtureModeNoNetwork confirms fixture mode emits deterministic records
// without any network call and requires no secret.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := brex.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "expenses", Config: cfg}, func(rec connectors.Record) error {
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
		if rec["id"] == nil {
			t.Fatalf("fixture record missing id: %+v", rec)
		}
	}
}

func TestCheckFixtureMode(t *testing.T) {
	c := brex.New()
	if err := c.Check(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

func TestCatalogStreams(t *testing.T) {
	c := brex.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"transactions": false, "users": false, "expenses": false, "vendors": false, "budgets": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q missing primary key", s.Name)
		}
		if len(s.Fields) == 0 {
			t.Fatalf("stream %q missing fields", s.Name)
		}
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("catalog missing core stream %q", name)
		}
	}
}

func TestRegisteredReadOnly(t *testing.T) {
	_ = brex.New() // ensure init ran
	caps := brex.New().Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("brex is read-only, Write must be false: %+v", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("brex"); !ok {
		t.Fatal("registry did not resolve brex (self-registration)")
	}
}
