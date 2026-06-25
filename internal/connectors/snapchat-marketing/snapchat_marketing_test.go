package snapchatmarketing_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	snapchatmarketing "polymetrics.ai/internal/connectors/snapchat-marketing"
)

// TestReadPaginatesAndAuthenticates is the red-first test: it stands up a fake
// Snapchat Ads API + OAuth2 token endpoint, then asserts (a) the refresh-token
// grant is exchanged for a bearer access token applied as Authorization, (b)
// cursor pagination follows paging.next_link across two pages, and (c) the
// {"campaigns":[{"campaign":{...}}]} envelope is unwrapped into flat records.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var sawGrant, sawRefresh string

	mux := http.NewServeMux()
	// OAuth2 token endpoint (refresh_token grant).
	mux.HandleFunc("/login/oauth2/access_token", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		sawGrant = r.Form.Get("grant_type")
		sawRefresh = r.Form.Get("refresh_token")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"act_test_999","token_type":"Bearer","expires_in":3600}`))
	})
	// Campaigns list endpoint with cursor pagination.
	mux.HandleFunc("/v1/adaccounts/ACC1/campaigns", func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Query().Get("cursor") {
		case "":
			// Page 1 advertises a next_link pointing back at this endpoint.
			next := "http://" + r.Host + "/v1/adaccounts/ACC1/campaigns?cursor=PAGE2"
			_, _ = w.Write([]byte(`{"request_status":"SUCCESS","campaigns":[` +
				`{"sub_request_status":"SUCCESS","campaign":{"id":"c1","updated_at":"2026-01-01T00:00:00.000Z","name":"Camp One","status":"ACTIVE","ad_account_id":"ACC1"}},` +
				`{"sub_request_status":"SUCCESS","campaign":{"id":"c2","updated_at":"2026-01-02T00:00:00.000Z","name":"Camp Two","status":"PAUSED","ad_account_id":"ACC1"}}` +
				`],"paging":{"next_link":"` + next + `"}}`))
		case "PAGE2":
			_, _ = w.Write([]byte(`{"request_status":"SUCCESS","campaigns":[` +
				`{"sub_request_status":"SUCCESS","campaign":{"id":"c3","updated_at":"2026-01-03T00:00:00.000Z","name":"Camp Three","status":"ACTIVE","ad_account_id":"ACC1"}}` +
				`]}`))
		default:
			t.Errorf("unexpected cursor=%q", r.URL.Query().Get("cursor"))
			_, _ = w.Write([]byte(`{"campaigns":[]}`))
		}
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := snapchatmarketing.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":       srv.URL + "/v1",
			"token_url":      srv.URL + "/login/oauth2/access_token",
			"ad_account_ids": "ACC1",
		},
		Secrets: map[string]string{
			"client_id":     "cid",
			"client_secret": "csecret",
			"refresh_token": "rtok",
		},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "campaigns", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawGrant != "refresh_token" {
		t.Fatalf("grant_type = %q, want refresh_token", sawGrant)
	}
	if sawRefresh != "rtok" {
		t.Fatalf("refresh_token = %q, want rtok", sawRefresh)
	}
	if sawAuth != "Bearer act_test_999" {
		t.Fatalf("Authorization = %q, want Bearer act_test_999", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	// Records must be unwrapped (flat) from the {"campaign":{...}} envelope.
	for _, rec := range got {
		if rec["id"] == nil || rec["updated_at"] == nil {
			t.Fatalf("record missing id/updated_at (envelope not unwrapped?): %+v", rec)
		}
		if _, wrapped := rec["campaign"]; wrapped {
			t.Fatalf("record still wrapped in campaign envelope: %+v", rec)
		}
	}
	if got[0]["name"] != "Camp One" {
		t.Fatalf("got[0].name = %v, want Camp One", got[0]["name"])
	}
}

// TestFixtureModeNoNetwork confirms fixture mode emits deterministic records
// without any network access so conformance can run credential-free.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := snapchatmarketing.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"organizations", "adaccounts", "campaigns", "adsquads", "ads"} {
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
}

// TestCheckFixtureMode confirms Check short-circuits in fixture mode (no creds).
func TestCheckFixtureMode(t *testing.T) {
	c := snapchatmarketing.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

// TestCatalogStreams confirms the published catalog exposes the core streams
// with primary keys.
func TestCatalogStreams(t *testing.T) {
	c := snapchatmarketing.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"organizations": false, "adaccounts": false, "campaigns": false, "adsquads": false, "ads": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
			if len(s.PrimaryKey) == 0 {
				t.Fatalf("stream %q missing primary key", s.Name)
			}
		}
	}
	for name, found := range want {
		if !found {
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

// TestRegistryResolves confirms the connector self-registers and is read-only.
func TestRegistryResolves(t *testing.T) {
	_ = snapchatmarketing.New()
	r := connectors.NewRegistry()
	got, ok := r.Get("snapchat-marketing")
	if !ok {
		t.Fatal("registry did not resolve snapchat-marketing (self-registration)")
	}
	if got.Name() != "snapchat-marketing" {
		t.Fatalf("Name() = %q, want snapchat-marketing", got.Name())
	}
	caps := got.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read+Catalog+Check", caps)
	}
	if caps.Write {
		t.Fatalf("snapchat-marketing should be read-only, got Write=true")
	}
}

// TestSSRFBaseURLValidation confirms a non-http(s) base_url override is rejected.
func TestSSRFBaseURLValidation(t *testing.T) {
	c := snapchatmarketing.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "file:///etc/passwd"},
		Secrets: map[string]string{"client_id": "a", "client_secret": "b", "refresh_token": "c"},
	}
	err := c.Check(context.Background(), cfg)
	if err == nil || !strings.Contains(err.Error(), "base_url") {
		t.Fatalf("Check with file:// base_url = %v, want base_url validation error", err)
	}
}
