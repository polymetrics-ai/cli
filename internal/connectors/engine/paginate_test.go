package engine

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"

	"polymetrics.ai/internal/connectors/connsdk"
)

// hitCounter is a thread-safe per-path request counter used to assert every
// page is fetched exactly once.
type hitCounter struct {
	mu   sync.Mutex
	hits map[string]int
}

func newHitCounter() *hitCounter { return &hitCounter{hits: make(map[string]int)} }

func (h *hitCounter) record(key string) int {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.hits[key]++
	return h.hits[key]
}

func (h *hitCounter) count(key string) int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.hits[key]
}

func requester(baseURL string) *connsdk.Requester {
	return &connsdk.Requester{BaseURL: baseURL}
}

// setBaseHost sets the SSRF guard's expected origin (scheme+host) on an
// SSRF-guarded paginator from a base URL string, mirroring how read.go
// derives it from requester.BaseURL before the first Harvest call.
func setBaseHost(t *testing.T, p connsdk.Paginator, baseURL string) {
	t.Helper()
	setter, ok := p.(baseHostSetter)
	if !ok {
		t.Fatalf("setBaseHost: paginator %T does not implement baseHostSetter", p)
	}
	u, err := url.Parse(baseURL)
	if err != nil {
		t.Fatalf("setBaseHost: parse %q: %v", baseURL, err)
	}
	setter.setBaseOrigin(u.Scheme, u.Host)
}

func collectPages(t *testing.T, r *connsdk.Requester, p connsdk.Paginator, recordsPath string) ([]connsdk.Record, error) {
	t.Helper()
	var out []connsdk.Record
	err := connsdk.Harvest(context.Background(), r, http.MethodGet, "list", url.Values{}, p, recordsPath, 100, func(rec connsdk.Record) error {
		out = append(out, rec)
		return nil
	})
	return out, err
}

// --- link_header ---

func TestNewPaginatorLinkHeader(t *testing.T) {
	hits := newHitCounter()
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := hits.record(r.URL.RawQuery + "#" + r.URL.Path)
		if n > 1 {
			t.Fatalf("page fetched more than once: %s", r.URL.String())
		}
		page := r.URL.Query().Get("page")
		switch page {
		case "":
			w.Header().Set("Link", `<`+srv.URL+`/list?page=2>; rel="next"`)
			_, _ = w.Write([]byte(`{"data":[{"id":1}]}`))
		case "2":
			w.Header().Set("Link", `<`+srv.URL+`/list?page=3>; rel="next"`)
			_, _ = w.Write([]byte(`{"data":[{"id":2}]}`))
		case "3":
			// no Link header: terminal page.
			_, _ = w.Write([]byte(`{"data":[{"id":3}]}`))
		default:
			t.Fatalf("unexpected page request: %s", r.URL.String())
		}
	}))
	defer srv.Close()

	p, err := newPaginator(PaginationSpec{Type: "link_header"}, 0, "data")
	if err != nil {
		t.Fatalf("newPaginator() error = %v", err)
	}
	setBaseHost(t, p, srv.URL)
	records, err := collectPages(t, requester(srv.URL), p, "data")
	if err != nil {
		t.Fatalf("collectPages() error = %v", err)
	}
	if len(records) != 3 {
		t.Fatalf("got %d records, want 3 (bounded, terminates at absent Link header)", len(records))
	}
	if records[0]["id"].(json.Number) != "1" || records[2]["id"].(json.Number) != "3" {
		t.Fatalf("unexpected record order: %+v", records)
	}
}

