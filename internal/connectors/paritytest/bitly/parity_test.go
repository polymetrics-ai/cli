// Package paritytest_bitly drives the legacy internal/connectors/bitly
// connector and the engine-backed connector built from
// internal/connectors/defs/bitly against ONE shared httptest.Server, per
// connector, asserting RAW reflect.DeepEqual record parity (conventions.md
// §"Parity suite minimum"). This file is the red-first test for wave1-pilot
// P-3 (bitly): it loads the bundle via engine.Load(defs.FS, "bitly") before
// the bundle exists, so the FIRST run of this file must FAIL red on a
// missing-bundle load error — captured in
// .planning/phases/wave1-pilot/traces/p3-bitly-ledger.md.
package paritytest_bitly

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/bitly"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
)

// jsonRoundTrip re-encodes v through encoding/json into a canonical
// map[string]any, so incidental Go type identity (e.g. int vs float64)
// never causes a false parity mismatch.
func jsonRoundTrip(v map[string]any) (map[string]any, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// loadBitlyBundle resolves the "bitly" bundle from defs.FS via engine.Load,
// the exact call the dispatch brief specifies (paritytest/<name> loads the
// bundle via engine.Load(defs.FS, "<name>")).
func loadBitlyBundle(t *testing.T) engine.Bundle {
	t.Helper()
	b, err := engine.Load(defs.FS, "bitly")
	if err != nil {
		t.Fatalf("engine.Load(defs.FS, %q): %v", "bitly", err)
	}
	return b
}

// withBitlyBaseURL returns a shallow copy of b with HTTP.URL pointed at
// baseURL (engine.Bundle is a value type; never mutates the loaded
// original), mirroring parity_stripe_test.go/parity_searxng_test.go's
// with<Name>BaseURL helper.
func withBitlyBaseURL(b engine.Bundle, baseURL string) engine.Bundle {
	b.HTTP.URL = baseURL
	return b
}

func bitlyRuntimeConfig(baseURL, apiKey string, extra map[string]string) connectors.RuntimeConfig {
	cfg := map[string]string{"base_url": baseURL}
	for k, v := range extra {
		cfg[k] = v
	}
	return connectors.RuntimeConfig{
		Config:  cfg,
		Secrets: map[string]string{"api_key": apiKey},
	}
}

func readAllBitlyRecords(t *testing.T, c connectors.Connector, req connectors.ReadRequest) []connectors.Record {
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

// normalizeBitlyRecord re-encodes r through encoding/json so both connectors
// compare canonical JSON shape rather than incidental Go type identity, then
// asserts RAW reflect.DeepEqual equality against legacy — no field
// stripping/normalization (conventions.md §"Red-first protocol": "never
// weaken an assertion to get green").
func normalizeBitlyRecord(t *testing.T, r connectors.Record) map[string]any {
	t.Helper()
	raw, err := jsonRoundTrip(map[string]any(r))
	if err != nil {
		t.Fatalf("json round-trip record: %v", err)
	}
	return raw
}

func normalizeBitlyRecords(t *testing.T, recs []connectors.Record) []map[string]any {
	t.Helper()
	out := make([]map[string]any, len(recs))
	for i, r := range recs {
		out[i] = normalizeBitlyRecord(t, r)
	}
	return out
}

// --- per-stream record parity: "groups" (non-paginated, no group_guid scoping) ---

func TestParityBitly_GroupsStreamRecords(t *testing.T) {
	bundle := loadBitlyBundle(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/groups" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"groups":[{"guid":"g1","name":"Acme","organization_guid":"o1","bsds":["bit.ly"],"is_active":true,"role":"admin","created":"2026-01-01T00:00:00+0000","modified":"2026-01-02T00:00:00+0000"},{"guid":"g2","name":"Beta","organization_guid":"o1","bsds":[],"is_active":false,"role":"member","created":"2026-01-03T00:00:00+0000","modified":"2026-01-04T00:00:00+0000"}]}`))
	}))
	defer srv.Close()

	legacy := bitly.New()
	legacyRecs := readAllBitlyRecords(t, legacy, connectors.ReadRequest{
		Stream: "groups",
		Config: bitlyRuntimeConfig(srv.URL, "tok_abc", nil),
	})
	if len(legacyRecs) != 2 {
		t.Fatalf("legacy records = %d, want 2 (test fixture bug)", len(legacyRecs))
	}

	eng := engine.New(withBitlyBaseURL(bundle, srv.URL), nil)
	engRecs := readAllBitlyRecords(t, eng, connectors.ReadRequest{
		Stream: "groups",
		Config: bitlyRuntimeConfig(srv.URL, "tok_abc", nil),
	})

	if len(engRecs) != len(legacyRecs) {
		t.Fatalf("record count = %d, want %d (legacy)\nengine: %+v\nlegacy: %+v", len(engRecs), len(legacyRecs), engRecs, legacyRecs)
	}
	gotNorm := normalizeBitlyRecords(t, engRecs)
	wantNorm := normalizeBitlyRecords(t, legacyRecs)
	for i := range wantNorm {
		if !reflect.DeepEqual(gotNorm[i], wantNorm[i]) {
			t.Fatalf("record %d mismatch:\nengine:  %+v\nlegacy:  %+v", i, gotNorm[i], wantNorm[i])
		}
	}
}

