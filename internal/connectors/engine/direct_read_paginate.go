package engine

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

// Direct reads used to issue one request with NO page-size parameter and
// return whatever the provider chose to put on page one. Against any API with
// a default page size — GitHub's is 30 — that silently discarded the rest of
// the collection at status 200 with nothing in the envelope to say so.
//
// A direct read stays ONE request: it is page-wise exploration, not bulk
// extraction (the ETL path is what stores data; a direct read does not). What
// changes is that the page is now the bundle's DECLARED page, and the result
// carries the page context needed to ask for the next one.
//
// Every strategy comes from the bundle's own declared pagination spec through
// newPaginator (paginate.go) — the same spec and the same seven strategies the
// ETL path consumes. There is no second pagination implementation and no
// per-connector hand-coding.
const (
	directReadReasonMorePages    = connectors.DirectReadPageReasonMorePages
	directReadReasonNoPagination = connectors.DirectReadPageReasonNoPagination
	directReadReasonAmbiguous    = connectors.DirectReadPageReasonAmbiguous
)

// addressableStrategies are the declared types that have a real page NUMBER a
// caller can name. The cursor/next_url/link_header families address pages only
// by an opaque token handed back by the provider, and start_index needs the
// previous response to compute its next window, so none of them can honour a
// "--page 7" request. Those return a next cursor instead.
func isAddressableStrategy(t string) bool {
	return t == "page_number" || t == "offset_limit"
}

// directReadWalk carries everything the shared pager needs. Both executors
// (DirectRead and OperationDirectRead) fill it identically, so neither can
// drift from the other's page contract.
type directReadWalk struct {
	method      string
	declaredPat string
	requestPath string
	query       url.Values
	body        any
	maxBytes    int
	page        int
	pageCursor  string
}

