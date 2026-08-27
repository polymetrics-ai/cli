package engine

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/safety"
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
	directReadReasonMorePages      = connectors.DirectReadPageReasonMorePages
	directReadReasonNoPagination   = connectors.DirectReadPageReasonNoPagination
	directReadReasonDeclaredNone   = connectors.DirectReadPageReasonDeclaredNone
	directReadReasonNotAddressable = connectors.DirectReadPageReasonNotAddressable
	directReadReasonInvalidSpec    = connectors.DirectReadPageReasonInvalidSpec
	directReadReasonAmbiguous      = connectors.DirectReadPageReasonAmbiguous
	directReadReasonSizeNotSent    = connectors.DirectReadPageReasonSizeNotRequested
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
	method          string
	declaredPat     string
	requestPath     string
	query           url.Values
	body            any
	bodyContentType string
	headers         http.Header
	operation       *OperationSpec
	outputPolicy    string
	maxBytes        int
	page            int
	pageCursor      string
	// pagination is present only for an operation-backed read. It wins over
	// the connector-level stream pagination because it is declared for this
	// exact provider endpoint.
	pagination *PaginationSpec
}

// directReadPageMode is what the bundle's declared pagination can actually do
// for THIS request. The three not-pageable cases are deliberately kept apart:
// telling a caller that a connector "declares no pagination strategy" when it
// declares a working one is the same class of untruth as an unsignalled
// truncation.
type directReadPageMode struct {
	// strategy is reported verbatim as DirectReadPage.Strategy — the bundle's
	// own declaration, or "none" when there is nothing (or nothing usable) to
	// page with.
	strategy string
	// pageable is true only when the declared strategy built a paginator that
	// can address another page of this request.
	pageable bool
	reason   string
	// detail carries the newPaginator failure for the invalid-spec case, so
	// the refusal names the declaration that is wrong.
	detail string
}

// describe renders the mode as the reason half of a refusal message.
func (m directReadPageMode) describe() string {
	switch m.reason {
	case directReadReasonNoPagination:
		return "this connector declares no pagination strategy"
	case directReadReasonDeclaredNone:
		return `this connector declares pagination type "none"`
	case directReadReasonNotAddressable:
		return fmt.Sprintf("the declared %q pagination cannot page this request", m.strategy)
	case directReadReasonInvalidSpec:
		return fmt.Sprintf("the declared %q pagination is unusable: %s", m.strategy, m.detail)
	default:
		return "this read cannot be paged"
	}
}

// resolveDirectReadPageMode decides, once, what the declared spec can do here.
//
// A malformed declaration degrades THAT connector to single-page behaviour
// with a reason naming it; it must never abort a read the provider would have
// answered, because the bundle's pagination spec is not what the caller asked
// about.
func resolveDirectReadPageMode(spec PaginationSpec, method string, paginatorErr error) directReadPageMode {
	declared := strings.TrimSpace(spec.Type)
	switch {
	case declared == "":
		return directReadPageMode{strategy: "none", reason: directReadReasonNoPagination}
	case declared == "none":
		return directReadPageMode{strategy: "none", reason: directReadReasonDeclaredNone}
	case paginatorErr != nil:
		return directReadPageMode{strategy: declared, reason: directReadReasonInvalidSpec, detail: paginatorErr.Error()}
	case method != http.MethodGet:
		// A POST read (provider_search) carries its selection in a JSON body;
		// the query-param strategies cannot express its next page.
		return directReadPageMode{strategy: declared, reason: directReadReasonNotAddressable}
	default:
		return directReadPageMode{strategy: declared, pageable: true}
	}
}

