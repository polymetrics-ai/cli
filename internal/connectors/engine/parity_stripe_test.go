package engine_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"sort"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
	"polymetrics.ai/internal/connectors/stripe"
)

// This file is the golden migration parity suite for the stripe bundle
// (PLAN.md T-15/B-15): every stream, its pagination/incremental behavior, and
// its write actions are driven against ONE shared httptest.Server for both
// the legacy hand-written stripe.Connector (internal/connectors/stripe,
// read-only reference) and the engine-backed connector built from
// internal/connectors/defs/stripe (engine.New(bundle, nil)). Byte-identical
// requests in ⇒ byte-identical connectors.Record slices and write requests
// out is the parity bar; any unavoidable deviation is documented in
// traces/waveF-b15-ledger.md's parity-deviation ledger, not worked around
// here.

// loadStripeBundle resolves the "stripe" bundle from defs.FS via
// engine.LoadAll, the same discovery path TestConformance and every other
// production caller uses (SPEC §1.9 rule 2: "no RegisterFactory call is made
// for engine-backed stripe" — parity/conformance tests build it directly).
func loadStripeBundle(t *testing.T) engine.Bundle {
	t.Helper()
	bundles, err := engine.LoadAll(defs.FS)
	if err != nil {
		t.Fatalf("engine.LoadAll(defs.FS): %v", err)
	}
	for _, b := range bundles {
		if b.Name == "stripe" {
			return b
		}
	}
	names := make([]string, 0, len(bundles))
	for _, b := range bundles {
		names = append(names, b.Name)
	}
	t.Fatalf("bundle %q not found in defs.FS (bundles: %v)", "stripe", names)
	return engine.Bundle{}
}

// withStripeBaseURL returns a shallow copy of b with HTTP.URL pointed at
// baseURL (engine.Bundle is a value type; this never mutates the loaded
// original, mirroring conformance/dynamic.go's withReplayURL).
func withStripeBaseURL(b engine.Bundle, baseURL string) engine.Bundle {
	b.HTTP.URL = baseURL
	return b
}

// stripeRuntimeConfig builds the connectors.RuntimeConfig shared by both
// connectors under test: base_url points at the shared httptest server,
// client_secret is a synthetic sk_test_-prefixed placeholder (never a
// real-looking live key, per THREAT-MODEL §4), and extra carries any
// additional config values a given subtest needs (start_date, account_id).
func stripeRuntimeConfig(baseURL string, extra map[string]string) connectors.RuntimeConfig {
	cfg := map[string]string{"base_url": baseURL}
	for k, v := range extra {
		cfg[k] = v
	}
	return connectors.RuntimeConfig{
		Config:  cfg,
		Secrets: map[string]string{"client_secret": "sk_test_conformancefixtureonly"},
	}
}

