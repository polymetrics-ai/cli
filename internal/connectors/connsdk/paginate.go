package connsdk

import (
	"context"
	"net/url"
	"strconv"
	"strings"
)

// NextPage describes how to request the page after the current one. Query params
// are merged into the base request params. URL, when set, overrides the path
// entirely (used for Link-header style absolute "next" URLs).
type NextPage struct {
	Query url.Values
	URL   string
}

// Paginator drives multi-page reads. Start returns the params for the first page;
// Next inspects each response and returns the next page, or nil when exhausted.
type Paginator interface {
	Start() *NextPage
	Next(resp *Response, recordCount int) *NextPage
}

// OffsetPaginator advances by offset until a short page is returned.
type OffsetPaginator struct {
	LimitParam  string // e.g. "limit"
	OffsetParam string // e.g. "offset"
	PageSize    int    // page size; also the stop threshold
	offset      int
}

func (p *OffsetPaginator) Start() *NextPage {
	p.offset = 0
	return &NextPage{Query: offsetQuery(p.LimitParam, p.PageSize, p.OffsetParam, 0)}
}

func (p *OffsetPaginator) Next(_ *Response, recordCount int) *NextPage {
	if p.PageSize <= 0 || recordCount < p.PageSize {
		return nil
	}
	p.offset += p.PageSize
	return &NextPage{Query: offsetQuery(p.LimitParam, p.PageSize, p.OffsetParam, p.offset)}
}

func offsetQuery(limitParam string, size int, offsetParam string, offset int) url.Values {
	q := url.Values{}
	if limitParam != "" && size > 0 {
		q.Set(limitParam, strconv.Itoa(size))
	}
	if offsetParam != "" {
		q.Set(offsetParam, strconv.Itoa(offset))
	}
	return q
}

// PageNumberPaginator advances by page number until a short page is returned.
type PageNumberPaginator struct {
	PageParam string // e.g. "page"
	SizeParam string // e.g. "per_page" (optional)
	StartPage int    // usually 1
	PageSize  int    // page size; also the stop threshold
	page      int
}

func (p *PageNumberPaginator) Start() *NextPage {
	p.page = p.StartPage
	if p.page == 0 {
		p.page = 1
	}
	return &NextPage{Query: pageQuery(p.PageParam, p.page, p.SizeParam, p.PageSize)}
}

func (p *PageNumberPaginator) Next(_ *Response, recordCount int) *NextPage {
	if p.PageSize <= 0 || recordCount < p.PageSize {
		return nil
	}
	p.page++
	return &NextPage{Query: pageQuery(p.PageParam, p.page, p.SizeParam, p.PageSize)}
}

func pageQuery(pageParam string, page int, sizeParam string, size int) url.Values {
	q := url.Values{}
	if pageParam != "" {
		q.Set(pageParam, strconv.Itoa(page))
	}
	if sizeParam != "" && size > 0 {
		q.Set(sizeParam, strconv.Itoa(size))
	}
	return q
}

// CursorPaginator reads the next page token from the response body at TokenPath
// and supplies it as CursorParam on the following request.
type CursorPaginator struct {
	CursorParam string // query param carrying the cursor, e.g. "cursor"
	TokenPath   string // dotted path to next token in body, e.g. "meta.next_cursor"
	// FirstQuery are params for the very first request (optional).
	FirstQuery url.Values
}

func (p *CursorPaginator) Start() *NextPage {
	return &NextPage{Query: cloneValues(p.FirstQuery)}
}

func (p *CursorPaginator) Next(resp *Response, _ int) *NextPage {
	if resp == nil {
		return nil
	}
	token, err := StringAt(resp.Body, p.TokenPath)
	if err != nil || strings.TrimSpace(token) == "" {
		return nil
	}
	q := url.Values{}
	q.Set(p.CursorParam, token)
	return &NextPage{Query: q}
}

// LinkHeaderPaginator follows the RFC 5988 Link header rel="next" (used by GitHub
// and many others).
type LinkHeaderPaginator struct {
	FirstQuery url.Values
}

func (p *LinkHeaderPaginator) Start() *NextPage {
	return &NextPage{Query: cloneValues(p.FirstQuery)}
}

func (p *LinkHeaderPaginator) Next(resp *Response, _ int) *NextPage {
	if resp == nil {
		return nil
	}
	next := parseLinkNext(resp.Header.Get("Link"))
	if next == "" {
		return nil
	}
	return &NextPage{URL: next}
}

// parseLinkNext extracts the rel="next" URL from a Link header value.
func parseLinkNext(header string) string {
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
		for _, attr := range segs[1:] {
			attr = strings.TrimSpace(attr)
			if attr == `rel="next"` || attr == "rel=next" {
				isNext = true
			}
		}
		if isNext {
			return urlPart[1 : len(urlPart)-1]
		}
	}
	return ""
}

func cloneValues(in url.Values) url.Values {
	out := url.Values{}
	for k, vs := range in {
		for _, v := range vs {
			out.Add(k, v)
		}
	}
	return out
}

func mergeValues(base, extra url.Values) url.Values {
	out := cloneValues(base)
	for k, vs := range extra {
		out.Del(k)
		for _, v := range vs {
			out.Add(k, v)
		}
	}
	return out
}

// Harvest reads every page of a paginated endpoint, extracting records at
// recordsPath and invoking emit for each. It stops when the paginator signals no
// next page, when maxPages is reached (0 = unlimited), or when ctx is cancelled.
//
// base holds query params common to all pages (filters, page size). method/path
// are the source endpoint.
func Harvest(ctx context.Context, r *Requester, method, path string, base url.Values, p Paginator, recordsPath string, maxPages int, emit func(Record) error) error {
	page := p.Start()
	for pageNum := 0; page != nil; pageNum++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if maxPages > 0 && pageNum >= maxPages {
			return nil
		}

		reqPath := path
		if page.URL != "" {
			reqPath = page.URL
		}
		query := mergeValues(base, page.Query)

		resp, err := r.Do(ctx, method, reqPath, query, nil)
		if err != nil {
			return err
		}
		records, err := RecordsAt(resp.Body, recordsPath)
		if err != nil {
			return err
		}
		for _, rec := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(rec); err != nil {
				return err
			}
		}
		page = p.Next(resp, len(records))
	}
	return nil
}
