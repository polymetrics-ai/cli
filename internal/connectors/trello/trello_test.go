package trello_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/trello"
)

// TestReadCardsAuthenticatesAndPaginates is the red-first test for the Trello
// connector: it asserts that the key/token credentials are sent as query params,
// that a board-scoped stream pages across two requests using the id-cursor
// `before` parameter, and that records are mapped with id/idBoard fields.
func TestReadCardsAuthenticatesAndPaginates(t *testing.T) {
	var sawKey, sawToken string
	var boardCalls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		sawKey = q.Get("key")
		sawToken = q.Get("token")
		switch r.URL.Path {
		case "/members/me/boards":
			boardCalls++
			_, _ = w.Write([]byte(`[{"id":"board_1","name":"Board One","idOrganization":"org_1"}]`))
		case "/boards/board_1/cards":
			switch q.Get("before") {
			case "":
				// First full page (limit is 2). Trello returns newest-first; the
				// last id (card_2) becomes the `before` cursor for the next page.
				_, _ = w.Write([]byte(`[{"id":"card_1","name":"Card One","idBoard":"board_1","dateLastActivity":"2026-01-03T00:00:00.000Z"},{"id":"card_2","name":"Card Two","idBoard":"board_1","dateLastActivity":"2026-01-02T00:00:00.000Z"}]`))
			case "card_2":
				// Second short page signals the end.
				_, _ = w.Write([]byte(`[{"id":"card_0","name":"Card Zero","idBoard":"board_1","dateLastActivity":"2026-01-01T00:00:00.000Z"}]`))
			default:
				t.Errorf("unexpected before=%q", q.Get("before"))
				_, _ = w.Write([]byte(`[]`))
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := trello.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"key": "k_test", "token": "t_test"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "cards", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawKey != "k_test" || sawToken != "t_test" {
		t.Fatalf("auth query = key %q token %q, want k_test/t_test", sawKey, sawToken)
	}
	if boardCalls != 1 {
		t.Fatalf("board discovery calls = %d, want 1", boardCalls)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["idBoard"] == nil {
			t.Fatalf("record missing id/idBoard: %+v", rec)
		}
	}
}

// TestReadBoardsUsesConfiguredBoardIDs verifies the boards stream honours the
// board_ids config (24-char hex IDs) without hitting member discovery.
func TestReadBoardsUsesConfiguredBoardIDs(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/boards/abcdef0123456789abcdef01" {
			_, _ = w.Write([]byte(`{"id":"abcdef0123456789abcdef01","name":"Configured Board","idOrganization":"org_9"}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	c := trello.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "board_ids": "abcdef0123456789abcdef01"},
		Secrets: map[string]string{"key": "k", "token": "t"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "boards", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read boards: %v", err)
	}
	if len(got) != 1 || got[0]["id"] != "abcdef0123456789abcdef01" {
		t.Fatalf("boards = %+v, want one configured board", got)
	}
}

// TestFixtureModeIsCredentialFree confirms the connector emits deterministic
// records with no network access so conformance can run without live creds.
func TestFixtureModeIsCredentialFree(t *testing.T) {
	c := trello.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"boards", "lists", "cards", "checklists", "actions"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture Read(%s) emitted no records", stream)
		}
		for _, rec := range got {
			if rec["id"] == nil {
				t.Fatalf("fixture %s record missing id: %+v", stream, rec)
			}
		}
	}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

// TestCatalogStreams checks the published catalog exposes the core streams with
// primary keys.
func TestCatalogStreams(t *testing.T) {
	c := trello.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"boards": false, "lists": false, "cards": false, "checklists": false, "actions": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q missing primary key", s.Name)
		}
	}
	for name, found := range want {
		if !found {
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

// TestRegisteredReadOnly verifies self-registration and the read-only capability
// surface (Trello has no safe reverse-ETL write set here).
func TestRegisteredReadOnly(t *testing.T) {
	c := trello.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write disabled", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("trello"); !ok {
		t.Fatal("registry did not resolve trello (self-registration)")
	}
}
