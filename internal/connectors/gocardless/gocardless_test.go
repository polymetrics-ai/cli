package gocardless_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/gocardless"
)

// TestReadPaginatesAndAuthenticates is the red-first test: it asserts Bearer
// auth, the required GoCardless-Version header, cursor pagination across two
// pages via meta.cursors.after, and record mapping for the payments stream.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth, sawVersion string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		sawVersion = r.Header.Get("GoCardless-Version")
		if r.URL.Path != "/payments" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Query().Get("after") {
		case "":
			_, _ = w.Write([]byte(`{"payments":[{"id":"PM0001","amount":1000,"currency":"GBP","status":"paid_out","created_at":"2026-01-01T00:00:00Z"},{"id":"PM0002","amount":2000,"currency":"GBP","status":"confirmed","created_at":"2026-01-02T00:00:00Z"}],"meta":{"cursors":{"after":"PM0002","before":null},"limit":50}}`))
		case "PM0002":
			_, _ = w.Write([]byte(`{"payments":[{"id":"PM0003","amount":3000,"currency":"GBP","status":"pending_submission","created_at":"2026-01-03T00:00:00Z"}],"meta":{"cursors":{"after":null,"before":"PM0003"},"limit":50}}`))
		default:
			t.Errorf("unexpected after=%q", r.URL.Query().Get("after"))
			_, _ = w.Write([]byte(`{"payments":[],"meta":{"cursors":{"after":null}}}`))
		}
	}))
	defer srv.Close()

	c := gocardless.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "gocardless_version": "2015-07-06"},
		Secrets: map[string]string{"access_token": "sandbox_abc123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "payments", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer sandbox_abc123" {
		t.Fatalf("Authorization = %q, want Bearer sandbox_abc123", sawAuth)
	}
	if sawVersion != "2015-07-06" {
		t.Fatalf("GoCardless-Version = %q, want 2015-07-06", sawVersion)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["status"] == nil {
			t.Fatalf("record missing id/status: %+v", rec)
		}
	}
	if got[0]["id"] != "PM0001" || got[2]["id"] != "PM0003" {
		t.Fatalf("unexpected record ordering: %v / %v", got[0]["id"], got[2]["id"])
	}
}

// TestFixtureModeReadIsCredentialFree confirms the deterministic fixture path
// works with no network and no secret, which conformance relies on.
func TestFixtureModeReadIsCredentialFree(t *testing.T) {
	c := gocardless.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	for _, stream := range []string{"payments", "mandates", "payouts", "refunds"} {
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
		for _, rec := range got {
			if rec["id"] == nil {
				t.Fatalf("fixture %s record missing id: %+v", stream, rec)
			}
		}
	}

	// Check must also short-circuit in fixture mode without creds.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

// TestCheckRequiresSecret verifies a real (non-fixture) Check rejects a missing
// access token before any network call.
func TestCheckRequiresSecret(t *testing.T) {
	c := gocardless.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"gocardless_version": "2015-07-06"}}
	if err := c.Check(context.Background(), cfg); err == nil {
		t.Fatal("Check should fail without access_token")
	}
}

// TestBaseURLRejectsBadScheme guards the SSRF check on base_url overrides.
func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := gocardless.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example.com", "gocardless_version": "2015-07-06"},
		Secrets: map[string]string{"access_token": "sandbox_abc"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "payments", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read should reject a non-http(s) base_url")
	}
}

// TestRegisteredReadOnly asserts self-registration and the read-only capability
// set (Write=false), resolvable through the shared registry.
func TestRegisteredReadOnly(t *testing.T) {
	c := gocardless.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Check || !caps.Catalog {
		t.Fatalf("capabilities = %+v, want Read && Check && Catalog", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("gocardless"); !ok {
		t.Fatal("registry did not resolve gocardless (self-registration)")
	}
}

// TestCatalogStreams confirms the published catalog exposes the core streams.
func TestCatalogStreams(t *testing.T) {
	c := gocardless.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"payments": false, "mandates": false, "payouts": false, "refunds": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %s missing primary key", s.Name)
		}
	}
	for name, found := range want {
		if !found {
			t.Fatalf("catalog missing expected stream %q", name)
		}
	}
}
