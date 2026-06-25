package metabase_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/metabase"
)

// TestReadAuthenticatesAndMapsRecords is the red-first test for the Metabase
// connector: it asserts the X-Metabase-Session auth header is sent, that a
// top-level JSON array response is mapped into records, and that record mapping
// projects the expected fields. Metabase list endpoints are not paginated, so
// the "two pages" requirement is satisfied here by the multi-endpoint catalog
// drive plus the explicit pagination test below for the data-wrapped shape.
func TestReadAuthenticatesAndMapsRecords(t *testing.T) {
	var sawSession string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawSession = r.Header.Get("X-Metabase-Session")
		if r.URL.Path != "/card" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`[
			{"id":1,"name":"Revenue","description":"r","collection_id":2,"database_id":3,"archived":false,"creator_id":7,"updated_at":"2026-01-02T00:00:00Z"},
			{"id":2,"name":"Signups","description":"s","collection_id":2,"database_id":3,"archived":false,"creator_id":8,"updated_at":"2026-01-03T00:00:00Z"}
		]`))
	}))
	defer srv.Close()

	c := metabase.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"instance_api_url": srv.URL},
		Secrets: map[string]string{"session_token": "sess_abc"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "cards", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawSession != "sess_abc" {
		t.Fatalf("X-Metabase-Session = %q, want sess_abc", sawSession)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2", len(got))
	}
	if got[0]["id"] == nil || got[0]["name"] != "Revenue" {
		t.Fatalf("record mapping wrong: %+v", got[0])
	}
}

// TestReadHandlesDataWrappedPaginationShape covers the {"data":[...],"total":N}
// shape some Metabase endpoints (e.g. /api/user) return, and drives two pages
// via an offset query so the pagination loop is exercised end to end.
func TestReadHandlesDataWrappedPagination(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/user" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("offset") {
		case "", "0":
			// total greater than page so a second page is requested.
			_, _ = w.Write([]byte(`{"data":[
				{"id":1,"email":"a@x.com","first_name":"A","common_name":"A","is_active":true},
				{"id":2,"email":"b@x.com","first_name":"B","common_name":"B","is_active":true}
			],"total":3,"limit":2,"offset":0}`))
		default:
			_, _ = w.Write([]byte(`{"data":[
				{"id":3,"email":"c@x.com","first_name":"C","common_name":"C","is_active":true}
			],"total":3,"limit":2,"offset":2}`))
		}
	}))
	defer srv.Close()

	c := metabase.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"instance_api_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"session_token": "sess_abc"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "users", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 across two pages", len(got))
	}
	if got[2]["email"] != "c@x.com" {
		t.Fatalf("second page record wrong: %+v", got[2])
	}
}

// TestFixtureModeNeedsNoNetwork ensures conformance can run without creds.
func TestFixtureModeNeedsNoNetwork(t *testing.T) {
	c := metabase.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
	for _, stream := range []string{"cards", "dashboards", "collections", "databases", "users"} {
		var n int
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(connectors.Record) error {
			n++
			return nil
		})
		if err != nil {
			t.Fatalf("Read fixture %s: %v", stream, err)
		}
		if n == 0 {
			t.Fatalf("fixture stream %s emitted no records", stream)
		}
	}
}

// TestCatalogStreams asserts the published catalog has the core streams.
func TestCatalogStreams(t *testing.T) {
	c := metabase.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"cards": false, "dashboards": false, "collections": false, "databases": false, "users": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %s missing primary key", s.Name)
		}
	}
	for name, found := range want {
		if !found {
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

// TestRegistryResolves confirms self-registration via the registry.
func TestRegistryResolves(t *testing.T) {
	_ = metabase.New() // ensure init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("metabase"); !ok {
		t.Fatal("registry did not resolve metabase (self-registration)")
	}
}

// TestReadMissingSecretErrors confirms a non-fixture read without creds fails
// fast rather than making an unauthenticated request.
func TestReadMissingSecretErrors(t *testing.T) {
	c := metabase.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"instance_api_url": "https://mb.example.com/api"}}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "cards", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("expected error when no session_token/password is supplied")
	}
}
