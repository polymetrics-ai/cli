package fleetio_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/fleetio"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Fleetio
// connector: it asserts the Fleetio auth headers (Authorization: Token <api_key>
// and Account-Token: <account_token>), cursor pagination over two pages using
// start_cursor/next_cursor wrapped in a "records" envelope, and record mapping.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth, sawAccount string
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		sawAccount = r.Header.Get("Account-Token")
		if r.URL.Path != "/vehicles" {
			http.NotFound(w, r)
			return
		}
		calls++
		switch r.URL.Query().Get("start_cursor") {
		case "":
			_, _ = w.Write([]byte(`{"records":[{"id":1,"name":"Truck 1"},{"id":2,"name":"Truck 2"}],"next_cursor":"abc123"}`))
		case "abc123":
			_, _ = w.Write([]byte(`{"records":[{"id":3,"name":"Truck 3"}],"next_cursor":null}`))
		default:
			t.Errorf("unexpected start_cursor=%q", r.URL.Query().Get("start_cursor"))
			_, _ = w.Write([]byte(`{"records":[],"next_cursor":null}`))
		}
	}))
	defer srv.Close()

	c := fleetio.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "key_abc", "account_token": "acct_xyz"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "vehicles", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Token key_abc" {
		t.Fatalf("Authorization = %q, want %q", sawAuth, "Token key_abc")
	}
	if sawAccount != "acct_xyz" {
		t.Fatalf("Account-Token = %q, want %q", sawAccount, "acct_xyz")
	}
	if calls != 2 {
		t.Fatalf("server calls = %d, want 2 (pagination across 2 pages)", calls)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("record missing id: %+v", rec)
		}
	}
}

func TestFixtureModeNoNetwork(t *testing.T) {
	c := fleetio.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	for _, stream := range []string{"vehicles", "contacts", "fuel_entries", "issues", "service_entries"} {
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

	// Check must short-circuit in fixture mode without credentials.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

func TestCheckRequiresSecrets(t *testing.T) {
	c := fleetio.New()
	err := c.Check(context.Background(), connectors.RuntimeConfig{
		Secrets: map[string]string{"api_key": "key_abc"},
	})
	if err == nil {
		t.Fatal("Check should fail when account_token secret is missing")
	}
}

func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := fleetio.New()
	err := c.Read(context.Background(), connectors.ReadRequest{
		Stream: "vehicles",
		Config: connectors.RuntimeConfig{
			Config:  map[string]string{"base_url": "ftp://evil.example.com"},
			Secrets: map[string]string{"api_key": "k", "account_token": "a"},
		},
	}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read should reject a non-http(s) base_url scheme")
	}
}

func TestCatalogStreams(t *testing.T) {
	c := fleetio.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "fleetio" {
		t.Fatalf("catalog connector = %q, want fleetio", cat.Connector)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("catalog streams = %d, want >= 3", len(cat.Streams))
	}
	for _, s := range cat.Streams {
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q has no primary key", s.Name)
		}
		if len(s.Fields) == 0 {
			t.Fatalf("stream %q has no fields", s.Name)
		}
	}
}

func TestRegisteredReadOnly(t *testing.T) {
	_ = fleetio.New() // ensure init ran
	c := fleetio.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog {
		t.Fatalf("capabilities = %+v, want Read && Catalog", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, Fleetio connector is read-only", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("fleetio"); !ok {
		t.Fatal("registry did not resolve fleetio (self-registration)")
	}
}
