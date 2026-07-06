// Package paritytest_twilio is the engine-vs-legacy parity suite for the
// twilio Tier-2 (StreamHook) migration (docs/migration/quarantine.json's
// OTHER/no-reason entry; investigation for this pass found the real
// blocker is host-relative next_page_uri pagination, not auth — see
// defs/twilio/docs.md's Overview). Both the legacy hand-written
// twilio.Connector (internal/connectors/twilio, read-only reference) and
// the engine-backed connector (engine.New(bundle,
// engine.HooksFor("twilio"))) are driven against the SAME httptest server;
// RAW connectors.Record reflect.DeepEqual equality is the parity bar,
// mirroring paritytest/sentry and paritytest/monday's precedent for a
// hook-backed pilot.
package paritytest_twilio

import (
	"context"
	"encoding/base64"
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
	_ "polymetrics.ai/internal/connectors/hooks/twilio" // registers the StreamHook via init()
	"polymetrics.ai/internal/connectors/twilio"
)

const bundleName = "twilio"

func loadBundle(t *testing.T) engine.Bundle {
	t.Helper()
	b, err := engine.Load(defs.FS, bundleName)
	if err != nil {
		t.Fatalf("engine.Load(defs.FS, %q): %v", bundleName, err)
	}
	return b
}

func withBaseURL(b engine.Bundle, baseURL string) engine.Bundle {
	b.HTTP.URL = baseURL
	return b
}

