// Package chargebeeparity_test is the engine-vs-legacy parity suite for the
// chargebee pilot migration (PLAN.md P-6, SPEC.md §5.4). It lives in its own
// package (not internal/connectors/engine) per SPEC.md §6's parity-test
// location decision: per-connector directories give clean 10-way DW-1
// parallelism with no shared Go package namespace collisions across pilot
// agents.
//
// This is the RED-first test: it fails to even compile/load (engine.LoadAll
// cannot find a bundle named "chargebee") until internal/connectors/defs/chargebee
// exists with a full bundle.
package chargebeeparity_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/chargebee"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
)

// loadChargebeeBundle resolves the "chargebee" bundle from defs.FS via
// engine.LoadAll, the same discovery path TestConformance and every other
// production caller uses.
func loadChargebeeBundle(t *testing.T) engine.Bundle {
	t.Helper()
	bundles, err := engine.LoadAll(defs.FS)
	if err != nil {
		t.Fatalf("engine.LoadAll(defs.FS): %v", err)
	}
	for _, b := range bundles {
		if b.Name == "chargebee" {
			return b
		}
	}
	names := make([]string, 0, len(bundles))
	for _, b := range bundles {
		names = append(names, b.Name)
	}
	t.Fatalf("bundle %q not found in defs.FS (bundles: %v)", "chargebee", names)
	return engine.Bundle{}
}

// withChargebeeBaseURL returns a shallow copy of b with HTTP.URL pointed at
// baseURL (engine.Bundle is a value type; this never mutates the loaded
// original).
func withChargebeeBaseURL(b engine.Bundle, baseURL string) engine.Bundle {
	b.HTTP.URL = baseURL
	return b
}

