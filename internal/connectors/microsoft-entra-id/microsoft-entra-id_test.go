package microsoftentraid_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	microsoftentraid "polymetrics.ai/internal/connectors/microsoft-entra-id"
)

// TestReadPaginatesAndAuthenticates is the red-first test: it asserts the
// connector exchanges the OAuth2 client-credentials grant for a bearer token,
// presents that token on the Graph request, follows @odata.nextLink pagination
// across two pages, and maps records from the value[] array.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var (
		sawAuth      string
		sawGrant     string
		sawScope     string
		tokenCalls   int
		usersPage1   bool
		usersPage2   bool
		nextLinkBase string
	)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/oauth2/v2.0/token"):
			tokenCalls++
			_ = r.ParseForm()
			sawGrant = r.Form.Get("grant_type")
			sawScope = r.Form.Get("scope")
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"token_type":"Bearer","expires_in":3600,"access_token":"graph_token_abc"}`))
		case r.URL.Path == "/v1.0/users":
			sawAuth = r.Header.Get("Authorization")
			if r.URL.Query().Get("$skiptoken") == "PAGE2" {
				usersPage2 = true
				_, _ = w.Write([]byte(`{"value":[{"id":"u3","displayName":"Cleo","userPrincipalName":"cleo@example.com"}]}`))
				return
			}
			usersPage1 = true
			next := nextLinkBase + "/v1.0/users?$skiptoken=PAGE2"
			payload := map[string]any{
				"@odata.nextLink": next,
				"value": []map[string]any{
					{"id": "u1", "displayName": "Ada", "userPrincipalName": "ada@example.com"},
					{"id": "u2", "displayName": "Grace", "userPrincipalName": "grace@example.com"},
				},
			}
			b, _ := json.Marshal(payload)
			_, _ = w.Write(b)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	nextLinkBase = srv.URL

	c := microsoftentraid.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":  srv.URL + "/v1.0",
			"token_url": srv.URL + "/tenant/oauth2/v2.0/token",
		},
		Secrets: map[string]string{
			"client_id":     "client-123",
			"client_secret": "secret-xyz",
			"tenant_id":     "tenant-123",
			"user_id":       "user-123",
		},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "users", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if !usersPage1 || !usersPage2 {
		t.Fatalf("expected both pages fetched: page1=%v page2=%v", usersPage1, usersPage2)
	}
	if tokenCalls == 0 {
		t.Fatal("expected the OAuth2 token endpoint to be called")
	}
	if sawGrant != "client_credentials" {
		t.Fatalf("grant_type = %q, want client_credentials", sawGrant)
	}
	if sawScope != "https://graph.microsoft.com/.default" {
		t.Fatalf("scope = %q, want graph default scope", sawScope)
	}
	if sawAuth != "Bearer graph_token_abc" {
		t.Fatalf("Authorization = %q, want Bearer graph_token_abc", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("record missing id: %+v", rec)
		}
	}
	if got[0]["display_name"] != "Ada" {
		t.Fatalf("record mapping failed, display_name = %v, want Ada", got[0]["display_name"])
	}
}

// TestFixtureModeNoNetwork asserts fixture mode emits deterministic records
// without any HTTP server (conformance runs without live creds).
func TestFixtureModeNoNetwork(t *testing.T) {
	c := microsoftentraid.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	for _, stream := range []string{"users", "groups", "applications", "serviceprincipals", "directoryroles"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if len(got) != 2 {
			t.Fatalf("fixture Read(%s) records = %d, want 2", stream, len(got))
		}
		for _, rec := range got {
			if rec["id"] == nil {
				t.Fatalf("fixture %s record missing id: %+v", stream, rec)
			}
		}
	}
}

// TestCheckFixture confirms Check short-circuits in fixture mode.
func TestCheckFixture(t *testing.T) {
	c := microsoftentraid.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

// TestCatalogStreams asserts the catalog exposes the core streams with primary keys.
func TestCatalogStreams(t *testing.T) {
	c := microsoftentraid.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"users": false, "groups": false, "applications": false, "serviceprincipals": false, "directoryroles": false}
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

// TestRegistryResolves confirms self-registration via the global registry.
func TestRegistryResolves(t *testing.T) {
	_ = microsoftentraid.New() // ensure init ran
	r := connectors.NewRegistry()
	got, ok := r.Get("microsoft-entra-id")
	if !ok {
		t.Fatal("registry did not resolve microsoft-entra-id (self-registration)")
	}
	caps := got.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read+Catalog+Check", caps)
	}
	if caps.Write {
		t.Fatalf("microsoft-entra-id is read-only; Write should be false")
	}
}
