// Package paritytest_youtubeanalytics is the engine-vs-legacy parity suite
// for the youtube-analytics Tier-2 (AuthHook) migration
// (docs/migration/quarantine.json's AUTH_COMPLEX entry). Mirrors
// paritytest/gmail's precedent byte-for-byte: both the legacy hand-written
// youtubeanalytics.Connector (internal/connectors/youtube-analytics,
// read-only reference) and the engine-backed connector
// (engine.New(bundle, engine.HooksFor("youtube-analytics"))) are driven
// against the SAME httptest Reporting-API-data server AND the SAME
// httptest TLS token-exchange server (THREAT-MODEL.md Delta 2: the hook
// requires token_url to be https); RAW connectors.Record reflect.DeepEqual
// equality is the parity bar.
package paritytest_youtubeanalytics

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
	"strings"
	"sync/atomic"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
	yahook "polymetrics.ai/internal/connectors/hooks/youtube-analytics" // registers the AuthHook via init(); also gives this test direct access to Hooks.Client for TLS trust
	youtubeanalytics "polymetrics.ai/internal/connectors/youtube-analytics"
)

const bundleName = "youtube-analytics"

// loadBundle resolves the "youtube-analytics" bundle from defs.FS via
// engine.Load, the same discovery path TestConformance and every other
// production caller uses.
func loadBundle(t *testing.T) engine.Bundle {
	t.Helper()
	b, err := engine.Load(defs.FS, bundleName)
	if err != nil {
		t.Fatalf("engine.Load(defs.FS, %q): %v", bundleName, err)
	}
	return b
}

// withBaseURL returns a shallow copy of b with HTTP.URL pointed at baseURL
// (engine.Bundle is a value type; this never mutates the loaded original).
func withBaseURL(b engine.Bundle, baseURL string) engine.Bundle {
	b.HTTP.URL = baseURL
	return b
}

// --- shared token-exchange server (both sides authenticate against it) ---

// tokenServer stands in for Google's OAuth token endpoint. It MUST be a TLS
// server: legacy's own token_url default is https, and the hook fails
// closed on a non-https token_url (THREAT-MODEL.md Delta 2) — so a plain
// httptest.Server would make the engine side fail before ever resolving an
// access token, which is not a fair parity comparison.
func tokenServer(t *testing.T, accessToken string) (*httptest.Server, *http.Client, *int32) {
	t.Helper()
	var hits int32
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		if err := r.ParseForm(); err != nil {
			t.Fatalf("token server: parse form: %v", err)
		}
		if got := r.PostForm.Get("grant_type"); got != "refresh_token" {
			t.Fatalf("token server: grant_type = %q, want refresh_token", got)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": accessToken,
			"token_type":   "Bearer",
			"expires_in":   3600,
		})
	}))
	t.Cleanup(srv.Close)
	return srv, srv.Client(), &hits
}

