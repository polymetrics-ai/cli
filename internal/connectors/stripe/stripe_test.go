package stripe_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/stripe"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Stripe
// connector: Bearer auth, Stripe has_more/starting_after pagination over data[],
// and record mapping. Red until internal/connectors/stripe exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/customers" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("starting_after") {
		case "":
			_, _ = w.Write([]byte(`{"object":"list","data":[{"id":"cus_1","created":1700000000},{"id":"cus_2","created":1700000100}],"has_more":true}`))
		case "cus_2":
			_, _ = w.Write([]byte(`{"object":"list","data":[{"id":"cus_3","created":1700000200}],"has_more":false}`))
		default:
			t.Errorf("unexpected starting_after=%q", r.URL.Query().Get("starting_after"))
			_, _ = w.Write([]byte(`{"object":"list","data":[],"has_more":false}`))
		}
	}))
	defer srv.Close()

	c := stripe.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"client_secret": "sk_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "customers", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer sk_test_123" {
		t.Fatalf("Authorization = %q, want Bearer sk_test_123", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["created"] == nil {
			t.Fatalf("record missing id/created: %+v", rec)
		}
	}
}

func TestWriteValidateAllowList(t *testing.T) {
	c := stripe.New()
	wv, ok := c.(connectors.WriteValidator)
	if !ok {
		t.Fatal("stripe connector must implement WriteValidator")
	}
	cfg := connectors.RuntimeConfig{Secrets: map[string]string{"client_secret": "sk_test_123"}}
	if err := wv.ValidateWrite(context.Background(), connectors.WriteRequest{Action: "create_customer", Config: cfg}, []connectors.Record{{"email": "a@example.com"}}); err != nil {
		t.Fatalf("ValidateWrite(create_customer) = %v, want nil", err)
	}
	err := wv.ValidateWrite(context.Background(), connectors.WriteRequest{Action: "delete_everything", Config: cfg}, []connectors.Record{{}})
	if err == nil || errors.Is(err, nil) {
		t.Fatal("ValidateWrite(unknown action) should be rejected")
	}
}

func TestRegisteredWithWriteCapability(t *testing.T) {
	_ = stripe.New() // ensure init ran
	c := stripe.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Write {
		t.Fatalf("capabilities = %+v, want Read && Write", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("stripe"); !ok {
		t.Fatal("registry did not resolve stripe (self-registration)")
	}
}
