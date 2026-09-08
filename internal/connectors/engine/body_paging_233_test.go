package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"testing/fstest"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

// Jira retained source d37a87d79d658bca06707a787dda1dfd0c0c78527140c3d8f1d8d26426175290,
// /rest/operations/30 and BulkChangelogRequestBean. Five literal fixture IDs
// extend the original three-ID witness; this is not a provider-chain rerun.
func bodyPagingOperation233() OperationSpec {
	return OperationSpec{ID: "acme.bulk", Kind: "rest_read", Summary: "Bulk read", Risk: "low", Approval: "none", OutputPolicy: "json_redacted", REST: &RESTOperationSpec{
		Method: "POST", Path: "/widgets", MaxBytes: 8192, ContentType: "application/json",
		Body:       map[string]any{"issueIdsOrKeys": []any{"10100", "10200", "10300", "10400", "10500"}},
		BodySchema: json.RawMessage(`{"type":"object","additionalProperties":false,"properties":{"issueIdsOrKeys":{"type":"array","minItems":1,"maxItems":1000,"items":{"type":"string"}},"nextPageToken":{"type":"string","maxLength":64},"maxResults":{"type":"integer","minimum":1,"maximum":10000,"default":2}},"required":["issueIdsOrKeys"]}`),
		Pagination: &PaginationSpec{Type: "cursor", CursorParam: "nextPageToken", TokenPath: "nextPageToken", BodyCursorField: "nextPageToken", SizeParam: "maxResults", BodyLimitField: "maxResults", PageSize: 2},
	}}
}

