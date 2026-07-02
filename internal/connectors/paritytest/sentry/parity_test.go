package sentryparity_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
	sentryhooks "polymetrics.ai/internal/connectors/hooks/sentry"
	"polymetrics.ai/internal/connectors/sentry"
)

// loadSentryBundle resolves the "sentry" bundle from defs.FS via
// engine.Load(defs.FS, "sentry") — the single-bundle discovery form
// SPEC.md §6 explicitly sanctions alongside engine.LoadAll (mirrors
// parity_stripe_test.go's loadStripeBundle, adapted to Load rather than
// LoadAll: during DW-1's fully-parallel dispatch, defs.FS legitimately
// contains other pilot agents' in-progress, structurally-incomplete
// directories at any given moment, and LoadAll fails hard on the FIRST
// malformed bundle anywhere in the tree — an unrelated sibling's
// in-flight write should never fail sentry's own parity suite. Load
// resolves only the sentry subtree, independent of every other
// connector's directory state.)
func loadSentryBundle(t *testing.T) engine.Bundle {
	t.Helper()
	b, err := engine.Load(defs.FS, "sentry")
	if err != nil {
		t.Fatalf("engine.Load(defs.FS, sentry): %v", err)
	}
	return b
}

func withSentryBaseURL(b engine.Bundle, baseURL string) engine.Bundle {
	b.HTTP.URL = baseURL
	return b
}

func sentryRuntimeConfig(baseURL string, extra map[string]string) connectors.RuntimeConfig {
	cfg := map[string]string{"base_url": baseURL, "organization": "acme", "project": "backend"}
	for k, v := range extra {
		cfg[k] = v
	}
	return connectors.RuntimeConfig{
		Config:  cfg,
		Secrets: map[string]string{"auth_token": "tok_abc"},
	}
}

func readAllRecords(t *testing.T, c connectors.Connector, req connectors.ReadRequest) []connectors.Record {
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

func newEngineSentry(bundle engine.Bundle, baseURL string) connectors.Connector {
	return engine.New(withSentryBaseURL(bundle, baseURL), sentryhooks.New())
}

// --- shared servers --------------------------------------------------------

// sentryProjectsServer answers the org/project-independent projects
// endpoint with a single page (no Link header at all -> both connectors
// must terminate after exactly one request).
func sentryProjectsServer(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/0/projects/" {
			http.NotFound(w, r)
			return
		}
		writeJSON(w, `[{"id":"10","slug":"backend","name":"Backend","platform":"python","status":"active","dateCreated":"2026-01-01T00:00:00Z","isPublic":false,"isBookmarked":true}]`)
	}))
	t.Cleanup(srv.Close)
	return srv
}

// sentryIssuesTwoPageServer is the dedicated Sentry Link-header twist
// fixture SPEC.md §5.3/PLAN.md P-5 calls out: page 1 sets a rel="next" Link
// with results="true" (real next page); page 2 ALSO sets rel="next" (Sentry
// ALWAYS emits it, sentry.go:7-9) but results="false" (the REAL stop
// signal) -- a page 3 request would be a defect on EITHER side, so
// receiving one at all fails the test outright via t.Errorf/http.NotFound.
func sentryIssuesTwoPageServer(t *testing.T) (*httptest.Server, *[]string) {
	t.Helper()
	var paths []string
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path+"?"+r.URL.RawQuery)
		if r.URL.Path != "/api/0/projects/acme/backend/issues/" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("cursor") {
		case "":
			w.Header().Set("Link", `<`+srv.URL+`/api/0/projects/acme/backend/issues/?cursor=0:100:0>; rel="next"; results="true"`)
			writeJSON(w, `[{"id":"1","shortId":"FIX-1","title":"NPE","culprit":"a.b in c","level":"error","status":"unresolved","type":"error","count":"3","userCount":2,"firstSeen":"2026-01-01T00:00:00Z","lastSeen":"2026-01-02T00:00:00Z"},{"id":"2","shortId":"FIX-2","title":"Timeout","culprit":"d.e in f","level":"warning","status":"resolved","type":"error","count":"1","userCount":1,"firstSeen":"2026-01-01T00:00:00Z","lastSeen":"2026-01-01T00:00:00Z"}]`)
		case "0:100:0":
			w.Header().Set("Link", `<`+srv.URL+`/api/0/projects/acme/backend/issues/?cursor=0:200:0>; rel="next"; results="false"`)
			writeJSON(w, `[{"id":"3","shortId":"FIX-3","title":"OOM","culprit":"g.h in i","level":"fatal","status":"unresolved","type":"error","count":"9","userCount":5,"firstSeen":"2026-01-01T00:00:00Z","lastSeen":"2026-01-03T00:00:00Z"}]`)
		default:
			t.Errorf("unexpected 3rd-page request (cursor=%q) -- Sentry's results=\"false\" must stop pagination without a trailing request", r.URL.Query().Get("cursor"))
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	return srv, &paths
}