// --- per-stream record parity: "organizations" ---

func TestParityBitly_OrganizationsStreamRecords(t *testing.T) {
	bundle := loadBitlyBundle(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/organizations" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"organizations":[{"guid":"o1","name":"Acme Org","is_active":true,"tier":"enterprise","tier_family":"enterprise","tier_display_name":"Enterprise","role":"admin","created":"2026-01-01T00:00:00+0000","modified":"2026-01-02T00:00:00+0000"}]}`))
	}))
	defer srv.Close()

	legacy := bitly.New()
	legacyRecs := readAllBitlyRecords(t, legacy, connectors.ReadRequest{
		Stream: "organizations",
		Config: bitlyRuntimeConfig(srv.URL, "tok_abc", nil),
	})
	if len(legacyRecs) != 1 {
		t.Fatalf("legacy records = %d, want 1 (test fixture bug)", len(legacyRecs))
	}

	eng := engine.New(withBitlyBaseURL(bundle, srv.URL), nil)
	engRecs := readAllBitlyRecords(t, eng, connectors.ReadRequest{
		Stream: "organizations",
		Config: bitlyRuntimeConfig(srv.URL, "tok_abc", nil),
	})

	gotNorm := normalizeBitlyRecords(t, engRecs)
	wantNorm := normalizeBitlyRecords(t, legacyRecs)
	if !reflect.DeepEqual(gotNorm, wantNorm) {
		t.Fatalf("records mismatch:\nengine:  %+v\nlegacy:  %+v", gotNorm, wantNorm)
	}
}

// --- per-stream record parity: "campaigns" ---

func TestParityBitly_CampaignsStreamRecords(t *testing.T) {
	bundle := loadBitlyBundle(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/campaigns" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"campaigns":[{"guid":"c1","name":"Launch","description":"launch campaign","group_guid":"g1","channel_guids":["ch1","ch2"],"created":"2026-01-01T00:00:00+0000","modified":"2026-01-02T00:00:00+0000"}]}`))
	}))
	defer srv.Close()

	legacy := bitly.New()
	legacyRecs := readAllBitlyRecords(t, legacy, connectors.ReadRequest{
		Stream: "campaigns",
		Config: bitlyRuntimeConfig(srv.URL, "tok_abc", nil),
	})
	if len(legacyRecs) != 1 {
		t.Fatalf("legacy records = %d, want 1 (test fixture bug)", len(legacyRecs))
	}

	eng := engine.New(withBitlyBaseURL(bundle, srv.URL), nil)
	engRecs := readAllBitlyRecords(t, eng, connectors.ReadRequest{
		Stream: "campaigns",
		Config: bitlyRuntimeConfig(srv.URL, "tok_abc", nil),
	})

	gotNorm := normalizeBitlyRecords(t, engRecs)
	wantNorm := normalizeBitlyRecords(t, legacyRecs)
	if !reflect.DeepEqual(gotNorm, wantNorm) {
		t.Fatalf("records mismatch:\nengine:  %+v\nlegacy:  %+v", gotNorm, wantNorm)
	}
}