func TestBodyPagingAdmission233(t *testing.T) {
	for _, paging := range []bool{false, true} {
		t.Run(fmt.Sprint(paging), func(t *testing.T) {
			op := bodyPagingOperation233()
			if !paging {
				op.REST.Pagination = nil
			}
			fsys := fullValidBundleFS("acme")
			raw, err := json.Marshal(map[string]any{"operations": []OperationSpec{op}})
			if err != nil {
				t.Fatal(err)
			}
			fsys["acme/operations.json"] = &fstest.MapFile{Data: raw}
			b, err := Load(fsys, "acme")
			if err != nil {
				t.Fatalf("typed safe POST actual admission: %v", err)
			}
			if _, err := operationDirectReadSpec(b, op.ID); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestBodyPagingDirect233(t *testing.T) {
	var mu sync.Mutex
	var bodies []string
	wants := []string{
		`{"issueIdsOrKeys":["10100","10200","10300","10400","10500"],"maxResults":2}`,
		`{"issueIdsOrKeys":["10100","10200","10300","10400","10500"],"maxResults":2,"nextPageToken":"jira-next-2"}`,
		`{"issueIdsOrKeys":["10100","10200","10300","10400","10500"],"maxResults":2,"nextPageToken":"jira-next-3"}`,
	}
	responses := []string{`{"data":[{"id":"10100"},{"id":"10200"}],"nextPageToken":"jira-next-2"}`, `{"data":[{"id":"10300"},{"id":"10400"}],"nextPageToken":"jira-next-3"}`, `{"data":[{"id":"10500"}],"nextPageToken":null}`}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, err := io.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
			return
		}
		mu.Lock()
		defer mu.Unlock()
		bodies = append(bodies, string(raw))
		i := len(bodies) - 1
		if i >= len(wants) || r.Method != "POST" || r.URL.RawQuery != "" || string(raw) != wants[i] {
			t.Errorf("request%d method=%s query=%s body=%s", i+1, r.Method, r.URL.RawQuery, raw)
			w.WriteHeader(400)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, responses[i])
	}))
	defer server.Close()
	b := newTestBundle(t, server, StreamSpec{})
	op := bodyPagingOperation233()
	b.Operations = []OperationSpec{op}
	req := connectors.OperationDirectReadRequest{Operation: op.ID}
	var ids []string
	for i := 0; i < 3; i++ {
		result, err := OperationDirectRead(t.Context(), b, req, nil)
		if err != nil {
			t.Fatal(err)
		}
		raw, err := json.Marshal(result.Body)
		if err != nil {
			t.Fatal(err)
		}
		var doc struct {
			Data []struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		if err := json.Unmarshal(raw, &doc); err != nil {
			t.Fatal(err)
		}
		for _, row := range doc.Data {
			ids = append(ids, row.ID)
		}
		mu.Lock()
		sends := len(bodies)
		mu.Unlock()
		if sends != i+1 {
			t.Fatalf("direct invocation sent %d requests, want total%d", sends, i+1)
		}
		if i < 2 && result.Page.NextCursor == "" {
			t.Fatalf("missing continuation: %+v", result.Page)
		}
		req.PageCursor = result.Page.NextCursor
	}
	if !bodyPagingObservationMatches233(ids, bodies) {
		t.Fatalf("identities=%v", ids)
	}
}

func TestBodyPagingInvalidAdmission233(t *testing.T) {
	cases := []struct {
		name   string
		change func(*OperationSpec)
	}{
		{"healthy", func(*OperationSpec) {}},
		{"schema_absent", func(op *OperationSpec) { op.REST.BodySchema = nil }},
		{"unsafe_mutation", func(op *OperationSpec) { op.Kind = "rest_write" }},
		{"text_body", func(op *OperationSpec) { op.REST.ContentType = "text/plain" }},
		{"get_body", func(op *OperationSpec) { op.REST.Method = "GET" }},
		{"missing_field", func(op *OperationSpec) { op.REST.Pagination.BodyCursorField = "missing" }},
		{"wrong_role_type", func(op *OperationSpec) { op.REST.Pagination.BodyCursorField = "maxResults" }},
		{"same_destination", func(op *OperationSpec) { op.REST.Pagination.BodyLimitField = "nextPageToken" }},
		{"nested_field", func(op *OperationSpec) { op.REST.Pagination.BodyCursorField = "paging.nextPageToken" }},
		{"wrong_strategy", func(op *OperationSpec) { op.REST.Pagination.BodyPageField = "maxResults" }},
		{"two_sizes", func(op *OperationSpec) { op.REST.Pagination.LimitParam = "otherSize" }},
		{"unaddressable_body_strategy", func(op *OperationSpec) {
			op.REST.Pagination.Type = "next_url"
			op.REST.Pagination.NextURLPath = "next"
			op.REST.Pagination.BodyCursorField = ""
			op.REST.Pagination.CursorParam = ""
			op.REST.Pagination.TokenPath = ""
		}},
		{"invalid_default", func(op *OperationSpec) {
			op.REST.BodySchema = json.RawMessage(strings.ReplaceAll(string(op.REST.BodySchema), `"maximum":10000`, `"maximum":1`))
		}},
		{"fake_query_source", func(op *OperationSpec) {
			op.REST.PaginationParameters = []OperationParameter{{Name: "nextPageToken", In: "query", Type: "string"}}
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			op := bodyPagingOperation233()
			tc.change(&op)
			fsys := fullValidBundleFS("acme")
			raw, err := json.Marshal(map[string]any{"operations": []OperationSpec{op}})
			if err != nil {
				t.Fatal(err)
			}
			fsys["acme/operations.json"] = &fstest.MapFile{Data: raw}
			_, err = Load(fsys, "acme")
			if tc.name == "healthy" {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil {
				t.Fatal("invalid source contract admitted")
			}
		})
	}
}

func TestBodyPagingInput233(t *testing.T) {
	cases := []struct {
		name   string
		body   map[string]any
		page   int
		cursor string
		raw    bool
		query  map[string]string
		bad    bool
	}{
		{name: "default"},
		{name: "explicit_size", body: map[string]any{"maxResults": 1}},
		{name: "zero", body: map[string]any{"maxResults": 0}, bad: true},
		{name: "outside", body: map[string]any{"maxResults": 10001}, bad: true},
		{name: "wrong_type", body: map[string]any{"maxResults": "2"}, bad: true},
		{name: "fraction", body: map[string]any{"maxResults": json.Number("1.5")}, bad: true},
		{name: "required_null", body: map[string]any{"issueIdsOrKeys": nil}, bad: true},
		{name: "empty_required", body: map[string]any{"issueIdsOrKeys": []any{}}, bad: true},
		{name: "bad_cursor_type", body: map[string]any{"nextPageToken": 1}, bad: true},
		{name: "cursor_too_long", body: map[string]any{"nextPageToken": strings.Repeat("x", 65)}, bad: true},
		{name: "raw_json", raw: true, bad: true},
		{name: "query_cursor", query: map[string]string{"nextPageToken": "x"}, bad: true},
		{name: "query_size", query: map[string]string{"maxResults": "2"}, bad: true},
		{name: "pm_page", page: 2, bad: true},
		{name: "raw_and_pm_cursor", body: map[string]any{"nextPageToken": "raw"}, cursor: "invalid", bad: true},
		{name: "malformed_capsule", cursor: "engine_body_v2:!", bad: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var sends atomic.Int32
			var wire string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				sends.Add(1)
				raw, _ := io.ReadAll(r.Body)
				wire = string(raw)
				w.Header().Set("Content-Type", "application/json")
				_, _ = io.WriteString(w, `{"data":[{"id":"healthy"}],"nextPageToken":null}`)
			}))
			defer server.Close()
			b := newTestBundle(t, server, StreamSpec{})
			op := bodyPagingOperation233()
			b.Operations = []OperationSpec{op}
			req := connectors.OperationDirectReadRequest{Operation: op.ID, Body: tc.body, Page: tc.page, PageCursor: tc.cursor, Query: tc.query}
			if tc.raw {
				v := "{}"
				req.RawBody = &v
			}
			_, err := OperationDirectRead(t.Context(), b, req, nil)
			if tc.bad {
				if err == nil || sends.Load() != 0 {
					t.Fatalf("invalid input err=%v sends=%d", err, sends.Load())
				}
				return
			}
			want := `{"issueIdsOrKeys":["10100","10200","10300","10400","10500"],"maxResults":2}`
			if tc.name == "explicit_size" {
				want = `{"issueIdsOrKeys":["10100","10200","10300","10400","10500"],"maxResults":1}`
			}
			if err != nil || sends.Load() != 1 || wire != want {
				t.Fatalf("healthy err=%v sends=%d body=%s", err, sends.Load(), wire)
			}
		})
	}
}