// TestNewPaginatorLinkHeaderCrossHostBlocked is M1 (SECURITY-REVIEW.md
// MAJOR): link_header pagination had NO SSRF guard at all — a hostile/
// compromised upstream's Link: rel="next" header pointing at an arbitrary
// host (e.g. a cloud metadata endpoint) was followed with the connector's
// auth applied. Before the fix, newPaginator's "link_header" case returns a
// bare *connsdk.LinkHeaderPaginator with no BaseHost concept at all.
func TestNewPaginatorLinkHeaderCrossHostBlocked(t *testing.T) {
	evil := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":[{"id":999}]}`))
	}))
	defer evil.Close()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Link", `<`+evil.URL+`/steal>; rel="next"`)
		_, _ = w.Write([]byte(`{"data":[{"id":1}]}`))
	}))
	defer srv.Close()

	p, err := newPaginator(PaginationSpec{Type: "link_header"}, 0, "data")
	if err != nil {
		t.Fatalf("newPaginator() error = %v", err)
	}
	setBaseHost(t, p, srv.URL)

	records, err := collectPages(t, requester(srv.URL), p, "data")
	if err != nil {
		t.Fatalf("Harvest() error = %v, want nil (guard violations surface via Err())", err)
	}
	if len(records) != 1 {
		t.Fatalf("got %d records, want 1 (only page 1; the cross-host Link header must never be followed)", len(records))
	}
	guardErr, ok := p.(interface{ Err() error })
	if !ok {
		t.Fatalf("paginator %T does not expose Err() for guard-violation reporting", p)
	}
	if guardErr.Err() == nil {
		t.Fatalf("Err() = nil, want SSRF guard error for cross-host Link header")
	}
}

// TestNewPaginatorLinkHeaderSameHostAllowed locks in that ordinary
// github-shaped same-host Link-header pagination keeps working after the
// guard is added (github.com's own paginator docstring example).
func TestNewPaginatorLinkHeaderSameHostAllowed(t *testing.T) {
	hits := newHitCounter()
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := hits.record(r.URL.RawQuery)
		if n > 1 {
			t.Fatalf("page fetched more than once: %s", r.URL.String())
		}
		if r.URL.Query().Get("page") == "" {
			w.Header().Set("Link", `<`+srv.URL+`/list?page=2>; rel="next"`)
			_, _ = w.Write([]byte(`{"data":[{"id":1}]}`))
			return
		}
		_, _ = w.Write([]byte(`{"data":[{"id":2}]}`))
	}))
	defer srv.Close()

	p, err := newPaginator(PaginationSpec{Type: "link_header"}, 0, "data")
	if err != nil {
		t.Fatalf("newPaginator() error = %v", err)
	}
	setBaseHost(t, p, srv.URL)

	records, err := collectPages(t, requester(srv.URL), p, "data")
	if err != nil {
		t.Fatalf("collectPages() error = %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("got %d records, want 2 (same-host Link-header pagination unaffected by the guard)", len(records))
	}
}

// TestNewPaginatorLinkHeaderAllowCrossHostEscape proves allow_cross_host
// opts a link_header paginator out of the guard, mirroring next_url's
// existing escape hatch.
func TestNewPaginatorLinkHeaderAllowCrossHostEscape(t *testing.T) {
	other := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":[{"id":2}]}`))
	}))
	defer other.Close()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Link", `<`+other.URL+`/list>; rel="next"`)
		_, _ = w.Write([]byte(`{"data":[{"id":1}]}`))
	}))
	defer srv.Close()

	p, err := newPaginator(PaginationSpec{Type: "link_header", AllowCrossHost: true}, 0, "data")
	if err != nil {
		t.Fatalf("newPaginator() error = %v", err)
	}
	setBaseHost(t, p, srv.URL)

	records, err := collectPages(t, requester(srv.URL), p, "data")
	if err != nil {
		t.Fatalf("Harvest() error = %v, want success when allow_cross_host is set", err)
	}
	if len(records) != 2 {
		t.Fatalf("got %d records, want 2", len(records))
	}
}

// --- page_number ---

