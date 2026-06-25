package bamboohr_test

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	bamboohr "polymetrics.ai/internal/connectors/bamboo-hr"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the BambooHR
// connector. It asserts HTTP Basic auth (api_key as username, "x" as password),
// the JSON Accept header, page-based pagination across two pages of the
// employees directory, and record mapping. Red until the bamboohr package exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var sawAccept string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		sawAccept = r.Header.Get("Accept")
		if r.URL.Path != "/employees/directory" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page") {
		case "", "1":
			// A full page (page size 2) signals there may be more.
			_, _ = w.Write([]byte(`{"fields":[{"id":"firstName"}],"employees":[{"id":"1","displayName":"Ada Lovelace","firstName":"Ada","lastName":"Lovelace","workEmail":"ada@example.com"},{"id":"2","displayName":"Grace Hopper","firstName":"Grace","lastName":"Hopper","workEmail":"grace@example.com"}]}`))
		case "2":
			// A short page (fewer than the page size) ends pagination.
			_, _ = w.Write([]byte(`{"fields":[{"id":"firstName"}],"employees":[{"id":"3","displayName":"Katherine Johnson","firstName":"Katherine","lastName":"Johnson","workEmail":"katherine@example.com"}]}`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
			_, _ = w.Write([]byte(`{"employees":[]}`))
		}
	}))
	defer srv.Close()

	c := bamboohr.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "subdomain": "acme", "page_size": "2"},
		Secrets: map[string]string{"api_key": "key_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "employees", Config: cfg}, func(rec connectors.Record) error {
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
	if !strings.Contains(sawAccept, "application/json") {
		t.Fatalf("Accept = %q, want it to contain application/json", sawAccept)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["display_name"] == nil {
			t.Fatalf("record missing id/display_name: %+v", rec)
		}
	}
	if got[0]["work_email"] != "ada@example.com" {
		t.Fatalf("record[0].work_email = %v, want ada@example.com", got[0]["work_email"])
	}
}

// TestReadTimeOffTypesNestedPath confirms a stream whose records live under a
// nested JSON key (timeOffTypes) is extracted and mapped correctly.
func TestReadTimeOffTypesNestedPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/meta/time_off/types" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"timeOffTypes":[{"id":"78","name":"Vacation","units":"days","icon":"time-off"},{"id":"79","name":"Sick","units":"hours","icon":"medical"}]}`))
	}))
	defer srv.Close()

	c := bamboohr.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "subdomain": "acme"},
		Secrets: map[string]string{"api_key": "key_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "time_off_types", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2", len(got))
	}
	if got[0]["id"] != "78" || got[0]["name"] != "Vacation" || got[0]["units"] != "days" {
		t.Fatalf("unexpected first time_off_type record: %+v", got[0])
	}
}

// TestReadMetaFieldsTopLevelArray confirms a stream whose response is a top-level
// JSON array (meta/fields) is read correctly.
func TestReadMetaFieldsTopLevelArray(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/meta/fields" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`[{"id":1,"name":"First Name","type":"text","alias":"firstName"},{"id":2,"name":"Last Name","type":"text","alias":"lastName"}]`))
	}))
	defer srv.Close()

	c := bamboohr.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "subdomain": "acme"},
		Secrets: map[string]string{"api_key": "key_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "meta_fields", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2", len(got))
	}
	if got[0]["alias"] != "firstName" || got[0]["name"] != "First Name" {
		t.Fatalf("unexpected first meta_field record: %+v", got[0])
	}
}

// TestFixtureModeNoNetwork confirms fixture mode emits deterministic records with
// no HTTP server configured, so conformance can run without credentials.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := bamboohr.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{"mode": "fixture", "subdomain": "acme"},
	}
	// Each stream's primary-key field must be populated by fixture mode.
	pkField := map[string]string{
		"employees":      "id",
		"meta_fields":    "id",
		"meta_lists":     "field_id",
		"time_off_types": "id",
	}
	for stream, pk := range pkField {
		var got []connectors.Record
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
		if got[0][pk] == nil || got[0][pk] == "" {
			t.Fatalf("Read(%s) fixture record missing primary key %q: %+v", stream, pk, got[0])
		}
	}
}

// TestCheckFixtureMode confirms Check short-circuits in fixture mode and that a
// missing api_key in live mode is an error.
func TestCheckFixtureMode(t *testing.T) {
	c := bamboohr.New()
	if err := c.Check(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture", "subdomain": "acme"}}); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
	err := c.Check(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"subdomain": "acme"}})
	if err == nil {
		t.Fatal("Check without api_key should error")
	}
}

// TestCatalogStreams confirms the published catalog exposes the core streams with
// primary keys.
func TestCatalogStreams(t *testing.T) {
	c := bamboohr.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"employees": false, "meta_fields": false, "meta_lists": false, "time_off_types": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q has no primary key", s.Name)
		}
	}
	for name, found := range want {
		if !found {
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

// TestRegisteredReadOnly confirms self-registration resolves via the registry and
// that the connector advertises read (not write) capability.
func TestRegisteredReadOnly(t *testing.T) {
	_ = bamboohr.New() // ensure init ran
	caps := bamboohr.New().Metadata().Capabilities
	if !caps.Read || !caps.Check || !caps.Catalog {
		t.Fatalf("capabilities = %+v, want Read && Check && Catalog", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only HR source)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("bamboo-hr"); !ok {
		t.Fatal("registry did not resolve bamboo-hr (self-registration)")
	}
}

// TestBaseURLSSRFValidation confirms an invalid base_url scheme is rejected.
func TestBaseURLSSRFValidation(t *testing.T) {
	c := bamboohr.New()
	err := c.Read(context.Background(), connectors.ReadRequest{
		Stream: "employees",
		Config: connectors.RuntimeConfig{
			Config:  map[string]string{"base_url": "file:///etc/passwd", "subdomain": "acme"},
			Secrets: map[string]string{"api_key": "key_123"},
		},
	}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read with file:// base_url should be rejected")
	}
}
