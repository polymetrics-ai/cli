package monday

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/engine"
)

// graphqlRequest mirrors the minimal GraphQL POST body shape used throughout
// this test file.
type graphqlRequest struct {
	Query string `json:"query"`
}

func newRuntime(baseURL string) *engine.Runtime {
	return &engine.Runtime{Requester: &connsdk.Requester{BaseURL: baseURL}}
}

// --- registration ---

func TestInit_RegistersHooks(t *testing.T) {
	h := engine.HooksFor("monday")
	if h == nil {
		t.Fatal("engine.HooksFor(\"monday\") = nil, want a registered hook set (init() must call engine.RegisterHooks)")
	}
	if h.ConnectorName() != "monday" {
		t.Fatalf("ConnectorName() = %q, want monday", h.ConnectorName())
	}
	if _, ok := h.(engine.StreamHook); !ok {
		t.Fatal("registered hooks do not implement StreamHook")
	}
	if _, ok := h.(engine.CheckHook); !ok {
		t.Fatal("registered hooks do not implement CheckHook")
	}
}

// --- ReadStream dispatch ---

func TestReadStream_UnknownStreamFallsBackToDeclarative(t *testing.T) {
	h := Hooks{}
	handled, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "not_a_real_stream"}, connectors.ReadRequest{}, newRuntime("http://example.invalid"), func(connectors.Record) error { return nil })
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if handled {
		t.Fatal("handled = true for an unrecognized stream name, want false (declarative fallback)")
	}
}

func TestReadStream_EmptyStreamNameDefaultsToBoards(t *testing.T) {
	pages := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body graphqlRequest
		_ = json.NewDecoder(r.Body).Decode(&body)
		if !strings.Contains(body.Query, "boards") {
			t.Errorf("query did not target boards: %q", body.Query)
		}
		pages++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"boards":[]}}`))
	}))
	defer srv.Close()

	h := Hooks{}
	var got []connectors.Record
	handled, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: ""}, connectors.ReadRequest{}, newRuntime(srv.URL), func(r connectors.Record) error {
		got = append(got, r)
		return nil
	})
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if !handled {
		t.Fatal("handled = false, want true (empty stream name defaults to boards)")
	}
	if pages == 0 {
		t.Fatal("server received zero requests")
	}
}

// --- boards/users/teams/tags: page-number pagination loop ---

func TestReadPaged_PaginatesUntilShortPage(t *testing.T) {
	var pagesSeen []int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body graphqlRequest
		_ = json.NewDecoder(r.Body).Decode(&body)
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.Contains(body.Query, "page: 1"):
			pagesSeen = append(pagesSeen, 1)
			_, _ = w.Write([]byte(`{"data":{"boards":[{"id":"1","name":"A"},{"id":"2","name":"B"}]}}`))
		case strings.Contains(body.Query, "page: 2"):
			pagesSeen = append(pagesSeen, 2)
			_, _ = w.Write([]byte(`{"data":{"boards":[{"id":"3","name":"C"}]}}`))
		default:
			t.Errorf("unexpected query: %q", body.Query)
			_, _ = w.Write([]byte(`{"data":{"boards":[]}}`))
		}
	}))
	defer srv.Close()

	h := Hooks{}
	var got []connectors.Record
	err := h.readPaged(context.Background(), &connsdk.Requester{BaseURL: srv.URL}, pageStreamSpecs["boards"], 2, 0, func(r connectors.Record) error {
		got = append(got, r)
		return nil
	})
	if err != nil {
		t.Fatalf("readPaged: %v", err)
	}
	if len(pagesSeen) != 2 {
		t.Fatalf("pages served = %v, want exactly [1 2] (short page 2 must stop the loop)", pagesSeen)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3", len(got))
	}
	if got[0]["id"] != "1" || got[0]["name"] != "A" {
		t.Fatalf("record 0 = %+v, want id=1 name=A", got[0])
	}
}

func TestReadPaged_MaxPagesHardStop(t *testing.T) {
	hits := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		w.Header().Set("Content-Type", "application/json")
		// Always return a FULL page so the short-page signal never fires on its
		// own; only the max_pages cap can stop this loop.
		_, _ = w.Write([]byte(`{"data":{"boards":[{"id":"1","name":"A"},{"id":"2","name":"B"}]}}`))
	}))
	defer srv.Close()

	h := Hooks{}
	err := h.readPaged(context.Background(), &connsdk.Requester{BaseURL: srv.URL}, pageStreamSpecs["boards"], 2, 1, func(connectors.Record) error { return nil })
	if err != nil {
		t.Fatalf("readPaged: %v", err)
	}
	if hits != 1 {
		t.Fatalf("server hits = %d, want exactly 1 (max_pages=1 hard stop)", hits)
	}
}

// --- items: cursor-based next_items_page pagination loop ---

func TestReadItems_FollowsCursorUntilNull(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body graphqlRequest
		_ = json.NewDecoder(r.Body).Decode(&body)
		calls++
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.Contains(body.Query, "next_items_page"):
			if !strings.Contains(body.Query, "CUR_2") {
				t.Errorf("next_items_page query missing cursor: %q", body.Query)
			}
			_, _ = w.Write([]byte(`{"data":{"next_items_page":{"cursor":null,"items":[{"id":"i3","name":"Item Three"}]}}}`))
		default:
			_, _ = w.Write([]byte(`{"data":{"boards":[{"id":"1","items_page":{"cursor":"CUR_2","items":[{"id":"i1","name":"Item One","group":{"id":"g1","title":"G1"},"board":{"id":"1","name":"B1"}},{"id":"i2","name":"Item Two"}]}}]}}`))
		}
	}))
	defer srv.Close()

	h := Hooks{}
	var got []connectors.Record
	err := h.readItems(context.Background(), &connsdk.Requester{BaseURL: srv.URL}, 2, 0, func(r connectors.Record) error {
		got = append(got, r)
		return nil
	})
	if err != nil {
		t.Fatalf("readItems: %v", err)
	}
	if calls != 2 {
		t.Fatalf("calls = %d, want 2 (first boards request + one next_items_page cursor request)", calls)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3", len(got))
	}
	if got[0]["group_id"] != "g1" || got[0]["group_title"] != "G1" {
		t.Fatalf("record 0 group hoist = %+v, want group_id=g1 group_title=G1", got[0])
	}
	if got[0]["board_id"] != "1" || got[0]["board_name"] != "B1" {
		t.Fatalf("record 0 board hoist = %+v, want board_id=1 board_name=B1", got[0])
	}
}

func TestReadItems_NoCursorStopsAfterFirstPage(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"boards":[{"id":"1","items_page":{"cursor":null,"items":[{"id":"i1","name":"Item One"}]}}]}}`))
	}))
	defer srv.Close()

	h := Hooks{}
	var got []connectors.Record
	err := h.readItems(context.Background(), &connsdk.Requester{BaseURL: srv.URL}, 2, 0, func(r connectors.Record) error {
		got = append(got, r)
		return nil
	})
	if err != nil {
		t.Fatalf("readItems: %v", err)
	}
	if calls != 1 {
		t.Fatalf("calls = %d, want 1 (null cursor stops immediately)", calls)
	}
	if len(got) != 1 {
		t.Fatalf("records = %d, want 1", len(got))
	}
}

