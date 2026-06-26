package lemlist_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/lemlist"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the lemlist
// connector: the API key is injected as the access_token query parameter, the
// campaigns stream pages via offset/limit over a root-level JSON array, and each
// object is mapped to a record keyed by _id. Red until internal/connectors/lemlist
// exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawToken string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawToken = r.URL.Query().Get("access_token")
		if r.URL.Path != "/campaigns" {
			http.NotFound(w, r)
			return
		}
		// page size is 2 in the test config; offset advances by 2.
		switch r.URL.Query().Get("offset") {
		case "", "0":
			_, _ = w.Write([]byte(`[{"_id":"cmp_1","name":"Alpha"},{"_id":"cmp_2","name":"Beta"}]`))
		case "2":
			_, _ = w.Write([]byte(`[{"_id":"cmp_3","name":"Gamma"}]`))
		default:
			t.Errorf("unexpected offset=%q", r.URL.Query().Get("offset"))
			_, _ = w.Write([]byte(`[]`))
		}
	}))
	defer srv.Close()

	c := lemlist.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"api_key": "key_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "campaigns", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawToken != "key_test_123" {
		t.Fatalf("access_token = %q, want key_test_123", sawToken)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["_id"] == nil {
			t.Fatalf("record missing _id: %+v", rec)
		}
	}
	if got[0]["_id"] != "cmp_1" || got[2]["_id"] != "cmp_3" {
		t.Fatalf("unexpected record order/ids: %+v", got)
	}
}

// TestReadTeamSingleObject verifies the team stream, whose API response is a
// single root-level JSON object (not an array), maps to exactly one record.
func TestReadTeamSingleObject(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/team" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"_id":"team_1","name":"Acme"}`))
	}))
	defer srv.Close()

	c := lemlist.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "key_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "team", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read team: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("team records = %d, want 1", len(got))
	}
	if got[0]["_id"] != "team_1" {
		t.Fatalf("team _id = %v, want team_1", got[0]["_id"])
	}
}

// TestFixtureModeNoNetwork verifies fixture mode emits deterministic records
// without any network access (base_url points nowhere reachable).
func TestFixtureModeNoNetwork(t *testing.T) {
	c := lemlist.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{"mode": "fixture", "base_url": "http://127.0.0.1:0"},
	}
	for _, stream := range []string{"team", "campaigns", "activities", "unsubscribes"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read %s: %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture %s emitted no records", stream)
		}
		for _, rec := range got {
			if rec["_id"] == nil {
				t.Fatalf("fixture %s record missing _id: %+v", stream, rec)
			}
		}
	}
}

func TestCheckFixtureMode(t *testing.T) {
	c := lemlist.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

func TestCatalogStreams(t *testing.T) {
	c := lemlist.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"team": true, "campaigns": true, "activities": true, "unsubscribes": true}
	for _, s := range cat.Streams {
		delete(want, s.Name)
		if len(s.PrimaryKey) == 0 || s.PrimaryKey[0] != "_id" {
			t.Fatalf("stream %q primary key = %v, want [_id]", s.Name, s.PrimaryKey)
		}
	}
	if len(want) != 0 {
		t.Fatalf("catalog missing streams: %v", want)
	}
}

func TestRegisteredReadOnly(t *testing.T) {
	_ = lemlist.New() // ensure init ran
	c := lemlist.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("lemlist is read-only; Write capability should be false")
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("lemlist"); !ok {
		t.Fatal("registry did not resolve lemlist (self-registration)")
	}
}
