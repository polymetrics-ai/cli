package metricool_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/metricool"
)

// TestReadAuthenticatesAndIteratesBlogs is the red-first test: it asserts the
// X-Mc-Auth header carries the user_token, that userId is propagated as a query
// param, that the connector issues one request per blog_id (the partition router,
// since the API itself is not paginated) and walks across both pages of blogs,
// and that records are mapped with blogId stamped on.
func TestReadAuthenticatesAndIteratesBlogs(t *testing.T) {
	var sawAuth string
	var sawUserID string
	sawBlogs := map[string]bool{}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("X-Mc-Auth")
		sawUserID = r.URL.Query().Get("userId")
		if r.URL.Path != "/stats/instagram/posts" {
			http.NotFound(w, r)
			return
		}
		blog := r.URL.Query().Get("blogId")
		sawBlogs[blog] = true
		switch blog {
		case "111":
			_, _ = w.Write([]byte(`[{"postId":"p1","interactions":10},{"postId":"p2","interactions":20}]`))
		case "222":
			_, _ = w.Write([]byte(`[{"postId":"p3","interactions":30}]`))
		default:
			t.Errorf("unexpected blogId=%q", blog)
			_, _ = w.Write([]byte(`[]`))
		}
	}))
	defer srv.Close()

	c := metricool.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url": srv.URL,
			"user_id":  "9999",
			"blog_ids": "111,222",
		},
		Secrets: map[string]string{"user_token": "tok_abc"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "instagram_posts", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "tok_abc" {
		t.Fatalf("X-Mc-Auth = %q, want tok_abc", sawAuth)
	}
	if sawUserID != "9999" {
		t.Fatalf("userId = %q, want 9999", sawUserID)
	}
	if !sawBlogs["111"] || !sawBlogs["222"] {
		t.Fatalf("expected requests for both blogs, saw %v", sawBlogs)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (across 2 blogs)", len(got))
	}
	blogsSeen := map[string]bool{}
	for _, rec := range got {
		if rec["postId"] == nil {
			t.Fatalf("record missing postId: %+v", rec)
		}
		if rec["blogId"] == nil {
			t.Fatalf("record missing stamped blogId: %+v", rec)
		}
		blogsSeen[rec["blogId"].(string)] = true
	}
	if !blogsSeen["111"] || !blogsSeen["222"] {
		t.Fatalf("records not stamped with both blogIds: %v", blogsSeen)
	}
}

// TestReadV2DataPath verifies that a v2 stream extracts records from the "data"
// envelope rather than the root array.
func TestReadV2DataPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/analytics/posts/tiktok" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"data":[{"videoId":"v1","views":5},{"videoId":"v2","views":6}]}`))
	}))
	defer srv.Close()

	c := metricool.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url": srv.URL,
			"user_id":  "9999",
			"blog_ids": "111",
		},
		Secrets: map[string]string{"user_token": "tok_abc"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "tiktok_posts", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2 from data[] envelope", len(got))
	}
	for _, rec := range got {
		if rec["videoId"] == nil {
			t.Fatalf("record missing videoId: %+v", rec)
		}
	}
}

func TestFixtureModeNeedsNoNetwork(t *testing.T) {
	c := metricool.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "instagram_posts", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("fixture records = %d, want 2", len(got))
	}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

func TestCatalogAndRegistry(t *testing.T) {
	c := metricool.New()
	if c.Name() != "metricool" {
		t.Fatalf("Name = %q, want metricool", c.Name())
	}
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read/Catalog/Check", caps)
	}
	if caps.Write {
		t.Fatalf("metricool is read-only; Write should be false")
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("want at least 3 streams, got %d", len(cat.Streams))
	}
	names := make([]string, 0, len(cat.Streams))
	for _, s := range cat.Streams {
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q missing primary key", s.Name)
		}
		names = append(names, s.Name)
	}
	sort.Strings(names)

	r := connectors.NewRegistry()
	if _, ok := r.Get("metricool"); !ok {
		t.Fatal("registry did not resolve metricool (self-registration)")
	}
}

func TestCheckRequiresToken(t *testing.T) {
	c := metricool.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"user_id": "9999", "blog_ids": "111"},
		Secrets: map[string]string{},
	}
	if err := c.Check(context.Background(), cfg); err == nil {
		t.Fatal("Check should fail without user_token")
	}
}
