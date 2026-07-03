// Package paritytest_akeneo is the engine-vs-legacy parity suite for the
// akeneo migration. Both the legacy hand-written akeneo.Connector
// (internal/connectors/akeneo, read-only reference) and the engine-backed
// connector (engine.New(bundle, engine.HooksFor("akeneo"))) are driven
// against the SAME httptest data server AND the SAME httptest token-exchange
// server; RAW connectors.Record reflect.DeepEqual equality is the parity
// bar, matching internal/connectors/paritytest/gmail's precedent for a
// hook-backed, skip_dynamic'd pilot.
package paritytest_akeneo

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
	"strings"
	"sync/atomic"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/akeneo"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
	akeneohook "polymetrics.ai/internal/connectors/hooks/akeneo" // registers the AuthHook via init()
)

// loadAkeneoBundle resolves the "akeneo" bundle from defs.FS via
// engine.Load, the same discovery path TestConformance and every other
// production caller uses.
func loadAkeneoBundle(t *testing.T) engine.Bundle {
	t.Helper()
	b, err := engine.Load(defs.FS, "akeneo")
	if err != nil {
		t.Fatalf("engine.Load(defs.FS, %q): %v", "akeneo", err)
	}
	return b
}

// withAkeneoBaseURL returns a shallow copy of b with HTTP.URL pointed at
// baseURL (engine.Bundle is a value type; never mutates the loaded
// original).
func withAkeneoBaseURL(b engine.Bundle, baseURL string) engine.Bundle {
	b.HTTP.URL = baseURL
	return b
}

// --- shared token-exchange + resource server (both sides authenticate and read against it) ---

// tokenAndChannelsServer stands in for a full Akeneo deployment: it serves
// BOTH /api/oauth/v1/token (capturing the Basic client_id:secret header and
// the request count) and /api/rest/v1/channels (an empty-items page,
// sufficient for auth-header-only assertions). Legacy's own akeneoBaseURL
// tolerates plain http (its own test suite drives a plain httptest.Server,
// not a TLS one), so this is a plain httptest.Server, unlike gmail's TLS
// requirement for the unrelated Google OAuth endpoint.
func tokenAndChannelsServer(t *testing.T, accessToken string) (*httptest.Server, *int32, *string) {
	t.Helper()
	var hits int32
	var gotBasic string
	mux := http.NewServeMux()
	mux.HandleFunc("/api/oauth/v1/token", func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		gotBasic = r.Header.Get("Authorization")
		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("token server: decode body: %v", err)
		}
		if body["grant_type"] != "password" {
			t.Errorf("token server: grant_type = %q, want password", body["grant_type"])
		}
		writeJSON(w, `{"access_token":"`+accessToken+`","token_type":"bearer","expires_in":3600}`)
	})
	mux.HandleFunc("/api/rest/v1/channels", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, `{"_links":{"self":{"href":"x"}},"_embedded":{"items":[]}}`)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv, &hits, &gotBasic
}

// akeneoRuntimeConfig builds the connectors.RuntimeConfig shared by both
// connectors. base_url doubles as both the token-exchange host AND the
// resource-data host for the legacy side (legacy derives token_url from
// base_url/host, akeneo.go:239); the engine side's AuthSpec.TokenURL
// template is likewise "{{ config.base_url }}/api/oauth/v1/token", so both
// sides hit the SAME token server whenever dataSrvURL == tokenSrvURL. Tests
// that need the data and token servers to be different hosts build the
// RuntimeConfig by hand instead.
func akeneoRuntimeConfig(baseURL string, extra map[string]string) connectors.RuntimeConfig {
	cfg := map[string]string{
		"base_url":     baseURL,
		"client_id":    "client-id-fixture",
		"api_username": "api-user-fixture",
		"page_size":    "100",
	}
	for k, v := range extra {
		cfg[k] = v
	}
	return connectors.RuntimeConfig{
		Config: cfg,
		Secrets: map[string]string{
			"secret":   "client-secret-fixture",
			"password": "user-password-fixture",
		},
	}
}

