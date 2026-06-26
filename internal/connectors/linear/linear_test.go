package linear_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/linear"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Linear
// connector: the personal API key is sent as a bare Authorization header (no
// "Bearer" prefix), the GraphQL issues connection is paginated across two pages
// via pageInfo.hasNextPage/endCursor + the `after` variable, and each node is
// mapped to a connectors.Record. Red until internal/connectors/linear exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var sawMethod string
	var afters []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		sawMethod = r.Method
		if r.URL.Path != "/graphql" {
			http.NotFound(w, r)
			return
		}
		body, _ := io.ReadAll(r.Body)
		var payload struct {
			Query     string         `json:"query"`
			Variables map[string]any `json:"variables"`
		}
		_ = json.Unmarshal(body, &payload)
		after, _ := payload.Variables["after"].(string)
		afters = append(afters, after)

		w.Header().Set("Content-Type", "application/json")
		switch after {
		case "":
			_, _ = w.Write([]byte(`{"data":{"issues":{"nodes":[` +
				`{"id":"iss_1","identifier":"ENG-1","title":"First","createdAt":"2026-01-01T00:00:00.000Z","updatedAt":"2026-01-02T00:00:00.000Z"},` +
				`{"id":"iss_2","identifier":"ENG-2","title":"Second","createdAt":"2026-01-03T00:00:00.000Z","updatedAt":"2026-01-04T00:00:00.000Z"}` +
				`],"pageInfo":{"hasNextPage":true,"endCursor":"cursor_page_2"}}}}`))
		case "cursor_page_2":
			_, _ = w.Write([]byte(`{"data":{"issues":{"nodes":[` +
				`{"id":"iss_3","identifier":"ENG-3","title":"Third","createdAt":"2026-01-05T00:00:00.000Z","updatedAt":"2026-01-06T00:00:00.000Z"}` +
				`],"pageInfo":{"hasNextPage":false,"endCursor":"cursor_page_3"}}}}`))
		default:
			t.Errorf("unexpected after=%q", after)
			_, _ = w.Write([]byte(`{"data":{"issues":{"nodes":[],"pageInfo":{"hasNextPage":false,"endCursor":null}}}}`))
		}
	}))
	defer srv.Close()

	c := linear.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "lin_api_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "issues", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	// Personal API keys must be sent WITHOUT the "Bearer" prefix.
	if sawAuth != "lin_api_test_123" {
		t.Fatalf("Authorization = %q, want bare api key (no Bearer prefix)", sawAuth)
	}
	if sawMethod != http.MethodPost {
		t.Fatalf("method = %q, want POST (GraphQL)", sawMethod)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	// Confirm the second page was fetched with the endCursor from page 1.
	foundCursor := false
	for _, a := range afters {
		if a == "cursor_page_2" {
			foundCursor = true
		}
	}
	if !foundCursor {
		t.Fatalf("after values = %v, want one request with after=cursor_page_2", afters)
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["identifier"] == nil {
			t.Fatalf("record missing id/identifier: %+v", rec)
		}
	}
	if got[0]["title"] != "First" {
		t.Fatalf("record mapping wrong: %+v", got[0])
	}
}

// TestFixtureModeNoNetwork verifies fixture mode emits deterministic records
// without any HTTP access (used by the conformance harness without creds).
func TestFixtureModeNoNetwork(t *testing.T) {
	c := linear.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"issues", "teams", "projects", "users"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if len(got) != 2 {
			t.Fatalf("fixture %s records = %d, want 2", stream, len(got))
		}
		if got[0]["id"] == nil {
			t.Fatalf("fixture %s record missing id: %+v", stream, got[0])
		}
	}
}

// TestCheckFixtureModeNoNetwork verifies Check short-circuits in fixture mode.
func TestCheckFixtureModeNoNetwork(t *testing.T) {
	c := linear.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

// TestBaseURLValidation rejects non-http(s) base URLs (SSRF guard).
func TestBaseURLValidation(t *testing.T) {
	c := linear.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "file:///etc/passwd"},
		Secrets: map[string]string{"api_key": "k"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "issues", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "base_url") {
		t.Fatalf("Read with bad base_url err = %v, want base_url validation error", err)
	}
}

// TestCatalogStreams asserts the published catalog contains the core streams.
func TestCatalogStreams(t *testing.T) {
	c := linear.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"issues": false, "teams": false, "projects": false, "users": false}
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
			t.Fatalf("catalog missing core stream %q", name)
		}
	}
}

// TestRegistryResolution confirms self-registration via init() resolves through
// the shared registry.
func TestRegistryResolution(t *testing.T) {
	_ = linear.New() // ensure package init ran
	r := connectors.NewRegistry()
	got, ok := r.Get("linear")
	if !ok {
		t.Fatal("registry did not resolve linear (self-registration)")
	}
	caps := got.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
}
