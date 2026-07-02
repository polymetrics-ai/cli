// Package paritytest_xkcd is the engine-vs-legacy parity suite for the xkcd
// pilot migration (wave1-pilot P-1, PLAN.md). Both the legacy hand-written
// xkcd.Connector (internal/connectors/xkcd, read-only reference) and the
// engine-backed connector built from internal/connectors/defs/xkcd
// (engine.New(bundle, nil)) are driven against ONE shared httptest.Server per
// subtest; RAW connectors.Record reflect.DeepEqual equality is the parity
// bar, matching internal/connectors/engine/parity_searxng_test.go /
// parity_stripe_test.go. Any unavoidable deviation is documented in
// .planning/phases/wave1-pilot/traces/p1-xkcd-ledger.md's parity-deviation
// ledger, not worked around here.
//
// xkcd has no auth, no pagination, and no incremental (SPEC.md §5.1): both
// streams ("latest", "comic") return a SINGLE JSON object
// (records.single_object: true), so the parity minimum here is per-stream
// record parity, the templated comic path, hostile-path fail-closed
// behavior on both sides, and error-path (non-2xx) mapping.
package paritytest_xkcd

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
	"polymetrics.ai/internal/connectors/xkcd"
)

// loadXkcdBundle resolves the "xkcd" bundle from defs.FS via engine.Load, the
// same discovery path TestConformance and every other production caller
// uses. This is the RED-first assertion: until internal/connectors/defs/xkcd
// exists with a structurally valid bundle, this call fails every subtest in
// this file.
func loadXkcdBundle(t *testing.T) engine.Bundle {
	t.Helper()
	b, err := engine.Load(defs.FS, "xkcd")
	if err != nil {
		t.Fatalf("engine.Load(defs.FS, %q): %v", "xkcd", err)
	}
	return b
}

// withXkcdBaseURL returns a shallow copy of b with HTTP.URL pointed at
// baseURL (engine.Bundle is a value type; this never mutates the loaded
// original, mirroring parity_searxng_test.go/parity_stripe_test.go).
func withXkcdBaseURL(b engine.Bundle, baseURL string) engine.Bundle {
	b.HTTP.URL = baseURL
	return b
}

func xkcdRuntimeConfig(baseURL string, extra map[string]string) connectors.RuntimeConfig {
	cfg := map[string]string{"base_url": baseURL}
	for k, v := range extra {
		cfg[k] = v
	}
	return connectors.RuntimeConfig{Config: cfg}
}