func readAllAkeneoRecords(t *testing.T, c connectors.Connector, req connectors.ReadRequest) []connectors.Record {
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

// normalizeAkeneoRecord re-encodes r through encoding/json with UseNumber so
// incidental Go type identity never causes a false parity mismatch (mirrors
// paritytest/gmail's normalizeGmailRecord).
func normalizeAkeneoRecord(t *testing.T, r connectors.Record) map[string]any {
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

func normalizeAkeneoRecords(t *testing.T, recs []connectors.Record) []map[string]any {
	t.Helper()
	out := make([]map[string]any, len(recs))
	for i, r := range recs {
		out[i] = normalizeAkeneoRecord(t, r)
	}
	return out
}

func writeJSON(w http.ResponseWriter, body string) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(body))
}

// newAkeneoLegacyConnector builds the legacy connector wired with client so
// the OAuth token exchange succeeds against the shared token server.
func newAkeneoLegacyConnector(client *http.Client) connectors.Connector {
	return akeneo.Connector{Client: client}
}

// newAkeneoEngineConnector builds the engine-backed connector with the real
// registered AuthHook, matching paritytest/gmail's precedent
// (engine.New(bundle, engine.HooksFor("gmail"))).
func newAkeneoEngineConnector(b engine.Bundle, h engine.Hooks) connectors.Connector {
	return engine.New(b, h)
}

func newHooksWithClient(client *http.Client) engine.Hooks {
	h := akeneohook.New().(*akeneohook.Hooks)
	h.Client = client
	return h
}

// --- single-host data+token server (products/categories/families/attributes/channels) ---

// akeneoDataServer serves both the token endpoint (delegated to
// tokenHandler) and the 5 REST resource endpoints under ONE httptest.Server,
// mirroring how a real Akeneo deployment serves both the OAuth token
// endpoint and the REST API from the same host.
func akeneoDataServer(t *testing.T, tokenHandler http.HandlerFunc) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/api/oauth/v1/token", tokenHandler)

	mux.HandleFunc("/api/rest/v1/products", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, `{"_links":{"self":{"href":"x"}},"_embedded":{"items":[
			{"identifier":"prod_1","uuid":"00000000-0000-0000-0000-000000000001","enabled":true,"family":"shoes","parent":null,"categories":["cat_1"],"groups":[],"values":{},"created":"2026-01-01T00:00:00+00:00","updated":"2026-01-01T00:00:00+00:00"},
			{"identifier":"prod_2","uuid":"00000000-0000-0000-0000-000000000002","enabled":false,"family":"hats","parent":"prod_1","categories":["cat_1"],"groups":[],"values":{},"created":"2026-01-02T00:00:00+00:00","updated":"2026-01-02T00:00:00+00:00"}
		]}}`)
	})
	mux.HandleFunc("/api/rest/v1/categories", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, `{"_links":{"self":{"href":"x"}},"_embedded":{"items":[
			{"code":"master","parent":null,"labels":{"en_US":"Master"},"updated":"2026-01-01T00:00:00+00:00"}
		]}}`)
	})
	mux.HandleFunc("/api/rest/v1/families", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, `{"_links":{"self":{"href":"x"}},"_embedded":{"items":[
			{"code":"shoes","attribute_as_label":"name","attribute_as_image":null,"attributes":["name"],"labels":{"en_US":"Shoes"}}
		]}}`)
	})
	mux.HandleFunc("/api/rest/v1/attributes", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, `{"_links":{"self":{"href":"x"}},"_embedded":{"items":[
			{"code":"name","type":"pim_catalog_text","group":"marketing","localizable":false,"scopable":false,"labels":{"en_US":"Name"}}
		]}}`)
	})
	mux.HandleFunc("/api/rest/v1/channels", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, `{"_links":{"self":{"href":"x"}},"_embedded":{"items":[
			{"code":"ecommerce","currencies":["USD"],"locales":["en_US"],"category_tree":"master","labels":{"en_US":"E-commerce"}}
		]}}`)
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func tokenHandlerReturning(accessToken string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body map[string]string
		_ = json.NewDecoder(r.Body).Decode(&body)
		writeJSON(w, `{"access_token":"`+accessToken+`","token_type":"bearer","expires_in":3600}`)
	}
}