// chargebeeRuntimeConfig builds the connectors.RuntimeConfig shared by both
// connectors under test.
func chargebeeRuntimeConfig(baseURL string, extra map[string]string) connectors.RuntimeConfig {
	cfg := map[string]string{"base_url": baseURL}
	for k, v := range extra {
		cfg[k] = v
	}
	return connectors.RuntimeConfig{
		Config:  cfg,
		Secrets: map[string]string{"site_api_key": "fixture_token_placeholder_123"},
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

// normalizeRecordStringify re-encodes every value to its string form so
// legacy's native Go numeric/boolean types (int64, bool from map literals)
// compare equal against the engine's computed_fields output, which always
// resolves to a string (engine.Interpolate's return type; see
// docs/migration/conventions.md §5 chargebee entry / this bundle's docs.md
// "Known limits" for the documented type-widening deviation: the envelope
// unwrap mechanism available to a Tier-1 bundle, per-field computed_fields,
// cannot preserve non-string raw JSON types when flattening a resource
// envelope). This is intentionally NOT the same raw-type DeepEqual bar
// stripe/searxng use — the deviation is real and is asserted explicitly by
// TestParityChargebee_ComputedFieldsStringifyNumericAndBooleanFields below,
// not hidden by this helper.
func normalizeRecordStringify(t *testing.T, r connectors.Record) map[string]any {
	t.Helper()
	out := make(map[string]any, len(r))
	for k, v := range r {
		out[k] = stringifyAny(v)
	}
	return out
}

func stringifyAny(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case json.Number:
		return t.String()
	default:
		raw, err := json.Marshal(t)
		if err != nil {
			return ""
		}
		var s string
		if err := json.Unmarshal(raw, &s); err == nil {
			return s
		}
		return string(raw)
	}
}

func normalizeRecordsStringify(t *testing.T, recs []connectors.Record) []map[string]any {
	t.Helper()
	out := make([]map[string]any, len(recs))
	for i, r := range recs {
		out[i] = normalizeRecordStringify(t, r)
	}
	return out
}

// --- stream fixtures: one deterministic page per stream, shaped exactly
// like legacy chargebee's real wire format (top-level {"list":[{<envelope>:
// {...}}, ...], "next_offset": "<token>"}) ---

func chargebeeStreamServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	mux.HandleFunc("/customers", func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("offset") {
		case "":
			writeJSON(w, `{"list":[
				{"customer":{"id":"cus_1","first_name":"Ada","last_name":"Lovelace","email":"ada@example.com","company":"Acme","phone":"+15550100","auto_collection":"on","net_term_days":0,"taxability":"taxable","created_at":1700000000,"updated_at":1700000000,"deleted":false}},
				{"customer":{"id":"cus_2","first_name":"Grace","last_name":"Hopper","email":"grace@example.com","company":"Acme","phone":"+15550101","auto_collection":"on","net_term_days":0,"taxability":"taxable","created_at":1700000100,"updated_at":1700000100,"deleted":false}}
			],"next_offset":"cus_2_offset"}`)
		case "cus_2_offset":
			writeJSON(w, `{"list":[
				{"customer":{"id":"cus_3","first_name":"Katherine","last_name":"Johnson","email":"katherine@example.com","company":"Acme","phone":"+15550102","auto_collection":"off","net_term_days":30,"taxability":"exempt","created_at":1700000200,"updated_at":1700000200,"deleted":true}}
			]}`)
		default:
			t.Errorf("unexpected offset=%q", r.URL.Query().Get("offset"))
			writeJSON(w, `{"list":[]}`)
		}
	})

	mux.HandleFunc("/subscriptions", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, `{"list":[
			{"subscription":{"id":"sub_1","customer_id":"cus_1","plan_id":"plan_basic","status":"active","currency_code":"USD","plan_quantity":1,"plan_amount":1000,"current_term_start":1700000000,"current_term_end":1702592000,"created_at":1700000000,"started_at":1700000000,"updated_at":1700000000,"deleted":false}}
		]}`)
	})

	mux.HandleFunc("/invoices", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, `{"list":[
			{"invoice":{"id":"inv_1","customer_id":"cus_1","subscription_id":"sub_1","status":"paid","currency_code":"USD","total":2000,"amount_paid":2000,"amount_due":0,"date":1700000000,"due_date":1700086400,"paid_at":1700003600,"updated_at":1700000000,"deleted":false}}
		]}`)
	})

	mux.HandleFunc("/plans", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, `{"list":[
			{"plan":{"id":"plan_basic","name":"Basic","invoice_name":"Basic Plan","price":1000,"currency_code":"USD","period":1,"period_unit":"month","status":"active","created_at":1700000000,"updated_at":1700000000}}
		]}`)
	})

	mux.HandleFunc("/items", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, `{"list":[
			{"item":{"id":"item_1","name":"Widget","type":"plan","item_family_id":"fam_1","status":"active","is_shippable":false,"enabled_for_checkout":true,"created_at":1700000000,"updated_at":1700000000}}
		]}`)
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func writeJSON(w http.ResponseWriter, body string) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(body))
}

// --- per-stream record parity across all 5 streams ---

func TestParityChargebee_StreamRecords(t *testing.T) {
	bundle := loadChargebeeBundle(t)

	streams := []string{"customers", "subscriptions", "invoices", "plans", "items"}
	for _, stream := range streams {
		stream := stream
		t.Run(stream, func(t *testing.T) {
			srv := chargebeeStreamServer(t)

			legacy := chargebee.New()
			legacyCfg := chargebeeRuntimeConfig(srv.URL, nil)
			legacyRecs := readAllRecords(t, legacy, connectors.ReadRequest{Stream: stream, Config: legacyCfg})

			eng := engine.New(withChargebeeBaseURL(bundle, srv.URL), nil)
			engCfg := chargebeeRuntimeConfig(srv.URL, nil)
			engRecs := readAllRecords(t, eng, connectors.ReadRequest{Stream: stream, Config: engCfg})

			if len(legacyRecs) == 0 {
				t.Fatalf("legacy chargebee emitted zero records for stream %q (test fixture bug)", stream)
			}
			if len(engRecs) != len(legacyRecs) {
				t.Fatalf("record count = %d, want %d (legacy)\nengine:  %+v\nlegacy:  %+v", len(engRecs), len(legacyRecs), engRecs, legacyRecs)
			}

			gotNorm := normalizeRecordsStringify(t, engRecs)
			wantNorm := normalizeRecordsStringify(t, legacyRecs)
			for i := range wantNorm {
				if !reflect.DeepEqual(gotNorm[i], wantNorm[i]) {
					t.Fatalf("stream %q record %d mismatch (stringified compare):\nengine:  %+v\nlegacy:  %+v", stream, i, gotNorm[i], wantNorm[i])
				}
			}
		})
	}
}

