package appsflyer

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/engine"
)

func newRuntime(baseURL string) *engine.Runtime {
	return &engine.Runtime{Requester: &connsdk.Requester{BaseURL: baseURL, Auth: connsdk.Bearer("test-token")}}
}

// --- registration ---

func TestInit_RegistersHooks(t *testing.T) {
	h := engine.HooksFor("appsflyer")
	if h == nil {
		t.Fatal("engine.HooksFor(\"appsflyer\") = nil, want a registered hook set (init() must call engine.RegisterHooks)")
	}
	if h.ConnectorName() != "appsflyer" {
		t.Fatalf("ConnectorName() = %q, want appsflyer", h.ConnectorName())
	}
	if _, ok := h.(engine.StreamHook); !ok {
		t.Fatal("registered hooks do not implement StreamHook")
	}
}

// --- ReadStream dispatch ---

func TestReadStream_UnknownStreamFallsBackToDeclarative(t *testing.T) {
	h := Hooks{}
	handled, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "not_a_real_stream"}, connectors.ReadRequest{}, newRuntime("http://example.invalid"), func(connectors.Record) error { return nil })
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if handled {
		t.Fatal("handled = true for an unrecognized stream name, want false (declarative fallback)")
	}
}

// TestReadStream_InstallsReportAuthenticatesAndMapsCSV mirrors legacy's own
// TestReadInstallsAuthenticatesAndMapsCSV (appsflyer_test.go): asserts the
// Bearer auth header, the exact report path, the from/to query params
// derived from start_date/end_date, and per-row CSV-to-record mapping with
// header snake-casing.
func TestReadStream_InstallsReportAuthenticatesAndMapsCSV(t *testing.T) {
	var sawAuth string
	var sawPath string
	var sawFrom, sawTo string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		sawPath = r.URL.Path
		sawFrom = r.URL.Query().Get("from")
		sawTo = r.URL.Query().Get("to")
		w.Header().Set("Content-Type", "text/csv")
		_, _ = w.Write([]byte("AppsFlyer ID,Event Time,Media Source,Campaign\naf_1,2026-01-01 00:00:00,network_a,winter\naf_2,2026-01-02 00:00:00,network_b,spring\n"))
	}))
	defer srv.Close()

	h := Hooks{}
	req := connectors.ReadRequest{
		Stream: "installs_report",
		Config: connectors.RuntimeConfig{Config: map[string]string{"app_id": "com.example", "start_date": "2026-01-01", "end_date": "2026-01-02"}},
	}
	var got []connectors.Record
	handled, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "installs_report"}, req, newRuntime(srv.URL), func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if !handled {
		t.Fatal("handled = false, want true (installs_report is always hook-handled)")
	}
	if sawPath != "/api/raw-data/export/app/com.example/installs_report/v5" {
		t.Fatalf("path = %q, want the app_id-scoped installs_report path", sawPath)
	}
	if sawAuth != "Bearer test-token" {
		t.Fatalf("Authorization = %q, want bearer token", sawAuth)
	}
	if sawFrom != "2026-01-01" {
		t.Fatalf("from = %q, want 2026-01-01", sawFrom)
	}
	if sawTo != "2026-01-02" {
		t.Fatalf("to = %q, want 2026-01-02", sawTo)
	}
	if len(got) != 2 || got[0]["appsflyer_id"] != "af_1" || got[1]["campaign"] != "spring" {
		t.Fatalf("csv records not mapped: %+v", got)
	}
}

// TestReadStream_InAppEventsReportUsesOwnPath asserts the second stream's
// report path segment differs from installs_report's.
func TestReadStream_InAppEventsReportUsesOwnPath(t *testing.T) {
	var sawPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawPath = r.URL.Path
		w.Header().Set("Content-Type", "text/csv")
		_, _ = w.Write([]byte("AppsFlyer ID,Event Time,Media Source,Campaign\n"))
	}))
	defer srv.Close()

	h := Hooks{}
	req := connectors.ReadRequest{Stream: "in_app_events_report", Config: connectors.RuntimeConfig{Config: map[string]string{"app_id": "com.example"}}}
	_, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "in_app_events_report"}, req, newRuntime(srv.URL), func(connectors.Record) error { return nil })
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if sawPath != "/api/raw-data/export/app/com.example/in_app_events_report/v5" {
		t.Fatalf("path = %q, want the app_id-scoped in_app_events_report path", sawPath)
	}
}

