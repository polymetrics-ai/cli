package pylon_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/pylon"
)

func TestReadIssuesAuthenticatesPaginatesAndMaps(t *testing.T) {
	var sawAuth string
	var cursors []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/issues" {
			http.NotFound(w, r)
			return
		}
		sawAuth = r.Header.Get("Authorization")
		cursors = append(cursors, r.URL.Query().Get("cursor"))
		switch r.URL.Query().Get("cursor") {
		case "":
			_, _ = w.Write([]byte(`{"data":[{"id":"iss_1","title":"First","state":"open","updated_at":"2026-01-01T00:00:00Z"}],"pagination":{"next_cursor":"next-1"}}`))
		case "next-1":
			_, _ = w.Write([]byte(`{"data":[{"id":"iss_2","title":"Second","state":"closed","updated_at":"2026-01-02T00:00:00Z"}],"pagination":{"next_cursor":null}}`))
		default:
			t.Fatalf("unexpected cursor %q", r.URL.Query().Get("cursor"))
		}
	}))
	defer srv.Close()

	c := pylon.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL}, Secrets: map[string]string{"api_token": "pylon-token"}}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "issues", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer pylon-token" {
		t.Fatalf("Authorization = %q", sawAuth)
	}
	if len(cursors) != 2 || cursors[0] != "" || cursors[1] != "next-1" {
		t.Fatalf("cursors = %v", cursors)
	}
	if len(got) != 2 || got[0]["id"] != "iss_1" || got[1]["title"] != "Second" {
		t.Fatalf("unexpected records: %+v", got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := pylon.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "issues", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) == 0 || got[0]["id"] == nil {
		t.Fatalf("fixture records not mapped: %+v", got)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "pylon" || len(cat.Streams) < 4 {
		t.Fatalf("unexpected catalog: %+v", cat)
	}
	for _, stream := range cat.Streams {
		if len(stream.PrimaryKey) == 0 {
			t.Fatalf("stream %q missing primary key", stream.Name)
		}
	}
	if _, ok := connectors.NewRegistry().Get("pylon"); !ok {
		t.Fatal("registry did not resolve pylon")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v", err)
	}
}
