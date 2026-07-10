package conformance

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"path"
	"sort"
	"sync"
)

// fixturePage is one recorded API page: fixtures/streams/<stream>/page_N.json,
// shape {"request":{"method","path","query"},"read_query":{...},"response":{"status","body"}}.
// read_query is optional harness input for parameterized streams whose runtime
// request needs fixture-specific values. Keys matching non-secret spec.json
// properties override RuntimeConfig.Config; the remaining keys populate
// connectors.ReadRequest.Query (for example GraphQL command flags embedded in
// the POST body).
// This shape is already load-bearing in the committed
// internal/connectors/engine/testdata/bundles/widget-demo fixture (Wave A/B
// reference), so it is reused verbatim here rather than invented fresh.
type fixturePage struct {
	file      string            // relative path, for error messages / hit tracking
	Request   fixtureRequest    `json:"request"`
	ReadQuery map[string]string `json:"read_query,omitempty"`
	Response  fixtureResponse   `json:"response"`
}

type fixtureRequest struct {
	Method string            `json:"method"`
	Path   string            `json:"path"`
	Query  map[string]string `json:"query"`
}

type fixtureResponse struct {
	Status int             `json:"status"`
	Body   json.RawMessage `json:"body"`
}

// loadFixturePages reads and parses every fixtures/streams/<stream>/page_*.json
// file, sorted by filename (page_1 before page_2, ...) so pages replay in
// recording order. A stream with no fixtures/streams/<stream>/ directory at
// all returns an empty (non-error) slice — callers decide whether that's
// acceptable (fixtures_present requires at least one page for the first
// stream; other streams may legitimately have none in wave0 goldens that
// only fixture their primary stream).
func loadFixturePages(fixtures fs.FS, stream string) ([]fixturePage, error) {
	if fixtures == nil {
		return nil, nil
	}
	dir := path.Join("streams", stream)
	entries, err := fs.ReadDir(fixtures, dir)
	if err != nil {
		return nil, nil //nolint:nilerr // missing dir = "no fixtures for this stream", not an error
	}

	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)

	pages := make([]fixturePage, 0, len(names))
	for _, name := range names {
		raw, err := fs.ReadFile(fixtures, path.Join(dir, name))
		if err != nil {
			return nil, fmt.Errorf("read fixture %s: %w", path.Join(dir, name), err)
		}
		var page fixturePage
		if err := json.Unmarshal(raw, &page); err != nil {
			return nil, fmt.Errorf("parse fixture %s: %w", path.Join(dir, name), err)
		}
		page.file = name
		pages = append(pages, page)
	}
	return pages, nil
}

// hitTracker records, per stream, how many times the replay server actually
// served a fixture page — used by pagination_terminates to assert every
// page was consumed exactly once (no duplicate fetch, no infinite loop).
type hitTracker struct {
	mu   sync.Mutex
	hits map[string]int
}

func newHitTracker() *hitTracker {
	return &hitTracker{hits: map[string]int{}}
}

func (h *hitTracker) record(stream string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.hits[stream]++
}

func (h *hitTracker) hitsFor(stream string) int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.hits[stream]
}

// newStreamReplayServer builds an httptest.Server that replays stream's
// fixture pages in order: each incoming request is matched against the
// first not-yet-served page whose recorded method+path+query matches, that
// page's response is written, and the page is marked served (a page is
// never served twice — the design §E.2 "consumed exactly once" contract).
// An unmatched request (no remaining fixture page matches) responds 404, a
// deliberate signal rather than a silent success: a bundle whose declared
// paging/query construction doesn't match its own recorded fixtures is
// itself a conformance defect worth surfacing as a failed check.
func newStreamReplayServer(fixtures fs.FS, stream string, tracker *hitTracker) (*httptest.Server, error) {
	pages, err := loadFixturePages(fixtures, stream)
	if err != nil {
		return nil, err
	}
	served := make([]bool, len(pages))

	handler := func(w http.ResponseWriter, r *http.Request) {
		idx := matchFixturePage(pages, served, r)
		if idx < 0 {
			http.NotFound(w, r)
			return
		}
		served[idx] = true
		if tracker != nil {
			tracker.record(stream)
		}
		page := pages[idx]
		w.Header().Set("Content-Type", "application/json")
		status := page.Response.Status
		if status == 0 {
			status = http.StatusOK
		}
		w.WriteHeader(status)
		if len(page.Response.Body) > 0 {
			_, _ = w.Write(page.Response.Body)
		} else {
			_, _ = w.Write([]byte("{}"))
		}
	}

	return httptest.NewServer(http.HandlerFunc(handler)), nil
}