func TestBodyPagingEmptyAndMixed233(t *testing.T) {
	for _, tc := range []string{"nil", "empty", "mixed_query", "exact_numeric"} {
		t.Run(tc, func(t *testing.T) {
			op := bodyPagingOperation233()
			op.REST.Body = nil
			op.REST.BodySchema = json.RawMessage(`{"type":"object","additionalProperties":false,"properties":{"nextPageToken":{"type":"string","maxLength":64},"maxResults":{"type":"integer","minimum":1,"maximum":5,"default":3},"exact":{"type":"number"}}}`)
			var body map[string]any
			if tc == "empty" {
				body = map[string]any{}
			}
			want := `{"maxResults":3}`
			if tc == "exact_numeric" {
				body = map[string]any{"exact": json.Number("0.1000000000000000200")}
				want = `{"exact":0.1000000000000000200,"maxResults":3}`
			}
			query := map[string]string{}
			wantQuery := ""
			if tc == "mixed_query" {
				op.REST.Parameters = []OperationParameter{{Name: "filter", In: "query", Type: "string"}}
				query["filter"] = "a+b c"
				wantQuery = "filter=a%2Bb+c"
			}
			var sends atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				raw, _ := io.ReadAll(r.Body)
				sends.Add(1)
				if string(raw) != want || r.URL.RawQuery != wantQuery {
					t.Errorf("body=%s query=%s", raw, r.URL.RawQuery)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = io.WriteString(w, `{"data":[{"id":"healthy"}],"nextPageToken":null}`)
			}))
			defer server.Close()
			b := newTestBundle(t, server, StreamSpec{})
			b.Operations = []OperationSpec{op}
			result, err := OperationDirectRead(t.Context(), b, connectors.OperationDirectReadRequest{Operation: op.ID, Body: body, Query: query}, nil)
			if err != nil || sends.Load() != 1 || result.Page.Size != 3 {
				t.Fatalf("err=%v sends=%d page=%+v", err, sends.Load(), result.Page)
			}
		})
	}
}