// --- pagination parity: "bitlinks" (group-scoped, next_url absolute) ---

// bitlyTwoPageServer serves page 1 with an ABSOLUTE pagination.next URL
// (bitly's real wire shape per legacy bitly.go:180-183's own comment, and
// bitly_test.go's TestReadBitlinksPaginates fixture) then a short/empty
// final page. Confirms the engine's next_url paginator follows the absolute
// URL and terminates (2-page fixture rule, conventions.md §4); also asserts
// the request-shape (group-scoped path on both pages). NOTE: unlike legacy
// (which resets to an empty url.Values{} once it follows an absolute
// next-page URL, bitly.go:180-183, so `size` is sent on page 1 only), the
// engine's declarative query merge re-sends `size=50` on EVERY page,
// including page 2 (engine/read.go's mergeQuery + connsdk's resolveURL
// Del+Add re-apply the stream's static query onto the absolute next URL) —
// this handler does not assert "no size param on page 2" because the engine
// legitimately sends one there (see docs.md's "Streams notes" and
// TestParityBitly_BitlinksSizeParamResentOnEveryPage below for the explicit,
// dedicated assertion of this real divergence).
func bitlyTwoPageServer(t *testing.T) (*httptest.Server, *[]string) {
	t.Helper()
	var paths []string
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path+"?"+r.URL.RawQuery)
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/groups/g1/bitlinks" && r.URL.Query().Get("search_after") == "":
			_, _ = w.Write([]byte(`{"links":[{"id":"bit.ly/a","link":"https://bit.ly/a","long_url":"https://example.com/a","title":"A","archived":false,"tags":["t1"],"deeplinks":[],"references":{"group":"g1"},"created_at":"2026-01-01T00:00:00+0000"}],"pagination":{"next":"` + srv.URL + `/groups/g1/bitlinks?search_after=tok2","search_after":"tok2","size":1}}`))
		case r.URL.Path == "/groups/g1/bitlinks" && r.URL.Query().Get("search_after") == "tok2":
			_, _ = w.Write([]byte(`{"links":[{"id":"bit.ly/b","link":"https://bit.ly/b","long_url":"https://example.com/b","title":"B","archived":false,"tags":[],"deeplinks":[],"references":{},"created_at":"2026-01-02T00:00:00+0000"}],"pagination":{"next":"","size":1}}`))
		default:
			t.Errorf("unexpected request: %s?%s", r.URL.Path, r.URL.RawQuery)
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	return srv, &paths
}