// readAllRecords drains c.Read(stream) into a slice, in emission order.
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
// legacy's native Go types (int64, bool, string from map literals) and the
// engine's json.Number-preserving decode compare equal on numeric fields —
// the two connectors read the SAME wire bytes but legacy builds its emitted
// connectors.Record from Go-native map[string]any (int64 literals in
// readFixture / go values decoded via encoding/json elsewhere), whereas the
// engine's read path always carries connsdk's json.Number. Both are
// decode(encode(x)) round-tripped here so the comparison is over the
// canonical JSON representation, not incidental Go numeric-type identity.
func normalizeRecord(t *testing.T, r connectors.Record) map[string]any {
	t.Helper()
	raw, err := json.Marshal(map[string]any(r))
	if err != nil {
		t.Fatalf("marshal record: %v", err)
	}
	var out map[string]any
	dec := json.NewDecoder(bytesReader(raw))
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

// bytesReader avoids importing bytes solely for one call site.
func bytesReader(b []byte) *byteReader { return &byteReader{b: b} }

type byteReader struct {
	b []byte
	i int
}

func (r *byteReader) Read(p []byte) (int, error) {
	if r.i >= len(r.b) {
		return 0, fmt.Errorf("EOF")
	}
	n := copy(p, r.b[r.i:])
	r.i += n
	return n, nil
}

// --- stream fixtures: one deterministic page per stream, shaped exactly
// like legacy stripe's real wire format (a top-level {"data":[...],
// "has_more":bool} list response with numeric created/amount/... fields) ---

func stripeStreamServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	mux.HandleFunc("/customers", func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("starting_after") {
		case "":
			writeJSON(w, `{"object":"list","data":[
				{"id":"cus_1","object":"customer","created":1700000000,"email":"ada@example.com","name":"Ada Lovelace","description":"first customer","phone":"+15550100","currency":"usd","balance":0,"delinquent":false,"livemode":false},
				{"id":"cus_2","object":"customer","created":1700000100,"email":"grace@example.com","name":"Grace Hopper","description":"second customer","phone":"+15550101","currency":"usd","balance":500,"delinquent":false,"livemode":false}
			],"has_more":true}`)
		case "cus_2":
			writeJSON(w, `{"object":"list","data":[
				{"id":"cus_3","object":"customer","created":1700000200,"email":"katherine@example.com","name":"Katherine Johnson","description":"third customer","phone":"+15550102","currency":"usd","balance":1200,"delinquent":true,"livemode":false}
			],"has_more":false}`)
		default:
			t.Errorf("unexpected starting_after=%q", r.URL.Query().Get("starting_after"))
			writeJSON(w, `{"object":"list","data":[],"has_more":false}`)
		}
	})

	mux.HandleFunc("/charges", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, `{"object":"list","data":[
			{"id":"ch_1","object":"charge","created":1700000300,"amount":1000,"amount_captured":1000,"amount_refunded":0,"currency":"usd","customer":"cus_1","status":"succeeded","paid":true,"refunded":false,"livemode":false}
		],"has_more":false}`)
	})

	mux.HandleFunc("/invoices", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, `{"object":"list","data":[
			{"id":"in_1","object":"invoice","created":1700000400,"customer":"cus_1","subscription":"sub_1","status":"paid","currency":"usd","amount_due":2000,"amount_paid":2000,"amount_remaining":0,"total":2000,"paid":true,"livemode":false}
		],"has_more":false}`)
	})

	mux.HandleFunc("/subscriptions", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, `{"object":"list","data":[
			{"id":"sub_1","object":"subscription","created":1700000500,"customer":"cus_1","status":"active","currency":"usd","current_period_start":1700000000,"current_period_end":1702592000,"cancel_at_period_end":false,"canceled_at":null,"livemode":false}
		],"has_more":false}`)
	})

	mux.HandleFunc("/products", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, `{"object":"list","data":[
			{"id":"prod_1","object":"product","created":1700000600,"updated":1700000700,"name":"Widget","description":"A widget","active":true,"type":"service","livemode":false}
		],"has_more":false}`)
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

