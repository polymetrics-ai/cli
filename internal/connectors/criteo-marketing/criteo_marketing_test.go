package criteomarketing_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	criteomarketing "polymetrics.ai/internal/connectors/criteo-marketing"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Criteo
// Marketing connector. It exercises:
//   - OAuth2 client-credentials token exchange (client_id/client_secret -> token)
//   - the resulting Bearer Authorization header on data requests
//   - offset/limit pagination across two pages of the ad-sets stream
//   - JSONAPI {data:[{id,attributes:{...}}]} record mapping
//
// Red until internal/connectors/criteo-marketing exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var (
		sawTokenForm string
		sawAuth      string
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/oauth2/token":
			_ = r.ParseForm()
			sawTokenForm = r.Form.Encode()
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"tok_abc","token_type":"Bearer","expires_in":900}`))
			return
		case strings.HasPrefix(r.URL.Path, "/2024-01/marketing-solutions/ad-sets/search"):
			sawAuth = r.Header.Get("Authorization")
			switch r.URL.Query().Get("offset") {
			case "", "0":
				_, _ = w.Write([]byte(`{"data":[{"id":"as_1","type":"AdSet","attributes":{"name":"Alpha","advertiserId":"adv_1","status":"Active"}},{"id":"as_2","type":"AdSet","attributes":{"name":"Beta","advertiserId":"adv_1","status":"Paused"}}]}`))
			case "2":
				_, _ = w.Write([]byte(`{"data":[{"id":"as_3","type":"AdSet","attributes":{"name":"Gamma","advertiserId":"adv_2","status":"Active"}}]}`))
			default:
				t.Errorf("unexpected offset=%q", r.URL.Query().Get("offset"))
				_, _ = w.Write([]byte(`{"data":[]}`))
			}
			return
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := criteomarketing.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":  srv.URL,
			"page_size": "2",
			"currency":  "USD",
		},
		Secrets: map[string]string{
			"client_id":     "cid_123",
			"client_secret": "csecret_456",
		},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "ad_sets", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer tok_abc" {
		t.Fatalf("Authorization = %q, want Bearer tok_abc", sawAuth)
	}
	if !strings.Contains(sawTokenForm, "grant_type=client_credentials") {
		t.Fatalf("token form = %q, want grant_type=client_credentials", sawTokenForm)
	}
	if !strings.Contains(sawTokenForm, "client_id=cid_123") {
		t.Fatalf("token form = %q, want client_id=cid_123", sawTokenForm)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["name"] == nil {
			t.Fatalf("record missing id/name: %+v", rec)
		}
	}
	if got[0]["id"] != "as_1" || got[0]["name"] != "Alpha" || got[0]["status"] != "Active" {
		t.Fatalf("first record mapping wrong: %+v", got[0])
	}
}

// TestFixtureModeReadsWithoutNetwork verifies fixture mode emits deterministic
// records with no network access, which is what the credential-free conformance
// harness exercises.
func TestFixtureModeReadsWithoutNetwork(t *testing.T) {
	c := criteomarketing.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture", "currency": "USD"}}

	// Each stream's deterministic record must populate its primary key fields.
	pk := map[string][]string{
		"ad_sets":     {"id"},
		"advertisers": {"id"},
		"campaigns":   {"id"},
		"audiences":   {"id"},
		"statistics":  {"AdvertiserId", "CampaignId", "Day"},
	}
	for _, stream := range []string{"ad_sets", "advertisers", "campaigns", "audiences", "statistics"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture Read(%s): no records", stream)
		}
		for _, rec := range got {
			for _, key := range pk[stream] {
				if rec[key] == nil {
					t.Fatalf("fixture %s record missing primary key %q: %+v", stream, key, rec)
				}
			}
		}
	}
}

// TestCheckFixtureMode confirms Check short-circuits in fixture mode (no creds).
func TestCheckFixtureMode(t *testing.T) {
	c := criteomarketing.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

// TestCheckRequiresSecrets confirms a non-fixture Check rejects missing creds
// without making a network call.
func TestCheckRequiresSecrets(t *testing.T) {
	c := criteomarketing.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"currency": "USD"}}
	if err := c.Check(context.Background(), cfg); err == nil {
		t.Fatal("Check without secrets should fail")
	}
}

// TestCatalogStreams confirms the published catalog exposes the core streams with
// primary keys.
func TestCatalogStreams(t *testing.T) {
	c := criteomarketing.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"ad_sets": false, "advertisers": false, "campaigns": false, "audiences": false, "statistics": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q missing primary key", s.Name)
		}
		if len(s.Fields) == 0 {
			t.Fatalf("stream %q missing fields", s.Name)
		}
	}
	for name, found := range want {
		if !found {
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

// TestBaseURLSSRFRejected confirms a non-http(s) base_url override is rejected.
func TestBaseURLSSRFRejected(t *testing.T) {
	c := criteomarketing.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "file:///etc/passwd"},
		Secrets: map[string]string{"client_id": "a", "client_secret": "b"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "ad_sets", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read with file:// base_url should be rejected")
	}
}

// TestRegisteredAndReadOnly verifies self-registration via the registry and the
// read-only capability profile.
func TestRegisteredAndReadOnly(t *testing.T) {
	_ = criteomarketing.New() // ensure init ran
	c := criteomarketing.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("criteo-marketing"); !ok {
		t.Fatal("registry did not resolve criteo-marketing (self-registration)")
	}
}
