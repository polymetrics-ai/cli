package jinaaireader_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	jinaaireader "polymetrics.ai/internal/connectors/jina-ai-reader"
)

func TestReadTraversesURLsAuthenticatesAndMapsPages(t *testing.T) {
	var sawAuth string
	var paths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		paths = append(paths, r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/http://example.com/a":
			_, _ = w.Write([]byte(`{"data":{"url":"http://example.com/a","title":"A","content":"Alpha"}}`))
		case "/http://example.com/b":
			_, _ = w.Write([]byte(`{"data":{"url":"http://example.com/b","title":"B","content":"Beta"}}`))
		default:
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
	}))
	defer srv.Close()

	c := jinaaireader.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "urls": "http://example.com/a,http://example.com/b"},
		Secrets: map[string]string{"api_key": "jina_key"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "pages", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer jina_key" {
		t.Fatalf("Authorization = %q", sawAuth)
	}
	if len(paths) != 2 || len(got) != 2 {
		t.Fatalf("paths=%v records=%+v", paths, got)
	}
	if got[0]["url"] != "http://example.com/a" || got[0]["content"] != "Alpha" {
		t.Fatalf("record not mapped: %+v", got[0])
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := jinaaireader.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture): %v", err)
	}
	var fixture []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "pages", Config: cfg}, func(rec connectors.Record) error {
		fixture = append(fixture, rec)
		return nil
	}); err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(fixture) == 0 || fixture[0]["fixture"] != true {
		t.Fatalf("fixture records = %+v", fixture)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil || cat.Connector != "jina-ai-reader" || len(cat.Streams) == 0 {
		t.Fatalf("catalog = %+v err=%v", cat, err)
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write err = %v", err)
	}
	if got, ok := connectors.NewRegistry().Get("jina-ai-reader"); !ok || got.Metadata().Capabilities.Write {
		t.Fatalf("registry/capabilities failed: ok=%v got=%T", ok, got)
	}
}
