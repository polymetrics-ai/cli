package dockerhub_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/dockerhub"
)

// TestReadRepositoriesPaginates is the red-first test for the DockerHub
// connector: the Docker Hub registry API is unauthenticated, paginates via a
// {count,next,previous,results} envelope where `next` is an absolute URL, and
// records live at results[]. Red until internal/connectors/dockerhub exists.
func TestReadRepositoriesPaginates(t *testing.T) {
	var sawAuth string
	var paths []string
	var srvURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		paths = append(paths, r.URL.Path+"?"+r.URL.RawQuery)
		if r.URL.Path != "/v2/repositories/upstream/" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page") {
		case "", "1":
			// next points at an absolute URL the connector must follow as-is.
			_, _ = w.Write([]byte(`{"count":3,"next":"` + srvURL + `/v2/repositories/upstream/?page=2&page_size=2","previous":null,"results":[{"name":"source-a","namespace":"upstream","pull_count":10,"star_count":1,"last_updated":"2026-01-01T00:00:00Z","is_private":false},{"name":"source-b","namespace":"upstream","pull_count":20,"star_count":2,"last_updated":"2026-01-02T00:00:00Z","is_private":false}]}`))
		case "2":
			_, _ = w.Write([]byte(`{"count":3,"next":null,"previous":null,"results":[{"name":"source-c","namespace":"upstream","pull_count":30,"star_count":3,"last_updated":"2026-01-03T00:00:00Z","is_private":false}]}`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
		}
	}))
	defer srv.Close()
	srvURL = srv.URL

	c := dockerhub.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{"base_url": srv.URL + "/v2", "docker_username": "upstream", "page_size": "2"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "repositories", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	// DockerHub public API takes no auth header.
	if sawAuth != "" {
		t.Fatalf("Authorization = %q, want empty (public API)", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 across 2 pages; paths=%v", len(got), paths)
	}
	for _, rec := range got {
		if rec["name"] == nil || rec["namespace"] == nil {
			t.Fatalf("record missing name/namespace: %+v", rec)
		}
	}
	if got[0]["name"] != "source-a" || got[2]["name"] != "source-c" {
		t.Fatalf("unexpected record order: %v", got)
	}
}

// TestReadTagsPerRepository exercises the tags stream, which reads
// /repositories/{username}/{repository}/tags.
func TestReadTagsPerRepository(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/repositories/upstream/source-dockerhub/tags" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"count":1,"next":null,"previous":null,"results":[{"name":"latest","full_size":1234,"last_updated":"2026-02-01T00:00:00Z","digest":"sha256:abc"}]}`))
	}))
	defer srv.Close()

	c := dockerhub.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{"base_url": srv.URL + "/v2", "docker_username": "upstream", "repository": "source-dockerhub"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "tags", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read tags: %v", err)
	}
	if len(got) != 1 || got[0]["name"] != "latest" {
		t.Fatalf("tags = %v, want one tag named latest", got)
	}
}

// TestFixtureMode confirms credential-free conformance reads emit deterministic
// records without any network access.
func TestFixtureMode(t *testing.T) {
	c := dockerhub.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture", "docker_username": "upstream"}}
	for _, stream := range []string{"repositories", "tags", "namespace"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture Read(%s) emitted no records", stream)
		}
	}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

// TestCheckRejectsSSRFBaseURL ensures an override base URL must be http/https.
func TestCheckRejectsSSRFBaseURL(t *testing.T) {
	c := dockerhub.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": "file:///etc/passwd", "docker_username": "upstream"}}
	if err := c.Check(context.Background(), cfg); err == nil {
		t.Fatal("Check should reject a non-http(s) base_url")
	}
}

func TestCatalogAndRegistry(t *testing.T) {
	c := dockerhub.New()
	if c.Name() != "dockerhub" {
		t.Fatalf("Name = %q, want dockerhub", c.Name())
	}
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || caps.Write {
		t.Fatalf("capabilities = %+v, want Read && Catalog && !Write", caps)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("streams = %d, want >= 3", len(cat.Streams))
	}
	names := map[string]bool{}
	for _, s := range cat.Streams {
		names[s.Name] = true
	}
	for _, want := range []string{"repositories", "tags", "namespace"} {
		if !names[want] {
			t.Fatalf("catalog missing stream %q (have %v)", want, names)
		}
	}

	r := connectors.NewRegistry()
	if _, ok := r.Get("dockerhub"); !ok {
		t.Fatal("registry did not resolve dockerhub (self-registration)")
	}
}

func TestUnknownStream(t *testing.T) {
	c := dockerhub.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture", "docker_username": "upstream"}}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "nope", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "nope") {
		t.Fatalf("expected unknown-stream error, got %v", err)
	}
}
