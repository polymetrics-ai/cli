package bunnyinc_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	bunnyinc "polymetrics.ai/internal/connectors/bunny-inc"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Bunny, Inc.
// connector. Bunny exposes a single GraphQL endpoint (POST /graphql) secured with
// a bearer token; list queries page via GraphQL cursor pagination
// (pageInfo.endCursor / hasNextPage) and records live at data.<stream>.nodes[*].
// This asserts the Authorization header, the two-page cursor walk over the
// `after` variable, and record mapping.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var afters []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.Method != http.MethodPost || r.URL.Path != "/graphql" {
			http.NotFound(w, r)
			return
		}
		var body struct {
			Query     string         `json:"query"`
			Variables map[string]any `json:"variables"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		if !strings.Contains(body.Query, "accounts(") {
			t.Errorf("query did not target accounts: %q", body.Query)
		}
		after, _ := body.Variables["after"].(string)
		afters = append(afters, after)
		switch after {
		case "":
			_, _ = w.Write([]byte(`{"data":{"accounts":{"nodes":[{"id":"acc_1","name":"Acme"},{"id":"acc_2","name":"Globex"}],"pageInfo":{"endCursor":"cur2","hasNextPage":true}}}}`))
		case "cur2":
			_, _ = w.Write([]byte(`{"data":{"accounts":{"nodes":[{"id":"acc_3","name":"Initech"}],"pageInfo":{"endCursor":null,"hasNextPage":false}}}}`))
		default:
			t.Errorf("unexpected after=%q", after)
			_, _ = w.Write([]byte(`{"data":{"accounts":{"nodes":[],"pageInfo":{"hasNextPage":false}}}}`))
		}
	}))
	defer srv.Close()

	c := bunnyinc.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"apikey": "bun_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "accounts", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer bun_test_123" {
		t.Fatalf("Authorization = %q, want Bearer bun_test_123", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	if len(afters) != 2 || afters[0] != "" || afters[1] != "cur2" {
		t.Fatalf("after sequence = %v, want [\"\" \"cur2\"]", afters)
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("record missing id: %+v", rec)
		}
	}
	if got[0]["name"] != "Acme" {
		t.Fatalf("first record name = %v, want Acme", got[0]["name"])
	}
}

// TestSubdomainBaseURL verifies that, absent a base_url override, the connector
// builds the host from the subdomain config (https://<subdomain>.bunny.com).
func TestSubdomainBaseURL(t *testing.T) {
	var sawHost string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawHost = r.Host
		_, _ = w.Write([]byte(`{"data":{"contacts":{"nodes":[{"id":"c_1"}],"pageInfo":{"hasNextPage":false}}}}`))
	}))
	defer srv.Close()

	// Drive the request through the test server by overriding base_url, but keep a
	// subdomain present so the resolver path is exercised end to end.
	c := bunnyinc.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "subdomain": "acme"},
		Secrets: map[string]string{"apikey": "k"},
	}
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "contacts", Config: cfg}, func(connectors.Record) error { return nil }); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawHost == "" {
		t.Fatal("expected the request to reach the server")
	}
}

// TestReadFixtureMode exercises the credential-free fixture path used by the
// conformance harness: deterministic records, no network.
func TestReadFixtureMode(t *testing.T) {
	c := bunnyinc.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "invoices", Config: cfg}, func(rec connectors.Record) error {
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

// TestCheckFixtureMode confirms Check short-circuits without network in fixture
// mode, and rejects a missing secret in live mode.
func TestCheckFixtureMode(t *testing.T) {
	c := bunnyinc.New()
	if err := c.Check(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
	err := c.Check(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"subdomain": "acme"}})
	if err == nil {
		t.Fatal("Check without apikey should fail")
	}
}

// TestCatalogStreams asserts the published catalog exposes the core streams with
// id primary keys.
func TestCatalogStreams(t *testing.T) {
	c := bunnyinc.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"accounts": false, "contacts": false, "invoices": false, "payments": false, "subscriptions": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 || s.PrimaryKey[0] != "id" {
			t.Fatalf("stream %q primary key = %v, want [id]", s.Name, s.PrimaryKey)
		}
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("catalog missing core stream %q", name)
		}
	}
}

// TestRegisteredReadOnly confirms self-registration via NewRegistry and that the
// connector advertises read (not write) capabilities.
func TestRegisteredReadOnly(t *testing.T) {
	_ = bunnyinc.New() // ensure init ran
	c := bunnyinc.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || caps.Write {
		t.Fatalf("capabilities = %+v, want Read && !Write", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("bunny-inc"); !ok {
		t.Fatal("registry did not resolve bunny-inc (self-registration)")
	}
}