func runtimeConfig(extra map[string]string) connectors.RuntimeConfig {
	cfg := map[string]string{}
	for k, v := range extra {
		cfg[k] = v
	}
	return connectors.RuntimeConfig{
		Config:  cfg,
		Secrets: map[string]string{"account_sid": "AC_test", "auth_token": "tok_secret"},
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

// dataServer builds a 2-page next_page_uri (host-relative) fixture server
// for messages, and single-page servers for the other 4 streams — mirrors
// legacy's own TestReadPaginatesAndAuthenticates fixture shape exactly.
func dataServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	messagesHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Query().Get("Page") {
		case "", "0":
			// next_page_uri carries Twilio's real host-relative convention:
			// it includes the /2010-04-01 API-version prefix (legacy's
			// absoluteURL treats any "/"-prefixed reference as host-relative
			// and takes it verbatim) — registered below under BOTH the
			// unprefixed and /2010-04-01-prefixed paths so page 1's request
			// (built from streams.json's plain stream path) and page 2's
			// request (following next_page_uri) both resolve.
			_, _ = w.Write([]byte(`{"messages":[
				{"sid":"SM1","account_sid":"AC_test","messaging_service_sid":null,"date_created":"Mon, 01 Jan 2024 00:00:00 +0000","date_sent":"Mon, 01 Jan 2024 00:00:00 +0000","date_updated":"Mon, 01 Jan 2024 00:00:00 +0000","from":"+1000","to":"+2000","body":"hi","status":"delivered","direction":"outbound-api","num_segments":"1","num_media":"0","price":"-0.0075","price_unit":"USD","error_code":null,"error_message":null},
				{"sid":"SM2","account_sid":"AC_test","messaging_service_sid":null,"date_created":"Mon, 01 Jan 2024 00:01:00 +0000","date_sent":"Mon, 01 Jan 2024 00:01:00 +0000","date_updated":"Mon, 01 Jan 2024 00:01:00 +0000","from":"+1000","to":"+2001","body":"hi again","status":"sent","direction":"outbound-api","num_segments":"1","num_media":"0","price":"-0.0075","price_unit":"USD","error_code":null,"error_message":null}
			],"next_page_uri":"/2010-04-01/Accounts/AC_test/Messages.json?Page=1&PageSize=2","page":0,"page_size":2}`))
		case "1":
			_, _ = w.Write([]byte(`{"messages":[{"sid":"SM3","account_sid":"AC_test","messaging_service_sid":null,"date_created":"Mon, 01 Jan 2024 00:02:00 +0000","date_sent":"Mon, 01 Jan 2024 00:02:00 +0000","date_updated":"Mon, 01 Jan 2024 00:02:00 +0000","from":"+1000","to":"+2002","body":"third","status":"queued","direction":"outbound-api","num_segments":"1","num_media":"0","price":"-0.0075","price_unit":"USD","error_code":null,"error_message":null}],"next_page_uri":null,"page":1,"page_size":2}`))
		default:
			t.Errorf("unexpected Page=%q for messages", r.URL.Query().Get("Page"))
			_, _ = w.Write([]byte(`{"messages":[],"next_page_uri":null}`))
		}
	}
	mux.HandleFunc("/Accounts/AC_test/Messages.json", messagesHandler)
	mux.HandleFunc("/2010-04-01/Accounts/AC_test/Messages.json", messagesHandler)

	mux.HandleFunc("/Accounts/AC_test/Calls.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"calls":[{"sid":"CA1","account_sid":"AC_test","date_created":"d","date_updated":"d","start_time":"t1","end_time":"t2","from":"+1000","to":"+2000","status":"completed","direction":"outbound-api","duration":"30","price":"-0.02","price_unit":"USD"}],"next_page_uri":null}`))
	})

	mux.HandleFunc("/Accounts/AC_test/Recordings.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"recordings":[{"sid":"RE1","account_sid":"AC_test","call_sid":"CA1","date_created":"d","date_updated":"d","start_time":"t1","duration":"30","status":"completed","channels":1,"source":"DialVerb","price":"-0.0025","price_unit":"USD"}],"next_page_uri":null}`))
	})

	mux.HandleFunc("/Accounts/AC_test/Conferences.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"conferences":[{"sid":"CF1","account_sid":"AC_test","friendly_name":"conf-1","date_created":"d","date_updated":"d","status":"completed","region":"us1"}],"next_page_uri":null}`))
	})

	mux.HandleFunc("/Accounts/AC_test/Usage/Records.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"usage_records":[{"account_sid":"AC_test","category":"sms-outbound","description":"SMS - Outbound","start_date":"2026-01-01","end_date":"2026-01-31","count":"2","count_unit":"messages","usage":"2","usage_unit":"messages","price":"-0.015","price_unit":"USD"}],"next_page_uri":null}`))
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func newLegacyConnector() connectors.Connector { return twilio.New() }

func newEngineConnector(b engine.Bundle, h engine.Hooks) connectors.Connector {
	return engine.New(b, h)
}

// --- per-stream record parity across all 5 streams ---

func TestParityTwilio_StreamRecords(t *testing.T) {
	bundle := loadBundle(t)

	streams := []string{"messages", "calls", "recordings", "conferences", "usage_records"}
	for _, stream := range streams {
		stream := stream
		t.Run(stream, func(t *testing.T) {
			srv := dataServer(t)

			legacy := newLegacyConnector()
			legacyRecs := readAllRecords(t, legacy, connectors.ReadRequest{Stream: stream, Config: withBase(runtimeConfig(nil), srv.URL)})

			eng := newEngineConnector(withBaseURL(bundle, srv.URL), engine.HooksFor(bundleName))
			engRecs := readAllRecords(t, eng, connectors.ReadRequest{Stream: stream, Config: runtimeConfig(nil)})

			if len(legacyRecs) == 0 {
				t.Fatalf("legacy twilio emitted zero records for stream %q (test fixture bug)", stream)
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

// withBase returns a copy of cfg with base_url set (legacy reads base_url
// from config directly; the engine reads it via bundle.HTTP.URL instead, so
// only the legacy side needs this).
func withBase(cfg connectors.RuntimeConfig, baseURL string) connectors.RuntimeConfig {
	out := connectors.RuntimeConfig{Config: map[string]string{}, Secrets: cfg.Secrets}
	for k, v := range cfg.Config {
		out.Config[k] = v
	}
	out.Config["base_url"] = baseURL
	return out
}

// --- pagination parity: messages 2-page host-relative next_page_uri ---

func TestParityTwilio_MessagesTwoPagePagination(t *testing.T) {
	bundle := loadBundle(t)
	srv := dataServer(t)

	legacy := newLegacyConnector()
	legacyRecs := readAllRecords(t, legacy, connectors.ReadRequest{Stream: "messages", Config: withBase(runtimeConfig(map[string]string{"page_size": "2"}), srv.URL)})
	if len(legacyRecs) != 3 {
		t.Fatalf("legacy messages records = %d, want 3 (2 pages)", len(legacyRecs))
	}

	eng := newEngineConnector(withBaseURL(bundle, srv.URL), engine.HooksFor(bundleName))
	engRecs := readAllRecords(t, eng, connectors.ReadRequest{Stream: "messages", Config: runtimeConfig(map[string]string{"page_size": "2"})})
	if len(engRecs) != 3 {
		t.Fatalf("engine messages records = %d, want 3 (2 pages)", len(engRecs))
	}

	gotIDs := recordSIDs(engRecs)
	wantIDs := recordSIDs(legacyRecs)
	if !reflect.DeepEqual(gotIDs, wantIDs) {
		t.Fatalf("messages record sid sequence = %v, want %v", gotIDs, wantIDs)
	}
	if !reflect.DeepEqual(gotIDs, []string{"SM1", "SM2", "SM3"}) {
		t.Fatalf("messages record sid sequence = %v, want [SM1 SM2 SM3]", gotIDs)
	}
}

func recordSIDs(recs []connectors.Record) []string {
	out := make([]string, len(recs))
	for i, r := range recs {
		sid, _ := r["sid"].(string)
		out[i] = sid
	}
	return out
}

// --- auth parity: HTTP Basic header ---

func TestParityTwilio_BasicAuthHeaderParity(t *testing.T) {
	bundle := loadBundle(t)

	var legacyAuth, engAuth string
	legacySrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		legacyAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"calls":[],"next_page_uri":null}`))
	}))
	t.Cleanup(legacySrv.Close)
	engSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		engAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"calls":[],"next_page_uri":null}`))
	}))
	t.Cleanup(engSrv.Close)

	legacy := newLegacyConnector()
	_ = readAllRecords(t, legacy, connectors.ReadRequest{Stream: "calls", Config: withBase(runtimeConfig(nil), legacySrv.URL)})

	eng := newEngineConnector(withBaseURL(bundle, engSrv.URL), engine.HooksFor(bundleName))
	_ = readAllRecords(t, eng, connectors.ReadRequest{Stream: "calls", Config: runtimeConfig(nil)})

	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("AC_test:tok_secret"))
	if legacyAuth != wantAuth {
		t.Fatalf("legacy Authorization = %q, want %q (test fixture bug)", legacyAuth, wantAuth)
	}
	if engAuth != wantAuth {
		t.Fatalf("engine Authorization = %q, want %q", engAuth, wantAuth)
	}
	if legacyAuth != engAuth {
		t.Fatalf("Authorization header mismatch: legacy=%q engine=%q", legacyAuth, engAuth)
	}
}

