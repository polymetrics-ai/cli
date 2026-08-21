package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"

	"polymetrics.ai/internal/connectors"
)

// Direct reads used to issue exactly one request and never ask for a full
// page, so any collection larger than the provider's DEFAULT page size came
// back truncated with status 200 and no signal of any kind. These tests assert
// on RETURNED RECORD COUNTS against a fixture that holds a known-larger
// collection — never on exit status, which was already 0 while data was being
// discarded.
//
// The fixture models the shape that produced the live GitHub finding: a
// provider that serves defaultFixturePage items when the client sends no
// page-size parameter, and up to fixtureMaxPage when it does.
const (
	fixtureTotalRecords = 120
	defaultFixturePage  = 30
	fixtureMaxPage      = 100
)

// pagedFixture is a provider with a default page size. It records every
// request it receives so a test can prove how many pages the executor walked
// and whether it ever asked for a full page.
type pagedFixture struct {
	mu       sync.Mutex
	requests []fixtureRequest

	// shape selects the response envelope: "array" (bare JSON array, github
	// pulls-files shape), "results" (object with a results array plus
	// has_more/next_cursor, notion shape), or "nested" (object whose
	// continuation token lives at a dotted path, gong shape).
	shape string
	total int
}

type fixtureRequest struct {
	path  string
	query string
	size  string
}

func (f *pagedFixture) count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.requests)
}

func (f *pagedFixture) sizesRequested() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]string, 0, len(f.requests))
	for _, r := range f.requests {
		out = append(out, r.size)
	}
	return out
}

func (f *pagedFixture) total_() int {
	if f.total > 0 {
		return f.total
	}
	return fixtureTotalRecords
}

func (f *pagedFixture) window(q map[string][]string) (offset, size int, explicit string) {
	size = defaultFixturePage
	for _, name := range []string{"per_page", "page_size", "count", "limit"} {
		if v, ok := q[name]; ok && len(v) > 0 {
			explicit = v[0]
			if n, err := strconv.Atoi(v[0]); err == nil && n > 0 {
				size = min(n, fixtureMaxPage)
			}
		}
	}
	for _, name := range []string{"cursor", "start_cursor", "offset", "startIndex"} {
		if v, ok := q[name]; ok && len(v) > 0 {
			if n, err := strconv.Atoi(v[0]); err == nil {
				offset = n
			}
		}
	}
	if v, ok := q["page"]; ok && len(v) > 0 {
		if n, err := strconv.Atoi(v[0]); err == nil && n > 0 {
			offset = (n - 1) * size
		}
	}
	return offset, size, explicit
}

func (f *pagedFixture) handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		offset, size, explicit := f.window(q)
		f.mu.Lock()
		f.requests = append(f.requests, fixtureRequest{path: r.URL.Path, query: r.URL.RawQuery, size: explicit})
		f.mu.Unlock()

		end := min(offset+size, f.total_())
		if offset > f.total_() {
			offset = f.total_()
		}
		if end < offset {
			end = offset
		}
		items := make([]any, 0, end-offset)
		for i := offset; i < end; i++ {
			items = append(items, map[string]any{"id": fmt.Sprintf("rec-%03d", i)})
		}
		more := end < f.total_()
		next := strconv.Itoa(end)

		var body any
		switch f.shape {
		case "results":
			envelope := map[string]any{
				"object":   "list",
				"results":  items,
				"has_more": more,
			}
			if more {
				envelope["next_cursor"] = next
			} else {
				envelope["next_cursor"] = nil
			}
			body = envelope
		case "nested":
			records := map[string]any{"totalRecords": f.total_()}
			if more {
				records["cursor"] = next
			}
			body = map[string]any{"logs": items, "records": records}
		default:
			body = items
		}

		raw, err := json.Marshal(body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if f.shape == "array" && more {
			w.Header().Set("Link", fmt.Sprintf(`<%s?page=%d>; rel="next"`, r.URL.Path, offset/max(size, 1)+2))
		}
		_, _ = w.Write(raw)
	})
}

func startPagedFixture(t *testing.T, shape string) (*pagedFixture, *httptest.Server) {
	t.Helper()
	fx := &pagedFixture{shape: shape}
	srv := httptest.NewServer(fx.handler())
	t.Cleanup(srv.Close)
	return fx, srv
}

// paginatedOperationBundle wires a rest_read operation onto a bundle that
// declares connector-level pagination — exactly the declaration the ETL path
// already consumes, which direct reads used to ignore.
func paginatedOperationBundle(baseURL string, pagination *PaginationSpec, endpointPath string) Bundle {
	return Bundle{
		Name: "acme",
		HTTP: HTTPBase{URL: baseURL, Pagination: pagination},
		Operations: []OperationSpec{{
			ID:           "acme.list",
			Kind:         "rest_read",
			Summary:      "List",
			Risk:         "low",
			Approval:     "none",
			OutputPolicy: "json_redacted",
			REST: &RESTOperationSpec{
				Method:   http.MethodGet,
				Path:     endpointPath,
				MaxBytes: 1 << 20,
			},
		}},
		Surface: &APISurface{Endpoints: []SurfaceEndpoint{{
			Method:    http.MethodGet,
			Path:      endpointPath,
			Operation: &SurfaceOperation{Model: "direct_read", Status: "allowed", Risk: "low", Reason: "fixture"},
		}}},
	}
}

// operationPaginatedBundle declares pagination on the operation itself. A
// connector can have both offset/page and index/count endpoints, so the
// connector-wide ETL paginator is not a truthful source for either operation
// direct read.
func operationPaginatedBundle(baseURL string, pagination *PaginationSpec, sourceParameters []OperationParameter, endpointPath string) Bundle {
	b := paginatedOperationBundle(baseURL, nil, endpointPath)
	b.Operations[0].REST.Pagination = pagination
	b.Operations[0].REST.PaginationParameters = sourceParameters
	return b
}

func paginatedDirectReadBundle(baseURL string, pagination *PaginationSpec, endpointPath string) Bundle {
	b := directReadBundle(baseURL, http.MethodGet, endpointPath)
	b.HTTP.Pagination = pagination
	return b
}

func rootArrayLen(t *testing.T, body any) int {
	t.Helper()
	items, ok := body.([]any)
	if !ok {
		t.Fatalf("body type = %T, want JSON array", body)
	}
	return len(items)
}

func arrayAtLen(t *testing.T, body any, key string) int {
	t.Helper()
	obj, ok := body.(map[string]any)
	if !ok {
		t.Fatalf("body type = %T, want JSON object", body)
	}
	items, ok := obj[key].([]any)
	if !ok {
		t.Fatalf("body[%q] type = %T, want JSON array", key, obj[key])
	}
	return len(items)
}

