package activecampaign_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics/internal/connectors"
	"polymetrics/internal/connectors/activecampaign"
)

// TestReadPaginatesAndAuthenticates is the red-first test: it asserts the
// Api-Token auth header, ActiveCampaign limit/offset pagination across two
// pages, and record mapping for the contacts stream.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawToken string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawToken = r.Header.Get("Api-Token")
		if r.URL.Path != "/contacts" {
			http.NotFound(w, r)
			return
		}
		// page size is 2 so two full records signals "more pages".
		offset := r.URL.Query().Get("offset")
		switch offset {
		case "", "0":
			_, _ = w.Write([]byte(`{"contacts":[{"id":"1","email":"a@example.com","cdate":"2026-01-01T00:00:00-05:00"},{"id":"2","email":"b@example.com","cdate":"2026-01-02T00:00:00-05:00"}],"meta":{"total":"3"}}`))
		case "2":
			_, _ = w.Write([]byte(`{"contacts":[{"id":"3","email":"c@example.com","cdate":"2026-01-03T00:00:00-05:00"}],"meta":{"total":"3"}}`))
		default:
			t.Errorf("unexpected offset=%q", offset)
			_, _ = w.Write([]byte(`{"contacts":[],"meta":{"total":"3"}}`))
		}
	}))
	defer srv.Close()

	c := activecampaign.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"api_key": "tok_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "contacts", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawToken != "tok_test_123" {
		t.Fatalf("Api-Token = %q, want tok_test_123", sawToken)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["email"] == nil {
			t.Fatalf("record missing id/email: %+v", rec)
		}
	}
}

// TestFixtureModeAndMissingSecret verifies that fixture mode short-circuits
// without network or credentials, and that a missing api_key is rejected by
// Check outside fixture mode.
func TestFixtureModeAndMissingSecret(t *testing.T) {
	c := activecampaign.New()

	// fixture mode: no network, no creds required.
	fixtureCfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), fixtureCfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
	var n int
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "deals", Config: fixtureCfg}, func(connectors.Record) error {
		n++
		return nil
	})
	if err != nil {
		t.Fatalf("Read(fixture) = %v", err)
	}
	if n == 0 {
		t.Fatal("fixture mode emitted no records")
	}

	// missing secret outside fixture mode is rejected by Check.
	bad := connectors.RuntimeConfig{Config: map[string]string{"account_username": "acme"}}
	if err := c.Check(context.Background(), bad); err == nil {
		t.Fatal("Check without api_key should fail")
	}
}

func TestCatalogStreams(t *testing.T) {
	c := activecampaign.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"contacts": false, "lists": false, "deals": false, "campaigns": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 {
			t.Errorf("stream %q missing primary key", s.Name)
		}
	}
	for name, seen := range want {
		if !seen {
			t.Errorf("catalog missing stream %q", name)
		}
	}
}

func TestRegisteredReadOnly(t *testing.T) {
	_ = activecampaign.New() // ensure init ran
	c := activecampaign.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("activecampaign"); !ok {
		t.Fatal("registry did not resolve activecampaign (self-registration)")
	}
}
