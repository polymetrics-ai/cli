// Package slack: hooks_test.go parity-tests both hook interfaces this
// bundle registers (CheckHook + StreamHook) against legacy
// internal/connectors/slack/slack_test.go's scenarios: cursor pagination,
// per-stream record mapping, the ok:false-at-HTTP-200 error signal, and
// channel_messages' channel_id requirement.
package slack

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/engine"
)

func newRuntime(baseURL string) *engine.Runtime {
	return &engine.Runtime{Requester: &connsdk.Requester{BaseURL: baseURL, Auth: connsdk.Bearer("xoxb-test-123")}}
}

// --- registration -----------------------------------------------------

func TestHooksRegisteredUnderSlack(t *testing.T) {
	h := engine.HooksFor("slack")
	if h == nil {
		t.Fatal("engine.HooksFor(\"slack\") = nil, want registered hooks")
	}
	if h.ConnectorName() != "slack" {
		t.Fatalf("ConnectorName() = %q, want %q", h.ConnectorName(), "slack")
	}
	if _, ok := h.(engine.CheckHook); !ok {
		t.Fatal("registered slack hooks does not implement engine.CheckHook")
	}
	if _, ok := h.(engine.StreamHook); !ok {
		t.Fatal("registered slack hooks does not implement engine.StreamHook")
	}
	if _, ok := h.(engine.AuthHook); ok {
		t.Fatal("registered slack hooks implements engine.AuthHook, want none (auth stays declarative bearer)")
	}
}

// --- StreamHook: cursor pagination + record mapping (legacy parity) -------

func TestReadStream_UsersPaginatesAndMapsRecords(t *testing.T) {
	var sawAuth string
	var sawCursors []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/users.list" {
			http.NotFound(w, r)
			return
		}
		cursor := r.URL.Query().Get("cursor")
		sawCursors = append(sawCursors, cursor)
		w.Header().Set("Content-Type", "application/json")
		switch cursor {
		case "":
			_, _ = w.Write([]byte(`{"ok":true,"members":[` +
				`{"id":"U1","team_id":"T1","name":"ada","real_name":"Ada","deleted":false,"is_bot":false,"updated":1700000000,"profile":{"email":"ada@example.com","display_name":"ada"}},` +
				`{"id":"U2","team_id":"T1","name":"grace","real_name":"Grace","deleted":false,"is_bot":false,"updated":1700000100,"profile":{"email":"grace@example.com","display_name":"grace"}}` +
				`],"response_metadata":{"next_cursor":"PAGE2"}}`))
		case "PAGE2":
			_, _ = w.Write([]byte(`{"ok":true,"members":[` +
				`{"id":"U3","team_id":"T1","name":"katherine","real_name":"Katherine","deleted":false,"is_bot":false,"updated":1700000200,"profile":{"email":"k@example.com","display_name":"katherine"}}` +
				`],"response_metadata":{"next_cursor":""}}`))
		default:
			t.Errorf("unexpected cursor=%q", cursor)
		}
	}))
	defer srv.Close()

	h := Hooks{}
	req := connectors.ReadRequest{Stream: "users", Config: connectors.RuntimeConfig{Config: map[string]string{}}}
	var got []connectors.Record
	handled, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "users"}, req, newRuntime(srv.URL), func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if !handled {
		t.Fatal("handled = false, want true")
	}
	if sawAuth != "Bearer xoxb-test-123" {
		t.Fatalf("Authorization = %q, want Bearer xoxb-test-123", sawAuth)
	}
	if len(sawCursors) != 2 || sawCursors[0] != "" || sawCursors[1] != "PAGE2" {
		t.Fatalf("cursors = %v, want [\"\" \"PAGE2\"]", sawCursors)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (across 2 pages)", len(got))
	}
	if got[0]["id"] != "U1" || got[0]["email"] != "ada@example.com" {
		t.Fatalf("record[0] mapping wrong: %+v", got[0])
	}
	if got[2]["id"] != "U3" {
		t.Fatalf("record[2] id = %v, want U3", got[2]["id"])
	}
}

