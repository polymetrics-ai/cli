package chargedesk_test

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics/internal/connectors"
	"polymetrics/internal/connectors/chargedesk"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the ChargeDesk
// connector: HTTP Basic auth (secret API key as username, blank password),
// ChargeDesk offset/count pagination over data[], and record mapping. The server
// returns a full page (count==2) then a short page (len<count) to terminate.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/charges" {
			http.NotFound(w, r)
			return
		}
		if got := r.URL.Query().Get("count"); got != "2" {
			t.Errorf("count = %q, want 2", got)
		}
		switch r.URL.Query().Get("offset") {
		case "", "0":
			_, _ = w.Write([]byte(`{"object":"list","count":2,"offset":0,"data":[{"charge_id":"chg_1","occurred":1700000000,"amount":"10.00","currency":"usd","status":"paid"},{"charge_id":"chg_2","occurred":1700000100,"amount":"20.00","currency":"usd","status":"paid"}]}`))
		case "2":
			_, _ = w.Write([]byte(`{"object":"list","count":2,"offset":2,"data":[{"charge_id":"chg_3","occurred":1700000200,"amount":"30.00","currency":"usd","status":"refunded"}]}`))
		default:
			t.Errorf("unexpected offset=%q", r.URL.Query().Get("offset"))
			_, _ = w.Write([]byte(`{"count":2,"offset":0,"data":[]}`))
		}
	}))
	defer srv.Close()

	c := chargedesk.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"password": "sk_test_chargedesk"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "charges", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("sk_test_chargedesk:"))
	if sawAuth != wantAuth {
		t.Fatalf("Authorization = %q, want %q", sawAuth, wantAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["charge_id"] == nil || rec["occurred"] == nil {
			t.Fatalf("record missing charge_id/occurred: %+v", rec)
		}
	}
	if got[0]["charge_id"] != "chg_1" || got[2]["status"] != "refunded" {
		t.Fatalf("unexpected mapped records: %+v", got)
	}
}

// TestReadUsernameOverride confirms that when a username config is supplied,
// it is used as the Basic auth username (matching ChargeDesk's documented
// secret-key-as-username scheme) with the password secret as the password.
func TestReadUsernameOverride(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		_, _ = w.Write([]byte(`{"count":20,"offset":0,"data":[{"customer_id":"cus_1","occurred":1700000000}]}`))
	}))
	defer srv.Close()

	c := chargedesk.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "username": "api_user"},
		Secrets: map[string]string{"password": "secret_pw"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "customers", Config: cfg}, func(connectors.Record) error { return nil })
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	want := "Basic " + base64.StdEncoding.EncodeToString([]byte("api_user:secret_pw"))
	if sawAuth != want {
		t.Fatalf("Authorization = %q, want %q", sawAuth, want)
	}
}

func TestFixtureModeReadsWithoutNetwork(t *testing.T) {
	c := chargedesk.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	for _, stream := range []string{"charges", "customers", "subscriptions", "products"} {
		got = got[:0]
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if len(got) != 2 {
			t.Fatalf("fixture Read(%s) = %d records, want 2", stream, len(got))
		}
	}
}

func TestCheckFixtureMode(t *testing.T) {
	c := chargedesk.New()
	if err := c.Check(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

func TestCatalogStreams(t *testing.T) {
	c := chargedesk.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "chargedesk" {
		t.Fatalf("catalog connector = %q, want chargedesk", cat.Connector)
	}
	want := map[string]bool{"charges": true, "customers": true, "subscriptions": true, "products": true}
	for _, s := range cat.Streams {
		delete(want, s.Name)
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q missing primary key", s.Name)
		}
	}
	if len(want) != 0 {
		t.Fatalf("catalog missing streams: %v", want)
	}
}

func TestMetadataReadOnly(t *testing.T) {
	caps := chargedesk.New().Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("chargedesk is read-only; Write must be false, got %+v", caps)
	}
}

func TestRegisteredInRegistry(t *testing.T) {
	_ = chargedesk.New() // ensure init() ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("chargedesk"); !ok {
		t.Fatal("registry did not resolve chargedesk (self-registration)")
	}
}

// TestBaseURLRejectsBadScheme guards the SSRF check on base_url overrides.
func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := chargedesk.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "file:///etc/passwd"},
		Secrets: map[string]string{"password": "x"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "charges", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read with file:// base_url should be rejected")
	}
}
