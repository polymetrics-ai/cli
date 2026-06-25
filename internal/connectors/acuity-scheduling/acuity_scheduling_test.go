package acuityscheduling_test

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics/internal/connectors"
	acuityscheduling "polymetrics/internal/connectors/acuity-scheduling"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Acuity
// Scheduling connector: HTTP Basic auth (User ID as username, API key as
// password), page-number pagination over a root JSON array, and record mapping.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var sawPaths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		sawPaths = append(sawPaths, r.URL.Path)
		if r.URL.Path != "/appointments" {
			http.NotFound(w, r)
			return
		}
		// Acuity returns a JSON array at the root. The connector advances
		// `page` until it sees a short (or empty) page.
		switch r.URL.Query().Get("page") {
		case "1":
			_, _ = w.Write([]byte(`[{"id":101,"firstName":"Ada","lastName":"Lovelace","email":"ada@example.com","datetime":"2026-01-01T09:00:00-0800","type":"Intro"},{"id":102,"firstName":"Grace","lastName":"Hopper","email":"grace@example.com","datetime":"2026-01-02T09:00:00-0800","type":"Intro"}]`))
		case "2":
			_, _ = w.Write([]byte(`[{"id":103,"firstName":"Katherine","lastName":"Johnson","email":"katherine@example.com","datetime":"2026-01-03T09:00:00-0800","type":"Intro"}]`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
			_, _ = w.Write([]byte(`[]`))
		}
	}))
	defer srv.Close()

	c := acuityscheduling.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "username": "12345", "page_size": "2"},
		Secrets: map[string]string{"password": "apikey_secret"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "appointments", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("12345:apikey_secret"))
	if sawAuth != wantAuth {
		t.Fatalf("Authorization = %q, want %q", sawAuth, wantAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	if len(sawPaths) != 2 {
		t.Fatalf("requests = %d (%v), want 2 pages", len(sawPaths), sawPaths)
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["email"] == nil {
			t.Fatalf("record missing id/email: %+v", rec)
		}
	}
	if got[0]["first_name"] != "Ada" {
		t.Fatalf("first record first_name = %v, want Ada (mapper flattens firstName->first_name)", got[0]["first_name"])
	}
}

// TestReadSinglePageStream confirms list endpoints that return a single root
// array (clients, appointment-types, calendars, forms) read in one request.
func TestReadSinglePageStream(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/clients" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`[{"firstName":"Ada","lastName":"Lovelace","email":"ada@example.com","phone":"555"}]`))
	}))
	defer srv.Close()

	c := acuityscheduling.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "username": "12345"},
		Secrets: map[string]string{"password": "apikey_secret"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "clients", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("records = %d, want 1", len(got))
	}
	if got[0]["email"] != "ada@example.com" {
		t.Fatalf("email = %v, want ada@example.com", got[0]["email"])
	}
}

// TestFixtureModeNoNetwork confirms fixture mode emits deterministic records
// with no network access so conformance runs credential-free.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := acuityscheduling.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "appointments", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read(fixture): %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("fixture records = %d, want 2", len(got))
	}
	if got[0]["id"] == nil {
		t.Fatalf("fixture record missing id: %+v", got[0])
	}
	// Check must also short-circuit in fixture mode.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

// TestCatalogStreams confirms the catalog exposes the core streams with primary
// keys.
func TestCatalogStreams(t *testing.T) {
	c := acuityscheduling.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"appointments": false, "clients": false, "appointment-types": false, "calendars": false, "forms": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
	}
	for name, found := range want {
		if !found {
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

// TestRegisteredReadOnly confirms self-registration via the registry and the
// read-only capability profile.
func TestRegisteredReadOnly(t *testing.T) {
	_ = acuityscheduling.New() // ensure init ran
	c := acuityscheduling.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write false (read-only source)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("acuity-scheduling"); !ok {
		t.Fatal("registry did not resolve acuity-scheduling (self-registration)")
	}
}
