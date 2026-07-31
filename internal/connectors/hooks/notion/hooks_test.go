package notion_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/engine"
	notion "polymetrics.ai/internal/connectors/hooks/notion"
)

// newTestRuntime builds a minimal *engine.Runtime pointed at srv with a
// Bundle whose Schemas[streamName] projects exactly props (mirrors
// hooks/sentry/hooks_test.go's helper).
func newTestRuntime(srv *httptest.Server, streamName string, props []string) *engine.Runtime {
	sch := schemaWithProperties(props)
	b := &engine.Bundle{
		Name:    "notion",
		Schemas: map[string]*engine.StreamSchema{streamName: sch},
	}
	return &engine.Runtime{
		Requester: &connsdk.Requester{BaseURL: srv.URL},
		Bundle:    b,
	}
}

func schemaWithProperties(props []string) *engine.StreamSchema {
	doc := map[string]any{
		"$schema":    "http://json-schema.org/draft-07/schema#",
		"type":       "object",
		"properties": map[string]any{},
	}
	propsMap := doc["properties"].(map[string]any)
	for _, p := range props {
		propsMap[p] = map[string]any{"type": []string{"string", "boolean", "object", "null"}}
	}
	sch, err := engine.CompileSchema(mustMarshal(doc))
	if err != nil {
		panic(err)
	}
	return &engine.StreamSchema{Schema: sch}
}

func mustMarshal(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

// TestReadStream_DatabasesSearchBodyCursorPagination is the red-first test:
// the hook must POST /search with the object=database filter and
// start_cursor in the JSON body, following the cursor across two pages.
func TestReadStream_DatabasesSearchBodyCursorPagination(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		w.Header().Set("Content-Type", "application/json")
		switch body.StartCursor {
		case "":
			_, _ = w.Write([]byte(`{"object":"list","results":[{"object":"database","id":"db_1","last_edited_time":"2026-01-02T00:00:00.000Z"},{"object":"database","id":"db_2","last_edited_time":"2026-01-04T00:00:00.000Z"}],"next_cursor":"cur_2","has_more":true}`))
		case "cur_2":
			_, _ = w.Write([]byte(`{"object":"list","results":[{"object":"database","id":"db_3","last_edited_time":"2026-01-06T00:00:00.000Z"}],"next_cursor":null,"has_more":false}`))
		default:
			t.Errorf("unexpected start_cursor=%q", body.StartCursor)
			_, _ = w.Write([]byte(`{"object":"list","results":[],"has_more":false}`))
		}
	}))
	defer srv.Close()

	h := notion.Hooks{}
	rt := newTestRuntime(srv, "databases", []string{"id", "object", "last_edited_time"})

	var got []connectors.Record
	handled, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "databases", Method: http.MethodPost, Path: "/search", Body: map[string]any{"filter": map[string]any{"property": "object", "value": "database"}}, Records: engine.RecordsSpec{Path: "results"}, Projection: "passthrough"}, connectors.ReadRequest{Stream: "databases"}, rt, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if !handled {
		t.Fatal("ReadStream returned handled=false, want true")
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["object"] != "database" {
			t.Fatalf("record missing id or wrong object: %+v", rec)
		}
	}
}

// TestReadStream_PagesUsesPageFilter verifies the pages stream sends the
// object=page filter (distinguishing it from databases on the same
// physical /search endpoint).
func TestReadStream_PagesUsesPageFilter(t *testing.T) {
	var gotValue string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Filter struct {
				Value string `json:"value"`
			} `json:"filter"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		gotValue = body.Filter.Value
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"object":"list","results":[{"object":"page","id":"pg_1"}],"has_more":false}`))
	}))
	defer srv.Close()

	h := notion.Hooks{}
	rt := newTestRuntime(srv, "pages", []string{"id", "object"})

	var got []connectors.Record
	handled, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "pages", Method: http.MethodPost, Path: "/search", Body: map[string]any{"filter": map[string]any{"property": "object", "value": "page"}}, Records: engine.RecordsSpec{Path: "results"}, Projection: "passthrough"}, connectors.ReadRequest{Stream: "pages"}, rt, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if !handled {
		t.Fatal("ReadStream returned handled=false, want true")
	}
	if gotValue != "page" {
		t.Fatalf("filter.value = %q, want page", gotValue)
	}
	if len(got) != 1 {
		t.Fatalf("records = %d, want 1", len(got))
	}
}

