package chift_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics/internal/connectors"
	"polymetrics/internal/connectors/chift"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Chift
// connector. It asserts the session-token exchange (POST /token with
// clientId/clientSecret/accountId), the Bearer header carried on data requests,
// offset/limit pagination over a top-level JSON array across two pages, and
// record mapping. Red until internal/connectors/chift exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var (
		tokenBody map[string]any
		sawAuth   string
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			defer r.Body.Close()
			raw, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(raw, &tokenBody)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"tok_abc","expires_in":3600}`))
		case "/consumers":
			sawAuth = r.Header.Get("Authorization")
			switch r.URL.Query().Get("offset") {
			case "", "0":
				_, _ = w.Write([]byte(`[{"consumerid":"con_1","name":"Acme"},{"consumerid":"con_2","name":"Globex"}]`))
			case "2":
				_, _ = w.Write([]byte(`[{"consumerid":"con_3","name":"Initech"}]`))
			default:
				t.Errorf("unexpected offset=%q", r.URL.Query().Get("offset"))
				_, _ = w.Write([]byte(`[]`))
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := chift.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{
			"client_id":     "cid",
			"client_secret": "csecret",
			"account_id":    "acct_1",
		},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "consumers", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer tok_abc" {
		t.Fatalf("Authorization = %q, want Bearer tok_abc", sawAuth)
	}
	if tokenBody["clientId"] != "cid" || tokenBody["clientSecret"] != "csecret" || tokenBody["accountId"] != "acct_1" {
		t.Fatalf("token request body = %+v, want clientId/clientSecret/accountId set", tokenBody)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	if got[0]["consumerid"] != "con_1" || got[0]["name"] != "Acme" {
		t.Fatalf("first record mapped wrong: %+v", got[0])
	}
	if got[2]["consumerid"] != "con_3" {
		t.Fatalf("last record mapped wrong: %+v", got[2])
	}
}

// TestFixtureModeNoNetwork asserts fixture mode emits deterministic records
// without any network access so conformance runs credential-free.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := chift.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"consumers", "connections", "syncs"} {
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

// TestCatalogStreams asserts the published catalog exposes the core streams with
// primary keys.
func TestCatalogStreams(t *testing.T) {
	c := chift.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]string{"consumers": "consumerid", "connections": "connectionid", "syncs": "syncid"}
	seen := map[string]bool{}
	for _, s := range cat.Streams {
		seen[s.Name] = true
		if pk, ok := want[s.Name]; ok {
			if len(s.PrimaryKey) == 0 || s.PrimaryKey[0] != pk {
				t.Fatalf("stream %s primary key = %v, want [%s]", s.Name, s.PrimaryKey, pk)
			}
		}
	}
	for name := range want {
		if !seen[name] {
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

// TestRegisteredReadOnly asserts self-registration via the package registry and
// the declared read-only capability set.
func TestRegisteredReadOnly(t *testing.T) {
	_ = chift.New() // ensure init ran
	caps := chift.New().Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only connector)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("chift"); !ok {
		t.Fatal("registry did not resolve chift (self-registration)")
	}
}