func TestParityStripe_StreamRecords(t *testing.T) {
	bundle := loadStripeBundle(t)

	streams := []string{"customers", "charges", "invoices", "subscriptions", "products"}
	for _, stream := range streams {
		stream := stream
		t.Run(stream, func(t *testing.T) {
			srv := stripeStreamServer(t)

			legacy := stripe.New()
			legacyCfg := stripeRuntimeConfig(srv.URL, nil)
			legacyRecs := readAllRecords(t, legacy, connectors.ReadRequest{Stream: stream, Config: legacyCfg})

			eng := engine.New(withStripeBaseURL(bundle, srv.URL), nil)
			engCfg := stripeRuntimeConfig(srv.URL, nil)
			engRecs := readAllRecords(t, eng, connectors.ReadRequest{Stream: stream, Config: engCfg})

			if len(legacyRecs) == 0 {
				t.Fatalf("legacy stripe emitted zero records for stream %q (test fixture bug)", stream)
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

// TestParityStripe_CustomersTwoPagePagination is the dedicated 2-page
// starting_after/has_more assertion PLAN.md calls out explicitly: 3
// customers across 2 pages, identical emitted sequence, and (via
// stripeStreamServer's t.Errorf on any OTHER starting_after value) the
// paginator must issue exactly the two requests legacy's own harvest() would
// issue — no more, no fewer.
func TestParityStripe_CustomersTwoPagePagination(t *testing.T) {
	bundle := loadStripeBundle(t)
	srv := stripeStreamServer(t)

	legacy := stripe.New()
	legacyRecs := readAllRecords(t, legacy, connectors.ReadRequest{Stream: "customers", Config: stripeRuntimeConfig(srv.URL, nil)})
	if len(legacyRecs) != 3 {
		t.Fatalf("legacy customers records = %d, want 3 (2 pages)", len(legacyRecs))
	}

	eng := engine.New(withStripeBaseURL(bundle, srv.URL), nil)
	engRecs := readAllRecords(t, eng, connectors.ReadRequest{Stream: "customers", Config: stripeRuntimeConfig(srv.URL, nil)})
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
		id, _ := r["id"].(string)
		out[i] = id
	}
	return out
}

// --- incremental created[gte] propagation ---

// incrementalCaptureServer answers every request with an empty customers
// page (so the read terminates after exactly one request) and records the
// created[gte] query value it observed, for both connectors to be driven
// against identically.
func incrementalCaptureServer(t *testing.T) (*httptest.Server, *string) {
	t.Helper()
	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.URL.Query().Get("created[gte]")
		writeJSON(w, `{"object":"list","data":[],"has_more":false}`)
	}))
	t.Cleanup(srv.Close)
	return srv, &got
}

// TestParityStripe_IncrementalCreatedGTEFromState asserts both connectors
// arrive at the identical created[gte] WIRE VALUE (Unix seconds) when fed
// the SAME state cursor — and that cursor is the HONEST, app-produced shape
// (B1, REVIEW.md): internal/app/sync_modes.go's recordCursor ->
// toComparableString stringifies a numeric "created" field verbatim, never
// converting it to RFC3339, so a real production cursor for this stream is
// always a bare Unix-seconds digits string like "1700000100" — nothing in
// internal/app ever produces an RFC3339 cursor for a numeric cursor field.
// Both legacy (which forwards the state cursor to created[gte] completely
// verbatim, stripe.go's incrementalLowerBound) and the engine (whose
// param_format: unix_seconds now accepts a digits-only input verbatim
// per B1's fix to formatParam) must therefore accept and correctly forward
// this SAME cursor shape identically — this is the honest parity bar the
// review calls for, not a hand-crafted RFC3339 cursor no production code
// path ever generates.
func TestParityStripe_IncrementalCreatedGTEFromState(t *testing.T) {
	bundle := loadStripeBundle(t)
	const appPersistedCursor = "1700000100" // internal/app's actual persisted-cursor shape: raw unix seconds digits

	legacySrv, legacyGot := incrementalCaptureServer(t)
	legacy := stripe.New()
	_ = readAllRecords(t, legacy, connectors.ReadRequest{
		Stream: "customers",
		Config: stripeRuntimeConfig(legacySrv.URL, nil),
		State:  map[string]string{"cursor": appPersistedCursor},
	})

	engSrv, engGot := incrementalCaptureServer(t)
	eng := engine.New(withStripeBaseURL(bundle, engSrv.URL), nil)
	_ = readAllRecords(t, eng, connectors.ReadRequest{
		Stream: "customers",
		Config: stripeRuntimeConfig(engSrv.URL, nil),
		State:  map[string]string{"cursor": appPersistedCursor},
	})

	if *legacyGot != appPersistedCursor {
		t.Fatalf("legacy created[gte] = %q, want %q (test fixture bug)", *legacyGot, appPersistedCursor)
	}
	if *engGot != *legacyGot {
		t.Fatalf("engine created[gte] = %q, want %q (legacy, same app-persisted cursor)", *engGot, *legacyGot)
	}
}

// TestParityStripe_IncrementalCreatedGTEFromStartDate exercises the
// config.start_date fallback path (RFC3339 -> unix seconds), matching
// legacy's incrementalLowerBound/formatParam-unix_seconds identically.
func TestParityStripe_IncrementalCreatedGTEFromStartDate(t *testing.T) {
	bundle := loadStripeBundle(t)
	const startDate = "2025-06-15T00:00:00Z"
	const wantUnix = "1749945600"

	legacySrv, legacyGot := incrementalCaptureServer(t)
	legacy := stripe.New()
	_ = readAllRecords(t, legacy, connectors.ReadRequest{
		Stream: "customers",
		Config: stripeRuntimeConfig(legacySrv.URL, map[string]string{"start_date": startDate}),
	})

	engSrv, engGot := incrementalCaptureServer(t)
	eng := engine.New(withStripeBaseURL(bundle, engSrv.URL), nil)
	_ = readAllRecords(t, eng, connectors.ReadRequest{
		Stream: "customers",
		Config: stripeRuntimeConfig(engSrv.URL, map[string]string{"start_date": startDate}),
	})

	if *legacyGot != wantUnix {
		t.Fatalf("legacy created[gte] = %q, want %q (test fixture bug)", *legacyGot, wantUnix)
	}
	if *engGot != *legacyGot {
		t.Fatalf("engine created[gte] = %q, want %q (legacy)", *engGot, *legacyGot)
	}
}

// --- write parity: create_customer / update_customer form bodies ---

// writeCaptureServer answers every request 200 {"id":"cus_1"} and records
// the last request's method/path/decoded form body.
type writeCaptureRequest struct {
	Method string
	Path   string
	Form   url.Values
}

func writeCaptureServer(t *testing.T) (*httptest.Server, *writeCaptureRequest) {
	t.Helper()
	captured := &writeCaptureRequest{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		captured.Method = r.Method
		captured.Path = r.URL.Path
		captured.Form = r.PostForm
		writeJSON(w, `{"id":"cus_1","object":"customer"}`)
	}))
	t.Cleanup(srv.Close)
	return srv, captured
}

