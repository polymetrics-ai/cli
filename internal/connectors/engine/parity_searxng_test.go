package engine_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
	"polymetrics.ai/internal/connectors/searxng"
)

// This file is the golden migration parity suite for the searxng bundle
// (PLAN.md T-16/B-16): both streams ("search", "reddit"), templated q
// propagation, pageno pagination, and manifest-surface equality are driven
// against ONE shared httptest.Server for both the legacy hand-written
// searxng.Connector (internal/connectors/searxng, read-only reference) and
// the engine-backed connector built from internal/connectors/defs/searxng
// (engine.New(bundle, nil)). Any unavoidable deviation is documented in
// traces/waveF-b16-ledger.md's parity-deviation ledger, not worked around
// here.

// loadSearxngBundle resolves the "searxng" bundle from defs.FS via
// engine.LoadAll, the same discovery path TestConformance and every other
// production caller uses.
//
// ENGINE HARDENING (hardening-ledger.md): LoadAll(defs.FS) now returns a
// non-nil error whenever ANY bundle in the fleet fails to load (a real,
// pre-existing, out-of-scope-to-fix-here defect in ~150 OTHER bundles —
// see hardening-ledger.md), while still returning every bundle that DID
// load cleanly, searxng included. This helper only fails the test if
// searxng itself is missing from the returned set, not merely because some
// unrelated bundle elsewhere in the fleet is broken.
func loadSearxngBundle(t *testing.T) engine.Bundle {
	t.Helper()
	bundles, _ := engine.LoadAll(defs.FS)
	for _, b := range bundles {
		if b.Name == "searxng" {
			return b
		}
	}
	names := make([]string, 0, len(bundles))
	for _, b := range bundles {
		names = append(names, b.Name)
	}
	t.Fatalf("bundle %q not found in defs.FS (bundles: %v)", "searxng", names)
	return engine.Bundle{}
}

// withSearxngBaseURL returns a shallow copy of b with HTTP.URL pointed at
// baseURL (engine.Bundle is a value type; this never mutates the loaded
// original, mirroring parity_stripe_test.go's withStripeBaseURL).
func withSearxngBaseURL(b engine.Bundle, baseURL string) engine.Bundle {
	b.HTTP.URL = baseURL
	return b
}

// withSearxngUnboundedMaxPages returns a shallow copy of b whose base
// pagination spec's MaxPages is 0 (unbounded), mirroring how the legacy side
// of a short-page-stop test feeds config max_pages: "all" (searxng.go:292) to
// isolate the short-page stop signal from the (now correctly enforced)
// max_pages hard-cap — see TestParitySearxng_MaxPagesStop for the dedicated
// max_pages-cap parity assertion. Does not mutate b.HTTP.Pagination in place
// (a fresh PaginationSpec value is assigned), so the loaded original from
// loadSearxngBundle is never affected.
func withSearxngUnboundedMaxPages(b engine.Bundle) engine.Bundle {
	if b.HTTP.Pagination == nil {
		return b
	}
	pag := *b.HTTP.Pagination
	pag.MaxPages = 0
	b.HTTP.Pagination = &pag
	return b
}

func searxngRuntimeConfig(baseURL string, extra map[string]string) connectors.RuntimeConfig {
	cfg := map[string]string{"base_url": baseURL}
	for k, v := range extra {
		cfg[k] = v
	}
	return connectors.RuntimeConfig{Config: cfg}
}

func readAllSearxngRecords(t *testing.T, c connectors.Connector, req connectors.ReadRequest) []connectors.Record {
	t.Helper()
	var out []connectors.Record
	if err := c.Read(context.Background(), req, func(r connectors.Record) error {
		out = append(out, r)
		return nil
	}); err != nil {
		t.Fatalf("Read(%s): %v", req.Stream, err)
	}
	return out
}

