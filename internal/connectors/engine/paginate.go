package engine

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"polymetrics.ai/internal/connectors/connsdk"
)

// baseHostSetter is satisfied by every engine-local SSRF-guarded paginator
// (nextURL, linkHeaderPaginator) so callers (read.go, and test helpers) can
// wire the resolved requester base origin generically instead of
// type-switching on the concrete paginator type.
type baseHostSetter interface {
	setBaseOrigin(scheme, host string)
}

// newPaginator maps a bundle PaginationSpec onto a connsdk.Paginator. Four of
// the six types reuse the existing connsdk constructors as-is; the other two
// (cursor's last_record_field variant and next_url) plus link_header's
// SSRF-guarded wrapper are engine-local implementations defined below that
// satisfy the same connsdk.Paginator contract — connsdk itself is not
// modified.
//
// recordsPath is the stream's effective RecordsSpec.Path (F3, REVIEW.md):
// lastRecordCursor needs to know where the records envelope lives in a page
// body to find the LAST record's cursor field, and hardcoding "data" (the
// stripe shape) silently truncates any bundle whose list key differs.
//
// The next_url/link_header guards (THREAT-MODEL §3, SECURITY-REVIEW.md M1)
// need the resolved base URL's scheme+host to enforce
// same-origin-unless-allow_cross_host; newPaginator has no requester in
// scope, so callers building one of these paginators for real reads
// (read.go) must call setBaseOrigin on the returned paginator (both
// implement baseHostSetter) from the resolved requester.BaseURL before the
// first Harvest call. Tests in this file do the same against their
// httptest.Server's own host.
func newPaginator(spec PaginationSpec, pageSize int, recordsPath string) (connsdk.Paginator, error) {
	size := spec.PageSize
	if size == 0 {
		size = pageSize
	}

	switch spec.Type {
	case "", "none":
		return &nonePaginator{}, nil

	case "link_header":
		return &linkHeaderPaginator{allowCrossHost: spec.AllowCrossHost}, nil

	case "page_number":
		return &pageNumberPaginator{
			pageParam: spec.PageParam,
			sizeParam: spec.SizeParam,
			startPage: startPageOrDefault(spec.StartPage),
			pageSize:  size,
		}, nil

	case "offset_limit":
		return &connsdk.OffsetPaginator{
			LimitParam:  spec.LimitParam,
			OffsetParam: spec.OffsetParam,
			PageSize:    size,
		}, nil

	case "cursor":
		return newCursorPaginator(spec, recordsPath)

	case "next_url":
		return newNextURLPaginator(spec)

	default:
		return nil, fmt.Errorf("new paginator: unknown pagination type %q", spec.Type)
	}
}

// newCursorPaginator builds the paginator for pagination.type == "cursor",
// which supports exactly one of two mutually exclusive token sources:
// token_path (a next-page token read from the response body, delegated to
// connsdk.CursorPaginator) or last_record_field (+ optional stop_path), the
// stripe starting_after/has_more shape implemented locally as
// lastRecordCursor.
func newCursorPaginator(spec PaginationSpec, recordsPath string) (connsdk.Paginator, error) {
	hasToken := spec.TokenPath != ""
	hasLastRecord := spec.LastRecordField != ""

	switch {
	case hasToken && hasLastRecord:
		return nil, fmt.Errorf("new paginator: cursor: token_path and last_record_field are mutually exclusive")
	case hasToken:
		return &tokenPathCursor{
			cursorParam: spec.CursorParam,
			tokenPath:   spec.TokenPath,
			stopPath:    spec.StopPath,
		}, nil
	case hasLastRecord:
		path := recordsPath
		if strings.TrimSpace(path) == "" {
			// Zero-value fallback for any construction path that never wires an
			// effective records path (e.g. an ad hoc caller): "data" matches the
			// paginator's original stripe-only behavior, so nothing that relied
			// on the old hardcoded default silently changes.
			path = "data"
		}
		return &lastRecordCursor{
			cursorParam:     spec.CursorParam,
			lastRecordField: spec.LastRecordField,
			stopPath:        spec.StopPath,
			recordsPath:     path,
		}, nil
	default:
		return nil, fmt.Errorf("new paginator: cursor: exactly one of token_path or last_record_field is required")
	}
}