// TestParityChargebee_CustomersTwoPagePagination is the dedicated 2-page
// offset/next_offset assertion: 3 customers across 2 pages, identical
// emitted id sequence, and (via chargebeeStreamServer's t.Errorf on any
// OTHER offset value) the paginator must issue exactly the two requests
// legacy's own harvest() would issue.
func TestParityChargebee_CustomersTwoPagePagination(t *testing.T) {
	bundle := loadChargebeeBundle(t)
	srv := chargebeeStreamServer(t)

	legacy := chargebee.New()
	legacyRecs := readAllRecords(t, legacy, connectors.ReadRequest{Stream: "customers", Config: chargebeeRuntimeConfig(srv.URL, nil)})
	if len(legacyRecs) != 3 {
		t.Fatalf("legacy customers records = %d, want 3 (2 pages)", len(legacyRecs))
	}

	eng := engine.New(withChargebeeBaseURL(bundle, srv.URL), nil)
	engRecs := readAllRecords(t, eng, connectors.ReadRequest{Stream: "customers", Config: chargebeeRuntimeConfig(srv.URL, nil)})
	if len(engRecs) != 3 {
		t.Fatalf("engine customers records = %d, want 3 (2 pages)", len(engRecs))
	}

	gotIDs := recordIDs(t, engRecs)
	wantIDs := recordIDs(t, legacyRecs)
	if !reflect.DeepEqual(gotIDs, wantIDs) {
		t.Fatalf("customers record id sequence = %v, want %v", gotIDs, wantIDs)
	}
	if !reflect.DeepEqual(gotIDs, []string{"cus_1", "cus_2", "cus_3"}) {
		t.Fatalf("customers record id sequence = %v, want [cus_1 cus_2 cus_3]", gotIDs)
	}
}

func recordIDs(t *testing.T, recs []connectors.Record) []string {
	t.Helper()
	out := make([]string, len(recs))
	for i, r := range recs {
		out[i] = stringifyAny(r["id"])
	}
	return out
}

// --- incremental updated_at[after] propagation ---

// incrementalCaptureServer answers every request with an empty customers
// page (so the read terminates after exactly one request) and records the
// updated_at[after] query value it observed, for both connectors to be
// driven against identically.
func incrementalCaptureServer(t *testing.T) (*httptest.Server, *string) {
	t.Helper()
	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.URL.Query().Get("updated_at[after]")
		writeJSON(w, `{"list":[]}`)
	}))
	t.Cleanup(srv.Close)
	return srv, &got
}

// TestParityChargebee_IncrementalUpdatedAtFromState asserts both connectors
// arrive at the identical updated_at[after] WIRE VALUE (Unix seconds) when
// fed the SAME state cursor — the honest, app-persisted cursor shape (B1,
// REVIEW.md): internal/app/sync_modes.go's recordCursor stringifies a
// numeric "updated_at" field verbatim as a bare Unix-seconds digits string
// like "1700000100", never RFC3339. Both legacy (which forwards the state
// cursor to updated_at[after] verbatim, chargebee.go's incrementalLowerBound)
// and the engine (param_format: unix_seconds accepts digits-only input
// verbatim) must forward this SAME cursor shape identically.
func TestParityChargebee_IncrementalUpdatedAtFromState(t *testing.T) {
	bundle := loadChargebeeBundle(t)
	const appPersistedCursor = "1700000100" // internal/app's actual persisted-cursor shape: raw unix seconds digits

	legacySrv, legacyGot := incrementalCaptureServer(t)
	legacy := chargebee.New()
	_ = readAllRecords(t, legacy, connectors.ReadRequest{
		Stream: "customers",
		Config: chargebeeRuntimeConfig(legacySrv.URL, nil),
		State:  map[string]string{"cursor": appPersistedCursor},
	})

	engSrv, engGot := incrementalCaptureServer(t)
	eng := engine.New(withChargebeeBaseURL(bundle, engSrv.URL), nil)
	_ = readAllRecords(t, eng, connectors.ReadRequest{
		Stream: "customers",
		Config: chargebeeRuntimeConfig(engSrv.URL, nil),
		State:  map[string]string{"cursor": appPersistedCursor},
	})

	if *legacyGot != appPersistedCursor {
		t.Fatalf("legacy updated_at[after] = %q, want %q (test fixture bug)", *legacyGot, appPersistedCursor)
	}
	if *engGot != *legacyGot {
		t.Fatalf("engine updated_at[after] = %q, want %q (legacy, same app-persisted cursor)", *engGot, *legacyGot)
	}
}

