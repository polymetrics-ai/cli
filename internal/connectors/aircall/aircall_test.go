package aircall_test

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics/internal/connectors"
	"polymetrics/internal/connectors/aircall"
)

// TestReadPaginatesAndAuthenticates is the red-first test: Basic auth
// (api_id:api_token), Aircall meta.next_page_link pagination over two pages, and
// record mapping for the calls stream.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/calls" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page") {
		case "", "1":
			// First page is full; advertise a next page link.
			w.Write([]byte(`{"meta":{"current_page":1,"per_page":2,"next_page_link":"` + srv.URL + `/calls?page=2&per_page=2"},"calls":[{"id":1,"direction":"inbound","status":"done","started_at":1700000000},{"id":2,"direction":"outbound","status":"done","started_at":1700000100}]}`))
		case "2":
			w.Write([]byte(`{"meta":{"current_page":2,"per_page":2,"next_page_link":null},"calls":[{"id":3,"direction":"inbound","status":"done","started_at":1700000200}]}`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
			w.Write([]byte(`{"meta":{"next_page_link":null},"calls":[]}`))
		}
	}))
	defer srv.Close()

	c := aircall.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "per_page": "2"},
		Secrets: map[string]string{"api_id": "id_123", "api_token": "tok_456"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "calls", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("id_123:tok_456"))
	if sawAuth != wantAuth {
		t.Fatalf("Authorization = %q, want %q", sawAuth, wantAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["started_at"] == nil {
			t.Fatalf("record missing id/started_at: %+v", rec)
		}
	}
	if got[0]["direction"] != "inbound" {
		t.Fatalf("first record direction = %v, want inbound", got[0]["direction"])
	}
}

// TestReadFixtureMode confirms the credential-free fixture path emits
// deterministic records without any network call.
func TestReadFixtureMode(t *testing.T) {
	c := aircall.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "users", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read fixture: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("fixture records = %d, want 2", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("fixture record missing id: %+v", rec)
		}
	}
}

// TestCheckFixtureMode confirms fixture mode short-circuits Check with no creds.
func TestCheckFixtureMode(t *testing.T) {
	c := aircall.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check fixture = %v, want nil", err)
	}
}

// TestCheckRequiresSecrets confirms a non-fixture Check rejects missing creds.
func TestCheckRequiresSecrets(t *testing.T) {
	c := aircall.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{}}
	if err := c.Check(context.Background(), cfg); err == nil {
		t.Fatal("Check without secrets should error")
	}
}

func TestCatalogStreams(t *testing.T) {
	c := aircall.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "aircall" {
		t.Fatalf("catalog connector = %q, want aircall", cat.Connector)
	}
	want := map[string]bool{"calls": true, "users": true, "contacts": true, "numbers": true, "teams": true}
	for _, s := range cat.Streams {
		delete(want, s.Name)
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q missing primary key", s.Name)
		}
	}
	if len(want) != 0 {
		t.Fatalf("catalog missing streams: %v", want)
	}
}

func TestRegistryResolution(t *testing.T) {
	_ = aircall.New() // ensure init() ran
	r := connectors.NewRegistry()
	conn, ok := r.Get("aircall")
	if !ok {
		t.Fatal("registry did not resolve aircall (self-registration)")
	}
	caps := conn.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || caps.Write {
		t.Fatalf("capabilities = %+v, want Read && Catalog && !Write", caps)
	}
}

func TestBaseURLSSRFValidation(t *testing.T) {
	c := aircall.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil"},
		Secrets: map[string]string{"api_id": "id", "api_token": "tok"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "calls", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "base_url") {
		t.Fatalf("Read with bad base_url scheme = %v, want base_url error", err)
	}
}