func TestNewPaginatorPageNumberShortPageStop(t *testing.T) {
	hits := newHitCounter()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.RawQuery
		if hits.record(key) > 1 {
			t.Fatalf("page fetched more than once: %s", key)
		}
		pageno := r.URL.Query().Get("pageno")
		if r.URL.Query().Get("per_page") != "" {
			t.Fatalf("size param should not be sent for searxng-shape pagination")
		}
		switch pageno {
		case "1", "":
			_, _ = w.Write([]byte(`{"data":[{"id":1},{"id":2}]}`))
		case "2":
			// short page (< page_size 2): stop condition.
			_, _ = w.Write([]byte(`{"data":[{"id":3}]}`))
		default:
			t.Fatalf("unexpected pageno: %s", pageno)
		}
	}))
	defer srv.Close()

	p, err := newPaginator(PaginationSpec{Type: "page_number", PageParam: "pageno", StartPage: 1, PageSize: 2}, 2, "data")
	if err != nil {
		t.Fatalf("newPaginator() error = %v", err)
	}
	records, err := collectPages(t, requester(srv.URL), p, "data")
	if err != nil {
		t.Fatalf("collectPages() error = %v", err)
	}
	if len(records) != 3 {
		t.Fatalf("got %d records, want 3 (short page at page 2 stops)", len(records))
	}
}

func TestNewPaginatorPageNumberMaxPagesStop(t *testing.T) {
	hits := newHitCounter()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.RawQuery
		hits.record(key)
		// Always returns a full page so max_pages is the only stop signal.
		_, _ = w.Write([]byte(`{"data":[{"id":1},{"id":2}]}`))
	}))
	defer srv.Close()

	p, err := newPaginator(PaginationSpec{Type: "page_number", PageParam: "pageno", StartPage: 1, PageSize: 2, MaxPages: 1}, 2, "data")
	if err != nil {
		t.Fatalf("newPaginator() error = %v", err)
	}

	var pages int
	err = connsdk.Harvest(context.Background(), requester(srv.URL), http.MethodGet, "list", url.Values{}, p, "data", 1, func(connsdk.Record) error {
		return nil
	})
	if err != nil {
		t.Fatalf("Harvest() error = %v", err)
	}
	pages = hits.count("pageno=1")
	if pages != 1 {
		t.Fatalf("page 1 fetched %d times, want 1", pages)
	}
	if hits.count("pageno=2") != 0 {
		t.Fatalf("page 2 should never be requested when max_pages=1")
	}
}

// --- offset_limit ---

func TestNewPaginatorOffsetLimitShortPageStop(t *testing.T) {
	hits := newHitCounter()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.RawQuery
		if hits.record(key) > 1 {
			t.Fatalf("page fetched more than once: %s", key)
		}
		offset := r.URL.Query().Get("offset")
		switch offset {
		case "0", "":
			_, _ = w.Write([]byte(`{"data":[{"id":1},{"id":2}]}`))
		case "2":
			_, _ = w.Write([]byte(`{"data":[{"id":3}]}`))
		default:
			t.Fatalf("unexpected offset: %s", offset)
		}
	}))
	defer srv.Close()

	p, err := newPaginator(PaginationSpec{Type: "offset_limit", LimitParam: "limit", OffsetParam: "offset", PageSize: 2}, 2, "data")
	if err != nil {
		t.Fatalf("newPaginator() error = %v", err)
	}
	records, err := collectPages(t, requester(srv.URL), p, "data")
	if err != nil {
		t.Fatalf("collectPages() error = %v", err)
	}
	if len(records) != 3 {
		t.Fatalf("got %d records, want 3", len(records))
	}
}

// --- cursor(token_path) ---

func TestNewPaginatorCursorTokenPathExhausts(t *testing.T) {
	hits := newHitCounter()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.RawQuery
		if hits.record(key) > 1 {
			t.Fatalf("page fetched more than once: %s", key)
		}
		cursor := r.URL.Query().Get("cursor")
		switch cursor {
		case "":
			_, _ = w.Write([]byte(`{"data":[{"id":1}],"meta":{"next_cursor":"abc"}}`))
		case "abc":
			_, _ = w.Write([]byte(`{"data":[{"id":2}],"meta":{"next_cursor":""}}`))
		default:
			t.Fatalf("unexpected cursor: %s", cursor)
		}
	}))
	defer srv.Close()

	p, err := newPaginator(PaginationSpec{Type: "cursor", CursorParam: "cursor", TokenPath: "meta.next_cursor"}, 0, "data")
	if err != nil {
		t.Fatalf("newPaginator() error = %v", err)
	}
	records, err := collectPages(t, requester(srv.URL), p, "data")
	if err != nil {
		t.Fatalf("collectPages() error = %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("got %d records, want 2 (token exhausts to empty string)", len(records))
	}
}

