package canny_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics/internal/connectors"
	"polymetrics/internal/connectors/canny"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Canny
// connector: the apiKey is sent in the POST form body, skip/limit offset
// pagination walks two pages via the hasMore flag, and records are mapped from
// the "posts" array. Red until internal/connectors/canny exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAPIKey string
	var sawMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawMethod = r.Method
		if err := r.ParseForm(); err != nil {
			t.Errorf("ParseForm: %v", err)
		}
		sawAPIKey = r.PostFormValue("apiKey")
		if r.URL.Path != "/posts/list" {
			http.NotFound(w, r)
			return
		}
		skip := r.PostFormValue("skip")
		switch skip {
		case "", "0":
			_, _ = w.Write([]byte(`{"posts":[{"id":"p1","title":"First","created":"2026-01-01T00:00:00Z"},{"id":"p2","title":"Second","created":"2026-01-02T00:00:00Z"}],"hasMore":true}`))
		case "2":
			_, _ = w.Write([]byte(`{"posts":[{"id":"p3","title":"Third","created":"2026-01-03T00:00:00Z"}],"hasMore":false}`))
		default:
			t.Errorf("unexpected skip=%q", skip)
			_, _ = w.Write([]byte(`{"posts":[],"hasMore":false}`))
		}
	}))
	defer srv.Close()

	c := canny.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"api_key": "secret_abc"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "posts", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawMethod != http.MethodPost {
		t.Fatalf("method = %q, want POST", sawMethod)
	}
	if sawAPIKey != "secret_abc" {
		t.Fatalf("apiKey form value = %q, want secret_abc", sawAPIKey)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["title"] == nil || rec["created"] == nil {
			t.Fatalf("record missing id/title/created: %+v", rec)
		}
	}
}

// TestReadCommentsArrayKey confirms a second stream maps from its own response
// key ("comments") rather than the posts key, exercising the per-stream routing.
func TestReadCommentsArrayKey(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/comments/list" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"comments":[{"id":"c1","value":"hi","created":"2026-01-01T00:00:00Z"}],"hasMore":false}`))
	}))
	defer srv.Close()

	c := canny.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "secret_abc"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "comments", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read comments: %v", err)
	}
	if len(got) != 1 || got[0]["id"] != "c1" {
		t.Fatalf("comments = %+v, want one record id c1", got)
	}
}

// TestFixtureMode confirms credential-free fixture reads work (conformance path).
func TestFixtureMode(t *testing.T) {
	c := canny.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture): %v", err)
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "boards", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read(fixture): %v", err)
	}
	if len(got) == 0 {
		t.Fatal("fixture mode emitted no records")
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("fixture record missing id: %+v", rec)
		}
	}
}

func TestCatalogStreams(t *testing.T) {
	c := canny.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "canny" {
		t.Fatalf("catalog connector = %q, want canny", cat.Connector)
	}
	want := map[string]bool{"boards": false, "posts": false, "comments": false, "categories": false, "companies": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q missing primary key", s.Name)
		}
		if len(s.Fields) == 0 {
			t.Fatalf("stream %q missing fields", s.Name)
		}
	}
	for name, found := range want {
		if !found {
			t.Fatalf("catalog missing expected stream %q", name)
		}
	}
}

func TestCheckRejectsMissingSecret(t *testing.T) {
	c := canny.New()
	err := c.Check(context.Background(), connectors.RuntimeConfig{Config: map[string]string{}})
	if err == nil {
		t.Fatal("Check without api_key should fail")
	}
}

func TestBaseURLSSRFRejected(t *testing.T) {
	c := canny.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil"},
		Secrets: map[string]string{"api_key": "secret_abc"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "posts", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("non-http base_url should be rejected")
	}
}

func TestReadOnlyNoWrite(t *testing.T) {
	c := canny.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read/Catalog/Check", caps)
	}
	if caps.Write {
		t.Fatalf("canny should be read-only, got Write=true")
	}
}

func TestRegisteredInRegistry(t *testing.T) {
	_ = canny.New() // ensure init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("canny"); !ok {
		t.Fatal("registry did not resolve canny (self-registration)")
	}
}
