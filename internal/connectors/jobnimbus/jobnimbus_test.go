package jobnimbus_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/jobnimbus"
)

// TestReadPaginatesAndAuthenticates is the red-first test: Bearer auth via
// api_key, JobNimbus offset pagination using the `from` parameter across two
// pages of the contacts stream (records under "results"), and jnid record
// mapping.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/contacts" {
			http.NotFound(w, r)
			return
		}
		from := r.URL.Query().Get("from")
		switch from {
		case "", "0":
			// A full page (page_size records) signals there may be more.
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"count":3,"results":[` +
				`{"jnid":"c_1","display_name":"Ada","date_updated":1700000000},` +
				`{"jnid":"c_2","display_name":"Grace","date_updated":1700000100}` +
				`]}`))
		case "2":
			// A short page (fewer than page_size) ends pagination.
			_, _ = w.Write([]byte(`{"count":3,"results":[` +
				`{"jnid":"c_3","display_name":"Katherine","date_updated":1700000200}` +
				`]}`))
		default:
			t.Errorf("unexpected from=%q", from)
			_, _ = w.Write([]byte(`{"results":[]}`))
		}
	}))
	defer srv.Close()

	c := jobnimbus.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"api_key": "jn_test_key"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "contacts", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer jn_test_key" {
		t.Fatalf("Authorization = %q, want Bearer jn_test_key", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["jnid"] == nil {
			t.Fatalf("record missing jnid: %+v", rec)
		}
	}
}

// TestReadActivitiesUsesActivityFieldPath confirms the per-stream record
// selector: activities are nested under "activity", not "results".
func TestReadActivitiesUsesActivityFieldPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/activities" {
			http.NotFound(w, r)
			return
		}
		if r.URL.Query().Get("from") == "" || r.URL.Query().Get("from") == "0" {
			_, _ = w.Write([]byte(`{"activity":[{"jnid":"a_1","note":"called"}]}`))
			return
		}
		_, _ = w.Write([]byte(`{"activity":[]}`))
	}))
	defer srv.Close()

	c := jobnimbus.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "1000"},
		Secrets: map[string]string{"api_key": "jn_test_key"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "activities", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 1 || got[0]["jnid"] != "a_1" {
		t.Fatalf("activities records = %+v, want one record jnid=a_1", got)
	}
}

// TestFixtureModeNeedsNoNetwork verifies credential-free conformance: fixture
// mode emits deterministic records without any HTTP call.
func TestFixtureModeNeedsNoNetwork(t *testing.T) {
	c := jobnimbus.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"contacts", "jobs", "tasks", "activities", "files"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture Read(%s) emitted no records", stream)
		}
		for _, rec := range got {
			if rec["jnid"] == nil {
				t.Fatalf("fixture %s record missing jnid: %+v", stream, rec)
			}
		}
	}
	// Check should also short-circuit in fixture mode without creds.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

func TestCheckRequiresAPIKey(t *testing.T) {
	c := jobnimbus.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{}}
	if err := c.Check(context.Background(), cfg); err == nil {
		t.Fatal("Check without api_key should fail")
	}
}

func TestRejectsUnknownStream(t *testing.T) {
	c := jobnimbus.New()
	cfg := connectors.RuntimeConfig{Secrets: map[string]string{"api_key": "k"}}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "nope", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read of unknown stream should fail")
	}
}

func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := jobnimbus.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example.com"},
		Secrets: map[string]string{"api_key": "k"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "contacts", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read with non-http base_url should fail (SSRF guard)")
	}
}

func TestCatalogListsCoreStreams(t *testing.T) {
	c := jobnimbus.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"jobs": false, "contacts": false, "tasks": false, "activities": false, "files": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) != 1 || s.PrimaryKey[0] != "jnid" {
			t.Fatalf("stream %s primary key = %v, want [jnid]", s.Name, s.PrimaryKey)
		}
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

func TestRegisteredReadOnly(t *testing.T) {
	_ = jobnimbus.New() // ensure init ran
	c := jobnimbus.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("jobnimbus"); !ok {
		t.Fatal("registry did not resolve jobnimbus (self-registration)")
	}
}
