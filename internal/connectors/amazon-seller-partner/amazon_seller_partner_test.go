package amazonsellerpartner_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	amazonsellerpartner "polymetrics.ai/internal/connectors/amazon-seller-partner"
)

// TestReadPaginatesAndAuthenticates is the red-first test: it asserts the LWA
// token exchange (refresh_token grant -> access_token), the x-amz-access-token
// header on the data request, NextToken pagination over payload.Orders across
// two pages, and record mapping.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var (
		sawAccessToken string
		sawGrantType   string
		sawRefresh     string
		sawMarketplace string
	)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/o2/token":
			_ = r.ParseForm()
			sawGrantType = r.PostForm.Get("grant_type")
			sawRefresh = r.PostForm.Get("refresh_token")
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"Atza|access-123","token_type":"bearer","expires_in":3600}`))
		case "/orders/v0/orders":
			sawAccessToken = r.Header.Get("x-amz-access-token")
			if sawMarketplace == "" {
				sawMarketplace = r.URL.Query().Get("MarketplaceIds")
			}
			switch r.URL.Query().Get("NextToken") {
			case "":
				_, _ = w.Write([]byte(`{"payload":{"Orders":[{"AmazonOrderId":"111-1","LastUpdateDate":"2026-01-01T00:00:00Z","OrderStatus":"Shipped"},{"AmazonOrderId":"111-2","LastUpdateDate":"2026-01-02T00:00:00Z","OrderStatus":"Pending"}],"NextToken":"PAGE2"}}`))
			case "PAGE2":
				_, _ = w.Write([]byte(`{"payload":{"Orders":[{"AmazonOrderId":"111-3","LastUpdateDate":"2026-01-03T00:00:00Z","OrderStatus":"Shipped"}]}}`))
			default:
				t.Errorf("unexpected NextToken=%q", r.URL.Query().Get("NextToken"))
				_, _ = w.Write([]byte(`{"payload":{"Orders":[]}}`))
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := amazonsellerpartner.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":       srv.URL,
			"lwa_token_url":  srv.URL + "/auth/o2/token",
			"marketplace_id": "ATVPDKIKX0DER",
		},
		Secrets: map[string]string{
			"lwa_app_id":        "amzn1.application-oa2-client.abc",
			"lwa_client_secret": "shhh",
			"refresh_token":     "Atzr|refresh-xyz",
		},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "orders", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if sawGrantType != "refresh_token" {
		t.Fatalf("token grant_type = %q, want refresh_token", sawGrantType)
	}
	if sawRefresh != "Atzr|refresh-xyz" {
		t.Fatalf("token refresh_token = %q, want the configured refresh token", sawRefresh)
	}
	if sawAccessToken != "Atza|access-123" {
		t.Fatalf("x-amz-access-token = %q, want Atza|access-123", sawAccessToken)
	}
	if sawMarketplace != "ATVPDKIKX0DER" {
		t.Fatalf("MarketplaceIds = %q, want ATVPDKIKX0DER", sawMarketplace)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["AmazonOrderId"] == nil {
			t.Fatalf("record missing AmazonOrderId: %+v", rec)
		}
	}
}

// TestInventoryPaginationTokenPath asserts the FBA inventory stream uses the
// distinct pagination.nextToken body path (vs payload.NextToken for orders).
func TestInventoryPaginationTokenPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/o2/token":
			_, _ = w.Write([]byte(`{"access_token":"Atza|inv","token_type":"bearer","expires_in":3600}`))
		case "/fba/inventory/v1/summaries":
			switch r.URL.Query().Get("nextToken") {
			case "":
				_, _ = w.Write([]byte(`{"pagination":{"nextToken":"NEXT"},"payload":{"inventorySummaries":[{"sellerSku":"SKU-1","asin":"B001","totalQuantity":5,"lastUpdatedTime":"2026-01-01T00:00:00Z"}]}}`))
			case "NEXT":
				_, _ = w.Write([]byte(`{"payload":{"inventorySummaries":[{"sellerSku":"SKU-2","asin":"B002","totalQuantity":9,"lastUpdatedTime":"2026-01-02T00:00:00Z"}]}}`))
			default:
				t.Errorf("unexpected nextToken=%q", r.URL.Query().Get("nextToken"))
				_, _ = w.Write([]byte(`{"payload":{"inventorySummaries":[]}}`))
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := amazonsellerpartner.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":       srv.URL,
			"lwa_token_url":  srv.URL + "/auth/o2/token",
			"marketplace_id": "ATVPDKIKX0DER",
		},
		Secrets: map[string]string{
			"lwa_app_id":        "id",
			"lwa_client_secret": "secret",
			"refresh_token":     "Atzr|rt",
		},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "inventory_summaries", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read inventory: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("inventory records = %d, want 2 (2 pages)", len(got))
	}
	if got[0]["sellerSku"] != "SKU-1" || got[1]["sellerSku"] != "SKU-2" {
		t.Fatalf("unexpected inventory mapping: %+v", got)
	}
}

// TestFixtureModeNoNetwork confirms credential-free conformance: fixture mode
// emits deterministic records with no token exchange and no network call.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := amazonsellerpartner.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	for _, stream := range []string{"orders", "inventory_summaries", "financial_event_groups"} {
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
	// Check must also short-circuit in fixture mode.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

func TestCatalogAndMetadata(t *testing.T) {
	c := amazonsellerpartner.New()
	if c.Name() != "amazon-seller-partner" {
		t.Fatalf("Name() = %q, want amazon-seller-partner", c.Name())
	}
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read+Catalog+Check", caps)
	}
	if caps.Write {
		t.Fatalf("amazon-seller-partner is read-only; Write should be false")
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("catalog streams = %d, want >= 3", len(cat.Streams))
	}
}

func TestRegistryResolves(t *testing.T) {
	_ = amazonsellerpartner.New() // ensure init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("amazon-seller-partner"); !ok {
		t.Fatal("registry did not resolve amazon-seller-partner (self-registration)")
	}
}

func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := amazonsellerpartner.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{"base_url": "ftp://evil.example.com"},
		Secrets: map[string]string{
			"lwa_app_id":        "id",
			"lwa_client_secret": "secret",
			"refresh_token":     "rt",
		},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "orders", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read with non-http base_url scheme should be rejected (SSRF guard)")
	}
}
