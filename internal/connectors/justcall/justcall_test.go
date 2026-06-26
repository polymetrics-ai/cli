package justcall_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/justcall"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the JustCall
// connector: the Authorization header carries the raw api_key_2 value, the
// page-increment paginator (page starts at 0, per_page=size) walks two pages of
// data[], and records are mapped. Red until internal/connectors/justcall exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var sawPages []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/v2.1/calls" {
			http.NotFound(w, r)
			return
		}
		page := r.URL.Query().Get("page")
		sawPages = append(sawPages, page)
		switch page {
		case "0":
			if r.URL.Query().Get("per_page") != "2" {
				t.Errorf("per_page = %q, want 2", r.URL.Query().Get("per_page"))
			}
			_, _ = w.Write([]byte(`{"data":[{"id":1,"call_date":"2026-01-01","agent_name":"A"},{"id":2,"call_date":"2026-01-02","agent_name":"B"}]}`))
		case "1":
			_, _ = w.Write([]byte(`{"data":[{"id":3,"call_date":"2026-01-03","agent_name":"C"}]}`))
		default:
			t.Errorf("unexpected page=%q", page)
			_, _ = w.Write([]byte(`{"data":[]}`))
		}
	}))
	defer srv.Close()

	c := justcall.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"api_key_2": "key:secret"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "calls", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "key:secret" {
		t.Fatalf("Authorization = %q, want raw api_key_2 value key:secret", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	if len(sawPages) != 2 || sawPages[0] != "0" || sawPages[1] != "1" {
		t.Fatalf("pages requested = %v, want [0 1]", sawPages)
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["call_date"] == nil {
			t.Fatalf("record missing id/call_date: %+v", rec)
		}
	}
}

// TestReadContactsUsesPost confirms the POST-list streams hit the right path
// with the POST method and still paginate over data[].
func TestReadContactsUsesPost(t *testing.T) {
	var sawMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/contacts/list" {
			http.NotFound(w, r)
			return
		}
		sawMethod = r.Method
		if r.URL.Query().Get("page") == "0" {
			_, _ = w.Write([]byte(`{"data":[{"id":10,"firstname":"Ada","email":"ada@example.com"}]}`))
			return
		}
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	defer srv.Close()

	c := justcall.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "100"},
		Secrets: map[string]string{"api_key_2": "key:secret"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "contacts", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read contacts: %v", err)
	}
	if sawMethod != http.MethodPost {
		t.Fatalf("contacts method = %q, want POST", sawMethod)
	}
	if len(got) != 1 || got[0]["firstname"] != "Ada" {
		t.Fatalf("contacts records = %+v, want one Ada record", got)
	}
}

// TestFixtureModeNoNetwork confirms fixture mode emits deterministic records
// without any network access (credential-free conformance).
func TestFixtureModeNoNetwork(t *testing.T) {
	c := justcall.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "calls", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("fixture records = %d, want 2", len(got))
	}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

// TestCatalogStreams confirms the published catalog exposes the core streams.
func TestCatalogStreams(t *testing.T) {
	c := justcall.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"users": false, "calls": false, "sms": false, "contacts": false, "phone_numbers": false}
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

// TestRegisteredReadOnly confirms self-registration and read-only capabilities.
func TestRegisteredReadOnly(t *testing.T) {
	c := justcall.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || caps.Write {
		t.Fatalf("capabilities = %+v, want Read && !Write", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("justcall"); !ok {
		t.Fatal("registry did not resolve justcall (self-registration)")
	}
}
