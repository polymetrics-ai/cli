package vitallyparity_test

// Engine-vs-legacy parity suite for the vitally pilot migration (PLAN.md P-2,
// SPEC.md §5.2). Drives BOTH connectors live against the SAME httptest
// server, asserting RAW connectors.Record equality (reflect.DeepEqual) plus
// legacy-side sanity assertions, following the wave0 goldens
// (internal/connectors/engine/parity_stripe_test.go,
// parity_searxng_test.go). Red-first protocol (conventions.md, PLAN.md P-1..
// P-10): this file was written FIRST and fails because
// internal/connectors/defs/vitally does not exist yet — see
// traces/p2-vitally-ledger.md for the recorded RED output.
//
// Legacy read (internal/connectors/vitally/vitally.go, read-only reference):
//   - auth: connsdk.APIKeyHeader("Authorization", auth, "") where auth is the
//     secret basic_auth_header verbatim (vitally.go:104) — the secret already
//     contains the FULL header value (e.g. "Basic dGVzdDp0ZXN0"), not just a
//     token; APIKeyHeader with an empty prefix sets the header to auth
//     unmodified. This is why the bundle must use api_key_header (not
//     engine "basic", which would base64-encode a username/password pair
//     the legacy connector never had) — see SPEC.md §5.2's "byte-exact
//     Authorization header parity" requirement.
//   - single stream "accounts": GET resources/accounts, optional "status"
//     query param (only sent when configured — vitally.go:72-74), records
//     extracted from top-level "results" (vitally.go:79), each record mapped
//     to exactly {id, name, traits} (vitally.go:84) — schema-as-projection,
//     no computed_fields renames needed.
//   - no pagination, no incremental, read-only (Write returns
//     ErrUnsupportedOperation, vitally.go:92).
//   - error path: any transport/non-2xx response surfaces as a Go error from
//     legacy's r.Do wrapped "read vitally accounts: %w"; the engine bundle
//     models this with metadata's default error handling (no bundle-level
//     error_map override changes emitted record data for any 2xx response,
//     so no ledger entry is needed there — a non-2xx response never emits
//     records on either side, which is exactly what this suite asserts).

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
	"polymetrics.ai/internal/connectors/vitally"
)

// loadVitallyBundle resolves the "vitally" bundle from defs.FS via
// engine.Load, the same discovery path TestConformance and every other
// production caller uses.
func loadVitallyBundle(t *testing.T) engine.Bundle {
	t.Helper()
	b, err := engine.Load(defs.FS, "vitally")
	if err != nil {
		t.Fatalf("engine.Load(defs.FS, %q): %v", "vitally", err)
	}
	return b
}

func withVitallyBaseURL(b engine.Bundle, baseURL string) engine.Bundle {
	b.HTTP.URL = baseURL
	return b
}

func vitallyRuntimeConfig(baseURL string, extra map[string]string) connectors.RuntimeConfig {
	cfg := map[string]string{"base_url": baseURL}
	for k, v := range extra {
		cfg[k] = v
	}
	return connectors.RuntimeConfig{Config: cfg, Secrets: map[string]string{"basic_auth_header": "Basic dGVzdDp0ZXN0"}}
}

func readAllVitallyRecords(t *testing.T, c connectors.Connector, req connectors.ReadRequest) []connectors.Record {
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

// --- per-stream record parity: "accounts" ---

func TestParityVitally_AccountsStreamRecords(t *testing.T) {
	bundle := loadVitallyBundle(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/resources/accounts" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"results":[{"id":"acct_1","name":"Acme","traits":{"plan":"pro"}},{"id":"acct_2","name":"Globex","traits":{"plan":"free"}}]}`))
	}))
	defer srv.Close()

	legacy := vitally.New()
	legacyRecs := readAllVitallyRecords(t, legacy, connectors.ReadRequest{
		Stream: "accounts",
		Config: vitallyRuntimeConfig(srv.URL, nil),
	})
	if len(legacyRecs) == 0 {
		t.Fatal("legacy vitally emitted zero records for stream accounts (test fixture bug)")
	}

	eng := engine.New(withVitallyBaseURL(bundle, srv.URL), nil)
	engRecs := readAllVitallyRecords(t, eng, connectors.ReadRequest{
		Stream: "accounts",
		Config: vitallyRuntimeConfig(srv.URL, nil),
	})

	if len(engRecs) != len(legacyRecs) {
		t.Fatalf("record count = %d, want %d (legacy)\nengine: %+v\nlegacy: %+v", len(engRecs), len(legacyRecs), engRecs, legacyRecs)
	}
	for i := range legacyRecs {
		if !reflect.DeepEqual(engRecs[i], legacyRecs[i]) {
			t.Fatalf("record %d mismatch:\nengine:  %+v\nlegacy:  %+v", i, engRecs[i], legacyRecs[i])
		}
	}
}