// runtimeConfig builds the connectors.RuntimeConfig shared by both
// connectors: base_url/token_url point at the shared httptest servers, and
// the three OAuth secrets are synthetic placeholders (never a real-looking
// credential, per THREAT-MODEL §4).
func runtimeConfig(baseURL, tokenURL string, extra map[string]string) connectors.RuntimeConfig {
	cfg := map[string]string{
		"base_url":  baseURL,
		"token_url": tokenURL,
		"scopes":    "https://www.googleapis.com/auth/yt-analytics.readonly",
		"page_size": "100",
	}
	for k, v := range extra {
		cfg[k] = v
	}
	return connectors.RuntimeConfig{
		Config: cfg,
		Secrets: map[string]string{
			"client_id":     "client-id-fixture",
			"client_secret": "client-secret-fixture",
			"refresh_token": "refresh-token-fixture",
		},
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

// normalizeRecord re-encodes r through encoding/json with UseNumber so
// legacy's native Go map-literal values and the engine's json.Number-
// preserving decode compare equal on numeric/boolean fields (mirrors
// paritytest/gmail's normalizeGmailRecord).
func normalizeRecord(t *testing.T, r connectors.Record) map[string]any {
	t.Helper()
	raw, err := json.Marshal(map[string]any(r))
	if err != nil {
		t.Fatalf("marshal record: %v", err)
	}
	var out map[string]any
	dec := json.NewDecoder(strings.NewReader(string(raw)))
	dec.UseNumber()
	if err := dec.Decode(&out); err != nil {
		t.Fatalf("decode record: %v", err)
	}
	return out
}

func normalizeRecords(t *testing.T, recs []connectors.Record) []map[string]any {
	t.Helper()
	out := make([]map[string]any, len(recs))
	for i, r := range recs {
		out[i] = normalizeRecord(t, r)
	}
	return out
}

// --- Reporting API data server fixtures (2-page pageToken/nextPageToken cursor) ---

func dataServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	mux.HandleFunc("/jobs", func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("pageToken") {
		case "":
			writeJSON(w, `{"jobs":[
				{"id":"job_1","reportTypeId":"channel_basic_a3","name":"Job One","createTime":"2026-01-01T00:00:00Z","expireTime":"2026-04-01T00:00:00Z","systemManaged":false},
				{"id":"job_2","reportTypeId":"channel_demographics_a1","name":"Job Two","createTime":"2026-01-02T00:00:00Z","expireTime":"2026-04-02T00:00:00Z","systemManaged":false}
			],"nextPageToken":"page2token"}`)
		case "page2token":
			writeJSON(w, `{"jobs":[{"id":"job_3","reportTypeId":"playlist_basic_a2","name":"Job Three","createTime":"2026-01-03T00:00:00Z","expireTime":"2026-04-03T00:00:00Z","systemManaged":true}]}`)
		default:
			t.Errorf("unexpected pageToken=%q for jobs", r.URL.Query().Get("pageToken"))
			writeJSON(w, `{"jobs":[]}`)
		}
	})

	mux.HandleFunc("/reportTypes", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, `{"reportTypes":[{"id":"channel_basic_a3","name":"User activity","deprecateTime":"","systemManaged":false}]}`)
	})

	mux.HandleFunc("/jobs/job_fixture_1/reports", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, `{"reports":[{"id":"report_1","jobId":"job_fixture_1","startTime":"2026-01-01T00:00:00Z","endTime":"2026-01-02T00:00:00Z","createTime":"2026-01-02T01:00:00Z","jobExpireTime":"2026-04-01T00:00:00Z","downloadUrl":"https://youtubereporting.googleapis.com/v1/media/report_1"}]}`)
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func writeJSON(w http.ResponseWriter, body string) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(body))
}

// newLegacyConnector builds the legacy connector wired with client (the
// TLS-trusting client for the shared token server) so the OAuth token
// exchange succeeds against tokenServer's self-signed cert.
func newLegacyConnector(client *http.Client) connectors.Connector {
	return youtubeanalytics.Connector{Client: client}
}

// newEngineConnector builds the engine-backed connector with the real
// registered AuthHook (mirrors paritytest/gmail's
// engine.New(b, engine.HooksFor("gmail")) precedent).
func newEngineConnector(b engine.Bundle, h engine.Hooks) connectors.Connector {
	return engine.New(b, h)
}

// --- per-stream record parity across all 3 streams ---