// normalizeSearxngRecord re-encodes r through encoding/json (no UseNumber
// needed: every field on this stream is a string/float/slice, never an
// int64 vs json.Number identity mismatch) so both connectors compare on
// canonical JSON shape rather than incidental Go type identity. Both the
// "engines" array-vs-comma-joined-string deviation and the dropped "stream"
// marker field (parity-deviation ledger entries 4 and 6, docs/migration/
// conventions.md) are RESOLVED via R1's join:<sep> filter and static-literal
// computed_fields (streams.json's computed_fields now emit
// "engines": "{{ record.engines | join:, }}" and "stream": "search"/"reddit"
// literally) — this comparison is now RAW record equality, with NO
// field-level adjustments/stripping.
func normalizeSearxngRecord(t *testing.T, r connectors.Record) map[string]any {
	t.Helper()
	raw, err := json.Marshal(map[string]any(r))
	if err != nil {
		t.Fatalf("marshal record: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("unmarshal record: %v", err)
	}
	return out
}

func normalizeSearxngRecords(t *testing.T, recs []connectors.Record) []map[string]any {
	t.Helper()
	out := make([]map[string]any, len(recs))
	for i, r := range recs {
		out[i] = normalizeSearxngRecord(t, r)
	}
	return out
}

// searxngShortPageServer answers /search with a SINGLE short page (2 results,
// below any plausible page_size threshold) so pagination naturally stops via
// the short-page signal on BOTH connectors regardless of max_pages — this
// isolates stream-record/templated-q parity from max_pages cap behavior,
// which TestParitySearxng_MaxPagesStop (below) covers directly. The
// max_pages hard-cap engine gap this comment used to document was closed by
// the wave0-engine-harness repair (see traces/waveF-repair-ledger.md);
// PaginationSpec.MaxPages is now enforced by read.go's readDeclarative loop.
func searxngShortPageServer(t *testing.T) (*httptest.Server, *string) {
	t.Helper()
	var sawQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/search" {
			http.NotFound(w, r)
			return
		}
		sawQuery = r.URL.Query().Get("q")
		if pn := r.URL.Query().Get("pageno"); pn != "" && pn != "1" {
			t.Errorf("unexpected pageno=%q on short-page fixture (should stop after page 1)", pn)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"query":"go etl","number_of_results":2,"results":[
			{"url":"https://reddit.com/r/dataengineering/a","title":"First","content":"c1","engine":"reddit","engines":["reddit"],"score":1.5,"category":"general","publishedDate":"2026-01-01T00:00:00","thumbnail":null},
			{"url":"https://reddit.com/r/dataengineering/b","title":"Second","content":"c2","engine":"duckduckgo","engines":["duckduckgo","brave"],"score":1.1,"category":"general","publishedDate":"2026-01-02T00:00:00","thumbnail":null}
		]}`))
	}))
	t.Cleanup(srv.Close)
	return srv, &sawQuery
}

// --- per-stream record parity: "search" ---

func TestParitySearxng_SearchStreamRecords(t *testing.T) {
	bundle := loadSearxngBundle(t)

	srv, _ := searxngShortPageServer(t)
	legacy := searxng.New()
	legacyRecs := readAllSearxngRecords(t, legacy, connectors.ReadRequest{
		Stream: "search",
		Config: searxngRuntimeConfig(srv.URL, map[string]string{"query": "go etl"}),
	})
	if len(legacyRecs) == 0 {
		t.Fatal("legacy searxng emitted zero records for stream search (test fixture bug)")
	}

	srv2, _ := searxngShortPageServer(t)
	eng := engine.New(withSearxngBaseURL(bundle, srv2.URL), nil)
	engRecs := readAllSearxngRecords(t, eng, connectors.ReadRequest{
		Stream: "search",
		Config: searxngRuntimeConfig(srv2.URL, map[string]string{"query": "go etl"}),
	})

	if len(engRecs) != len(legacyRecs) {
		t.Fatalf("record count = %d, want %d (legacy)\nengine: %+v\nlegacy: %+v", len(engRecs), len(legacyRecs), engRecs, legacyRecs)
	}

	gotNorm := normalizeSearxngRecords(t, engRecs)
	wantNorm := normalizeSearxngRecords(t, legacyRecs)
	for i := range wantNorm {
		if !reflect.DeepEqual(gotNorm[i], wantNorm[i]) {
			t.Fatalf("record %d mismatch:\nengine:  %+v\nlegacy:  %+v", i, gotNorm[i], wantNorm[i])
		}
	}
}

// --- per-stream record parity + query scoping: "reddit" ---

// TestParitySearxng_RedditStreamScopesQuery drives the "reddit" stream's
// base-case query scoping (no subreddit configured, matching legacy's own
// searxng.go:225-242 fallback: site:reddit.com plus the user's terms). The
// subreddit-narrowing variant (site:reddit.com/r/<sub>) is a documented,
// out-of-scope simplification for this golden bundle (docs.md "Known
// limits"): the engine's declarative query templating has no
// conditional/default-value filter, so a subreddit-present-vs-absent branch
// cannot be expressed purely via stream.Query templating without risking an
// unresolved-key error when subreddit is unset — see
// traces/waveF-b16-ledger.md.
func TestParitySearxng_RedditStreamScopesQuery(t *testing.T) {
	bundle := loadSearxngBundle(t)

	legacySrv, legacyQ := searxngShortPageServer(t)
	legacy := searxng.New()
	legacyRecs := readAllSearxngRecords(t, legacy, connectors.ReadRequest{
		Stream: "reddit",
		Config: searxngRuntimeConfig(legacySrv.URL, map[string]string{"query": "best etl tool"}),
	})
	if len(legacyRecs) == 0 {
		t.Fatal("legacy searxng emitted zero records for stream reddit (test fixture bug)")
	}
	if !strings.Contains(*legacyQ, "site:reddit.com") || !strings.Contains(*legacyQ, "best etl tool") {
		t.Fatalf("legacy reddit query not scoped as expected: %q", *legacyQ)
	}

	engSrv, engQ := searxngShortPageServer(t)
	eng := engine.New(withSearxngBaseURL(bundle, engSrv.URL), nil)
	engRecs := readAllSearxngRecords(t, eng, connectors.ReadRequest{
		Stream: "reddit",
		Config: searxngRuntimeConfig(engSrv.URL, map[string]string{"query": "best etl tool"}),
	})

	if *engQ != *legacyQ {
		t.Fatalf("engine reddit q = %q, want %q (legacy templated q propagation)", *engQ, *legacyQ)
	}
	if len(engRecs) != len(legacyRecs) {
		t.Fatalf("record count = %d, want %d (legacy)", len(engRecs), len(legacyRecs))
	}

	gotNorm := normalizeSearxngRecords(t, engRecs)
	wantNorm := normalizeSearxngRecords(t, legacyRecs)
	for i := range wantNorm {
		if !reflect.DeepEqual(gotNorm[i], wantNorm[i]) {
			t.Fatalf("reddit record %d mismatch:\nengine:  %+v\nlegacy:  %+v", i, gotNorm[i], wantNorm[i])
		}
	}
}

// --- optional bearer-proxy auth (api_key) parity (F6, REVIEW.md) ---

// searxngAuthCaptureServer answers /search with a single short page and
// records the Authorization header it observed, for both connectors to be
// driven against identically.
func searxngAuthCaptureServer(t *testing.T) (*httptest.Server, *string) {
	t.Helper()
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"results":[{"url":"https://x/1","title":"t","content":"c","engine":"e","engines":["e"],"score":1,"category":"general","publishedDate":"2026-01-01T00:00:00","thumbnail":null}]}`))
	}))
	t.Cleanup(srv.Close)
	return srv, &gotAuth
}