func readAllXkcdRecords(t *testing.T, c connectors.Connector, req connectors.ReadRequest) []connectors.Record {
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

// normalizeXkcdRecord re-encodes r through encoding/json so both connectors
// compare on canonical JSON shape rather than incidental Go numeric-type
// identity (legacy decodes "num" via encoding/json into connectors.Record
// directly -> float64; the engine's read path also produces a JSON-decoded
// map, but this keeps the comparison honest regardless).
func normalizeXkcdRecord(t *testing.T, r connectors.Record) map[string]any {
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

// xkcdFixtureBody is the REAL XKCD JSON API response shape (11 fields:
// month, num, link, year, news, safe_title, transcript, alt, img, title,
// day — https://xkcd.com/json.html, confirmed by a live sample), not the
// 6-field subset the bundle's schema previously declared. Legacy's read path
// is a raw passthrough (`json.Unmarshal(resp.Body, &rec); emit(rec)`,
// xkcd.go:93-97) — every one of these 11 fields reaches a real caller on
// every real read. REVIEW-B.md finding 1 (xkcd BLOCKER): a 6-field fixture
// masks the "schema" projection mode silently dropping link/news/transcript/
// alt/img, so this fixture (and the two golden fixture files under
// defs/xkcd/fixtures/streams/**) must carry the full realistic shape for the
// parity/conformance suites to ever exercise the divergence.
const xkcdFixtureBody = `{"month":"1","num":42,"link":"","year":"2006","news":"","safe_title":"Geography","transcript":"transcript text","alt":"alt text","img":"https://imgs.xkcd.com/comics/geography.png","title":"Geography","day":"1"}`

// xkcdAllRealFieldNames is the complete real-API field set (order per
// https://xkcd.com/json.html's own documented field list); used to assert
// passthrough parity doesn't silently narrow the record.
var xkcdAllRealFieldNames = []string{"month", "num", "link", "year", "news", "safe_title", "transcript", "alt", "img", "title", "day"}

func xkcdSingleObjectServer(t *testing.T, wantPath string, body string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != wantPath {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv
}

// --- per-stream record parity: "latest" (info.0.json, single_object) ---

func TestParityXkcd_LatestStreamRecord(t *testing.T) {
	bundle := loadXkcdBundle(t)

	legacySrv := xkcdSingleObjectServer(t, "/info.0.json", xkcdFixtureBody)
	legacy := xkcd.New()
	legacyRecs := readAllXkcdRecords(t, legacy, connectors.ReadRequest{
		Stream: "latest",
		Config: xkcdRuntimeConfig(legacySrv.URL, nil),
	})
	if len(legacyRecs) != 1 {
		t.Fatalf("legacy xkcd emitted %d records for stream latest, want 1 (test fixture bug)", len(legacyRecs))
	}

	engSrv := xkcdSingleObjectServer(t, "/info.0.json", xkcdFixtureBody)
	eng := engine.New(withXkcdBaseURL(bundle, engSrv.URL), nil)
	engRecs := readAllXkcdRecords(t, eng, connectors.ReadRequest{
		Stream: "latest",
		Config: xkcdRuntimeConfig(engSrv.URL, nil),
	})
	if len(engRecs) != 1 {
		t.Fatalf("engine xkcd emitted %d records for stream latest, want 1", len(engRecs))
	}

	got := normalizeXkcdRecord(t, engRecs[0])
	want := normalizeXkcdRecord(t, legacyRecs[0])
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("latest record mismatch:\nengine:  %+v\nlegacy:  %+v", got, want)
	}
}

// --- per-stream record parity + templated path: "comic" ---

// TestParityXkcd_ComicStreamTemplatedPath drives the "comic" stream's
// templated path ({{ config.comic_number }}/info.0.json, legacy
// xkcd.go:82-88) against a server that only answers the expected
// "/614/info.0.json" path, so any request-shape divergence fails loudly on
// either side (wave0's F1 stream-path-interpolation fix's first real
// consumer per SPEC.md §5.1).
func TestParityXkcd_ComicStreamTemplatedPath(t *testing.T) {
	bundle := loadXkcdBundle(t)
	const comicNumber = "614"
	const comicBody = `{"month":"9","num":614,"link":"","year":"2009","news":"","safe_title":"Woodpecker","transcript":"woodpecker transcript","alt":"woodpecker alt text","img":"https://imgs.xkcd.com/comics/woodpecker.png","title":"Woodpecker","day":"9"}`

	legacySrv := xkcdSingleObjectServer(t, "/"+comicNumber+"/info.0.json", comicBody)
	legacy := xkcd.New()
	legacyRecs := readAllXkcdRecords(t, legacy, connectors.ReadRequest{
		Stream: "comic",
		Config: xkcdRuntimeConfig(legacySrv.URL, map[string]string{"comic_number": comicNumber}),
	})
	if len(legacyRecs) != 1 {
		t.Fatalf("legacy xkcd emitted %d records for stream comic, want 1 (test fixture bug)", len(legacyRecs))
	}

	engSrv := xkcdSingleObjectServer(t, "/"+comicNumber+"/info.0.json", comicBody)
	eng := engine.New(withXkcdBaseURL(bundle, engSrv.URL), nil)
	engRecs := readAllXkcdRecords(t, eng, connectors.ReadRequest{
		Stream: "comic",
		Config: xkcdRuntimeConfig(engSrv.URL, map[string]string{"comic_number": comicNumber}),
	})
	if len(engRecs) != 1 {
		t.Fatalf("engine xkcd emitted %d records for stream comic, want 1", len(engRecs))
	}

	got := normalizeXkcdRecord(t, engRecs[0])
	want := normalizeXkcdRecord(t, legacyRecs[0])
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("comic record mismatch:\nengine:  %+v\nlegacy:  %+v", got, want)
	}
}

// --- full 11-field record fidelity: previously-dropped fields survive ---

// TestParityXkcd_AllElevenRealFieldsSurvivePassthrough is the REVIEW-B.md
// finding-1 regression test: legacy's read path is a raw passthrough
// (xkcd.go:93-97) that emits every field of the real API response body, not
// just the 6 the bundle's schema used to declare. Before the fix this failed
// on the ENGINE side only (the "schema" projection mode silently dropped
// link/news/transcript/alt/img), proving the divergence the fixed-fixture
// alone would have masked. Both connectors must now emit the identical
// 11-field set for the SAME real-shape response.
func TestParityXkcd_AllElevenRealFieldsSurvivePassthrough(t *testing.T) {
	bundle := loadXkcdBundle(t)

	legacySrv := xkcdSingleObjectServer(t, "/info.0.json", xkcdFixtureBody)
	legacy := xkcd.New()
	legacyRecs := readAllXkcdRecords(t, legacy, connectors.ReadRequest{
		Stream: "latest",
		Config: xkcdRuntimeConfig(legacySrv.URL, nil),
	})
	if len(legacyRecs) != 1 {
		t.Fatalf("legacy xkcd emitted %d records for stream latest, want 1 (test fixture bug)", len(legacyRecs))
	}

	engSrv := xkcdSingleObjectServer(t, "/info.0.json", xkcdFixtureBody)
	eng := engine.New(withXkcdBaseURL(bundle, engSrv.URL), nil)
	engRecs := readAllXkcdRecords(t, eng, connectors.ReadRequest{
		Stream: "latest",
		Config: xkcdRuntimeConfig(engSrv.URL, nil),
	})
	if len(engRecs) != 1 {
		t.Fatalf("engine xkcd emitted %d records for stream latest, want 1", len(engRecs))
	}

	got := normalizeXkcdRecord(t, engRecs[0])
	want := normalizeXkcdRecord(t, legacyRecs[0])

	for _, field := range xkcdAllRealFieldNames {
		if _, ok := want[field]; !ok {
			t.Fatalf("test fixture bug: legacy record missing real field %q entirely: %+v", field, want)
		}
		if _, ok := got[field]; !ok {
			t.Fatalf("engine record dropped real API field %q (schema-projection silently discarding a field legacy passes through) — got %+v, want field present as in legacy %+v", field, got, want)
		}
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("11-field record mismatch:\nengine:  %+v\nlegacy:  %+v", got, want)
	}
	if len(got) != len(xkcdAllRealFieldNames) {
		t.Fatalf("engine record has %d fields, want exactly the %d real API fields %v: got %+v", len(got), len(xkcdAllRealFieldNames), xkcdAllRealFieldNames, got)
	}
}

// --- hostile comic_number fails closed on BOTH sides (SPEC.md §5.1) ---

// TestParityXkcd_HostileComicNumberFailsClosedOnBothSides asserts a
// path-traversal-shaped comic_number ("../../etc/passwd") never reaches the
// filesystem/traverses outside the intended path segment on EITHER
// connector, and that both surface an error rather than silently emitting a
// record. The two sides reach this outcome via genuinely different
// mechanisms (a documented, ACCEPTABLE parity deviation per
// conventions.md §5 — see .planning/phases/wave1-pilot/traces/
// p1-xkcd-ledger.md): legacy hand-rolls a pre-flight guard (xkcd.go:84:
// rejects any comic_number containing "/?#" outright, erroring BEFORE
// dialing out — legacyHits stays 0), whereas the engine's InterpolatePath
// urlencodes the ENTIRE resolved comic_number as one opaque path segment by
// default (engine/interpolate.go's urlencodeSegment: "../../etc/passwd"
// becomes the literal segment "..%2F..%2Fetc%2Fpasswd", with every "/"
// percent-encoded — never split into constituent ".."/"etc"/"passwd"
// segments), so the dot-dot guard (containsDotDotSegment, which only
// rejects a segment that IS, or decodes to, exactly "..") does not trip;
// the engine therefore DOES issue one request, but to a safe, non-traversing
// encoded URL (RequestURI literally ".../..%2F..%2Fetc%2Fpasswd/info.0.json"
// — the request never leaves the intended path prefix on the wire), which
// then 404s and surfaces as a read error. Neither side ever traverses
// outside the base path or leaks data for this input; the request-count
// delta (0 vs 1) is the same class of ACCEPTABLE deviation conventions.md's
// meta-rule calls "request-count delta with identical record DATA" — here
// there is no data at all on either side; see sentry's link_header analog.
//
// Note: the engine's dot-dot guard is actually STRICTER than legacy for a
// BARE ".." value (contains no "/?#", so legacy's own guard lets it through
// and legacy sends a real, unencoded "/../info.0.json" that traverses one
// directory level on the wire — a genuine pre-existing legacy gap, not
// something this migration introduces or needs to reproduce); this bundle
// does not regress that case, it closes it.
func TestParityXkcd_HostileComicNumberFailsClosedOnBothSides(t *testing.T) {
	bundle := loadXkcdBundle(t)
	const hostileComicNumber = "../../etc/passwd"

	legacyHits := 0
	legacySrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		legacyHits++
		http.NotFound(w, r)
	}))
	t.Cleanup(legacySrv.Close)

	legacy := xkcd.New()
	legacyErr := legacy.Read(context.Background(), connectors.ReadRequest{
		Stream: "comic",
		Config: xkcdRuntimeConfig(legacySrv.URL, map[string]string{"comic_number": hostileComicNumber}),
	}, func(connectors.Record) error { return nil })
	if legacyErr == nil {
		t.Fatal("legacy Read(comic, hostile comic_number) succeeded, want a fail-closed error (test fixture bug)")
	}
	if legacyHits != 0 {
		t.Fatalf("legacy issued %d requests for a hostile comic_number, want 0 (must fail closed before ever dialing out)", legacyHits)
	}

	var engRequestURI string
	engSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		engRequestURI = r.RequestURI
		http.NotFound(w, r)
	}))
	t.Cleanup(engSrv.Close)

	eng := engine.New(withXkcdBaseURL(bundle, engSrv.URL), nil)
	var engEmitted []connectors.Record
	engErr := eng.Read(context.Background(), connectors.ReadRequest{
		Stream: "comic",
		Config: xkcdRuntimeConfig(engSrv.URL, map[string]string{"comic_number": hostileComicNumber}),
	}, func(r connectors.Record) error {
		engEmitted = append(engEmitted, r)
		return nil
	})
	if engErr == nil {
		t.Fatal("engine Read(comic, hostile comic_number) succeeded, want a fail-closed error")
	}
	if len(engEmitted) != 0 {
		t.Fatalf("engine emitted %d records for a hostile comic_number, want 0", len(engEmitted))
	}
	// The engine issues one request (unlike legacy's zero), but it must never
	// traverse outside the intended path prefix: every "/" in the hostile
	// value must have been percent-encoded (as %2F), not sent as a literal
	// path separator.
	if strings.Contains(engRequestURI, "/etc/passwd") || strings.Contains(engRequestURI, "/../") {
		t.Fatalf("engine RequestURI = %q, want the hostile value's slashes percent-encoded (no literal traversal reaching the wire)", engRequestURI)
	}
	if !strings.Contains(engRequestURI, "%2F") {
		t.Fatalf("engine RequestURI = %q, want percent-encoded slashes (%%2F) proving the hostile segment was treated as one opaque token, not split", engRequestURI)
	}
}

