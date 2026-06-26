package mendeley_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/mendeley"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Mendeley
// connector. It exercises the OAuth2 refresh-token grant (POST /oauth/token ->
// Bearer access token), the resource-specific Accept header, RFC 5988 Link-header
// pagination across two pages, and record mapping. Red until
// internal/connectors/mendeley exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var (
		sawTokenForm http.Request
		sawAuth      string
		sawAccept    string
		tokenHits    int
	)

	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/oauth/token":
			tokenHits++
			_ = r.ParseForm()
			sawTokenForm = *r
			if r.Form.Get("grant_type") != "refresh_token" {
				t.Errorf("token grant_type = %q, want refresh_token", r.Form.Get("grant_type"))
			}
			if r.Form.Get("refresh_token") != "rt_123" {
				t.Errorf("refresh_token = %q, want rt_123", r.Form.Get("refresh_token"))
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"at_abc","token_type":"bearer","expires_in":3600}`))
		case r.URL.Path == "/documents":
			sawAuth = r.Header.Get("Authorization")
			sawAccept = r.Header.Get("Accept")
			switch r.URL.Query().Get("marker") {
			case "":
				next := fmt.Sprintf("%s/documents?limit=2&marker=doc_2", srv.URL)
				w.Header().Set("Link", "<"+next+`>; rel="next"`)
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`[{"id":"doc_1","title":"First","last_modified":"2026-01-01T00:00:00.000Z"},{"id":"doc_2","title":"Second","last_modified":"2026-01-02T00:00:00.000Z"}]`))
			case "doc_2":
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`[{"id":"doc_3","title":"Third","last_modified":"2026-01-03T00:00:00.000Z"}]`))
			default:
				t.Errorf("unexpected marker=%q", r.URL.Query().Get("marker"))
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := mendeley.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":  srv.URL,
			"token_url": srv.URL + "/oauth/token",
		},
		Secrets: map[string]string{
			"client_id":            "cid",
			"client_secret":        "csec",
			"client_refresh_token": "rt_123",
		},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "documents", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if tokenHits == 0 {
		t.Fatal("expected at least one call to the token endpoint")
	}
	_ = sawTokenForm
	if sawAuth != "Bearer at_abc" {
		t.Fatalf("Authorization = %q, want Bearer at_abc", sawAuth)
	}
	if !strings.Contains(sawAccept, "vnd.mendeley-document") {
		t.Fatalf("Accept = %q, want resource-specific mendeley document type", sawAccept)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	if got[0]["id"] != "doc_1" || got[2]["id"] != "doc_3" {
		t.Fatalf("unexpected record ids: %v", got)
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["last_modified"] == nil {
			t.Fatalf("record missing id/last_modified: %+v", rec)
		}
	}
}

// TestFixtureModeNoNetwork verifies fixture mode emits deterministic records
// without any network access so conformance can run credential-free.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := mendeley.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	for _, stream := range []string{"documents", "folders", "groups", "annotations"} {
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

	// Check must succeed in fixture mode without creds.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
}

func TestCheckRequiresSecrets(t *testing.T) {
	c := mendeley.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{}}
	if err := c.Check(context.Background(), cfg); err == nil {
		t.Fatal("Check without secrets should fail")
	}
}

func TestCatalogStreams(t *testing.T) {
	c := mendeley.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "mendeley" {
		t.Fatalf("catalog connector = %q, want mendeley", cat.Connector)
	}
	want := map[string]bool{"documents": false, "folders": false, "groups": false, "annotations": false}
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
	_ = mendeley.New() // ensure init ran
	c := mendeley.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog {
		t.Fatalf("capabilities = %+v, want Read && Catalog", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("mendeley"); !ok {
		t.Fatal("registry did not resolve mendeley (self-registration)")
	}
}
