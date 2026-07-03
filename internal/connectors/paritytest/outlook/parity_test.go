// Package paritytest_outlook is the engine-vs-legacy parity suite for the
// outlook AUTH_COMPLEX quarantine repair. Both the legacy hand-written
// outlook.Connector (internal/connectors/outlook, read-only reference) and
// the engine-backed connector (engine.New(bundle, engine.HooksFor("outlook")))
// are driven against the SAME httptest Graph-data server AND the SAME
// httptest TLS token-exchange server; RAW connectors.Record reflect.DeepEqual
// equality is the parity bar, matching internal/connectors/paritytest/gmail
// and paritytest/strava's precedent for a hook-backed connector
// authenticating via an OAuth2 refresh-token grant.
package paritytest_outlook

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
	outlookhook "polymetrics.ai/internal/connectors/hooks/outlook" // registers the hooks via init(); also gives direct access to Hooks.Client for TLS trust
	"polymetrics.ai/internal/connectors/outlook"
)

func loadOutlookBundle(t *testing.T) engine.Bundle {
	t.Helper()
	b, err := engine.Load(defs.FS, "outlook")
	if err != nil {
		t.Fatalf("engine.Load(defs.FS, %q): %v", "outlook", err)
	}
	return b
}

// withOutlookBaseURL returns a shallow copy of b with HTTP.URL pointed at
// baseURL (engine.Bundle is a value type; never mutates the loaded original).
func withOutlookBaseURL(b engine.Bundle, baseURL string) engine.Bundle {
	b.HTTP.URL = baseURL
	return b
}

// --- shared token-exchange server (both sides authenticate against it) ---

