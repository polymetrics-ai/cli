package engine

import (
	"fmt"
	"net/url"
	"strings"

	"polymetrics.ai/internal/connectors/connsdk"
)

// newPaginator maps a bundle PaginationSpec onto a connsdk.Paginator. Four of
// the six types reuse the existing connsdk constructors as-is; the other two
// (cursor's last_record_field variant and next_url) are engine-local
// implementations defined below that satisfy the same connsdk.Paginator
// contract — connsdk itself is not modified.
//
// The next_url guard (THREAT-MODEL §3) needs the resolved base URL's host to
// enforce same-host-unless-allow_cross_host; newPaginator has no requester
// in scope, so callers building a next_url paginator for real reads (read.go,
// wave C) must set the returned *nextURL's BaseHost field from
// requester.BaseURL before the first Harvest call. Tests in this file do the
// same against their httptest.Server's own host.
func newPaginator(spec PaginationSpec, pageSize int) (connsdk.Paginator, error) {
	size := spec.PageSize
	if size == 0 {
		size = pageSize
	}

	switch spec.Type {
	case "", "none":
		return &nonePaginator{}, nil

	case "link_header":
		return &connsdk.LinkHeaderPaginator{}, nil

	case "page_number":
		start := spec.StartPage
		if start == 0 {
			start = 1
		}
		return &connsdk.PageNumberPaginator{
			PageParam: spec.PageParam,
			SizeParam: spec.SizeParam,
			StartPage: start,
			PageSize:  size,
		}, nil

	case "offset_limit":
		return &connsdk.OffsetPaginator{
			LimitParam:  spec.LimitParam,
			OffsetParam: spec.OffsetParam,
			PageSize:    size,
		}, nil

	case "cursor":
		return newCursorPaginator(spec)

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
func newCursorPaginator(spec PaginationSpec) (connsdk.Paginator, error) {
	hasToken := spec.TokenPath != ""
	hasLastRecord := spec.LastRecordField != ""

	switch {
	case hasToken && hasLastRecord:
		return nil, fmt.Errorf("new paginator: cursor: token_path and last_record_field are mutually exclusive")
	case hasToken:
		return &connsdk.CursorPaginator{
			CursorParam: spec.CursorParam,
			TokenPath:   spec.TokenPath,
		}, nil
	case hasLastRecord:
		return &lastRecordCursor{
			cursorParam:     spec.CursorParam,
			lastRecordField: spec.LastRecordField,
			stopPath:        spec.StopPath,
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

// lastRecordCursor implements the stripe starting_after/has_more pagination
// shape (today hand-written at internal/connectors/stripe/stripe.go:147):
// the next page's cursor param is the value of lastRecordField on the LAST
// record of the current page's "data" list. stopPath, when set, names a
// body path whose falsy value stops pagination (stripe's has_more); its
// absence is treated as "continue" only if a last-record id was actually
// found (defensive: an empty or malformed page can never advance the
// cursor, so it can never loop forever regardless of what the API claims).
type lastRecordCursor struct {
	cursorParam     string
	lastRecordField string
	stopPath        string
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

	lastID, ok := lastRecordFieldValue(resp.Body, p.lastRecordField)
	if !ok || lastID == "" {
		return nil
	}

	q := url.Values{}
	q.Set(p.cursorParam, lastID)
	return &connsdk.NextPage{Query: q}
}

// lastRecordFieldValue extracts the value of field from the LAST element of
// the response body's top-level "data" array (the stripe/list-response
// convention this paginator targets). ok is false when the body has no
// records or the field is absent (or null) on the last one.
func lastRecordFieldValue(body []byte, field string) (string, bool) {
	records, err := connsdk.RecordsAt(body, "data")
	if err != nil || len(records) == 0 {
		return "", false
	}
	last := records[len(records)-1]
	v, present := last[field]
	if !present || v == nil {
		return "", false
	}
	s, ok := v.(string)
	if !ok || s == "" {
		return "", false
	}
	return s, true
}

// nextURL implements pagination.type == "next_url" (aircall style): the
// next page's absolute URL is read from a body path on the current
// response. It enforces the SSRF guard from THREAT-MODEL §3 — a next URL
// whose host differs from BaseHost is rejected unless allowCrossHost is set
// — and a loop guard that refuses to follow the same URL twice (defends
// against a hostile/buggy API looping pagination forever).
//
// BaseHost is exported (despite the type being unexported) so callers that
// construct a next_url paginator via newPaginator can set it from the
// resolved requester.BaseURL host before the first Harvest call; newPaginator
// itself has no requester in scope to derive it automatically.
type nextURL struct {
	BaseHost string

	path           string
	allowCrossHost bool

	seen    map[string]bool
	lastErr error
}

func (p *nextURL) Start() *connsdk.NextPage {
	p.seen = map[string]bool{}
	p.lastErr = nil
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

	if !p.allowCrossHost && p.BaseHost != "" {
		if host := urlHost(next); host != "" && host != p.BaseHost {
			p.lastErr = fmt.Errorf("next_url: cross-host redirect to %q blocked (base host %q); set allow_cross_host: true to permit", next, p.BaseHost)
			return nil
		}
	}

	if p.seen[next] {
		p.lastErr = fmt.Errorf("next_url: loop detected — %q requested twice", next)
		return nil
	}
	p.seen[next] = true

	return &connsdk.NextPage{URL: next}
}

// Err returns the sticky error from the most recent guard violation (cross-
// host block or loop detection), or nil when pagination stopped for a
// benign reason (absent/empty next_url value). connsdk.Paginator.Next has no
// error return, so callers that need to distinguish "no more pages" from
// "a guard blocked further pagination" call Err() after Harvest returns.
func (p *nextURL) Err() error { return p.lastErr }

func urlHost(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	return u.Host
}