func TestParityYoutubeAnalytics_StreamRecords(t *testing.T) {
	bundle := loadBundle(t)

	streams := []string{"jobs", "report_types", "reports"}
	for _, stream := range streams {
		stream := stream
		t.Run(stream, func(t *testing.T) {
			dataSrv := dataServer(t)
			tokenSrv, tlsClient, _ := tokenServer(t, "tok_"+stream)

			extra := map[string]string{"job_id": "job_fixture_1"}

			legacy := newLegacyConnector(tlsClient)
			legacyCfg := runtimeConfig(dataSrv.URL, tokenSrv.URL, extra)
			legacyRecs := readAllRecords(t, legacy, connectors.ReadRequest{Stream: stream, Config: legacyCfg})

			hooksHook := newHooksWithClient(tlsClient)
			eng := newEngineConnector(withBaseURL(bundle, dataSrv.URL), hooksHook)
			engCfg := runtimeConfig(dataSrv.URL, tokenSrv.URL, extra)
			engRecs := readAllRecords(t, eng, connectors.ReadRequest{Stream: stream, Config: engCfg})

			if len(legacyRecs) == 0 {
				t.Fatalf("legacy youtube-analytics emitted zero records for stream %q (test fixture bug)", stream)
			}
			if len(engRecs) != len(legacyRecs) {
				t.Fatalf("record count = %d, want %d (legacy)\nengine:  %+v\nlegacy:  %+v", len(engRecs), len(legacyRecs), engRecs, legacyRecs)
			}

			gotNorm := normalizeRecords(t, engRecs)
			wantNorm := normalizeRecords(t, legacyRecs)
			for i := range wantNorm {
				if !reflect.DeepEqual(gotNorm[i], wantNorm[i]) {
					t.Fatalf("stream %q record %d mismatch:\nengine:  %+v\nlegacy:  %+v", stream, i, gotNorm[i], wantNorm[i])
				}
			}
		})
	}
}

// --- pagination parity: jobs 2-page pageToken/nextPageToken ---

func TestParityYoutubeAnalytics_JobsTwoPagePagination(t *testing.T) {
	bundle := loadBundle(t)

	dataSrv := dataServer(t)
	tokenSrv, tlsClient, _ := tokenServer(t, "tok_jobs")

	legacy := newLegacyConnector(tlsClient)
	legacyRecs := readAllRecords(t, legacy, connectors.ReadRequest{Stream: "jobs", Config: runtimeConfig(dataSrv.URL, tokenSrv.URL, nil)})
	if len(legacyRecs) != 3 {
		t.Fatalf("legacy jobs records = %d, want 3 (2 pages)", len(legacyRecs))
	}

	eng := newEngineConnector(withBaseURL(bundle, dataSrv.URL), newHooksWithClient(tlsClient))
	engRecs := readAllRecords(t, eng, connectors.ReadRequest{Stream: "jobs", Config: runtimeConfig(dataSrv.URL, tokenSrv.URL, nil)})
	if len(engRecs) != 3 {
		t.Fatalf("engine jobs records = %d, want 3 (2 pages)", len(engRecs))
	}

	gotIDs := recordIDs(t, engRecs)
	wantIDs := recordIDs(t, legacyRecs)
	if !reflect.DeepEqual(gotIDs, wantIDs) {
		t.Fatalf("jobs record id sequence = %v, want %v", gotIDs, wantIDs)
	}
	if !reflect.DeepEqual(gotIDs, []string{"job_1", "job_2", "job_3"}) {
		t.Fatalf("jobs record id sequence = %v, want [job_1 job_2 job_3]", gotIDs)
	}
}

func recordIDs(t *testing.T, recs []connectors.Record) []string {
	t.Helper()
	out := make([]string, len(recs))
	for i, r := range recs {
		id, _ := r["id"].(string)
		out[i] = id
	}
	return out
}

// --- content_owner_id optional query parameter parity ---