// --- per-stream record parity across all 5 streams ---

func TestParityAkeneo_StreamRecords(t *testing.T) {
	bundle := loadAkeneoBundle(t)

	streams := []string{"products", "categories", "families", "attributes", "channels"}
	for _, stream := range streams {
		stream := stream
		t.Run(stream, func(t *testing.T) {
			srv := akeneoDataServer(t, tokenHandlerReturning("tok_"+stream))

			legacy := newAkeneoLegacyConnector(nil)
			legacyCfg := akeneoRuntimeConfig(srv.URL, nil)
			legacyRecs := readAllAkeneoRecords(t, legacy, connectors.ReadRequest{Stream: stream, Config: legacyCfg})

			eng := newAkeneoEngineConnector(withAkeneoBaseURL(bundle, srv.URL), newHooksWithClient(nil))
			engCfg := akeneoRuntimeConfig(srv.URL, nil)
			engRecs := readAllAkeneoRecords(t, eng, connectors.ReadRequest{Stream: stream, Config: engCfg})

			if len(legacyRecs) == 0 {
				t.Fatalf("legacy akeneo emitted zero records for stream %q (test fixture bug)", stream)
			}
			if len(engRecs) != len(legacyRecs) {
				t.Fatalf("record count = %d, want %d (legacy)\nengine:  %+v\nlegacy:  %+v", len(engRecs), len(legacyRecs), engRecs, legacyRecs)
			}

			gotNorm := normalizeAkeneoRecords(t, engRecs)
			wantNorm := normalizeAkeneoRecords(t, legacyRecs)
			for i := range wantNorm {
				if !reflect.DeepEqual(gotNorm[i], wantNorm[i]) {
					t.Fatalf("stream %q record %d mismatch:\nengine:  %+v\nlegacy:  %+v", stream, i, gotNorm[i], wantNorm[i])
				}
			}
		})
	}
}

// --- pagination parity: products across 2 HAL pages (_links.next.href) ---

// akeneoTwoPageServer serves page 1 of /api/rest/v1/products with an
// ABSOLUTE _links.next.href pointing back at the SAME server (mirrors
// legacy's harvest loop, akeneo.go:141-181, and akeneo_test.go's own
// TestReadPaginatesAndAuthenticates fixture shape), then a final page with
// no next link.
func akeneoTwoPageServer(t *testing.T, tokenHandler http.HandlerFunc) (*httptest.Server, *[]string) {
	t.Helper()
	var paths []string
	var srv *httptest.Server
	mux := http.NewServeMux()
	mux.HandleFunc("/api/oauth/v1/token", tokenHandler)
	mux.HandleFunc("/api/rest/v1/products", func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path+"?"+r.URL.RawQuery)
		if r.URL.Query().Get("page") == "2" {
			writeJSON(w, `{"_links":{"self":{"href":"x"}},"_embedded":{"items":[
				{"identifier":"prod_3","uuid":"00000000-0000-0000-0000-000000000003","enabled":true,"family":"shoes","parent":null,"categories":["cat_1"],"groups":[],"values":{},"created":"2026-01-03T00:00:00+00:00","updated":"2026-01-03T00:00:00+00:00"}
			]}}`)
			return
		}
		writeJSON(w, `{"_links":{"self":{"href":"x"},"next":{"href":"`+srv.URL+`/api/rest/v1/products?page=2&limit=100"}},"_embedded":{"items":[
			{"identifier":"prod_1","uuid":"00000000-0000-0000-0000-000000000001","enabled":true,"family":"shoes","parent":null,"categories":["cat_1"],"groups":[],"values":{},"created":"2026-01-01T00:00:00+00:00","updated":"2026-01-01T00:00:00+00:00"},
			{"identifier":"prod_2","uuid":"00000000-0000-0000-0000-000000000002","enabled":false,"family":"hats","parent":"prod_1","categories":["cat_1"],"groups":[],"values":{},"created":"2026-01-02T00:00:00+00:00","updated":"2026-01-02T00:00:00+00:00"}
		]}}`)
	})
	srv = httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv, &paths
}

