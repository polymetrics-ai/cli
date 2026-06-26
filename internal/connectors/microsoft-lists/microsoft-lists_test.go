package microsoftlists_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	microsoftlists "polymetrics.ai/internal/connectors/microsoft-lists"
)

// tokenResponse is the canned OAuth2 client-credentials token reply the test
// servers return so the connector's Authenticator can mint a bearer token.
const tokenResponse = `{"token_type":"Bearer","access_token":"graph_test_token","expires_in":3600}`

// TestReadPaginatesAndAuthenticates is the red-first test: OAuth2 client
// credentials token fetch, Bearer auth on the Graph request, @odata.nextLink
// pagination over value[], and record mapping for the lists stream.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var sawTokenForm string

	mux := http.NewServeMux()

	// Token endpoint (login.microsoftonline.com stand-in).
	mux.HandleFunc("/tenant-123/oauth2/v2.0/token", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		sawTokenForm = r.Form.Encode()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(tokenResponse))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	// Graph endpoint (graph.microsoft.com stand-in). Registered after the
	// server exists so the absolute nextLink can point back at this server.
	mux.HandleFunc("/sites/site-1/lists", func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Query().Get("$skiptoken") {
		case "":
			_, _ = w.Write([]byte(`{"value":[` +
				`{"id":"l1","displayName":"Tasks","createdDateTime":"2026-01-01T00:00:00Z"},` +
				`{"id":"l2","displayName":"Issues","createdDateTime":"2026-01-02T00:00:00Z"}` +
				`],"@odata.nextLink":"` + srv.URL + `/sites/site-1/lists?%24skiptoken=page2"}`))
		case "page2":
			_, _ = w.Write([]byte(`{"value":[` +
				`{"id":"l3","displayName":"Docs","createdDateTime":"2026-01-03T00:00:00Z"}` +
				`]}`))
		default:
			t.Errorf("unexpected $skiptoken=%q", r.URL.Query().Get("$skiptoken"))
			_, _ = w.Write([]byte(`{"value":[]}`))
		}
	})

	c := microsoftlists.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":  srv.URL,
			"token_url": srv.URL + "/tenant-123/oauth2/v2.0/token",
			"site_id":   "site-1",
			"page_size": "2",
		},
		Secrets: map[string]string{
			"client_id":     "app-123",
			"client_secret": "shhh",
			"tenant_id":     "tenant-123",
			"domain":        "contoso.sharepoint.com",
		},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "lists", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer graph_test_token" {
		t.Fatalf("Authorization = %q, want Bearer graph_test_token", sawAuth)
	}
	if !strings.Contains(sawTokenForm, "grant_type=client_credentials") {
		t.Fatalf("token form = %q, want client_credentials grant", sawTokenForm)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["display_name"] == nil {
			t.Fatalf("record missing id/display_name: %+v", rec)
		}
	}
	if got[0]["id"] != "l1" || got[2]["id"] != "l3" {
		t.Fatalf("unexpected ids: %v ... %v", got[0]["id"], got[2]["id"])
	}
}

// TestFixtureMode verifies the credential-free fixture path emits deterministic
// records without any network access, so conformance can run without secrets.
func TestFixtureMode(t *testing.T) {
	c := microsoftlists.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture", "site_id": "demo"}}

	for _, stream := range []string{"lists", "list_items", "columns", "content_types"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("Read(%s) fixture: %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("Read(%s) fixture emitted no records", stream)
		}
		for _, rec := range got {
			if rec["id"] == nil {
				t.Fatalf("fixture %s record missing id: %+v", stream, rec)
			}
		}
	}
}

// TestCheckFixtureMode ensures Check short-circuits in fixture mode (no network).
func TestCheckFixtureMode(t *testing.T) {
	c := microsoftlists.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
}

// TestCatalogStreams verifies the published catalog exposes the core streams.
func TestCatalogStreams(t *testing.T) {
	c := microsoftlists.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"lists": false, "list_items": false, "columns": false, "content_types": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q has no primary key", s.Name)
		}
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

// TestRegisteredReadOnly verifies self-registration and read-only capabilities.
func TestRegisteredReadOnly(t *testing.T) {
	c := microsoftlists.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("microsoft-lists"); !ok {
		t.Fatal("registry did not resolve microsoft-lists (self-registration)")
	}
}
