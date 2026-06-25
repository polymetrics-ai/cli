package airtable_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/airtable"
)

// TestReadBasesPaginatesAndAuthenticates is the red-first test: Bearer auth,
// Airtable body-offset pagination over bases[], and record mapping across two
// pages.
func TestReadBasesPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/meta/bases" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("offset") {
		case "":
			_, _ = w.Write([]byte(`{"bases":[{"id":"app1","name":"Base One","permissionLevel":"create"},{"id":"app2","name":"Base Two","permissionLevel":"edit"}],"offset":"page2tok"}`))
		case "page2tok":
			_, _ = w.Write([]byte(`{"bases":[{"id":"app3","name":"Base Three","permissionLevel":"read"}]}`))
		default:
			t.Errorf("unexpected offset=%q", r.URL.Query().Get("offset"))
			_, _ = w.Write([]byte(`{"bases":[]}`))
		}
	}))
	defer srv.Close()

	c := airtable.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"credentials.api_key": "patABC123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "bases", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer patABC123" {
		t.Fatalf("Authorization = %q, want Bearer patABC123", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	if got[0]["id"] != "app1" || got[0]["name"] != "Base One" {
		t.Fatalf("first record mapped wrong: %+v", got[0])
	}
	if got[2]["id"] != "app3" {
		t.Fatalf("last record id = %v, want app3", got[2]["id"])
	}
}

// TestReadRecordsPaginates exercises the records stream which targets
// /{baseId}/{tableId}, extracts records[], flattens fields[], and follows the
// body offset.
func TestReadRecordsPaginates(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/appXYZ/tblABC" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("offset") {
		case "":
			_, _ = w.Write([]byte(`{"records":[{"id":"rec1","createdTime":"2026-01-01T00:00:00.000Z","fields":{"Name":"Ada","Score":42}}],"offset":"next"}`))
		case "next":
			_, _ = w.Write([]byte(`{"records":[{"id":"rec2","createdTime":"2026-01-02T00:00:00.000Z","fields":{"Name":"Grace"}}]}`))
		default:
			t.Errorf("unexpected offset=%q", r.URL.Query().Get("offset"))
			_, _ = w.Write([]byte(`{"records":[]}`))
		}
	}))
	defer srv.Close()

	c := airtable.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "base_id": "appXYZ", "table_id": "tblABC"},
		Secrets: map[string]string{"credentials.access_token": "oauthTok"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "records", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2 (2 pages)", len(got))
	}
	if got[0]["id"] != "rec1" {
		t.Fatalf("first record id = %v, want rec1", got[0]["id"])
	}
	// fields[] should be flattened/exposed under "fields".
	fields, ok := got[0]["fields"].(map[string]any)
	if !ok {
		t.Fatalf("record fields not preserved as object: %+v", got[0]["fields"])
	}
	if fields["Name"] != "Ada" {
		t.Fatalf("fields.Name = %v, want Ada", fields["Name"])
	}
}

func TestFixtureModeNoNetwork(t *testing.T) {
	c := airtable.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	for _, stream := range []string{"bases", "tables", "records"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture Read(%s) produced no records", stream)
		}
		if got[0]["id"] == nil {
			t.Fatalf("fixture Read(%s) record missing id: %+v", stream, got[0])
		}
	}
}

func TestCheckFixtureMode(t *testing.T) {
	c := airtable.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

func TestRegisteredReadOnly(t *testing.T) {
	_ = airtable.New() // ensure init ran
	caps := airtable.New().Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("airtable should be read-only, got Write=true")
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("airtable"); !ok {
		t.Fatal("registry did not resolve airtable (self-registration)")
	}
}