// --- GraphQL error surfacing (monday returns HTTP 200 for query errors) ---

func TestExecute_GraphQLErrorEnvelopeSurfacesAsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"errors":[{"message":"field arguments are invalid"}]}`))
	}))
	defer srv.Close()

	_, err := execute(context.Background(), &connsdk.Requester{BaseURL: srv.URL}, "query { boards { id } }")
	if err == nil {
		t.Fatal("execute did not return an error for a GraphQL errors envelope")
	}
	if !strings.Contains(err.Error(), "field arguments are invalid") {
		t.Fatalf("error = %v, want it to contain the GraphQL error message", err)
	}
}

func TestExecute_LegacyErrorMessageEnvelope(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"error_message":"Not Authenticated"}`))
	}))
	defer srv.Close()

	_, err := execute(context.Background(), &connsdk.Requester{BaseURL: srv.URL}, "query { me { id } }")
	if err == nil {
		t.Fatal("execute did not return an error for a legacy error_message envelope")
	}
	if !strings.Contains(err.Error(), "Not Authenticated") {
		t.Fatalf("error = %v, want it to contain the legacy error message", err)
	}
}

// --- CheckHook ---

func TestCheck_SendsMeQuery(t *testing.T) {
	var sawQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body graphqlRequest
		_ = json.NewDecoder(r.Body).Decode(&body)
		sawQuery = body.Query
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"me":{"id":"1"}}}`))
	}))
	defer srv.Close()

	h := Hooks{}
	handled, err := h.Check(context.Background(), connectors.RuntimeConfig{}, newRuntime(srv.URL))
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if !handled {
		t.Fatal("handled = false, want true")
	}
	if !strings.Contains(sawQuery, "me") || !strings.Contains(sawQuery, "id") {
		t.Fatalf("check query = %q, want a bounded query containing me/id", sawQuery)
	}
}

func TestCheck_GraphQLErrorPropagates(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"errors":[{"message":"Not Authenticated"}]}`))
	}))
	defer srv.Close()

	h := Hooks{}
	handled, err := h.Check(context.Background(), connectors.RuntimeConfig{}, newRuntime(srv.URL))
	if err == nil {
		t.Fatal("Check did not error on a GraphQL error envelope")
	}
	if !handled {
		t.Fatal("handled = false on error path, want true (CheckHook owns the check outcome even on failure)")
	}
}

