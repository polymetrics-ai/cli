package brevo_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics/internal/connectors"
	"polymetrics/internal/connectors/brevo"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Brevo
// connector: api-key header auth, offset/limit pagination over the contacts[]
// array (two pages), and record mapping. Red until internal/connectors/brevo
// is implemented.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAPIKey string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAPIKey = r.Header.Get("api-key")
		if r.URL.Path != "/contacts" {
			http.NotFound(w, r)
			return
		}
		// page size is 2 in this test; the connector advances offset by the
		// returned record count and stops once a short page comes back.
		switch r.URL.Query().Get("offset") {
		case "", "0":
			_, _ = w.Write([]byte(`{"contacts":[{"id":1,"email":"a@example.com","modifiedAt":"2026-01-01T00:00:00.000+00:00"},{"id":2,"email":"b@example.com","modifiedAt":"2026-01-02T00:00:00.000+00:00"}],"count":3}`))
		case "2":
			_, _ = w.Write([]byte(`{"contacts":[{"id":3,"email":"c@example.com","modifiedAt":"2026-01-03T00:00:00.000+00:00"}],"count":3}`))
		default:
			t.Errorf("unexpected offset=%q", r.URL.Query().Get("offset"))
			_, _ = w.Write([]byte(`{"contacts":[],"count":3}`))
		}
	}))
	defer srv.Close()

	c := brevo.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"api_key": "xkeysib-test-123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "contacts", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAPIKey != "xkeysib-test-123" {
		t.Fatalf("api-key header = %q, want xkeysib-test-123", sawAPIKey)
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

// TestReadCampaignsPath confirms the connector reads from a non-default JSON
// records path (campaigns[]) for a second stream.
func TestReadCampaignsPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/emailCampaigns" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"campaigns":[{"id":10,"name":"Welcome","status":"sent"}],"count":1}`))
	}))
	defer srv.Close()

	c := brevo.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "50"},
		Secrets: map[string]string{"api_key": "xkeysib-test-123"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "emailCampaigns", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 1 || got[0]["name"] != "Welcome" {
		t.Fatalf("campaigns = %+v, want one Welcome campaign", got)
	}
}

// TestFixtureMode confirms credential-free fixture reads work for conformance.
func TestFixtureMode(t *testing.T) {
	c := brevo.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "contacts", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read(fixture): %v", err)
	}
	if len(got) == 0 {
		t.Fatal("fixture read produced no records")
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("fixture record missing id: %+v", rec)
		}
	}
}

func TestCatalogStreams(t *testing.T) {
	c := brevo.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("streams = %d, want at least 3", len(cat.Streams))
	}
	want := map[string]bool{"contacts": false, "emailCampaigns": false, "contacts_lists": false, "senders": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("catalog missing expected stream %q", name)
		}
	}
}

func TestRegistryResolves(t *testing.T) {
	_ = brevo.New() // ensure init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("brevo"); !ok {
		t.Fatal("registry did not resolve brevo (self-registration)")
	}
	caps := brevo.New().Metadata().Capabilities
	if !caps.Read || !caps.Check || !caps.Catalog {
		t.Fatalf("capabilities = %+v, want Check && Catalog && Read", caps)
	}
}