// TestOperationDirectReadPageOneUsesDeclaredPageSize is the direct
// reproduction of the live GitHub finding. The read must return the bundle's
// DECLARED page, not the provider's default of 30, and must say that more
// records exist.
func TestOperationDirectReadPageOneUsesDeclaredPageSize(t *testing.T) {
	fx, srv := startPagedFixture(t, "array")
	b := paginatedOperationBundle(srv.URL, &PaginationSpec{
		Type:      "page_number",
		PageParam: "page",
		SizeParam: "per_page",
		PageSize:  fixtureMaxPage,
	}, "/repos/octo/hello/pulls/1/files")

	result, err := OperationDirectRead(context.Background(), b, connectors.OperationDirectReadRequest{
		Operation: "acme.list",
	}, nil)
	if err != nil {
		t.Fatalf("OperationDirectRead: %v", err)
	}
	if got := rootArrayLen(t, result.Body); got != fixtureMaxPage {
		t.Fatalf("records = %d, want the declared page of %d (provider default is %d)", got, fixtureMaxPage, defaultFixturePage)
	}
	if result.Page.Records != fixtureMaxPage {
		t.Fatalf("page.records = %d, want %d", result.Page.Records, fixtureMaxPage)
	}
	if result.Page.Complete {
		t.Fatalf("page.complete = true while %d of %d records were returned", result.Page.Records, fixtureTotalRecords)
	}
	if !result.Page.HasMore {
		t.Fatal("page.has_more = false, want true")
	}
	if result.Page.Reason != directReadReasonMorePages {
		t.Fatalf("page.reason = %q, want %q", result.Page.Reason, directReadReasonMorePages)
	}
	if result.Page.Number != 1 || result.Page.NextNumber != 2 {
		t.Fatalf("page number/next = %d/%d, want 1/2", result.Page.Number, result.Page.NextNumber)
	}
	if fx.count() != 1 {
		t.Fatalf("requests = %d, want exactly 1 — a direct read is one page", fx.count())
	}
	for _, size := range fx.sizesRequested() {
		if size != strconv.Itoa(fixtureMaxPage) {
			t.Fatalf("page size requested = %q, want the declared %d", size, fixtureMaxPage)
		}
	}
}

func TestOperationDirectReadUsesOperationDeclaredPagination(t *testing.T) {
	fx, srv := startPagedFixture(t, "array")
	b := operationPaginatedBundle(srv.URL, &PaginationSpec{
		Type:      "page_number",
		PageParam: "page",
		SizeParam: "page_size",
		PageSize:  fixtureMaxPage,
	}, []OperationParameter{
		{Name: "page", In: "query", Type: "integer"},
		{Name: "page_size", In: "query", Type: "integer"},
	}, "/v2/audit-logs")

	result, err := OperationDirectRead(context.Background(), b, connectors.OperationDirectReadRequest{Operation: "acme.list"}, nil)
	if err != nil {
		t.Fatalf("OperationDirectRead: %v", err)
	}
	if got := rootArrayLen(t, result.Body); got != fixtureMaxPage {
		t.Fatalf("records = %d, want operation declared page of %d", got, fixtureMaxPage)
	}
	if result.Page.Strategy != "page_number" || !result.Page.HasMore {
		t.Fatalf("page = %+v, want operation page_number with more records", result.Page)
	}
	if fx.count() != 1 {
		t.Fatalf("requests = %d, want exactly one", fx.count())
	}
}

func TestOperationDirectReadRejectsPaginationThatDisagreesWithSourceBeforeIO(t *testing.T) {
	requests := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[]`))
	}))
	t.Cleanup(srv.Close)
	b := operationPaginatedBundle(srv.URL, &PaginationSpec{
		Type:      "page_number",
		PageParam: "page",
		SizeParam: "page_size",
		PageSize:  2,
	}, []OperationParameter{
		{Name: "offset", In: "query", Type: "integer"},
		{Name: "limit", In: "query", Type: "integer"},
	}, "/v2/audit-logs")

	_, err := OperationDirectRead(context.Background(), b, connectors.OperationDirectReadRequest{Operation: "acme.list"}, nil)
	if err == nil || !strings.Contains(err.Error(), "pagination") || !strings.Contains(err.Error(), "source") {
		t.Fatalf("OperationDirectRead() error = %v, want source pagination refusal", err)
	}
	if requests != 0 {
		t.Fatalf("requests = %d, want source disagreement rejected before I/O", requests)
	}
}

// TestOperationDirectReadReachesEveryRecordByPageNumber proves the collection
// is fully reachable through the reported page context: the record counts of
// the pages a caller is told to fetch must add up to the whole fixture.
func TestOperationDirectReadReachesEveryRecordByPageNumber(t *testing.T) {
	_, srv := startPagedFixture(t, "array")
	b := paginatedOperationBundle(srv.URL, &PaginationSpec{
		Type:      "page_number",
		PageParam: "page",
		SizeParam: "per_page",
		PageSize:  fixtureMaxPage,
	}, "/repos/octo/hello/pulls/1/files")

	seen := 0
	page := 1
	for pages := 0; pages < 10; pages++ {
		result, err := OperationDirectRead(context.Background(), b, connectors.OperationDirectReadRequest{
			Operation: "acme.list",
			Page:      page,
		}, nil)
		if err != nil {
			t.Fatalf("OperationDirectRead(page=%d): %v", page, err)
		}
		seen += rootArrayLen(t, result.Body)
		if result.Page.Complete {
			break
		}
		if result.Page.NextNumber <= page {
			t.Fatalf("next_number = %d does not advance past %d", result.Page.NextNumber, page)
		}
		page = result.Page.NextNumber
	}
	if seen != fixtureTotalRecords {
		t.Fatalf("records reached across pages = %d, want %d", seen, fixtureTotalRecords)
	}
}

// TestOperationDirectReadCursorStrategyHandsBackNextCursor covers the notion
// shape: no addressable page number, so the caller is given a token instead —
// and following it reaches every record.
func TestOperationDirectReadCursorStrategyHandsBackNextCursor(t *testing.T) {
	_, srv := startPagedFixture(t, "results")
	b := paginatedOperationBundle(srv.URL, &PaginationSpec{
		Type:        "cursor",
		CursorParam: "start_cursor",
		TokenPath:   "next_cursor",
		StopPath:    "has_more",
		SizeParam:   "page_size",
		PageSize:    fixtureMaxPage,
	}, "/v1/comments")

	result, err := OperationDirectRead(context.Background(), b, connectors.OperationDirectReadRequest{
		Operation: "acme.list",
	}, nil)
	if err != nil {
		t.Fatalf("OperationDirectRead: %v", err)
	}
	if got := arrayAtLen(t, result.Body, "results"); got != fixtureMaxPage {
		t.Fatalf("results = %d, want the declared page of %d", got, fixtureMaxPage)
	}
	if result.Page.Complete || !result.Page.HasMore {
		t.Fatalf("page complete/has_more = %v/%v, want false/true", result.Page.Complete, result.Page.HasMore)
	}
	if result.Page.Number != 0 {
		t.Fatalf("page.number = %d, want 0 — a cursor strategy has no addressable page number", result.Page.Number)
	}
	if result.Page.NextCursor == "" {
		t.Fatal("page.next_cursor is empty, want the provider's continuation token")
	}

	seen := arrayAtLen(t, result.Body, "results")
	next, err := OperationDirectRead(context.Background(), b, connectors.OperationDirectReadRequest{
		Operation:  "acme.list",
		PageCursor: result.Page.NextCursor,
	}, nil)
	if err != nil {
		t.Fatalf("OperationDirectRead(cursor): %v", err)
	}
	seen += arrayAtLen(t, next.Body, "results")
	if seen != fixtureTotalRecords {
		t.Fatalf("records reached across pages = %d, want %d", seen, fixtureTotalRecords)
	}
	if !next.Page.Complete {
		t.Fatal("final page.complete = false, want true once the collection is exhausted")
	}
}

// TestDirectReadNestedCursorHandsBackNextCursor covers the legacy
// (non-operation) executor and the gong shape, whose token sits at a dotted
// body path.
func TestDirectReadNestedCursorHandsBackNextCursor(t *testing.T) {
	_, srv := startPagedFixture(t, "nested")
	b := paginatedDirectReadBundle(srv.URL, &PaginationSpec{
		Type:        "cursor",
		CursorParam: "cursor",
		TokenPath:   "records.cursor",
		SizeParam:   "limit",
		PageSize:    fixtureMaxPage,
	}, "/v2/logs")

	result, err := DirectRead(context.Background(), b, connectors.DirectReadRequest{
		Method:       http.MethodGet,
		Path:         "/v2/logs",
		OutputPolicy: "json_redacted",
	}, nil)
	if err != nil {
		t.Fatalf("DirectRead: %v", err)
	}
	if got := arrayAtLen(t, result.Body, "logs"); got != fixtureMaxPage {
		t.Fatalf("logs = %d, want the declared page of %d", got, fixtureMaxPage)
	}
	if result.Page.NextCursor == "" {
		t.Fatal("page.next_cursor is empty, want the provider's continuation token")
	}

	seen := arrayAtLen(t, result.Body, "logs")
	next, err := DirectRead(context.Background(), b, connectors.DirectReadRequest{
		Method:       http.MethodGet,
		Path:         "/v2/logs",
		OutputPolicy: "json_redacted",
		PageCursor:   result.Page.NextCursor,
	}, nil)
	if err != nil {
		t.Fatalf("DirectRead(cursor): %v", err)
	}
	seen += arrayAtLen(t, next.Body, "logs")
	if seen != fixtureTotalRecords {
		t.Fatalf("records reached across pages = %d, want %d", seen, fixtureTotalRecords)
	}
	if !next.Page.Complete {
		t.Fatal("final page.complete = false, want true")
	}
}

// TestOperationDirectReadOffsetLimitIsAddressableByNumber covers the mailchimp
// shape: offset_limit has a real page number, so --page must work on it.
func TestOperationDirectReadOffsetLimitIsAddressableByNumber(t *testing.T) {
	_, srv := startPagedFixture(t, "results")
	b := paginatedOperationBundle(srv.URL, &PaginationSpec{
		Type:        "offset_limit",
		LimitParam:  "count",
		OffsetParam: "offset",
		PageSize:    fixtureMaxPage,
	}, "/3.0/lists")

	first, err := OperationDirectRead(context.Background(), b, connectors.OperationDirectReadRequest{Operation: "acme.list"}, nil)
	if err != nil {
		t.Fatalf("OperationDirectRead: %v", err)
	}
	if got := arrayAtLen(t, first.Body, "results"); got != fixtureMaxPage {
		t.Fatalf("results = %d, want %d", got, fixtureMaxPage)
	}
	if first.Page.NextNumber != 2 {
		t.Fatalf("next_number = %d, want 2", first.Page.NextNumber)
	}

	second, err := OperationDirectRead(context.Background(), b, connectors.OperationDirectReadRequest{
		Operation: "acme.list",
		Page:      first.Page.NextNumber,
	}, nil)
	if err != nil {
		t.Fatalf("OperationDirectRead(page=2): %v", err)
	}
	seen := arrayAtLen(t, first.Body, "results") + arrayAtLen(t, second.Body, "results")
	if seen != fixtureTotalRecords {
		t.Fatalf("records reached across pages = %d, want %d", seen, fixtureTotalRecords)
	}
	if !second.Page.Complete {
		t.Fatal("final page.complete = false, want true")
	}
}

// TestOperationDirectReadSingleObjectIsComplete guards the reads that return
// one object: they stay a single request and are reported as complete, not as
// an incomplete collection.
func TestOperationDirectReadSingleObjectIsComplete(t *testing.T) {
	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"name":"hello","stars":7}`))
	}))
	t.Cleanup(srv.Close)

	b := paginatedOperationBundle(srv.URL, &PaginationSpec{
		Type:      "page_number",
		PageParam: "page",
		SizeParam: "per_page",
		PageSize:  fixtureMaxPage,
	}, "/repos/octo/hello")

	result, err := OperationDirectRead(context.Background(), b, connectors.OperationDirectReadRequest{Operation: "acme.list"}, nil)
	if err != nil {
		t.Fatalf("OperationDirectRead: %v", err)
	}
	if hits != 1 {
		t.Fatalf("requests = %d, want exactly 1", hits)
	}
	if !result.Page.Complete {
		t.Fatal("page.complete = false, want true for a single-object read")
	}
	obj, ok := result.Body.(map[string]any)
	if !ok || obj["name"] != "hello" {
		t.Fatalf("body = %#v, want the unmodified object", result.Body)
	}
}