func TestReadStream_Channels(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/conversations.list" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true,"channels":[` +
			`{"id":"C1","name":"general","is_channel":true,"is_archived":false,"created":1700000000,"num_members":5,"is_private":false}` +
			`],"response_metadata":{"next_cursor":""}}`))
	}))
	defer srv.Close()

	h := Hooks{}
	req := connectors.ReadRequest{Stream: "channels", Config: connectors.RuntimeConfig{Config: map[string]string{}}}
	var got []connectors.Record
	_, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "channels"}, req, newRuntime(srv.URL), func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if len(got) != 1 || got[0]["id"] != "C1" || got[0]["name"] != "general" {
		t.Fatalf("channels mapping wrong: %+v", got)
	}
}

func TestReadStream_ChannelMessagesRequiresChannelID(t *testing.T) {
	h := Hooks{}
	req := connectors.ReadRequest{Stream: "channel_messages", Config: connectors.RuntimeConfig{Config: map[string]string{}}}
	handled, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "channel_messages"}, req, newRuntime("http://example.invalid"), func(connectors.Record) error { return nil })
	if !handled {
		t.Fatal("handled = false, want true (a known stream with bad config is still a handled error, not a declarative fallback)")
	}
	if err == nil {
		t.Fatal("ReadStream error = nil, want an error for a missing channel_id")
	}
}

func TestReadStream_ChannelMessagesSendsChannelParam(t *testing.T) {
	var sawChannel string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawChannel = r.URL.Query().Get("channel")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true,"messages":[{"ts":"1700000000.000100","type":"message","user":"U1","text":"hi"}],"response_metadata":{"next_cursor":""}}`))
	}))
	defer srv.Close()

	h := Hooks{}
	req := connectors.ReadRequest{Stream: "channel_messages", Config: connectors.RuntimeConfig{Config: map[string]string{"channel_id": "C123"}}}
	_, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "channel_messages"}, req, newRuntime(srv.URL), func(connectors.Record) error { return nil })
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if sawChannel != "C123" {
		t.Fatalf("channel param = %q, want C123", sawChannel)
	}
}

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

func TestReadStream_EmptyStreamNameDefaultsToUsers(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/users.list" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"ok":true,"members":[],"response_metadata":{"next_cursor":""}}`))
	}))
	defer srv.Close()

	h := Hooks{}
	req := connectors.ReadRequest{Config: connectors.RuntimeConfig{Config: map[string]string{}}}
	handled, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: ""}, req, newRuntime(srv.URL), func(connectors.Record) error { return nil })
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if !handled {
		t.Fatal("handled = false, want true (empty stream name should default to users)")
	}
}

func TestReadStream_MaxPagesCapsRequestCount(t *testing.T) {
	hits := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		_, _ = w.Write([]byte(`{"ok":true,"members":[{"id":"U1"},{"id":"U2"}],"response_metadata":{"next_cursor":"MORE"}}`))
	}))
	defer srv.Close()

	h := Hooks{}
	req := connectors.ReadRequest{Stream: "users", Config: connectors.RuntimeConfig{Config: map[string]string{"page_size": "2", "max_pages": "1"}}}
	_, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "users"}, req, newRuntime(srv.URL), func(connectors.Record) error { return nil })
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if hits != 1 {
		t.Fatalf("hits = %d, want exactly 1 (max_pages=1 cap), even though next_cursor kept advancing", hits)
	}
}

// --- ok:false at HTTP 200 (the connector's core blocker) -------------------

func TestReadStream_SlackAPIErrorSurfacesEvenAtHTTP200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":false,"error":"invalid_auth"}`))
	}))
	defer srv.Close()

	h := Hooks{}
	req := connectors.ReadRequest{Stream: "users", Config: connectors.RuntimeConfig{Config: map[string]string{}}}
	_, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "users"}, req, newRuntime(srv.URL), func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("expected error for ok:false response even though HTTP status was 200")
	}
	if got := err.Error(); !strings.Contains(got, "invalid_auth") {
		t.Fatalf("error = %q, want it to name the slack error code invalid_auth", got)
	}
}

// --- CheckHook --------------------------------------------------------------

func TestCheck_Passes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/auth.test" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"ok":true,"team":"T1","user":"bot"}`))
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

func TestCheck_FailsOnOKFalseEvenAtHTTP200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":false,"error":"invalid_auth"}`))
	}))
	defer srv.Close()

	h := Hooks{}
	handled, err := h.Check(context.Background(), connectors.RuntimeConfig{}, newRuntime(srv.URL))
	if !handled {
		t.Fatal("handled = false, want true")
	}
	if err == nil {
		t.Fatal("Check error = nil, want an error for ok:false even at HTTP 200")
	}
}
