package ebayfulfillment_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	ebayfulfillment "polymetrics.ai/internal/connectors/ebay-fulfillment"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the eBay
// Fulfillment connector. It asserts:
//   - the OAuth2 refresh-token grant is exchanged against refresh_token_endpoint,
//   - the resulting access token is sent as a Bearer Authorization header,
//   - getOrders pagination over the "next" URL walks two pages,
//   - order records are mapped (orderId becomes id-style key).
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var sawTokenGrant string

	mux := http.NewServeMux()
	mux.HandleFunc("/identity/v1/oauth2/token", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		sawTokenGrant = r.Form.Get("grant_type")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"ACCESS_123","token_type":"Bearer","expires_in":7200}`))
	})
	mux.HandleFunc("/sell/fulfillment/v1/order", func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Query().Get("offset") == "2" {
			_, _ = w.Write([]byte(`{"orders":[{"orderId":"03-00003","creationDate":"2026-02-03T00:00:00.000Z","orderFulfillmentStatus":"FULFILLED"}],"total":3,"limit":2,"offset":2}`))
			return
		}
		// Build an absolute "next" URL that points back at this same test server
		// so the connector's Link-style follow lands on the offset=2 page above.
		next := "http://" + r.Host + "/sell/fulfillment/v1/order?limit=2&offset=2"
		_, _ = w.Write([]byte(`{"orders":[{"orderId":"03-00001","creationDate":"2026-02-01T00:00:00.000Z","orderFulfillmentStatus":"NOT_STARTED"},{"orderId":"03-00002","creationDate":"2026-02-02T00:00:00.000Z","orderFulfillmentStatus":"IN_PROGRESS"}],"total":3,"limit":2,"offset":0,"next":"` + next + `"}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := ebayfulfillment.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"api_host":               srv.URL,
			"refresh_token_endpoint": srv.URL + "/identity/v1/oauth2/token",
			"page_size":              "2",
		},
		Secrets: map[string]string{
			"refresh_token": "REFRESH_ABC",
			"password":      "client_secret_xyz",
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
	if sawTokenGrant != "refresh_token" {
		t.Fatalf("token grant_type = %q, want refresh_token", sawTokenGrant)
	}
	if sawAuth != "Bearer ACCESS_123" {
		t.Fatalf("Authorization = %q, want Bearer ACCESS_123", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["order_id"] == nil {
			t.Fatalf("record missing order_id: %+v", rec)
		}
	}
}

func TestFixtureModeNoNetwork(t *testing.T) {
	c := ebayfulfillment.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "orders", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("fixture records = %d, want 2", len(got))
	}
	if got[0]["order_id"] == nil {
		t.Fatalf("fixture record missing order_id: %+v", got[0])
	}
}

func TestCatalogStreams(t *testing.T) {
	c := ebayfulfillment.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "ebay-fulfillment" {
		t.Fatalf("catalog connector = %q, want ebay-fulfillment", cat.Connector)
	}
	want := map[string]bool{"orders": false, "order_line_items": false, "shipping_fulfillments": false, "payment_disputes": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 {
			t.Errorf("stream %q missing primary key", s.Name)
		}
	}
	for name, seen := range want {
		if !seen {
			t.Errorf("catalog missing expected stream %q", name)
		}
	}
}

func TestBaseURLSSRFValidation(t *testing.T) {
	c := ebayfulfillment.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"api_host": "ftp://evil"},
		Secrets: map[string]string{"refresh_token": "x", "password": "y"},
	}
	err := c.Check(context.Background(), cfg)
	if err == nil || !strings.Contains(err.Error(), "http") {
		t.Fatalf("Check with bad scheme = %v, want scheme error", err)
	}
}

func TestRegisteredReadOnly(t *testing.T) {
	_ = ebayfulfillment.New() // ensure init ran
	c := ebayfulfillment.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("ebay-fulfillment"); !ok {
		t.Fatal("registry did not resolve ebay-fulfillment (self-registration)")
	}
}