// TestOperationDirectReadWithoutDeclaredPaginationIsReportedIncomplete covers a
// bundle that declares no strategy: one request is all the engine can honestly
// do, so the result must say so instead of implying completeness.
func TestOperationDirectReadWithoutDeclaredPaginationIsReportedIncomplete(t *testing.T) {
	fx, srv := startPagedFixture(t, "array")
	b := paginatedOperationBundle(srv.URL, nil, "/repos/octo/hello/pulls/1/files")

	result, err := OperationDirectRead(context.Background(), b, connectors.OperationDirectReadRequest{Operation: "acme.list"}, nil)
	if err != nil {
		t.Fatalf("OperationDirectRead: %v", err)
	}
	if got := rootArrayLen(t, result.Body); got != defaultFixturePage {
		t.Fatalf("records = %d, want the provider default of %d", got, defaultFixturePage)
	}
	if fx.count() != 1 {
		t.Fatalf("requests = %d, want exactly 1", fx.count())
	}
	if result.Page.Complete {
		t.Fatal("page.complete = true without a declared strategy, want false")
	}
	if result.Page.Reason != directReadReasonNoPagination {
		t.Fatalf("page.reason = %q, want %q", result.Page.Reason, directReadReasonNoPagination)
	}
}

// TestOperationDirectReadRefusesNavigationTheStrategyCannotHonour is the
// anti-silence guard: an unsupported page request must fail loudly rather than
// quietly returning page one, which is the very failure mode being fixed.
func TestOperationDirectReadRefusesNavigationTheStrategyCannotHonour(t *testing.T) {
	_, srv := startPagedFixture(t, "results")
	cursorBundle := paginatedOperationBundle(srv.URL, &PaginationSpec{
		Type:        "cursor",
		CursorParam: "start_cursor",
		TokenPath:   "next_cursor",
		StopPath:    "has_more",
		SizeParam:   "page_size",
		PageSize:    fixtureMaxPage,
	}, "/v1/comments")

	if _, err := OperationDirectRead(context.Background(), cursorBundle, connectors.OperationDirectReadRequest{
		Operation: "acme.list",
		Page:      3,
	}, nil); err == nil {
		t.Fatal("page 3 on a cursor strategy returned no error, want a refusal instead of a silent page one")
	}

	_, arraySrv := startPagedFixture(t, "array")
	numbered := paginatedOperationBundle(arraySrv.URL, &PaginationSpec{
		Type:      "page_number",
		PageParam: "page",
		SizeParam: "per_page",
		PageSize:  fixtureMaxPage,
	}, "/repos/octo/hello/pulls/1/files")

	if _, err := OperationDirectRead(context.Background(), numbered, connectors.OperationDirectReadRequest{
		Operation:  "acme.list",
		PageCursor: "abc",
	}, nil); err == nil {
		t.Fatal("cursor on a page_number strategy returned no error, want a refusal")
	}
}