func TestParityAkeneo_ProductsStreamPaginatesAcrossHALPages(t *testing.T) {
	bundle := loadAkeneoBundle(t)

	legacySrv, legacyPaths := akeneoTwoPageServer(t, tokenHandlerReturning("tok_legacy"))
	legacy := newAkeneoLegacyConnector(nil)
	legacyRecs := readAllAkeneoRecords(t, legacy, connectors.ReadRequest{Stream: "products", Config: akeneoRuntimeConfig(legacySrv.URL, nil)})
	if len(legacyRecs) != 3 {
		t.Fatalf("legacy records = %d, want 3 (across 2 pages); paths=%v", len(legacyRecs), *legacyPaths)
	}
	if len(*legacyPaths) != 2 {
		t.Fatalf("legacy requested %d pages, want 2: %v", len(*legacyPaths), *legacyPaths)
	}

	engSrv, engPaths := akeneoTwoPageServer(t, tokenHandlerReturning("tok_engine"))
	eng := newAkeneoEngineConnector(withAkeneoBaseURL(bundle, engSrv.URL), newHooksWithClient(nil))
	engRecs := readAllAkeneoRecords(t, eng, connectors.ReadRequest{Stream: "products", Config: akeneoRuntimeConfig(engSrv.URL, nil)})
	if len(engRecs) != 3 {
		t.Fatalf("engine records = %d, want 3 (across 2 pages); paths=%v", len(engRecs), *engPaths)
	}
	if len(*engPaths) != 2 {
		t.Fatalf("engine requested %d pages, want 2: %v", len(*engPaths), *engPaths)
	}

	gotNorm := normalizeAkeneoRecords(t, engRecs)
	wantNorm := normalizeAkeneoRecords(t, legacyRecs)
	if !reflect.DeepEqual(gotNorm, wantNorm) {
		t.Fatalf("records mismatch:\nengine:  %+v\nlegacy:  %+v", gotNorm, wantNorm)
	}
}

// --- auth parity: Basic token-exchange header ---

// TestParityAkeneo_AuthHeadersParity asserts BOTH connectors send the
// identical HTTP Basic client_id:secret header on the token-exchange
// request. Legacy always derives both the token-exchange and resource
// requests from the same base_url/host (akeneo.go:230-252), so both
// requests land on the SAME httptest.Server here; the resource-request
// Bearer header is separately, explicitly asserted by
// TestParityAkeneo_BearerHeaderAfterExchange.
func TestParityAkeneo_AuthHeadersParity(t *testing.T) {
	bundle := loadAkeneoBundle(t)
	const accessToken = "tok_shared_fixture_value"

	legacyTokenSrv, legacyHits, legacyBasic := tokenAndChannelsServer(t, accessToken)
	legacy := newAkeneoLegacyConnector(nil)
	_ = readAllAkeneoRecords(t, legacy, connectors.ReadRequest{Stream: "channels", Config: akeneoRuntimeConfig(legacyTokenSrv.URL, nil)})

	engTokenSrv, engHits, engBasic := tokenAndChannelsServer(t, accessToken)
	eng := newAkeneoEngineConnector(withAkeneoBaseURL(bundle, engTokenSrv.URL), newHooksWithClient(nil))
	_ = readAllAkeneoRecords(t, eng, connectors.ReadRequest{Stream: "channels", Config: akeneoRuntimeConfig(engTokenSrv.URL, nil)})

	wantBasic := "Basic " + base64.StdEncoding.EncodeToString([]byte("client-id-fixture:client-secret-fixture"))
	if *legacyBasic != wantBasic {
		t.Fatalf("legacy token Authorization = %q, want %q (test fixture bug)", *legacyBasic, wantBasic)
	}
	if *engBasic != wantBasic {
		t.Fatalf("engine token Authorization = %q, want %q (legacy)", *engBasic, wantBasic)
	}
	if *legacyHits != 1 {
		t.Fatalf("legacy token endpoint hits = %d, want 1", *legacyHits)
	}
	if *engHits != 1 {
		t.Fatalf("engine token endpoint hits = %d, want 1", *engHits)
	}
}