// --- error-path parity: non-2xx mapping ---

// TestParityXkcd_NotFoundErrorPathParity asserts a 404 from the upstream API
// surfaces as an error on BOTH connectors (neither silently emits zero
// records on a hard failure).
func TestParityXkcd_NotFoundErrorPathParity(t *testing.T) {
	bundle := loadXkcdBundle(t)

	legacySrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	t.Cleanup(legacySrv.Close)

	legacy := xkcd.New()
	legacyErr := legacy.Read(context.Background(), connectors.ReadRequest{
		Stream: "latest",
		Config: xkcdRuntimeConfig(legacySrv.URL, nil),
	}, func(connectors.Record) error { return nil })
	if legacyErr == nil {
		t.Fatal("legacy Read against a 404 upstream succeeded, want an error (test fixture bug)")
	}

	engSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	t.Cleanup(engSrv.Close)

	eng := engine.New(withXkcdBaseURL(bundle, engSrv.URL), nil)
	engErr := eng.Read(context.Background(), connectors.ReadRequest{
		Stream: "latest",
		Config: xkcdRuntimeConfig(engSrv.URL, nil),
	}, func(connectors.Record) error { return nil })
	if engErr == nil {
		t.Fatal("engine Read against a 404 upstream succeeded, want an error")
	}
}