// TestDirectReadFollowsCursorOntoANonFinalPage guards the paginator state a
// cursor-supplied read depends on. The cursor strategies allocate their
// loop-guard set in Start(), so a read that jumps straight to a cursor and
// then lands on a page that still has successors must not fault — the case an
// exhausting fixture hides, because its followed page is always the last one.
func TestDirectReadFollowsCursorOntoANonFinalPage(t *testing.T) {
	const declaredPage = 30
	_, srv := startPagedFixture(t, "nested")
	b := paginatedDirectReadBundle(srv.URL, &PaginationSpec{
		Type:        "cursor",
		CursorParam: "cursor",
		TokenPath:   "records.cursor",
		SizeParam:   "limit",
		PageSize:    declaredPage,
	}, "/v2/logs")

	first, err := DirectRead(context.Background(), b, connectors.DirectReadRequest{
		Method:       http.MethodGet,
		Path:         "/v2/logs",
		OutputPolicy: "json_redacted",
	}, nil)
	if err != nil {
		t.Fatalf("DirectRead: %v", err)
	}
	if got := arrayAtLen(t, first.Body, "logs"); got != declaredPage {
		t.Fatalf("logs = %d, want the declared page of %d", got, declaredPage)
	}

	// Page two of four: it has successors, so Next() runs its loop guard.
	second, err := DirectRead(context.Background(), b, connectors.DirectReadRequest{
		Method:       http.MethodGet,
		Path:         "/v2/logs",
		OutputPolicy: "json_redacted",
		PageCursor:   first.Page.NextCursor,
	}, nil)
	if err != nil {
		t.Fatalf("DirectRead(cursor): %v", err)
	}
	if got := arrayAtLen(t, second.Body, "logs"); got != declaredPage {
		t.Fatalf("second page logs = %d, want %d", got, declaredPage)
	}
	if !second.Page.HasMore || second.Page.Complete {
		t.Fatalf("second page has_more/complete = %v/%v, want true/false", second.Page.HasMore, second.Page.Complete)
	}
	if second.Page.NextCursor == "" {
		t.Fatal("second page next_cursor is empty, want a further continuation token")
	}

	// And the whole collection stays reachable by following the tokens.
	seen := arrayAtLen(t, first.Body, "logs") + arrayAtLen(t, second.Body, "logs")
	cursor := second.Page.NextCursor
	for i := 0; i < 10 && cursor != ""; i++ {
		next, err := DirectRead(context.Background(), b, connectors.DirectReadRequest{
			Method:       http.MethodGet,
			Path:         "/v2/logs",
			OutputPolicy: "json_redacted",
			PageCursor:   cursor,
		}, nil)
		if err != nil {
			t.Fatalf("DirectRead(cursor=%s): %v", cursor, err)
		}
		seen += arrayAtLen(t, next.Body, "logs")
		cursor = next.Page.NextCursor
	}
	if seen != fixtureTotalRecords {
		t.Fatalf("records reached by following cursors = %d, want %d", seen, fixtureTotalRecords)
	}
}

// linkHeaderFixture serves a two-page collection whose next page arrives as an
// ABSOLUTE Link-header URL — the shape a caller is handed back as next_cursor
// and can then type on the command line. It records every path it is asked
// for, so a test can prove a refused cursor never reached the wire.
type linkHeaderFixture struct {
	mu      sync.Mutex
	paths   []string
	queries []string
	base    string
}

const linkHeaderFixturePath = "/repos/octo/hello/issues"

func (f *linkHeaderFixture) requested() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.paths...)
}

func (f *linkHeaderFixture) queried() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.queries...)
}

func (f *linkHeaderFixture) handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		f.paths = append(f.paths, r.URL.Path)
		f.queries = append(f.queries, r.URL.RawQuery)
		f.mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path != linkHeaderFixturePath {
			// Any other same-origin path answers happily, so a test that
			// asserts a refusal is asserting the guard and not a 404.
			_, _ = w.Write([]byte(`[{"id":"off-endpoint"}]`))
			return
		}
		if r.URL.Query().Get("page") == "2" {
			_, _ = w.Write([]byte(`[{"id":"b-1"}]`))
			return
		}
		w.Header().Set("Link", fmt.Sprintf(`<%s%s?page=2>; rel="next"`, f.base, linkHeaderFixturePath))
		_, _ = w.Write([]byte(`[{"id":"a-1"},{"id":"a-2"}]`))
	})
}

func startLinkHeaderFixture(t *testing.T) (*linkHeaderFixture, *httptest.Server) {
	t.Helper()
	fx := &linkHeaderFixture{}
	srv := httptest.NewServer(fx.handler())
	fx.base = srv.URL
	t.Cleanup(srv.Close)
	return fx, srv
}

func linkHeaderBundle(baseURL string, allowCrossHost bool) Bundle {
	return paginatedDirectReadBundle(baseURL, &PaginationSpec{
		Type:           "link_header",
		AllowCrossHost: allowCrossHost,
	}, linkHeaderFixturePath)
}

func readLinkHeaderPage(b Bundle, cursor string) (connectors.DirectReadResult, error) {
	return DirectRead(context.Background(), b, connectors.DirectReadRequest{
		Method:       http.MethodGet,
		Path:         linkHeaderFixturePath,
		OutputPolicy: "json_redacted",
		PageCursor:   cursor,
	}, nil)
}

// TestDirectReadCursorURLFollowsTheProviderToken is the working case the guard
// below must not break: the cursor a page reported varies only the query, so
// it is the admitted endpoint and it fetches the next page.
func TestDirectReadCursorURLFollowsTheProviderToken(t *testing.T) {
	fx, srv := startLinkHeaderFixture(t)
	b := linkHeaderBundle(srv.URL, false)

	first, err := readLinkHeaderPage(b, "")
	if err != nil {
		t.Fatalf("DirectRead: %v", err)
	}
	if got := rootArrayLen(t, first.Body); got != 2 {
		t.Fatalf("records = %d, want 2", got)
	}
	if first.Page.NextCursor == "" {
		t.Fatal("page.next_cursor is empty, want the provider's Link URL")
	}

	second, err := readLinkHeaderPage(b, first.Page.NextCursor)
	if err != nil {
		t.Fatalf("DirectRead(cursor): %v", err)
	}
	if got := rootArrayLen(t, second.Body); got != 1 {
		t.Fatalf("second page records = %d, want 1", got)
	}
	if !second.Page.Complete {
		t.Fatal("final page.complete = false, want true")
	}
	for _, path := range fx.requested() {
		if path != linkHeaderFixturePath {
			t.Fatalf("requested path = %q, want only %q", path, linkHeaderFixturePath)
		}
	}
}

// TestDirectReadRefusesACursorURLOutsideTheAdmittedEndpoint is the security
// guard. DirectRead rejects an absolute req.Path and admits the relative one
// against api_surface; a typed page cursor arrives after both checks and
// becomes the request target, so it has to clear the same admission or a
// direct read would be a generic authenticated GET.
func TestDirectReadRefusesACursorURLOutsideTheAdmittedEndpoint(t *testing.T) {
	for _, allowCrossHost := range []bool{false, true} {
		t.Run(fmt.Sprintf("allow_cross_host=%v", allowCrossHost), func(t *testing.T) {
			fx, srv := startLinkHeaderFixture(t)
			b := linkHeaderBundle(srv.URL, allowCrossHost)

			_, err := readLinkHeaderPage(b, srv.URL+"/admin/secrets")
			if err == nil {
				t.Fatal("off-endpoint page cursor returned no error, want a refusal")
			}
			if !strings.Contains(err.Error(), "pagination") {
				t.Fatalf("error = %q, want it reported as a pagination refusal", err.Error())
			}
			for _, path := range fx.requested() {
				if path == "/admin/secrets" {
					t.Fatal("the refused cursor still reached the provider")
				}
			}
		})
	}
}