// readDirectPage issues exactly one request for the requested page and reports
// where that page sits in the collection.
func readDirectPage(ctx context.Context, b Bundle, rt *Runtime, w directReadWalk) (any, connectors.DirectReadPage, *connsdk.Response, error) {
	var spec PaginationSpec
	if b.HTTP.Pagination != nil {
		spec = *b.HTTP.Pagination
	}
	pageSize := spec.PageSize
	if pageSize == 0 {
		pageSize = defaultPageSize
	}

	requester, err := rt.requesterFor(w.method, w.declaredPat)
	if err != nil {
		return nil, connectors.DirectReadPage{}, nil, err
	}

	// A POST read (provider_search) carries its selection in a JSON body; the
	// query-param strategies cannot express its next page, so it is reported as
	// a single unpaged request rather than implying completeness.
	strategy := spec.Type
	if strategy == "" {
		strategy = "none"
	}
	if w.method != http.MethodGet {
		strategy = "none"
	}

	paginator, err := newPaginator(spec, pageSize, "")
	if err != nil {
		return nil, connectors.DirectReadPage{}, nil, fmt.Errorf("direct read pagination: %w", err)
	}
	if setter, ok := paginator.(baseHostSetter); ok {
		scheme, host := requesterOrigin(requester.BaseURL)
		setter.setBaseOrigin(scheme, host)
	}

	if err := validateDirectReadPageRequest(strategy, w); err != nil {
		return nil, connectors.DirectReadPage{}, nil, err
	}

	// Start() is what initialises a paginator's internal state — the cursor and
	// URL strategies allocate their loop-guard set there. It must run even when
	// the caller supplied a cursor and this read does not need Start's query,
	// or the later Next() call writes to a nil map.
	startPage := paginator.Start()

	number := 1
	if isAddressableStrategy(strategy) && w.page > 0 {
		number = w.page
	}

	reqPath := w.requestPath
	// The page a direct read returns must be the bundle's DECLARED page, not
	// whatever default the provider applies to a request that names no size.
	// The cursor families' paginators carry a token but never a size, so the
	// declared size_param is applied here for every strategy that declares one.
	query := mergeQuery(w.query, declaredSizeQuery(spec, pageSize))
	switch {
	case strategy == "none":
		// Nothing declared to page with: send the caller's query unchanged.
		query = w.query
	case w.pageCursor != "" && (strategy == "next_url" || strategy == "link_header"):
		// These strategies hand back an absolute next-page URL; the token IS
		// that URL, already origin-checked when it was issued below.
		if err := checkOrigin(w.pageCursor, originOf(requester.BaseURL), spec.AllowCrossHost); err != nil {
			return nil, connectors.DirectReadPage{}, nil, fmt.Errorf("direct read pagination: %w", err)
		}
		reqPath = w.pageCursor
		query = nil
	case w.pageCursor != "":
		query = mergeQuery(query, cursorQuery(spec, w.pageCursor))
	default:
		query = mergeQuery(query, seekPageQuery(paginator, startPage, number))
	}

	resp, err := requester.DoLimited(ctx, w.method, reqPath, query, w.body, w.maxBytes)
	if err != nil {
		return nil, connectors.DirectReadPage{}, nil, err
	}
	if len(resp.Body) > w.maxBytes {
		return nil, connectors.DirectReadPage{}, resp, errDirectReadTooLarge{got: len(resp.Body), limit: w.maxBytes}
	}
	decoded, err := decodeDirectReadBody(resp.Body, w.maxBytes)
	if err != nil {
		return nil, connectors.DirectReadPage{}, resp, err
	}

	page := connectors.DirectReadPage{Strategy: strategy}
	collection, isList, ambiguous := directReadCollectionKey(decoded)
	if !isList {
		// A single object is the whole answer; there is no collection to page.
		page.Complete = true
		if ambiguous {
			page.Complete = false
			page.Reason = directReadReasonAmbiguous
		}
		return decoded, page, resp, nil
	}
	items, _ := directReadItemsAt(decoded, collection)
	page.Records = len(items)

	if strategy == "none" {
		// One request is all the engine can honestly do, and it cannot prove
		// this is the whole collection — so it says so instead of implying it.
		page.Complete = false
		page.Reason = directReadReasonNoPagination
		return decoded, page, resp, nil
	}

	page.Size = pageSize
	if isAddressableStrategy(strategy) {
		page.Number = number
	}

	// The declared strategy itself decides whether another page exists: its own
	// Next() is the single source of truth, exactly as it is for the ETL path.
	if lrc, ok := paginator.(*lastRecordCursor); ok && collection != "" {
		lrc.recordsPath = collection
	}
	next := paginator.Next(resp, len(items))
	if guard, ok := paginator.(interface{ Err() error }); ok {
		if err := guard.Err(); err != nil {
			return nil, connectors.DirectReadPage{}, resp, fmt.Errorf("direct read pagination: %w", err)
		}
	}
	if next == nil {
		page.Complete = true
		return decoded, page, resp, nil
	}

	page.HasMore = true
	page.Complete = false
	page.Reason = directReadReasonMorePages
	if isAddressableStrategy(strategy) {
		page.NextNumber = number + 1
	} else if token := nextCursorToken(spec, next); token != "" {
		page.NextCursor = token
	}
	return decoded, page, resp, nil
}

// validateDirectReadPageRequest refuses navigation input the declared strategy
// cannot honour, rather than accepting it and silently returning page one — the
// same class of quiet wrongness this whole change exists to remove.
func validateDirectReadPageRequest(strategy string, w directReadWalk) error {
	if w.page > 0 && w.pageCursor != "" {
		return fmt.Errorf("direct read accepts a page number or a page cursor, not both")
	}
	if w.page < 0 {
		return fmt.Errorf("direct read page must be a positive page number, got %d", w.page)
	}
	if strategy == "none" && (w.page > 1 || w.pageCursor != "") {
		return fmt.Errorf("direct read connector declares no pagination strategy, so it cannot address page %s", describeRequestedPage(w))
	}
	if w.page > 1 && !isAddressableStrategy(strategy) {
		return fmt.Errorf("direct read pagination strategy %q has no addressable page number; pass the previous page's next_cursor instead", strategy)
	}
	if w.pageCursor != "" && isAddressableStrategy(strategy) {
		return fmt.Errorf("direct read pagination strategy %q addresses pages by number; pass a page number instead of a cursor", strategy)
	}
	return nil
}

func describeRequestedPage(w directReadWalk) string {
	if w.pageCursor != "" {
		return "by cursor"
	}
	return strconv.Itoa(w.page)
}

