package fastbill_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/fastbill"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the FastBill
// connector: HTTP Basic auth (username + api_key), the JSON SERVICE envelope,
// LIMIT/OFFSET pagination over RESPONSE.CUSTOMERS, and record mapping.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var sawServices []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		var body struct {
			Service string      `json:"SERVICE"`
			Limit   json.Number `json:"LIMIT"`
			Offset  json.Number `json:"OFFSET"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		sawServices = append(sawServices, body.Service)
		w.Header().Set("Content-Type", "application/json")
		switch body.Offset.String() {
		case "0":
			// A full page (LIMIT records) signals there may be more.
			_, _ = w.Write([]byte(`{"RESPONSE":{"CUSTOMERS":[` +
				`{"CUSTOMER_ID":"1","CUSTOMER_NUMBER":"K-1","ORGANIZATION":"Acme","EMAIL":"a@example.com"},` +
				`{"CUSTOMER_ID":"2","CUSTOMER_NUMBER":"K-2","ORGANIZATION":"Globex","EMAIL":"b@example.com"}]}}`))
		case "2":
			// A short page ends pagination.
			_, _ = w.Write([]byte(`{"RESPONSE":{"CUSTOMERS":[` +
				`{"CUSTOMER_ID":"3","CUSTOMER_NUMBER":"K-3","ORGANIZATION":"Initech","EMAIL":"c@example.com"}]}}`))
		default:
			t.Errorf("unexpected OFFSET=%q", body.Offset.String())
			_, _ = w.Write([]byte(`{"RESPONSE":{"CUSTOMERS":[]}}`))
		}
	}))
	defer srv.Close()

	c := fastbill.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":  srv.URL,
			"username":  "user@example.com",
			"page_size": "2",
		},
		Secrets: map[string]string{"api_key": "secret_key"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "customers", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("user@example.com:secret_key"))
	if sawAuth != wantAuth {
		t.Fatalf("Authorization = %q, want %q", sawAuth, wantAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, svc := range sawServices {
		if svc != "customer.get" {
			t.Fatalf("SERVICE = %q, want customer.get", svc)
		}
	}
	if got[0]["customer_id"] != "1" || got[0]["organization"] != "Acme" || got[0]["email"] != "a@example.com" {
		t.Fatalf("mapped record 0 = %+v", got[0])
	}
	if got[2]["customer_id"] != "3" {
		t.Fatalf("mapped record 2 = %+v", got[2])
	}
}

// TestReadFixtureMode confirms the credential-free fixture path emits
// deterministic records without any network call.
func TestReadFixtureMode(t *testing.T) {
	c := fastbill.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "invoices", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read fixture: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("fixture records = %d, want 2", len(got))
	}
	if got[0]["invoice_id"] == nil {
		t.Fatalf("fixture record missing invoice_id: %+v", got[0])
	}
}

func TestCheckFixtureNoNetwork(t *testing.T) {
	c := fastbill.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check fixture = %v, want nil", err)
	}
}

func TestCatalogStreams(t *testing.T) {
	c := fastbill.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"customers": false, "invoices": false, "products": false, "recurring_invoices": false, "revenues": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := fastbill.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "file:///etc/passwd", "username": "u"},
		Secrets: map[string]string{"api_key": "k"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "customers", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read with file:// base_url should be rejected (SSRF guard)")
	}
}

func TestRegisteredReadOnly(t *testing.T) {
	_ = fastbill.New() // ensure init ran
	caps := fastbill.New().Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("fastbill is read-only, Write should be false")
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("fastbill"); !ok {
		t.Fatal("registry did not resolve fastbill (self-registration)")
	}
}