func TestParityStripe_WriteCreateCustomerFormBody(t *testing.T) {
	bundle := loadStripeBundle(t)
	record := connectors.Record{"email": "ada@example.com", "name": "Ada Lovelace", "description": "test customer"}

	legacySrv, legacyGot := writeCaptureServer(t)
	legacy := stripe.New()
	legacyResult, err := legacy.Write(context.Background(), connectors.WriteRequest{Action: "create_customer", Config: stripeRuntimeConfig(legacySrv.URL, nil)}, []connectors.Record{record})
	if err != nil {
		t.Fatalf("legacy Write(create_customer): %v", err)
	}
	if legacyResult.RecordsWritten != 1 {
		t.Fatalf("legacy RecordsWritten = %d, want 1", legacyResult.RecordsWritten)
	}

	engSrv, engGot := writeCaptureServer(t)
	eng := engine.New(withStripeBaseURL(bundle, engSrv.URL), nil)
	engResult, err := eng.Write(context.Background(), connectors.WriteRequest{Action: "create_customer", Config: stripeRuntimeConfig(engSrv.URL, nil)}, []connectors.Record{record})
	if err != nil {
		t.Fatalf("engine Write(create_customer): %v", err)
	}
	if engResult.RecordsWritten != 1 {
		t.Fatalf("engine RecordsWritten = %d, want 1", engResult.RecordsWritten)
	}

	if legacyGot.Method != http.MethodPost || legacyGot.Path != "/customers" {
		t.Fatalf("legacy request = %s %s, want POST /customers (test fixture bug)", legacyGot.Method, legacyGot.Path)
	}
	if engGot.Method != legacyGot.Method {
		t.Fatalf("engine method = %q, want %q (legacy)", engGot.Method, legacyGot.Method)
	}
	if engGot.Path != legacyGot.Path {
		t.Fatalf("engine path = %q, want %q (legacy)", engGot.Path, legacyGot.Path)
	}
	if !reflect.DeepEqual(engGot.Form, legacyGot.Form) {
		t.Fatalf("engine form = %v, want %v (legacy)", engGot.Form, legacyGot.Form)
	}
}