// TestParitySearxng_ApiKeySecretSendsBearerAuth locks in F6's resolution: an
// optional api_key secret, when set, is applied as a Bearer Authorization
// header on BOTH connectors (legacy's searxng.go:184-189 requester ->
// connsdk.Bearer(token); the engine bundle's streams.json base.auth now
// declares a `when`-gated bearer spec: `{"mode":"bearer","token":"{{
// secrets.api_key }}","when":"{{ secrets.api_key }}"}` falling back to
// `{"mode":"none"}, using R1's absent-key-falsy `when` tolerance so an unset
// secret safely selects "none" instead of erroring).
func TestParitySearxng_ApiKeySecretSendsBearerAuth(t *testing.T) {
	bundle := loadSearxngBundle(t)
	const token = "proxy-token-12345"

	legacySrv, legacyAuth := searxngAuthCaptureServer(t)
	legacy := searxng.New()
	legacyCfg := searxngRuntimeConfig(legacySrv.URL, map[string]string{"query": "go etl"})
	legacyCfg.Secrets = map[string]string{"api_key": token}
	_ = readAllSearxngRecords(t, legacy, connectors.ReadRequest{Stream: "search", Config: legacyCfg})

	engSrv, engAuth := searxngAuthCaptureServer(t)
	eng := engine.New(withSearxngBaseURL(bundle, engSrv.URL), nil)
	engCfg := searxngRuntimeConfig(engSrv.URL, map[string]string{"query": "go etl"})
	engCfg.Secrets = map[string]string{"api_key": token}
	_ = readAllSearxngRecords(t, eng, connectors.ReadRequest{Stream: "search", Config: engCfg})

	wantAuth := "Bearer " + token
	if *legacyAuth != wantAuth {
		t.Fatalf("legacy Authorization = %q, want %q (test fixture bug)", *legacyAuth, wantAuth)
	}
	if *engAuth != wantAuth {
		t.Fatalf("engine Authorization = %q, want %q (legacy, api_key secret configured)", *engAuth, wantAuth)
	}
}