func writeJSON(w http.ResponseWriter, body string) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(body))
}

// --- per-stream record parity ----------------------------------------------

func TestParitySentry_ProjectsStreamRecords(t *testing.T) {
	bundle := loadSentryBundle(t)
	srv := sentryProjectsServer(t)

	legacy := sentry.New()
	legacyRecs := readAllRecords(t, legacy, connectors.ReadRequest{Stream: "projects", Config: sentryRuntimeConfig(srv.URL, nil)})
	if len(legacyRecs) == 0 {
		t.Fatal("legacy sentry emitted zero projects records (test fixture bug)")
	}

	eng := newEngineSentry(bundle, srv.URL)
	engRecs := readAllRecords(t, eng, connectors.ReadRequest{Stream: "projects", Config: sentryRuntimeConfig(srv.URL, nil)})

	if len(engRecs) != len(legacyRecs) {
		t.Fatalf("record count = %d, want %d (legacy)\nengine: %+v\nlegacy: %+v", len(engRecs), len(legacyRecs), engRecs, legacyRecs)
	}
	for i := range legacyRecs {
		if !reflect.DeepEqual(engRecs[i], legacyRecs[i]) {
			t.Fatalf("projects record %d mismatch:\nengine: %+v\nlegacy: %+v", i, engRecs[i], legacyRecs[i])
		}
	}
}

// TestParitySentry_IssuesTwoPagePaginationAndResultsFalseStop is the
// dedicated 2-page Link-header + results= assertion PLAN.md P-5 requires:
// identical emitted sequence across 3 records/2 pages, and (via
// sentryIssuesTwoPageServer's t.Errorf on any 3rd-page request) EXACTLY 2
// requests on each side -- proving both legacy's hand-rolled harvest and the
// engine's StreamHook honor results="false" as the real stop signal despite
// rel="next" always being present.
func TestParitySentry_IssuesTwoPagePaginationAndResultsFalseStop(t *testing.T) {
	bundle := loadSentryBundle(t)

	legacySrv, legacyPaths := sentryIssuesTwoPageServer(t)
	legacy := sentry.New()
	legacyRecs := readAllRecords(t, legacy, connectors.ReadRequest{Stream: "issues", Config: sentryRuntimeConfig(legacySrv.URL, nil)})
	if len(legacyRecs) != 3 {
		t.Fatalf("legacy issues records = %d, want 3 (2 pages, test fixture bug)", len(legacyRecs))
	}
	if len(*legacyPaths) != 2 {
		t.Fatalf("legacy requests = %d, want exactly 2 (results=false must stop pagination without a trailing request); paths=%v", len(*legacyPaths), *legacyPaths)
	}

	engSrv, engPaths := sentryIssuesTwoPageServer(t)
	eng := newEngineSentry(bundle, engSrv.URL)
	engRecs := readAllRecords(t, eng, connectors.ReadRequest{Stream: "issues", Config: sentryRuntimeConfig(engSrv.URL, nil)})
	if len(engRecs) != 3 {
		t.Fatalf("engine issues records = %d, want 3 (2 pages)", len(engRecs))
	}
	if len(*engPaths) != 2 {
		t.Fatalf("engine requests = %d, want exactly 2 (results=false must stop pagination without a trailing request); paths=%v", len(*engPaths), *engPaths)
	}

	gotIDs := recordIDs(engRecs)
	wantIDs := recordIDs(legacyRecs)
	if !reflect.DeepEqual(gotIDs, wantIDs) {
		t.Fatalf("issues record id sequence = %v, want %v (legacy)", gotIDs, wantIDs)
	}
	if !reflect.DeepEqual(gotIDs, []string{"1", "2", "3"}) {
		t.Fatalf("issues record id sequence = %v, want [1 2 3]", gotIDs)
	}
	for i := range legacyRecs {
		if !reflect.DeepEqual(engRecs[i], legacyRecs[i]) {
			t.Fatalf("issues record %d mismatch:\nengine: %+v\nlegacy: %+v", i, engRecs[i], legacyRecs[i])
		}
	}
}