// reusableStreamReplayServer serves one stream's fixture pages at a time while
// keeping the same loopback listener. Large bundles can have hundreds of
// streams; using a fresh httptest.Server for every stream exhausts local TCP
// ports when the full repository test suite runs package tests concurrently.
type reusableStreamReplayServer struct {
	*httptest.Server

	mu      sync.Mutex
	stream  string
	pages   []fixturePage
	served  []bool
	tracker *hitTracker
}

func newReusableStreamReplayServer() *reusableStreamReplayServer {
	rs := &reusableStreamReplayServer{}
	rs.Server = httptest.NewServer(http.HandlerFunc(rs.serveHTTP))
	return rs
}

func (rs *reusableStreamReplayServer) reset(stream string, pages []fixturePage, tracker *hitTracker) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.stream = stream
	rs.pages = pages
	rs.served = make([]bool, len(pages))
	rs.tracker = tracker
}

func (rs *reusableStreamReplayServer) serveHTTP(w http.ResponseWriter, r *http.Request) {
	rs.mu.Lock()
	idx := matchFixturePage(rs.pages, rs.served, r)
	if idx < 0 {
		rs.mu.Unlock()
		http.NotFound(w, r)
		return
	}
	rs.served[idx] = true
	stream := rs.stream
	tracker := rs.tracker
	page := rs.pages[idx]
	rs.mu.Unlock()

	if tracker != nil {
		tracker.record(stream)
	}
	w.Header().Set("Content-Type", "application/json")
	status := page.Response.Status
	if status == 0 {
		status = http.StatusOK
	}
	w.WriteHeader(status)
	if len(page.Response.Body) > 0 {
		_, _ = w.Write(page.Response.Body)
	} else {
		_, _ = w.Write([]byte("{}"))
	}
}

// matchFixturePage returns the index of the first not-yet-served page whose
// recorded request shape matches r, or -1 if none match. Matching compares
// path exactly and every recorded query key/value (extra query params on
// the incoming request beyond what's recorded are ignored, since a spec may
// legitimately add optional params a hand-authored fixture didn't bother
// recording); method is compared case-insensitively, defaulting the
// recorded method to GET when unset.
func matchFixturePage(pages []fixturePage, served []bool, r *http.Request) int {
	for i, page := range pages {
		if served[i] {
			continue
		}
		if !requestMatchesFixture(r, page.Request) {
			continue
		}
		return i
	}
	return -1
}

func requestMatchesFixture(r *http.Request, want fixtureRequest) bool {
	wantMethod := want.Method
	if wantMethod == "" {
		wantMethod = http.MethodGet
	}
	if r.Method != wantMethod {
		return false
	}
	if want.Path != "" && r.URL.Path != want.Path {
		return false
	}
	if len(want.Query) == 0 {
		return true
	}
	got := r.URL.Query()
	for k, v := range want.Query {
		if got.Get(k) != v {
			return false
		}
	}
	return true
}

// checkFixtureFile is fixtures/check.json's shape:
// {"response":{"status","body"}}, now also carrying an OPTIONAL "request"
// field: {"request":{"method","path","query"},"response":{...}}. Check()
// always hits the bundle's single declared HTTP.Check request, so unlike
// stream pages there is exactly one file, not a numbered sequence, and
// Request.Method/Request.Path are NOT matched against the incoming request
// (checkquery-ledger.md item 5 scoped this to query only, deliberately): a
// pre-existing, repo-wide authoring convention (e.g.
// internal/connectors/defs/github/fixtures/check.json, recorded against
// "/repos/octocat/hello-world") records a realistic EXAMPLE path/method in
// "request" purely as human-readable documentation of what was captured,
// while conformance's runtimeConfigForEngine synthesizes a DIFFERENT,
// generic placeholder value ("synthetic-conformance-value") for every
// config property — so a templated check.Path (e.g.
// "/repos/{{ config.owner }}/{{ config.repo }}") never equals the fixture's
// documentary path at replay time, unlike stream page fixtures, which
// separately follow the OPPOSITE convention of recording the synthetic
// placeholder verbatim precisely so page-fixture path matching works.
// Enforcing path/method equality here would break that pre-existing,
// widespread convention for a dimension nothing asked to change; query is
// the one dimension RequestSpec.Query newly makes runtime-significant, so
// only query is compared, and (per newCheckReplayServer's doc comment) only
// for the KEYS the bundle's own base.check.query declares — never the full
// query string, which may also carry an auth-injected param
// (api_key_query's Param, e.g. nasa/openweather/aviationstack's "api_key")
// that has nothing to do with RequestSpec.Query and predates it entirely.
type checkFixtureFile struct {
	Request  fixtureRequest  `json:"request,omitempty"`
	Response fixtureResponse `json:"response"`
}

