package imagga_test

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/imagga"
)

// TestReadTagsBasicAuthAndMultiImage is the red-first test for the Imagga
// connector. Imagga's /tags endpoint analyzes one image per request and returns
// a result object with a tags array. The connector iterates over each configured
// image URL (a request per image == a "page"), authenticates with HTTP Basic
// (api_key:api_secret), and maps result.tags into one record per tag.
func TestReadTagsBasicAuthAndMultiImage(t *testing.T) {
	var sawAuth string
	seenImages := map[string]bool{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/tags" {
			http.NotFound(w, r)
			return
		}
		img := r.URL.Query().Get("image_url")
		seenImages[img] = true
		switch img {
		case "https://example.com/a.jpg":
			_, _ = w.Write([]byte(`{"result":{"tags":[{"confidence":99.1,"tag":{"en":"cat"}},{"confidence":42.0,"tag":{"en":"pet"}}]},"status":{"type":"success"}}`))
		case "https://example.com/b.jpg":
			_, _ = w.Write([]byte(`{"result":{"tags":[{"confidence":88.5,"tag":{"en":"dog"}}]},"status":{"type":"success"}}`))
		default:
			t.Errorf("unexpected image_url=%q", img)
			_, _ = w.Write([]byte(`{"result":{"tags":[]},"status":{"type":"success"}}`))
		}
	}))
	defer srv.Close()

	c := imagga.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":   srv.URL,
			"image_urls": "https://example.com/a.jpg,https://example.com/b.jpg",
		},
		Secrets: map[string]string{"api_key": "acc_123", "api_secret": "sec_456"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "tags", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("acc_123:sec_456"))
	if sawAuth != wantAuth {
		t.Fatalf("Authorization = %q, want %q", sawAuth, wantAuth)
	}
	// Both images must have been requested (two-request "pagination").
	if !seenImages["https://example.com/a.jpg"] || !seenImages["https://example.com/b.jpg"] {
		t.Fatalf("did not request both images: %+v", seenImages)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 + 1 tags across 2 images)", len(got))
	}
	for _, rec := range got {
		if rec["tag"] == nil || rec["image_url"] == nil || rec["confidence"] == nil {
			t.Fatalf("record missing tag/image_url/confidence: %+v", rec)
		}
	}
	// Verify a mapped value flattened from tag.en.
	foundCat := false
	for _, rec := range got {
		if rec["tag"] == "cat" && rec["image_url"] == "https://example.com/a.jpg" {
			foundCat = true
		}
	}
	if !foundCat {
		t.Fatalf("expected a record with tag=cat for image a.jpg, got %+v", got)
	}
}

// TestCheckBasicAuth confirms Check hits the usage endpoint with Basic auth.
func TestCheckBasicAuth(t *testing.T) {
	var sawAuth string
	var sawPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		sawPath = r.URL.Path
		_, _ = w.Write([]byte(`{"result":{"monthly_processed":10},"status":{"type":"success"}}`))
	}))
	defer srv.Close()

	c := imagga.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "acc_123", "api_secret": "sec_456"},
	}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check: %v", err)
	}
	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("acc_123:sec_456"))
	if sawAuth != wantAuth {
		t.Fatalf("Authorization = %q, want %q", sawAuth, wantAuth)
	}
	if !strings.HasSuffix(sawPath, "/usage") {
		t.Fatalf("Check path = %q, want .../usage", sawPath)
	}
}

// TestFixtureModeNoNetwork confirms fixture mode emits deterministic records with
// no network access so credential-free conformance passes.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := imagga.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture): %v", err)
	}
	for _, stream := range []string{"tags", "categories", "colors", "faces_detections", "usage"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("Read(fixture, %s): %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture stream %s emitted no records", stream)
		}
	}
}

// TestCatalogStreams confirms the published catalog exposes the core streams.
func TestCatalogStreams(t *testing.T) {
	c := imagga.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"tags": false, "categories": false, "colors": false, "faces_detections": false, "usage": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
	}
	for name, found := range want {
		if !found {
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

// TestRegisteredReadOnly confirms self-registration and read-only capabilities.
func TestRegisteredReadOnly(t *testing.T) {
	_ = imagga.New() // ensure init ran
	c := imagga.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || caps.Write {
		t.Fatalf("capabilities = %+v, want Read && !Write", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("imagga"); !ok {
		t.Fatal("registry did not resolve imagga (self-registration)")
	}
}