// --- cursor(last_record_field + stop_path) — stripe shape ---

func TestNewPaginatorCursorLastRecordFieldStripeShape(t *testing.T) {
	hits := newHitCounter()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.RawQuery
		if hits.record(key) > 1 {
			t.Fatalf("page fetched more than once: %s", key)
		}
		after := r.URL.Query().Get("starting_after")
		switch after {
		case "":
			_, _ = w.Write([]byte(`{"data":[{"id":"cus_1"},{"id":"cus_2"}],"has_more":true}`))
		case "cus_2":
			_, _ = w.Write([]byte(`{"data":[{"id":"cus_3"}],"has_more":false}`))
		default:
			t.Fatalf("unexpected starting_after: %s", after)
		}
	}))
	defer srv.Close()

	p, err := newPaginator(PaginationSpec{
		Type:            "cursor",
		CursorParam:     "starting_after",
		LastRecordField: "id",
		StopPath:        "has_more",
	}, 0, "data")
	if err != nil {
		t.Fatalf("newPaginator() error = %v", err)
	}
	records, err := collectPages(t, requester(srv.URL), p, "data")
	if err != nil {
		t.Fatalf("collectPages() error = %v", err)
	}
	if len(records) != 3 {
		t.Fatalf("got %d records, want 3 (has_more=false on page 2 stops)", len(records))
	}
}

func TestNewPaginatorCursorLastRecordFieldEmptyPageWithHasMoreTrueDefensiveStop(t *testing.T) {
	hits := newHitCounter()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.RawQuery
		if hits.record(key) > 1 {
			t.Fatalf("page fetched more than once: %s", key)
		}
		after := r.URL.Query().Get("starting_after")
		if after != "" {
			t.Fatalf("no second page should ever be requested (defensive stop): got starting_after=%s", after)
		}
		// Empty data but has_more:true is a malformed/defensive-stop case — must
		// not infinite loop since there is no last record id to advance with.
		_, _ = w.Write([]byte(`{"data":[],"has_more":true}`))
	}))
	defer srv.Close()

	p, err := newPaginator(PaginationSpec{
		Type:            "cursor",
		CursorParam:     "starting_after",
		LastRecordField: "id",
		StopPath:        "has_more",
	}, 0, "data")
	if err != nil {
		t.Fatalf("newPaginator() error = %v", err)
	}
	records, err := collectPages(t, requester(srv.URL), p, "data")
	if err != nil {
		t.Fatalf("collectPages() error = %v", err)
	}
	if len(records) != 0 {
		t.Fatalf("got %d records, want 0", len(records))
	}
}

func TestNewPaginatorCursorLastRecordFieldMissingIDFieldStops(t *testing.T) {
	hits := newHitCounter()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.RawQuery
		if hits.record(key) > 1 {
			t.Fatalf("page fetched more than once: %s", key)
		}
		after := r.URL.Query().Get("starting_after")
		if after != "" {
			t.Fatalf("no second page should ever be requested when the id field is missing: got starting_after=%s", after)
		}
		// has_more:true but the last record lacks the "id" field entirely.
		_, _ = w.Write([]byte(`{"data":[{"name":"no id here"}],"has_more":true}`))
	}))
	defer srv.Close()

	p, err := newPaginator(PaginationSpec{
		Type:            "cursor",
		CursorParam:     "starting_after",
		LastRecordField: "id",
		StopPath:        "has_more",
	}, 0, "data")
	if err != nil {
		t.Fatalf("newPaginator() error = %v", err)
	}
	records, err := collectPages(t, requester(srv.URL), p, "data")
	if err != nil {
		t.Fatalf("collectPages() error = %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("got %d records, want 1 (stop when last record has no id field)", len(records))
	}
}