func TestBodyPagingCapsule233(t *testing.T) {
	for _, change := range []string{"healthy", "definition", "body", "query", "origin", "malformed_later", "repeated_later"} {
		t.Run(change, func(t *testing.T) {
			var sends atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				n := sends.Add(1)
				raw, _ := io.ReadAll(r.Body)
				w.Header().Set("Content-Type", "application/json")
				if n == 1 {
					token := "next"
					if change == "malformed_later" {
						token = strings.Repeat("x", 65)
					}
					_, _ = fmt.Fprintf(w, `{"data":[{"id":"one"}],"nextPageToken":%q}`, token)
					return
				}
				if !strings.Contains(string(raw), `"nextPageToken":"next"`) {
					t.Errorf("resumed body=%s", raw)
				}
				if change == "repeated_later" {
					_, _ = io.WriteString(w, `{"data":[{"id":"two"}],"nextPageToken":"next"}`)
				} else {
					_, _ = io.WriteString(w, `{"data":[{"id":"two"}],"nextPageToken":null}`)
				}
			}))
			defer server.Close()
			b := newTestBundle(t, server, StreamSpec{})
			op := bodyPagingOperation233()
			op.REST.Parameters = []OperationParameter{{Name: "filter", In: "query", Type: "string"}}
			b.Operations = []OperationSpec{op}
			req := connectors.OperationDirectReadRequest{Operation: op.ID, Query: map[string]string{"filter": "a"}}
			first, err := OperationDirectRead(t.Context(), b, req, nil)
			if change == "malformed_later" {
				if err == nil || sends.Load() != 1 {
					t.Fatalf("malformed next err=%v sends=%d", err, sends.Load())
				}
				return
			}
			if err != nil || sends.Load() != 1 || first.Page.NextCursor == "" {
				t.Fatalf("initial err=%v sends=%d", err, sends.Load())
			}
			req.PageCursor = first.Page.NextCursor
			switch change {
			case "definition":
				op.REST.Pagination.PageSize = 4
			case "body":
				req.Body = map[string]any{"maxResults": 1}
			case "query":
				req.Query["filter"] = "b"
			case "origin":
				b.HTTP.URL = "http://127.0.0.1:1"
			}
			_, err = OperationDirectRead(t.Context(), b, req, nil)
			if change == "healthy" {
				if err != nil || sends.Load() != 2 {
					t.Fatalf("healthy resume err=%v sends=%d", err, sends.Load())
				}
			} else {
				want := int32(1)
				if change == "repeated_later" {
					want = 2
				}
				if err == nil || sends.Load() != want {
					t.Fatalf("drift/loop err=%v sends=%d want%d", err, sends.Load(), want)
				}
			}
		})
	}
}

