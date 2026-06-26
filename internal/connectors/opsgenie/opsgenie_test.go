package opsgenie_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/opsgenie"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Opsgenie
// connector: it asserts the GenieKey Authorization header, paging.next URL
// pagination across two pages of data[], and record mapping. Red until
// internal/connectors/opsgenie exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var srvURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/alerts" {
			http.NotFound(w, r)
			return
		}
		// First page advertises a next URL pointing at offset=2; second page
		// omits paging.next to terminate the loop.
		switch r.URL.Query().Get("offset") {
		case "", "0":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"data":[{"id":"a1","tinyId":"1","message":"disk full","status":"open","priority":"P1","createdAt":"2026-01-01T00:00:00Z"},{"id":"a2","tinyId":"2","message":"cpu high","status":"open","priority":"P2","createdAt":"2026-01-01T01:00:00Z"}],"paging":{"next":"` + srvURL + `/alerts?offset=2&limit=2"},"totalCount":3}`))
		case "2":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"data":[{"id":"a3","tinyId":"3","message":"oom","status":"closed","priority":"P3","createdAt":"2026-01-01T02:00:00Z"}],"paging":{},"totalCount":3}`))
		default:
			t.Errorf("unexpected offset=%q", r.URL.Query().Get("offset"))
			_, _ = w.Write([]byte(`{"data":[],"paging":{}}`))
		}
	}))
	defer srv.Close()
	srvURL = srv.URL

	c := opsgenie.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"api_token": "tok_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "alerts", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "GenieKey tok_123" {
		t.Fatalf("Authorization = %q, want GenieKey tok_123", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["created_at"] == nil {
			t.Fatalf("record missing id/created_at: %+v", rec)
		}
	}
	if got[0]["priority"] != "P1" || got[2]["status"] != "closed" {
		t.Fatalf("unexpected mapping: %+v ... %+v", got[0], got[2])
	}
}

func TestFixtureModeReadsWithoutNetwork(t *testing.T) {
	c := opsgenie.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	for _, stream := range []string{"alerts", "incidents", "users", "teams", "services"} {
		got = got[:0]
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
		if got[0]["id"] == nil {
			t.Fatalf("fixture %s record missing id: %+v", stream, got[0])
		}
	}
}

func TestCheckFixtureMode(t *testing.T) {
	c := opsgenie.New()
	if err := c.Check(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
}

func TestCatalogStreams(t *testing.T) {
	c := opsgenie.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "opsgenie" {
		t.Fatalf("catalog connector = %q, want opsgenie", cat.Connector)
	}
	want := map[string]bool{"alerts": false, "incidents": false, "users": false, "teams": false, "services": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %s has no primary key", s.Name)
		}
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := opsgenie.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example.com"},
		Secrets: map[string]string{"api_token": "tok"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "alerts", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "base_url") {
		t.Fatalf("expected base_url scheme rejection, got %v", err)
	}
}

func TestRegisteredReadOnly(t *testing.T) {
	_ = opsgenie.New() // ensure init ran
	caps := opsgenie.New().Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("opsgenie is read-only, Write should be false")
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("opsgenie"); !ok {
		t.Fatal("registry did not resolve opsgenie (self-registration)")
	}
}