// --- F3 (REVIEW.md flag): lastRecordCursor hardcoded the records envelope
// path to "data" and required the last-record id field to be a Go string —
// an API whose list key isn't "data", or whose ids are numeric, silently
// stopped after page 1 (data truncation, no error). ---

// TestNewPaginatorCursorLastRecordFieldNonDataEnvelope proves the paginator
// derives its records-envelope path from the effective RecordsSpec (wired
// via newPaginator's recordsPath parameter) instead of hardcoding "data".
// Before the fix, lastRecordFieldValue always calls
// connsdk.RecordsAt(body, "data"), so a "results"-enveloped page silently
// truncates after page 1 (no error, no second request).
func TestNewPaginatorCursorLastRecordFieldNonDataEnvelope(t *testing.T) {
	hits := newHitCounter()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.RawQuery
		if hits.record(key) > 1 {
			t.Fatalf("page fetched more than once: %s", key)
		}
		after := r.URL.Query().Get("starting_after")
		switch after {
		case "":
			_, _ = w.Write([]byte(`{"results":[{"id":"r_1"},{"id":"r_2"}],"has_more":true}`))
		case "r_2":
			_, _ = w.Write([]byte(`{"results":[{"id":"r_3"}],"has_more":false}`))
		default:
			t.Fatalf("unexpected starting_after: %s", after)
		}
	}))
	defer srv.Close()

	p, err := newPaginator(PaginationSpec{
		Type:            "cursor",
		CursorParam:     "starting_after",
		LastRecordField: "id",
		StopPath:        "has_more",
	}, 0, "results")
	if err != nil {
		t.Fatalf("newPaginator() error = %v", err)
	}
	records, err := collectPages(t, requester(srv.URL), p, "results")
	if err != nil {
		t.Fatalf("collectPages() error = %v", err)
	}
	if len(records) != 3 {
		t.Fatalf("got %d records, want 3 (both pages fetched via the non-\"data\" envelope path)", len(records))
	}
}

// TestNewPaginatorCursorLastRecordFieldNumericID proves a numeric (not Go
// string) last-record id field still advances pagination instead of
// silently stopping. Before the fix, lastRecordFieldValue's v.(string) type
// assertion fails for a json.Number id, so ok=false and pagination halts
// after page 1 with no error.
func TestNewPaginatorCursorLastRecordFieldNumericID(t *testing.T) {
	hits := newHitCounter()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.RawQuery
		if hits.record(key) > 1 {
			t.Fatalf("page fetched more than once: %s", key)
		}
		after := r.URL.Query().Get("starting_after")
		switch after {
		case "":
			_, _ = w.Write([]byte(`{"data":[{"id":1001},{"id":1002}],"has_more":true}`))
		case "1002":
			_, _ = w.Write([]byte(`{"data":[{"id":1003}],"has_more":false}`))
		default:
			t.Fatalf("unexpected starting_after: %s", after)
		}
	}))
	defer srv.Close()

	p, err := newPaginator(PaginationSpec{
		Type:            "cursor",
		CursorParam:     "starting_after",
		LastRecordField: "id",
		StopPath:        "has_more",
	}, 0, "data")
	if err != nil {
		t.Fatalf("newPaginator() error = %v", err)
	}
	records, err := collectPages(t, requester(srv.URL), p, "data")
	if err != nil {
		t.Fatalf("collectPages() error = %v", err)
	}
	if len(records) != 3 {
		t.Fatalf("got %d records, want 3 (numeric last-record id must still advance pagination)", len(records))
	}
}