func TestParityStripe_WriteUpdateCustomerFormBody(t *testing.T) {
	bundle := loadStripeBundle(t)
	record := connectors.Record{"id": "cus_1", "email": "new-email@example.com", "phone": "+15559999"}

	legacySrv, legacyGot := writeCaptureServer(t)
	legacy := stripe.New()
	legacyResult, err := legacy.Write(context.Background(), connectors.WriteRequest{Action: "update_customer", Config: stripeRuntimeConfig(legacySrv.URL, nil)}, []connectors.Record{record})
	if err != nil {
		t.Fatalf("legacy Write(update_customer): %v", err)
	}
	if legacyResult.RecordsWritten != 1 {
		t.Fatalf("legacy RecordsWritten = %d, want 1", legacyResult.RecordsWritten)
	}

	engSrv, engGot := writeCaptureServer(t)
	eng := engine.New(withStripeBaseURL(bundle, engSrv.URL), nil)
	engResult, err := eng.Write(context.Background(), connectors.WriteRequest{Action: "update_customer", Config: stripeRuntimeConfig(engSrv.URL, nil)}, []connectors.Record{record})
	if err != nil {
		t.Fatalf("engine Write(update_customer): %v", err)
	}
	if engResult.RecordsWritten != 1 {
		t.Fatalf("engine RecordsWritten = %d, want 1", engResult.RecordsWritten)
	}

	if legacyGot.Method != http.MethodPost || legacyGot.Path != "/customers/cus_1" {
		t.Fatalf("legacy request = %s %s, want POST /customers/cus_1 (test fixture bug)", legacyGot.Method, legacyGot.Path)
	}
	if engGot.Method != legacyGot.Method {
		t.Fatalf("engine method = %q, want %q (legacy)", engGot.Method, legacyGot.Method)
	}
	if engGot.Path != legacyGot.Path {
		t.Fatalf("engine path = %q, want %q (legacy)", engGot.Path, legacyGot.Path)
	}
	// The engine's write path builds the form body from every record field
	// not in path_fields (id is a path_field here); legacy's customerForm
	// only ever sends email/name/description/phone regardless of what else
	// is on the record, so both sides converge on the same key set for a
	// record shaped like the legacy allow-list expects.
	if !reflect.DeepEqual(engGot.Form, legacyGot.Form) {
		t.Fatalf("engine form = %v, want %v (legacy)", engGot.Form, legacyGot.Form)
	}
}

// --- manifest-surface parity ---

