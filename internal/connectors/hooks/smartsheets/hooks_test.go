package smartsheets

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/engine"
)

func newRuntime(baseURL string) *engine.Runtime {
	return &engine.Runtime{Requester: &connsdk.Requester{BaseURL: baseURL}}
}

// --- registration ---

func TestInit_RegistersHooks(t *testing.T) {
	h := engine.HooksFor("smartsheets")
	if h == nil {
		t.Fatal(`engine.HooksFor("smartsheets") = nil, want a registered hook set (init() must call engine.RegisterHooks)`)
	}
	if h.ConnectorName() != "smartsheets" {
		t.Fatalf("ConnectorName() = %q, want smartsheets", h.ConnectorName())
	}
	if _, ok := h.(engine.StreamHook); !ok {
		t.Fatal("registered hooks do not implement StreamHook")
	}
}

// --- ReadStream dispatch ---

func TestReadStream_UnknownStreamFallsBackToDeclarative(t *testing.T) {
	h := Hooks{}
	handled, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "sheets"}, connectors.ReadRequest{}, newRuntime("http://example.invalid"), func(connectors.Record) error { return nil })
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if handled {
		t.Fatal("handled = true for the sheets stream, want false (declarative fallback)")
	}
}

func TestReadStream_MissingSpreadsheetIDErrors(t *testing.T) {
	h := Hooks{}
	cfg := connectors.RuntimeConfig{Config: map[string]string{}}
	handled, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "sheet_rows"}, connectors.ReadRequest{Config: cfg}, newRuntime("http://example.invalid"), func(connectors.Record) error { return nil })
	if !handled {
		t.Fatal("handled = false for sheet_rows with a spreadsheet_id validation error, want true (the error itself is the handled outcome)")
	}
	if err == nil {
		t.Fatal("ReadStream: want error for missing spreadsheet_id, got nil")
	}
}

// TestReadStream_SheetRowsPaginatesAndFlattensColumns ports
// smartsheets_test.go's TestReadRowsPaginatesAuthenticatesAndMaps: a 2-page
// read whose second page is only requested once page 1 is consumed, and
// whose rows are flattened using the page's sibling columns[] title lookup.
func TestReadStream_SheetRowsPaginatesAndFlattensColumns(t *testing.T) {
	var pages []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sheets/900" {
			http.NotFound(w, r)
			return
		}
		pages = append(pages, r.URL.Query().Get("page"))
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Query().Get("page") {
		case "1":
			_, _ = w.Write([]byte(`{"id":900,"name":"Plan","pageNumber":1,"totalPages":2,"columns":[{"id":11,"title":"Name"}],"rows":[{"id":1,"rowNumber":1,"modifiedAt":"2026-01-01T00:00:00Z","cells":[{"columnId":11,"value":"Alpha"}]}]}`))
		case "2":
			_, _ = w.Write([]byte(`{"id":900,"name":"Plan","pageNumber":2,"totalPages":2,"columns":[{"id":11,"title":"Name"}],"rows":[{"id":2,"rowNumber":2,"modifiedAt":"2026-01-02T00:00:00Z","cells":[{"columnId":11,"value":"Beta"}]}]}`))
		default:
			t.Fatalf("unexpected page=%q", r.URL.Query().Get("page"))
		}
	}))
	defer srv.Close()

	h := Hooks{}
	cfg := connectors.RuntimeConfig{Config: map[string]string{"spreadsheet_id": "900", "page_size": "1"}}
	var got []connectors.Record
	handled, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "sheet_rows"}, connectors.ReadRequest{Config: cfg}, newRuntime(srv.URL), func(r connectors.Record) error {
		got = append(got, r)
		return nil
	})
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if !handled {
		t.Fatal("handled = false, want true")
	}
	if len(pages) != 2 {
		t.Fatalf("pages requested = %v, want exactly 2", pages)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2", len(got))
	}
	if got[0]["row_id"] != float64(1) || got[0]["Name"] != "Alpha" {
		t.Fatalf("row 0 = %+v, want row_id=1 Name=Alpha", got[0])
	}
	if got[1]["row_id"] != float64(2) || got[1]["Name"] != "Beta" {
		t.Fatalf("row 1 = %+v, want row_id=2 Name=Beta", got[1])
	}
	if got[0]["sheet_id"] != float64(900) || got[0]["sheet_name"] != "Plan" {
		t.Fatalf("row 0 sheet fields = %+v", got[0])
	}
}

func TestReadStream_UnknownColumnIDFallsBackToCellPrefix(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":900,"name":"Plan","pageNumber":1,"totalPages":1,"columns":[],"rows":[{"id":1,"rowNumber":1,"modifiedAt":"2026-01-01T00:00:00Z","cells":[{"columnId":42,"value":"Untitled"}]}]}`))
	}))
	defer srv.Close()

	h := Hooks{}
	cfg := connectors.RuntimeConfig{Config: map[string]string{"spreadsheet_id": "900"}}
	var got []connectors.Record
	handled, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "sheet_rows"}, connectors.ReadRequest{Config: cfg}, newRuntime(srv.URL), func(r connectors.Record) error {
		got = append(got, r)
		return nil
	})
	if err != nil || !handled {
		t.Fatalf("ReadStream: handled=%v err=%v", handled, err)
	}
	if len(got) != 1 || got[0]["cell_42"] != "Untitled" {
		t.Fatalf("got = %+v, want cell_42=Untitled fallback key", got)
	}
}

// --- pageSize ---

func TestPageSize(t *testing.T) {
	cases := []struct {
		raw  string
		want int
	}{
		{"", defaultPageSize},
		{"0", defaultPageSize},
		{"-1", defaultPageSize},
		{"not-a-number", defaultPageSize},
		{"25", 25},
	}
	for _, tc := range cases {
		if got := pageSize(connectors.RuntimeConfig{Config: map[string]string{"page_size": tc.raw}}); got != tc.want {
			t.Errorf("pageSize(%q) = %d, want %d", tc.raw, got, tc.want)
		}
	}
}