// TestParityYoutubeAnalytics_ContentOwnerScope asserts the optional
// content_owner_id config is sent as the onBehalfOfContentOwner query
// parameter on BOTH sides when configured, and omitted entirely on BOTH
// sides when unset (legacy's applyContentOwner, youtube_analytics.go:254-258;
// this bundle's omit_when_absent-dialect query entry, conventions.md §3).
func TestParityYoutubeAnalytics_ContentOwnerScope(t *testing.T) {
	bundle := loadBundle(t)

	t.Run("set", func(t *testing.T) {
		var legacyOwner, engOwner string
		legacySrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			legacyOwner = r.URL.Query().Get("onBehalfOfContentOwner")
			writeJSON(w, `{"reportTypes":[{"id":"channel_basic_a3","name":"User activity"}]}`)
		}))
		t.Cleanup(legacySrv.Close)
		engSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			engOwner = r.URL.Query().Get("onBehalfOfContentOwner")
			writeJSON(w, `{"reportTypes":[{"id":"channel_basic_a3","name":"User activity"}]}`)
		}))
		t.Cleanup(engSrv.Close)

		tokenSrv, tlsClient, _ := tokenServer(t, "tok_owner")

		legacy := newLegacyConnector(tlsClient)
		_ = readAllRecords(t, legacy, connectors.ReadRequest{Stream: "report_types", Config: runtimeConfig(legacySrv.URL, tokenSrv.URL, map[string]string{"content_owner_id": "owner_42"})})
		if legacyOwner != "owner_42" {
			t.Fatalf("legacy onBehalfOfContentOwner = %q, want owner_42 (test fixture bug)", legacyOwner)
		}

		eng := newEngineConnector(withBaseURL(bundle, engSrv.URL), newHooksWithClient(tlsClient))
		_ = readAllRecords(t, eng, connectors.ReadRequest{Stream: "report_types", Config: runtimeConfig(engSrv.URL, tokenSrv.URL, map[string]string{"content_owner_id": "owner_42"})})
		if engOwner != "owner_42" {
			t.Fatalf("engine onBehalfOfContentOwner = %q, want owner_42", engOwner)
		}
	})

	t.Run("unset", func(t *testing.T) {
		var legacySawParam, engSawParam bool
		legacySrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, legacySawParam = r.URL.Query()["onBehalfOfContentOwner"]
			writeJSON(w, `{"reportTypes":[{"id":"channel_basic_a3","name":"User activity"}]}`)
		}))
		t.Cleanup(legacySrv.Close)
		engSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, engSawParam = r.URL.Query()["onBehalfOfContentOwner"]
			writeJSON(w, `{"reportTypes":[{"id":"channel_basic_a3","name":"User activity"}]}`)
		}))
		t.Cleanup(engSrv.Close)

		tokenSrv, tlsClient, _ := tokenServer(t, "tok_owner_unset")

		legacy := newLegacyConnector(tlsClient)
		_ = readAllRecords(t, legacy, connectors.ReadRequest{Stream: "report_types", Config: runtimeConfig(legacySrv.URL, tokenSrv.URL, nil)})
		if legacySawParam {
			t.Fatal("legacy sent onBehalfOfContentOwner despite content_owner_id being unset (test fixture bug)")
		}

		eng := newEngineConnector(withBaseURL(bundle, engSrv.URL), newHooksWithClient(tlsClient))
		_ = readAllRecords(t, eng, connectors.ReadRequest{Stream: "report_types", Config: runtimeConfig(engSrv.URL, tokenSrv.URL, nil)})
		if engSawParam {
			t.Fatal("engine sent onBehalfOfContentOwner despite content_owner_id being unset, want omitted (omit_when_absent dialect)")
		}
	})
}

// --- auth parity: Authorization header after refresh ---