// TestLastRecordFieldValueDirectUnitCases exercises lastRecordFieldValue
// directly (unit-level, bypassing the HTTP/paginator plumbing) to cover the
// full type-dispatch table: string, empty string (rejected), json.Number
// (the real connsdk-decoded shape reached through a JSON body), and an
// unsupported type (bool) / absent / null field / empty records array.
func TestLastRecordFieldValueDirectUnitCases(t *testing.T) {
	cases := []struct {
		name   string
		body   []byte
		wantID string
		wantOK bool
	}{
		{"string id", []byte(`{"data":[{"id":"cus_1"}]}`), "cus_1", true},
		{"empty string id rejected", []byte(`{"data":[{"id":""}]}`), "", false},
		{"json.Number integer id", []byte(`{"data":[{"id":1002}]}`), "1002", true},
		{"unsupported bool id type", []byte(`{"data":[{"id":true}]}`), "", false},
		{"no records", []byte(`{"data":[]}`), "", false},
		{"field absent", []byte(`{"data":[{"name":"x"}]}`), "", false},
		{"field null", []byte(`{"data":[{"id":null}]}`), "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			id, ok := lastRecordFieldValue(tc.body, "data", "id")
			if id != tc.wantID || ok != tc.wantOK {
				t.Fatalf("lastRecordFieldValue() = (%q, %v), want (%q, %v)", id, ok, tc.wantID, tc.wantOK)
			}
		})
	}
}

// TestStringifyLastRecordIDFloat64Cases exercises stringifyLastRecordID's
// float64 branch directly — connsdk.RecordsAt itself always decodes with
// UseNumber (so lastRecordFieldValue's real call path never actually
// produces a float64), but stringifyLastRecordID is written defensively for
// any OTHER caller that builds a records map by hand (e.g. a future
// non-UseNumber decode path); this locks in that defensive behavior.
func TestStringifyLastRecordIDFloat64Cases(t *testing.T) {
	if id, ok := stringifyLastRecordID(float64(1002)); !ok || id != "1002" {
		t.Fatalf("stringifyLastRecordID(1002.0) = (%q, %v), want (1002, true) — integral float64 must not carry a redundant .0", id, ok)
	}
	if id, ok := stringifyLastRecordID(1002.5); !ok || id != "1002.5" {
		t.Fatalf("stringifyLastRecordID(1002.5) = (%q, %v), want (1002.5, true)", id, ok)
	}
}

// --- next_url ---

func TestNewPaginatorNextURLFollowsAbsoluteURL(t *testing.T) {
	hits := newHitCounter()
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Path + "?" + r.URL.RawQuery
		if hits.record(key) > 1 {
			t.Fatalf("page fetched more than once: %s", key)
		}
		if r.URL.Path == "/list" {
			_, _ = w.Write([]byte(`{"data":[{"id":1}],"next":"` + srv.URL + `/list/page2"}`))
			return
		}
		if r.URL.Path == "/list/page2" {
			_, _ = w.Write([]byte(`{"data":[{"id":2}],"next":""}`))
			return
		}
		t.Fatalf("unexpected path: %s", r.URL.Path)
	}))
	defer srv.Close()

	p, err := newPaginator(PaginationSpec{Type: "next_url", NextURLPath: "next"}, 0, "data")
	if err != nil {
		t.Fatalf("newPaginator() error = %v", err)
	}
	setBaseHost(t, p, srv.URL)
	records, err := collectPages(t, requester(srv.URL), p, "data")
	if err != nil {
		t.Fatalf("collectPages() error = %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("got %d records, want 2", len(records))
	}
}

func TestNewPaginatorNextURLLoopGuardSameURLTwice(t *testing.T) {
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Always points back at itself: a hostile/buggy API loop.
		_, _ = w.Write([]byte(`{"data":[{"id":1}],"next":"` + srv.URL + `/list"}`))
	}))
	defer srv.Close()

	p, err := newPaginator(PaginationSpec{Type: "next_url", NextURLPath: "next"}, 0, "data")
	if err != nil {
		t.Fatalf("newPaginator() error = %v", err)
	}
	setBaseHost(t, p, srv.URL)

	err = connsdk.Harvest(context.Background(), requester(srv.URL), http.MethodGet, "list", url.Values{}, p, "data", 100, func(connsdk.Record) error {
		return nil
	})
	if err != nil {
		t.Fatalf("Harvest() error = %v, want nil (guard violations surface via Err(), not Harvest's return)", err)
	}
	if guardErr := p.(*nextURL).Err(); guardErr == nil {
		t.Fatalf("nextURL.Err() = nil, want loop-guard error when next_url repeats")
	}
}