// readDirectPage issues exactly one request for the requested page and reports
// where that page sits in the collection.
func readDirectPage(ctx context.Context, b Bundle, rt *Runtime, w directReadWalk) (any, connectors.DirectReadPage, *connsdk.Response, error) {
	var spec PaginationSpec
	if w.pagination != nil {
		spec = *w.pagination
	} else if b.HTTP.Pagination != nil {
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
	if w.operation != nil {
		requester, err = requesterWithOperationHeaders(requester, *w.operation, w.headers)
		if err != nil {
			return nil, connectors.DirectReadPage{}, nil, err
		}
	}

	// The paginator's stop rule for page_number/offset_limit is "this page held
	// fewer records than the page size", so its threshold has to be the size
	// that actually reaches the wire. Caller-wins means the declared size is no
	// longer always that size: an explicit --page-size 5 against a declared 100
	// would make every page look short and assert complete on page one. Resolve
	// the effective size BEFORE building the paginator so the threshold and the
	// request cannot disagree by construction.
	pageSize = effectiveDirectReadPageSize(spec, pageSize, w.query)
	walkSpec := spec
	walkSpec.PageSize = pageSize

	paginator, paginatorErr := newPaginator(walkSpec, pageSize, "")
	mode := resolveDirectReadPageMode(spec, w.method, paginatorErr)
	strategy := mode.strategy

	if err := validateDirectReadPageRequest(mode, w); err != nil {
		return nil, connectors.DirectReadPage{}, nil, errDirectReadPagination{err: err}
	}
	if err := connectors.ValidateDirectReadPageCursor(w.pageCursor); err != nil {
		return nil, connectors.DirectReadPage{}, nil, errDirectReadPagination{err: err}
	}

	number := 1
	reqPath := w.requestPath
	query := w.query
	callerNavigated := false

	if mode.pageable {
		if setter, ok := paginator.(baseHostSetter); ok {
			scheme, host := requesterOrigin(requester.BaseURL)
			setter.setBaseOrigin(scheme, host)
		}

		// Start() is what initialises a paginator's internal state — the cursor
		// and URL strategies allocate their loop-guard set there. It must run
		// even when the caller supplied a cursor and this read does not need
		// Start's query, or the later Next() call writes to a nil map.
		startPage := paginator.Start()

		if isAddressableStrategy(strategy) && w.page > 0 {
			number = w.page
		}

		supplied := callerPagingParams(spec, strategy, w.query)
		if err := refuseConflictingPagingInput(supplied, w, strategy); err != nil {
			return nil, connectors.DirectReadPage{}, nil, errDirectReadPagination{err: err}
		}
		callerNavigated = len(supplied) > 0

		// The page a direct read returns must be the bundle's DECLARED page,
		// not whatever default the provider applies to a request that names no
		// size. The cursor families' paginators carry a token but never a size,
		// so the declared size_param is applied here for every strategy that
		// declares one.
		//
		// Every value the engine derives goes in the BASE position, so a value
		// the CALLER named for the same parameter wins. A declared default that
		// overwrote an explicit `--page-size 5` with the bundle's 100, or an
		// explicit `--start-index 100` with the paginator's 0, would answer a
		// question the caller did not ask and never say so. The one pairing
		// that cannot be resolved this way — a raw paging parameter alongside
		// --page/--page-cursor — is refused above rather than silently ranked.
		query = mergeQuery(declaredSizeQuery(spec, pageSize), w.query)
		switch {
		case w.pageCursor != "" && (strategy == "next_url" || strategy == "link_header"):
			// These strategies hand back an absolute next-page URL, so the
			// token IS a URL. A caller can type one, which means it has to
			// clear the same admission the declared endpoint cleared rather
			// than becoming a generic authenticated GET.
			admitted, err := admitDirectReadCursorURL(requester.BaseURL, w.requestPath, w.pageCursor, spec, strategy, w.query)
			if err != nil {
				return nil, connectors.DirectReadPage{}, nil, errDirectReadPagination{err: err}
			}
			reqPath = admitted
			// The cursor URL carries the query of the page it continues, and
			// the requester merges these over it. Dropping the caller's own
			// flags here would silently narrow the next page to something the
			// caller never asked for.
			query = w.query
		case w.pageCursor != "":
			query = mergeQuery(cursorQuery(spec, w.pageCursor), query)
		default:
			query = mergeQuery(seekPageQuery(paginator, startPage, number), query)
		}
	}

	// Only a size that is actually on the wire may be reported as the page
	// size: 143 of the paginating bundles declare no size/limit/count param at
	// all, and those requests still receive the provider's own default.
	sizeSent := directReadRequestedSize(spec, strategy, query)

	var resp *connsdk.Response
	if w.bodyContentType == "text/plain" {
		text, ok := w.body.(string)
		if !ok {
			return nil, connectors.DirectReadPage{}, nil, fmt.Errorf("declared text/plain direct-read body is not text")
		}
		resp, err = requester.DoTextLimited(ctx, w.method, reqPath, query, text, w.maxBytes)
	} else {
		resp, err = requester.DoLimited(ctx, w.method, reqPath, query, w.body, w.maxBytes)
	}
	if err != nil {
		return nil, connectors.DirectReadPage{}, nil, err
	}
	if len(resp.Body) > w.maxBytes {
		return nil, connectors.DirectReadPage{}, resp, errDirectReadTooLarge{got: len(resp.Body), limit: w.maxBytes}
	}
	decoded, err := decodeDirectReadResponse(w.outputPolicy, resp.Body, w.maxBytes)
	if err != nil {
		return nil, connectors.DirectReadPage{}, resp, err
	}

	// Two separate questions, deliberately not one boolean. Whether the
	// strategy addresses pages by NUMBER decides which navigation field the
	// result may carry at all — an addressable strategy never hands back a
	// cursor, because validateDirectReadPageRequest would refuse that cursor on
	// the way back in. Whether the CALLER chose this window decides only
	// whether the engine may name a number for it: the paginator still starts
	// at its own first page, so reporting number 1 for a request that asked for
	// startIndex=100 is a measured-looking lie.
	addressableStrategy := mode.pageable && isAddressableStrategy(strategy)
	engineChosePage := addressableStrategy && !callerNavigated

	page := connectors.DirectReadPage{Strategy: strategy, Size: sizeSent}
	if engineChosePage {
		page.Number = number
	}
	collection, isList, ambiguous := directReadCollectionKey(decoded)
	if ambiguous {
		// Which array pages cannot be known, so this read will not guess. What
		// it must not do is claim zero rows for a body that plainly carries
		// them, nor drop the page context the caller needs to ask for more.
		page.Records = directReadAmbiguousRecordCount(decoded)
		page.Complete = false
		page.Reason = directReadReasonAmbiguous
		return decoded, page, resp, nil
	}
	if !isList {
		// A single object is the whole answer; there is no collection to page.
		page.Size = 0
		page.Number = 0
		page.Complete = true
		return decoded, page, resp, nil
	}
	items, _ := directReadItemsAt(decoded, collection)
	page.Records = len(items)

	if !mode.pageable {
		// One request is all the engine can honestly do, and it cannot prove
		// this is the whole collection — so it says so, naming which of the
		// declaration states it is in.
		page.Complete = false
		page.Reason = mode.reason
		return decoded, page, resp, nil
	}

	// The declared strategy itself decides whether another page exists: its own
	// Next() is the single source of truth, exactly as it is for the ETL path.
	if lrc, ok := paginator.(*lastRecordCursor); ok && collection != "" {
		lrc.recordsPath = collection
	}
	next := paginator.Next(resp, len(items))
	if guard, ok := paginator.(interface{ Err() error }); ok {
		if err := guard.Err(); err != nil {
			return nil, connectors.DirectReadPage{}, resp, errDirectReadPagination{err: err}
		}
	}
	if next == nil {
		if isAddressableStrategy(strategy) && sizeSent != pageSize {
			// page_number and offset_limit stop on a SHORT page: fewer records
			// than the page size means the collection ended. That inference is
			// only sound when the size the paginator compared against is the
			// size the request actually carried. When the declared spec names
			// no size/limit param nothing was sent at all (sizeSent is 0) and
			// the provider chose the page, so a "short" page proves nothing;
			// any other mismatch means the threshold and the wire disagree.
			// Asserting completeness from either is the false claim this
			// contract exists to remove.
			page.Complete = false
			page.Reason = directReadReasonSizeNotSent
			return decoded, page, resp, nil
		}
		page.Complete = true
		return decoded, page, resp, nil
	}

	page.HasMore = true
	page.Complete = false
	page.Reason = directReadReasonMorePages
	switch {
	case engineChosePage:
		page.NextNumber = number + 1
	case addressableStrategy:
		// The caller drives this walk through the connector's own paging
		// parameter. The paginator's next window is computed from an offset it
		// never advanced, so it names a page behind the caller — and a cursor
		// is not an input this strategy accepts anyway.
	default:
		if token := nextCursorToken(spec, next); token != "" {
			page.NextCursor = token
		}
	}
	return decoded, page, resp, nil
}

// effectiveDirectReadPageSize resolves the page size this request will actually
// carry, given that a caller's own size flag overrides the declared default.
//
// A value that is not a usable positive size leaves the declared size in place
// for the paginator; directReadRequestedSize then reads 0 back off the wire, the
// two disagree, and completeness is not asserted — which is the honest outcome
// for a size nothing can interpret.
func effectiveDirectReadPageSize(spec PaginationSpec, declared int, query url.Values) int {
	if size := directReadRequestedSize(spec, strings.TrimSpace(spec.Type), query); size > 0 {
		return size
	}
	return declared
}

// pagingParamsForStrategy names the query parameters the declared strategy
// itself writes, split by role.
//
// The distinction is load-bearing. A SIZE parameter is a default: a caller who
// names it is asking for a different page size, which the engine can simply
// honour and then report back from the wire. A NAVIGATION parameter selects
// WHICH page, so a caller naming it at the same time as --page/--page-cursor
// has asked for two different pages in one request, and only a refusal is
// honest.
func pagingParamsForStrategy(spec PaginationSpec, strategy string) (navigation, size []string) {
	switch strategy {
	case "page_number":
		return []string{spec.PageParam}, []string{spec.SizeParam}
	case "offset_limit":
		return []string{spec.OffsetParam}, []string{spec.LimitParam, spec.SizeParam}
	case "cursor":
		return []string{spec.CursorParam}, []string{spec.SizeParam}
	case "start_index":
		return []string{valueOrDefault(spec.StartIndexParam, defaultStartIndexParam)},
			[]string{valueOrDefault(spec.CountParam, defaultStartIndexCount), spec.SizeParam}
	case "next_url":
		// A next URL is still the sole public navigation channel. These names
		// are only the provider-owned query controls the engine may put on page
		// one and admit from a returned continuation URL.
		return []string{spec.OffsetParam}, []string{spec.SizeParam, spec.LimitParam}
	case "link_header":
		return nil, []string{spec.SizeParam}
	default:
		return nil, nil
	}
}

// callerPagingParams reports which of the strategy's navigation parameters the
// caller set directly through the command's own flags.
func callerPagingParams(spec PaginationSpec, strategy string, query url.Values) []string {
	navigation, _ := pagingParamsForStrategy(spec, strategy)
	var out []string
	for _, param := range navigation {
		if param == "" {
			continue
		}
		if query.Get(param) != "" {
			out = append(out, param)
		}
	}
	sort.Strings(out)
	return out
}

// refuseConflictingPagingInput rejects the one pairing the caller-wins rule
// cannot resolve: a raw paging parameter set alongside --page or --page-cursor.
// Preferring either one would answer a question the caller did not ask.
//
// The parameter is named as the REQUEST PARAMETER it is, never dressed up as a
// flag: the engine knows the query name a command's flag maps onto, not the
// flag, and bahmni's `--start-index` maps onto `startIndex`. Rendering a flag
// from the query name printed `--startIndex`, which no caller can type.
func refuseConflictingPagingInput(supplied []string, w directReadWalk, strategy string) error {
	if len(supplied) == 0 || (w.page == 0 && w.pageCursor == "") {
		return nil
	}
	flag := "--page-cursor"
	if w.page > 0 {
		flag = "--page"
	}
	return fmt.Errorf("direct read received %s and a command flag setting the request parameter %q, which the declared %q pagination uses to select a page; they select different pages, so pass one of them", flag, supplied[0], strategy)
}

// validateDirectReadPageRequest refuses navigation input the declared strategy
// cannot honour, rather than accepting it and silently returning page one — the
// same class of quiet wrongness this whole change exists to remove.
func validateDirectReadPageRequest(mode directReadPageMode, w directReadWalk) error {
	if w.page > 0 && w.pageCursor != "" {
		return fmt.Errorf("direct read accepts a page number or a page cursor, not both")
	}
	if w.page < 0 {
		return fmt.Errorf("direct read page must be a positive page number, got %d", w.page)
	}
	if !mode.pageable {
		if w.page > 1 || w.pageCursor != "" {
			return fmt.Errorf("direct read cannot address page %s: %s", describeRequestedPage(w), mode.describe())
		}
		return nil
	}
	if w.page > 1 && !isAddressableStrategy(mode.strategy) {
		return fmt.Errorf("direct read pagination strategy %q has no addressable page number; pass the previous page's next_cursor instead", mode.strategy)
	}
	if w.pageCursor != "" && isAddressableStrategy(mode.strategy) {
		return fmt.Errorf("direct read pagination strategy %q addresses pages by number; pass a page number instead of a cursor", mode.strategy)
	}
	return nil
}

func describeRequestedPage(w directReadWalk) string {
	if w.pageCursor != "" {
		return "by cursor"
	}
	return strconv.Itoa(w.page)
}

// admitDirectReadCursorURL is the endpoint admission for the one navigation
// input that is an absolute URL.
//
// DirectRead rejects an absolute req.Path and admits the relative one against
// api_surface / the operation ledger. A next_url or link_header cursor arrives
// after both of those checks and becomes the request target, so without this
// it would be a generic authenticated GET against any same-origin path — the
// passthrough this connector layer deliberately does not offer.
//
// Two rules, both derived from the request that was already admitted: the
// cursor must be same-origin with the resolved base URL, and it must address
// the SAME endpoint. allow_cross_host governs a provider-supplied Link header
// discovered during a walk; it can never widen what a caller may type, so it
// is deliberately not consulted here.
func admitDirectReadCursorURL(baseURL, requestPath, cursor string, spec PaginationSpec, strategy string, callerQuery url.Values) (string, error) {
	if !isAbsoluteHTTPURL(cursor) {
		return "", fmt.Errorf("page cursor must be the absolute next-page URL a previous page reported, got %q", cursor)
	}
	next, err := url.Parse(cursor)
	if err != nil {
		return "", fmt.Errorf("page cursor %q is not a valid URL", cursor)
	}
	if next.User != nil {
		return "", fmt.Errorf("page cursor must not carry userinfo")
	}
	target, err := directReadRequestTarget(baseURL, requestPath)
	if err != nil || target.Host == "" {
		return "", fmt.Errorf("page cursor cannot be admitted: this connector has no resolvable base origin to check it against")
	}
	if !strings.EqualFold(next.Host, target.Host) || !strings.EqualFold(next.Scheme, target.Scheme) {
		return "", fmt.Errorf("page cursor %q is not same-origin with %s://%s", cursor, target.Scheme, target.Host)
	}
	if directReadPathKey(next.EscapedPath()) != directReadPathKey(target.EscapedPath()) {
		return "", fmt.Errorf("page cursor %q addresses %q, but this command is admitted only for %q; a page cursor may vary the query, not the endpoint", cursor, directReadPathKey(next.EscapedPath()), directReadPathKey(target.EscapedPath()))
	}
	values, err := url.ParseQuery(next.RawQuery)
	if err != nil {
		return "", fmt.Errorf("page cursor has invalid query encoding: %w", err)
	}
	allowed := cursorURLAllowedQueryKeys(spec, strategy, callerQuery)
	for name, entries := range values {
		if _, ok := allowed[name]; !ok {
			return "", fmt.Errorf("page cursor query parameter %q is not declared by this command's continuation contract", name)
		}
		if len(entries) != 1 {
			return "", fmt.Errorf("page cursor repeats continuation query parameter %q", name)
		}
		if err := safety.RejectDangerousChars(name, "page cursor query parameter"); err != nil {
			return "", err
		}
		if err := safety.RejectDangerousChars(entries[0], "page cursor query value"); err != nil {
			return "", err
		}
	}
	return cursor, nil
}

// cursorURLAllowedQueryKeys has one source of authority: the fixed command's
// declared query flags plus the declared pagination controls. Link-header and
// next-URL pagination use `page` as their provider-owned continuation key
// when an older source declaration has no explicit CursorParam; this preserves
// the established closed link-header path without admitting arbitrary query
// authority.
func cursorURLAllowedQueryKeys(spec PaginationSpec, strategy string, callerQuery url.Values) map[string]struct{} {
	allowed := make(map[string]struct{}, len(callerQuery)+4)
	for name := range callerQuery {
		allowed[name] = struct{}{}
	}
	navigation, size := pagingParamsForStrategy(spec, strategy)
	for _, name := range append(navigation, size...) {
		if strings.TrimSpace(name) != "" {
			allowed[name] = struct{}{}
		}
	}
	if strategy == "link_header" || strategy == "next_url" {
		name := strings.TrimSpace(spec.CursorParam)
		if name == "" {
			name = "page"
		}
		allowed[name] = struct{}{}
	}
	return allowed
}

// directReadRequestTarget reproduces connsdk Requester.resolveURL's base+path
// join, so the endpoint a cursor is compared against is exactly the one this
// request would otherwise have hit.
func directReadRequestTarget(baseURL, requestPath string) (*url.URL, error) {
	raw := requestPath
	if !isAbsoluteHTTPURL(requestPath) {
		raw = strings.TrimRight(baseURL, "/") + "/" + strings.TrimLeft(requestPath, "/")
	}
	return url.Parse(raw)
}

func directReadPathKey(path string) string {
	trimmed := strings.TrimRight(path, "/")
	if trimmed == "" {
		return "/"
	}
	return trimmed
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

// directReadRequestedSize reads back the page size this request actually
// carries, from the final query rather than from the declaration.
//
// The declared-or-defaulted size is NOT the size on the wire: declaredSizeQuery
// skips an empty size_param, and the page_number/offset_limit/start_index
// paginators skip an empty size/limit/count param too. Reporting the
// declaration for those bundles would claim a page the provider never applied.
func directReadRequestedSize(spec PaginationSpec, strategy string, query url.Values) int {
	for _, param := range directReadSizeParams(spec, strategy) {
		if param == "" {
			continue
		}
		raw := query.Get(param)
		if raw == "" {
			continue
		}
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			return n
		}
	}
	return 0
}

// directReadSizeParams names every param through which a strategy can put a
// page size on the wire, mirroring the paginators in paginate.go.
func directReadSizeParams(spec PaginationSpec, strategy string) []string {
	_, size := pagingParamsForStrategy(spec, strategy)
	return size
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

// errDirectReadTooLarge lets the shared pager report the pre-existing
// oversize-response failure while each executor keeps its own message prefix.
type errDirectReadTooLarge struct {
	got   int
	limit int
}

func (e errDirectReadTooLarge) Error() string {
	return fmt.Sprintf("response too large: %d bytes exceeds limit %d", e.got, e.limit)
}

// errDirectReadPagination marks a failure that belongs to pagination — a
// refused navigation input, an unadmittable page cursor, or a paginator guard
// that fired. Both executors report it as such: a guard failure arrives with
// the response already fetched and decoded, so without this marker it was
// wrapped as "direct read response is not JSON", which is a wrong diagnosis of
// a body that parsed fine.
type errDirectReadPagination struct {
	err error
}

func (e errDirectReadPagination) Error() string { return e.err.Error() }

func (e errDirectReadPagination) Unwrap() error { return e.err }

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

// directReadAmbiguousRecordCount totals every top-level array element of an
// envelope whose paged collection could not be identified. Which array pages
// is unknown, but how many rows came back is not — and reporting 0 for a body
// carrying hundreds is exactly the untruth this change exists to remove.
func directReadAmbiguousRecordCount(decoded any) int {
	obj, ok := decoded.(map[string]any)
	if !ok {
		return 0
	}
	total := 0
	for _, value := range obj {
		if items, ok := value.([]any); ok {
			total += len(items)
		}
	}
	return total
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