func newNextURLPaginator(spec PaginationSpec) (connsdk.Paginator, error) {
	if strings.TrimSpace(spec.NextURLPath) == "" {
		return nil, fmt.Errorf("new paginator: next_url: next_url_path is required")
	}
	return &nextURL{
		path:           spec.NextURLPath,
		allowCrossHost: spec.AllowCrossHost,
	}, nil
}

// nonePaginator issues exactly one request: Start returns the first (and
// only) page; Next always signals exhaustion.
type nonePaginator struct{}

func (p *nonePaginator) Start() *connsdk.NextPage { return &connsdk.NextPage{} }

func (p *nonePaginator) Next(*connsdk.Response, int) *connsdk.NextPage { return nil }

// startPageOrDefault returns the effective start page for a page_number
// paginator (S4 engine mini-wave item 1): a nil StartPage (never declared in
// streams.json) defaults to 1, matching every pre-existing bundle's implicit
// assumption; a non-nil pointer — INCLUDING one pointing at 0 — is honored
// verbatim, since only the pointer's nil-ness (not its pointee's value)
// signals "unset".
func startPageOrDefault(startPage *int) int {
	if startPage == nil {
		return 1
	}
	return *startPage
}

// pageNumberPaginator is an engine-local reimplementation of
// connsdk.PageNumberPaginator's exact Start/Next shape (S4 engine mini-wave
// item 1), differing in exactly one respect: it honors an explicitly-0
// startPage instead of unconditionally coercing 0 to 1
// (connsdk.PageNumberPaginator.Start() does `p.page = p.StartPage; if
// p.page == 0 { p.page = 1 }`, which cannot distinguish "explicitly 0" from
// "the Go zero value because it was never set" — exactly the ambiguity a
// pointer-typed PaginationSpec.StartPage now resolves one layer up).
// connsdk itself is intentionally not modified (out of scope; every other
// connsdk.PageNumberPaginator caller, all legacy Go connectors, keeps its
// existing 1-coercion behavior unchanged).
type pageNumberPaginator struct {
	pageParam string
	sizeParam string
	startPage int
	pageSize  int
	page      int
}

func (p *pageNumberPaginator) Start() *connsdk.NextPage {
	p.page = p.startPage
	return &connsdk.NextPage{Query: pageNumberQuery(p.pageParam, p.page, p.sizeParam, p.pageSize)}
}

func (p *pageNumberPaginator) Next(_ *connsdk.Response, recordCount int) *connsdk.NextPage {
	if p.pageSize <= 0 || recordCount < p.pageSize {
		return nil
	}
	p.page++
	return &connsdk.NextPage{Query: pageNumberQuery(p.pageParam, p.page, p.sizeParam, p.pageSize)}
}

// pageNumberQuery mirrors connsdk.paginate.go's unexported pageQuery helper
// exactly (duplicated here since that helper is unexported and connsdk is
// read-only in this task).
func pageNumberQuery(pageParam string, page int, sizeParam string, size int) url.Values {
	q := url.Values{}
	if pageParam != "" {
		q.Set(pageParam, strconv.Itoa(page))
	}
	if sizeParam != "" && size > 0 {
		q.Set(sizeParam, strconv.Itoa(size))
	}
	return q
}

// lastRecordCursor implements the stripe starting_after/has_more pagination
// shape (today hand-written at internal/connectors/stripe/stripe.go:147):
// the next page's cursor param is the value of lastRecordField on the LAST
// record of the current page's records envelope (recordsPath — F3: this is
// NOT hardcoded to "data" any more; it is wired from the stream's effective
// RecordsSpec.Path by newCursorPaginator). stopPath, when set, names a body
// path whose falsy value stops pagination (stripe's has_more); its absence
// is treated as "continue" only if a last-record id was actually found
// (defensive: an empty or malformed page can never advance the cursor, so it
// can never loop forever regardless of what the API claims).
type lastRecordCursor struct {
	cursorParam     string
	lastRecordField string
	stopPath        string
	recordsPath     string
}

func (p *lastRecordCursor) Start() *connsdk.NextPage {
	return &connsdk.NextPage{Query: url.Values{}}
}

func (p *lastRecordCursor) Next(resp *connsdk.Response, recordCount int) *connsdk.NextPage {
	if resp == nil || recordCount == 0 {
		return nil
	}

	if p.stopPath != "" {
		stopVal, err := connsdk.StringAt(resp.Body, p.stopPath)
		if err != nil || stopVal != "true" {
			return nil
		}
	}

	path := p.recordsPath
	if strings.TrimSpace(path) == "" {
		path = "data"
	}
	lastID, ok := lastRecordFieldValue(resp.Body, path, p.lastRecordField)
	if !ok || lastID == "" {
		return nil
	}

	q := url.Values{}
	q.Set(p.cursorParam, lastID)
	return &connsdk.NextPage{Query: q}
}