// TestDirectReadRefusesACrossOriginCursorURLEvenWhenCrossHostIsAllowed pins the
// asymmetry: allow_cross_host exists for a provider-supplied Link header
// discovered during a walk. It must never widen what a caller can type.
func TestDirectReadRefusesACrossOriginCursorURLEvenWhenCrossHostIsAllowed(t *testing.T) {
	var elsewhereHits int
	elsewhere := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		elsewhereHits++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":"leaked"}]`))
	}))
	t.Cleanup(elsewhere.Close)

	_, srv := startLinkHeaderFixture(t)
	b := linkHeaderBundle(srv.URL, true)

	_, err := readLinkHeaderPage(b, elsewhere.URL+linkHeaderFixturePath)
	if err == nil {
		t.Fatal("cross-origin page cursor returned no error, want a refusal")
	}
	if !strings.Contains(err.Error(), "same-origin") {
		t.Fatalf("error = %q, want a same-origin refusal", err.Error())
	}
	if elsewhereHits != 0 {
		t.Fatalf("cross-origin host received %d requests, want 0", elsewhereHits)
	}
}

// TestDirectReadCursorURLAdmitsOnlyBoundedDeclaredContinuationQuery proves a
// manually supplied same-origin URL cannot add an undeclared authenticated
// query control or a duplicate page selector after the normal command binding
// has finished. The fixture's `page` is the link-header continuation field;
// `state` is the caller's already-bound query input.
func TestDirectReadCursorURLAdmitsOnlyBoundedDeclaredContinuationQuery(t *testing.T) {
	fx, srv := startLinkHeaderFixture(t)
	b := linkHeaderBundle(srv.URL, false)
	for _, cursor := range []string{
		srv.URL + linkHeaderFixturePath + "?page=2&admin=true",
		srv.URL + linkHeaderFixturePath + "?page=2&page=3",
		srv.URL + linkHeaderFixturePath + "?page=%0Dforged",
	} {
		before := len(fx.requested())
		_, err := DirectRead(context.Background(), b, connectors.DirectReadRequest{
			Method:       http.MethodGet,
			Path:         linkHeaderFixturePath,
			Query:        map[string]string{"state": "open"},
			PageCursor:   cursor,
			OutputPolicy: "json_redacted",
		}, nil)
		if err == nil || !strings.Contains(err.Error(), "page cursor") {
			t.Fatalf("DirectRead(%q) error = %v, want cursor admission refusal", cursor, err)
		}
		if got := len(fx.requested()); got != before {
			t.Fatalf("refused cursor %q reached provider: requests %d -> %d", cursor, before, got)
		}
	}

	oversized := strings.Repeat("x", 16<<10+1)
	before := len(fx.requested())
	_, err := DirectRead(context.Background(), b, connectors.DirectReadRequest{
		Method:       http.MethodGet,
		Path:         linkHeaderFixturePath,
		PageCursor:   oversized,
		OutputPolicy: "json_redacted",
	}, nil)
	if err == nil || !strings.Contains(err.Error(), "page cursor") {
		t.Fatalf("DirectRead oversized cursor error = %v, want bounded cursor refusal", err)
	}
	if got := len(fx.requested()); got != before {
		t.Fatalf("oversized cursor reached provider: requests %d -> %d", before, got)
	}
}

// TestDirectReadPaginationGuardIsNotReportedAsADecodeFailure covers the
// diagnosis: the body parsed fine and only the NEXT-page guard fired, so
// "response is not JSON" would name the wrong thing entirely.
func TestDirectReadPaginationGuardIsNotReportedAsADecodeFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Link", `<https://attacker.invalid/repos/octo/hello/issues?page=2>; rel="next"`)
		_, _ = w.Write([]byte(`[{"id":"a-1"}]`))
	}))
	t.Cleanup(srv.Close)

	_, err := readLinkHeaderPage(linkHeaderBundle(srv.URL, false), "")
	if err == nil {
		t.Fatal("cross-host Link header returned no error, want a pagination guard failure")
	}
	if strings.Contains(err.Error(), "not JSON") {
		t.Fatalf("error = %q, want a pagination diagnosis rather than a decode failure", err.Error())
	}
	if !strings.Contains(err.Error(), "pagination") {
		t.Fatalf("error = %q, want it reported as a pagination failure", err.Error())
	}
}

