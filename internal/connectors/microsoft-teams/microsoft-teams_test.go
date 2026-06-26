package microsoftteams_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	microsoftteams "polymetrics.ai/internal/connectors/microsoft-teams"
)

// TestReadPaginatesAndAuthenticates is the red-first test: the connector must
// exchange the OAuth2 client-credentials grant for a bearer token, send it as
// Authorization: Bearer on every Microsoft Graph request, follow the
// @odata.nextLink across two pages of value[], and map records.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var (
		sawTokenRequest bool
		sawAuth         string
		graphCalls      int
	)
	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	defer srv.Close()

	// OAuth2 token endpoint (client credentials grant).
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		sawTokenRequest = true
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse token form: %v", err)
		}
		if got := r.PostForm.Get("grant_type"); got != "client_credentials" {
			t.Errorf("grant_type = %q, want client_credentials", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"graph_tok_123","token_type":"Bearer","expires_in":3600}`))
	})

	// Microsoft Graph /users collection with @odata.nextLink pagination.
	mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		graphCalls++
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Query().Get("$skiptoken") {
		case "":
			next := srv.URL + "/users?$skiptoken=page2"
			_, _ = w.Write([]byte(`{"value":[{"id":"u1","displayName":"Ada","userPrincipalName":"ada@x.com"},{"id":"u2","displayName":"Grace","userPrincipalName":"grace@x.com"}],"@odata.nextLink":"` + next + `"}`))
		case "page2":
			_, _ = w.Write([]byte(`{"value":[{"id":"u3","displayName":"Kay","userPrincipalName":"kay@x.com"}]}`))
		default:
			t.Errorf("unexpected $skiptoken=%q", r.URL.Query().Get("$skiptoken"))
			_, _ = w.Write([]byte(`{"value":[]}`))
		}
	})

	c := microsoftteams.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":  srv.URL,
			"token_url": srv.URL + "/token",
			"client_id": "app-id",
		},
		Secrets: map[string]string{
			"client_secret": "shh",
			"refresh_token": "rt",
			"tenant_id":     "tenant-guid",
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
	if !sawTokenRequest {
		t.Fatal("connector never called the OAuth2 token endpoint")
	}
	if sawAuth != "Bearer graph_tok_123" {
		t.Fatalf("Authorization = %q, want Bearer graph_tok_123", sawAuth)
	}
	if graphCalls != 2 {
		t.Fatalf("graph calls = %d, want 2 (nextLink pagination)", graphCalls)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 across 2 pages", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["display_name"] == nil {
			t.Fatalf("record missing mapped id/display_name: %+v", rec)
		}
	}
}

// TestFixtureModeNeedsNoNetwork verifies the credential-free conformance path.
func TestFixtureModeNeedsNoNetwork(t *testing.T) {
	c := microsoftteams.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	for _, stream := range []string{"users", "groups", "channels", "team_device_usage_report"} {
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
	}
	if len(got) == 0 {
		t.Fatal("fixture mode emitted no records")
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("fixture record missing id: %+v", rec)
		}
	}
	// Check must also short-circuit without network in fixture mode.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

func TestCatalogStreams(t *testing.T) {
	c := microsoftteams.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("expected >=3 streams, got %d", len(cat.Streams))
	}
	for _, s := range cat.Streams {
		if len(s.PrimaryKey) == 0 {
			t.Errorf("stream %q missing primary key", s.Name)
		}
	}
}

func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := microsoftteams.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil"},
		Secrets: map[string]string{"client_secret": "x", "refresh_token": "y", "tenant_id": "z"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "users", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "base_url") {
		t.Fatalf("expected base_url scheme error, got %v", err)
	}
}

func TestRegisteredReadOnly(t *testing.T) {
	_ = microsoftteams.New()
	c := microsoftteams.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("microsoft-teams should be read-only, Write=%v", caps.Write)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("microsoft-teams"); !ok {
		t.Fatal("registry did not resolve microsoft-teams (self-registration)")
	}
}