// This bridge exercises exported Read against the selected234 compiled-plan
// interface. Full generated request_inputs loader/population remains parent234
// ownership; the synthetic plan below is explicit, not loaded-source proof.
func TestBodyPagingSaved233(t *testing.T) {
	op := bodyPagingOperation233()
	var sends atomic.Int32
	wants := []string{`{"issueIdsOrKeys":["10100","10200","10300","10400","10500"],"maxResults":2}`, `{"issueIdsOrKeys":["10100","10200","10300","10400","10500"],"maxResults":2,"nextPageToken":"jira-next-2"}`, `{"issueIdsOrKeys":["10100","10200","10300","10400","10500"],"maxResults":2,"nextPageToken":"jira-next-3"}`}
	responses := []string{`{"data":[{"id":"10100"},{"id":"10200"}],"nextPageToken":"jira-next-2"}`, `{"data":[{"id":"10300"},{"id":"10400"}],"nextPageToken":"jira-next-3"}`, `{"data":[{"id":"10500"}],"nextPageToken":null}`}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		i := int(sends.Add(1)) - 1
		raw, _ := io.ReadAll(r.Body)
		if i >= 3 || r.Method != "POST" || r.URL.RawQuery != "" || string(raw) != wants[i] {
			t.Errorf("send%d %s %s body=%s", i+1, r.Method, r.URL.RawQuery, raw)
			w.WriteHeader(400)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, responses[i])
	}))
	defer server.Close()
	b := newTestBundle(t, server, StreamSpec{Method: "POST", BodyType: "json", Pagination: op.REST.Pagination})
	b.Streams[0].RequestInputs = &RequestInputContract{Version: 1, Schema: "schemas/input.json"}
	b.Streams[0].inputPlan = &compiledRequestInputPlan{bodySchema: op.REST.BodySchema}
	b.Streams[0].preparedReadBody = op.REST.Body
	b.Streams[0].preparedReadBodyPresent = true
	rows, err := readAll(t, t.Context(), b, connectors.ReadRequest{Stream: "widgets"}, nil)
	var ids []string
	for _, row := range rows {
		ids = append(ids, fmt.Sprint(row["id"]))
	}
	if err != nil || sends.Load() != 3 || !reflect.DeepEqual(ids, []string{"10100", "10200", "10300", "10400", "10500"}) {
		t.Fatalf("saved err=%v sends=%d ids=%v", err, sends.Load(), ids)
	}
}

type bodyPagingFrontier233 struct{ calls int }

func (h *bodyPagingFrontier233) ConnectorName() string { return "acme" }
func (h *bodyPagingFrontier233) Authenticator(context.Context, connectors.RuntimeConfig, AuthSpec) (connsdk.Authenticator, error) {
	h.calls++
	return nil, fmt.Errorf("healthy authentication frontier")
}

func TestBodyPagingPreRuntime233(t *testing.T) {
	for _, mode := range []string{"direct", "saved"} {
		for _, invalid := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/invalid_%v", mode, invalid), func(t *testing.T) {
				var sends atomic.Int32
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { sends.Add(1); w.WriteHeader(500) }))
				defer server.Close()
				op := bodyPagingOperation233()
				body := cloneAnyMap(op.REST.Body)
				if invalid {
					body["maxResults"] = "invalid"
				}
				b := newTestBundle(t, server, StreamSpec{Method: "POST", BodyType: "json", Pagination: op.REST.Pagination})
				b.HTTP.Auth = []AuthSpec{{Mode: "custom", Hook: "acme"}}
				h := &bodyPagingFrontier233{}
				var err error
				if mode == "direct" {
					b.Operations = []OperationSpec{op}
					_, err = OperationDirectRead(t.Context(), b, connectors.OperationDirectReadRequest{Operation: op.ID, Body: body}, h)
				} else {
					b.Streams[0].RequestInputs = &RequestInputContract{Version: 1, Schema: "schemas/input.json"}
					b.Streams[0].inputPlan = &compiledRequestInputPlan{bodySchema: op.REST.BodySchema}
					b.Streams[0].preparedReadBody = body
					b.Streams[0].preparedReadBodyPresent = true
					_, err = readAll(t, t.Context(), b, connectors.ReadRequest{Stream: "widgets"}, h)
				}
				want := 1
				if invalid {
					want = 0
				}
				if err == nil || h.calls != want || sends.Load() != 0 {
					t.Fatalf("err=%v auth=%d want%d sends=%d", err, h.calls, want, sends.Load())
				}
			})
		}
	}
}

