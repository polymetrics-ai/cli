package chargify_test

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics/internal/connectors"
	"polymetrics/internal/connectors/chargify"
)

// TestReadPaginatesAndAuthenticates is the red-first test: Basic auth (api_key as
// username, "x" as password), page/per_page pagination across two pages, and
// mapping of Chargify's wrapped {"customer":{...}} list elements.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/customers.json" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page") {
		case "", "1":
			_, _ = w.Write([]byte(`[{"customer":{"id":1,"email":"a@example.com","updated_at":"2026-01-01T00:00:00Z"}},{"customer":{"id":2,"email":"b@example.com","updated_at":"2026-01-02T00:00:00Z"}}]`))
		case "2":
			_, _ = w.Write([]byte(`[{"customer":{"id":3,"email":"c@example.com","updated_at":"2026-01-03T00:00:00Z"}}]`))
		case "3":
			_, _ = w.Write([]byte(`[]`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
			_, _ = w.Write([]byte(`[]`))
		}
	}))
	defer srv.Close()

	c := chargify.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"api_key": "key_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "customers", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("key_123:x"))
	if sawAuth != wantAuth {
		t.Fatalf("Authorization = %q, want %q", sawAuth, wantAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["email"] == nil {
			t.Fatalf("record missing unwrapped fields: %+v", rec)
		}
	}
}

// TestReadUsernamePasswordAuth verifies that an explicit username/password pair
// overrides the api_key:x default Basic credentials.
func TestReadUsernamePasswordAuth(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		_, _ = w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	c := chargify.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "username": "user1"},
		Secrets: map[string]string{"password": "pass1"},
	}
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "customers", Config: cfg}, func(connectors.Record) error { return nil }); err != nil {
		t.Fatalf("Read: %v", err)
	}
	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("user1:pass1"))
	if sawAuth != wantAuth {
		t.Fatalf("Authorization = %q, want %q", sawAuth, wantAuth)
	}
}

// TestFixtureModeReadNoNetwork exercises the credential-free fixture path used by
// the conformance harness.
func TestFixtureModeReadNoNetwork(t *testing.T) {
	c := chargify.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	for _, stream := range []string{"customers", "subscriptions", "products", "coupons", "transactions"} {
		got = got[:0]
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("Read(%s) fixture: %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("Read(%s) fixture emitted no records", stream)
		}
		if got[0]["id"] == nil {
			t.Fatalf("Read(%s) fixture record missing id: %+v", stream, got[0])
		}
	}
}

func TestCheckFixtureMode(t *testing.T) {
	c := chargify.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

func TestCatalogStreams(t *testing.T) {
	c := chargify.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "chargify" {
		t.Fatalf("Catalog.Connector = %q, want chargify", cat.Connector)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("Catalog streams = %d, want >= 3", len(cat.Streams))
	}
	for _, s := range cat.Streams {
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q missing primary key", s.Name)
		}
	}
}

func TestBaseURLValidation(t *testing.T) {
	c := chargify.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil"},
		Secrets: map[string]string{"api_key": "k"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "customers", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "base_url") {
		t.Fatalf("Read with bad base_url err = %v, want base_url validation error", err)
	}
}

func TestRegistryResolvesChargify(t *testing.T) {
	_ = chargify.New() // ensure init ran
	r := connectors.NewRegistry()
	got, ok := r.Get("chargify")
	if !ok {
		t.Fatal("registry did not resolve chargify (self-registration)")
	}
	if got.Name() != "chargify" {
		t.Fatalf("resolved connector Name = %q, want chargify", got.Name())
	}
	caps := got.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read/Catalog/Check", caps)
	}
}