func TestParityYoutubeAnalytics_AuthorizationHeaderAfterRefresh(t *testing.T) {
	bundle := loadBundle(t)
	const accessToken = "tok_shared_fixture_value"

	var legacyAuthHeader, engAuthHeader string

	legacySrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		legacyAuthHeader = r.Header.Get("Authorization")
		writeJSON(w, `{"reportTypes":[]}`)
	}))
	t.Cleanup(legacySrv.Close)

	engSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		engAuthHeader = r.Header.Get("Authorization")
		writeJSON(w, `{"reportTypes":[]}`)
	}))
	t.Cleanup(engSrv.Close)

	tokenSrv, tlsClient, hits := tokenServer(t, accessToken)

	legacy := newLegacyConnector(tlsClient)
	_ = readAllRecords(t, legacy, connectors.ReadRequest{Stream: "report_types", Config: runtimeConfig(legacySrv.URL, tokenSrv.URL, nil)})

	eng := newEngineConnector(withBaseURL(bundle, engSrv.URL), newHooksWithClient(tlsClient))
	_ = readAllRecords(t, eng, connectors.ReadRequest{Stream: "report_types", Config: runtimeConfig(engSrv.URL, tokenSrv.URL, nil)})

	wantHeader := "Bearer " + accessToken
	if legacyAuthHeader != wantHeader {
		t.Fatalf("legacy Authorization header = %q, want %q (test fixture bug)", legacyAuthHeader, wantHeader)
	}
	if engAuthHeader != wantHeader {
		t.Fatalf("engine Authorization header = %q, want %q (legacy, same shared token exchange)", engAuthHeader, wantHeader)
	}
	if *hits != 2 {
		t.Fatalf("token endpoint hits = %d, want 2 (one refresh exchange per connector)", *hits)
	}
}

// TestParityYoutubeAnalytics_TokenEndpointFailureSurfacesAsAuthError asserts
// a token endpoint failure surfaces as an error on BOTH sides (never a
// silent unauthenticated request to the Reporting API).
func TestParityYoutubeAnalytics_TokenEndpointFailureSurfacesAsAuthError(t *testing.T) {
	bundle := loadBundle(t)

	var dataHits int32
	dataSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&dataHits, 1)
		writeJSON(w, `{"reportTypes":[]}`)
	}))
	t.Cleanup(dataSrv.Close)

	failingTokenSrv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "invalid_grant"})
	}))
	t.Cleanup(failingTokenSrv.Close)
	tlsClient := failingTokenSrv.Client()

	legacy := newLegacyConnector(tlsClient)
	legacyErr := legacy.Read(context.Background(), connectors.ReadRequest{Stream: "report_types", Config: runtimeConfig(dataSrv.URL, failingTokenSrv.URL, nil)}, func(connectors.Record) error { return nil })
	if legacyErr == nil {
		t.Fatal("legacy Read succeeded despite a failing token endpoint, want an error (test fixture bug)")
	}

	eng := newEngineConnector(withBaseURL(bundle, dataSrv.URL), newHooksWithClient(tlsClient))
	engErr := eng.Read(context.Background(), connectors.ReadRequest{Stream: "report_types", Config: runtimeConfig(dataSrv.URL, failingTokenSrv.URL, nil)}, func(connectors.Record) error { return nil })
	if engErr == nil {
		t.Fatal("engine Read succeeded despite a failing token endpoint, want an error")
	}

	if dataHits != 0 {
		t.Fatalf("Reporting API received %d requests despite a failed token exchange, want 0 (no silent unauthenticated fallback)", dataHits)
	}
}

// --- write parity: both sides reject writes (read-only connector) ---

func TestParityYoutubeAnalytics_WriteUnsupportedOnBothSides(t *testing.T) {
	bundle := loadBundle(t)

	legacy := youtubeanalytics.New()
	if _, err := legacy.Write(context.Background(), connectors.WriteRequest{}, nil); err == nil {
		t.Fatal("legacy Write succeeded, want ErrUnsupportedOperation (test fixture bug)")
	}

	eng := engine.New(bundle, engine.HooksFor(bundleName))
	if _, err := eng.Write(context.Background(), connectors.WriteRequest{}, nil); err == nil {
		t.Fatal("engine Write succeeded, want an error (youtube-analytics bundle declares capabilities.write: false, no writes.json)")
	}
	if bundle.Metadata.Capabilities.Write {
		t.Fatal("bundle metadata.capabilities.write = true, want false (youtube-analytics is read-only)")
	}
	if len(bundle.Writes) != 0 {
		t.Fatalf("bundle write actions = %v, want none (youtube-analytics is read-only, no writes.json)", bundle.Writes)
	}
}

// --- manifest-surface parity ---