// TestReadStream_UsersQueryCursorPagination verifies GET /users paginates
// via the start_cursor query parameter (no request body at all).
func TestReadStream_UsersQueryCursorPagination(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/users" || r.Method != http.MethodGet {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Query().Get("start_cursor") {
		case "":
			_, _ = w.Write([]byte(`{"object":"list","results":[{"object":"user","id":"u_1","name":"Ada"}],"next_cursor":"u_cur","has_more":true}`))
		case "u_cur":
			_, _ = w.Write([]byte(`{"object":"list","results":[{"object":"user","id":"u_2","name":"Grace"}],"next_cursor":null,"has_more":false}`))
		default:
			t.Errorf("unexpected start_cursor=%q", r.URL.Query().Get("start_cursor"))
			_, _ = w.Write([]byte(`{"object":"list","results":[],"has_more":false}`))
		}
	}))
	defer srv.Close()

	h := notion.Hooks{}
	rt := newTestRuntime(srv, "users", []string{"id", "object", "name"})

	var got []connectors.Record
	handled, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "users", Method: http.MethodGet, Path: "/users", Records: engine.RecordsSpec{Path: "results"}, Projection: "passthrough"}, connectors.ReadRequest{Stream: "users"}, rt, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if !handled {
		t.Fatal("ReadStream returned handled=false, want true")
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2 (2 pages)", len(got))
	}
	if got[0]["id"] != "u_1" || got[1]["id"] != "u_2" {
		t.Fatalf("user ids = %v, %v, want u_1, u_2", got[0]["id"], got[1]["id"])
	}
}

// TestReadStream_MaxPagesCapsRequests verifies the config-driven max_pages
// cap stops pagination early even when the server still offers a next
// cursor.
func TestReadStream_MaxPagesCapsRequests(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"object":"list","results":[{"object":"database","id":"db_x"}],"next_cursor":"more","has_more":true}`))
	}))
	defer srv.Close()

	h := notion.Hooks{}
	rt := newTestRuntime(srv, "databases", []string{"id", "object"})
	cfg := connectors.RuntimeConfig{Config: map[string]string{"max_pages": "1"}}

	var got []connectors.Record
	handled, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "databases", Method: http.MethodPost, Path: "/search", Body: map[string]any{"filter": map[string]any{"property": "object", "value": "database"}}, Records: engine.RecordsSpec{Path: "results"}, Projection: "passthrough"}, connectors.ReadRequest{Stream: "databases", Config: cfg}, rt, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if !handled {
		t.Fatal("ReadStream returned handled=false, want true")
	}
	if calls != 1 {
		t.Fatalf("calls = %d, want 1 (max_pages=1 cap)", calls)
	}
	if len(got) != 1 {
		t.Fatalf("records = %d, want 1", len(got))
	}
}

func TestReadStream_MeetingNotesUsesLimitPaginationField(t *testing.T) {
	var sawLimit, sawPageSize bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/blocks/meeting_notes/query" || r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		_, sawLimit = body["limit"]
		_, sawPageSize = body["page_size"]
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"object":"list","results":[{"object":"block","id":"blk_1"}],"has_more":false}`))
	}))
	defer srv.Close()

	h := notion.Hooks{}
	rt := newTestRuntime(srv, "meeting_notes_query", []string{"id", "object"})
	cfg := connectors.RuntimeConfig{Config: map[string]string{"page_size": "7"}}

	var got []connectors.Record
	_, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "meeting_notes_query", Method: http.MethodPost, Path: "/blocks/meeting_notes/query", Body: map[string]any{"limit": "{{ config.page_size }}"}, Records: engine.RecordsSpec{Path: "results"}, Projection: "passthrough"}, connectors.ReadRequest{Stream: "meeting_notes_query", Config: cfg}, rt, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if !sawLimit {
		t.Fatal("request body missing limit")
	}
	if sawPageSize {
		t.Fatal("request body included page_size, want only limit for meeting notes query")
	}
	if len(got) != 1 {
		t.Fatalf("records = %d, want 1", len(got))
	}
}

func TestReadStream_RepeatedCursorFailsClosed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"object":"list","results":[{"object":"database","id":"db_x"}],"next_cursor":"same","has_more":true}`))
	}))
	defer srv.Close()

	h := notion.Hooks{}
	rt := newTestRuntime(srv, "databases", []string{"id", "object"})
	_, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "databases", Method: http.MethodPost, Path: "/search", Body: map[string]any{"filter": map[string]any{"property": "object", "value": "database"}}, Records: engine.RecordsSpec{Path: "results"}, Projection: "passthrough"}, connectors.ReadRequest{Stream: "databases"}, rt, func(connectors.Record) error {
		return nil
	})
	if err == nil {
		t.Fatal("ReadStream returned nil error for repeated next_cursor, want fail-closed error")
	}
}

// TestReadStream_UnknownStreamFallsBack verifies an unrecognized stream
// name returns handled=false rather than erroring, keeping the declarative
// path an honest fallback.
func TestReadStream_UnknownStreamFallsBack(t *testing.T) {
	h := notion.Hooks{}
	rt := &engine.Runtime{Bundle: &engine.Bundle{Name: "notion"}}
	handled, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "nonexistent", Method: http.MethodGet, Path: "/nonexistent", Records: engine.RecordsSpec{Path: "results"}, Projection: "passthrough"}, connectors.ReadRequest{Stream: "nonexistent"}, rt, func(connectors.Record) error { return nil })
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if handled {
		t.Fatal("ReadStream returned handled=true for an unknown stream, want false")
	}
}