// --- write parity: legacy remains read-only; bundle declares Pass B writes ---

func TestParityTwilio_WriteUnsupportedOnBothSides(t *testing.T) {
	bundle := loadBundle(t)

	legacy := twilio.New()
	if _, err := legacy.Write(context.Background(), connectors.WriteRequest{}, nil); err == nil {
		t.Fatal("legacy Write succeeded, want ErrUnsupportedOperation (test fixture bug)")
	}

	eng := engine.New(bundle, engine.HooksFor(bundleName))
	if _, err := eng.Write(context.Background(), connectors.WriteRequest{}, nil); err == nil {
		t.Fatal("engine Write succeeded with no action, want an error")
	}
	if !bundle.Metadata.Capabilities.Write {
		t.Fatal("bundle metadata.capabilities.write = false, want true (Pass B write actions are modeled)")
	}
	if len(bundle.Writes) == 0 {
		t.Fatal("bundle write actions = 0, want Pass B write actions")
	}
}

// --- manifest-surface parity ---

func TestParityTwilio_ManifestSurface(t *testing.T) {
	bundle := loadBundle(t)

	legacyCatalog, err := twilio.New().Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("legacy Catalog: %v", err)
	}

	eng := engine.New(bundle, engine.HooksFor(bundleName))
	engManifest := connectors.ManifestOf(eng)

	wantStreams := manifestStreamSurface(legacyCatalog.Streams)
	gotStreams := manifestStreamSurface(engManifest.Streams)
	if missing := missingTwilioStreamSurface(gotStreams, wantStreams); len(missing) != 0 {
		t.Fatalf("engine manifest missing legacy stream surface entries %+v; got %+v", missing, gotStreams)
	}

	if len(engManifest.WriteActions) == 0 {
		t.Fatal("engine write actions = 0, want Pass B write actions")
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

func missingTwilioStreamSurface(got, want []streamSurface) []streamSurface {
	byName := make(map[string]streamSurface, len(got))
	for _, s := range got {
		byName[s.Name] = s
	}
	var missing []streamSurface
	for _, s := range want {
		if gotS, ok := byName[s.Name]; !ok || !reflect.DeepEqual(gotS.PrimaryKey, s.PrimaryKey) {
			missing = append(missing, s)
		}
	}
	return missing
}

func missingStrings(got, want []string) []string {
	seen := make(map[string]bool, len(got))
	for _, s := range got {
		seen[s] = true
	}
	var missing []string
	for _, s := range want {
		if !seen[s] {
			missing = append(missing, s)
		}
	}
	return missing
}

// --- bundle load smoke guard ---

func TestParityTwilio_BundleLoadsAndValidates(t *testing.T) {
	bundle := loadBundle(t)

	wantStreams := []string{"calls", "conferences", "messages", "recordings", "usage_records"}
	gotStreams := make([]string, 0, len(bundle.Streams))
	for _, s := range bundle.Streams {
		gotStreams = append(gotStreams, s.Name)
	}
	sort.Strings(gotStreams)
	if missing := missingStrings(gotStreams, wantStreams); len(missing) != 0 {
		t.Fatalf("bundle streams missing legacy streams %v; got %v", missing, gotStreams)
	}

	if len(bundle.Writes) == 0 {
		t.Fatal("bundle write actions = 0, want Pass B write actions")
	}
	if !bundle.Metadata.Capabilities.Write {
		t.Fatal("bundle metadata.capabilities.write = false, want true")
	}
	for _, s := range bundle.Streams {
		if s.Incremental != nil {
			t.Errorf("stream %q declares an incremental block, want none (legacy never applies a state-based filter)", s.Name)
		}
	}
}

// --- AuthSpec shape guard ---

// TestParityTwilio_AuthSpecIsDeclarativeBasic locks in the decision that
// twilio needs NO AuthHook: legacy's HTTP Basic account_sid/auth_token pair
// is fully declarative-expressible.
func TestParityTwilio_AuthSpecIsDeclarativeBasic(t *testing.T) {
	bundle := loadBundle(t)

	if len(bundle.HTTP.Auth) != 1 {
		t.Fatalf("len(bundle.HTTP.Auth) = %d, want 1", len(bundle.HTTP.Auth))
	}
	spec := bundle.HTTP.Auth[0]
	if spec.Mode != "basic" {
		t.Fatalf("auth spec mode = %q, want %q (no AuthHook needed for twilio)", spec.Mode, "basic")
	}
	if spec.Hook != "" {
		t.Fatalf("auth spec hook = %q, want empty (declarative basic auth, not a custom hook)", spec.Hook)
	}
}

// TestParityTwilio_HooksImplementsStreamHookOnly asserts the registered
// hooks implement StreamHook (pagination) but NOT AuthHook (auth stays
// fully declarative) — the exact Tier-2 shape docs.md's Overview describes.
func TestParityTwilio_HooksImplementsStreamHookOnly(t *testing.T) {
	h := engine.HooksFor(bundleName)
	if h == nil {
		t.Fatal("engine.HooksFor(\"twilio\") = nil, want registered hooks")
	}
	if _, ok := h.(engine.StreamHook); !ok {
		t.Fatal("registered twilio hooks does not implement engine.StreamHook")
	}
	if _, ok := h.(engine.AuthHook); ok {
		t.Fatal("registered twilio hooks implements engine.AuthHook, want none")
	}
}
