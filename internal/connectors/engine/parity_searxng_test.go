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
func loadSearxngBundle(t *testing.T) engine.Bundle {
	t.Helper()
	bundles, err := engine.LoadAll(defs.FS)
	if err != nil {
		t.Fatalf("engine.LoadAll(defs.FS): %v", err)
	}
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
// canonical JSON shape rather than incidental Go type identity, with two
// documented, deliberate field-level adjustments (see docs.md "Known
// limits" and traces/waveF-b16-ledger.md):
//
//   - "stream" is dropped entirely: legacy's searxngResultRecord stamps a
//     derived "stream" key (streams.go:68) naming which stream the record
//     came from, but that value is neither present on the raw SearXNG API
//     response nor expressible via the engine's computed_fields (which
//     resolves only against record.* — there is no static-literal/
//     stream-name namespace for injecting "search"/"reddit" declaratively).
//     Dropping it does not affect dedup/incremental semantics (PK is url,
//     cursor is published_date).
//   - "engines" is normalized to a canonical sorted, comma-joined string
//     on BOTH sides: legacy's joinAny (streams.go:75-90) comma-joins the
//     raw engines[] array; the engine's declarative dialect has no
//     array-join filter (interpolate.go's applyFilter supports only
//     urlencode/unix_seconds/base64), so this bundle's schema passes the
//     raw array through unjoined. Normalizing both representations to the
//     same canonical form here compares the underlying DATA (which engines
//     contributed to a result) rather than an incidental string-formatting
//     difference the engine's dialect cannot produce.
func normalizeSearxngRecord(t *testing.T, r connectors.Record) map[string]any {
	t.Helper()
	delete(r, "stream")
	if v, ok := r["engines"]; ok {
		r["engines"] = canonicalEngines(v)
	}
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

// canonicalEngines renders v (either legacy's comma-joined string or the
// engine's raw []any/[]string) as a sorted, comma-joined string so both
// representations of "which engines contributed" compare equal.
func canonicalEngines(v any) string {
	var parts []string
	switch t := v.(type) {
	case string:
		if t == "" {
			return ""
		}
		parts = strings.Split(t, ",")
	case []any:
		for _, e := range t {
			parts = append(parts, fmt.Sprintf("%v", e))
		}
	case []string:
		parts = append(parts, t...)
	default:
		return fmt.Sprintf("%v", v)
	}
	sort.Strings(parts)
	return strings.Join(parts, ",")
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
// the short-page signal on BOTH connectors regardless of max_pages wiring —
// this isolates stream-record/templated-q parity from the max_pages engine
// gap (see TestParityStripe_MaxPagesStopEngineGap below, and
// traces/waveF-b16-ledger.md).
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
	eng := engine.New(withSearxngBaseURL(bundle, engSrv.URL), nil)
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

// --- KNOWN, DOCUMENTED ENGINE_GAP: max_pages hard-stop is not enforced by
// the declarative read path ---
//
// Legacy's default max_pages=1 is a HARD request-count cap enforced by
// connsdk.Harvest's own maxPages parameter (searxng.go:149), independent of
// page fullness: even if the source keeps returning full pages, legacy
// issues at most maxPages requests. internal/connectors/engine/read.go's
// readDeclarative loop (the ONLY production read path engine.Connector.Read
// dispatches to) never reads PaginationSpec.MaxPages at all — grep confirms
// zero references to "MaxPages" in read.go — so the engine's page_number
// paginator stops ONLY on a short/empty page (recordCount < PageSize), never
// on a request-count cap. This is confirmed both by reading read.go in full
// and by paginate_test.go's own TestNewPaginatorPageNumberMaxPagesStop: that
// test drives connsdk.Harvest directly with maxPages hard-coded to 1 as
// Harvest's explicit parameter — it does NOT exercise
// PaginationSpec.MaxPages being read out of the spec by any production
// caller, because no such caller exists.
//
// This test documents the gap directly rather than papering over it: with a
// source that ALWAYS returns a full page (so the short-page stop signal
// never fires), the engine's real Read() keeps paginating past what
// max_pages:1 would have bounded, while legacy's real Read() (fed the same
// max_pages=1 config, its default) stops at exactly one request. This is an
// ENGINE_GAP (see traces/waveF-b16-ledger.md), reported to the coordinator,
// NOT worked around in this bundle or this test file — read.go is out of
// this task's sanctioned file set.
func TestParitySearxng_MaxPagesStopEngineGap(t *testing.T) {
	bundle := loadSearxngBundle(t)

	// A page is "full" relative to EACH side's own effective page_size
	// threshold (searxngPageSize=10, matching both the bundle's declared
	// streams.json page_size and legacy's own default): serving exactly
	// searxngPageSize records on every page means the short-page stop signal
	// (which both sides correctly implement) never fires on its own, so any
	// stop that DOES happen is attributable only to a max_pages cap.
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
			// Safety valve so this documented-gap test cannot hang the suite
			// if the gap is ever silently widened: degrade to an empty page
			// so the engine's short-page stop signal eventually fires and
			// the test completes deterministically either way.
			_, _ = w.Write([]byte(`{"results":[]}`))
			return
		}
		_, _ = w.Write(fullPageBody)
	}))
	t.Cleanup(engSrv.Close)

	eng := engine.New(withSearxngBaseURL(bundle, engSrv.URL), nil)
	// The bundle's own max_pages:1 (spec.json/streams.json default) has NO
	// effect on the engine read path today (the documented gap): assert
	// current, real behavior — the engine does NOT stop at 1 request the way
	// legacy does.
	_ = readAllSearxngRecords(t, eng, connectors.ReadRequest{
		Stream: "search",
		Config: searxngRuntimeConfig(engSrv.URL, map[string]string{"query": "go etl"}),
	})

	if engHits <= 1 {
		t.Fatalf("engine issued %d requests; expected this documented gap test to demonstrate engHits > 1 (max_pages not enforced) — if this now fails, PaginationSpec.MaxPages has been wired into read.go and traces/waveF-b16-ledger.md's ENGINE_GAP entry should be closed, not this test loosened silently", engHits)
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