func TestNewPaginatorNextURLSSRFGuardDifferentHostRejected(t *testing.T) {
	evil := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":[{"id":999}]}`))
	}))
	defer evil.Close()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":[{"id":1}],"next":"` + evil.URL + `/steal"}`))
	}))
	defer srv.Close()

	p, err := newPaginator(PaginationSpec{Type: "next_url", NextURLPath: "next"}, 0, "data")
	if err != nil {
		t.Fatalf("newPaginator() error = %v", err)
	}
	setBaseHost(t, p, srv.URL)

	records, err := collectPages(t, requester(srv.URL), p, "data")
	if err != nil {
		t.Fatalf("Harvest() error = %v, want nil (guard violations surface via Err())", err)
	}
	if len(records) != 1 {
		t.Fatalf("got %d records, want 1 (only page 1; the cross-host next must never be followed)", len(records))
	}
	if guardErr := p.(*nextURL).Err(); guardErr == nil {
		t.Fatalf("nextURL.Err() = nil, want SSRF guard error for cross-host next_url")
	}
}

func TestNewPaginatorNextURLAllowCrossHostEscape(t *testing.T) {
	other := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":[{"id":2}],"next":""}`))
	}))
	defer other.Close()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":[{"id":1}],"next":"` + other.URL + `/list"}`))
	}))
	defer srv.Close()

	p, err := newPaginator(PaginationSpec{Type: "next_url", NextURLPath: "next", AllowCrossHost: true}, 0, "data")
	if err != nil {
		t.Fatalf("newPaginator() error = %v", err)
	}
	setBaseHost(t, p, srv.URL)

	records, err := collectPages(t, requester(srv.URL), p, "data")
	if err != nil {
		t.Fatalf("Harvest() error = %v, want success when allow_cross_host is set", err)
	}
	if len(records) != 2 {
		t.Fatalf("got %d records, want 2", len(records))
	}
}

// TestNewPaginatorNextURLSSRFGuardSchemeDowngradeBlocked is m2
// (SECURITY-REVIEW.md MINOR): the same-host guard compared host only, never
// scheme — a hostile API returning "http://<same-host>/..." when the base
// was "https://<same-host>" passed the guard, sending the follow-up request
// (potentially carrying Authorization) in cleartext. Before the fix, this
// same-host/different-scheme next URL is followed without error.
func TestNewPaginatorNextURLSSRFGuardSchemeDowngradeBlocked(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Same host as the base, but http:// instead of https://.
		downgraded := "http://" + r.Host + "/steal"
		_, _ = w.Write([]byte(`{"data":[{"id":1}],"next":"` + downgraded + `"}`))
	}))
	defer srv.Close()

	p, err := newPaginator(PaginationSpec{Type: "next_url", NextURLPath: "next"}, 0, "data")
	if err != nil {
		t.Fatalf("newPaginator() error = %v", err)
	}
	// The test server itself is http://, so simulate the base being https by
	// setting the guard's origin to https explicitly on the same host string.
	nu := p.(*nextURL)
	u, _ := url.Parse(srv.URL)
	nu.setBaseOrigin("https", u.Host)

	records, err := collectPages(t, requester(srv.URL), p, "data")
	if err != nil {
		t.Fatalf("Harvest() error = %v, want nil (guard violations surface via Err())", err)
	}
	if len(records) != 1 {
		t.Fatalf("got %d records, want 1 (a same-host but scheme-downgraded next URL must never be followed)", len(records))
	}
	if guardErr := nu.Err(); guardErr == nil {
		t.Fatalf("nextURL.Err() = nil, want scheme-downgrade guard error")
	}
}

