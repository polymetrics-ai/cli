package pinterest_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics/internal/connectors"
	"polymetrics/internal/connectors/pinterest"
)

// testConfig builds a RuntimeConfig pointed at the test server with the three
// OAuth secrets Pinterest requires.
func testConfig(baseURL string) connectors.RuntimeConfig {
	return connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":   baseURL,
			"account_id": "549755885175",
		},
		Secrets: map[string]string{
			"credentials.client_id":     "client-abc",
			"credentials.client_secret": "secret-xyz",
			"credentials.refresh_token": "refresh-123",
		},
	}
}

// TestReadAuthenticatesViaRefreshTokenAndPaginates is the red-first test: it
// asserts the connector exchanges the refresh token for an access token (HTTP
// Basic on the token endpoint, refresh_token grant), then sends that access
// token as a Bearer header on the data request, and follows Pinterest's
// bookmark-cursor pagination across two pages of items[].
func TestReadAuthenticatesViaRefreshTokenAndPaginates(t *testing.T) {
	var sawTokenAuth string
	var sawTokenGrant string
	var sawDataAuth string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth/token":
			sawTokenAuth = r.Header.Get("Authorization")
			_ = r.ParseForm()
			sawTokenGrant = r.Form.Get("grant_type")
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"access-tok","token_type":"bearer","expires_in":3600}`))
		case "/boards":
			sawDataAuth = r.Header.Get("Authorization")
			switch r.URL.Query().Get("bookmark") {
			case "":
				_, _ = w.Write([]byte(`{"items":[{"id":"b1","name":"Board One"},{"id":"b2","name":"Board Two"}],"bookmark":"page2"}`))
			case "page2":
				_, _ = w.Write([]byte(`{"items":[{"id":"b3","name":"Board Three"}],"bookmark":null}`))
			default:
				t.Errorf("unexpected bookmark=%q", r.URL.Query().Get("bookmark"))
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := pinterest.New()
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "boards", Config: testConfig(srv.URL)}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if !strings.HasPrefix(sawTokenAuth, "Basic ") {
		t.Fatalf("token endpoint Authorization = %q, want Basic auth", sawTokenAuth)
	}
	if sawTokenGrant != "refresh_token" {
		t.Fatalf("token grant_type = %q, want refresh_token", sawTokenGrant)
	}
	if sawDataAuth != "Bearer access-tok" {
		t.Fatalf("data Authorization = %q, want Bearer access-tok", sawDataAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 across 2 pages", len(got))
	}
	if got[0]["id"] != "b1" || got[2]["id"] != "b3" {
		t.Fatalf("unexpected records: %+v", got)
	}
}

// TestReadAccountScopedStreamUsesAccountID asserts account-scoped streams embed
// the configured account_id in the path.
func TestReadAccountScopedStreamUsesAccountID(t *testing.T) {
	var sawPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/token" {
			_, _ = w.Write([]byte(`{"access_token":"access-tok","expires_in":3600}`))
			return
		}
		sawPath = r.URL.Path
		_, _ = w.Write([]byte(`{"items":[{"id":"c1","name":"Campaign","status":"ACTIVE"}],"bookmark":null}`))
	}))
	defer srv.Close()

	c := pinterest.New()
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "campaigns", Config: testConfig(srv.URL)}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawPath != "/ad_accounts/549755885175/campaigns" {
		t.Fatalf("campaigns path = %q, want /ad_accounts/549755885175/campaigns", sawPath)
	}
	if len(got) != 1 || got[0]["id"] != "c1" {
		t.Fatalf("unexpected records: %+v", got)
	}
}

// TestAccountScopedStreamRequiresAccountID asserts a missing account_id is a
// clear configuration error for account-scoped streams.
func TestAccountScopedStreamRequiresAccountID(t *testing.T) {
	c := pinterest.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{},
		Secrets: map[string]string{"credentials.client_id": "a", "credentials.client_secret": "b", "credentials.refresh_token": "c"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "campaigns", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("expected error for account-scoped stream without account_id")
	}
}

// TestFixtureModeReadsWithoutNetwork asserts fixture mode emits deterministic
// records with no live credentials or network access (conformance path).
func TestFixtureModeReadsWithoutNetwork(t *testing.T) {
	c := pinterest.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture", "account_id": "123"}}
	for _, stream := range []string{"boards", "ad_accounts", "campaigns", "ad_groups", "audiences"} {
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

// TestCheckFixtureModeNoNetwork asserts Check short-circuits in fixture mode.
func TestCheckFixtureModeNoNetwork(t *testing.T) {
	c := pinterest.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check fixture mode = %v, want nil", err)
	}
}

// TestCatalogStreams asserts the published catalog includes the core streams
// with primary keys.
func TestCatalogStreams(t *testing.T) {
	c := pinterest.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"boards": false, "ad_accounts": false, "campaigns": false, "ad_groups": false, "audiences": false}
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
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

// TestRegisteredReadOnly asserts self-registration and read-only capabilities.
func TestRegisteredReadOnly(t *testing.T) {
	_ = pinterest.New()
	caps := pinterest.New().Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read/Catalog/Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("pinterest"); !ok {
		t.Fatal("registry did not resolve pinterest (self-registration)")
	}
}
