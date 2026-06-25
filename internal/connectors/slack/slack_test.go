package slack_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics/internal/connectors"
	"polymetrics/internal/connectors/slack"
)

// TestReadPaginatesAndAuthenticates is the red-first test: Bearer auth, Slack
// cursor pagination (response_metadata.next_cursor -> cursor param) across two
// pages, records extracted from the per-stream list key (members), and field
// mapping. Red until internal/connectors/slack exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
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
				`{"id":"U1","team_id":"T1","name":"ada","real_name":"Ada","deleted":false,"is_bot":false,"updated":1700000000,"profile":{"email":"ada@example.com","real_name":"Ada Lovelace","display_name":"ada"}},` +
				`{"id":"U2","team_id":"T1","name":"grace","real_name":"Grace","deleted":false,"is_bot":false,"updated":1700000100,"profile":{"email":"grace@example.com","real_name":"Grace Hopper","display_name":"grace"}}` +
				`],"response_metadata":{"next_cursor":"PAGE2"}}`))
		case "PAGE2":
			_, _ = w.Write([]byte(`{"ok":true,"members":[` +
				`{"id":"U3","team_id":"T1","name":"katherine","real_name":"Katherine","deleted":false,"is_bot":false,"updated":1700000200,"profile":{"email":"k@example.com","real_name":"Katherine Johnson","display_name":"katherine"}}` +
				`],"response_metadata":{"next_cursor":""}}`))
		default:
			t.Errorf("unexpected cursor=%q", cursor)
			_, _ = w.Write([]byte(`{"ok":true,"members":[],"response_metadata":{"next_cursor":""}}`))
		}
	}))
	defer srv.Close()

	c := slack.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_token": "xoxb-test-123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "users", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
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

// TestReadChannels confirms the channels stream maps conversations.list ->
// channels[] with the right list key.
func TestReadChannels(t *testing.T) {
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

	c := slack.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_token": "xoxb-test-123"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "channels", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 1 || got[0]["id"] != "C1" || got[0]["name"] != "general" {
		t.Fatalf("channels mapping wrong: %+v", got)
	}
}

// TestSlackAPIError verifies a Slack {"ok":false,"error":...} body is surfaced
// as an error even though the HTTP status is 200.
func TestSlackAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":false,"error":"invalid_auth"}`))
	}))
	defer srv.Close()

	c := slack.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_token": "xoxb-bad"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "users", Config: cfg}, func(rec connectors.Record) error {
		return nil
	})
	if err == nil {
		t.Fatal("expected error for ok:false response, got nil")
	}
}

// TestFixtureMode confirms a credential-free, network-free deterministic read.
func TestFixtureMode(t *testing.T) {
	c := slack.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "users", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) == 0 {
		t.Fatal("fixture mode produced no records")
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("fixture record missing id: %+v", rec)
		}
	}
	// Check must succeed in fixture mode with no secret.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

// TestCatalogAndMetadata verifies the published streams and read-only caps.
func TestCatalogAndMetadata(t *testing.T) {
	c := slack.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("slack is read-only; Write must be false, got %+v", caps)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"users": false, "channels": false, "channel_messages": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q missing primary key", s.Name)
		}
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("catalog missing required stream %q", name)
		}
	}
}

// TestRegistryResolution confirms self-registration via init().
func TestRegistryResolution(t *testing.T) {
	_ = slack.New() // ensure package init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("slack"); !ok {
		t.Fatal("registry did not resolve slack (self-registration)")
	}
}

// TestBaseURLValidation rejects SSRF-unsafe overrides.
func TestBaseURLValidation(t *testing.T) {
	c := slack.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil"},
		Secrets: map[string]string{"api_token": "xoxb-x"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "users", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("expected error for non-http(s) base_url scheme")
	}
}