// TestReadStream_EndDateDefaultsToStartDate mirrors legacy's first(end_date,
// start_date) fallback: an unset end_date sends the start_date value as
// "to" as well.
func TestReadStream_EndDateDefaultsToStartDate(t *testing.T) {
	var sawTo string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawTo = r.URL.Query().Get("to")
		w.Header().Set("Content-Type", "text/csv")
		_, _ = w.Write([]byte("AppsFlyer ID,Event Time,Media Source,Campaign\n"))
	}))
	defer srv.Close()

	h := Hooks{}
	req := connectors.ReadRequest{Stream: "installs_report", Config: connectors.RuntimeConfig{Config: map[string]string{"app_id": "com.example", "start_date": "2026-02-01"}}}
	_, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "installs_report"}, req, newRuntime(srv.URL), func(connectors.Record) error { return nil })
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if sawTo != "2026-02-01" {
		t.Fatalf("to = %q, want 2026-02-01 (defaults to start_date)", sawTo)
	}
}

// TestReadStream_TimezoneSentWhenConfigured asserts the optional timezone
// query param is only sent when configured.
func TestReadStream_TimezoneSentWhenConfigured(t *testing.T) {
	var sawTZ string
	var sawHasTZ bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawTZ = r.URL.Query().Get("timezone")
		sawHasTZ = r.URL.Query().Has("timezone")
		w.Header().Set("Content-Type", "text/csv")
		_, _ = w.Write([]byte("AppsFlyer ID,Event Time,Media Source,Campaign\n"))
	}))
	defer srv.Close()

	h := Hooks{}
	req := connectors.ReadRequest{Stream: "installs_report", Config: connectors.RuntimeConfig{Config: map[string]string{"app_id": "com.example", "timezone": "UTC"}}}
	_, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "installs_report"}, req, newRuntime(srv.URL), func(connectors.Record) error { return nil })
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if !sawHasTZ || sawTZ != "UTC" {
		t.Fatalf("timezone query = (has=%v, val=%q), want UTC", sawHasTZ, sawTZ)
	}
}

// --- CSV decode / header snake-casing ---

func TestEmitCSV_HeaderSnakeCasingAndAppsFlyerCorrection(t *testing.T) {
	var got []connectors.Record
	body := []byte("AppsFlyer ID,Event Time,Media Source,Campaign\naf_1,2026-01-01 00:00:00,network_a,winter\n")
	if err := emitCSV(context.Background(), body, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("emitCSV: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("records = %d, want 1", len(got))
	}
	if _, ok := got[0]["appsflyer_id"]; !ok {
		t.Fatalf("record missing appsflyer_id key (apps_flyer->appsflyer correction failed): %+v", got[0])
	}
	if _, ok := got[0]["apps_flyer_id"]; ok {
		t.Fatalf("record has uncorrected apps_flyer_id key: %+v", got[0])
	}
	if got[0]["event_time"] != "2026-01-01 00:00:00" {
		t.Fatalf("event_time = %v, want the raw timestamp string", got[0]["event_time"])
	}
}

func TestEmitCSV_EmptyBodyEmitsNoRecords(t *testing.T) {
	var n int
	if err := emitCSV(context.Background(), []byte(""), func(connectors.Record) error { n++; return nil }); err != nil {
		t.Fatalf("emitCSV: %v", err)
	}
	if n != 0 {
		t.Fatalf("records = %d, want 0 for an empty body", n)
	}
}

func TestEmitCSV_RaggedRowsTolerated(t *testing.T) {
	// FieldsPerRecord: -1 (mirroring legacy) tolerates a short row; missing
	// trailing columns are simply absent from that record, not an error.
	var got []connectors.Record
	body := []byte("AppsFlyer ID,Event Time,Media Source,Campaign\naf_1,2026-01-01 00:00:00\n")
	if err := emitCSV(context.Background(), body, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("emitCSV: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("records = %d, want 1", len(got))
	}
	if _, ok := got[0]["media_source"]; ok {
		t.Fatalf("record has media_source key despite a short row: %+v", got[0])
	}
}

// --- snake() unit coverage ---

func TestSnake(t *testing.T) {
	cases := map[string]string{
		"AppsFlyer ID": "appsflyer_id",
		"Event Time":   "event_time",
		"Media Source": "media_source",
		"Campaign":     "campaign",
		"  Foo-Bar!! ": "foo_bar",
	}
	for in, want := range cases {
		if got := snake(in); got != want {
			t.Errorf("snake(%q) = %q, want %q", in, got, want)
		}
	}
}

// --- reportPath / reportQuery unit coverage ---

func TestReportPath_EscapesAppID(t *testing.T) {
	cfg := connectors.RuntimeConfig{Config: map[string]string{"app_id": "com/example weird"}}
	got := reportPath(cfg, "installs_report")
	want := "/api/raw-data/export/app/com%2Fexample%20weird/installs_report/v5"
	if got != want {
		t.Fatalf("reportPath = %q, want %q", got, want)
	}
}

func TestReportQuery_NoDatesConfiguredSendsNoParams(t *testing.T) {
	cfg := connectors.RuntimeConfig{Config: map[string]string{"app_id": "com.example"}}
	q := reportQuery(cfg)
	if q.Has("from") || q.Has("to") || q.Has("timezone") {
		t.Fatalf("query = %v, want no from/to/timezone params when unset", q)
	}
}
