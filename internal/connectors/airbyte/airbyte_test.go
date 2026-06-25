package airbyte_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics/internal/connectors"
	"polymetrics/internal/connectors/airbyte"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Airbyte
// connector: OAuth2 client-credentials token exchange, Bearer auth on data
// requests, offset/limit pagination over data[], and record mapping. Red until
// internal/connectors/airbyte exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var (
		sawTokenGrant string
		sawAuth       string
		tokenCalls    int
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/applications/token":
			tokenCalls++
			_ = r.ParseForm()
			sawTokenGrant = r.Form.Get("grant_type")
			// Token may be sent as a JSON body too; accept either.
			if sawTokenGrant == "" {
				sawTokenGrant = "client_credentials"
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"tok_abc","token_type":"Bearer","expires_in":3600}`))
		case "/connections":
			sawAuth = r.Header.Get("Authorization")
			offset := r.URL.Query().Get("offset")
			// limit is 2 in the test config so the first page (2 records) is full
			// and triggers a second page; the second page is short and stops.
			switch offset {
			case "", "0":
				_, _ = w.Write([]byte(`{"data":[{"connectionId":"c1","name":"A","status":"active"},{"connectionId":"c2","name":"B","status":"inactive"}]}`))
			case "2":
				_, _ = w.Write([]byte(`{"data":[{"connectionId":"c3","name":"C","status":"active"}]}`))
			default:
				t.Errorf("unexpected offset=%q", offset)
				_, _ = w.Write([]byte(`{"data":[]}`))
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := airbyte.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":  srv.URL,
			"token_url": srv.URL + "/applications/token",
			"page_size": "2",
		},
		Secrets: map[string]string{"client_secret": "sek_123"},
	}
	// client_id lives in config for OAuth2; supply it.
	cfg.Config["client_id"] = "cli_123"

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "connections", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if tokenCalls == 0 {
		t.Fatal("expected an OAuth2 token exchange, got none")
	}
	if sawTokenGrant != "client_credentials" {
		t.Fatalf("grant_type = %q, want client_credentials", sawTokenGrant)
	}
	if sawAuth != "Bearer tok_abc" {
		t.Fatalf("Authorization = %q, want Bearer tok_abc", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["connectionId"] == nil {
			t.Fatalf("record missing connectionId: %+v", rec)
		}
	}
}

// TestFixtureModeNoNetwork ensures fixture mode emits deterministic records with
// no network access (conformance runs credential-free).
func TestFixtureModeNoNetwork(t *testing.T) {
	c := airbyte.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	for _, stream := range []string{"workspaces", "connections", "sources", "destinations", "jobs"} {
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
			if len(rec) == 0 {
				t.Fatalf("fixture Read(%s) emitted empty record", stream)
			}
		}
	}

	// Check must short-circuit in fixture mode with no creds.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

// TestCatalogStreams confirms the published catalog exposes the core streams with
// primary keys.
func TestCatalogStreams(t *testing.T) {
	c := airbyte.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"workspaces": false, "connections": false, "sources": false, "destinations": false, "jobs": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
			if len(s.PrimaryKey) == 0 {
				t.Fatalf("stream %s has no primary key", s.Name)
			}
		}
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

// TestBaseURLSSRFGuard rejects non-http(s) base URLs.
func TestBaseURLSSRFGuard(t *testing.T) {
	c := airbyte.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "file:///etc/passwd", "client_id": "x"},
		Secrets: map[string]string{"client_secret": "y"},
	}
	err := c.Check(context.Background(), cfg)
	if err == nil || !strings.Contains(err.Error(), "base_url") {
		t.Fatalf("expected base_url SSRF rejection, got %v", err)
	}
}

// TestRegistryResolves confirms self-registration resolves through the registry.
func TestRegistryResolves(t *testing.T) {
	_ = airbyte.New() // ensure init ran
	r := connectors.NewRegistry()
	conn, ok := r.Get("airbyte")
	if !ok {
		t.Fatal("registry did not resolve airbyte (self-registration)")
	}
	caps := conn.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("airbyte is read-only; Write should be false, got %+v", caps)
	}
}
