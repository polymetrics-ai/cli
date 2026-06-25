package buzzsprout_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/buzzsprout"
)

// TestReadEpisodesAuthAndPagination is the red-first test: it asserts the
// Buzzsprout token auth header, the .json episodes path scoped by podcast_id,
// pagination across two pages (the connector follows page=N until a short page),
// and record mapping of the episode fields.
func TestReadEpisodesAuthAndPagination(t *testing.T) {
	var sawAuth string
	var sawUA string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		sawUA = r.Header.Get("User-Agent")
		if r.URL.Path != "/api/4321/episodes.json" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Query().Get("page") {
		case "", "1":
			// A full page (page size 2) signals there may be more.
			_, _ = w.Write([]byte(`[{"id":1,"title":"Ep One","published_at":"2026-01-01T00:00:00Z","episode_number":1,"duration":1800,"total_plays":10},{"id":2,"title":"Ep Two","published_at":"2026-01-02T00:00:00Z","episode_number":2,"duration":1200,"total_plays":20}]`))
		case "2":
			// A short page (1 < page size) ends pagination.
			_, _ = w.Write([]byte(`[{"id":3,"title":"Ep Three","published_at":"2026-01-03T00:00:00Z","episode_number":3,"duration":900,"total_plays":30}]`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
			_, _ = w.Write([]byte(`[]`))
		}
	}))
	defer srv.Close()

	c := buzzsprout.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":   srv.URL,
			"podcast_id": "4321",
			"page_size":  "2",
		},
		Secrets: map[string]string{"api_key": "tok_abc123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "episodes", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Token token=tok_abc123" {
		t.Fatalf("Authorization = %q, want Token token=tok_abc123", sawAuth)
	}
	if sawUA == "" {
		t.Fatal("expected a User-Agent header to be sent (Buzzsprout blocks empty UA)")
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	first := got[0]
	if fmt.Sprintf("%v", first["id"]) != "1" {
		t.Fatalf("first record id = %v, want 1", first["id"])
	}
	if first["title"] != "Ep One" {
		t.Fatalf("first record title = %v, want Ep One", first["title"])
	}
	if first["published_at"] != "2026-01-01T00:00:00Z" {
		t.Fatalf("first record published_at = %v, want 2026-01-01T00:00:00Z", first["published_at"])
	}
}

// TestReadPodcastsAccountLevel checks the podcasts stream hits the account-level
// /api/podcasts.json path (not scoped to a podcast_id) and maps records.
func TestReadPodcastsAccountLevel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/podcasts.json" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":99,"title":"My Show","author":"Ada","language":"en"}]`))
	}))
	defer srv.Close()

	c := buzzsprout.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "podcast_id": "4321"},
		Secrets: map[string]string{"api_key": "tok_abc123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "podcasts", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("records = %d, want 1", len(got))
	}
	if got[0]["title"] != "My Show" {
		t.Fatalf("title = %v, want My Show", got[0]["title"])
	}
}

// TestFixtureModeNoNetwork verifies fixture mode emits deterministic records
// without any network access, so conformance can run credential-free.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := buzzsprout.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	for _, stream := range []string{"episodes", "podcasts"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("Read(%s) fixture: %v", stream, err)
		}
		if len(got) != 2 {
			t.Fatalf("fixture %s records = %d, want 2", stream, len(got))
		}
		for _, rec := range got {
			if rec["id"] == nil {
				t.Fatalf("fixture %s record missing id: %+v", stream, rec)
			}
		}
	}

	// Check in fixture mode must succeed without creds or network.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

func TestCheckRequiresSecretAndPodcastID(t *testing.T) {
	c := buzzsprout.New()
	// Missing api_key.
	err := c.Check(context.Background(), connectors.RuntimeConfig{
		Config: map[string]string{"podcast_id": "4321"},
	})
	if err == nil {
		t.Fatal("Check without api_key should fail")
	}
}

func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := buzzsprout.New()
	err := c.Read(context.Background(), connectors.ReadRequest{
		Stream: "episodes",
		Config: connectors.RuntimeConfig{
			Config:  map[string]string{"base_url": "file:///etc/passwd", "podcast_id": "4321"},
			Secrets: map[string]string{"api_key": "tok"},
		},
	}, func(connectors.Record) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "base_url") {
		t.Fatalf("expected base_url scheme rejection, got %v", err)
	}
}

func TestCatalogAndMetadata(t *testing.T) {
	c := buzzsprout.New()
	md := c.Metadata()
	if md.Name != "buzzsprout" {
		t.Fatalf("metadata name = %q, want buzzsprout", md.Name)
	}
	if !md.Capabilities.Read || !md.Capabilities.Catalog || !md.Capabilities.Check {
		t.Fatalf("capabilities = %+v, want Read/Catalog/Check", md.Capabilities)
	}
	if md.Capabilities.Write {
		t.Fatalf("buzzsprout is read-only, Write should be false")
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 2 {
		t.Fatalf("expected at least 2 streams, got %d", len(cat.Streams))
	}
	names := map[string]bool{}
	for _, s := range cat.Streams {
		names[s.Name] = true
	}
	if !names["episodes"] || !names["podcasts"] {
		t.Fatalf("missing core streams, got %v", names)
	}
}

func TestRegistryResolvesBuzzsprout(t *testing.T) {
	_ = buzzsprout.New() // ensure package init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("buzzsprout"); !ok {
		t.Fatal("registry did not resolve buzzsprout (self-registration)")
	}
}