// TestNewPaginatorNextURLUnparseableNextURLFailsClosed is m2: an unparseable
// next-URL body value must FAIL CLOSED (rejected), not silently pass through
// the guard. Before the fix, urlHost("") on a parse failure returns "",
// and the guard condition `host != "" && host != BaseHost` treats an empty
// host as "no host to compare, allow it".
func TestNewPaginatorNextURLUnparseableNextURLFailsClosed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// A control character makes this an invalid URL per net/url.Parse.
		_, _ = w.Write([]byte(`{"data":[{"id":1}],"next":"http://` + "\x7f" + `evil.example.com/steal"}`))
	}))
	defer srv.Close()

	p, err := newPaginator(PaginationSpec{Type: "next_url", NextURLPath: "next"}, 0, "data")
	if err != nil {
		t.Fatalf("newPaginator() error = %v", err)
	}
	setBaseHost(t, p, srv.URL)

	records, err := collectPages(t, requester(srv.URL), p, "data")
	if err != nil {
		t.Fatalf("Harvest() error = %v, want nil (guard violations surface via Err())", err)
	}
	if len(records) != 1 {
		t.Fatalf("got %d records, want 1 (an unparseable next URL must fail closed, never be followed)", len(records))
	}
	if guardErr := p.(*nextURL).Err(); guardErr == nil {
		t.Fatalf("nextURL.Err() = nil, want fail-closed guard error for an unparseable next URL")
	}
}

// --- none ---

func TestNewPaginatorNoneSingleRequest(t *testing.T) {
	hits := newHitCounter()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.record("only")
		_, _ = w.Write([]byte(`{"data":[{"id":1},{"id":2}]}`))
	}))
	defer srv.Close()

	p, err := newPaginator(PaginationSpec{Type: "none"}, 0, "data")
	if err != nil {
		t.Fatalf("newPaginator() error = %v", err)
	}
	records, err := collectPages(t, requester(srv.URL), p, "data")
	if err != nil {
		t.Fatalf("collectPages() error = %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("got %d records, want 2", len(records))
	}
	if hits.count("only") != 1 {
		t.Fatalf("requests made = %d, want exactly 1 for pagination type none", hits.count("only"))
	}
}

// --- malformed spec errors ---

func TestNewPaginatorUnknownTypeIsError(t *testing.T) {
	_, err := newPaginator(PaginationSpec{Type: "bogus"}, 0, "data")
	if err == nil {
		t.Fatalf("newPaginator() error = nil, want error for unknown pagination type")
	}
}

func TestNewPaginatorCursorWithBothTokenSourcesIsError(t *testing.T) {
	_, err := newPaginator(PaginationSpec{
		Type:            "cursor",
		CursorParam:     "cursor",
		TokenPath:       "meta.next",
		LastRecordField: "id",
	}, 0, "data")
	if err == nil {
		t.Fatalf("newPaginator() error = nil, want error when cursor spec sets both token_path and last_record_field")
	}
}

func TestNewPaginatorCursorWithNeitherTokenSourceIsError(t *testing.T) {
	_, err := newPaginator(PaginationSpec{Type: "cursor", CursorParam: "cursor"}, 0, "data")
	if err == nil {
		t.Fatalf("newPaginator() error = nil, want error when cursor spec sets neither token_path nor last_record_field")
	}
}

func TestNewPaginatorNextURLMissingPathIsError(t *testing.T) {
	_, err := newPaginator(PaginationSpec{Type: "next_url"}, 0, "data")
	if err == nil {
		t.Fatalf("newPaginator() error = nil, want error when next_url spec has no next_url_path")
	}
}

// sanity: verify the malformed-spec table drives newPaginator uniformly.
func TestNewPaginatorMalformedSpecTable(t *testing.T) {
	cases := []struct {
		name string
		spec PaginationSpec
	}{
		{"unknown type", PaginationSpec{Type: "not-a-type"}},
		{"cursor both sources", PaginationSpec{Type: "cursor", TokenPath: "a", LastRecordField: "b"}},
		{"cursor neither source", PaginationSpec{Type: "cursor"}},
		{"next_url no path", PaginationSpec{Type: "next_url"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := newPaginator(tc.spec, 0, "data"); err == nil {
				t.Fatalf("newPaginator(%+v) error = nil, want error", tc.spec)
			}
		})
	}
}
