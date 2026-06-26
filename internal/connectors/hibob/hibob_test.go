package hibob_test

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/hibob"
)

// TestReadProfilesAuthenticatesAndPaginates is the red-first test for the HiBob
// connector. HiBob authenticates with HTTP Basic (service-user-id:token) and the
// /profiles endpoint returns {"employees":[...]}. The connector pages a
// follow-on read via the includeHumanReadable-independent paging param the test
// server expects, exercising auth header, two pages, and record mapping.
func TestReadProfilesAuthenticatesAndPaginates(t *testing.T) {
	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("svc-user:tok-123"))
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/profiles" {
			http.NotFound(w, r)
			return
		}
		// Page over employees using the connector's offset param.
		switch r.URL.Query().Get("offset") {
		case "", "0":
			_, _ = w.Write([]byte(`{"employees":[
				{"id":"1","email":"a@ex.com","displayName":"A","firstName":"Al","surname":"Pha"},
				{"id":"2","email":"b@ex.com","displayName":"B","firstName":"Be","surname":"Ta"}
			]}`))
		case "2":
			_, _ = w.Write([]byte(`{"employees":[
				{"id":"3","email":"c@ex.com","displayName":"C","firstName":"Ce","surname":"Ta"}
			]}`))
		default:
			_, _ = w.Write([]byte(`{"employees":[]}`))
		}
	}))
	defer srv.Close()

	c := hibob.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "username": "svc-user", "page_size": "2"},
		Secrets: map[string]string{"password": "tok-123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "profiles", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != wantAuth {
		t.Fatalf("Authorization = %q, want %q", sawAuth, wantAuth)
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

// TestFixtureModeNoNetwork exercises the credential-free fixture path used by the
// conformance harness.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := hibob.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var n int
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "profiles", Config: cfg}, func(rec connectors.Record) error {
		if rec["id"] == nil {
			t.Fatalf("fixture record missing id: %+v", rec)
		}
		n++
		return nil
	})
	if err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if n == 0 {
		t.Fatal("fixture mode emitted no records")
	}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

// TestRegisteredReadOnly confirms self-registration and read-only capabilities.
func TestRegisteredReadOnly(t *testing.T) {
	c := hibob.New()
	if c.Name() != "hibob" {
		t.Fatalf("Name = %q, want hibob", c.Name())
	}
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read/Catalog/Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("hibob"); !ok {
		t.Fatal("registry did not resolve hibob (self-registration)")
	}
}

// TestCatalogStreams confirms the published catalog includes profiles.
func TestCatalogStreams(t *testing.T) {
	c := hibob.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("streams = %d, want >= 3", len(cat.Streams))
	}
	var found bool
	for _, s := range cat.Streams {
		if s.Name == "profiles" {
			found = true
			if len(s.PrimaryKey) == 0 {
				t.Fatal("profiles stream missing primary key")
			}
		}
	}
	if !found {
		t.Fatal("catalog missing profiles stream")
	}
}