// lastRecordFieldValue extracts the value of field from the LAST element of
// the response body's recordsPath-addressed array (F3: recordsPath is the
// stream's own records envelope location, no longer hardcoded to "data").
// ok is false when the body has no records or the field is absent (or null)
// on the last one, or its value is not a stringifiable id (see
// stringifyLastRecordID).
func lastRecordFieldValue(body []byte, recordsPath, field string) (string, bool) {
	records, err := connsdk.RecordsAt(body, recordsPath)
	if err != nil || len(records) == 0 {
		return "", false
	}
	last := records[len(records)-1]
	v, present := last[field]
	if !present || v == nil {
		return "", false
	}
	return stringifyLastRecordID(v)
}

// stringifyLastRecordID converts a last-record cursor field's raw value into
// the string used as the next page's cursor param (F3, REVIEW.md: the prior
// version only accepted a Go string, silently stopping pagination for any
// API whose ids are numeric). A string must be non-empty. A JSON number is
// stringified canonically via its own text form (connsdk decodes with
// UseNumber, so json.Number is the shape actually reachable through
// lastRecordFieldValue's connsdk.RecordsAt call); float64 is handled too,
// defensively, for any caller that constructs a records map by hand (e.g. a
// unit test or a future non-UseNumber decode path) rather than through
// connsdk's JSON extraction, formatted without a redundant ".0" when the
// value is integral. Any other type is not a supported id shape.
func stringifyLastRecordID(v any) (string, bool) {
	switch t := v.(type) {
	case string:
		if t == "" {
			return "", false
		}
		return t, true
	case json.Number:
		return t.String(), true
	case float64:
		if t == float64(int64(t)) {
			return fmt.Sprintf("%d", int64(t)), true
		}
		return fmt.Sprintf("%v", t), true
	default:
		return "", false
	}
}

// tokenPathCursor implements pagination.type == "cursor" with the token_path
// token source (a next-page token read from the response body at a fixed
// dotted path, e.g. Zendesk's meta.after_cursor) — an engine-local
// reimplementation of connsdk.CursorPaginator's exact Start/Next shape, plus
// two additions connsdk's version does not have (gap-loop cycle-1 item 5,
// REVIEW-B.md zendesk-support finding 2; connsdk/paginate.go itself is
// outside this task's editable file set):
//
//  1. optional stopPath (mirrors lastRecordCursor's stripe-shape stopPath):
//     when set, a FALSY body value at that path stops pagination
//     unconditionally, even when the token itself is still non-empty.
//     Zendesk's own cursor-pagination docs warn the cursor properties "may
//     be populated even when has_more is false" — legacy stops on
//     `has_more != "true" || after_cursor == ""`, and without this the
//     engine would issue one extra (or infinitely looping, absent any other
//     guard) request past the real last page.
//  2. a loop guard refusing to follow the same token twice in a row
//     (defends against a hostile/buggy API that never actually advances the
//     cursor), mirroring nextURL's/linkHeaderPaginator's existing "seen" map
//     — connsdk.CursorPaginator has no such guard today.
//
// A stop_path body value is read via connsdk.StringAt exactly like
// lastRecordCursor's stopPath check: any value other than the literal string
// "true" (including a JSON `false`, a missing path, or a read error) is
// falsy and stops pagination. A spec that never sets stop_path (the
// zero-value/absent case) preserves the exact prior behavior: stop only on
// an absent/empty token, no stop_path check at all.
type tokenPathCursor struct {
	cursorParam string
	tokenPath   string
	stopPath    string

	seen    map[string]bool
	lastErr error
}

func (p *tokenPathCursor) Start() *connsdk.NextPage {
	p.seen = map[string]bool{}
	p.lastErr = nil
	return &connsdk.NextPage{}
}