// --- base-case query parity: no "status" param sent when unset ---

// TestParityVitally_NoStatusParamSentWhenUnset asserts the BASE case both
// sides agree on: legacy only appends ?status=<value> when a non-empty
// status config is set (vitally.go:72-74); when it is unset (the common
// case, and the only case this bundle's spec.json even allows since
// "status" is deliberately not declared as a config property — see docs.md
// "Known limits"), neither side sends a status query param at all.
//
// The status-CONFIGURED case is a documented, out-of-scope deviation, not a
// silently-broken parity claim: the engine dialect's stream.Query templating
// has no absent-key-falsy tolerance (conventions.md §3), so an unconditional
// "status": "{{ config.status }}" would hard-error whenever status is unset
// — the common path. This mirrors searxng's identical, already-accepted
// "optional passthrough filter" limitation (docs/migration/conventions.md
// §3's stream.Query note; searxng's own docs.md "Known limits"). Asserting a
// status-configured parity claim the bundle cannot honestly satisfy would be
// exactly the "weakened assertion to get green" conventions.md forbids in
// the OTHER direction — this test instead proves the accepted base case.
func TestParityVitally_NoStatusParamSentWhenUnset(t *testing.T) {
	bundle := loadVitallyBundle(t)

	legacySawStatus := false
	legacySrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := r.URL.Query()["status"]; ok {
			legacySawStatus = true
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"results":[{"id":"acct_1","name":"Acme","traits":{"plan":"pro"}}]}`))
	}))
	defer legacySrv.Close()

	legacy := vitally.New()
	_ = readAllVitallyRecords(t, legacy, connectors.ReadRequest{
		Stream: "accounts",
		Config: vitallyRuntimeConfig(legacySrv.URL, nil),
	})
	if legacySawStatus {
		t.Fatal("legacy sent a status query param with no status configured (test fixture bug)")
	}

	engSawStatus := false
	engSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := r.URL.Query()["status"]; ok {
			engSawStatus = true
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"results":[{"id":"acct_1","name":"Acme","traits":{"plan":"pro"}}]}`))
	}))
	defer engSrv.Close()

	eng := engine.New(withVitallyBaseURL(bundle, engSrv.URL), nil)
	_ = readAllVitallyRecords(t, eng, connectors.ReadRequest{
		Stream: "accounts",
		Config: vitallyRuntimeConfig(engSrv.URL, nil),
	})

	if engSawStatus {
		t.Fatal("engine sent a status query param with no status configured, want none (bundle declares no status query wiring)")
	}
}

// --- byte-exact Authorization header parity (SPEC.md §5.2) ---

// TestParityVitally_AuthorizationHeaderByteExact locks in SPEC.md §5.2's
// explicit requirement: legacy's basic_auth_header secret already contains
// the FULL Authorization header value (vitally.go:100-104
// connsdk.APIKeyHeader("Authorization", auth, "")); the bundle must reproduce
// this verbatim, not re-derive/re-encode it (an engine "basic" auth mode
// would base64-encode a username:password pair vitally never had).
func TestParityVitally_AuthorizationHeaderByteExact(t *testing.T) {
	bundle := loadVitallyBundle(t)
	const rawHeaderValue = "Basic dXNlcjpwYXNz"

	legacyAuth := ""
	legacySrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		legacyAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"results":[]}`))
	}))
	defer legacySrv.Close()

	legacy := vitally.New()
	legacyCfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": legacySrv.URL}, Secrets: map[string]string{"basic_auth_header": rawHeaderValue}}
	_ = readAllVitallyRecords(t, legacy, connectors.ReadRequest{Stream: "accounts", Config: legacyCfg})
	if legacyAuth != rawHeaderValue {
		t.Fatalf("legacy Authorization = %q, want %q (test fixture bug)", legacyAuth, rawHeaderValue)
	}

	engAuth := ""
	engSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		engAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"results":[]}`))
	}))
	defer engSrv.Close()

	eng := engine.New(withVitallyBaseURL(bundle, engSrv.URL), nil)
	engCfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": engSrv.URL}, Secrets: map[string]string{"basic_auth_header": rawHeaderValue}}
	_ = readAllVitallyRecords(t, eng, connectors.ReadRequest{Stream: "accounts", Config: engCfg})

	if engAuth != rawHeaderValue {
		t.Fatalf("engine Authorization = %q, want %q (byte-exact, matching legacy's verbatim secret pass-through)", engAuth, rawHeaderValue)
	}
	if engAuth != legacyAuth {
		t.Fatalf("engine Authorization = %q, legacy Authorization = %q; must be byte-identical", engAuth, legacyAuth)
	}
}

// --- error-path parity: non-2xx response emits no records, surfaces an error ---