func TestBodyPagingSavedCapsule233(t *testing.T) {
	for _, change := range []string{"healthy", "body", "query", "schema", "malformed_later"} {
		t.Run(change, func(t *testing.T) {
			var sends atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				n := sends.Add(1)
				raw, _ := io.ReadAll(r.Body)
				w.Header().Set("Content-Type", "application/json")
				if n == 1 {
					token := "next"
					if change == "malformed_later" {
						token = strings.Repeat("x", 65)
					}
					_, _ = fmt.Fprintf(w, `{"data":[{"id":"10100"},{"id":"10200"}],"nextPageToken":%q}`, token)
					return
				}
				if n == 2 {
					if !strings.Contains(string(raw), `"nextPageToken":"next"`) {
						t.Errorf("resume body=%s", raw)
					}
					_, _ = io.WriteString(w, `{"data":[{"id":"10300"},{"id":"10400"}],"nextPageToken":"last"}`)
				} else {
					_, _ = io.WriteString(w, `{"data":[{"id":"10500"}],"nextPageToken":null}`)
				}
			}))
			defer server.Close()
			op := bodyPagingOperation233()
			b := newTestBundle(t, server, StreamSpec{Method: "POST", BodyType: "json", Pagination: op.REST.Pagination})
			b.Streams[0].RequestInputs = &RequestInputContract{Version: 1, Schema: "schemas/input.json"}
			b.Streams[0].inputPlan = &compiledRequestInputPlan{bodySchema: op.REST.BodySchema}
			b.Streams[0].preparedReadBody = op.REST.Body
			b.Streams[0].preparedReadBodyPresent = true
			req := connectors.ReadRequest{Stream: "widgets", MaxPages: 1}
			var ids []string
			emit := func(r connectors.Record) error { ids = append(ids, fmt.Sprint(r["id"])); return nil }
			err := ReadWithOutcome(t.Context(), b, req, nil, emit)
			if change == "malformed_later" {
				if err == nil || sends.Load() != 1 {
					t.Fatalf("malformed err=%v sends=%d", err, sends.Load())
				}
				return
			}
			var stopped *connectors.ReadBudgetStoppedError
			if !errors.As(err, &stopped) || sends.Load() != 1 {
				t.Fatalf("cap err=%v sends=%d", err, sends.Load())
			}
			req.Continuation = &stopped.Continuation
			req.MaxPages = 0
			switch change {
			case "body":
				body := cloneAnyMap(op.REST.Body)
				body["maxResults"] = 1
				b.Streams[0].preparedReadBody = body
			case "query":
				req.Query = map[string]string{"filter": "changed"}
			case "schema":
				b.Streams[0].inputPlan = &compiledRequestInputPlan{bodySchema: json.RawMessage(strings.ReplaceAll(string(op.REST.BodySchema), `"maximum":10000`, `"maximum":9999`))}
			}
			err = ReadWithOutcome(t.Context(), b, req, nil, emit)
			if change != "healthy" {
				if err == nil || sends.Load() != 1 {
					t.Fatalf("changed input err=%v sends=%d", err, sends.Load())
				}
				return
			}
			if err != nil || sends.Load() != 3 || !reflect.DeepEqual(ids, []string{"10100", "10200", "10300", "10400", "10500"}) {
				t.Fatalf("resumed err=%v sends=%d ids=%v", err, sends.Load(), ids)
			}
		})
	}
}

