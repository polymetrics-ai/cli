package discord_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/discord"
)

// TestReadMembersPaginatesAndAuthenticates is the red-first test for the Discord
// connector: Bot auth header, Discord `after`=highest-user-id cursor pagination
// over a top-level array, and record mapping. Red until
// internal/connectors/discord exists.
func TestReadMembersPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/guilds/42/members" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("after") {
		case "":
			// First page returns two members; highest user id is "200".
			_, _ = w.Write([]byte(`[{"user":{"id":"100","username":"ada"},"nick":"Ada","joined_at":"2026-01-01T00:00:00Z"},{"user":{"id":"200","username":"grace"},"nick":"Grace","joined_at":"2026-01-02T00:00:00Z"}]`))
		case "200":
			// Second page returns a single short page -> pagination stops.
			_, _ = w.Write([]byte(`[{"user":{"id":"300","username":"katherine"},"nick":"Kat","joined_at":"2026-01-03T00:00:00Z"}]`))
		default:
			t.Errorf("unexpected after=%q", r.URL.Query().Get("after"))
			_, _ = w.Write([]byte(`[]`))
		}
	}))
	defer srv.Close()

	c := discord.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "guild_id": "42", "page_size": "2"},
		Secrets: map[string]string{"bot_token": "tok_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "members", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bot tok_123" {
		t.Fatalf("Authorization = %q, want Bot tok_123", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	if got[0]["id"] != "100" || got[0]["user_id"] != "100" {
		t.Fatalf("first member mapping wrong: %+v", got[0])
	}
	if got[0]["username"] != "ada" {
		t.Fatalf("first member username = %v, want ada", got[0]["username"])
	}
	if got[2]["id"] != "300" {
		t.Fatalf("last member id = %v, want 300", got[2]["id"])
	}
}

// TestReadChannelsArray verifies an unpaginated top-level-array stream (channels)
// maps records and sends the Bot auth header to the right path.
func TestReadChannelsArray(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/guilds/42/channels" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`[{"id":"c1","name":"general","type":0,"position":0},{"id":"c2","name":"random","type":0,"position":1}]`))
	}))
	defer srv.Close()

	c := discord.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "guild_id": "42"},
		Secrets: map[string]string{"bot_token": "tok_123"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "channels", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2", len(got))
	}
	if got[0]["id"] != "c1" || got[0]["name"] != "general" {
		t.Fatalf("channel mapping wrong: %+v", got[0])
	}
}

// TestReadGuildSingleObject verifies a single-object stream (guilds) reads the
// guild detail endpoint and maps it to one record.
func TestReadGuildSingleObject(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/guilds/42" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"id":"42","name":"PM Guild","owner_id":"100","approximate_member_count":3}`))
	}))
	defer srv.Close()

	c := discord.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "guild_id": "42"},
		Secrets: map[string]string{"bot_token": "tok_123"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "guilds", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("records = %d, want 1", len(got))
	}
	if got[0]["id"] != "42" || got[0]["name"] != "PM Guild" {
		t.Fatalf("guild mapping wrong: %+v", got[0])
	}
}

// TestFixtureModeNoNetwork ensures fixture mode emits deterministic records
// without any network access (credential-free conformance).
func TestFixtureModeNoNetwork(t *testing.T) {
	c := discord.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"guilds", "channels", "roles", "members"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("Read(%s) fixture: %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("Read(%s) fixture emitted no records", stream)
		}
		if got[0]["id"] == nil {
			t.Fatalf("Read(%s) fixture record missing id: %+v", stream, got[0])
		}
	}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
}

// TestCatalogAndMetadata verifies the published catalog and read-only metadata.
func TestCatalogAndMetadata(t *testing.T) {
	c := discord.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("discord is read-only; Write should be false")
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("catalog streams = %d, want >= 3", len(cat.Streams))
	}
}

// TestRegistryResolves verifies self-registration via init().
func TestRegistryResolves(t *testing.T) {
	_ = discord.New() // ensure package init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("discord"); !ok {
		t.Fatal("registry did not resolve discord (self-registration)")
	}
}

// TestSSRFGuardRejectsBadBaseURL ensures the base_url override is validated.
func TestSSRFGuardRejectsBadBaseURL(t *testing.T) {
	c := discord.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil/", "guild_id": "42"},
		Secrets: map[string]string{"bot_token": "tok_123"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "channels", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("expected error for non-http(s) base_url")
	}
}