// TestParitySearxng_ApiKeyAbsentSendsNoAuth locks in the OTHER half of F6's
// resolution: when api_key is unset (the common public-instance case), both
// connectors send NO Authorization header at all (not an empty one, not an
// error).
func TestParitySearxng_ApiKeyAbsentSendsNoAuth(t *testing.T) {
	bundle := loadSearxngBundle(t)

	legacySrv, legacyAuth := searxngAuthCaptureServer(t)
	legacy := searxng.New()
	_ = readAllSearxngRecords(t, legacy, connectors.ReadRequest{
		Stream: "search",
		Config: searxngRuntimeConfig(legacySrv.URL, map[string]string{"query": "go etl"}),
	})

	engSrv, engAuth := searxngAuthCaptureServer(t)
	eng := engine.New(withSearxngBaseURL(bundle, engSrv.URL), nil)
	_ = readAllSearxngRecords(t, eng, connectors.ReadRequest{
		Stream: "search",
		Config: searxngRuntimeConfig(engSrv.URL, map[string]string{"query": "go etl"}),
	})

	if *legacyAuth != "" {
		t.Fatalf("legacy Authorization = %q, want empty (test fixture bug: api_key unset)", *legacyAuth)
	}
	if *engAuth != "" {
		t.Fatalf("engine Authorization = %q, want empty (legacy, api_key secret unset)", *engAuth)
	}
}

// --- pageno pagination sequence + short-page stop ---

// searxngPageSize is the bundle's declared streams.json base pagination
// page_size (10, matching legacy's own searxngDefaultPageSize). Unlike
// legacy, the engine has no runtime (config-driven) override for a bundle's
// declared PaginationSpec.PageSize — it is a fixed value baked into
// streams.json, read straight from the loaded spec with no interpolation or
// config lookup (engine/read.go's readDeclarative: `if pag.PageSize > 0 {
// pageSize = pag.PageSize }`, never consulting req.Config). This test
// therefore feeds LEGACY the equivalent page_size config value (its own
// native override mechanism) rather than relying on unsupported per-request
// engine configurability, so both sides use the identical effective
// short-page threshold — the correct parity bar, matching how
// parity_stripe_test.go's IncrementalCreatedGTEFromState feeds each side its
// own native cursor representation for the same logical instant.
const searxngPageSize = 10

