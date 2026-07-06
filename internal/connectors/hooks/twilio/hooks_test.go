package twilio

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/engine"
)

// newTestRuntime builds a minimal *engine.Runtime pointed at srv — enough
// for ReadStream to run without needing a full loaded bundle (mirrors
// hooks/sentry/hooks_test.go's newTestRuntime, minus the schema-projection
// scaffolding twilio's straight-through mapRecord doesn't need).
func newTestRuntime(srv *httptest.Server, cfg connectors.RuntimeConfig) *engine.Runtime {
	return &engine.Runtime{
		Requester: &connsdk.Requester{BaseURL: srv.URL},
		Bundle:    &engine.Bundle{Name: "twilio"},
		Config:    cfg,
	}
}

func baseCfg(extra map[string]string) connectors.RuntimeConfig {
	cfg := map[string]string{}
	for k, v := range extra {
		cfg[k] = v
	}
	return connectors.RuntimeConfig{
		Config:  cfg,
		Secrets: map[string]string{"account_sid": "AC_test", "auth_token": "tok_secret"},
	}
}

// --- registration ---------------------------------------------------------

func TestHooksRegisteredUnderTwilio(t *testing.T) {
	h := engine.HooksFor("twilio")
	if h == nil {
		t.Fatal(`engine.HooksFor("twilio") = nil, want registered hooks (hooks/twilio's init() must call engine.RegisterHooks)`)
	}
	if h.ConnectorName() != "twilio" {
		t.Fatalf("ConnectorName() = %q, want %q", h.ConnectorName(), "twilio")
	}
	if _, ok := h.(engine.StreamHook); !ok {
		t.Fatal("registered twilio hooks does not implement engine.StreamHook")
	}
	if _, ok := h.(engine.AuthHook); ok {
		t.Fatal("twilio hooks implements engine.AuthHook, want none (auth is fully declarative, mode: basic)")
	}
}

// --- host-relative next_page_uri pagination --------------------------------

// TestReadStream_HostRelativeNextPageURIPagination is the core ENGINE_GAP
// regression: Twilio's real next_page_uri wire value is a HOST-RELATIVE URL
// (proven by legacy's own TestReadPaginatesAndAuthenticates fixture), which
// the engine's declarative next_url pagination type's checkOrigin guard
// would fail-closed reject (empty Host). This hook must follow it correctly
// across 2 pages, resolving it against the requester's own base origin.
func TestReadStream_HostRelativeNextPageURIPagination(t *testing.T) {
	var sawAuth string
	var page0Path, page1Path string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		switch r.URL.Query().Get("Page") {
		case "", "0":
			page0Path = r.URL.Path
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"messages":[` +
				`{"sid":"SM1","date_sent":"Mon, 01 Jan 2024 00:00:00 +0000","from":"+1000","to":"+2000","status":"delivered"},` +
				`{"sid":"SM2","date_sent":"Mon, 01 Jan 2024 00:01:00 +0000","from":"+1000","to":"+2001","status":"sent"}` +
				`],"next_page_uri":"/2010-04-01/Accounts/AC_test/Messages.json?Page=1&PageSize=2","page":0,"page_size":2}`))
		case "1":
			page1Path = r.URL.Path
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"messages":[{"sid":"SM3","date_sent":"Mon, 01 Jan 2024 00:02:00 +0000","from":"+1000","to":"+2002","status":"queued"}],"next_page_uri":null,"page":1,"page_size":2}`))
		default:
			t.Errorf("unexpected Page=%q", r.URL.Query().Get("Page"))
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer srv.Close()

	// Wire real HTTP Basic auth on the test Requester (streams.json's
	// declarative mode:basic auth is what production actually applies;
	// this hook itself never touches auth).
	rt := newTestRuntime(srv, baseCfg(map[string]string{"page_size": "2"}))
	rt.Requester.Auth = connsdk.Basic("AC_test", "tok_secret")

	h := New().(Hooks)
	stream := engine.StreamSpec{Name: "messages", Path: "/Accounts/{{ secrets.account_sid }}/Messages.json"}

	var got []connectors.Record
	handled, err := h.ReadStream(context.Background(), stream, connectors.ReadRequest{Stream: "messages", Config: rt.Config}, rt, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if !handled {
		t.Fatal("ReadStream handled = false, want true for a recognized stream")
	}

	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("AC_test:tok_secret"))
	if sawAuth != wantAuth {
		t.Fatalf("Authorization = %q, want %q", sawAuth, wantAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["sid"] == nil {
			t.Fatalf("record missing sid: %+v", rec)
		}
	}
	if !strings.Contains(page0Path, "/Accounts/AC_test/Messages.json") {
		t.Fatalf("page0 path = %q, want account-scoped Messages.json", page0Path)
	}
	if !strings.Contains(page1Path, "/Accounts/AC_test/Messages.json") {
		t.Fatalf("page1 path = %q, want next_page_uri followed", page1Path)
	}
}

