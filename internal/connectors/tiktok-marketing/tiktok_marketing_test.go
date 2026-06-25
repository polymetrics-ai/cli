package tiktokmarketing_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	tiktokmarketing "polymetrics.ai/internal/connectors/tiktok-marketing"
)

// TestReadPaginatesAndAuthenticates is the red-first test: it asserts the custom
// Access-Token auth header, TikTok page/page_info pagination across 2 pages over
// data.list[], and record mapping for the campaigns stream.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawToken string
	var sawAdvertiser string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawToken = r.Header.Get("Access-Token")
		sawAdvertiser = r.URL.Query().Get("advertiser_id")
		if r.URL.Path != "/campaign/get/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Query().Get("page") {
		case "", "1":
			_, _ = w.Write([]byte(`{"code":0,"message":"OK","data":{"list":[{"campaign_id":"c1","campaign_name":"Alpha"},{"campaign_id":"c2","campaign_name":"Beta"}],"page_info":{"page":1,"page_size":2,"total_number":3,"total_page":2}}}`))
		case "2":
			_, _ = w.Write([]byte(`{"code":0,"message":"OK","data":{"list":[{"campaign_id":"c3","campaign_name":"Gamma"}],"page_info":{"page":2,"page_size":2,"total_number":3,"total_page":2}}}`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
			_, _ = w.Write([]byte(`{"code":0,"message":"OK","data":{"list":[],"page_info":{"page":3,"page_size":2,"total_number":3,"total_page":2}}}`))
		}
	}))
	defer srv.Close()

	c := tiktokmarketing.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "advertiser_id": "adv_123"},
		Secrets: map[string]string{"credentials.access_token": "tok_secret"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "campaigns", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawToken != "tok_secret" {
		t.Fatalf("Access-Token = %q, want tok_secret", sawToken)
	}
	if sawAdvertiser != "adv_123" {
		t.Fatalf("advertiser_id = %q, want adv_123", sawAdvertiser)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["campaign_id"] == nil || rec["campaign_name"] == nil {
			t.Fatalf("record missing campaign_id/campaign_name: %+v", rec)
		}
	}
}

// TestFixtureModeReadsWithoutNetwork confirms fixture mode emits deterministic
// records without any network access (no base_url, no creds needed).
func TestFixtureModeReadsWithoutNetwork(t *testing.T) {
	c := tiktokmarketing.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "ads", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) == 0 {
		t.Fatal("fixture Read emitted no records")
	}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

// TestCheckRequiresAccessToken confirms non-fixture Check rejects a missing token.
func TestCheckRequiresAccessToken(t *testing.T) {
	c := tiktokmarketing.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"advertiser_id": "adv_1"}}
	if err := c.Check(context.Background(), cfg); err == nil {
		t.Fatal("Check with no access token should fail")
	}
}

// TestRegisteredReadOnly confirms self-registration via the registry and the
// read-only capability set.
func TestRegisteredReadOnly(t *testing.T) {
	c := tiktokmarketing.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || caps.Write {
		t.Fatalf("capabilities = %+v, want Read && !Write", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("tiktok-marketing"); !ok {
		t.Fatal("registry did not resolve tiktok-marketing (self-registration)")
	}
}

func TestCatalogStreams(t *testing.T) {
	c := tiktokmarketing.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("streams = %d, want >= 3", len(cat.Streams))
	}
	want := map[string]bool{"advertisers": false, "campaigns": false, "adgroups": false, "ads": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q has no primary key", s.Name)
		}
	}
	for name, found := range want {
		if !found {
			t.Fatalf("catalog missing expected stream %q", name)
		}
	}
}
