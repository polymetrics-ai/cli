package monday_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics/internal/connectors"
	"polymetrics/internal/connectors/monday"
)

// graphqlRequest is the minimal shape of a monday.com GraphQL POST body.
type graphqlRequest struct {
	Query string `json:"query"`
}

// TestReadPaginatesAndAuthenticates is the red-first test: the monday connector
// must POST GraphQL to /v2, authenticate with the raw token in the Authorization
// header (NO Bearer prefix), paginate the boards stream across two pages using
// the page argument, and map records out of data.boards.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var sawMethod string
	pages := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		sawMethod = r.Method
		if strings.TrimSuffix(r.URL.Path, "/") != "/v2" {
			http.NotFound(w, r)
			return
		}
		var body graphqlRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode body: %v", err)
		}
		if !strings.Contains(body.Query, "boards") {
			t.Errorf("query did not target boards: %q", body.Query)
		}
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.Contains(body.Query, "page: 1"):
			pages++
			_, _ = w.Write([]byte(`{"data":{"boards":[{"id":"1","name":"Board One","state":"active","board_kind":"public"},{"id":"2","name":"Board Two","state":"active","board_kind":"private"}]}}`))
		case strings.Contains(body.Query, "page: 2"):
			pages++
			_, _ = w.Write([]byte(`{"data":{"boards":[{"id":"3","name":"Board Three","state":"archived","board_kind":"public"}]}}`))
		case strings.Contains(body.Query, "page: 3"):
			pages++
			_, _ = w.Write([]byte(`{"data":{"boards":[]}}`))
		default:
			t.Errorf("unexpected query page: %q", body.Query)
			_, _ = w.Write([]byte(`{"data":{"boards":[]}}`))
		}
	}))
	defer srv.Close()

	c := monday.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":  srv.URL + "/v2",
			"page_size": "2",
		},
		Secrets: map[string]string{"credentials.api_token": "tok_abc123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "boards", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawMethod != http.MethodPost {
		t.Fatalf("method = %q, want POST", sawMethod)
	}
	if sawAuth != "tok_abc123" {
		t.Fatalf("Authorization = %q, want raw token tok_abc123 (no Bearer prefix)", sawAuth)
	}
	if pages < 2 {
		t.Fatalf("server served %d pages, want at least 2 (pagination)", pages)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 across 2 pages", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["name"] == nil {
			t.Fatalf("record missing id/name: %+v", rec)
		}
	}
	if got[0]["name"] != "Board One" {
		t.Fatalf("first record name = %v, want Board One", got[0]["name"])
	}
}

// TestFixtureModeNeedsNoNetwork confirms credential-free conformance: fixture
// mode emits deterministic records without any HTTP call.
func TestFixtureModeNeedsNoNetwork(t *testing.T) {
	c := monday.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	for _, stream := range []string{"boards", "items", "users", "teams", "tags"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture Read(%s) produced no records", stream)
		}
		for _, rec := range got {
			if rec["id"] == nil {
				t.Fatalf("fixture %s record missing id: %+v", stream, rec)
			}
		}
	}

	// Check in fixture mode must not error or hit the network.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

// TestItemsCursorPagination verifies the items stream uses monday's cursor-based
// next_items_page pagination model under the data.boards[].items_page envelope.
func TestItemsCursorPagination(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body graphqlRequest
		_ = json.NewDecoder(r.Body).Decode(&body)
		w.Header().Set("Content-Type", "application/json")
		calls++
		switch {
		case strings.Contains(body.Query, "next_items_page"):
			// Second page requested via cursor.
			if !strings.Contains(body.Query, "CUR_2") {
				t.Errorf("next_items_page query missing cursor: %q", body.Query)
			}
			_, _ = w.Write([]byte(`{"data":{"next_items_page":{"cursor":null,"items":[{"id":"i3","name":"Item Three"}]}}}`))
		default:
			_, _ = w.Write([]byte(`{"data":{"boards":[{"id":"1","items_page":{"cursor":"CUR_2","items":[{"id":"i1","name":"Item One"},{"id":"i2","name":"Item Two"}]}}]}}`))
		}
	}))
	defer srv.Close()

	c := monday.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"credentials.api_token": "tok"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "items", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read items: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("items = %d, want 3 (cursor pagination)", len(got))
	}
}

// TestRegistryResolves confirms self-registration and read-only capabilities.
func TestRegistryResolves(t *testing.T) {
	_ = monday.New() // ensure init ran
	c := monday.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog {
		t.Fatalf("capabilities = %+v, want Read && Catalog", caps)
	}
	if caps.Write {
		t.Fatalf("monday should be read-only, got Write=true")
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("monday"); !ok {
		t.Fatal("registry did not resolve monday (self-registration)")
	}
}

// TestCatalogStreams confirms the published catalog exposes the core streams.
func TestCatalogStreams(t *testing.T) {
	c := monday.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"boards": false, "items": false, "users": false, "teams": false, "tags": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %s has no primary key", s.Name)
		}
	}
	for name, found := range want {
		if !found {
			t.Fatalf("catalog missing core stream %q", name)
		}
	}
}