// TestParityBitly_BitlinksSizeParamResentOnEveryPage locks in the real,
// documented divergence from legacy (docs.md "Streams notes"): legacy resets
// its query to an empty url.Values{} once it follows bitlinks' absolute
// pagination.next URL, so `size=50` is sent on page 1 only (bitly.go:180-183).
// The engine's declarative next_url paginator re-merges the stream's static
// `query` onto every page request (engine/read.go's mergeQuery +
// connsdk.Requester.resolveURL's Del+Add re-apply), so `size=50` IS present
// on page 2 as well. This is verified benign in DATA terms only because
// Bitly's own next URL already carries the identical size value the engine
// re-applies (a no-op replace) — asserted explicitly here so the divergence
// itself (not just "records still match") is pinned, per
// docs/migration/conventions.md's stripping/normalization companion-assertion
// discipline.
func TestParityBitly_BitlinksSizeParamResentOnEveryPage(t *testing.T) {
	bundle := loadBitlyBundle(t)

	srv, paths := bitlyTwoPageServer(t)
	eng := engine.New(withBitlyBaseURL(bundle, srv.URL), nil)
	recs := readAllBitlyRecords(t, eng, connectors.ReadRequest{
		Stream: "bitlinks",
		Config: bitlyRuntimeConfig(srv.URL, "tok_abc", map[string]string{"group_guid": "g1"}),
	})
	if len(recs) != 2 {
		t.Fatalf("records = %d, want 2; paths=%v", len(recs), *paths)
	}
	if len(*paths) != 2 {
		t.Fatalf("requested %d pages, want 2: %v", len(*paths), *paths)
	}
	page1, err := url.ParseQuery(strings.SplitN((*paths)[0], "?", 2)[1])
	if err != nil {
		t.Fatalf("parse page 1 query: %v", err)
	}
	if got := page1.Get("size"); got != "50" {
		t.Fatalf("page 1 size = %q, want %q", got, "50")
	}
	page2, err := url.ParseQuery(strings.SplitN((*paths)[1], "?", 2)[1])
	if err != nil {
		t.Fatalf("parse page 2 query: %v", err)
	}
	if got := page2.Get("size"); got != "50" {
		t.Fatalf("page 2 size = %q, want %q (engine re-sends the static query on every page, unlike legacy — see docs.md)", got, "50")
	}
}

func TestParityBitly_BitlinksStreamPaginates(t *testing.T) {
	bundle := loadBitlyBundle(t)

	legacySrv, legacyPaths := bitlyTwoPageServer(t)
	legacy := bitly.New()
	legacyRecs := readAllBitlyRecords(t, legacy, connectors.ReadRequest{
		Stream: "bitlinks",
		Config: bitlyRuntimeConfig(legacySrv.URL, "tok_abc", map[string]string{"group_guid": "g1"}),
	})
	if len(legacyRecs) != 2 {
		t.Fatalf("legacy records = %d, want 2 (across 2 pages); paths=%v", len(legacyRecs), *legacyPaths)
	}
	if len(*legacyPaths) != 2 {
		t.Fatalf("legacy requested %d pages, want 2: %v", len(*legacyPaths), *legacyPaths)
	}

	engSrv, engPaths := bitlyTwoPageServer(t)
	eng := engine.New(withBitlyBaseURL(bundle, engSrv.URL), nil)
	engRecs := readAllBitlyRecords(t, eng, connectors.ReadRequest{
		Stream: "bitlinks",
		Config: bitlyRuntimeConfig(engSrv.URL, "tok_abc", map[string]string{"group_guid": "g1"}),
	})
	if len(engRecs) != 2 {
		t.Fatalf("engine records = %d, want 2 (across 2 pages); paths=%v", len(engRecs), *engPaths)
	}
	if len(*engPaths) != 2 {
		t.Fatalf("engine requested %d pages, want 2: %v", len(*engPaths), *engPaths)
	}

	gotNorm := normalizeBitlyRecords(t, engRecs)
	wantNorm := normalizeBitlyRecords(t, legacyRecs)
	if !reflect.DeepEqual(gotNorm, wantNorm) {
		t.Fatalf("records mismatch:\nengine:  %+v\nlegacy:  %+v", gotNorm, wantNorm)
	}
}

// TestParityBitly_BitlinksAbsoluteNextURLNotRelative locks in the SPEC.md N3
// note: bitly's pagination.next is an ABSOLUTE URL (verified against
// legacy's own bitly_test.go:64's fixture, which serves a full srv.URL-
// prefixed next link) — the wave0 N3 relative-next-url fail-closed guard
// does not bite here since the fixture never emits a relative path.
func TestParityBitly_BitlinksAbsoluteNextURLNotRelative(t *testing.T) {
	bundle := loadBitlyBundle(t)

	srv, paths := bitlyTwoPageServer(t)
	eng := engine.New(withBitlyBaseURL(bundle, srv.URL), nil)
	recs := readAllBitlyRecords(t, eng, connectors.ReadRequest{
		Stream: "bitlinks",
		Config: bitlyRuntimeConfig(srv.URL, "tok_abc", map[string]string{"group_guid": "g1"}),
	})
	if len(recs) != 2 {
		t.Fatalf("records = %d, want 2; paths=%v", len(recs), *paths)
	}
}

