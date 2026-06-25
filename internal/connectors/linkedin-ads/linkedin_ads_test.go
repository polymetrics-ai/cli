package linkedinads_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics/internal/connectors"
	linkedinads "polymetrics/internal/connectors/linkedin-ads"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the LinkedIn Ads
// connector: it asserts Bearer auth on the access token, the LinkedIn-Version
// header, start/count offset pagination across two pages of elements[], and
// record mapping. Red until internal/connectors/linkedin-ads exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth, sawVersion, sawRestli string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		sawVersion = r.Header.Get("LinkedIn-Version")
		sawRestli = r.Header.Get("X-Restli-Protocol-Version")
		if r.URL.Path != "/adAccounts" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("start") {
		case "", "0":
			// Full first page of count=2 -> there is a next page.
			_, _ = w.Write([]byte(`{"elements":[{"id":1,"name":"Acct One","status":"ACTIVE"},{"id":2,"name":"Acct Two","status":"ACTIVE"}]}`))
		case "2":
			// Short page -> pagination stops after this.
			_, _ = w.Write([]byte(`{"elements":[{"id":3,"name":"Acct Three","status":"DRAFT"}]}`))
		default:
			t.Errorf("unexpected start=%q", r.URL.Query().Get("start"))
			_, _ = w.Write([]byte(`{"elements":[]}`))
		}
	}))
	defer srv.Close()

	c := linkedinads.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"credentials.access_token": "li_token_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "accounts", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer li_token_123" {
		t.Fatalf("Authorization = %q, want Bearer li_token_123", sawAuth)
	}
	if sawVersion == "" {
		t.Fatalf("LinkedIn-Version header was not set")
	}
	if sawRestli != "2.0.0" {
		t.Fatalf("X-Restli-Protocol-Version = %q, want 2.0.0", sawRestli)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["name"] == nil {
			t.Fatalf("record missing id/name: %+v", rec)
		}
	}
}

// TestFixtureModeReadsWithoutNetwork confirms fixture mode emits deterministic
// records with no creds and no network for every core stream.
func TestFixtureModeReadsWithoutNetwork(t *testing.T) {
	c := linkedinads.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"accounts", "campaign_groups", "campaigns", "creatives"} {
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

// TestCheckFixtureMode confirms Check short-circuits in fixture mode with no creds.
func TestCheckFixtureMode(t *testing.T) {
	c := linkedinads.New()
	if err := c.Check(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

// TestCheckRequiresSecret confirms Check fails without an access token when not
// in fixture mode.
func TestCheckRequiresSecret(t *testing.T) {
	c := linkedinads.New()
	err := c.Check(context.Background(), connectors.RuntimeConfig{Config: map[string]string{}})
	if err == nil {
		t.Fatal("Check without access token should fail")
	}
}

// TestCatalogStreams confirms the published catalog exposes the core streams.
func TestCatalogStreams(t *testing.T) {
	c := linkedinads.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"accounts": false, "campaign_groups": false, "campaigns": false, "creatives": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %s has no primary key", s.Name)
		}
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

// TestBaseURLRejectsBadScheme confirms SSRF guard on base_url override.
func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := linkedinads.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "file:///etc/passwd"},
		Secrets: map[string]string{"credentials.access_token": "x"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "accounts", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read with file:// base_url should be rejected")
	}
}

// TestRegisteredReadOnly confirms self-registration and read-only capabilities.
func TestRegisteredReadOnly(t *testing.T) {
	_ = linkedinads.New() // ensure init ran
	c := linkedinads.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("linkedin-ads"); !ok {
		t.Fatal("registry did not resolve linkedin-ads (self-registration)")
	}
}