// TestDirectReadAmbiguousEnvelopeStillReportsItsRowCount: which array pages is
// unknown, but how many rows arrived is not. Reporting 0 for a body that
// plainly carries them is the same untruth as an unsignalled truncation.
func TestDirectReadAmbiguousEnvelopeStillReportsItsRowCount(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"results":[{"id":1},{"id":2},{"id":3}],"links":[{"rel":"self"}]}`))
	}))
	t.Cleanup(srv.Close)

	b := paginatedDirectReadBundle(srv.URL, &PaginationSpec{
		Type:        "offset_limit",
		LimitParam:  "limit",
		OffsetParam: "startIndex",
		PageSize:    50,
	}, "/openmrs/ws/rest/v1/patient")

	result, err := DirectRead(context.Background(), b, connectors.DirectReadRequest{
		Method:       http.MethodGet,
		Path:         "/openmrs/ws/rest/v1/patient",
		OutputPolicy: "json_redacted",
	}, nil)
	if err != nil {
		t.Fatalf("DirectRead: %v", err)
	}
	if result.Page.Reason != directReadReasonAmbiguous {
		t.Fatalf("page.reason = %q, want %q", result.Page.Reason, directReadReasonAmbiguous)
	}
	if result.Page.Complete {
		t.Fatal("page.complete = true for an unidentifiable collection, want false")
	}
	if result.Page.Records != 4 {
		t.Fatalf("page.records = %d, want 4 — every top-level array element that arrived", result.Page.Records)
	}
	if result.Page.Number != 1 {
		t.Fatalf("page.number = %d, want 1 — an addressable strategy keeps its page context", result.Page.Number)
	}
	if result.Page.Size != 50 {
		t.Fatalf("page.size = %d, want the 50 actually requested", result.Page.Size)
	}
}

// TestDirectReadDegradesWhenTheDeclaredPaginationSpecIsUnusable: a malformed
// declaration is a bundle defect, not a reason to refuse a read the provider
// would have answered. It degrades that connector's paging alone, and says so.
func TestDirectReadDegradesWhenTheDeclaredPaginationSpecIsUnusable(t *testing.T) {
	fx, srv := startPagedFixture(t, "array")
	// cursor with neither token_path nor last_record_field — newPaginator
	// rejects it, exactly as inflowinventory's streams.json declares it.
	b := paginatedDirectReadBundle(srv.URL, &PaginationSpec{
		Type:        "cursor",
		CursorParam: "after",
	}, "/v2/logs")

	result, err := DirectRead(context.Background(), b, connectors.DirectReadRequest{
		Method:       http.MethodGet,
		Path:         "/v2/logs",
		OutputPolicy: "json_redacted",
	}, nil)
	if err != nil {
		t.Fatalf("DirectRead: %v — an unusable spec must not abort the read", err)
	}
	if fx.count() != 1 {
		t.Fatalf("requests = %d, want the request to have been served", fx.count())
	}
	if got := rootArrayLen(t, result.Body); got != defaultFixturePage {
		t.Fatalf("records = %d, want the provider default of %d", got, defaultFixturePage)
	}
	if result.Page.Complete {
		t.Fatal("page.complete = true on an unusable spec, want false")
	}
	if result.Page.Reason != directReadReasonInvalidSpec {
		t.Fatalf("page.reason = %q, want %q", result.Page.Reason, directReadReasonInvalidSpec)
	}
	if result.Page.Strategy != "cursor" {
		t.Fatalf("page.strategy = %q, want the declared %q", result.Page.Strategy, "cursor")
	}
	if _, err := DirectRead(context.Background(), b, connectors.DirectReadRequest{
		Method:       http.MethodGet,
		Path:         "/v2/logs",
		OutputPolicy: "json_redacted",
		PageCursor:   "abc",
	}, nil); err == nil {
		t.Fatal("navigation on an unusable spec returned no error, want a refusal")
	}
}

// TestDirectReadOmitsAPageSizeItNeverSent covers the gong/xero/recurly shape:
// no size param is declared, so the provider applied its own default and the
// envelope must not claim a size that was never requested.
func TestDirectReadOmitsAPageSizeItNeverSent(t *testing.T) {
	_, srv := startPagedFixture(t, "nested")
	b := paginatedDirectReadBundle(srv.URL, &PaginationSpec{
		Type:        "cursor",
		CursorParam: "cursor",
		TokenPath:   "records.cursor",
		PageSize:    fixtureMaxPage,
	}, "/v2/logs")

	result, err := DirectRead(context.Background(), b, connectors.DirectReadRequest{
		Method:       http.MethodGet,
		Path:         "/v2/logs",
		OutputPolicy: "json_redacted",
	}, nil)
	if err != nil {
		t.Fatalf("DirectRead: %v", err)
	}
	if got := arrayAtLen(t, result.Body, "logs"); got != defaultFixturePage {
		t.Fatalf("logs = %d, want the provider default of %d — no size param is declared", got, defaultFixturePage)
	}
	if result.Page.Size != 0 {
		t.Fatalf("page.size = %d, want it omitted: no size parameter was put on the wire", result.Page.Size)
	}
}

// TestDirectReadDistinguishesDeclaredNoneFromNoDeclaration: 68 bundles declare
// "none" outright. Telling their callers the connector declares nothing is
// false, and a caller who reads that goes looking for a declaration to fix.
func TestDirectReadDistinguishesDeclaredNoneFromNoDeclaration(t *testing.T) {
	_, srv := startPagedFixture(t, "array")
	b := paginatedDirectReadBundle(srv.URL, &PaginationSpec{Type: "none"}, "/v2/logs")

	result, err := DirectRead(context.Background(), b, connectors.DirectReadRequest{
		Method:       http.MethodGet,
		Path:         "/v2/logs",
		OutputPolicy: "json_redacted",
	}, nil)
	if err != nil {
		t.Fatalf("DirectRead: %v", err)
	}
	if result.Page.Complete {
		t.Fatal("page.complete = true, want false — a declaration is not proof")
	}
	if result.Page.Reason != directReadReasonDeclaredNone {
		t.Fatalf("page.reason = %q, want %q", result.Page.Reason, directReadReasonDeclaredNone)
	}
}

// TestOperationDirectReadPOSTReportsAStrategyItCannotUse covers the ~39 POST
// direct reads on connectors that DO declare a strategy: the request cannot be
// paged, but the connector declares plenty.
func TestOperationDirectReadPOSTReportsAStrategyItCannotUse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"calls":[{"id":"c-1"},{"id":"c-2"}]}`))
	}))
	t.Cleanup(srv.Close)

	b := Bundle{
		Name: "gong",
		HTTP: HTTPBase{URL: srv.URL, Pagination: &PaginationSpec{
			Type:        "cursor",
			CursorParam: "cursor",
			TokenPath:   "records.cursor",
		}},
		Operations: []OperationSpec{{
			ID:           "gong.calls_extensive",
			Kind:         "rest_read",
			Summary:      "List calls",
			Risk:         "low",
			Approval:     "none",
			OutputPolicy: "json_redacted",
			REST: &RESTOperationSpec{
				Method:      http.MethodPost,
				Path:        "/v2/calls/extensive",
				ContentType: "application/json",
				MaxBytes:    1 << 20,
				BodySchema:  json.RawMessage(`{"type":"object","additionalProperties":false,"properties":{"filter":{"type":"object","additionalProperties":false,"properties":{}}}}`),
			},
		}},
		Surface: &APISurface{Endpoints: []SurfaceEndpoint{{
			Method:    http.MethodPost,
			Path:      "/v2/calls/extensive",
			Operation: &SurfaceOperation{Model: "direct_read", Status: "allowed", Risk: "low", Reason: "fixture"},
		}}},
	}

	result, err := OperationDirectRead(context.Background(), b, connectors.OperationDirectReadRequest{
		Operation: "gong.calls_extensive",
	}, nil)
	if err != nil {
		t.Fatalf("OperationDirectRead: %v", err)
	}
	if result.Page.Reason != directReadReasonNotAddressable {
		t.Fatalf("page.reason = %q, want %q — the connector declares a strategy, it just cannot page a POST", result.Page.Reason, directReadReasonNotAddressable)
	}
	if result.Page.Strategy != "cursor" {
		t.Fatalf("page.strategy = %q, want the declared %q", result.Page.Strategy, "cursor")
	}
	if result.Page.Records != 2 {
		t.Fatalf("page.records = %d, want 2", result.Page.Records)
	}
}

// queriesRequested returns the raw query string of every request the fixture
// received, so a test can assert on what actually reached the wire rather than
// on what the executor intended to send.
func (f *pagedFixture) queriesRequested() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]string, 0, len(f.requests))
	for _, r := range f.requests {
		out = append(out, r.query)
	}
	return out
}

// TestDirectReadCallerPageSizeWinsOverTheDeclaredDefault reproduces notion:
// the bundle declares size_param page_size / page_size 100 and its commands
// declare a --page-size flag mapped to query.page_size. The declared default
// used to overwrite the caller's value, so `--page-size 5` fetched 100 rows
// while reporting a size the caller never asked for.
func TestDirectReadCallerPageSizeWinsOverTheDeclaredDefault(t *testing.T) {
	fx, srv := startPagedFixture(t, "results")
	b := paginatedDirectReadBundle(srv.URL, &PaginationSpec{
		Type:        "cursor",
		CursorParam: "start_cursor",
		TokenPath:   "next_cursor",
		StopPath:    "has_more",
		SizeParam:   "page_size",
		PageSize:    fixtureMaxPage,
	}, "/v1/blocks/abc/children")

	result, err := DirectRead(context.Background(), b, connectors.DirectReadRequest{
		Method:       http.MethodGet,
		Path:         "/v1/blocks/abc/children",
		Query:        map[string]string{"page_size": "5"},
		OutputPolicy: "json_redacted",
	}, nil)
	if err != nil {
		t.Fatalf("DirectRead: %v", err)
	}
	if got := arrayAtLen(t, result.Body, "results"); got != 5 {
		t.Fatalf("results = %d, want the 5 the caller asked for, not the declared %d", got, fixtureMaxPage)
	}
	if result.Page.Records != 5 {
		t.Fatalf("page.records = %d, want 5", result.Page.Records)
	}
	if result.Page.Size != 5 {
		t.Fatalf("page.size = %d, want the 5 that was actually sent", result.Page.Size)
	}
	for _, q := range fx.queriesRequested() {
		if !strings.Contains(q, "page_size=5") || strings.Contains(q, "page_size=100") {
			t.Fatalf("query sent = %q, want page_size=5", q)
		}
	}
}

