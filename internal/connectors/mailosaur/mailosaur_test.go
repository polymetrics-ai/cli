package mailosaur_test

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/mailosaur"
)

// TestReadMessagesPaginatesAndAuthenticates is the red-first test: HTTP Basic
// auth (API key as the password), page/itemsPerPage pagination over the
// {"items":[...]} payload across two pages, and record mapping.
func TestReadMessagesPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/messages" {
			http.NotFound(w, r)
			return
		}
		if r.URL.Query().Get("server") != "srv1" {
			t.Errorf("missing server param, got %q", r.URL.Query().Get("server"))
		}
		switch r.URL.Query().Get("page") {
		case "0":
			_, _ = w.Write([]byte(`{"items":[{"id":"msg_1","received":"2026-01-01T00:00:00Z","subject":"Hello","type":"Email"},{"id":"msg_2","received":"2026-01-02T00:00:00Z","subject":"World","type":"Email"}]}`))
		case "1":
			_, _ = w.Write([]byte(`{"items":[{"id":"msg_3","received":"2026-01-03T00:00:00Z","subject":"Again","type":"Email"}]}`))
		default:
			_, _ = w.Write([]byte(`{"items":[]}`))
		}
	}))
	defer srv.Close()

	c := mailosaur.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":       srv.URL,
			"server":         "srv1",
			"items_per_page": "2",
		},
		Secrets: map[string]string{"password": "key_abc"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "messages", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("api:key_abc"))
	if sawAuth != wantAuth {
		t.Fatalf("Authorization = %q, want %q", sawAuth, wantAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["received"] == nil {
			t.Fatalf("record missing id/received: %+v", rec)
		}
	}
}

// TestReadServers covers the top-level-array stream (servers).
func TestReadServers(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/servers" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`[{"id":"srv1","name":"Test","messages":5},{"id":"srv2","name":"Prod","messages":9}]`))
	}))
	defer srv.Close()

	c := mailosaur.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"password": "key_abc"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "servers", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read servers: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("servers = %d, want 2", len(got))
	}
	if got[0]["id"] != "srv1" || got[0]["name"] != "Test" {
		t.Fatalf("unexpected first server: %+v", got[0])
	}
}

// TestFixtureMode confirms the credential-free deterministic path.
func TestFixtureMode(t *testing.T) {
	c := mailosaur.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	// Primary key field per stream — transactions is keyed by timestamp, not id.
	pkField := map[string]string{"servers": "id", "messages": "id", "transactions": "timestamp"}
	for _, stream := range []string{"servers", "messages", "transactions"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture stream %s emitted no records", stream)
		}
		if got[0][pkField[stream]] == nil {
			t.Fatalf("fixture stream %s record missing %s: %+v", stream, pkField[stream], got[0])
		}
	}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

// TestSSRFGuard ensures a bad base_url scheme is rejected.
func TestSSRFGuard(t *testing.T) {
	c := mailosaur.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "file:///etc/passwd"},
		Secrets: map[string]string{"password": "key_abc"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "servers", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "base_url") {
		t.Fatalf("expected base_url scheme rejection, got %v", err)
	}
}

// TestRegistryResolution confirms self-registration and read-only capabilities.
func TestRegistryResolution(t *testing.T) {
	_ = mailosaur.New()
	caps := mailosaur.New().Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("mailosaur should be read-only, got Write=true")
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("mailosaur"); !ok {
		t.Fatal("registry did not resolve mailosaur (self-registration)")
	}
}

// TestCatalog confirms the published streams.
func TestCatalog(t *testing.T) {
	c := mailosaur.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "mailosaur" {
		t.Fatalf("connector = %q, want mailosaur", cat.Connector)
	}
	names := map[string]bool{}
	for _, s := range cat.Streams {
		names[s.Name] = true
	}
	for _, want := range []string{"servers", "messages", "transactions"} {
		if !names[want] {
			t.Fatalf("catalog missing stream %q (got %v)", want, names)
		}
	}
}
