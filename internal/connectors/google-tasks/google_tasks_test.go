package googletasks_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	googletasks "polymetrics.ai/internal/connectors/google-tasks"
)

// TestReadTaskListsPaginatesAndAuthenticates is the red-first test: it asserts
// Bearer auth from the api_key secret, nextPageToken/pageToken pagination across
// two pages of the users/@me/lists endpoint, and record mapping.
func TestReadTaskListsPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/users/@me/lists" {
			http.NotFound(w, r)
			return
		}
		calls++
		switch r.URL.Query().Get("pageToken") {
		case "":
			_, _ = w.Write([]byte(`{"kind":"tasks#taskLists","items":[{"kind":"tasks#taskList","id":"L1","title":"Inbox","updated":"2026-01-01T00:00:00.000Z"},{"kind":"tasks#taskList","id":"L2","title":"Work","updated":"2026-01-02T00:00:00.000Z"}],"nextPageToken":"PAGE2"}`))
		case "PAGE2":
			_, _ = w.Write([]byte(`{"kind":"tasks#taskLists","items":[{"kind":"tasks#taskList","id":"L3","title":"Personal","updated":"2026-01-03T00:00:00.000Z"}]}`))
		default:
			t.Errorf("unexpected pageToken=%q", r.URL.Query().Get("pageToken"))
		}
	}))
	defer srv.Close()

	c := googletasks.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "tok_abc123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "tasklists", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer tok_abc123" {
		t.Fatalf("Authorization = %q, want Bearer tok_abc123", sawAuth)
	}
	if calls != 2 {
		t.Fatalf("server calls = %d, want 2 (two pages)", calls)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 across 2 pages", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["title"] == nil {
			t.Fatalf("record missing id/title: %+v", rec)
		}
	}
}

// TestReadTasksIteratesEveryTaskList asserts the tasks stream first lists the
// task lists, then harvests tasks from each list endpoint, mapping records.
func TestReadTasksIteratesEveryTaskList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/users/@me/lists":
			_, _ = w.Write([]byte(`{"items":[{"id":"L1","title":"Inbox"},{"id":"L2","title":"Work"}]}`))
		case "/lists/L1/tasks":
			_, _ = w.Write([]byte(`{"items":[{"id":"T1","title":"Task one","status":"needsAction","updated":"2026-02-01T00:00:00.000Z"}]}`))
		case "/lists/L2/tasks":
			_, _ = w.Write([]byte(`{"items":[{"id":"T2","title":"Task two","status":"completed","updated":"2026-02-02T00:00:00.000Z"}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := googletasks.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "tok_abc123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "tasks", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2 (one per task list)", len(got))
	}
	ids := map[string]bool{}
	for _, rec := range got {
		ids[rec["id"].(string)] = true
		if rec["tasklist_id"] == nil {
			t.Fatalf("task record missing tasklist_id: %+v", rec)
		}
	}
	if !ids["T1"] || !ids["T2"] {
		t.Fatalf("missing expected task ids, got %+v", ids)
	}
}

// TestFixtureModeReadsWithoutNetwork confirms the credential-free fixture path
// emits deterministic records for conformance.
func TestFixtureModeReadsWithoutNetwork(t *testing.T) {
	c := googletasks.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"tasklists", "tasks"} {
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
		for _, rec := range got {
			if rec["id"] == nil {
				t.Fatalf("fixture %s record missing id: %+v", stream, rec)
			}
		}
	}
}

func TestCheckFixtureModeNoNetwork(t *testing.T) {
	c := googletasks.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

func TestCatalogStreams(t *testing.T) {
	c := googletasks.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 2 {
		t.Fatalf("expected at least 2 streams, got %d", len(cat.Streams))
	}
	names := map[string]bool{}
	for _, s := range cat.Streams {
		names[s.Name] = true
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %s missing primary key", s.Name)
		}
	}
	if !names["tasklists"] || !names["tasks"] {
		t.Fatalf("missing core streams, got %+v", names)
	}
}

func TestRegisteredReadOnly(t *testing.T) {
	c := googletasks.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write false (read-only API)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("google-tasks"); !ok {
		t.Fatal("registry did not resolve google-tasks (self-registration)")
	}
}