func TestParityYoutubeAnalytics_ManifestSurface(t *testing.T) {
	bundle := loadBundle(t)

	legacyCatalog, err := youtubeanalytics.New().Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("legacy Catalog: %v", err)
	}

	eng := engine.New(bundle, engine.HooksFor(bundleName))
	engManifest := connectors.ManifestOf(eng)

	wantStreams := manifestStreamSurface(legacyCatalog.Streams)
	gotStreams := manifestStreamSurface(engManifest.Streams)
	if !reflect.DeepEqual(gotStreams, wantStreams) {
		t.Fatalf("stream surface = %+v, want %+v (legacy Catalog())", gotStreams, wantStreams)
	}

	if len(engManifest.WriteActions) != 0 {
		t.Fatalf("engine write actions = %v, want none (youtube-analytics is read-only)", engManifest.WriteActions)
	}
}

type streamSurface struct {
	Name       string
	PrimaryKey []string
}

func manifestStreamSurface(streams []connectors.Stream) []streamSurface {
	out := make([]streamSurface, len(streams))
	for i, s := range streams {
		out[i] = streamSurface{Name: s.Name, PrimaryKey: append([]string{}, s.PrimaryKey...)}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// --- bundle load smoke guard ---

func TestParityYoutubeAnalytics_BundleLoadsAndValidates(t *testing.T) {
	bundle := loadBundle(t)

	wantStreams := []string{"jobs", "report_types", "reports"}
	gotStreams := make([]string, 0, len(bundle.Streams))
	for _, s := range bundle.Streams {
		gotStreams = append(gotStreams, s.Name)
	}
	sort.Strings(gotStreams)
	if !reflect.DeepEqual(gotStreams, wantStreams) {
		t.Fatalf("bundle streams = %v, want %v", gotStreams, wantStreams)
	}

	if len(bundle.Writes) != 0 {
		t.Fatalf("bundle write actions = %v, want none (youtube-analytics is read-only, no writes.json)", bundle.Writes)
	}
	if bundle.Metadata.Capabilities.Write {
		t.Fatal("bundle metadata.capabilities.write = true, want false (youtube-analytics has no mutation API)")
	}
	for _, s := range bundle.Streams {
		if s.Incremental != nil {
			t.Errorf("stream %q declares an incremental block, want none (legacy never applies a state-based filter, matching sentry/gmail precedent)", s.Name)
		}
	}
}

// --- AuthSpec shape guard ---

// TestParityYoutubeAnalytics_AuthSpecIsSoleCustomCandidate locks in the
// "no roster swap needed" decision (gmail precedent): legacy has no
// alternate auth path, so the bundle declares exactly one auth candidate
// (mode custom, hook youtube-analytics), not a when-gated fallback list.
func TestParityYoutubeAnalytics_AuthSpecIsSoleCustomCandidate(t *testing.T) {
	bundle := loadBundle(t)

	if len(bundle.HTTP.Auth) != 1 {
		t.Fatalf("len(bundle.HTTP.Auth) = %d, want 1 (no alternate auth path exists in legacy youtube-analytics)", len(bundle.HTTP.Auth))
	}
	spec := bundle.HTTP.Auth[0]
	if spec.Mode != "custom" || spec.Hook != bundleName {
		t.Fatalf("auth spec = %+v, want mode=custom hook=%s", spec, bundleName)
	}
}

// --- helper: a Hooks instance carrying the shared TLS-trusting client -----

// newHooksWithClient constructs the real youtube-analytics AuthHook
// (yahook.New().(*yahook.Hooks), the exact same type
// engine.RegisterHooks("youtube-analytics", ...) constructs) but overrides
// its exported Client field to trust the shared tokenServer's self-signed
// TLS certificate — mirrors how engine.HooksFor("youtube-analytics") behaves
// in production EXCEPT for the test certificate trust, which production
// never needs (a real Google token endpoint has a publicly trusted cert).
func newHooksWithClient(client *http.Client) engine.Hooks {
	h := yahook.New().(*yahook.Hooks)
	h.Client = client
	return h
}