func (p *tokenPathCursor) Next(resp *connsdk.Response, _ int) *connsdk.NextPage {
	if resp == nil {
		return nil
	}

	if p.stopPath != "" {
		stopVal, err := connsdk.StringAt(resp.Body, p.stopPath)
		if err != nil || stopVal != "true" {
			return nil
		}
	}

	token, err := connsdk.StringAt(resp.Body, p.tokenPath)
	if err != nil || strings.TrimSpace(token) == "" {
		return nil
	}

	if p.seen[token] {
		p.lastErr = fmt.Errorf("cursor(token_path): loop detected — token %q requested twice", token)
		return nil
	}
	p.seen[token] = true

	q := url.Values{}
	q.Set(p.cursorParam, token)
	return &connsdk.NextPage{Query: q}
}

// Err returns the sticky loop-detection error, or nil when pagination
// stopped for a benign reason (absent/empty token, or a falsy stop_path
// value) — mirrors nextURL.Err()/linkHeaderPaginator.Err().
func (p *tokenPathCursor) Err() error { return p.lastErr }

// baseOrigin is the (scheme, host) pair an SSRF-guarded paginator compares a
// discovered next-page URL against. Both nextURL and linkHeaderPaginator
// share this shape and its guard logic via checkOrigin below (M1/m2,
// SECURITY-REVIEW.md).
type baseOrigin struct {
	scheme string
	host   string
}

func (o baseOrigin) set() bool { return o.host != "" }

// checkOrigin enforces the same-origin-unless-allow_cross_host SSRF guard
// shared by nextURL and linkHeaderPaginator: an unparseable next URL FAILS
// CLOSED (m2: never silently allowed through just because its host could
// not be determined); once parsed, both scheme AND host must match base
// (m2: a same-host but http-downgraded-from-https next URL is rejected,
// not just a different host) unless allowCrossHost is set or base was never
// configured (base.set() == false — e.g. a test that never wired BaseHost;
// existing behavior for that case is preserved: no guard is enforced). Error
// messages keep the historical "cross-host" wording for a differing HOST
// (existing callers/tests match on that substring) and use a distinct
// "scheme downgrade" wording when only the scheme differs on the same host.
func checkOrigin(next string, base baseOrigin, allowCrossHost bool) error {
	if allowCrossHost || !base.set() {
		return nil
	}
	u, err := url.Parse(next)
	if err != nil {
		return fmt.Errorf("cross-host follow: next URL %q could not be parsed; rejecting (fail closed)", next)
	}
	if u.Host == "" {
		return fmt.Errorf("cross-host follow: next URL %q has no host; rejecting (fail closed)", next)
	}
	if u.Host != base.host {
		return fmt.Errorf("cross-host redirect to %q blocked (base host %q); set allow_cross_host: true to permit", next, base.host)
	}
	if u.Scheme != base.scheme {
		return fmt.Errorf("scheme downgrade to %q blocked (base origin %s://%s); set allow_cross_host: true to permit", next, base.scheme, base.host)
	}
	return nil
}

// nextURL implements pagination.type == "next_url" (aircall style): the
// next page's absolute URL is read from a body path on the current
// response. It enforces the SSRF guard from THREAT-MODEL §3 — a next URL
// whose origin (scheme+host) differs from the base is rejected unless
// allowCrossHost is set — and a loop guard that refuses to follow the same
// URL twice (defends against a hostile/buggy API looping pagination
// forever).
type nextURL struct {
	// BaseHost is kept as an exported field for backward-compatible direct
	// assignment (existing call sites set p.BaseHost = host directly); it is
	// mirrored into base.host by setBaseOrigin, which also carries scheme.
	// Setting BaseHost directly (without scheme) preserves prior
	// host-only-guard behavior only insofar as base.scheme stays "" — but
	// Next always parses the actual next URL's scheme, so a "" base scheme
	// simply never matches any real scheme and the guard degrades to
	// rejecting cross-scheme follows too; callers migrating to the new
	// scheme-aware guard should call setBaseOrigin instead.
	BaseHost string

	path           string
	allowCrossHost bool

	base baseOrigin

	seen    map[string]bool
	lastErr error
}

// setBaseOrigin sets the SSRF guard's expected scheme+host (read.go derives
// this from the resolved requester.BaseURL before the first Harvest call).
func (p *nextURL) setBaseOrigin(scheme, host string) {
	p.base = baseOrigin{scheme: scheme, host: host}
	p.BaseHost = host
}

func (p *nextURL) Start() *connsdk.NextPage {
	p.seen = map[string]bool{}
	p.lastErr = nil
	// BaseHost may have been set directly (legacy call sites) rather than via
	// setBaseOrigin; keep base.host in sync so the guard still fires even
	// without a known scheme (see BaseHost's doc comment).
	if p.BaseHost != "" && p.base.host == "" {
		p.base.host = p.BaseHost
	}
	return &connsdk.NextPage{}
}

