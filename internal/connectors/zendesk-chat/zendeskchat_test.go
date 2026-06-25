package zendeskchat_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	zendeskchat "polymetrics.ai/internal/connectors/zendesk-chat"
)

// TestReadArrayStreamAuthenticates verifies Bearer auth and top-level-array
// extraction for a simple list stream (agents). Red until the package exists.
func TestReadArrayStreamAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/agents" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`[{"id":1,"display_name":"Ada","email":"ada@example.com","enabled":true},{"id":2,"display_name":"Grace","email":"grace@example.com","enabled":false}]`))
	}))
	defer srv.Close()

	c := zendeskchat.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"credentials.access_token": "tok_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "agents", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer tok_123" {
		t.Fatalf("Authorization = %q, want Bearer tok_123", sawAuth)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2", len(got))
	}
	if got[0]["id"] == nil || got[0]["display_name"] == nil {
		t.Fatalf("record missing fields: %+v", got[0])
	}
}

// TestReadChatsPaginatesNextURL verifies the chats incremental-export stream
// follows next_url across two pages and stops on a short page.
func TestReadChatsPaginatesNextURL(t *testing.T) {
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chats" {
			http.NotFound(w, r)
			return
		}
		// Page 1: full page with a next_url pointing back at the same server.
		if r.URL.Query().Get("cursor") == "" {
			_, _ = w.Write([]byte(`{"chats":[{"id":"c1","type":"chat"},{"id":"c2","type":"chat"}],"next_url":"` + srv.URL + `/chats?cursor=PAGE2","count":2}`))
			return
		}
		// Page 2: final page, no further next_url.
		_, _ = w.Write([]byte(`{"chats":[{"id":"c3","type":"chat"}],"next_url":null,"count":1}`))
	}))
	defer srv.Close()

	c := zendeskchat.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "start_date": "2021-02-01T00:00:00Z"},
		Secrets: map[string]string{"credentials.access_token": "tok_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "chats", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	if got[2]["id"] != "c3" {
		t.Fatalf("last record id = %v, want c3", got[2]["id"])
	}
}

// TestFixtureModeNoNetwork confirms credential-free fixture reads work for
// conformance without a live server.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := zendeskchat.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "agents", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) == 0 {
		t.Fatal("fixture mode emitted no records")
	}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

// TestCatalogStreams confirms the published stream set is non-empty and that
// every stream carries a primary key.
func TestCatalogStreams(t *testing.T) {
	c := zendeskchat.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("streams = %d, want >= 3", len(cat.Streams))
	}
	for _, s := range cat.Streams {
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q missing primary key", s.Name)
		}
	}
}

// TestRegistryResolution confirms self-registration via NewRegistry.
func TestRegistryResolution(t *testing.T) {
	_ = zendeskchat.New() // ensure init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("zendesk-chat"); !ok {
		t.Fatal("registry did not resolve zendesk-chat (self-registration)")
	}
	caps := zendeskchat.New().Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
}
