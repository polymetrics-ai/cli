package googleads_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	googleads "polymetrics.ai/internal/connectors/google-ads"
)

func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth, sawDeveloperToken, sawLoginCustomerID string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		sawDeveloperToken = r.Header.Get("developer-token")
		sawLoginCustomerID = r.Header.Get("login-customer-id")
		if r.Method != http.MethodPost || r.URL.Path != "/v24/customers/1234567890/googleAds:search" {
			http.NotFound(w, r)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		switch body["pageToken"] {
		case nil, "":
			_, _ = w.Write([]byte(`{"results":[{"campaign":{"id":"111","name":"Brand","status":"ENABLED","resourceName":"customers/1234567890/campaigns/111"}}],"nextPageToken":"next-page"}`))
		case "next-page":
			_, _ = w.Write([]byte(`{"results":[{"campaign":{"id":"222","name":"Demand","status":"PAUSED","resourceName":"customers/1234567890/campaigns/222"}}]}`))
		default:
			t.Fatalf("unexpected pageToken=%v", body["pageToken"])
		}
	}))
	defer srv.Close()

	c := googleads.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":          srv.URL + "/v24",
			"customer_id":       "1234567890",
			"login_customer_id": "9999999999",
			"page_size":         "1",
		},
		Secrets: map[string]string{"access_token": "test-access-token", "developer_token": "test-developer-token"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "campaigns", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer test-access-token" || sawDeveloperToken != "test-developer-token" || sawLoginCustomerID != "9999999999" {
		t.Fatalf("auth headers not applied: auth=%q developer=%q login=%q", sawAuth, sawDeveloperToken, sawLoginCustomerID)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2", len(got))
	}
	if got[0]["id"] != "111" || got[0]["name"] != "Brand" || got[0]["status"] != "ENABLED" {
		t.Fatalf("first record not mapped: %+v", got[0])
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := googleads.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
	cat, err := c.Catalog(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "google-ads" || len(cat.Streams) == 0 {
		t.Fatalf("catalog = %+v, want google-ads streams", cat)
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "campaigns", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read fixture: %v", err)
	}
	if len(got) == 0 || got[0]["id"] == nil {
		t.Fatalf("fixture records not mapped: %+v", got)
	}
	if _, ok := connectors.NewRegistry().Get("google-ads"); !ok {
		t.Fatal("registry did not resolve google-ads")
	}
	_, err = c.Write(context.Background(), connectors.WriteRequest{Config: cfg}, []connectors.Record{{"id": "1"}})
	if !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
}
