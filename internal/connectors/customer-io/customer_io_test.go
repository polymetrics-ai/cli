package customerio_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	customerio "polymetrics.ai/internal/connectors/customer-io"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Customer.io
// connector: Bearer auth with the App API Key, cursor pagination over the
// {newsletters:[...], next:"..."} body (next token passed back as ?start=), and
// record mapping. Red until internal/connectors/customer-io exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/newsletters" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("start") {
		case "":
			_, _ = w.Write([]byte(`{"newsletters":[{"id":1,"name":"Welcome","updated":1700000000},{"id":2,"name":"Promo","updated":1700000100}],"next":"page2"}`))
		case "page2":
			_, _ = w.Write([]byte(`{"newsletters":[{"id":3,"name":"Digest","updated":1700000200}],"next":null}`))
		default:
			t.Errorf("unexpected start=%q", r.URL.Query().Get("start"))
			_, _ = w.Write([]byte(`{"newsletters":[],"next":null}`))
		}
	}))
	defer srv.Close()

	c := customerio.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"app_api_key": "appkey_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "newsletters", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer appkey_123" {
		t.Fatalf("Authorization = %q, want Bearer appkey_123", sawAuth)
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

// TestReadCampaignsRootSelector verifies the campaigns stream selects records at
// the "campaigns" field path and maps the expected fields.
func TestReadCampaignsRootSelector(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/campaigns" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"campaigns":[{"id":42,"name":"Onboarding","type":"triggered","active":true,"updated":1700000000}]}`))
	}))
	defer srv.Close()

	c := customerio.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"app_api_key": "appkey_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "campaigns", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("records = %d, want 1", len(got))
	}
	rec := got[0]
	if rec["name"] != "Onboarding" || rec["type"] != "triggered" {
		t.Fatalf("unexpected campaign record: %+v", rec)
	}
}

// TestFixtureModeNoNetwork ensures fixture mode emits deterministic records with
// no network access so conformance can run without credentials.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := customerio.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	for _, stream := range []string{"campaigns", "newsletters", "segments", "broadcasts"} {
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
}

// TestCheckFixtureMode verifies Check short-circuits in fixture mode.
func TestCheckFixtureMode(t *testing.T) {
	c := customerio.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

// TestRegistryResolves confirms self-registration via init() and read-only
// capabilities.
func TestRegistryResolves(t *testing.T) {
	_ = customerio.New() // ensure init ran
	c := customerio.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("customer-io"); !ok {
		t.Fatal("registry did not resolve customer-io (self-registration)")
	}
}
