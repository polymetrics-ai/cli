package akeneo_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/akeneo"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Akeneo
// connector. It exercises the full live path against an httptest server:
//   - the OAuth2 password-grant token exchange (Basic client_id:secret header +
//     JSON {grant_type:password,username,password} body),
//   - Bearer auth on the resource request using the returned access_token,
//   - Akeneo _links.next.href pagination across two pages, and
//   - record mapping out of _embedded.items.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawTokenBasic string
	var sawTokenBody map[string]string
	var sawResourceAuth string
	pageRequests := 0

	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	defer srv.Close()

	mux.HandleFunc("/api/oauth/v1/token", func(w http.ResponseWriter, r *http.Request) {
		sawTokenBasic = r.Header.Get("Authorization")
		_ = json.NewDecoder(r.Body).Decode(&sawTokenBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"tok_abc","expires_in":3600,"token_type":"bearer","refresh_token":"ref_xyz"}`))
	})

	mux.HandleFunc("/api/rest/v1/products", func(w http.ResponseWriter, r *http.Request) {
		sawResourceAuth = r.Header.Get("Authorization")
		pageRequests++
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Query().Get("page") == "2" {
			_, _ = w.Write([]byte(`{
				"_links":{"self":{"href":"x"}},
				"_embedded":{"items":[
					{"identifier":"prod_3","enabled":true,"family":"shoes","updated":"2026-01-03T00:00:00+00:00"}
				]}
			}`))
			return
		}
		// First page advertises a next link pointing at page 2 (absolute URL).
		_, _ = w.Write([]byte(`{
			"_links":{"self":{"href":"x"},"next":{"href":"` + srv.URL + `/api/rest/v1/products?page=2&limit=2"}},
			"_embedded":{"items":[
				{"identifier":"prod_1","enabled":true,"family":"shoes","updated":"2026-01-01T00:00:00+00:00"},
				{"identifier":"prod_2","enabled":false,"family":"hats","updated":"2026-01-02T00:00:00+00:00"}
			]}
		}`))
	})

	c := akeneo.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"host":         srv.URL,
			"api_username": "api_user",
			"client_id":    "cid_123",
		},
		Secrets: map[string]string{
			"password": "user_pw",
			"secret":   "client_secret_456",
		},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "products", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	// Token request must use Basic auth over base64(client_id:secret).
	wantBasic := "Basic " + base64.StdEncoding.EncodeToString([]byte("cid_123:client_secret_456"))
	if sawTokenBasic != wantBasic {
		t.Fatalf("token Authorization = %q, want %q", sawTokenBasic, wantBasic)
	}
	if sawTokenBody["grant_type"] != "password" || sawTokenBody["username"] != "api_user" || sawTokenBody["password"] != "user_pw" {
		t.Fatalf("token body = %+v, want password grant with api_user/user_pw", sawTokenBody)
	}
	// Resource request must carry the bearer access token.
	if sawResourceAuth != "Bearer tok_abc" {
		t.Fatalf("resource Authorization = %q, want Bearer tok_abc", sawResourceAuth)
	}
	if pageRequests != 2 {
		t.Fatalf("resource page requests = %d, want 2", pageRequests)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 across 2 pages", len(got))
	}
	if got[0]["id"] != "prod_1" {
		t.Fatalf("record 0 id = %v, want prod_1 (identifier mapped to id)", got[0]["id"])
	}
	if got[2]["id"] != "prod_3" {
		t.Fatalf("record 2 id = %v, want prod_3", got[2]["id"])
	}
}

func TestFixtureModeReadNoNetwork(t *testing.T) {
	c := akeneo.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	for _, stream := range []string{"products", "categories", "families", "attributes"} {
		got = got[:0]
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("Read(%s) fixture: %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("Read(%s) fixture emitted no records", stream)
		}
		for _, rec := range got {
			if rec["id"] == nil || rec["id"] == "" {
				t.Fatalf("Read(%s) fixture record missing id: %+v", stream, rec)
			}
		}
	}
}

func TestCheckFixtureModeNoCreds(t *testing.T) {
	c := akeneo.New()
	if err := c.Check(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

func TestCatalogStreams(t *testing.T) {
	c := akeneo.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "akeneo" {
		t.Fatalf("catalog connector = %q, want akeneo", cat.Connector)
	}
	want := map[string]bool{"products": false, "categories": false, "families": false, "attributes": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q has no primary key", s.Name)
		}
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("catalog missing core stream %q", name)
		}
	}
}

func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := akeneo.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"host": "ftp://evil.example.com"},
		Secrets: map[string]string{"password": "p", "secret": "s"},
	}
	err := c.Check(context.Background(), cfg)
	if err == nil || !strings.Contains(err.Error(), "host") {
		t.Fatalf("Check with ftp host = %v, want host scheme error", err)
	}
}

func TestRegisteredReadOnly(t *testing.T) {
	_ = akeneo.New() // ensure init() ran
	caps := akeneo.New().Metadata().Capabilities
	if !caps.Read || !caps.Catalog {
		t.Fatalf("capabilities = %+v, want Read && Catalog", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write false (PIM is read-only here)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("akeneo"); !ok {
		t.Fatal("registry did not resolve akeneo (self-registration)")
	}
}