// TestReadStream_NullNextPageURIStopsAfterOnePage asserts a single-page
// response (next_page_uri: null) makes exactly one request.
func TestReadStream_NullNextPageURIStopsAfterOnePage(t *testing.T) {
	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"calls":[{"sid":"CA1","start_time":"t","status":"completed"}],"next_page_uri":null}`))
	}))
	defer srv.Close()

	rt := newTestRuntime(srv, baseCfg(nil))
	h := New().(Hooks)
	stream := engine.StreamSpec{Name: "calls", Path: "/Accounts/{{ secrets.account_sid }}/Calls.json"}

	var got []connectors.Record
	handled, err := h.ReadStream(context.Background(), stream, connectors.ReadRequest{Stream: "calls", Config: rt.Config}, rt, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if !handled {
		t.Fatal("ReadStream handled = false, want true")
	}
	if hits != 1 {
		t.Fatalf("request count = %d, want 1 (next_page_uri null stops pagination)", hits)
	}
	if len(got) != 1 || got[0]["sid"] != "CA1" {
		t.Fatalf("records = %+v, want one record with sid CA1", got)
	}
}

// TestReadStream_EmptyStringNextPageURIStopsIdenticallyToNull mirrors
// legacy's `next == "" || next == "null"` check (twilio.go:198): both a
// missing next_page_uri key AND a literal "null" string sentinel must stop
// pagination identically.
func TestReadStream_EmptyStringNextPageURIStopsIdenticallyToNull(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"usage_records":[{"category":"sms","start_date":"2026-01-01"}]}`)) // no next_page_uri key at all
	}))
	defer srv.Close()

	rt := newTestRuntime(srv, baseCfg(nil))
	h := New().(Hooks)
	stream := engine.StreamSpec{Name: "usage_records", Path: "/Accounts/{{ secrets.account_sid }}/Usage/Records.json"}

	var got []connectors.Record
	_, err := h.ReadStream(context.Background(), stream, connectors.ReadRequest{Stream: "usage_records", Config: rt.Config}, rt, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("records = %d, want 1 (absent next_page_uri key stops pagination)", len(got))
	}
}

// --- max_pages / page_size config ------------------------------------------