// searxngTwoPageServer serves a full first page (searxngPageSize results)
// then a short second page (1 result), so BOTH connectors must issue exactly
// two requests (pageno=1 then pageno=2) and stop — no size param is ever
// sent (legacy never sends one; PageNumberPaginator.SizeParam is left empty
// in the bundle, matching searxng.go:141-145's "no per-page size param is
// sent" comment).
func searxngTwoPageServer(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/search" {
			http.NotFound(w, r)
			return
		}
		for k := range r.URL.Query() {
			if k != "q" && k != "format" && k != "pageno" {
				t.Errorf("unexpected extra query param %q (searxng sends no size param)", k)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Query().Get("pageno") {
		case "", "1":
			_, _ = w.Write([]byte(searxngFullPageBody(searxngPageSize, "page1")))
		case "2":
			_, _ = w.Write([]byte(`{"results":[
				{"url":"https://x/page2-0","title":"last","content":"c","engine":"e","engines":["e"],"score":1,"category":"general","publishedDate":"2026-02-01T00:00:00","thumbnail":null}
			]}`))
		default:
			t.Errorf("unexpected pageno=%q", r.URL.Query().Get("pageno"))
			_, _ = w.Write([]byte(`{"results":[]}`))
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

// searxngFullPageBody builds a {"results":[...]} JSON body with exactly n
// result objects (every field a schema-parity-safe non-null value), each
// with a distinct URL derived from label+index.
func searxngFullPageBody(n int, label string) string {
	var b strings.Builder
	b.WriteString(`{"results":[`)
	for i := 0; i < n; i++ {
		if i > 0 {
			b.WriteString(",")
		}
		fmt.Fprintf(&b, `{"url":"https://x/%s-%d","title":"t%d","content":"c%d","engine":"e","engines":["e"],"score":1,"category":"general","publishedDate":"2026-01-01T00:00:0%dZ","thumbnail":null}`, label, i, i, i, i%10)
	}
	b.WriteString(`]}`)
	return b.String()
}

func TestParitySearxng_PagenoSequenceAndShortPageStop(t *testing.T) {
	bundle := loadSearxngBundle(t)

	legacySrv := searxngTwoPageServer(t)
	legacy := searxng.New()
	legacyRecs := readAllSearxngRecords(t, legacy, connectors.ReadRequest{
		Stream: "search",
		Config: searxngRuntimeConfig(legacySrv.URL, map[string]string{"query": "go etl", "page_size": strconv.Itoa(searxngPageSize), "max_pages": "all"}),
	})
	if len(legacyRecs) != searxngPageSize+1 {
		t.Fatalf("legacy records = %d, want %d (2 pages: full then short)", len(legacyRecs), searxngPageSize+1)
	}

	engSrv := searxngTwoPageServer(t)
	// Unbounded max_pages on this side, matching legacy's own max_pages:"all"
	// above: this test's concern is the pageno sequence + short-page stop,
	// not the max_pages cap (which TestParitySearxng_MaxPagesStop covers
	// directly) — without this override the bundle's declared max_pages:1
	// would stop the engine after page 1, before the short-page signal on
	// page 2 is ever reached.
	eng := engine.New(withSearxngUnboundedMaxPages(withSearxngBaseURL(bundle, engSrv.URL)), nil)
	engRecs := readAllSearxngRecords(t, eng, connectors.ReadRequest{
		Stream: "search",
		Config: searxngRuntimeConfig(engSrv.URL, map[string]string{"query": "go etl"}),
	})
	if len(engRecs) != searxngPageSize+1 {
		t.Fatalf("engine records = %d, want %d (2 pages: full then short, same page_size=%d as the bundle declares)", len(engRecs), searxngPageSize+1, searxngPageSize)
	}

	gotURLs := recordURLs(t, engRecs)
	wantURLs := recordURLs(t, legacyRecs)
	if !reflect.DeepEqual(gotURLs, wantURLs) {
		t.Fatalf("record url sequence = %v, want %v (legacy)", gotURLs, wantURLs)
	}
}

func recordURLs(t *testing.T, recs []connectors.Record) []string {
	t.Helper()
	out := make([]string, len(recs))
	for i, r := range recs {
		u, _ := r["url"].(string)
		out[i] = u
	}
	return out
}

// --- max_pages hard-stop parity (formerly a documented ENGINE_GAP, closed by
// the wave0-engine-harness repair — see traces/waveF-repair-ledger.md) ---
//
// Legacy's default max_pages=1 is a HARD request-count cap enforced by
// connsdk.Harvest's own maxPages parameter (searxng.go:149), independent of
// page fullness: even if the source keeps returning full pages, legacy
// issues at most maxPages requests. internal/connectors/engine/read.go's
// readDeclarative loop now enforces the SAME cap by consulting
// PaginationSpec.MaxPages directly (the effective spec after stream-overrides
// -base resolution), so a page_number stream whose source always returns a
// full page stops at exactly max_pages requests on BOTH connectors.
//
// Only the DEFAULT case (max_pages=1, legacy's default and this bundle's
// declared streams.json base value) is asserted for parity here. Legacy also
// supports a CONFIG-driven max_pages override (max_pages: "all"/"unlimited"/N
// read from cfg.Config at request time, searxng.go:287-303) — the engine's
// PaginationSpec.MaxPages is a static int with no template support, so a
// config-driven override is NOT modeled by this bundle. This remains a
// documented, deliberate deviation (docs.md "Known limits"; not re-litigated
// here since it does not affect the DEFAULT-case parity this test asserts).
func TestParitySearxng_MaxPagesStop(t *testing.T) {
	bundle := loadSearxngBundle(t)

	// A page is "full" relative to EACH side's own effective page_size
	// threshold (searxngPageSize=10, matching both the bundle's declared
	// streams.json page_size and legacy's own default): serving exactly
	// searxngPageSize records on every page means the short-page stop signal
	// (which both sides correctly implement) never fires on its own, so any
	// stop that DOES happen is attributable only to the max_pages cap.
	const capServerMaxHits = 5 // safety valve: never serve unboundedly many full pages
	fullPageBody := []byte(searxngFullPageBody(searxngPageSize, "full"))

	legacyHits := 0
	legacySrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		legacyHits++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(fullPageBody)
	}))
	t.Cleanup(legacySrv.Close)

	legacy := searxng.New()
	legacyRecs := readAllSearxngRecords(t, legacy, connectors.ReadRequest{
		Stream: "search",
		// max_pages omitted: legacy's default is 1 (searxngDefaultMaxPages).
		Config: searxngRuntimeConfig(legacySrv.URL, map[string]string{"query": "go etl", "page_size": strconv.Itoa(searxngPageSize)}),
	})
	if legacyHits != 1 {
		t.Fatalf("legacy issued %d requests against an always-full page source, want exactly 1 (max_pages default = 1 hard stop)", legacyHits)
	}
	if len(legacyRecs) != searxngPageSize {
		t.Fatalf("legacy records = %d, want %d (one page only)", len(legacyRecs), searxngPageSize)
	}

	engHits := 0
	engSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		engHits++
		w.Header().Set("Content-Type", "application/json")
		if engHits > capServerMaxHits {
			// Safety valve so a regression cannot hang the suite: degrade to
			// an empty page so the engine's short-page stop signal eventually
			// fires and the test completes deterministically either way.
			_, _ = w.Write([]byte(`{"results":[]}`))
			return
		}
		_, _ = w.Write(fullPageBody)
	}))
	t.Cleanup(engSrv.Close)

	eng := engine.New(withSearxngBaseURL(bundle, engSrv.URL), nil)
	// The bundle's own max_pages:1 (streams.json base pagination block) must
	// now stop the engine's read at exactly 1 request, matching legacy.
	engRecs := readAllSearxngRecords(t, eng, connectors.ReadRequest{
		Stream: "search",
		Config: searxngRuntimeConfig(engSrv.URL, map[string]string{"query": "go etl"}),
	})

	if engHits != 1 {
		t.Fatalf("engine issued %d requests against an always-full page source, want exactly 1 (max_pages:1 hard stop, matching legacy's default) — if this fails, PaginationSpec.MaxPages wiring in read.go has regressed", engHits)
	}
	if len(engRecs) != searxngPageSize {
		t.Fatalf("engine records = %d, want %d (one page only, matching legacy)", len(engRecs), searxngPageSize)
	}

	gotURLs := recordURLs(t, engRecs)
	wantURLs := recordURLs(t, legacyRecs)
	if !reflect.DeepEqual(gotURLs, wantURLs) {
		t.Fatalf("record url sequence = %v, want %v (legacy) — max_pages stop must yield byte-identical record sets on both sides", gotURLs, wantURLs)
	}
}