func (p *nextURL) Next(resp *connsdk.Response, _ int) *connsdk.NextPage {
	if resp == nil {
		return nil
	}
	next, err := connsdk.StringAt(resp.Body, p.path)
	if err != nil || strings.TrimSpace(next) == "" {
		return nil
	}

	if guardErr := checkOrigin(next, p.base, p.allowCrossHost); guardErr != nil {
		p.lastErr = fmt.Errorf("next_url: %w", guardErr)
		return nil
	}

	if p.seen[next] {
		p.lastErr = fmt.Errorf("next_url: loop detected — %q requested twice", next)
		return nil
	}
	p.seen[next] = true

	return &connsdk.NextPage{URL: next}
}

// Err returns the sticky error from the most recent guard violation (cross-
// origin block or loop detection), or nil when pagination stopped for a
// benign reason (absent/empty next_url value). connsdk.Paginator.Next has no
// error return, so callers that need to distinguish "no more pages" from
// "a guard blocked further pagination" call Err() after Harvest returns.
func (p *nextURL) Err() error { return p.lastErr }

// linkHeaderPaginator wraps connsdk.LinkHeaderPaginator's RFC 5988
// Link-header-follow semantics with the SAME SSRF guard nextURL enforces
// (M1, SECURITY-REVIEW.md MAJOR): before this type existed, newPaginator's
// "link_header" case returned a bare *connsdk.LinkHeaderPaginator with NO
// host validation at all, so a compromised/malicious upstream could redirect
// pagination to an arbitrary internal host via a crafted Link response
// header — the exact vector THREAT-MODEL §3 claims is covered for every
// pagination type, but wasn't for this one (github-shaped connectors use
// link_header pagination).
type linkHeaderPaginator struct {
	allowCrossHost bool

	base baseOrigin

	seen    map[string]bool
	lastErr error
}

// setBaseOrigin sets the SSRF guard's expected scheme+host, mirroring
// nextURL.setBaseOrigin.
func (p *linkHeaderPaginator) setBaseOrigin(scheme, host string) {
	p.base = baseOrigin{scheme: scheme, host: host}
}

func (p *linkHeaderPaginator) Start() *connsdk.NextPage {
	p.seen = map[string]bool{}
	p.lastErr = nil
	return &connsdk.NextPage{}
}

func (p *linkHeaderPaginator) Next(resp *connsdk.Response, _ int) *connsdk.NextPage {
	if resp == nil {
		return nil
	}
	next := parseLinkNextHeader(resp.Header.Get("Link"))
	if next == "" {
		return nil
	}

	if guardErr := checkOrigin(next, p.base, p.allowCrossHost); guardErr != nil {
		p.lastErr = fmt.Errorf("link_header: %w", guardErr)
		return nil
	}

	if p.seen[next] {
		p.lastErr = fmt.Errorf("link_header: loop detected — %q requested twice", next)
		return nil
	}
	p.seen[next] = true

	return &connsdk.NextPage{URL: next}
}

// Err mirrors nextURL.Err(): the sticky guard-violation error, or nil for a
// benign stop (no Link header / no rel="next").
func (p *linkHeaderPaginator) Err() error { return p.lastErr }

// parseLinkNextHeader extracts the rel="next" URL from a Link header value,
// duplicating connsdk's own (unexported) parseLinkNext logic — connsdk
// itself is not modified (out of scope; see this file's package doc),
// wave0's parity/format is small enough that a one-off reimplementation here
// is preferable to changing connsdk's exported surface.
func parseLinkNextHeader(header string) string {
	if header == "" {
		return ""
	}
	for _, part := range strings.Split(header, ",") {
		segs := strings.Split(part, ";")
		if len(segs) < 2 {
			continue
		}
		urlPart := strings.TrimSpace(segs[0])
		if !strings.HasPrefix(urlPart, "<") || !strings.HasSuffix(urlPart, ">") {
			continue
		}
		isNext := false
		for _, param := range segs[1:] {
			param = strings.TrimSpace(param)
			if param == `rel="next"` || param == "rel=next" {
				isNext = true
				break
			}
		}
		if isNext {
			return urlPart[1 : len(urlPart)-1]
		}
	}
	return ""
}