// TestDirectReadCallerOffsetWinsOverThePaginatorStart reproduces bahmni
// `bahmnicore patient-search`: an offset_limit connector whose command
// declares its own offset flag. The paginator's Start() reset it to 0, so
// asking for rows 100+ silently returned rows 0-49 at exit 0.
func TestDirectReadCallerOffsetWinsOverThePaginatorStart(t *testing.T) {
	fx, srv := startPagedFixture(t, "results")
	b := paginatedDirectReadBundle(srv.URL, &PaginationSpec{
		Type:        "offset_limit",
		LimitParam:  "limit",
		OffsetParam: "offset",
		PageSize:    50,
	}, "/openmrs/ws/rest/v1/bahmnicore/search/patient")

	result, err := DirectRead(context.Background(), b, connectors.DirectReadRequest{
		Method:       http.MethodGet,
		Path:         "/openmrs/ws/rest/v1/bahmnicore/search/patient",
		Query:        map[string]string{"offset": "100"},
		OutputPolicy: "json_redacted",
	}, nil)
	if err != nil {
		t.Fatalf("DirectRead: %v", err)
	}
	// 120 fixture records, window 100..120.
	if got := arrayAtLen(t, result.Body, "results"); got != 20 {
		t.Fatalf("results = %d, want the 20 records at offset 100 — offset 0 would return 50", got)
	}
	for _, q := range fx.queriesRequested() {
		if !strings.Contains(q, "offset=100") || strings.Contains(q, "offset=0") {
			t.Fatalf("query sent = %q, want offset=100", q)
		}
	}
	if result.Page.Number != 0 || result.Page.NextNumber != 0 {
		t.Fatalf("page number/next = %d/%d, want both unset: the caller chose the window, so the engine has no page number to name", result.Page.Number, result.Page.NextNumber)
	}
}

// TestDirectReadRefusesARawPagingParameterAlongsidePageNavigation covers the
// one pairing caller-wins cannot resolve. Preferring either value would answer
// a question the caller did not ask, so nothing is sent at all.
func TestDirectReadRefusesARawPagingParameterAlongsidePageNavigation(t *testing.T) {
	fx, srv := startPagedFixture(t, "results")
	b := paginatedDirectReadBundle(srv.URL, &PaginationSpec{
		Type:        "offset_limit",
		LimitParam:  "limit",
		OffsetParam: "offset",
		PageSize:    50,
	}, "/openmrs/ws/rest/v1/bahmnicore/search/patient")

	_, err := DirectRead(context.Background(), b, connectors.DirectReadRequest{
		Method:       http.MethodGet,
		Path:         "/openmrs/ws/rest/v1/bahmnicore/search/patient",
		Query:        map[string]string{"offset": "100"},
		Page:         2,
		OutputPolicy: "json_redacted",
	}, nil)
	if err == nil {
		t.Fatal("--page alongside a raw offset returned no error, want a refusal")
	}
	if !strings.Contains(err.Error(), `"offset"`) || !strings.Contains(err.Error(), "--page") {
		t.Fatalf("error = %q, want it to name both inputs", err.Error())
	}
	if fx.count() != 0 {
		t.Fatalf("requests = %d, want 0 — a refused navigation must not reach the provider", fx.count())
	}
}

// TestDirectReadConflictRefusalNamesSomethingTheCallerCanType: the engine knows
// the QUERY parameter a command's flag maps onto, not the flag. Rendering one
// from the other printed "--startIndex" for bahmni, whose flag is
// --start-index — a spelling no caller can type and the guide never showed.
func TestDirectReadConflictRefusalNamesSomethingTheCallerCanType(t *testing.T) {
	fx, srv := startPagedFixture(t, "results")
	b := paginatedDirectReadBundle(srv.URL, bahmniPatientSearchPagination(), bahmniPatientSearchPath)

	_, err := DirectRead(context.Background(), b, connectors.DirectReadRequest{
		Method:       http.MethodGet,
		Path:         bahmniPatientSearchPath,
		Query:        map[string]string{"startIndex": "100"},
		Page:         2,
		OutputPolicy: "json_redacted",
	}, nil)
	if err == nil {
		t.Fatal("--page alongside a raw startIndex returned no error, want a refusal")
	}
	if strings.Contains(err.Error(), "--startIndex") {
		t.Fatalf("error = %q, want no invented flag spelling: the real flag is --start-index", err.Error())
	}
	for _, want := range []string{"--page", `"startIndex"`, `"offset_limit"`} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error = %q, want it to contain %s", err.Error(), want)
		}
	}
	if fx.count() != 0 {
		t.Fatalf("requests = %d, want 0", fx.count())
	}
}

// bahmniPatientSearchPagination is bahmni `bahmnicore patient-search` exactly as
// the bundle declares it. The camelCase offset param matters: it collides with
// nextCursorToken's start_index default, which is how an offset_limit read came
// to report a cursor.
const bahmniPatientSearchPath = "/openmrs/ws/rest/v1/bahmnicore/search/patient"

func bahmniPatientSearchPagination() *PaginationSpec {
	return &PaginationSpec{
		Type:        "offset_limit",
		LimitParam:  "limit",
		OffsetParam: "startIndex",
		PageSize:    50,
	}
}

// TestDirectReadCallerNavigatedAddressableStrategyEmitsNoCursor: an addressable
// strategy must never hand back next_cursor, because
// validateDirectReadPageRequest refuses a cursor for that strategy on the way
// back in. On bahmni the token was also WRONG — derived from a paginator whose
// offset was never advanced, so it named the window just read.
func TestDirectReadCallerNavigatedAddressableStrategyEmitsNoCursor(t *testing.T) {
	fx, srv := startPagedFixture(t, "results")
	b := paginatedDirectReadBundle(srv.URL, bahmniPatientSearchPagination(), bahmniPatientSearchPath)

	result, err := DirectRead(context.Background(), b, connectors.DirectReadRequest{
		Method:       http.MethodGet,
		Path:         bahmniPatientSearchPath,
		Query:        map[string]string{"startIndex": "50"},
		OutputPolicy: "json_redacted",
	}, nil)
	if err != nil {
		t.Fatalf("DirectRead: %v", err)
	}
	if got := arrayAtLen(t, result.Body, "results"); got != 50 {
		t.Fatalf("results = %d, want the 50 records at startIndex 50", got)
	}
	for _, q := range fx.queriesRequested() {
		if !strings.Contains(q, "startIndex=50") {
			t.Fatalf("query sent = %q, want startIndex=50", q)
		}
	}
	if !result.Page.HasMore || result.Page.Complete {
		t.Fatalf("page has_more/complete = %v/%v, want true/false — 20 records remain", result.Page.HasMore, result.Page.Complete)
	}
	if result.Page.NextCursor != "" {
		t.Fatalf("page.next_cursor = %q, want none: %q addresses pages by number and would refuse this cursor on the way back in", result.Page.NextCursor, result.Page.Strategy)
	}
	if result.Page.Number != 0 || result.Page.NextNumber != 0 {
		t.Fatalf("page number/next = %d/%d, want both unset: the caller chose the window", result.Page.Number, result.Page.NextNumber)
	}
}