// loadCheckFixture reads fixtures/check.json. A bundle with no such file
// (ok=false) has no dedicated check fixture; callers Skip the check_fixture
// check rather than treating this as an error.
func loadCheckFixture(fixtures fs.FS) (checkFixtureFile, bool, error) {
	if fixtures == nil {
		return checkFixtureFile{}, false, nil
	}
	raw, err := fs.ReadFile(fixtures, "check.json")
	if err != nil {
		return checkFixtureFile{}, false, nil //nolint:nilerr // missing file = "no check fixture", not an error
	}
	var f checkFixtureFile
	if err := json.Unmarshal(raw, &f); err != nil {
		return checkFixtureFile{}, false, fmt.Errorf("parse fixture check.json: %w", err)
	}
	return f, true, nil
}

// newCheckReplayServer builds an httptest.Server that answers EVERY request
// whose query matches fx's recorded request for each key in checkQueryKeys
// with fx's recorded response — used for engine.Check(), which issues
// exactly one request to a single declared path (method/path are
// deliberately not matched; see checkFixtureFile's doc comment).
// checkQueryKeys is the bundle's OWN base.check.query key set (b.HTTP.Check.
// Query, not fx's): matching is scoped to exactly those keys, both because
// that is the one dimension RequestSpec.Query makes newly runtime-
// significant (checkquery-ledger.md item 5), and because the live request's
// FULL query string may carry an auth-injected param (api_key_query mode)
// that predates this dialect entirely and has nothing to do with it — a
// bundle with query-param auth but no check.query must still pass exactly
// as it always did. For each key in checkQueryKeys: if fx's recorded
// request.query has that key, its value must match the live request's
// value for that key; if fx's recorded request.query does NOT have that
// key (including when it records no query at all — every check.json that
// predates RequestSpec.Query), the match fails — this is precisely the
// ledger's named scenario: a fixture recorded before base.check.query
// existed (or never updated to match it) must fail loudly rather than
// silently pass by ignoring the query entirely. A key the bundle does NOT
// declare in check.query is never compared, regardless of what fx.Request.
// Query happens to record for it (aspirational/stale fixture data outside
// this dialect's concern). An unmatched request responds 404, which
// Requester.Do surfaces as a non-2xx error and therefore a failing
// check_fixture CheckResult, not a silent pass.
func newCheckReplayServer(fx checkFixtureFile, checkQueryKeys []string) *httptest.Server {
	status := fx.Response.Status
	if status == 0 {
		status = http.StatusOK
	}
	body := fx.Response.Body
	if len(body) == 0 {
		body = json.RawMessage("{}")
	}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !checkQueryMatchesFixture(r, fx.Request.Query, checkQueryKeys) {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write(body)
	}))
}

// checkQueryMatchesFixture reports whether r's query values, for exactly the
// keys named in checkQueryKeys, match wantQuery's recorded values for those
// same keys — see newCheckReplayServer's doc comment for the full rationale.
// A checkQueryKeys entry absent from wantQuery (nil/zero-value map included)
// always fails the match.
func checkQueryMatchesFixture(r *http.Request, wantQuery map[string]string, checkQueryKeys []string) bool {
	if len(checkQueryKeys) == 0 {
		return true
	}
	got := r.URL.Query()
	for _, k := range checkQueryKeys {
		want, ok := wantQuery[k]
		if !ok || got.Get(k) != want {
			return false
		}
	}
	return true
}
