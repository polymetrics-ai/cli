package notion_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/notion"
)

// TestSearchPaginatesAndAuthenticates is the red-first test: it asserts the
// Bearer token, the required Notion-Version header, POST /search body-cursor
// pagination over results[], and record mapping for the databases stream.
func TestSearchPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth, sawVersion string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		sawVersion = r.Header.Get("Notion-Version")
		if r.URL.Path != "/search" || r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		var body struct {
			StartCursor string `json:"start_cursor"`
			Filter      struct {
				Property string `json:"property"`
				Value    string `json:"value"`
			} `json:"filter"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body.Filter.Property != "object" || body.Filter.Value != "database" {
			t.Errorf("filter = %+v, want object=database", body.Filter)
		}
		switch body.StartCursor {
		case "":
			_, _ = w.Write([]byte(`{"object":"list","results":[{"object":"database","id":"db_1","created_time":"2026-01-01T00:00:00.000Z","last_edited_time":"2026-01-02T00:00:00.000Z"},{"object":"database","id":"db_2","created_time":"2026-01-03T00:00:00.000Z","last_edited_time":"2026-01-04T00:00:00.000Z"}],"next_cursor":"cur_2","has_more":true}`))
		case "cur_2":
			_, _ = w.Write([]byte(`{"object":"list","results":[{"object":"database","id":"db_3","created_time":"2026-01-05T00:00:00.000Z","last_edited_time":"2026-01-06T00:00:00.000Z"}],"next_cursor":null,"has_more":false}`))
		default:
			t.Errorf("unexpected start_cursor=%q", body.StartCursor)
			_, _ = w.Write([]byte(`{"object":"list","results":[],"has_more":false}`))
		}
	}))
	defer srv.Close()

	c := notion.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"credentials.access_token": "secret_abc"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "databases", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer secret_abc" {
		t.Fatalf("Authorization = %q, want Bearer secret_abc", sawAuth)
	}
	if sawVersion != "2022-06-28" {
		t.Fatalf("Notion-Version = %q, want 2022-06-28", sawVersion)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["last_edited_time"] == nil {
			t.Fatalf("record missing id/last_edited_time: %+v", rec)
		}
		if rec["object"] != "database" {
			t.Fatalf("record object = %v, want database", rec["object"])
		}
	}
}

// TestUsersPaginatesViaQueryCursor verifies the GET /users stream authenticates
// and paginates via the start_cursor query parameter.
func TestUsersPaginatesViaQueryCursor(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/users" || r.Method != http.MethodGet {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("start_cursor") {
		case "":
			_, _ = w.Write([]byte(`{"object":"list","results":[{"object":"user","id":"u_1","name":"Ada","type":"person"}],"next_cursor":"u_cur","has_more":true}`))
		case "u_cur":
			_, _ = w.Write([]byte(`{"object":"list","results":[{"object":"user","id":"u_2","name":"Grace","type":"bot"}],"next_cursor":null,"has_more":false}`))
		default:
			t.Errorf("unexpected start_cursor=%q", r.URL.Query().Get("start_cursor"))
			_, _ = w.Write([]byte(`{"object":"list","results":[],"has_more":false}`))
		}
	}))
	defer srv.Close()

	c := notion.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"token": "secret_xyz"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "users", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2 (2 pages)", len(got))
	}
	if got[0]["id"] != "u_1" || got[1]["id"] != "u_2" {
		t.Fatalf("user ids = %v, %v, want u_1, u_2", got[0]["id"], got[1]["id"])
	}
}

// TestFixtureModeNoNetwork ensures the fixture path emits deterministic records
// without any credentials or network access.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := notion.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"databases", "pages", "users"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("Read(%s) fixture: %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture stream %s produced no records", stream)
		}
		for _, rec := range got {
			if rec["id"] == nil {
				t.Fatalf("fixture %s record missing id: %+v", stream, rec)
			}
		}
	}
}

// TestCheckFixtureNoSecret confirms Check short-circuits in fixture mode.
func TestCheckFixtureNoSecret(t *testing.T) {
	c := notion.New()
	if err := c.Check(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
}

// TestCheckRequiresSecret confirms a non-fixture Check rejects a missing token.
func TestCheckRequiresSecret(t *testing.T) {
	c := notion.New()
	if err := c.Check(context.Background(), connectors.RuntimeConfig{}); err == nil {
		t.Fatal("Check without token should fail")
	}
}

// TestBaseURLRejectsBadScheme guards the SSRF validation on base_url overrides.
func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := notion.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "file:///etc/passwd"},
		Secrets: map[string]string{"token": "secret_xyz"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "users", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read with file:// base_url should be rejected")
	}
}

// TestReadOnlyCapabilities asserts Notion is read-only and registry-resolvable.
func TestReadOnlyCapabilities(t *testing.T) {
	c := notion.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Check || !caps.Catalog {
		t.Fatalf("capabilities = %+v, want Read && Check && Catalog", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("notion"); !ok {
		t.Fatal("registry did not resolve notion (self-registration)")
	}
}

// TestCatalogStreams asserts the published catalog exposes the core streams.
func TestCatalogStreams(t *testing.T) {
	c := notion.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"databases": false, "pages": false, "users": false}
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