// --- record mapping helpers ---

func TestStringField_CoercesNumberAndString(t *testing.T) {
	if got := stringField(map[string]any{"id": "abc"}, "id"); got != "abc" {
		t.Fatalf("stringField(string) = %q, want abc", got)
	}
	if got := stringField(map[string]any{"id": json.Number("42")}, "id"); got != "42" {
		t.Fatalf("stringField(json.Number) = %q, want 42", got)
	}
	if got := stringField(map[string]any{"id": nil}, "id"); got != "" {
		t.Fatalf("stringField(nil) = %q, want empty", got)
	}
	if got := stringField(map[string]any{}, "id"); got != "" {
		t.Fatalf("stringField(absent) = %q, want empty", got)
	}
}

func TestBoardRecord_MapsAllFields(t *testing.T) {
	rec := boardRecord(map[string]any{
		"id": "1", "name": "Board", "state": "active", "board_kind": "public",
		"description": "d", "type": "board", "updated_at": "2026-01-01T00:00:00Z", "workspace_id": "ws1",
	})
	want := connectors.Record{
		"id": "1", "name": "Board", "state": "active", "board_kind": "public",
		"description": "d", "type": "board", "updated_at": "2026-01-01T00:00:00Z", "workspace_id": "ws1",
	}
	for k, v := range want {
		if rec[k] != v {
			t.Fatalf("boardRecord[%q] = %v, want %v", k, rec[k], v)
		}
	}
}

func TestItemRecord_HoistsGroupAndBoard(t *testing.T) {
	rec := itemRecord(map[string]any{
		"id": "i1", "name": "Item", "state": "active",
		"created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-02T00:00:00Z",
		"group": map[string]any{"id": "g1", "title": "Group"},
		"board": map[string]any{"id": "b1", "name": "Board"},
	})
	if rec["group_id"] != "g1" || rec["group_title"] != "Group" {
		t.Fatalf("group hoist = group_id=%v group_title=%v, want g1/Group", rec["group_id"], rec["group_title"])
	}
	if rec["board_id"] != "b1" || rec["board_name"] != "Board" {
		t.Fatalf("board hoist = board_id=%v board_name=%v, want b1/Board", rec["board_id"], rec["board_name"])
	}
}

func TestItemRecord_MissingGroupBoardOmitsHoistedFields(t *testing.T) {
	rec := itemRecord(map[string]any{"id": "i1", "name": "Item"})
	if _, ok := rec["group_id"]; ok {
		t.Fatalf("group_id present without a group object: %+v", rec)
	}
	if _, ok := rec["board_id"]; ok {
		t.Fatalf("board_id present without a board object: %+v", rec)
	}
}

// --- config parsing helpers ---

func TestMondayPageSize_DefaultsAndParses(t *testing.T) {
	if got := mondayPageSize(connectors.RuntimeConfig{Config: map[string]string{}}); got != mondayDefaultPageSize {
		t.Fatalf("mondayPageSize(empty) = %d, want default %d", got, mondayDefaultPageSize)
	}
	if got := mondayPageSize(connectors.RuntimeConfig{Config: map[string]string{"page_size": "25"}}); got != 25 {
		t.Fatalf("mondayPageSize(25) = %d, want 25", got)
	}
	if got := mondayPageSize(connectors.RuntimeConfig{Config: map[string]string{"page_size": "not-a-number"}}); got != mondayDefaultPageSize {
		t.Fatalf("mondayPageSize(invalid) = %d, want default %d", got, mondayDefaultPageSize)
	}
}

func TestMondayMaxPages_UnboundedVariants(t *testing.T) {
	for _, raw := range []string{"", "all", "ALL", "unlimited"} {
		if got := mondayMaxPages(connectors.RuntimeConfig{Config: map[string]string{"max_pages": raw}}); got != 0 {
			t.Fatalf("mondayMaxPages(%q) = %d, want 0 (unbounded)", raw, got)
		}
	}
	if got := mondayMaxPages(connectors.RuntimeConfig{Config: map[string]string{"max_pages": "3"}}); got != 3 {
		t.Fatalf("mondayMaxPages(3) = %d, want 3", got)
	}
}
