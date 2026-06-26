package tmdb_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/tmdb"
)

func TestReadPopularMoviesPaginatesAndAuthenticates(t *testing.T) {
	var sawKey, sawLanguage string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawKey = r.URL.Query().Get("api_key")
		sawLanguage = r.URL.Query().Get("language")
		if r.URL.Path != "/movie/popular" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page") {
		case "1":
			_, _ = w.Write([]byte(`{"results":[{"id":1,"title":"One"},{"id":2,"title":"Two"}]}`))
		case "2":
			_, _ = w.Write([]byte(`{"results":[{"id":3,"title":"Three"}]}`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
			_, _ = w.Write([]byte(`{"results":[]}`))
		}
	}))
	defer srv.Close()

	c := tmdb.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "language": "en-US", "page_size": "2"}, Secrets: map[string]string{"api_key": "tmdb-key"}}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "popular_movies", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawKey != "tmdb-key" || sawLanguage != "en-US" {
		t.Fatalf("query auth/language = %q/%q", sawKey, sawLanguage)
	}
	if len(got) != 3 || got[0]["title"] == nil {
		t.Fatalf("records = %+v", got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := tmdb.New()
	assertConnector(t, c, "tmdb")
}

func assertConnector(t *testing.T, c connectors.Connector, name string) {
	t.Helper()
	if c.Name() != name {
		t.Fatalf("Name = %q, want %q", c.Name(), name)
	}
	caps := c.Metadata().Capabilities
	if !caps.Check || !caps.Catalog || !caps.Read || caps.Write {
		t.Fatalf("capabilities = %+v, want check/catalog/read and no write", caps)
	}
	fixture := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), fixture); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	cat, err := c.Catalog(context.Background(), fixture)
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != name || len(cat.Streams) == 0 {
		t.Fatalf("catalog = %+v", cat)
	}
	for _, stream := range cat.Streams {
		var got []connectors.Record
		if err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream.Name, Config: fixture}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		}); err != nil {
			t.Fatalf("fixture Read(%s): %v", stream.Name, err)
		}
		if len(got) == 0 || got[0]["id"] == nil {
			t.Fatalf("fixture %s records = %+v", stream.Name, got)
		}
	}
	result, err := c.Write(context.Background(), connectors.WriteRequest{Config: fixture}, []connectors.Record{{"id": "1"}})
	if !errors.Is(err, connectors.ErrUnsupportedOperation) || result.RecordsFailed != 1 {
		t.Fatalf("Write = (%+v, %v), want unsupported with one failed record", result, err)
	}
	if _, ok := connectors.NewRegistry().Get(name); !ok {
		t.Fatalf("registry did not resolve %s", name)
	}
}