// --- write parity: both sides reject writes (read-only connector) ---

func TestParityXkcd_WriteUnsupportedOnBothSides(t *testing.T) {
	bundle := loadXkcdBundle(t)

	legacy := xkcd.New()
	if _, err := legacy.Write(context.Background(), connectors.WriteRequest{}, nil); err == nil {
		t.Fatal("legacy Write succeeded, want ErrUnsupportedOperation (test fixture bug)")
	}

	eng := engine.New(bundle, nil)
	if _, err := eng.Write(context.Background(), connectors.WriteRequest{}, nil); err == nil {
		t.Fatal("engine Write succeeded, want an error (xkcd bundle declares capabilities.write: false, no writes.json)")
	}
	if bundle.Metadata.Capabilities.Write {
		t.Fatal("bundle metadata.capabilities.write = true, want false (xkcd has no mutation API)")
	}
	if len(bundle.Writes) != 0 {
		t.Fatalf("bundle write actions = %v, want none (xkcd is read-only, no writes.json)", bundle.Writes)
	}
}

// --- manifest-surface parity ---

// TestParityXkcd_ManifestSurface compares the engine-synthesized Manifest's
// stream surface against legacy's OWN Catalog() result, mirroring
// parity_searxng_test.go's TestParitySearxng_ManifestSurface: legacy
// xkcd.Connector does not implement connectors.ManifestProvider, so
// ManifestOf falls back to the generic default path (Streams: nil) — the
// correct ground truth for "what streams does legacy xkcd actually expose"
// is c.Catalog(ctx, cfg).Streams.
func TestParityXkcd_ManifestSurface(t *testing.T) {
	bundle := loadXkcdBundle(t)

	legacy := xkcd.New()
	legacyCatalog, err := legacy.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("legacy Catalog: %v", err)
	}

	eng := engine.New(bundle, nil)
	engManifest := connectors.ManifestOf(eng)

	wantStreams := xkcdManifestStreamSurface(legacyCatalog.Streams)
	gotStreams := xkcdManifestStreamSurface(engManifest.Streams)
	if !reflect.DeepEqual(gotStreams, wantStreams) {
		t.Fatalf("stream surface = %+v, want %+v (legacy Catalog())", gotStreams, wantStreams)
	}

	if len(engManifest.WriteActions) != 0 {
		t.Fatalf("engine write actions = %v, want none (xkcd is read-only)", engManifest.WriteActions)
	}
}

