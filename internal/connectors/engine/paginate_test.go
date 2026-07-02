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

// setBaseHost sets the SSRF guard's expected host on a next_url paginator
// from a base URL string, mirroring how read.go (wave C) derives it from
// requester.BaseURL before the first Harvest call.
func setBaseHost(t *testing.T, p connsdk.Paginator, baseURL string) {
	t.Helper()
	nu, ok := p.(*nextURL)
	if !ok {
		t.Fatalf("setBaseHost: paginator is not *nextURL (%T)", p)
	}
	u, err := url.Parse(baseURL)
	if err != nil {
		t.Fatalf("setBaseHost: parse %q: %v", baseURL, err)
	}
	nu.BaseHost = u.Host
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

	p, err := newPaginator(PaginationSpec{Type: "link_header"}, 0)
	if err != nil {
		t.Fatalf("newPaginator() error = %v", err)
	}
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

	p, err := newPaginator(PaginationSpec{Type: "page_number", PageParam: "pageno", StartPage: 1, PageSize: 2}, 2)
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

	p, err := newPaginator(PaginationSpec{Type: "page_number", PageParam: "pageno", StartPage: 1, PageSize: 2, MaxPages: 1}, 2)
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

	p, err := newPaginator(PaginationSpec{Type: "offset_limit", LimitParam: "limit", OffsetParam: "offset", PageSize: 2}, 2)
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

	p, err := newPaginator(PaginationSpec{Type: "cursor", CursorParam: "cursor", TokenPath: "meta.next_cursor"}, 0)
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
	}, 0)
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
	}, 0)
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
	}, 0)
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

	p, err := newPaginator(PaginationSpec{Type: "next_url", NextURLPath: "next"}, 0)
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

	p, err := newPaginator(PaginationSpec{Type: "next_url", NextURLPath: "next"}, 0)
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

	p, err := newPaginator(PaginationSpec{Type: "next_url", NextURLPath: "next"}, 0)
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

	p, err := newPaginator(PaginationSpec{Type: "next_url", NextURLPath: "next", AllowCrossHost: true}, 0)
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

// --- none ---

func TestNewPaginatorNoneSingleRequest(t *testing.T) {
	hits := newHitCounter()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.record("only")
		_, _ = w.Write([]byte(`{"data":[{"id":1},{"id":2}]}`))
	}))
	defer srv.Close()

	p, err := newPaginator(PaginationSpec{Type: "none"}, 0)
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
	_, err := newPaginator(PaginationSpec{Type: "bogus"}, 0)
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
	}, 0)
	if err == nil {
		t.Fatalf("newPaginator() error = nil, want error when cursor spec sets both token_path and last_record_field")
	}
}

func TestNewPaginatorCursorWithNeitherTokenSourceIsError(t *testing.T) {
	_, err := newPaginator(PaginationSpec{Type: "cursor", CursorParam: "cursor"}, 0)
	if err == nil {
		t.Fatalf("newPaginator() error = nil, want error when cursor spec sets neither token_path nor last_record_field")
	}
}

func TestNewPaginatorNextURLMissingPathIsError(t *testing.T) {
	_, err := newPaginator(PaginationSpec{Type: "next_url"}, 0)
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
			if _, err := newPaginator(tc.spec, 0); err == nil {
				t.Fatalf("newPaginator(%+v) error = nil, want error", tc.spec)
			}
		})
	}
}