func TestParityVitally_NonSuccessStatusErrorsOnBothSides(t *testing.T) {
	bundle := loadVitallyBundle(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
	}))
	defer srv.Close()

	legacy := vitally.New()
	var legacyRecs []connectors.Record
	legacyErr := legacy.Read(context.Background(), connectors.ReadRequest{
		Stream: "accounts",
		Config: vitallyRuntimeConfig(srv.URL, nil),
	}, func(r connectors.Record) error {
		legacyRecs = append(legacyRecs, r)
		return nil
	})
	if legacyErr == nil {
		t.Fatal("legacy Read returned nil error on a 401 response (test fixture bug)")
	}
	if len(legacyRecs) != 0 {
		t.Fatalf("legacy emitted %d records on a 401 response, want 0", len(legacyRecs))
	}

	eng := engine.New(withVitallyBaseURL(bundle, srv.URL), nil)
	var engRecs []connectors.Record
	engErr := eng.Read(context.Background(), connectors.ReadRequest{
		Stream: "accounts",
		Config: vitallyRuntimeConfig(srv.URL, nil),
	}, func(r connectors.Record) error {
		engRecs = append(engRecs, r)
		return nil
	})
	if engErr == nil {
		t.Fatal("engine Read returned nil error on a 401 response, want a non-nil error (parity with legacy)")
	}
	if len(engRecs) != 0 {
		t.Fatalf("engine emitted %d records on a 401 response, want 0", len(engRecs))
	}
}

// --- capabilities / write-unsupported parity ---

func TestParityVitally_WriteUnsupportedOnBothSides(t *testing.T) {
	bundle := loadVitallyBundle(t)

	legacy := vitally.New()
	if _, err := legacy.Write(context.Background(), connectors.WriteRequest{}, nil); err != connectors.ErrUnsupportedOperation {
		t.Fatalf("legacy Write error = %v, want ErrUnsupportedOperation", err)
	}
	if legacy.Metadata().Capabilities.Write {
		t.Fatal("legacy capabilities.write = true, want false (read-only connector)")
	}

	eng := engine.New(bundle, nil)
	if _, err := eng.Write(context.Background(), connectors.WriteRequest{}, nil); err != connectors.ErrUnsupportedOperation {
		t.Fatalf("engine Write error = %v, want ErrUnsupportedOperation", err)
	}
	if eng.Metadata().Capabilities.Write {
		t.Fatal("engine capabilities.write = true, want false (bundle metadata.json capabilities.write must be false, no writes.json)")
	}
	if len(bundle.Writes) != 0 {
		t.Fatalf("bundle write actions = %v, want none (vitally is read-only, no writes.json)", bundle.Writes)
	}
}

// --- bundle load / manifest surface smoke guard ---

func TestParityVitally_BundleLoadsWithSingleAccountsStream(t *testing.T) {
	bundle := loadVitallyBundle(t)

	if len(bundle.Streams) != 1 || bundle.Streams[0].Name != "accounts" {
		names := make([]string, 0, len(bundle.Streams))
		for _, s := range bundle.Streams {
			names = append(names, s.Name)
		}
		t.Fatalf("bundle streams = %v, want exactly [accounts]", names)
	}

	legacy := vitally.New()
	legacyCatalog, err := legacy.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("legacy Catalog: %v", err)
	}
	if len(legacyCatalog.Streams) != 1 || legacyCatalog.Streams[0].Name != "accounts" {
		t.Fatalf("legacy catalog streams = %+v, want exactly [accounts]", legacyCatalog.Streams)
	}

	schema, ok := bundle.Schemas["accounts"]
	if !ok {
		t.Fatal("bundle has no compiled schema for stream accounts")
	}
	if !reflect.DeepEqual(schema.PrimaryKey, []string{"id"}) {
		t.Fatalf("accounts x-primary-key = %v, want [id]", schema.PrimaryKey)
	}
}

// --- basic_auth_header secret required parity (Check) ---

func TestParityVitally_CheckRequiresAuthSecretOnBothSides(t *testing.T) {
	bundle := loadVitallyBundle(t)

	legacy := vitally.New()
	legacyErr := legacy.Check(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"base_url": "https://example.invalid"}})
	if legacyErr == nil {
		t.Fatal("legacy Check with no basic_auth_header secret = nil error, want an error (test fixture bug)")
	}
	if !strings.Contains(legacyErr.Error(), "basic_auth_header") {
		t.Fatalf("legacy Check error = %q, want it to name basic_auth_header", legacyErr.Error())
	}

	eng := engine.New(withVitallyBaseURL(bundle, "https://example.invalid"), nil)
	engErr := eng.Check(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"base_url": "https://example.invalid"}})
	if engErr == nil {
		t.Fatal("engine Check with no basic_auth_header secret = nil error, want an error (parity with legacy)")
	}
}