// TestParityAkeneo_BearerHeaderAfterExchange asserts BOTH connectors send
// the identical "Authorization: Bearer <access_token>" header (derived from
// the SAME token-exchange response) on the resource request.
func TestParityAkeneo_BearerHeaderAfterExchange(t *testing.T) {
	bundle := loadAkeneoBundle(t)
	const accessToken = "tok_bearer_fixture"

	var legacyAuth, engAuth string
	legacySrv := akeneoDataServerCapturingAuth(t, accessToken, &legacyAuth)
	legacy := newAkeneoLegacyConnector(nil)
	_ = readAllAkeneoRecords(t, legacy, connectors.ReadRequest{Stream: "channels", Config: akeneoRuntimeConfig(legacySrv.URL, nil)})

	engSrv := akeneoDataServerCapturingAuth(t, accessToken, &engAuth)
	eng := newAkeneoEngineConnector(withAkeneoBaseURL(bundle, engSrv.URL), newHooksWithClient(nil))
	_ = readAllAkeneoRecords(t, eng, connectors.ReadRequest{Stream: "channels", Config: akeneoRuntimeConfig(engSrv.URL, nil)})

	wantAuth := "Bearer " + accessToken
	if legacyAuth != wantAuth {
		t.Fatalf("legacy resource Authorization = %q, want %q (test fixture bug)", legacyAuth, wantAuth)
	}
	if engAuth != wantAuth {
		t.Fatalf("engine resource Authorization = %q, want %q (legacy, same shared token exchange)", engAuth, wantAuth)
	}
}

// akeneoDataServerCapturingAuth serves BOTH the token endpoint (always
// returning accessToken) and a single /api/rest/v1/channels resource
// endpoint that records its Authorization header into gotAuth.
func akeneoDataServerCapturingAuth(t *testing.T, accessToken string, gotAuth *string) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/api/oauth/v1/token", tokenHandlerReturning(accessToken))
	mux.HandleFunc("/api/rest/v1/channels", func(w http.ResponseWriter, r *http.Request) {
		*gotAuth = r.Header.Get("Authorization")
		writeJSON(w, `{"_links":{"self":{"href":"x"}},"_embedded":{"items":[]}}`)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

// TestParityAkeneo_TokenEndpointFailureSurfacesAsAuthError asserts a token
// endpoint failure surfaces as an error on BOTH sides (never a silent
// unauthenticated request to the Akeneo resource API).
func TestParityAkeneo_TokenEndpointFailureSurfacesAsAuthError(t *testing.T) {
	bundle := loadAkeneoBundle(t)

	failingSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/oauth/v1/token" {
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": "invalid_grant"})
			return
		}
		t.Errorf("unexpected resource request %s despite a failed token exchange", r.URL.Path)
		writeJSON(w, `{"_embedded":{"items":[]}}`)
	}))
	t.Cleanup(failingSrv.Close)

	legacy := newAkeneoLegacyConnector(nil)
	legacyErr := legacy.Read(context.Background(), connectors.ReadRequest{Stream: "channels", Config: akeneoRuntimeConfig(failingSrv.URL, nil)}, func(connectors.Record) error { return nil })
	if legacyErr == nil {
		t.Fatal("legacy Read succeeded despite a failing token endpoint, want an error (test fixture bug)")
	}

	eng := newAkeneoEngineConnector(withAkeneoBaseURL(bundle, failingSrv.URL), newHooksWithClient(nil))
	engErr := eng.Read(context.Background(), connectors.ReadRequest{Stream: "channels", Config: akeneoRuntimeConfig(failingSrv.URL, nil)}, func(connectors.Record) error { return nil })
	if engErr == nil {
		t.Fatal("engine Read succeeded despite a failing token endpoint, want an error")
	}
}

// --- write parity: both sides reject writes (read-only connector) ---