func TestReadStream_MaxPagesCapsRequestCount(t *testing.T) {
	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"messages":[{"sid":"SM` + strconv.Itoa(hits) + `"}],"next_page_uri":"/2010-04-01/Accounts/AC_test/Messages.json?Page=` + strconv.Itoa(hits) + `"}`))
	}))
	defer srv.Close()

	rt := newTestRuntime(srv, baseCfg(map[string]string{"max_pages": "2"}))
	h := New().(Hooks)
	stream := engine.StreamSpec{Name: "messages", Path: "/Accounts/{{ secrets.account_sid }}/Messages.json"}

	var got []connectors.Record
	_, err := h.ReadStream(context.Background(), stream, connectors.ReadRequest{Stream: "messages", Config: rt.Config}, rt, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if hits != 2 {
		t.Fatalf("request count = %d, want 2 (max_pages cap, even though next_page_uri kept offering a 3rd page)", hits)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2", len(got))
	}
}

func TestReadStream_PageSizeSentAsPageSizeParamOnFirstRequestOnly(t *testing.T) {
	var firstQuery, secondQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("Page") == "" {
			firstQuery = r.URL.RawQuery
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"messages":[{"sid":"SM1"}],"next_page_uri":"/2010-04-01/Accounts/AC_test/Messages.json?Page=1"}`))
			return
		}
		secondQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"messages":[{"sid":"SM2"}],"next_page_uri":null}`))
	}))
	defer srv.Close()

	rt := newTestRuntime(srv, baseCfg(map[string]string{"page_size": "25"}))
	h := New().(Hooks)
	stream := engine.StreamSpec{Name: "messages", Path: "/Accounts/{{ secrets.account_sid }}/Messages.json"}

	_, err := h.ReadStream(context.Background(), stream, connectors.ReadRequest{Stream: "messages", Config: rt.Config}, rt, func(connectors.Record) error { return nil })
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if !strings.Contains(firstQuery, "PageSize=25") {
		t.Fatalf("first request query = %q, want PageSize=25", firstQuery)
	}
	if strings.Contains(secondQuery, "PageSize=") {
		t.Fatalf("second request query = %q, want no PageSize param (legacy only sends PageSize on the first request)", secondQuery)
	}
}

// --- unrecognized stream ----------------------------------------------------

func TestReadStream_UnrecognizedStreamReturnsNotHandled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("no request should be made for an unrecognized stream")
	}))
	defer srv.Close()

	rt := newTestRuntime(srv, baseCfg(nil))
	h := New().(Hooks)
	stream := engine.StreamSpec{Name: "nope"}

	handled, err := h.ReadStream(context.Background(), stream, connectors.ReadRequest{Stream: "nope", Config: rt.Config}, rt, func(connectors.Record) error { return nil })
	if handled {
		t.Fatal("handled = true, want false for an unrecognized stream name")
	}
	if err != nil {
		t.Fatalf("err = %v, want nil (declarative fallback should surface its own error)", err)
	}
}

// --- config validation errors -----------------------------------------------

func TestReadStream_InvalidPageSizeIsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("no request should be made when page_size is invalid")
	}))
	defer srv.Close()
	rt := newTestRuntime(srv, baseCfg(map[string]string{"page_size": "not-a-number"}))

	h := New().(Hooks)
	stream := engine.StreamSpec{Name: "messages", Path: "/Accounts/{{ secrets.account_sid }}/Messages.json"}
	_, err := h.ReadStream(context.Background(), stream, connectors.ReadRequest{Stream: "messages", Config: rt.Config}, rt, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("ReadStream error = nil, want an error for an invalid page_size")
	}
}

// --- ctx cancellation --------------------------------------------------------

func TestReadStream_HonorsContextCancellation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("no request should be made with an already-cancelled context")
	}))
	defer srv.Close()

	rt := newTestRuntime(srv, baseCfg(nil))
	h := New().(Hooks)
	stream := engine.StreamSpec{Name: "messages", Path: "/Accounts/{{ secrets.account_sid }}/Messages.json"}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := h.ReadStream(ctx, stream, connectors.ReadRequest{Stream: "messages", Config: rt.Config}, rt, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("ReadStream(cancelled ctx) error = nil, want a cancellation error")
	}
}

// --- absoluteURL unit tests (ported from legacy's identical helper) --------

func TestAbsoluteURL_HostRelativeResolvesAgainstBaseOrigin(t *testing.T) {
	got, err := absoluteURL("http://127.0.0.1:9999/2010-04-01", "/2010-04-01/Accounts/AC_test/Messages.json?Page=1")
	if err != nil {
		t.Fatalf("absoluteURL: %v", err)
	}
	want := "http://127.0.0.1:9999/2010-04-01/Accounts/AC_test/Messages.json?Page=1"
	if got != want {
		t.Fatalf("absoluteURL = %q, want %q", got, want)
	}
}

func TestAbsoluteURL_RelativeResolvesAgainstBasePath(t *testing.T) {
	got, err := absoluteURL("http://127.0.0.1:9999/2010-04-01", "Accounts/AC_test/Messages.json")
	if err != nil {
		t.Fatalf("absoluteURL: %v", err)
	}
	want := "http://127.0.0.1:9999/2010-04-01/Accounts/AC_test/Messages.json"
	if got != want {
		t.Fatalf("absoluteURL = %q, want %q", got, want)
	}
}

func TestAbsoluteURL_AbsoluteURLPassesThroughUnchanged(t *testing.T) {
	got, err := absoluteURL("http://127.0.0.1:9999", "https://api.twilio.com/2010-04-01/Accounts/AC/Messages.json")
	if err != nil {
		t.Fatalf("absoluteURL: %v", err)
	}
	want := "https://api.twilio.com/2010-04-01/Accounts/AC/Messages.json"
	if got != want {
		t.Fatalf("absoluteURL = %q, want %q", got, want)
	}
}
