// Package stigg's hooks_test.go proves hooks.go's StreamHook ports legacy
// internal/connectors/stigg/stigg.go's GraphQL query construction and
// data.<field> record extraction byte-for-byte.
package stigg

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/engine"
)

type graphQLRequest struct {
	Query string `json:"query"`
}

func newRuntime(baseURL string) *engine.Runtime {
	return &engine.Runtime{Requester: &connsdk.Requester{BaseURL: baseURL}}
}

// --- registration ---

func TestInit_RegistersHooks(t *testing.T) {
	h := engine.HooksFor("stigg")
	if h == nil {
		t.Fatal("engine.HooksFor(\"stigg\") = nil, want a registered hook set (init() must call engine.RegisterHooks)")
	}
	if h.ConnectorName() != "stigg" {
		t.Fatalf("ConnectorName() = %q, want stigg", h.ConnectorName())
	}
	if _, ok := h.(engine.StreamHook); !ok {
		t.Fatal("registered hooks do not implement StreamHook")
	}
	if _, ok := h.(engine.AuthHook); ok {
		t.Fatal("registered hooks implement AuthHook, want none (auth is fully declarative via bearer mode)")
	}
}

// --- ReadStream dispatch ---

func TestReadStream_UnknownStreamFallsBackToDeclarative(t *testing.T) {
	h := Hooks{}
	handled, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "not_a_real_stream"}, connectors.ReadRequest{}, newRuntime("http://example.invalid"), func(connectors.Record) error { return nil })
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if handled {
		t.Fatal("handled = true for an unrecognized stream name, want false (declarative fallback)")
	}
}

func TestReadStream_EmptyStreamNameDefaultsToProducts(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body graphQLRequest
		_ = json.NewDecoder(r.Body).Decode(&body)
		if !strings.Contains(body.Query, "products") {
			t.Errorf("query did not target products: %q", body.Query)
		}
		_, _ = w.Write([]byte(`{"data":{"products":[]}}`))
	}))
	defer srv.Close()

	h := Hooks{}
	handled, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: ""}, connectors.ReadRequest{}, newRuntime(srv.URL), func(connectors.Record) error { return nil })
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if !handled {
		t.Fatal("handled = false for empty stream name, want true (defaults to products)")
	}
}

// --- per-stream GraphQL query + record extraction ---

func TestReadStream_ProductsQueryAndRecords(t *testing.T) {
	var sawMethod string
	var sawPath string
	var sawQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawMethod = r.Method
		sawPath = r.URL.Path
		var body graphQLRequest
		_ = json.NewDecoder(r.Body).Decode(&body)
		sawQuery = body.Query
		_, _ = w.Write([]byte(`{"data":{"products":[{"id":"prod1","refId":"starter","displayName":"Starter","status":"ACTIVE"},{"id":"prod2","refId":"pro","displayName":"Pro","status":"ACTIVE"}]}}`))
	}))
	defer srv.Close()

	var got []connectors.Record
	h := Hooks{}
	handled, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "products"}, connectors.ReadRequest{}, newRuntime(srv.URL), func(r connectors.Record) error {
		got = append(got, r)
		return nil
	})
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if !handled {
		t.Fatal("handled = false, want true")
	}
	if sawMethod != http.MethodPost {
		t.Fatalf("method = %q, want POST", sawMethod)
	}
	if sawPath != "/graphql" {
		t.Fatalf("path = %q, want /graphql", sawPath)
	}
	if !strings.Contains(sawQuery, "PolymetricsProducts") || !strings.Contains(sawQuery, "id refId displayName status") {
		t.Fatalf("query = %q, want the exact ported PolymetricsProducts query", sawQuery)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2", len(got))
	}
	if got[0]["id"] != "prod1" || got[0]["refId"] != "starter" || got[0]["displayName"] != "Starter" || got[0]["status"] != "ACTIVE" {
		t.Fatalf("record 0 = %+v, want full field mapping", got[0])
	}
}