func tokenServer(t *testing.T, accessToken string) (*httptest.Server, *http.Client, *int32) {
	t.Helper()
	var hits int32
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
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

func outlookRuntimeConfig(baseURL, tokenURL string, extra map[string]string) connectors.RuntimeConfig {
	cfg := map[string]string{
		"base_url":  baseURL,
		"token_url": tokenURL,
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

func readAllOutlookRecords(t *testing.T, c connectors.Connector, req connectors.ReadRequest) []connectors.Record {
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

func normalizeOutlookRecord(t *testing.T, r connectors.Record) map[string]any {
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

func normalizeOutlookRecords(t *testing.T, recs []connectors.Record) []map[string]any {
	t.Helper()
	out := make([]map[string]any, len(recs))
	for i, r := range recs {
		out[i] = normalizeOutlookRecord(t, r)
	}
	return out
}

// --- Graph data server: 2-page @odata.nextLink pagination per stream -----

func outlookDataServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	var srv *httptest.Server

	mux.HandleFunc("/me/messages", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Query().Get("$skiptoken") {
		case "":
			writeJSON(w, `{"value":[{"id":"msg_1","subject":"Hello","receivedDateTime":"2026-01-01T00:00:00Z","lastModifiedDateTime":"2026-01-01T00:00:00Z","webLink":"https://example.com/1"}],"@odata.nextLink":"`+srv.URL+`/me/messages?$skiptoken=page2"}`)
		case "page2":
			writeJSON(w, `{"value":[{"id":"msg_2","subject":"World","receivedDateTime":"2026-01-02T00:00:00Z","lastModifiedDateTime":"2026-01-02T00:00:00Z","webLink":"https://example.com/2"}]}`)
		default:
			t.Fatalf("unexpected $skiptoken=%q for messages", r.URL.Query().Get("$skiptoken"))
		}
	})

	mux.HandleFunc("/me/mailFolders", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		writeJSON(w, `{"value":[{"id":"folder_1","displayName":"Inbox","totalItemCount":42,"unreadItemCount":3}]}`)
	})

	mux.HandleFunc("/me/events", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Query().Get("$skiptoken") {
		case "":
			writeJSON(w, `{"value":[{"id":"evt_1","subject":"Standup","createdDateTime":"2026-01-01T00:00:00Z","lastModifiedDateTime":"2026-01-01T00:00:00Z","webLink":"https://example.com/e1"}],"@odata.nextLink":"`+srv.URL+`/me/events?$skiptoken=page2"}`)
		case "page2":
			writeJSON(w, `{"value":[{"id":"evt_2","subject":"Review","createdDateTime":"2026-01-02T00:00:00Z","lastModifiedDateTime":"2026-01-02T00:00:00Z","webLink":"https://example.com/e2"}]}`)
		default:
			t.Fatalf("unexpected $skiptoken=%q for events", r.URL.Query().Get("$skiptoken"))
		}
	})

	srv = httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func writeJSON(w http.ResponseWriter, body string) {
	_, _ = w.Write([]byte(body))
}

func newOutlookLegacyConnector(client *http.Client) connectors.Connector {
	return outlook.Connector{Client: client}
}

func newOutlookEngineConnector(b engine.Bundle, h engine.Hooks) connectors.Connector {
	return engine.New(b, h)
}

// newHooksWithClient constructs the real outlook Hooks
// (outlookhook.New().(*outlookhook.Hooks), the exact same type
// engine.RegisterHooks("outlook", ...) constructs) but overrides its
// exported Client field to trust the shared tokenServer's self-signed TLS
// certificate.
func newHooksWithClient(client *http.Client) engine.Hooks {
	h := outlookhook.New().(*outlookhook.Hooks)
	h.Client = client
	return h
}

// --- per-stream record parity across all 3 streams -----------------------

func TestParityOutlook_StreamRecords(t *testing.T) {
	bundle := loadOutlookBundle(t)

	streams := []string{"messages", "mail_folders", "events"}
	for _, stream := range streams {
		stream := stream
		t.Run(stream, func(t *testing.T) {
			dataSrv := outlookDataServer(t)
			tokenSrv, tlsClient, _ := tokenServer(t, "tok_"+stream)

			legacy := newOutlookLegacyConnector(tlsClient)
			legacyCfg := outlookRuntimeConfig(dataSrv.URL, tokenSrv.URL, nil)
			legacyRecs := readAllOutlookRecords(t, legacy, connectors.ReadRequest{Stream: stream, Config: legacyCfg})

			eng := newOutlookEngineConnector(withOutlookBaseURL(bundle, dataSrv.URL), newHooksWithClient(tlsClient))
			engCfg := outlookRuntimeConfig(dataSrv.URL, tokenSrv.URL, nil)
			engRecs := readAllOutlookRecords(t, eng, connectors.ReadRequest{Stream: stream, Config: engCfg})

			if len(legacyRecs) == 0 {
				t.Fatalf("legacy outlook emitted zero records for stream %q (test fixture bug)", stream)
			}
			if len(engRecs) != len(legacyRecs) {
				t.Fatalf("record count = %d, want %d (legacy)\nengine:  %+v\nlegacy:  %+v", len(engRecs), len(legacyRecs), engRecs, legacyRecs)
			}

			gotNorm := normalizeOutlookRecords(t, engRecs)
			wantNorm := normalizeOutlookRecords(t, legacyRecs)
			if !reflect.DeepEqual(gotNorm, wantNorm) {
				t.Fatalf("stream %q records mismatch:\nengine:  %+v\nlegacy:  %+v", stream, gotNorm, wantNorm)
			}
		})
	}
}

// --- pagination parity: messages/events follow @odata.nextLink -----------

func TestParityOutlook_MessagesTwoPageNextLinkPagination(t *testing.T) {
	bundle := loadOutlookBundle(t)

	dataSrv := outlookDataServer(t)
	tokenSrv, tlsClient, _ := tokenServer(t, "tok_messages")

	legacy := newOutlookLegacyConnector(tlsClient)
	legacyRecs := readAllOutlookRecords(t, legacy, connectors.ReadRequest{Stream: "messages", Config: outlookRuntimeConfig(dataSrv.URL, tokenSrv.URL, nil)})
	if len(legacyRecs) != 2 {
		t.Fatalf("legacy messages records = %d, want 2 (2 pages via @odata.nextLink)", len(legacyRecs))
	}

	eng := newOutlookEngineConnector(withOutlookBaseURL(bundle, dataSrv.URL), newHooksWithClient(tlsClient))
	engRecs := readAllOutlookRecords(t, eng, connectors.ReadRequest{Stream: "messages", Config: outlookRuntimeConfig(dataSrv.URL, tokenSrv.URL, nil)})
	if len(engRecs) != 2 {
		t.Fatalf("engine messages records = %d, want 2 (2 pages via @odata.nextLink, StreamHook-driven)", len(engRecs))
	}
	if engRecs[0]["id"] != "msg_1" || engRecs[1]["id"] != "msg_2" {
		t.Fatalf("engine messages id sequence wrong: %+v", engRecs)
	}
}

func TestParityOutlook_MailFoldersSinglePageDoesNotPaginate(t *testing.T) {
	bundle := loadOutlookBundle(t)

	var legacyHits, engHits int32
	mkSrv := func(hits *int32) *httptest.Server {
		mux := http.NewServeMux()
		mux.HandleFunc("/me/mailFolders", func(w http.ResponseWriter, r *http.Request) {
			*hits++
			writeJSON(w, `{"value":[{"id":"folder_1","displayName":"Inbox","totalItemCount":1,"unreadItemCount":0}]}`)
		})
		srv := httptest.NewServer(mux)
		t.Cleanup(srv.Close)
		return srv
	}
	legacySrv := mkSrv(&legacyHits)
	engSrv := mkSrv(&engHits)

	tokenSrv, tlsClient, _ := tokenServer(t, "tok_folders")

	legacy := newOutlookLegacyConnector(tlsClient)
	_ = readAllOutlookRecords(t, legacy, connectors.ReadRequest{Stream: "mail_folders", Config: outlookRuntimeConfig(legacySrv.URL, tokenSrv.URL, nil)})
	if legacyHits != 1 {
		t.Fatalf("legacy mail_folders request count = %d, want 1 (no @odata.nextLink in fixture)", legacyHits)
	}

	eng := newOutlookEngineConnector(withOutlookBaseURL(bundle, engSrv.URL), newHooksWithClient(tlsClient))
	_ = readAllOutlookRecords(t, eng, connectors.ReadRequest{Stream: "mail_folders", Config: outlookRuntimeConfig(engSrv.URL, tokenSrv.URL, nil)})
	if engHits != 1 {
		t.Fatalf("engine mail_folders request count = %d, want 1", engHits)
	}
}

// --- auth parity: Authorization header after refresh ----------------------

func TestParityOutlook_AuthorizationHeaderAfterRefresh(t *testing.T) {
	bundle := loadOutlookBundle(t)
	const accessToken = "tok_shared_fixture_value"

	var legacyAuthHeader, engAuthHeader string

	legacySrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		legacyAuthHeader = r.Header.Get("Authorization")
		writeJSON(w, `{"value":[{"id":"folder_1"}]}`)
	}))
	t.Cleanup(legacySrv.Close)

	engSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		engAuthHeader = r.Header.Get("Authorization")
		writeJSON(w, `{"value":[{"id":"folder_1"}]}`)
	}))
	t.Cleanup(engSrv.Close)

	tokenSrv, tlsClient, hits := tokenServer(t, accessToken)

	legacy := newOutlookLegacyConnector(tlsClient)
	_ = readAllOutlookRecords(t, legacy, connectors.ReadRequest{Stream: "mail_folders", Config: outlookRuntimeConfig(legacySrv.URL, tokenSrv.URL, nil)})

	eng := newOutlookEngineConnector(withOutlookBaseURL(bundle, engSrv.URL), newHooksWithClient(tlsClient))
	_ = readAllOutlookRecords(t, eng, connectors.ReadRequest{Stream: "mail_folders", Config: outlookRuntimeConfig(engSrv.URL, tokenSrv.URL, nil)})

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

