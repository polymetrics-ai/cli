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
// shape {"request":{"method","path","query"},"response":{"status","body"}}.
// This shape is already load-bearing in the committed
// internal/connectors/engine/testdata/bundles/widget-demo fixture (Wave A/B
// reference), so it is reused verbatim here rather than invented fresh.
type fixturePage struct {
	file     string          // relative path, for error messages / hit tracking
	Request  fixtureRequest  `json:"request"`
	Response fixtureResponse `json:"response"`
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

// checkFixtureFile is fixtures/check.json's shape: {"response":{"status","body"}}.
// Check() always hits the bundle's single declared HTTP.Check request, so
// unlike stream pages there is exactly one file, not a numbered sequence,
// and no "request" field to match against (any request the connector's
// Check() sends is answered the same way).
type checkFixtureFile struct {
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
// with fx's recorded response — used for engine.Check(), which issues
// exactly one request to a single declared path with no pagination/query
// variation worth matching against.
func newCheckReplayServer(fx checkFixtureFile) *httptest.Server {
	status := fx.Response.Status
	if status == 0 {
		status = http.StatusOK
	}
	body := fx.Response.Body
	if len(body) == 0 {
		body = json.RawMessage("{}")
	}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write(body)
	}))
}
