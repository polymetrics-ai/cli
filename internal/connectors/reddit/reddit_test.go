package reddit_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/reddit"
)

func TestReadPostsAuthenticatesPaginatesAndMapsRecords(t *testing.T) {
	var sawAuth string
	var afters []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/r/golang/new" {
			http.NotFound(w, r)
			return
		}
		afters = append(afters, r.URL.Query().Get("after"))
		switch r.URL.Query().Get("after") {
		case "":
			_, _ = w.Write([]byte(`{"data":{"children":[{"kind":"t3","data":{"id":"p1","name":"t3_p1","title":"Go 1.25","subreddit":"golang"}},{"kind":"t3","data":{"id":"p2","name":"t3_p2","title":"Concurrency","subreddit":"golang"}}],"after":"t3_next"}}`))
		case "t3_next":
			_, _ = w.Write([]byte(`{"data":{"children":[{"kind":"t3","data":{"id":"p3","name":"t3_p3","title":"Interfaces","subreddit":"golang"}}],"after":null}}`))
		default:
			t.Fatalf("unexpected after %q", r.URL.Query().Get("after"))
		}
	}))
	defer srv.Close()

	c := reddit.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "subreddit": "golang", "page_size": "2"}, Secrets: map[string]string{"access_token": "reddit_token"}}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "posts", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer reddit_token" {
		t.Fatalf("Authorization = %q", sawAuth)
	}
	if len(afters) != 2 || afters[0] != "" || afters[1] != "t3_next" {
		t.Fatalf("afters = %v", afters)
	}
	if len(got) != 3 || got[0]["id"] != "p1" || got[0]["title"] != "Go 1.25" {
		t.Fatalf("unexpected records: %+v", got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := reddit.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
	cat, err := c.Catalog(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "reddit" || len(cat.Streams) == 0 {
		t.Fatalf("catalog = %+v", cat)
	}
	for _, stream := range cat.Streams {
		var got []connectors.Record
		if err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream.Name, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		}); err != nil {
			t.Fatalf("Read fixture %s: %v", stream.Name, err)
		}
		if len(got) == 0 || got[0]["id"] == nil {
			t.Fatalf("fixture %s records = %+v", stream.Name, got)
		}
	}
	if _, ok := connectors.NewRegistry().Get("reddit"); !ok {
		t.Fatal("registry did not resolve reddit")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{Config: cfg}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v", err)
	}
}
