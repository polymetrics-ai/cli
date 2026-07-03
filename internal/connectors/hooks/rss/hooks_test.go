package rss

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/engine"
)

const fixtureFeed = `<?xml version="1.0"?><rss version="2.0"><channel><title>Fixture Feed</title><link>https://example.test</link><description>A fixture RSS channel</description><lastBuildDate>Mon, 02 Jan 2006 15:04:05 MST</lastBuildDate><item><guid>one</guid><title>First item</title><link>https://example.test/one</link><description>First item description</description><pubDate>Mon, 02 Jan 2006 15:04:05 MST</pubDate></item><item><guid>two</guid><title>Second item</title><link>https://example.test/two</link></item></channel></rss>`

func newRuntime(baseURL string) *engine.Runtime {
	return &engine.Runtime{Requester: &connsdk.Requester{BaseURL: baseURL}}
}

// --- registration ---

func TestInit_RegistersHooks(t *testing.T) {
	h := engine.HooksFor("rss")
	if h == nil {
		t.Fatal(`engine.HooksFor("rss") = nil, want a registered hook set (init() must call engine.RegisterHooks)`)
	}
	if h.ConnectorName() != "rss" {
		t.Fatalf("ConnectorName() = %q, want rss", h.ConnectorName())
	}
	if _, ok := h.(engine.StreamHook); !ok {
		t.Fatal("registered hooks do not implement StreamHook")
	}
	if _, ok := h.(engine.CheckHook); !ok {
		t.Fatal("registered hooks do not implement CheckHook")
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

func TestReadStream_EmptyStreamNameDefaultsToItems(t *testing.T) {
	var sawAuth bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" {
			sawAuth = true
		}
		w.Header().Set("Content-Type", "application/rss+xml")
		_, _ = w.Write([]byte(fixtureFeed))
	}))
	defer srv.Close()

	h := Hooks{}
	var got []connectors.Record
	handled, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: ""}, connectors.ReadRequest{}, newRuntime(srv.URL), func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if !handled {
		t.Fatal("handled = false, want true for empty stream name (defaults to items)")
	}
	if sawAuth {
		t.Fatal("rss hook sent credentials")
	}
	if len(got) != 2 || got[0]["id"] != "one" || got[0]["title"] != "First item" {
		t.Fatalf("items not parsed: %+v", got)
	}
	if got[1]["id"] != "two" {
		t.Fatalf("second item id fallback (link) failed: %+v", got[1])
	}
}

func TestReadStream_Items(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		_, _ = w.Write([]byte(fixtureFeed))
	}))
	defer srv.Close()

	h := Hooks{}
	var got []connectors.Record
	handled, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "items"}, connectors.ReadRequest{}, newRuntime(srv.URL), func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if !handled {
		t.Fatal("handled = false, want true")
	}
	if len(got) != 2 {
		t.Fatalf("want 2 items, got %d: %+v", len(got), got)
	}
	if got[0]["published_at"] != "Mon, 02 Jan 2006 15:04:05 MST" {
		t.Fatalf("published_at not verbatim: %+v", got[0])
	}
}

func TestReadStream_Channel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		_, _ = w.Write([]byte(fixtureFeed))
	}))
	defer srv.Close()

	h := Hooks{}
	var got []connectors.Record
	handled, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "channel"}, connectors.ReadRequest{}, newRuntime(srv.URL), func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if !handled {
		t.Fatal("handled = false, want true")
	}
	if len(got) != 1 || got[0]["title"] != "Fixture Feed" || got[0]["id"] != "https://example.test" {
		t.Fatalf("channel not parsed: %+v", got)
	}
}

func TestReadStream_MissingChannelDataErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		_, _ = w.Write([]byte(`<?xml version="1.0"?><rss version="2.0"><channel></channel></rss>`))
	}))
	defer srv.Close()

	h := Hooks{}
	_, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "items"}, connectors.ReadRequest{}, newRuntime(srv.URL), func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("ReadStream: want error for a feed with no title and no items")
	}
}

// --- Check ---

func TestCheck_DecodesFeed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		_, _ = w.Write([]byte(fixtureFeed))
	}))
	defer srv.Close()

	h := Hooks{}
	handled, err := h.Check(context.Background(), connectors.RuntimeConfig{}, newRuntime(srv.URL))
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if !handled {
		t.Fatal("handled = false, want true")
	}
}

func TestCheck_PropagatesLoadError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	h := Hooks{}
	handled, err := h.Check(context.Background(), connectors.RuntimeConfig{}, newRuntime(srv.URL))
	if err == nil {
		t.Fatal("Check: want error on 500 response")
	}
	if !handled {
		t.Fatal("handled = false, want true even on error (CheckHook always claims the check)")
	}
}

// --- record mapping ---

func TestItemRecord_IDFallbackChain(t *testing.T) {
	if got := itemRecord(rssItem{GUID: "g", Link: "l", Title: "t"}); got["id"] != "g" {
		t.Fatalf("id = %v, want guid to win", got["id"])
	}
	if got := itemRecord(rssItem{Link: "l", Title: "t"}); got["id"] != "l" {
		t.Fatalf("id = %v, want link fallback", got["id"])
	}
	if got := itemRecord(rssItem{Title: "t"}); got["id"] != "t" {
		t.Fatalf("id = %v, want title fallback", got["id"])
	}
}

func TestChannelRecord_IDFallbackChain(t *testing.T) {
	if got := channelRecord(rssChannel{Link: "l", Title: "t"}); got["id"] != "l" {
		t.Fatalf("id = %v, want link to win", got["id"])
	}
	if got := channelRecord(rssChannel{Title: "t"}); got["id"] != "t" {
		t.Fatalf("id = %v, want title fallback", got["id"])
	}
}

func TestLoad_WrapsRequestError(t *testing.T) {
	r := &connsdk.Requester{BaseURL: "http://127.0.0.1:0"}
	_, err := load(context.Background(), r)
	if err == nil {
		t.Fatal("load: want error for an unreachable server")
	}
}
