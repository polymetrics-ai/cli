package judgemereviews_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	judgemereviews "polymetrics.ai/internal/connectors/judge-me-reviews"
)

// TestReadPaginatesAndAuthenticates is the red-first test: Judge.me uses
// api_token + shop_domain query-param auth and page-number pagination over the
// resource-keyed array (e.g. {"reviews":[...]}). It asserts the auth params,
// two-page pagination, and record mapping.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawToken, sawShop string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/reviews" {
			http.NotFound(w, r)
			return
		}
		q := r.URL.Query()
		sawToken = q.Get("api_token")
		sawShop = q.Get("shop_domain")
		switch q.Get("page") {
		case "", "1":
			_, _ = w.Write([]byte(`{"reviews":[{"id":1,"rating":5,"title":"Great","reviewer":{"name":"Ada","email":"ada@example.com"},"created_at":"2026-01-01T00:00:00Z"},{"id":2,"rating":4,"title":"Good","reviewer":{"name":"Grace"},"created_at":"2026-01-02T00:00:00Z"}]}`))
		case "2":
			_, _ = w.Write([]byte(`{"reviews":[{"id":3,"rating":3,"title":"Okay","reviewer":{"name":"Kat"},"created_at":"2026-01-03T00:00:00Z"}]}`))
		case "3":
			_, _ = w.Write([]byte(`{"reviews":[]}`))
		default:
			t.Errorf("unexpected page=%q", q.Get("page"))
			_, _ = w.Write([]byte(`{"reviews":[]}`))
		}
	}))
	defer srv.Close()

	c := judgemereviews.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":    srv.URL,
			"shop_domain": "example.myshopify.com",
			"page_size":   "2",
		},
		Secrets: map[string]string{"api_key": "tok_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "reviews", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawToken != "tok_123" {
		t.Fatalf("api_token = %q, want tok_123", sawToken)
	}
	if sawShop != "example.myshopify.com" {
		t.Fatalf("shop_domain = %q, want example.myshopify.com", sawShop)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages of data)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["rating"] == nil {
			t.Fatalf("record missing id/rating: %+v", rec)
		}
	}
	// reviewer.name should be flattened onto the record.
	if got[0]["reviewer_name"] != "Ada" {
		t.Fatalf("reviewer_name = %v, want Ada", got[0]["reviewer_name"])
	}
}

func TestFixtureModeNoNetwork(t *testing.T) {
	c := judgemereviews.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "reviews", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("fixture records = %d, want 2", len(got))
	}
	if got[0]["id"] == nil {
		t.Fatalf("fixture record missing id: %+v", got[0])
	}

	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

func TestCheckRequiresSecretAndShop(t *testing.T) {
	c := judgemereviews.New()
	// Missing api_key.
	err := c.Check(context.Background(), connectors.RuntimeConfig{
		Config: map[string]string{"shop_domain": "example.myshopify.com"},
	})
	if err == nil {
		t.Fatal("Check should fail without api_key secret")
	}
	// Missing shop_domain.
	err = c.Check(context.Background(), connectors.RuntimeConfig{
		Secrets: map[string]string{"api_key": "tok"},
	})
	if err == nil {
		t.Fatal("Check should fail without shop_domain")
	}
}

func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := judgemereviews.New()
	err := c.Check(context.Background(), connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil", "shop_domain": "x"},
		Secrets: map[string]string{"api_key": "tok"},
	})
	if err == nil {
		t.Fatal("Check should reject non-http(s) base_url")
	}
}

func TestCatalogAndMetadata(t *testing.T) {
	c := judgemereviews.New()
	if c.Name() != "judge-me-reviews" {
		t.Fatalf("Name = %q, want judge-me-reviews", c.Name())
	}
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatal("judge-me-reviews is read-only; Write should be false")
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("streams = %d, want >= 3", len(cat.Streams))
	}
	found := false
	for _, s := range cat.Streams {
		if s.Name == "reviews" {
			found = true
			if len(s.PrimaryKey) == 0 {
				t.Fatal("reviews stream must have a primary key")
			}
		}
	}
	if !found {
		t.Fatal("catalog missing reviews stream")
	}
}

func TestRegistryResolves(t *testing.T) {
	_ = judgemereviews.New() // ensure init() ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("judge-me-reviews"); !ok {
		t.Fatal("registry did not resolve judge-me-reviews (self-registration)")
	}
}