// --- auth header parity ---

func bitlyAuthCaptureServer(t *testing.T) (*httptest.Server, *string) {
	t.Helper()
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"groups":[]}`))
	}))
	t.Cleanup(srv.Close)
	return srv, &gotAuth
}

func TestParityBitly_BearerAuthHeaderParity(t *testing.T) {
	bundle := loadBitlyBundle(t)
	const token = "tok_secret_12345"

	legacySrv, legacyAuth := bitlyAuthCaptureServer(t)
	legacy := bitly.New()
	_ = readAllBitlyRecords(t, legacy, connectors.ReadRequest{
		Stream: "groups",
		Config: bitlyRuntimeConfig(legacySrv.URL, token, nil),
	})

	engSrv, engAuth := bitlyAuthCaptureServer(t)
	eng := engine.New(withBitlyBaseURL(bundle, engSrv.URL), nil)
	_ = readAllBitlyRecords(t, eng, connectors.ReadRequest{
		Stream: "groups",
		Config: bitlyRuntimeConfig(engSrv.URL, token, nil),
	})

	wantAuth := "Bearer " + token
	if *legacyAuth != wantAuth {
		t.Fatalf("legacy Authorization = %q, want %q (test fixture bug)", *legacyAuth, wantAuth)
	}
	if *engAuth != wantAuth {
		t.Fatalf("engine Authorization = %q, want %q (legacy)", *engAuth, wantAuth)
	}
}

// --- error-path parity (non-2xx mapping) ---

func TestParityBitly_ErrorPathParity(t *testing.T) {
	bundle := loadBitlyBundle(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message":"FORBIDDEN"}`))
	}))
	defer srv.Close()

	legacy := bitly.New()
	legacyErr := legacy.Read(context.Background(), connectors.ReadRequest{
		Stream: "groups",
		Config: bitlyRuntimeConfig(srv.URL, "tok_abc", nil),
	}, func(connectors.Record) error { return nil })
	if legacyErr == nil {
		t.Fatal("legacy Read on a 401 response = nil error, want non-nil (test fixture bug)")
	}

	eng := engine.New(withBitlyBaseURL(bundle, srv.URL), nil)
	engErr := eng.Read(context.Background(), connectors.ReadRequest{
		Stream: "groups",
		Config: bitlyRuntimeConfig(srv.URL, "tok_abc", nil),
	}, func(connectors.Record) error { return nil })
	if engErr == nil {
		t.Fatal("engine Read on a 401 response = nil error, want non-nil")
	}
}

// --- bundle load smoke guard ---

func TestParityBitly_BundleLoadsAndValidates(t *testing.T) {
	bundle := loadBitlyBundle(t)

	wantStreams := map[string]bool{"organizations": true, "groups": true, "campaigns": true, "bitlinks": true}
	if len(bundle.Streams) != len(wantStreams) {
		t.Fatalf("bundle streams = %d, want %d", len(bundle.Streams), len(wantStreams))
	}
	for _, s := range bundle.Streams {
		if !wantStreams[s.Name] {
			t.Fatalf("unexpected bundle stream %q", s.Name)
		}
	}

	if len(bundle.Writes) != 0 {
		t.Fatalf("bundle write actions = %v, want none (bitly is read-only, no writes.json)", bundle.Writes)
	}
	if bundle.Metadata.Capabilities.Write {
		t.Fatal("bundle metadata.capabilities.write = true, want false (bitly has no mutation API)")
	}
}
