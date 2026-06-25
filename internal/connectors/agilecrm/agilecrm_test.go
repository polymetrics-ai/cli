package agilecrm_test

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/agilecrm"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the AgileCRM
// connector: HTTP Basic auth (email:api_key), AgileCRM cursor pagination where
// the next cursor is read off the LAST record of a top-level JSON array and
// supplied back as ?cursor=..., and record mapping. Red until
// internal/connectors/agilecrm exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/contacts" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("cursor") {
		case "":
			// First page: last record carries the cursor to the next page.
			_, _ = w.Write([]byte(`[{"id":1,"type":"PERSON","created_time":1700000000},{"id":2,"type":"PERSON","created_time":1700000100,"cursor":"abc"}]`))
		case "abc":
			// Second page: no cursor on the last record => end of list.
			_, _ = w.Write([]byte(`[{"id":3,"type":"PERSON","created_time":1700000200}]`))
		default:
			t.Errorf("unexpected cursor=%q", r.URL.Query().Get("cursor"))
			_, _ = w.Write([]byte(`[]`))
		}
	}))
	defer srv.Close()

	c := agilecrm.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "email": "user@example.com"},
		Secrets: map[string]string{"api_key": "secret_key_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "contacts", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("user@example.com:secret_key_123"))
	if sawAuth != wantAuth {
		t.Fatalf("Authorization = %q, want %q", sawAuth, wantAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("record missing id: %+v", rec)
		}
	}
}

// TestReadFixtureMode confirms the credential-free deterministic path used by
// the conformance harness emits records without any network call.
func TestReadFixtureMode(t *testing.T) {
	c := agilecrm.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "deals", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read(fixture): %v", err)
	}
	if len(got) == 0 {
		t.Fatal("fixture mode emitted no records")
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("fixture record missing id: %+v", rec)
		}
	}
}

// TestCheckFixtureMode confirms Check short-circuits without a network call.
func TestCheckFixtureMode(t *testing.T) {
	c := agilecrm.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

// TestCatalogStreams confirms the published catalog exposes the core streams
// with primary keys.
func TestCatalogStreams(t *testing.T) {
	c := agilecrm.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "agilecrm" {
		t.Fatalf("catalog connector = %q, want agilecrm", cat.Connector)
	}
	want := map[string]bool{"contacts": false, "deals": false, "tasks": false, "milestone": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q missing primary key", s.Name)
		}
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("catalog missing core stream %q", name)
		}
	}
}

// TestRegistryResolution confirms self-registration resolves via the registry.
func TestRegistryResolution(t *testing.T) {
	_ = agilecrm.New() // ensure init ran
	r := connectors.NewRegistry()
	got, ok := r.Get("agilecrm")
	if !ok {
		t.Fatal("registry did not resolve agilecrm (self-registration)")
	}
	caps := got.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
}
