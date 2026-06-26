package onfleet_test

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/onfleet"
)

// TestReadTasksPaginatesAndAuthenticates is the red-first test: HTTP Basic auth
// with the API key as username and empty password, Onfleet lastId pagination
// over {lastId, tasks:[]} across two pages, and record mapping.
func TestReadTasksPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/tasks/all" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("lastId") {
		case "":
			_, _ = w.Write([]byte(`{"lastId":"task_2","tasks":[{"id":"task_1","state":0},{"id":"task_2","state":1}]}`))
		case "task_2":
			_, _ = w.Write([]byte(`{"lastId":"","tasks":[{"id":"task_3","state":2}]}`))
		default:
			t.Errorf("unexpected lastId=%q", r.URL.Query().Get("lastId"))
			_, _ = w.Write([]byte(`{"lastId":"","tasks":[]}`))
		}
	}))
	defer srv.Close()

	c := onfleet.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "key_abc", "password": ""},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "tasks", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("key_abc:"))
	if sawAuth != wantAuth {
		t.Fatalf("Authorization = %q, want %q", sawAuth, wantAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("record missing id: %+v", rec)
		}
	}
}

// TestReadWorkersTopLevelArray covers the non-paginated streams that return a
// top-level JSON array (workers, teams, hubs, administrators).
func TestReadWorkersTopLevelArray(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/workers" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`[{"id":"w_1","name":"Alice"},{"id":"w_2","name":"Bob"}]`))
	}))
	defer srv.Close()

	c := onfleet.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "key_abc"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "workers", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2", len(got))
	}
	if got[0]["name"] != "Alice" {
		t.Fatalf("first worker name = %v, want Alice", got[0]["name"])
	}
}

// TestFixtureModeNoNetwork ensures fixture mode emits deterministic records with
// no network access, so conformance can run credential-free.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := onfleet.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	for _, stream := range []string{"tasks", "workers", "teams", "hubs", "administrators"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("Read(%s) fixture: %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("Read(%s) fixture emitted 0 records", stream)
		}
		for _, rec := range got {
			if rec["id"] == nil {
				t.Fatalf("fixture %s record missing id: %+v", stream, rec)
			}
		}
	}

	// Check must also short-circuit in fixture mode.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
}

func TestCatalogStreams(t *testing.T) {
	c := onfleet.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "onfleet" {
		t.Fatalf("catalog connector = %q, want onfleet", cat.Connector)
	}
	want := map[string]bool{"tasks": false, "workers": false, "teams": false, "hubs": false, "administrators": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q has no primary key", s.Name)
		}
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := onfleet.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://onfleet.com/api/v2"},
		Secrets: map[string]string{"api_key": "key_abc"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "workers", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "base_url") {
		t.Fatalf("Read with ftp base_url err = %v, want base_url scheme rejection", err)
	}
}

func TestRegisteredReadOnly(t *testing.T) {
	_ = onfleet.New() // ensure init ran
	c := onfleet.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog {
		t.Fatalf("capabilities = %+v, want Read && Catalog", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("onfleet"); !ok {
		t.Fatal("registry did not resolve onfleet (self-registration)")
	}
}