// TestParityChargebee_IncrementalUpdatedAtFromStartDate exercises the
// config.start_date fallback path (RFC3339 -> unix seconds), matching
// legacy's incrementalLowerBound/formatParam-unix_seconds identically.
func TestParityChargebee_IncrementalUpdatedAtFromStartDate(t *testing.T) {
	bundle := loadChargebeeBundle(t)
	const startDate = "2025-06-15T00:00:00Z"
	const wantUnix = "1749945600"

	legacySrv, legacyGot := incrementalCaptureServer(t)
	legacy := chargebee.New()
	_ = readAllRecords(t, legacy, connectors.ReadRequest{
		Stream: "customers",
		Config: chargebeeRuntimeConfig(legacySrv.URL, map[string]string{"start_date": startDate}),
	})

	engSrv, engGot := incrementalCaptureServer(t)
	eng := engine.New(withChargebeeBaseURL(bundle, engSrv.URL), nil)
	_ = readAllRecords(t, eng, connectors.ReadRequest{
		Stream: "customers",
		Config: chargebeeRuntimeConfig(engSrv.URL, map[string]string{"start_date": startDate}),
	})

	if *legacyGot != wantUnix {
		t.Fatalf("legacy updated_at[after] = %q, want %q (test fixture bug)", *legacyGot, wantUnix)
	}
	if *engGot != *legacyGot {
		t.Fatalf("engine updated_at[after] = %q, want %q (legacy)", *engGot, *legacyGot)
	}
}

// --- auth header parity: HTTP Basic, site API key as username, empty password ---

func TestParityChargebee_BasicAuthHeader(t *testing.T) {
	bundle := loadChargebeeBundle(t)

	var legacyAuth, engAuth string
	legacySrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		legacyAuth = r.Header.Get("Authorization")
		writeJSON(w, `{"list":[]}`)
	}))
	defer legacySrv.Close()
	engSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		engAuth = r.Header.Get("Authorization")
		writeJSON(w, `{"list":[]}`)
	}))
	defer engSrv.Close()

	legacy := chargebee.New()
	_ = readAllRecords(t, legacy, connectors.ReadRequest{Stream: "customers", Config: chargebeeRuntimeConfig(legacySrv.URL, nil)})

	eng := engine.New(withChargebeeBaseURL(bundle, engSrv.URL), nil)
	_ = readAllRecords(t, eng, connectors.ReadRequest{Stream: "customers", Config: chargebeeRuntimeConfig(engSrv.URL, nil)})

	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("fixture_token_placeholder_123:"))
	if legacyAuth != wantAuth {
		t.Fatalf("legacy Authorization = %q, want %q (test fixture bug)", legacyAuth, wantAuth)
	}
	if engAuth != legacyAuth {
		t.Fatalf("engine Authorization = %q, want %q (legacy, byte-exact)", engAuth, legacyAuth)
	}
}

// --- error-path parity: non-2xx mapping ---