func TestBodyPagingAddressable233(t *testing.T) {
	for _, strategy := range []string{"page_number", "offset_limit"} {
		t.Run(strategy, func(t *testing.T) {
			op := bodyPagingOperation233()
			op.REST.Body = nil
			op.REST.BodySchema = json.RawMessage(`{"type":"object","additionalProperties":false,"properties":{"position":{"type":"integer","minimum":0,"maximum":3},"maxResults":{"type":"integer","minimum":1,"maximum":5,"default":3}},"required":["position","maxResults"]}`)
			spec := &PaginationSpec{Type: strategy, PageSize: 2, BodyLimitField: "maxResults"}
			want := `{"maxResults":3,"position":2}`
			if strategy == "page_number" {
				spec.PageParam = "page"
				spec.SizeParam = "size"
				spec.BodyPageField = "position"
			} else {
				spec.OffsetParam = "offset"
				spec.LimitParam = "limit"
				spec.BodyOffsetField = "position"
				want = `{"maxResults":3,"position":3}`
			}
			op.REST.Pagination = spec
			var sends atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				sends.Add(1)
				raw, _ := io.ReadAll(r.Body)
				if string(raw) != want || r.URL.RawQuery != "" {
					t.Errorf("body=%s query=%s", raw, r.URL.RawQuery)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = io.WriteString(w, `{"data":[{"id":"last"}]}`)
			}))
			defer server.Close()
			b := newTestBundle(t, server, StreamSpec{})
			b.Operations = []OperationSpec{op}
			result, err := OperationDirectRead(t.Context(), b, connectors.OperationDirectReadRequest{Operation: op.ID, Page: 2}, nil)
			if err != nil || sends.Load() != 1 || result.Page.Size != 3 || !result.Page.Complete {
				t.Fatalf("err=%v sends=%d page=%+v", err, sends.Load(), result.Page)
			}
		})
	}
}

