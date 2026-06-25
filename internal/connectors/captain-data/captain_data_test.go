package captaindata_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics/internal/connectors"
	captaindata "polymetrics/internal/connectors/captain-data"
)

// TestReadWorkflowsAuthenticates is the red-first test for the Captain Data
// connector: x-api-key auth, x-project-id header, and root-array record mapping
// on a top-level stream. Red until internal/connectors/captain-data exists.
func TestReadWorkflowsAuthenticates(t *testing.T) {
	var sawAPIKey, sawProject string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAPIKey = r.Header.Get("X-API-Key")
		sawProject = r.Header.Get("X-Project-Id")
		if r.URL.Path != "/workflows" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`[{"uid":"wf_1","name":"Scrape","status":"active"},{"uid":"wf_2","name":"Enrich","status":"paused"}]`))
	}))
	defer srv.Close()

	c := captaindata.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "project_uid": "proj_42"},
		Secrets: map[string]string{"api_key": "key_abc"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "workflows", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAPIKey != "key_abc" {
		t.Fatalf("X-API-Key = %q, want key_abc", sawAPIKey)
	}
	if sawProject != "proj_42" {
		t.Fatalf("X-Project-Id = %q, want proj_42", sawProject)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2", len(got))
	}
	if got[0]["uid"] != "wf_1" || got[1]["uid"] != "wf_2" {
		t.Fatalf("unexpected records: %+v", got)
	}
}

// TestReadJobResultsPaginates exercises the cursor pagination on job_results:
// the API returns {results:[...], paging:{next, have_next_page}} and the next
// page is fetched with the cursor query param until have_next_page is false.
func TestReadJobResultsPaginates(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/jobs/job_1/results" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("cursor") {
		case "":
			_, _ = w.Write([]byte(`{"results":[{"uid":"res_1"},{"uid":"res_2"}],"paging":{"next":"cur_2","have_next_page":true}}`))
		case "cur_2":
			_, _ = w.Write([]byte(`{"results":[{"uid":"res_3"}],"paging":{"next":null,"have_next_page":false}}`))
		default:
			t.Errorf("unexpected cursor=%q", r.URL.Query().Get("cursor"))
			_, _ = w.Write([]byte(`{"results":[],"paging":{"have_next_page":false}}`))
		}
	}))
	defer srv.Close()

	c := captaindata.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "project_uid": "proj_42", "job_uid": "job_1"},
		Secrets: map[string]string{"api_key": "key_abc"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "job_results", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["uid"] == nil {
			t.Fatalf("record missing uid: %+v", rec)
		}
	}
}

// TestFixtureModeReadsWithoutNetwork confirms fixture mode emits deterministic
// records with no network access, so conformance can run credential-free.
func TestFixtureModeReadsWithoutNetwork(t *testing.T) {
	c := captaindata.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"workspace", "workflows", "jobs", "job_results"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("Read(%s) fixture: %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture stream %s emitted no records", stream)
		}
		if got[0]["uid"] == nil {
			t.Fatalf("fixture stream %s record missing uid: %+v", stream, got[0])
		}
	}
}

func TestCheckFixtureModeNoNetwork(t *testing.T) {
	c := captaindata.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

func TestCatalogStreams(t *testing.T) {
	c := captaindata.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"workspace": false, "workflows": false, "jobs": false, "job_results": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %s missing primary key", s.Name)
		}
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

func TestRegisteredReadOnly(t *testing.T) {
	_ = captaindata.New() // ensure init ran
	c := captaindata.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog {
		t.Fatalf("capabilities = %+v, want Read && Catalog", caps)
	}
	if caps.Write {
		t.Fatalf("captain-data should be read-only, got Write=true")
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("captain-data"); !ok {
		t.Fatal("registry did not resolve captain-data (self-registration)")
	}
}
