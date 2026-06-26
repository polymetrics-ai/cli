package merge_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/merge"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Merge
// connector: dual-token auth (Authorization: Bearer <api_token> +
// X-Account-Token: <account_token>), Merge cursor pagination over results[]
// driven by the body "next" field, and record mapping. Red until
// internal/connectors/merge exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth, sawAccount string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		sawAccount = r.Header.Get("X-Account-Token")
		if r.URL.Path != "/candidates" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("cursor") {
		case "":
			_, _ = w.Write([]byte(`{"next":"cursorPage2","previous":null,"results":[{"id":"cand_1","first_name":"Ada"},{"id":"cand_2","first_name":"Alan"}]}`))
		case "cursorPage2":
			_, _ = w.Write([]byte(`{"next":null,"previous":"cursorPage1","results":[{"id":"cand_3","first_name":"Grace"}]}`))
		default:
			t.Errorf("unexpected cursor=%q", r.URL.Query().Get("cursor"))
			_, _ = w.Write([]byte(`{"next":null,"previous":null,"results":[]}`))
		}
	}))
	defer srv.Close()

	c := merge.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_token": "key_test_123", "account_token": "acct_test_456"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "candidates", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer key_test_123" {
		t.Fatalf("Authorization = %q, want Bearer key_test_123", sawAuth)
	}
	if sawAccount != "acct_test_456" {
		t.Fatalf("X-Account-Token = %q, want acct_test_456", sawAccount)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	if got[0]["id"] != "cand_1" || got[2]["id"] != "cand_3" {
		t.Fatalf("unexpected ids: %v / %v", got[0]["id"], got[2]["id"])
	}
}

// TestReadFixtureMode confirms the credential-free fixture path emits
// deterministic records without any network access.
func TestReadFixtureMode(t *testing.T) {
	c := merge.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "jobs", Config: cfg}, func(rec connectors.Record) error {
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

// TestCheckFixtureMode confirms Check short-circuits without network in fixture
// mode.
func TestCheckFixtureMode(t *testing.T) {
	c := merge.New()
	if err := c.Check(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}); err != nil {
		t.Fatalf("Check fixture = %v, want nil", err)
	}
}

// TestCatalogStreams confirms the published catalog exposes the core streams
// with primary keys.
func TestCatalogStreams(t *testing.T) {
	c := merge.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "merge" {
		t.Fatalf("catalog connector = %q, want merge", cat.Connector)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("streams = %d, want >= 3", len(cat.Streams))
	}
	for _, s := range cat.Streams {
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q missing primary key", s.Name)
		}
	}
}

// TestRegisteredReadOnly confirms self-registration via the registry and that
// the connector is read-only (no reverse-ETL writes).
func TestRegisteredReadOnly(t *testing.T) {
	_ = merge.New() // ensure init ran
	c := merge.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("merge"); !ok {
		t.Fatal("registry did not resolve merge (self-registration)")
	}
}