func TestBodyPagingRawCursor233(t *testing.T) {
	op := bodyPagingOperation233()
	op.REST.Pagination.BodyLimitField = ""
	op.REST.Pagination.SizeParam = ""
	// A raw source cursor is an initial position; repeating that response token
	// must not offer the same page again, even without any size query to allocate.
	var sends atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sends.Add(1)
		raw, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(raw), `"nextPageToken":"raw"`) {
			t.Errorf("raw cursor location=%s", raw)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"data":[{"id":"same"}],"nextPageToken":"raw"}`)
	}))
	defer server.Close()
	b := newTestBundle(t, server, StreamSpec{})
	b.Operations = []OperationSpec{op}
	_, err := OperationDirectRead(t.Context(), b, connectors.OperationDirectReadRequest{Operation: op.ID, Body: map[string]any{"nextPageToken": "raw"}}, nil)
	if err == nil || sends.Load() != 1 {
		t.Fatalf("raw resume loop err=%v sends=%d", err, sends.Load())
	}
}

func bodyPagingObservationMatches233(ids, bodies []string) bool {
	return reflect.DeepEqual(ids, []string{"10100", "10200", "10300", "10400", "10500"}) && reflect.DeepEqual(bodies, []string{
		`{"issueIdsOrKeys":["10100","10200","10300","10400","10500"],"maxResults":2}`,
		`{"issueIdsOrKeys":["10100","10200","10300","10400","10500"],"maxResults":2,"nextPageToken":"jira-next-2"}`,
		`{"issueIdsOrKeys":["10100","10200","10300","10400","10500"],"maxResults":2,"nextPageToken":"jira-next-3"}`,
	})
}
func TestBodyPagingOracle233(t *testing.T) {
	for _, change := range []string{"healthy", "same_count_wrong_ids", "duplicate_id", "omitted_id", "changed_body_location"} {
		t.Run(change, func(t *testing.T) {
			ids := []string{"10100", "10200", "10300", "10400", "10500"}
			bodies := []string{`{"issueIdsOrKeys":["10100","10200","10300","10400","10500"],"maxResults":2}`, `{"issueIdsOrKeys":["10100","10200","10300","10400","10500"],"maxResults":2,"nextPageToken":"jira-next-2"}`, `{"issueIdsOrKeys":["10100","10200","10300","10400","10500"],"maxResults":2,"nextPageToken":"jira-next-3"}`}
			switch change {
			case "same_count_wrong_ids":
				ids[2] = "other"
			case "duplicate_id":
				ids[2] = ids[1]
			case "omitted_id":
				ids = ids[:4]
			case "changed_body_location":
				bodies[1] = strings.ReplaceAll(bodies[1], `"nextPageToken"`, `"cursor"`)
			}
			if bodyPagingObservationMatches233(ids, bodies) != (change == "healthy") {
				t.Fatal("literal ID/body oracle failed independent perturbation")
			}
		})
	}
}

func TestBodyPagingSourceDefault233(t *testing.T) {
	for _, explicit := range []bool{false, true} {
		t.Run(fmt.Sprintf("explicit_%v", explicit), func(t *testing.T) {
			op := bodyPagingOperation233()
			op.REST.BodySchema = json.RawMessage(strings.ReplaceAll(string(op.REST.BodySchema), `"default":2`, `"default":1000`))
			want := `{"issueIdsOrKeys":["10100","10200","10300","10400","10500"],"maxResults":1000}`
			var body map[string]any
			if explicit {
				body = map[string]any{"maxResults": 2}
				want = `{"issueIdsOrKeys":["10100","10200","10300","10400","10500"],"maxResults":2}`
			}
			var sends atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				sends.Add(1)
				raw, _ := io.ReadAll(r.Body)
				if string(raw) != want {
					t.Errorf("source default/override body=%s", raw)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = io.WriteString(w, `{"data":[{"id":"source-control"}],"nextPageToken":null}`)
			}))
			defer server.Close()
			b := newTestBundle(t, server, StreamSpec{})
			b.Operations = []OperationSpec{op}
			result, err := OperationDirectRead(t.Context(), b, connectors.OperationDirectReadRequest{Operation: op.ID, Body: body}, nil)
			wantSize := 1000
			if explicit {
				wantSize = 2
			}
			if err != nil || sends.Load() != 1 || result.Page.Size != wantSize {
				t.Fatalf("err=%v sends=%d page=%+v", err, sends.Load(), result.Page)
			}
		})
	}
}

func TestBodyPagingRequiredAndBounds233(t *testing.T) {
	for _, tc := range []string{"healthy", "required_omitted", "request_bound", "wrong_body_root", "unselected_variant"} {
		t.Run(tc, func(t *testing.T) {
			var sends atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				sends.Add(1)
				w.Header().Set("Content-Type", "application/json")
				_, _ = io.WriteString(w, `{"data":[{"id":"healthy"}],"nextPageToken":null}`)
			}))
			defer server.Close()
			op := bodyPagingOperation233()
			switch tc {
			case "required_omitted":
				op.REST.Body = nil
			case "request_bound":
				op.REST.MaxBytes = 32
			case "wrong_body_root":
				op.REST.BodySchema = json.RawMessage(`{"type":"array","items":{"type":"string"},"maxItems":5}`)
			case "unselected_variant":
				op.REST.BodySchema = json.RawMessage(`{"oneOf":[{"type":"object"},{"type":"null"}]}`)
			}
			b := newTestBundle(t, server, StreamSpec{})
			b.Operations = []OperationSpec{op}
			_, err := OperationDirectRead(t.Context(), b, connectors.OperationDirectReadRequest{Operation: op.ID}, nil)
			if tc == "healthy" {
				if err != nil || sends.Load() != 1 {
					t.Fatalf("healthy err=%v sends=%d", err, sends.Load())
				}
			} else if err == nil || sends.Load() != 0 {
				t.Fatalf("invalid err=%v sends=%d", err, sends.Load())
			}
		})
	}
}