func recordIDs(recs []connectors.Record) []string {
	out := make([]string, len(recs))
	for i, r := range recs {
		id, _ := r["id"].(string)
		out[i] = id
	}
	return out
}

func TestParitySentry_EventsAndReleasesStreamRecords(t *testing.T) {
	bundle := loadSentryBundle(t)

	cases := []struct {
		stream  string
		path    string
		body    string
		idField string
	}{
		{
			stream:  "events",
			path:    "/api/0/projects/acme/backend/events/",
			body:    `[{"id":"e1","eventID":"evt_1","groupID":"grp_1","title":"NPE","message":"boom","platform":"python","type":"error","dateCreated":"2026-01-01T00:00:00Z"}]`,
			idField: "id",
		},
		{
			stream:  "releases",
			path:    "/api/0/organizations/acme/releases/",
			body:    `[{"version":"1.0.0","shortVersion":"1.0.0","ref":"main","url":"https://example.com","status":"open","dateCreated":"2026-01-01T00:00:00Z","dateReleased":"2026-01-02T00:00:00Z"}]`,
			idField: "version",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.stream, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != tc.path {
					http.NotFound(w, r)
					return
				}
				writeJSON(w, tc.body)
			}))
			t.Cleanup(srv.Close)

			legacy := sentry.New()
			legacyRecs := readAllRecords(t, legacy, connectors.ReadRequest{Stream: tc.stream, Config: sentryRuntimeConfig(srv.URL, nil)})
			if len(legacyRecs) == 0 || legacyRecs[0][tc.idField] == nil {
				t.Fatalf("legacy %s emitted no usable records (test fixture bug): %+v", tc.stream, legacyRecs)
			}

			eng := newEngineSentry(bundle, srv.URL)
			engRecs := readAllRecords(t, eng, connectors.ReadRequest{Stream: tc.stream, Config: sentryRuntimeConfig(srv.URL, nil)})

			if len(engRecs) != len(legacyRecs) {
				t.Fatalf("%s record count = %d, want %d (legacy)\nengine: %+v\nlegacy: %+v", tc.stream, len(engRecs), len(legacyRecs), engRecs, legacyRecs)
			}
			for i := range legacyRecs {
				if !reflect.DeepEqual(engRecs[i], legacyRecs[i]) {
					t.Fatalf("%s record %d mismatch:\nengine: %+v\nlegacy: %+v", tc.stream, i, engRecs[i], legacyRecs[i])
				}
			}
		})
	}
}

// --- auth header parity ----------------------------------------------------

func TestParitySentry_BearerAuthHeaderParity(t *testing.T) {
	bundle := loadSentryBundle(t)

	var legacyAuth, engAuth string
	legacySrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		legacyAuth = r.Header.Get("Authorization")
		writeJSON(w, `[]`)
	}))
	defer legacySrv.Close()
	engSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		engAuth = r.Header.Get("Authorization")
		writeJSON(w, `[]`)
	}))
	defer engSrv.Close()

	legacy := sentry.New()
	_ = readAllRecords(t, legacy, connectors.ReadRequest{Stream: "projects", Config: sentryRuntimeConfig(legacySrv.URL, nil)})

	eng := newEngineSentry(bundle, engSrv.URL)
	_ = readAllRecords(t, eng, connectors.ReadRequest{Stream: "projects", Config: sentryRuntimeConfig(engSrv.URL, nil)})

	if legacyAuth != "Bearer tok_abc" {
		t.Fatalf("legacy Authorization = %q, want %q (test fixture bug)", legacyAuth, "Bearer tok_abc")
	}
	if engAuth != legacyAuth {
		t.Fatalf("engine Authorization = %q, want %q (legacy, byte-exact)", engAuth, legacyAuth)
	}
}

// --- error-path parity ------------------------------------------------------

