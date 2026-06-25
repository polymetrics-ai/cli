package intercom_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics/internal/connectors"
	"polymetrics/internal/connectors/intercom"
)

// TestReadPaginatesAndAuthenticates is the red-first test: Bearer auth, Intercom
// pages.next.starting_after cursor pagination over data[], and record mapping.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/contacts" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Query().Get("starting_after") {
		case "":
			_, _ = w.Write([]byte(`{"type":"list","data":[
				{"id":"c1","type":"contact","created_at":1700000000,"updated_at":1700000001,"email":"a@example.com"},
				{"id":"c2","type":"contact","created_at":1700000002,"updated_at":1700000003,"email":"b@example.com"}
			],"pages":{"type":"pages","next":{"starting_after":"cursorAAA"}},"total_count":3}`))
		case "cursorAAA":
			_, _ = w.Write([]byte(`{"type":"list","data":[
				{"id":"c3","type":"contact","created_at":1700000004,"updated_at":1700000005,"email":"c@example.com"}
			],"pages":{"type":"pages"},"total_count":3}`))
		default:
			t.Errorf("unexpected starting_after=%q", r.URL.Query().Get("starting_after"))
			_, _ = w.Write([]byte(`{"type":"list","data":[],"pages":{"type":"pages"}}`))
		}
	}))
	defer srv.Close()

	c := intercom.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"access_token": "tok_abc123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "contacts", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer tok_abc123" {
		t.Fatalf("Authorization = %q, want Bearer tok_abc123", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["created_at"] == nil {
			t.Fatalf("record missing id/created_at: %+v", rec)
		}
	}
	if got[0]["email"] != "a@example.com" {
		t.Fatalf("first record email = %v, want a@example.com", got[0]["email"])
	}
}

// TestFixtureModeRead exercises the credential-free fixture path used by the
// conformance harness.
func TestFixtureModeRead(t *testing.T) {
	c := intercom.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "companies", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("fixture records = %d, want 2", len(got))
	}
	if got[0]["id"] == nil {
		t.Fatalf("fixture record missing id: %+v", got[0])
	}
}

func TestCheckFixtureMode(t *testing.T) {
	c := intercom.New()
	if err := c.Check(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

func TestCatalogStreams(t *testing.T) {
	c := intercom.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "intercom" {
		t.Fatalf("catalog connector = %q, want intercom", cat.Connector)
	}
	want := map[string]bool{"contacts": false, "companies": false, "conversations": false, "admins": false, "tags": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q missing primary key", s.Name)
		}
	}
	for name, found := range want {
		if !found {
			t.Fatalf("catalog missing expected stream %q", name)
		}
	}
}

func TestUnknownStream(t *testing.T) {
	c := intercom.New()
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "does_not_exist", Config: connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read of unknown stream should error")
	}
}

func TestBaseURLValidation(t *testing.T) {
	c := intercom.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example.com"},
		Secrets: map[string]string{"access_token": "tok"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "contacts", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("non-http base_url scheme should be rejected")
	}
}

func TestRegisteredReadOnly(t *testing.T) {
	c := intercom.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("intercom is read-only, want Write=false, got %+v", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("intercom"); !ok {
		t.Fatal("registry did not resolve intercom (self-registration)")
	}
}
