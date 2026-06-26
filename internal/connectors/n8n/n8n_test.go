package n8n_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/n8n"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the n8n connector:
// X-N8N-API-KEY header auth, n8n nextCursor pagination over data[], and record
// mapping. n8n list responses are {data:[...], nextCursor:"..."}; the next page
// is requested with ?cursor=<nextCursor>.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("X-N8N-API-KEY")
		if r.URL.Path != "/api/v1/workflows" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("cursor") {
		case "":
			_, _ = w.Write([]byte(`{"data":[{"id":"1","name":"Alpha","active":true},{"id":"2","name":"Beta","active":false}],"nextCursor":"MQ"}`))
		case "MQ":
			_, _ = w.Write([]byte(`{"data":[{"id":"3","name":"Gamma","active":true}],"nextCursor":null}`))
		default:
			t.Errorf("unexpected cursor=%q", r.URL.Query().Get("cursor"))
			_, _ = w.Write([]byte(`{"data":[],"nextCursor":null}`))
		}
	}))
	defer srv.Close()

	c := n8n.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "n8n_test_key"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "workflows", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "n8n_test_key" {
		t.Fatalf("X-N8N-API-KEY = %q, want n8n_test_key", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["name"] == nil {
			t.Fatalf("record missing id/name: %+v", rec)
		}
	}
	if got[0]["id"] != "1" || got[2]["id"] != "3" {
		t.Fatalf("unexpected ids: %v, %v", got[0]["id"], got[2]["id"])
	}
}

// TestHostConfigDerivesBaseURL verifies that the n8n-native `host` config (an
// instance hostname without the /api/v1 suffix) is honored and the API version
// path is appended.
func TestHostConfigDerivesBaseURL(t *testing.T) {
	var sawPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawPath = r.URL.Path
		_, _ = w.Write([]byte(`{"data":[{"id":"42","name":"X"}],"nextCursor":null}`))
	}))
	defer srv.Close()

	c := n8n.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"host": srv.URL},
		Secrets: map[string]string{"api_key": "k"},
	}
	var n int
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "workflows", Config: cfg}, func(connectors.Record) error {
		n++
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawPath != "/api/v1/workflows" {
		t.Fatalf("path = %q, want /api/v1/workflows", sawPath)
	}
	if n != 1 {
		t.Fatalf("records = %d, want 1", n)
	}
}

// TestFixtureModeNoNetwork verifies the deterministic fixture path emits records
// with no network access so conformance can run credential-free.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := n8n.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	for _, stream := range []string{"workflows", "executions", "tags", "users"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("Read(%s) fixture: %v", stream, err)
		}
		if len(got) != 2 {
			t.Fatalf("fixture %s records = %d, want 2", stream, len(got))
		}
		if got[0]["id"] == nil {
			t.Fatalf("fixture %s record missing id: %+v", stream, got[0])
		}
	}

	// Check should short-circuit in fixture mode without creds.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
}

func TestCatalogStreams(t *testing.T) {
	c := n8n.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"workflows": false, "executions": false, "tags": false, "users": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q missing primary key", s.Name)
		}
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

func TestRegisteredReadOnly(t *testing.T) {
	_ = n8n.New() // ensure init ran
	caps := n8n.New().Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("n8n is read-only; Write should be false, got %+v", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("n8n"); !ok {
		t.Fatal("registry did not resolve n8n (self-registration)")
	}
}

func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := n8n.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example.com"},
		Secrets: map[string]string{"api_key": "k"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "workflows", Config: cfg}, func(connectors.Record) error {
		return nil
	})
	if err == nil {
		t.Fatal("expected error for non-http base_url scheme")
	}
}