func TestParityOutlook_TokenEndpointFailureSurfacesAsAuthError(t *testing.T) {
	bundle := loadOutlookBundle(t)

	var dataHits int32
	dataSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dataHits++
		writeJSON(w, `{"value":[]}`)
	}))
	t.Cleanup(dataSrv.Close)

	failingTokenSrv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "invalid_grant"})
	}))
	t.Cleanup(failingTokenSrv.Close)
	tlsClient := failingTokenSrv.Client()

	legacy := newOutlookLegacyConnector(tlsClient)
	legacyErr := legacy.Read(context.Background(), connectors.ReadRequest{Stream: "mail_folders", Config: outlookRuntimeConfig(dataSrv.URL, failingTokenSrv.URL, nil)}, func(connectors.Record) error { return nil })
	if legacyErr == nil {
		t.Fatal("legacy Read succeeded despite a failing token endpoint, want an error (test fixture bug)")
	}

	eng := newOutlookEngineConnector(withOutlookBaseURL(bundle, dataSrv.URL), newHooksWithClient(tlsClient))
	engErr := eng.Read(context.Background(), connectors.ReadRequest{Stream: "mail_folders", Config: outlookRuntimeConfig(dataSrv.URL, failingTokenSrv.URL, nil)}, func(connectors.Record) error { return nil })
	if engErr == nil {
		t.Fatal("engine Read succeeded despite a failing token endpoint, want an error")
	}

	if dataHits != 0 {
		t.Fatalf("Graph data API received %d requests despite a failed token exchange, want 0 (no silent unauthenticated fallback)", dataHits)
	}
}

