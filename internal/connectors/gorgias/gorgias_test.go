package gorgias_test

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/gorgias"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Gorgias
// connector: HTTP Basic auth (username + password), Gorgias cursor pagination
// via meta.next_cursor over data[], and record mapping. Gorgias list responses
// look like {"data":[...],"meta":{"next_cursor":"..."}} and the next page is
// requested with ?cursor=<next_cursor>.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/api/tickets" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("cursor") {
		case "":
			_, _ = w.Write([]byte(`{"data":[{"id":1,"subject":"first","status":"open"},{"id":2,"subject":"second","status":"closed"}],"meta":{"next_cursor":"CURSOR2"}}`))
		case "CURSOR2":
			_, _ = w.Write([]byte(`{"data":[{"id":3,"subject":"third","status":"open"}],"meta":{"next_cursor":null}}`))
		default:
			t.Errorf("unexpected cursor=%q", r.URL.Query().Get("cursor"))
			_, _ = w.Write([]byte(`{"data":[],"meta":{"next_cursor":null}}`))
		}
	}))
	defer srv.Close()

	c := gorgias.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"password": "api_key_123"},
	}
	// Username is a non-secret config value for Gorgias basic auth.
	cfg.Config["username"] = "agent@example.com"

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "tickets", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("agent@example.com:api_key_123"))
	if sawAuth != wantAuth {
		t.Fatalf("Authorization = %q, want %q", sawAuth, wantAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["subject"] == nil {
			t.Fatalf("record missing id/subject: %+v", rec)
		}
	}
}

// TestFixtureModeReadsWithoutNetwork verifies fixture mode emits deterministic
// records with no credentials and no network access (credential-free conformance).
func TestFixtureModeReadsWithoutNetwork(t *testing.T) {
	c := gorgias.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "tickets", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("fixture records = %d, want 2", len(got))
	}
	if got[0]["id"] == nil {
		t.Fatalf("fixture record missing id: %+v", got[0])
	}
}

// TestCheckFixtureMode confirms Check short-circuits in fixture mode without creds.
func TestCheckFixtureMode(t *testing.T) {
	c := gorgias.New()
	if err := c.Check(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

// TestCatalogStreams verifies the published catalog exposes the core streams.
func TestCatalogStreams(t *testing.T) {
	c := gorgias.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"tickets": false, "customers": false, "messages": false, "satisfaction-surveys": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 {
			t.Errorf("stream %q has no primary key", s.Name)
		}
	}
	for name, seen := range want {
		if !seen {
			t.Errorf("catalog missing stream %q", name)
		}
	}
}

// TestRegisteredReadOnly confirms self-registration and read-only capabilities.
func TestRegisteredReadOnly(t *testing.T) {
	_ = gorgias.New() // ensure init ran
	c := gorgias.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("gorgias should be read-only, got Write=true")
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("gorgias"); !ok {
		t.Fatal("registry did not resolve gorgias (self-registration)")
	}
}
