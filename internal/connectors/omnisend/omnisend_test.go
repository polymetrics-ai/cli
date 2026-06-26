package omnisend_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/omnisend"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Omnisend
// connector: X-API-KEY auth, Omnisend cursor pagination (paging.next is a full
// URL), record extraction from the per-stream field path, and record mapping.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAPIKey string
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAPIKey = r.Header.Get("X-API-KEY")
		if r.URL.Path != "/contacts" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		// Page 1 carries a full-URL next pointer back to this server; page 2
		// (identified by the cursor query param) terminates with paging.next null.
		if r.URL.Query().Get("cursor") == "" {
			next := srv.URL + "/contacts?cursor=abc&limit=100"
			_, _ = w.Write([]byte(`{"contacts":[{"contactID":"c_1","email":"a@example.com","status":"subscribed"},{"contactID":"c_2","email":"b@example.com","status":"subscribed"}],"paging":{"next":"` + next + `"}}`))
			return
		}
		_, _ = w.Write([]byte(`{"contacts":[{"contactID":"c_3","email":"c@example.com","status":"unsubscribed"}],"paging":{"next":null}}`))
	}))
	defer srv.Close()

	c := omnisend.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "omni_test_key"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "contacts", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAPIKey != "omni_test_key" {
		t.Fatalf("X-API-KEY = %q, want omni_test_key", sawAPIKey)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["contactID"] == nil {
			t.Fatalf("record missing contactID: %+v", rec)
		}
	}
	if got[0]["contactID"] != "c_1" || got[2]["contactID"] != "c_3" {
		t.Fatalf("unexpected record order/ids: %+v", got)
	}
}

// TestCampaignsFieldPath verifies the campaigns stream extracts from the
// singular "campaign" array (a quirk of the Omnisend API).
func TestCampaignsFieldPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/campaigns" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"campaign":[{"campaignID":"camp_1","name":"Spring","status":"sent"}],"paging":{"next":null}}`))
	}))
	defer srv.Close()

	c := omnisend.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "k"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "campaigns", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read campaigns: %v", err)
	}
	if len(got) != 1 || got[0]["campaignID"] != "camp_1" {
		t.Fatalf("campaigns = %+v, want one record camp_1", got)
	}
}

// TestFixtureModeNoNetwork verifies that fixture mode emits deterministic
// records with no network access (conformance runs without creds).
func TestFixtureModeNoNetwork(t *testing.T) {
	c := omnisend.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"contacts", "campaigns", "carts", "orders", "products"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read %s: %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture Read %s emitted no records", stream)
		}
		for _, rec := range got {
			if len(rec) == 0 {
				t.Fatalf("fixture %s emitted empty record", stream)
			}
		}
	}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

func TestCheckRequiresAPIKey(t *testing.T) {
	c := omnisend.New()
	err := c.Check(context.Background(), connectors.RuntimeConfig{})
	if err == nil {
		t.Fatal("Check with no api_key should error")
	}
}

func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := omnisend.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example.com"},
		Secrets: map[string]string{"api_key": "k"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "contacts", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "base_url") {
		t.Fatalf("expected base_url scheme rejection, got %v", err)
	}
}

func TestCatalogAndMetadata(t *testing.T) {
	c := omnisend.New()
	md := c.Metadata()
	if md.Name != "omnisend" {
		t.Fatalf("Name = %q, want omnisend", md.Name)
	}
	if !md.Capabilities.Read || !md.Capabilities.Catalog || !md.Capabilities.Check {
		t.Fatalf("capabilities = %+v, want Read/Catalog/Check", md.Capabilities)
	}
	if md.Capabilities.Write {
		t.Fatalf("omnisend is read-only; Write should be false")
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]string{
		"contacts":  "contactID",
		"campaigns": "campaignID",
		"carts":     "cartID",
		"orders":    "orderID",
		"products":  "productID",
	}
	if len(cat.Streams) != len(want) {
		t.Fatalf("streams = %d, want %d", len(cat.Streams), len(want))
	}
	for _, s := range cat.Streams {
		pk, ok := want[s.Name]
		if !ok {
			t.Fatalf("unexpected stream %q", s.Name)
		}
		if len(s.PrimaryKey) != 1 || s.PrimaryKey[0] != pk {
			t.Fatalf("stream %q primary key = %v, want [%s]", s.Name, s.PrimaryKey, pk)
		}
	}
}

func TestRegisteredInRegistry(t *testing.T) {
	_ = omnisend.New() // ensure init ran
	r := connectors.NewRegistry()
	got, ok := r.Get("omnisend")
	if !ok {
		t.Fatal("registry did not resolve omnisend (self-registration)")
	}
	if got.Name() != "omnisend" {
		t.Fatalf("resolved connector Name = %q, want omnisend", got.Name())
	}
}