type xkcdStreamSurface struct {
	Name       string
	PrimaryKey []string
}

func xkcdManifestStreamSurface(streams []connectors.Stream) []xkcdStreamSurface {
	out := make([]xkcdStreamSurface, len(streams))
	for i, s := range streams {
		out[i] = xkcdStreamSurface{Name: s.Name, PrimaryKey: append([]string{}, s.PrimaryKey...)}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// --- bundle load smoke guard ---

func TestParityXkcd_BundleLoadsAndValidates(t *testing.T) {
	bundle := loadXkcdBundle(t)

	wantStreams := []string{"comic", "latest"}
	gotStreams := make([]string, 0, len(bundle.Streams))
	for _, s := range bundle.Streams {
		gotStreams = append(gotStreams, s.Name)
	}
	sort.Strings(gotStreams)
	if !reflect.DeepEqual(gotStreams, wantStreams) {
		t.Fatalf("bundle streams = %v, want %v", gotStreams, wantStreams)
	}

	if len(bundle.Writes) != 0 {
		t.Fatalf("bundle write actions = %v, want none (xkcd is read-only, no writes.json)", bundle.Writes)
	}
	if bundle.Metadata.Capabilities.Write {
		t.Fatal("bundle metadata.capabilities.write = true, want false (xkcd has no mutation API)")
	}

	for _, s := range bundle.Streams {
		if !s.Records.SingleObject {
			t.Errorf("stream %q records.single_object = false, want true (xkcd returns a single JSON object per SPEC.md §5.1)", s.Name)
		}
	}
}