func TestParityChargebee_ErrorPathNon2xx(t *testing.T) {
	bundle := loadChargebeeBundle(t)

	legacySrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		writeJSON(w, `{"message":"invalid api key"}`)
	}))
	defer legacySrv.Close()
	engSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		writeJSON(w, `{"message":"invalid api key"}`)
	}))
	defer engSrv.Close()

	legacy := chargebee.New()
	legacyErr := legacy.Read(context.Background(), connectors.ReadRequest{Stream: "customers", Config: chargebeeRuntimeConfig(legacySrv.URL, nil)}, func(connectors.Record) error { return nil })
	if legacyErr == nil {
		t.Fatal("legacy Read on 401 = nil error, want non-nil (test fixture bug)")
	}

	eng := engine.New(withChargebeeBaseURL(bundle, engSrv.URL), nil)
	engErr := eng.Read(context.Background(), connectors.ReadRequest{Stream: "customers", Config: chargebeeRuntimeConfig(engSrv.URL, nil)}, func(connectors.Record) error { return nil })
	if engErr == nil {
		t.Fatal("engine Read on 401 = nil error, want non-nil (matching legacy's non-2xx failure)")
	}
}

// --- write parity: chargebee is read-only; both sides must reject Write identically ---

func TestParityChargebee_WriteUnsupported(t *testing.T) {
	bundle := loadChargebeeBundle(t)

	legacy := chargebee.New()
	_, legacyErr := legacy.Write(context.Background(), connectors.WriteRequest{Action: "create_customer"}, nil)
	if legacyErr == nil {
		t.Fatal("legacy Write = nil error, want ErrUnsupportedOperation (test fixture bug)")
	}

	eng := engine.New(bundle, nil)
	_, engErr := eng.Write(context.Background(), connectors.WriteRequest{Action: "create_customer"}, nil)
	if engErr == nil {
		t.Fatal("engine Write = nil error, want an error (bundle declares no writes.json / capabilities.write=false)")
	}
}

// --- catalog-surface parity ---

// TestParityChargebee_CatalogSurface compares Catalog() stream
// name/primary-key/cursor-field surface between legacy and the engine.
// chargebee.Connector has no hand-authored Manifest() (unlike stripe), so
// Catalog() — a method both sides genuinely implement — is the real
// "same stream set" parity claim here, not connectors.ManifestOf (which
// would silently fall back to a zero-stream default for legacy and produce
// a false failure unrelated to the migration's actual correctness).
func TestParityChargebee_CatalogSurface(t *testing.T) {
	bundle := loadChargebeeBundle(t)

	legacyCat, err := chargebee.New().Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("legacy Catalog: %v", err)
	}
	eng := engine.New(bundle, nil)
	engCat, err := eng.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("engine Catalog: %v", err)
	}

	wantStreams := manifestStreamSurface(legacyCat.Streams)
	gotStreams := manifestStreamSurface(engCat.Streams)
	if !reflect.DeepEqual(gotStreams, wantStreams) {
		t.Fatalf("stream surface = %+v, want %+v (legacy)", gotStreams, wantStreams)
	}
}

type streamSurface struct {
	Name         string
	PrimaryKey   []string
	CursorFields []string
}

