package nytimes_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/nytimes"
)

// TestArchivePaginatesAndAuthenticates is the red-first test: the archive stream
// iterates one request per month between start_date and end_date (here two
// months => two pages), passes the api-key query parameter, and maps records out
// of response.docs.
func TestArchivePaginatesAndAuthenticates(t *testing.T) {
	var sawKey string
	var paths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawKey = r.URL.Query().Get("api-key")
		paths = append(paths, r.URL.Path)
		switch {
		case strings.HasSuffix(r.URL.Path, "/2022/1.json"):
			_, _ = w.Write([]byte(`{"response":{"docs":[{"_id":"a1","web_url":"https://nyt.com/a1","pub_date":"2022-01-05T00:00:00Z","headline":{"main":"Jan A"}},{"_id":"a2","web_url":"https://nyt.com/a2","pub_date":"2022-01-06T00:00:00Z","headline":{"main":"Jan B"}}]}}`))
		case strings.HasSuffix(r.URL.Path, "/2022/2.json"):
			_, _ = w.Write([]byte(`{"response":{"docs":[{"_id":"a3","web_url":"https://nyt.com/a3","pub_date":"2022-02-01T00:00:00Z","headline":{"main":"Feb A"}}]}}`))
		default:
			t.Errorf("unexpected path %q", r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := nytimes.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":   srv.URL,
			"start_date": "2022-01",
			"end_date":   "2022-02",
			"period":     "7",
		},
		Secrets: map[string]string{"api_key": "secret_key_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "archive", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawKey != "secret_key_123" {
		t.Fatalf("api-key = %q, want secret_key_123", sawKey)
	}
	if len(paths) != 2 {
		t.Fatalf("requested %d months, want 2: %v", len(paths), paths)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (two months)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["web_url"] == nil {
			t.Fatalf("record missing id/web_url: %+v", rec)
		}
		if rec["headline"] == nil {
			t.Fatalf("record missing flattened headline: %+v", rec)
		}
	}
}

// TestMostPopularAuthenticates checks the most-popular viewed stream hits the
// correct period-keyed endpoint, passes api-key, and maps results[].
func TestMostPopularAuthenticates(t *testing.T) {
	var sawPath, sawKey string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawPath = r.URL.Path
		sawKey = r.URL.Query().Get("api-key")
		_, _ = w.Write([]byte(`{"status":"OK","num_results":2,"results":[{"id":1,"url":"https://nyt.com/1","title":"One","published_date":"2022-01-01"},{"id":2,"url":"https://nyt.com/2","title":"Two","published_date":"2022-01-02"}]}`))
	}))
	defer srv.Close()

	c := nytimes.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":   srv.URL,
			"start_date": "2022-01",
			"period":     "7",
		},
		Secrets: map[string]string{"api_key": "secret_key_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "most_popular_viewed", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawKey != "secret_key_123" {
		t.Fatalf("api-key = %q, want secret_key_123", sawKey)
	}
	if !strings.HasSuffix(sawPath, "/mostpopular/v2/viewed/7.json") {
		t.Fatalf("path = %q, want suffix /mostpopular/v2/viewed/7.json", sawPath)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2", len(got))
	}
	if got[0]["id"] == nil || got[0]["url"] == nil || got[0]["title"] == nil {
		t.Fatalf("record missing id/url/title: %+v", got[0])
	}
}

// TestFixtureModeNoNetwork ensures the conformance-friendly fixture path emits
// deterministic records with no network access.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := nytimes.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture", "start_date": "2022-01", "period": "7"}}

	for _, stream := range []string{"archive", "most_popular_viewed", "most_popular_emailed", "most_popular_shared"} {
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
}

func TestCheckFixtureMode(t *testing.T) {
	c := nytimes.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
}

func TestCatalogStreams(t *testing.T) {
	c := nytimes.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "nytimes" {
		t.Fatalf("catalog connector = %q, want nytimes", cat.Connector)
	}
	want := map[string]bool{"archive": false, "most_popular_viewed": false, "most_popular_emailed": false, "most_popular_shared": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %s missing primary key", s.Name)
		}
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

func TestRegisteredReadOnly(t *testing.T) {
	_ = nytimes.New() // ensure init ran
	c := nytimes.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("nytimes"); !ok {
		t.Fatal("registry did not resolve nytimes (self-registration)")
	}
}

func TestBaseURLSSRFValidation(t *testing.T) {
	c := nytimes.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example", "start_date": "2022-01", "period": "7"},
		Secrets: map[string]string{"api_key": "k"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "most_popular_viewed", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("expected error for non-http base_url scheme")
	}
}