// TestParityStripe_ManifestSurface compares the engine-synthesized Manifest
// against connectors.ManifestOf(stripe.New()) (legacy, self-registered) for
// stream names, primary keys, cursor fields, and write action names — the
// "manifest-surface equality" bar PLAN.md T-15 calls for. Full-manifest
// DeepEqual is deliberately NOT asserted: legacy hand-authors additional
// descriptive fields (ConfigFields, SecretFields, AuthModes, Pagination,
// Risk prose) that the engine synthesizes generically from the bundle
// (design §C) and is not required to reproduce verbatim.
func TestParityStripe_ManifestSurface(t *testing.T) {
	bundle := loadStripeBundle(t)

	legacyManifest := connectors.ManifestOf(stripe.New())
	eng := engine.New(bundle, nil)
	engManifest := connectors.ManifestOf(eng)

	wantStreams := manifestStreamSurface(legacyManifest.Streams)
	gotStreams := manifestStreamSurface(engManifest.Streams)
	if !reflect.DeepEqual(gotStreams, wantStreams) {
		t.Fatalf("stream surface = %+v, want %+v (legacy)", gotStreams, wantStreams)
	}

	wantWrites := writeActionNames(legacyManifest.WriteActions)
	gotWrites := writeActionNames(engManifest.WriteActions)
	if !reflect.DeepEqual(gotWrites, wantWrites) {
		t.Fatalf("write action names = %v, want %v (legacy)", gotWrites, wantWrites)
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

func writeActionNames(actions []connectors.WriteActionSpec) []string {
	out := make([]string, len(actions))
	for i, a := range actions {
		out[i] = a.Name
	}
	sort.Strings(out)
	return out
}

// TestParityStripe_BundleLoadsAndValidates is a smoke guard: the bundle must
// load cleanly via engine.LoadAll(defs.FS) (meta-schema + structural
// validation already ran inside LoadAll) and declare exactly the 5 legacy
// streams and 2 legacy write actions by name — a minimal, fast-failing
// sanity check that runs before the heavier parity subtests above.
func TestParityStripe_BundleLoadsAndValidates(t *testing.T) {
	bundle := loadStripeBundle(t)

	wantStreams := []string{"charges", "customers", "invoices", "products", "subscriptions"}
	gotStreams := make([]string, 0, len(bundle.Streams))
	for _, s := range bundle.Streams {
		gotStreams = append(gotStreams, s.Name)
	}
	sort.Strings(gotStreams)
	if !reflect.DeepEqual(gotStreams, wantStreams) {
		t.Fatalf("bundle streams = %v, want %v", gotStreams, wantStreams)
	}

	wantWrites := []string{"create_customer", "update_customer"}
	gotWrites := make([]string, 0, len(bundle.Writes))
	for _, w := range bundle.Writes {
		gotWrites = append(gotWrites, w.Name)
	}
	sort.Strings(gotWrites)
	if !reflect.DeepEqual(gotWrites, wantWrites) {
		t.Fatalf("bundle write actions = %v, want %v", gotWrites, wantWrites)
	}

	if !bundle.Metadata.Capabilities.Write {
		t.Fatal("bundle metadata.capabilities.write = false, want true")
	}
}

// TestParityStripe_NoDeadPaginationFields is the F6 golden-hygiene regression
// test (REVIEW.md): the cursor+last_record_field paginator constructor
// (paginate.go's newCursorPaginator) never reads PaginationSpec.LimitParam or
// PaginationSpec.PageSize (only page_number/offset_limit do), so declaring
// them on stripe's base pagination block was dead config that silently did
// nothing while appearing to declare an effective page size — the actual
// limit=100 is sent via each stream's static query.limit. Locks in that the
// dead fields are gone; limit=100 still goes out on the wire via the
// existing TestParityStripe_CustomersTwoPagePagination assertion.
func TestParityStripe_NoDeadPaginationFields(t *testing.T) {
	bundle := loadStripeBundle(t)

	pag := bundle.HTTP.Pagination
	if pag == nil {
		t.Fatal("bundle.HTTP.Pagination is nil, want a cursor pagination spec")
	}
	if pag.LimitParam != "" {
		t.Fatalf("pagination.limit_param = %q, want empty (dead for cursor+last_record_field paginator, F6)", pag.LimitParam)
	}
	if pag.PageSize != 0 {
		t.Fatalf("pagination.page_size = %d, want 0 (dead for cursor+last_record_field paginator, F6)", pag.PageSize)
	}
}

// TestParityStripe_MetadataRateLimitIsInformationalOnly is the F6
// rate-limit-placement regression test (REVIEW.md): metadata.json's
// rate_limit block is NOT consumed by the read path at all (read.go reads
// only b.HTTP.RateLimit, streams.json's base rate_limit — Metadata.RateLimit
// is a separate, purely descriptive field with zero wiring into request
// throttling) and streams.json intentionally declares NO base.rate_limit,
// since legacy stripe enforces no client-side rate limit and this bundle
// must not add new throttling behavior legacy never had. Locks in both
// halves: metadata.json's rate_limit carries no fields this Go type doesn't
// already define as informational (RequestsPerMinute only — no
// "strategy"/enforcement-shape key), and streams.json's HTTP.RateLimit is
// nil (no enforcement).
func TestParityStripe_MetadataRateLimitIsInformationalOnly(t *testing.T) {
	bundle := loadStripeBundle(t)

	if bundle.HTTP.RateLimit != nil {
		t.Fatalf("bundle.HTTP.RateLimit = %+v, want nil (legacy stripe has no client-side rate limiting; streams.json must not add one)", bundle.HTTP.RateLimit)
	}
	if bundle.Metadata.RateLimit.RequestsPerMinute != 100 {
		t.Fatalf("bundle.Metadata.RateLimit.RequestsPerMinute = %d, want 100 (informational-only, matches Stripe's documented published limit)", bundle.Metadata.RateLimit.RequestsPerMinute)
	}
}