// TestDirectReadCallerPageSizeBecomesTheCompletenessThreshold is the second way
// into the false-completeness claim. The paginator's stop rule is
// "recordCount < pageSize"; built from the DECLARED 100 while the wire carried
// the caller's 5, every page looked short, so page one of 120 records was
// reported complete with a size that had genuinely been sent.
func TestDirectReadCallerPageSizeBecomesTheCompletenessThreshold(t *testing.T) {
	fx, srv := startPagedFixture(t, "array")
	b := paginatedDirectReadBundle(srv.URL, &PaginationSpec{
		Type:      "page_number",
		PageParam: "page",
		SizeParam: "per_page",
		PageSize:  fixtureMaxPage,
	}, "/v2/logs")

	result, err := DirectRead(context.Background(), b, connectors.DirectReadRequest{
		Method:       http.MethodGet,
		Path:         "/v2/logs",
		Query:        map[string]string{"per_page": "5"},
		OutputPolicy: "json_redacted",
	}, nil)
	if err != nil {
		t.Fatalf("DirectRead: %v", err)
	}
	if got := rootArrayLen(t, result.Body); got != 5 {
		t.Fatalf("records = %d, want the 5 the caller asked for", got)
	}
	for _, q := range fx.queriesRequested() {
		if !strings.Contains(q, "per_page=5") || strings.Contains(q, "per_page=100") {
			t.Fatalf("query sent = %q, want per_page=5", q)
		}
	}
	if result.Page.Complete {
		t.Fatalf("page.complete = true after returning %d of %d records", result.Page.Records, fixtureTotalRecords)
	}
	if !result.Page.HasMore || result.Page.NextNumber != 2 {
		t.Fatalf("page has_more/next_number = %v/%d, want true/2", result.Page.HasMore, result.Page.NextNumber)
	}
	if result.Page.Size != 5 {
		t.Fatalf("page.size = %d, want the 5 that was sent", result.Page.Size)
	}

	second, err := DirectRead(context.Background(), b, connectors.DirectReadRequest{
		Method:       http.MethodGet,
		Path:         "/v2/logs",
		Query:        map[string]string{"per_page": "5"},
		Page:         result.Page.NextNumber,
		OutputPolicy: "json_redacted",
	}, nil)
	if err != nil {
		t.Fatalf("DirectRead(page=2): %v", err)
	}
	if got := rootArrayLen(t, second.Body); got != 5 {
		t.Fatalf("page 2 records = %d, want 5", got)
	}
}

// TestDirectReadCallerPageSizeAtTheDeclaredSizeStillProvesCompleteness is the
// other side of the threshold rule: the guard must not turn every caller-sized
// read into an unprovable one.
func TestDirectReadCallerPageSizeAtTheDeclaredSizeStillProvesCompleteness(t *testing.T) {
	fx := &pagedFixture{shape: "array", total: 4}
	srv := httptest.NewServer(fx.handler())
	t.Cleanup(srv.Close)
	b := paginatedDirectReadBundle(srv.URL, &PaginationSpec{
		Type:      "page_number",
		PageParam: "page",
		SizeParam: "per_page",
		PageSize:  fixtureMaxPage,
	}, "/v2/logs")

	result, err := DirectRead(context.Background(), b, connectors.DirectReadRequest{
		Method:       http.MethodGet,
		Path:         "/v2/logs",
		Query:        map[string]string{"per_page": "10"},
		OutputPolicy: "json_redacted",
	}, nil)
	if err != nil {
		t.Fatalf("DirectRead: %v", err)
	}
	if got := rootArrayLen(t, result.Body); got != 4 {
		t.Fatalf("records = %d, want the whole 4-record collection", got)
	}
	if !result.Page.Complete {
		t.Fatalf("page.complete = false (reason %q) after a genuinely short page against the size that was sent", result.Page.Reason)
	}
}

// TestDirectReadCannotClaimCompletenessWithoutSendingAPageSize covers the 13
// bundles that declare page_number or offset_limit with NO size parameter.
// Their requests carry no page size, so the provider applies its own default
// and the short-page heuristic is comparing that default against a page_size
// the request never asked for. Asserting complete from it is a false claim.
func TestDirectReadCannotClaimCompletenessWithoutSendingAPageSize(t *testing.T) {
	fx, srv := startPagedFixture(t, "array")
	b := paginatedDirectReadBundle(srv.URL, &PaginationSpec{
		Type:      "page_number",
		PageParam: "page",
		PageSize:  fixtureMaxPage,
	}, "/v2/logs")

	result, err := DirectRead(context.Background(), b, connectors.DirectReadRequest{
		Method:       http.MethodGet,
		Path:         "/v2/logs",
		OutputPolicy: "json_redacted",
	}, nil)
	if err != nil {
		t.Fatalf("DirectRead: %v", err)
	}
	if got := rootArrayLen(t, result.Body); got != defaultFixturePage {
		t.Fatalf("records = %d, want the provider default of %d", got, defaultFixturePage)
	}
	if result.Page.Complete {
		t.Fatalf("page.complete = true while %d of %d records were returned and no page size was sent", result.Page.Records, fixtureTotalRecords)
	}
	if result.Page.Reason != directReadReasonSizeNotSent {
		t.Fatalf("page.reason = %q, want %q", result.Page.Reason, directReadReasonSizeNotSent)
	}
	if result.Page.Size != 0 {
		t.Fatalf("page.size = %d, want it omitted: no size reached the wire", result.Page.Size)
	}
	for _, q := range fx.queriesRequested() {
		if strings.Contains(q, "per_page") || strings.Contains(q, "page_size") {
			t.Fatalf("query sent = %q, want no size parameter — the spec declares none", q)
		}
	}
}

// TestDirectReadCursorURLKeepsTheCallerQuery: a next_url/link_header cursor
// replaces the request target wholesale. Dropping the caller's own flags there
// would narrow the next page to something they never asked for.
func TestDirectReadCursorURLKeepsTheCallerQuery(t *testing.T) {
	fx, srv := startLinkHeaderFixture(t)
	b := linkHeaderBundle(srv.URL, false)

	first, err := DirectRead(context.Background(), b, connectors.DirectReadRequest{
		Method:       http.MethodGet,
		Path:         linkHeaderFixturePath,
		Query:        map[string]string{"state": "open"},
		OutputPolicy: "json_redacted",
	}, nil)
	if err != nil {
		t.Fatalf("DirectRead: %v", err)
	}
	if _, err := DirectRead(context.Background(), b, connectors.DirectReadRequest{
		Method:       http.MethodGet,
		Path:         linkHeaderFixturePath,
		Query:        map[string]string{"state": "open"},
		PageCursor:   first.Page.NextCursor,
		OutputPolicy: "json_redacted",
	}, nil); err != nil {
		t.Fatalf("DirectRead(cursor): %v", err)
	}
	queries := fx.queried()
	if len(queries) != 2 {
		t.Fatalf("requests = %d, want 2", len(queries))
	}
	if !strings.Contains(queries[1], "state=open") {
		t.Fatalf("cursor request query = %q, want the caller's state=open carried onto it", queries[1])
	}
	if !strings.Contains(queries[1], "page=2") {
		t.Fatalf("cursor request query = %q, want the provider's page=2 preserved", queries[1])
	}
}
