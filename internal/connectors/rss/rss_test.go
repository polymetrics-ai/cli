package rss_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/rss"
)

func TestReadItemsParsesRSSFeedWithoutCredentials(t *testing.T) {
	var sawAuth bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" {
			sawAuth = true
		}
		w.Header().Set("Content-Type", "application/rss+xml")
		_, _ = w.Write([]byte(`<?xml version="1.0"?><rss version="2.0"><channel><title>Fixture Feed</title><link>https://example.test</link><item><guid>one</guid><title>First item</title><link>https://example.test/one</link><pubDate>Mon, 02 Jan 2006 15:04:05 MST</pubDate></item><item><guid>two</guid><title>Second item</title><link>https://example.test/two</link></item></channel></rss>`))
	}))
	defer srv.Close()

	c := rss.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"feed_url": srv.URL}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "items", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth {
		t.Fatal("rss connector sent credentials")
	}
	if len(got) != 2 || got[0]["id"] != "one" || got[0]["title"] != "First item" {
		t.Fatalf("items not parsed: %+v", got)
	}
	got = nil
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "channel", Config: cfg}, func(rec connectors.Record) error { got = append(got, rec); return nil }); err != nil {
		t.Fatalf("Read channel: %v", err)
	}
	if len(got) != 1 || got[0]["title"] != "Fixture Feed" {
		t.Fatalf("channel not parsed: %+v", got)
	}
}

func TestFixtureCatalogRegistrationAndReadOnly(t *testing.T) {
	c := rss.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "rss" || len(cat.Streams) != 2 {
		t.Fatalf("unexpected catalog: %+v", cat)
	}
	for _, stream := range cat.Streams {
		count := 0
		if err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream.Name, Config: cfg}, func(connectors.Record) error { count++; return nil }); err != nil {
			t.Fatalf("fixture Read(%s): %v", stream.Name, err)
		}
		if count == 0 || len(stream.PrimaryKey) == 0 {
			t.Fatalf("stream %q not fixture-ready: count=%d pk=%v", stream.Name, count, stream.PrimaryKey)
		}
	}
	if _, ok := connectors.NewRegistry().Get("rss"); !ok {
		t.Fatal("registry did not resolve rss")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v", err)
	}
}
