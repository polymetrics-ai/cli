package youtubedata_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	youtubedata "polymetrics.ai/internal/connectors/youtube-data"
)

func TestReadChannelsAuthenticatesAndMaps(t *testing.T) {
	var sawKey, sawPart, sawID string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawKey = r.URL.Query().Get("key")
		sawPart = r.URL.Query().Get("part")
		sawID = r.URL.Query().Get("id")
		if r.URL.Path != "/youtube/v3/channels" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"items":[{"id":"chan_1","snippet":{"title":"Fixture Channel"},"statistics":{"viewCount":"42"}}]}`))
	}))
	defer srv.Close()

	c := youtubedata.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL + "/youtube/v3", "channel_ids": "chan_1"}, Secrets: map[string]string{"api_key": "yt_key"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "channels", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawKey != "yt_key" || sawPart == "" || sawID != "chan_1" {
		t.Fatalf("query key=%q part=%q id=%q, want API-key channel request", sawKey, sawPart, sawID)
	}
	if len(got) != 1 || got[0]["id"] != "chan_1" || got[0]["title"] != "Fixture Channel" {
		t.Fatalf("records = %+v, want mapped channel", got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := youtubedata.New()
	fixture := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), fixture); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
	var rows []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "videos", Config: fixture}, func(rec connectors.Record) error {
		rows = append(rows, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read fixture: %v", err)
	}
	if len(rows) == 0 || rows[0]["id"] == nil {
		t.Fatalf("fixture rows = %+v, want records with id", rows)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "youtube-data" || len(cat.Streams) < 3 {
		t.Fatalf("catalog = %+v, want youtube-data streams", cat)
	}
	if _, ok := connectors.NewRegistry().Get("youtube-data"); !ok {
		t.Fatal("registry did not resolve youtube-data")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
}