func TestParityAkeneo_WriteUnsupportedOnBothSides(t *testing.T) {
	bundle := loadAkeneoBundle(t)

	legacy := akeneo.New()
	if _, err := legacy.Write(context.Background(), connectors.WriteRequest{}, nil); err == nil {
		t.Fatal("legacy Write succeeded, want ErrUnsupportedOperation (test fixture bug)")
	}

	eng := engine.New(bundle, engine.HooksFor("akeneo"))
	if _, err := eng.Write(context.Background(), connectors.WriteRequest{}, nil); err == nil {
		t.Fatal("engine Write succeeded, want an error (akeneo bundle declares capabilities.write: false, no writes.json)")
	}
	if bundle.Metadata.Capabilities.Write {
		t.Fatal("bundle metadata.capabilities.write = true, want false (akeneo is read-only)")
	}
	if len(bundle.Writes) != 0 {
		t.Fatalf("bundle write actions = %v, want none (akeneo is read-only, no writes.json)", bundle.Writes)
	}
}

// --- manifest-surface parity ---

func TestParityAkeneo_ManifestSurface(t *testing.T) {
	bundle := loadAkeneoBundle(t)

	legacyCatalog, err := akeneo.New().Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("legacy Catalog: %v", err)
	}

	eng := engine.New(bundle, engine.HooksFor("akeneo"))
	engManifest := connectors.ManifestOf(eng)

	wantStreams := akeneoManifestStreamSurface(legacyCatalog.Streams)
	gotStreams := akeneoManifestStreamSurface(engManifest.Streams)
	if !reflect.DeepEqual(gotStreams, wantStreams) {
		t.Fatalf("stream surface = %+v, want %+v (legacy Catalog())", gotStreams, wantStreams)
	}

	if len(engManifest.WriteActions) != 0 {
		t.Fatalf("engine write actions = %v, want none (akeneo is read-only)", engManifest.WriteActions)
	}
}

type akeneoStreamSurface struct {
	Name       string
	PrimaryKey []string
}

func akeneoManifestStreamSurface(streams []connectors.Stream) []akeneoStreamSurface {
	out := make([]akeneoStreamSurface, len(streams))
	for i, s := range streams {
		out[i] = akeneoStreamSurface{Name: s.Name, PrimaryKey: append([]string{}, s.PrimaryKey...)}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// --- bundle load smoke guard ---

func TestParityAkeneo_BundleLoadsAndValidates(t *testing.T) {
	bundle := loadAkeneoBundle(t)

	wantStreams := []string{"attributes", "categories", "channels", "families", "products"}
	gotStreams := make([]string, 0, len(bundle.Streams))
	for _, s := range bundle.Streams {
		gotStreams = append(gotStreams, s.Name)
	}
	sort.Strings(gotStreams)
	if !reflect.DeepEqual(gotStreams, wantStreams) {
		t.Fatalf("bundle streams = %v, want %v", gotStreams, wantStreams)
	}

	if len(bundle.Writes) != 0 {
		t.Fatalf("bundle write actions = %v, want none (akeneo is read-only, no writes.json)", bundle.Writes)
	}
	if bundle.Metadata.Capabilities.Write {
		t.Fatal("bundle metadata.capabilities.write = true, want false (akeneo has no mutation API)")
	}
}

// --- AuthSpec shape guard ---

// TestParityAkeneo_AuthSpecIsSoleCustomCandidate locks in "no roster swap
// needed": legacy has no alternate auth path, so the bundle declares exactly
// one auth candidate (mode custom, hook akeneo), not a when-gated fallback
// list.
func TestParityAkeneo_AuthSpecIsSoleCustomCandidate(t *testing.T) {
	bundle := loadAkeneoBundle(t)

	if len(bundle.HTTP.Auth) != 1 {
		t.Fatalf("len(bundle.HTTP.Auth) = %d, want 1 (no alternate auth path exists in legacy akeneo)", len(bundle.HTTP.Auth))
	}
	spec := bundle.HTTP.Auth[0]
	if spec.Mode != "custom" || spec.Hook != "akeneo" {
		t.Fatalf("auth spec = %+v, want mode=custom hook=akeneo", spec)
	}
}