// --- manifest-surface parity ---

// TestParitySearxng_ManifestSurface compares the engine-synthesized
// Manifest's stream surface against legacy's OWN Catalog() result (not
// connectors.ManifestOf(searxng.New()).Streams): unlike stripe, legacy
// searxng.Connector does not implement connectors.ManifestProvider (no
// hand-written Manifest() method — grep confirms
// internal/connectors/searxng has no Manifest() func), so ManifestOf falls
// back to connectors.ManifestOf's generic default path, which never calls
// Catalog() and always returns Streams: nil. That is legacy's real,
// legitimate (if minimal) behavior — not a bug to work around — so the
// correct ground truth for "what streams does legacy searxng actually
// expose" is c.Catalog(ctx, cfg).Streams, exactly what the CLI's own
// catalog command would call.
func TestParitySearxng_ManifestSurface(t *testing.T) {
	bundle := loadSearxngBundle(t)

	legacy := searxng.New()
	legacyCatalog, err := legacy.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("legacy Catalog: %v", err)
	}

	eng := engine.New(bundle, nil)
	engManifest := connectors.ManifestOf(eng)

	wantStreams := manifestSearxngStreamSurface(legacyCatalog.Streams)
	gotStreams := manifestSearxngStreamSurface(engManifest.Streams)
	if !reflect.DeepEqual(gotStreams, wantStreams) {
		t.Fatalf("stream surface = %+v, want %+v (legacy Catalog())", gotStreams, wantStreams)
	}

	if len(engManifest.WriteActions) != 0 {
		t.Fatalf("engine write actions = %v, want none (searxng is read-only)", engManifest.WriteActions)
	}
	legacyManifest := connectors.ManifestOf(legacy)
	if len(legacyManifest.WriteActions) != 0 {
		t.Fatalf("legacy write actions = %v, want none (test fixture bug: searxng.Metadata().Capabilities.Write should be false)", legacyManifest.WriteActions)
	}
}