func TestReadStream_PlansQueryAndRecords(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body graphQLRequest
		_ = json.NewDecoder(r.Body).Decode(&body)
		if !strings.Contains(body.Query, "PolymetricsPlans") {
			t.Errorf("query = %q, want PolymetricsPlans", body.Query)
		}
		_, _ = w.Write([]byte(`{"data":{"plans":[{"id":"plan1","refId":"starter-monthly","displayName":"Starter Monthly","status":"ACTIVE"}]}}`))
	}))
	defer srv.Close()

	var got []connectors.Record
	h := Hooks{}
	_, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "plans"}, connectors.ReadRequest{}, newRuntime(srv.URL), func(r connectors.Record) error {
		got = append(got, r)
		return nil
	})
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if len(got) != 1 || got[0]["id"] != "plan1" {
		t.Fatalf("records = %+v, want one plan record", got)
	}
}

func TestReadStream_CustomersQueryAndRecords(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body graphQLRequest
		_ = json.NewDecoder(r.Body).Decode(&body)
		if !strings.Contains(body.Query, "PolymetricsCustomers") {
			t.Errorf("query = %q, want PolymetricsCustomers", body.Query)
		}
		_, _ = w.Write([]byte(`{"data":{"customers":[{"id":"cust1","refId":"fixture-customer","displayName":"Fixture Customer","status":"ACTIVE"}]}}`))
	}))
	defer srv.Close()

	var got []connectors.Record
	h := Hooks{}
	_, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "customers"}, connectors.ReadRequest{}, newRuntime(srv.URL), func(r connectors.Record) error {
		got = append(got, r)
		return nil
	})
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if len(got) != 1 || got[0]["id"] != "cust1" {
		t.Fatalf("records = %+v, want one customer record", got)
	}
}

func TestReadStream_SubscriptionsQueryAndRecords(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body graphQLRequest
		_ = json.NewDecoder(r.Body).Decode(&body)
		if !strings.Contains(body.Query, "PolymetricsSubscriptions") || !strings.Contains(body.Query, "customerId") {
			t.Errorf("query = %q, want PolymetricsSubscriptions with customerId field", body.Query)
		}
		_, _ = w.Write([]byte(`{"data":{"subscriptions":[{"id":"sub1","refId":"fixture-sub","customerId":"cust1","status":"ACTIVE"}]}}`))
	}))
	defer srv.Close()

	var got []connectors.Record
	h := Hooks{}
	_, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "subscriptions"}, connectors.ReadRequest{}, newRuntime(srv.URL), func(r connectors.Record) error {
		got = append(got, r)
		return nil
	})
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if len(got) != 1 || got[0]["customerId"] != "cust1" {
		t.Fatalf("records = %+v, want one subscription record with customerId", got)
	}
}

// --- no GraphQL-errors detection, matching legacy exactly ---

func TestReadStream_GraphQLErrorsEnvelopeIsNotDetected(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"errors":[{"message":"boom"}]}`))
	}))
	defer srv.Close()

	h := Hooks{}
	handled, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "products"}, connectors.ReadRequest{}, newRuntime(srv.URL), func(connectors.Record) error { return nil })
	if err != nil {
		t.Fatalf("ReadStream: %v, want no error (legacy stigg.go never inspects the response for a GraphQL errors array)", err)
	}
	if !handled {
		t.Fatal("handled = false, want true")
	}
}

// --- HTTP error path ---

func TestReadStream_NonSuccessResponseIsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	h := Hooks{}
	_, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "products"}, connectors.ReadRequest{}, newRuntime(srv.URL), func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("ReadStream error = nil, want an error for a 500 response")
	}
}

// --- ctx cancellation ---

func TestReadStream_HonorsContextCancellation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":{"products":[]}}`))
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	h := Hooks{}
	_, err := h.ReadStream(ctx, engine.StreamSpec{Name: "products"}, connectors.ReadRequest{}, newRuntime(srv.URL), func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("ReadStream(cancelled ctx) error = nil, want a cancellation error")
	}
}
