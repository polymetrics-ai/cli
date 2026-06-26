package gutendex_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/gutendex"
)

// TestReadPaginatesAndMapsRecords is the red-first test for the Gutendex
// connector: no auth header is sent (the API is public), DRF "next" URL
// pagination is followed across two pages, query filters for the stream are
// applied, and records are flattened. Red until internal/connectors/gutendex
// is implemented.
func TestReadPaginatesAndMapsRecords(t *testing.T) {
	var (
		sawAuth string
		sawSort string
		paths   []string
	)
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The public Gutendex API requires no credentials; assert we never
		// leak an Authorization header.
		if v := r.Header.Get("Authorization"); v != "" {
			sawAuth = v
		}
		paths = append(paths, r.URL.Path)
		if r.URL.Path != "/books/" && r.URL.Path != "/books" {
			http.NotFound(w, r)
			return
		}
		if s := r.URL.Query().Get("sort"); s != "" {
			sawSort = s
		}
		switch r.URL.Query().Get("page") {
		case "", "1":
			// next is an absolute URL pointing back at this test server.
			_, _ = w.Write([]byte(`{"count":3,"next":"` + srv.URL + `/books/?page=2","previous":null,"results":[` +
				`{"id":2701,"title":"Moby Dick","authors":[{"name":"Melville, Herman","birth_year":1819,"death_year":1891}],"languages":["en"],"copyright":false,"media_type":"Text","download_count":135445},` +
				`{"id":1342,"title":"Pride and Prejudice","authors":[{"name":"Austen, Jane","birth_year":1775,"death_year":1817}],"languages":["en"],"copyright":false,"media_type":"Text","download_count":117126}]}`))
		case "2":
			_, _ = w.Write([]byte(`{"count":3,"next":null,"previous":"` + srv.URL + `/books/?page=1","results":[` +
				`{"id":84,"title":"Frankenstein","authors":[{"name":"Shelley, Mary","birth_year":1797,"death_year":1851}],"languages":["en"],"copyright":false,"media_type":"Text","download_count":80000}]}`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
			_, _ = w.Write([]byte(`{"count":0,"next":null,"previous":null,"results":[]}`))
		}
	}))
	defer srv.Close()

	c := gutendex.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL}}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "popular_books", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "" {
		t.Fatalf("unexpected Authorization header %q; Gutendex is unauthenticated", sawAuth)
	}
	if sawSort != "popular" {
		t.Fatalf("sort = %q, want popular for popular_books stream", sawSort)
	}
	if len(paths) != 2 {
		t.Fatalf("requested %d pages (%v), want 2", len(paths), paths)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	first := got[0]
	if first["id"] == nil || first["title"] == nil {
		t.Fatalf("record missing id/title: %+v", first)
	}
	if first["author_name"] != "Melville, Herman" {
		t.Fatalf("author_name = %v, want Melville, Herman", first["author_name"])
	}
	langs, ok := first["languages"].(string)
	if !ok || !strings.Contains(langs, "en") {
		t.Fatalf("languages = %v, want a string containing en", first["languages"])
	}
}

// TestFixtureModeNoNetwork confirms fixture mode emits deterministic records
// without any network access, so conformance can run credential-free.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := gutendex.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "books", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) == 0 {
		t.Fatal("fixture mode emitted no records")
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["title"] == nil {
			t.Fatalf("fixture record missing id/title: %+v", rec)
		}
	}
}

// TestCheckFixtureMode confirms Check short-circuits in fixture mode.
func TestCheckFixtureMode(t *testing.T) {
	c := gutendex.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

// TestCatalogStreams confirms the catalog exposes the core streams.
func TestCatalogStreams(t *testing.T) {
	c := gutendex.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"books": false, "popular_books": false, "latest_books": false, "english_books": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q missing primary key", s.Name)
		}
	}
	for name, found := range want {
		if !found {
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

// TestBaseURLRejectsBadScheme confirms SSRF guard on base_url override.
func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := gutendex.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": "file:///etc/passwd"}}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "books", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read with file:// base_url should be rejected")
	}
}

// TestRegistryResolves confirms self-registration resolves via the registry.
func TestRegistryResolves(t *testing.T) {
	_ = gutendex.New() // ensure init ran
	r := connectors.NewRegistry()
	conn, ok := r.Get("gutendex")
	if !ok {
		t.Fatal("registry did not resolve gutendex (self-registration)")
	}
	caps := conn.Metadata().Capabilities
	if !caps.Read || !caps.Catalog {
		t.Fatalf("capabilities = %+v, want Read && Catalog", caps)
	}
	if caps.Write {
		t.Fatalf("gutendex is read-only; Write should be false, got %+v", caps)
	}
}