type searxngStreamSurface struct {
	Name         string
	PrimaryKey   []string
	CursorFields []string
}

func manifestSearxngStreamSurface(streams []connectors.Stream) []searxngStreamSurface {
	out := make([]searxngStreamSurface, len(streams))
	for i, s := range streams {
		out[i] = searxngStreamSurface{Name: s.Name, PrimaryKey: append([]string{}, s.PrimaryKey...), CursorFields: append([]string{}, s.CursorFields...)}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// --- bundle load smoke guard ---

func TestParitySearxng_BundleLoadsAndValidates(t *testing.T) {
	bundle := loadSearxngBundle(t)

	wantStreams := []string{"reddit", "search"}
	gotStreams := make([]string, 0, len(bundle.Streams))
	for _, s := range bundle.Streams {
		gotStreams = append(gotStreams, s.Name)
	}
	sort.Strings(gotStreams)
	if !reflect.DeepEqual(gotStreams, wantStreams) {
		t.Fatalf("bundle streams = %v, want %v", gotStreams, wantStreams)
	}

	if len(bundle.Writes) != 0 {
		t.Fatalf("bundle write actions = %v, want none (searxng is read-only, no writes.json)", bundle.Writes)
	}
	if bundle.Metadata.Capabilities.Write {
		t.Fatal("bundle metadata.capabilities.write = true, want false (searxng has no mutation API)")
	}
}

// --- registrygen regression: legacy searxng stays live-registered ---

// TestSearxngRegistrygenSkipMapRegression is the T-16 regression test PLAN.md
// calls out explicitly: even after cmd/registrygen's skip map grows to
// include defs/engine/hooks/native/conformance/certify (none of which are
// connector packages), legacy internal/connectors/searxng must still be
// discovered by registrygen (it is a real connector package directory, not a
// skip-listed one) and self-register via RegisterNativeLive
// (searxng.go:43-47) into connectors.NewLiveRegistry(), the registry the CLI
// uses in production.
func TestSearxngRegistrygenSkipMapRegression(t *testing.T) {
	r := connectors.NewLiveRegistry()
	if _, ok := r.Get("searxng"); !ok {
		t.Fatal("live registry did not resolve searxng after registrygen skip-map edit (RegisterNativeLive regression)")
	}
}