// --- write parity: both sides reject writes (read-only connector) --------

func TestParityOutlook_WriteUnsupportedOnBothSides(t *testing.T) {
	bundle := loadOutlookBundle(t)

	legacy := outlook.New()
	if _, err := legacy.Write(context.Background(), connectors.WriteRequest{}, nil); err == nil {
		t.Fatal("legacy Write succeeded, want ErrUnsupportedOperation (test fixture bug)")
	}

	eng := engine.New(bundle, engine.HooksFor("outlook"))
	if _, err := eng.Write(context.Background(), connectors.WriteRequest{}, nil); err == nil {
		t.Fatal("engine Write succeeded, want an error (outlook bundle declares capabilities.write: false, no writes.json)")
	}
	if bundle.Metadata.Capabilities.Write {
		t.Fatal("bundle metadata.capabilities.write = true, want false (outlook is read-only)")
	}
	if len(bundle.Writes) != 0 {
		t.Fatalf("bundle write actions = %v, want none (outlook is read-only, no writes.json)", bundle.Writes)
	}
}

// --- manifest-surface parity ----------------------------------------------

func TestParityOutlook_ManifestSurface(t *testing.T) {
	bundle := loadOutlookBundle(t)

	legacyCatalog, err := outlook.New().Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("legacy Catalog: %v", err)
	}

	eng := engine.New(bundle, engine.HooksFor("outlook"))
	engManifest := connectors.ManifestOf(eng)

	wantStreams := outlookManifestStreamSurface(legacyCatalog.Streams)
	gotStreams := outlookManifestStreamSurface(engManifest.Streams)
	if !reflect.DeepEqual(gotStreams, wantStreams) {
		t.Fatalf("stream surface = %+v, want %+v (legacy Catalog())", gotStreams, wantStreams)
	}

	if len(engManifest.WriteActions) != 0 {
		t.Fatalf("engine write actions = %v, want none (outlook is read-only)", engManifest.WriteActions)
	}
}

type outlookStreamSurface struct {
	Name       string
	PrimaryKey []string
}

func outlookManifestStreamSurface(streams []connectors.Stream) []outlookStreamSurface {
	out := make([]outlookStreamSurface, len(streams))
	for i, s := range streams {
		out[i] = outlookStreamSurface{Name: s.Name, PrimaryKey: append([]string{}, s.PrimaryKey...)}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// --- bundle load smoke guard ------------------------------------------------

func TestParityOutlook_BundleLoadsAndValidates(t *testing.T) {
	bundle := loadOutlookBundle(t)

	wantStreams := []string{"events", "mail_folders", "messages"}
	gotStreams := make([]string, 0, len(bundle.Streams))
	for _, s := range bundle.Streams {
		gotStreams = append(gotStreams, s.Name)
	}
	sort.Strings(gotStreams)
	if !reflect.DeepEqual(gotStreams, wantStreams) {
		t.Fatalf("bundle streams = %v, want %v", gotStreams, wantStreams)
	}

	if len(bundle.Writes) != 0 {
		t.Fatalf("bundle write actions = %v, want none (outlook is read-only, no writes.json)", bundle.Writes)
	}
	if bundle.Metadata.Capabilities.Write {
		t.Fatal("bundle metadata.capabilities.write = true, want false (outlook has no reverse-ETL write surface)")
	}
}

// --- AuthSpec shape guard ----------------------------------------------------

// TestParityOutlook_AuthSpecIsSoleCustomCandidate locks in the "no roster
// swap needed" decision: legacy has no alternate auth path, so the bundle
// declares exactly one auth candidate (mode custom, hook outlook).
func TestParityOutlook_AuthSpecIsSoleCustomCandidate(t *testing.T) {
	bundle := loadOutlookBundle(t)

	if len(bundle.HTTP.Auth) != 1 {
		t.Fatalf("len(bundle.HTTP.Auth) = %d, want 1 (no alternate auth path exists in legacy outlook)", len(bundle.HTTP.Auth))
	}
	spec := bundle.HTTP.Auth[0]
	if spec.Mode != "custom" || spec.Hook != "outlook" {
		t.Fatalf("auth spec = %+v, want mode=custom hook=outlook", spec)
	}
}
