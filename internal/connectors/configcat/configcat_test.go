package configcat_test

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/configcat"
)

// TestReadProductsAuthenticates is the red-first test: HTTP Basic auth on the
// Authorization header and root-array record mapping for the products stream.
func TestReadProductsAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/v1/products" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`[
			{"productId":"p1","name":"Alpha","description":"first","order":0},
			{"productId":"p2","name":"Beta","description":"second","order":1}
		]`))
	}))
	defer srv.Close()

	c := configcat.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"password": "pw_secret", "username": "user_id"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "products", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	want := "Basic " + base64.StdEncoding.EncodeToString([]byte("user_id:pw_secret"))
	if sawAuth != want {
		t.Fatalf("Authorization = %q, want %q", sawAuth, want)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2", len(got))
	}
	if got[0]["product_id"] != "p1" || got[1]["product_id"] != "p2" {
		t.Fatalf("unexpected product_id mapping: %+v", got)
	}
}

// TestReadConfigsFansOutAcrossProducts asserts the nested configs stream fans
// out across every product (two product pages of work), proving the connector
// walks /v1/products then /v1/products/{id}/configs for each.
func TestReadConfigsFansOutAcrossProducts(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/products":
			_, _ = w.Write([]byte(`[{"productId":"p1","name":"Alpha"},{"productId":"p2","name":"Beta"}]`))
		case "/v1/products/p1/configs":
			_, _ = w.Write([]byte(`[{"configId":"c1","name":"Main","order":0}]`))
		case "/v1/products/p2/configs":
			_, _ = w.Write([]byte(`[{"configId":"c2","name":"Mobile","order":0},{"configId":"c3","name":"Web","order":1}]`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := configcat.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"password": "pw", "username": "u"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "configs", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("configs = %d, want 3 (1 from p1 + 2 from p2)", len(got))
	}
	// fan-out should annotate each config with its owning product id.
	for _, rec := range got {
		if rec["config_id"] == nil {
			t.Fatalf("config record missing config_id: %+v", rec)
		}
		if rec["product_id"] == nil {
			t.Fatalf("config record missing product_id from fan-out: %+v", rec)
		}
	}
}

func TestFixtureModeNeedsNoNetwork(t *testing.T) {
	c := configcat.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"organizations", "products", "configs", "environments", "tags"} {
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
}

func TestRegisteredReadOnly(t *testing.T) {
	_ = configcat.New()
	caps := configcat.New().Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("configcat is read-only; Write should be false, got %+v", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("configcat"); !ok {
		t.Fatal("registry did not resolve configcat (self-registration)")
	}
}

func TestCheckFixtureMode(t *testing.T) {
	c := configcat.New()
	if err := c.Check(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}