func manifestStreamSurface(streams []connectors.Stream) []streamSurface {
	out := make([]streamSurface, len(streams))
	for i, s := range streams {
		out[i] = streamSurface{Name: s.Name, PrimaryKey: append([]string{}, s.PrimaryKey...), CursorFields: append([]string{}, s.CursorFields...)}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// TestParityChargebee_BundleLoadsAndValidates is a smoke guard: the bundle
// must load cleanly via engine.LoadAll(defs.FS) and declare exactly the 5
// legacy streams by name, with no write actions (chargebee is read-only).
func TestParityChargebee_BundleLoadsAndValidates(t *testing.T) {
	bundle := loadChargebeeBundle(t)

	wantStreams := []string{"customers", "invoices", "items", "plans", "subscriptions"}
	gotStreams := make([]string, 0, len(bundle.Streams))
	for _, s := range bundle.Streams {
		gotStreams = append(gotStreams, s.Name)
	}
	sort.Strings(gotStreams)
	if !reflect.DeepEqual(gotStreams, wantStreams) {
		t.Fatalf("bundle streams = %v, want %v", gotStreams, wantStreams)
	}

	if len(bundle.Writes) != 0 {
		t.Fatalf("bundle write actions = %v, want none (chargebee is read-only)", bundle.Writes)
	}
	if bundle.Metadata.Capabilities.Write {
		t.Fatal("bundle metadata.capabilities.write = true, want false (read-only connector)")
	}
}

// TestParityChargebee_ComputedFieldsPreserveNativeNumericAndBooleanTypes locks
// in the RESOLVED state of the formerly-documented parity deviation
// (docs/migration/conventions.md §5 chargebee entry, gap-loop cycle-1 A1):
// every one of chargebee's computed_fields templates is a single bare
// `{{ record.<envelope>.<field> }}` reference with no filter stage, so the
// engine's typed computed_fields extraction (engine/read.go's
// bareRecordPathReference + resolveRecordPathValue, gap-loop cycle-1 item 1)
// now copies the raw typed JSON value straight through instead of routing it
// through Interpolate's stringify — matching legacy's connsdk.RecordsAt
// decode (json.Number for a numeric field via json.Decoder.UseNumber, a
// native Go bool for a JSON boolean) byte-for-byte in TYPE as well as value.
// This test previously asserted the (now-superseded) stringified deviation;
// it is kept as a dedicated, named companion assertion per REVIEW-A.md A2
// rule 4 (a type-shape guarantee deserves a pinned test whether it documents
// a known deviation or proves one resolved) so no future engine change can
// silently regress chargebee back to string-typed numeric/boolean fields
// without a test catching it.
func TestParityChargebee_ComputedFieldsPreserveNativeNumericAndBooleanTypes(t *testing.T) {
	bundle := loadChargebeeBundle(t)
	srv := chargebeeStreamServer(t)

	legacy := chargebee.New()
	legacyRecs := readAllRecords(t, legacy, connectors.ReadRequest{Stream: "customers", Config: chargebeeRuntimeConfig(srv.URL, nil)})
	if len(legacyRecs) == 0 {
		t.Fatal("legacy emitted zero customers records (test fixture bug)")
	}
	if n, ok := legacyRecs[0]["created_at"].(json.Number); !ok || n.String() != "1700000000" {
		t.Fatalf("legacy customers[0].created_at = %#v (%T), want json.Number(\"1700000000\") (test fixture bug: connsdk's UseNumber decode)", legacyRecs[0]["created_at"], legacyRecs[0]["created_at"])
	}
	if _, ok := legacyRecs[0]["deleted"].(bool); !ok {
		t.Fatalf("legacy customers[0].deleted = %#v (%T), want bool (test fixture bug: legacy's native Go type)", legacyRecs[0]["deleted"], legacyRecs[0]["deleted"])
	}

	engSrv := chargebeeStreamServer(t)
	eng := engine.New(withChargebeeBaseURL(bundle, engSrv.URL), nil)
	engRecs := readAllRecords(t, eng, connectors.ReadRequest{Stream: "customers", Config: chargebeeRuntimeConfig(engSrv.URL, nil)})
	if len(engRecs) == 0 {
		t.Fatal("engine emitted zero customers records")
	}
	n, ok := engRecs[0]["created_at"].(json.Number)
	if !ok {
		t.Fatalf("engine customers[0].created_at = %#v (%T), want json.Number (typed computed_fields extraction; conventions.md §5 chargebee entry now RESOLVED)", engRecs[0]["created_at"], engRecs[0]["created_at"])
	}
	if n.String() != "1700000000" {
		t.Fatalf("engine customers[0].created_at = %q, want %q (same DATA, now same TYPE as legacy)", n.String(), "1700000000")
	}
	b, ok := engRecs[0]["deleted"].(bool)
	if !ok {
		t.Fatalf("engine customers[0].deleted = %#v (%T), want bool (typed computed_fields extraction)", engRecs[0]["deleted"], engRecs[0]["deleted"])
	}
	if b != false {
		t.Fatalf("engine customers[0].deleted = %v, want false (same DATA, now same TYPE as legacy)", b)
	}
}