// seekPageQuery advances the DECLARED paginator to the requested page without
// issuing a request for each one. This is only sound for the addressable
// strategies, whose Next() derives the following page from the record count
// alone and never reads the response body — which is precisely what makes them
// addressable. Every other strategy pages by token and never reaches here.
func seekPageQuery(paginator connsdk.Paginator, page *connsdk.NextPage, number int) url.Values {
	if page == nil {
		return nil
	}
	for i := 1; i < number; i++ {
		next := paginator.Next(nil, pageAdvanceRecordCount(paginator))
		if next == nil {
			return page.Query
		}
		page = next
	}
	return page.Query
}

// pageAdvanceRecordCount reports the record count that means "this page was
// full" to an addressable paginator, so seeking forward advances instead of
// stopping on what looks like a short final page.
func pageAdvanceRecordCount(paginator connsdk.Paginator) int {
	switch p := paginator.(type) {
	case *pageNumberPaginator:
		return p.pageSize
	case *connsdk.OffsetPaginator:
		return p.PageSize
	default:
		return 0
	}
}

// cursorQuery builds the request query for a caller-supplied cursor token
// using the param the bundle itself declares.
// declaredSizeQuery renders the bundle's declared page size onto the request.
// An absent size_param leaves the query untouched — the bundle is then saying
// its provider has no size parameter to name.
func declaredSizeQuery(spec PaginationSpec, pageSize int) url.Values {
	q := url.Values{}
	if spec.SizeParam != "" && pageSize > 0 {
		q.Set(spec.SizeParam, strconv.Itoa(pageSize))
	}
	return q
}

func cursorQuery(spec PaginationSpec, token string) url.Values {
	q := url.Values{}
	param := spec.CursorParam
	if param == "" {
		param = "cursor"
	}
	if spec.Type == "start_index" {
		param = valueOrDefault(spec.StartIndexParam, defaultStartIndexParam)
	}
	q.Set(param, token)
	return q
}

// nextCursorToken extracts the opaque token from whatever the declared
// strategy produced for the next page: a URL for next_url/link_header, or the
// cursor param's value for the cursor and start_index families.
func nextCursorToken(spec PaginationSpec, next *connsdk.NextPage) string {
	if next == nil {
		return ""
	}
	if next.URL != "" {
		return next.URL
	}
	if len(next.Query) == 0 {
		return ""
	}
	for _, param := range []string{spec.CursorParam, valueOrDefault(spec.StartIndexParam, defaultStartIndexParam), "cursor"} {
		if param == "" {
			continue
		}
		if v := next.Query.Get(param); v != "" {
			return v
		}
	}
	return ""
}

func originOf(baseURL string) baseOrigin {
	scheme, host := requesterOrigin(baseURL)
	return baseOrigin{scheme: scheme, host: host}
}

// errDirectReadTooLarge lets the shared pager report the pre-existing
// oversize-response failure while each executor keeps its own message prefix.
type errDirectReadTooLarge struct {
	got   int
	limit int
}

func (e errDirectReadTooLarge) Error() string {
	return fmt.Sprintf("response too large: %d bytes exceeds limit %d", e.got, e.limit)
}

// directReadCollectionKey locates the paged collection in a decoded page. A
// root array is the collection itself (key ""). An object is a collection only
// when exactly ONE of its members is an array — with two or more there is no
// way to know which one pages, so the read reports the ambiguity instead of
// guessing.
func directReadCollectionKey(decoded any) (key string, isList bool, ambiguous bool) {
	switch v := decoded.(type) {
	case []any:
		return "", true, false
	case map[string]any:
		found := ""
		count := 0
		for name, value := range v {
			if _, ok := value.([]any); ok {
				found = name
				count++
			}
		}
		switch count {
		case 0:
			return "", false, false
		case 1:
			return found, true, false
		default:
			return "", false, true
		}
	default:
		return "", false, false
	}
}

func directReadItemsAt(decoded any, key string) ([]any, bool) {
	if key == "" {
		items, ok := decoded.([]any)
		return items, ok
	}
	obj, ok := decoded.(map[string]any)
	if !ok {
		return nil, false
	}
	items, ok := obj[key].([]any)
	return items, ok
}