// TestParitySentry_NonSuccessStatusErrorsBothSides asserts a non-2xx
// response fails Read() on both connectors (neither emits partial records
// past a failing request).
func TestParitySentry_NonSuccessStatusErrorsBothSides(t *testing.T) {
	bundle := loadSentryBundle(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"detail":"Invalid token"}`))
	}))
	defer srv.Close()

	legacy := sentry.New()
	legacyErr := legacy.Read(context.Background(), connectors.ReadRequest{Stream: "issues", Config: sentryRuntimeConfig(srv.URL, nil)}, func(connectors.Record) error { return nil })
	if legacyErr == nil {
		t.Fatal("legacy Read did not error on a 401 response (test fixture bug)")
	}

	eng := newEngineSentry(bundle, srv.URL)
	engErr := eng.Read(context.Background(), connectors.ReadRequest{Stream: "issues", Config: sentryRuntimeConfig(srv.URL, nil)}, func(connectors.Record) error { return nil })
	if engErr == nil {
		t.Fatal("engine Read did not error on a 401 response, want an error matching legacy")
	}
}

// --- manifest-surface / bundle-shape parity ---------------------------------

func TestParitySentry_BundleLoadsAndValidates(t *testing.T) {
	bundle := loadSentryBundle(t)

	wantStreams := []string{"events", "issues", "projects", "releases"}
	gotStreams := make([]string, 0, len(bundle.Streams))
	for _, s := range bundle.Streams {
		gotStreams = append(gotStreams, s.Name)
	}
	sort.Strings(gotStreams)
	if !reflect.DeepEqual(gotStreams, wantStreams) {
		t.Fatalf("bundle streams = %v, want %v", gotStreams, wantStreams)
	}

	if bundle.Metadata.Capabilities.Write {
		t.Fatal("bundle metadata.capabilities.write = true, want false (sentry source is read-only)")
	}
	if len(bundle.Writes) != 0 {
		t.Fatalf("bundle.Writes = %+v, want none (read-only connector, no writes.json)", bundle.Writes)
	}
}

// TestParitySentry_CatalogSurface compares the engine-synthesized Manifest's
// stream surface against legacy's PUBLISHED Catalog() surface, not
// connectors.ManifestOf(legacy): legacy sentry has no hand-written
// manifest.go (unlike stripe's ManifestProvider), so ManifestOf(legacy)
// silently falls back to a zero-streams default (connectors.ManifestOf,
// manifest.go:70-82) that says nothing about sentry's real stream shape.
// Catalog() (sentry.go:101-106, sentryStreams() in streams.go) is legacy's
// actual published stream/PK/cursor-field surface and the correct
// comparison target.
func TestParitySentry_CatalogSurface(t *testing.T) {
	bundle := loadSentryBundle(t)

	legacyCat, err := sentry.New().Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("legacy Catalog: %v", err)
	}
	wantStreams := manifestStreamNames(legacyCat.Streams)

	eng := newEngineSentry(bundle, "http://example.invalid")
	engManifest := connectors.ManifestOf(eng)
	gotStreams := manifestStreamNames(engManifest.Streams)

	if !reflect.DeepEqual(gotStreams, wantStreams) {
		t.Fatalf("stream surface = %v, want %v (legacy Catalog)", gotStreams, wantStreams)
	}

	wantPK := map[string][]string{}
	for _, s := range legacyCat.Streams {
		wantPK[s.Name] = s.PrimaryKey
	}
	for _, s := range engManifest.Streams {
		want, ok := wantPK[s.Name]
		if !ok {
			continue
		}
		if !reflect.DeepEqual(s.PrimaryKey, want) {
			t.Fatalf("stream %q primary key = %v, want %v (legacy Catalog)", s.Name, s.PrimaryKey, want)
		}
	}
}

func manifestStreamNames(streams []connectors.Stream) []string {
	out := make([]string, len(streams))
	for i, s := range streams {
		out[i] = s.Name
	}
	sort.Strings(out)
	return out
}

// TestParitySentry_HostileBaseURLFailsClosedBothSides is a defense-in-depth
// SSRF sanity check (not the primary escape hatch under test, but cheap
// insurance): a non-http(s) base_url must fail Check() on both sides.
func TestParitySentry_HostileBaseURLFailsClosedBothSides(t *testing.T) {
	bundle := loadSentryBundle(t)
	cfg := sentryRuntimeConfig("file:///etc/passwd", nil)

	legacy := sentry.New()
	if err := legacy.Check(context.Background(), cfg); err == nil {
		t.Fatal("legacy Check did not reject non-http(s) base_url (test fixture bug)")
	}

	eng := newEngineSentry(bundle, "file:///etc/passwd")
	if err := eng.Check(context.Background(), cfg); err == nil {
		t.Fatal("engine Check did not reject non-http(s) base_url, want an error matching legacy")
	}
}

var _ = strings.TrimSpace // keep strings import if unused helpers are trimmed later
