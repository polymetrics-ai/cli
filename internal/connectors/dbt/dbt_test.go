package dbt_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/dbt"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the dbt connector:
// it asserts the dbt Cloud "Token <key>" Authorization header, offset/limit
// pagination over the data[] array (driven by extra.pagination.total_count), the
// account-scoped path, and record mapping.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var sawPaths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		sawPaths = append(sawPaths, r.URL.Path)
		if r.URL.Path != "/accounts/42/projects/" {
			http.NotFound(w, r)
			return
		}
		offset := r.URL.Query().Get("offset")
		switch offset {
		case "", "0":
			_, _ = w.Write([]byte(`{"status":{"code":200},"data":[{"id":1,"name":"analytics","account_id":42,"state":1},{"id":2,"name":"finance","account_id":42,"state":1}],"extra":{"pagination":{"count":2,"total_count":3},"filters":{"limit":2,"offset":0}}}`))
		case "2":
			_, _ = w.Write([]byte(`{"status":{"code":200},"data":[{"id":3,"name":"marketing","account_id":42,"state":1}],"extra":{"pagination":{"count":1,"total_count":3},"filters":{"limit":2,"offset":2}}}`))
		default:
			t.Errorf("unexpected offset=%q", offset)
			_, _ = w.Write([]byte(`{"status":{"code":200},"data":[],"extra":{"pagination":{"count":0,"total_count":3}}}`))
		}
	}))
	defer srv.Close()

	c := dbt.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "account_id": "42", "page_size": "2"},
		Secrets: map[string]string{"api_key_2": "dbtu_secret"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "projects", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Token dbtu_secret" {
		t.Fatalf("Authorization = %q, want Token dbtu_secret", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 across 2 pages; paths=%v", len(got), sawPaths)
	}
	if len(sawPaths) != 2 {
		t.Fatalf("requests = %d, want 2 pages", len(sawPaths))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["name"] == nil {
			t.Fatalf("record missing id/name: %+v", rec)
		}
	}
}

// TestReadStopsAtTotalCount confirms the paginator halts once the full page is
// shorter than the page size even when total_count is consistent.
func TestReadStopsWhenShortPage(t *testing.T) {
	hits := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		_, _ = w.Write([]byte(`{"status":{"code":200},"data":[{"id":7,"name":"only","account_id":1}],"extra":{"pagination":{"count":1,"total_count":1}}}`))
	}))
	defer srv.Close()

	c := dbt.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "account_id": "1", "page_size": "100"},
		Secrets: map[string]string{"api_key_2": "k"},
	}
	var n int
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "runs", Config: cfg}, func(connectors.Record) error {
		n++
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if n != 1 {
		t.Fatalf("records = %d, want 1", n)
	}
	if hits != 1 {
		t.Fatalf("requests = %d, want 1 (short page stops pagination)", hits)
	}
}

func TestFixtureModeNoNetwork(t *testing.T) {
	c := dbt.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture", "account_id": "1"}}
	for _, stream := range []string{"projects", "runs", "repositories", "users", "environments"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture Read(%s) returned no records", stream)
		}
		for _, rec := range got {
			if rec["id"] == nil {
				t.Fatalf("fixture %s record missing id: %+v", stream, rec)
			}
		}
	}
	// Check must short-circuit in fixture mode without creds.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

func TestCheckRequiresSecretAndAccount(t *testing.T) {
	c := dbt.New()
	// Missing secret.
	err := c.Check(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"account_id": "1"}})
	if err == nil {
		t.Fatal("Check should fail without api_key_2 secret")
	}
	// Missing account_id.
	err = c.Check(context.Background(), connectors.RuntimeConfig{Secrets: map[string]string{"api_key_2": "k"}})
	if err == nil {
		t.Fatal("Check should fail without account_id")
	}
}

func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := dbt.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example.com", "account_id": "1"},
		Secrets: map[string]string{"api_key_2": "k"},
	}
	if err := c.Check(context.Background(), cfg); err == nil {
		t.Fatal("Check should reject non-http(s) base_url")
	}
}

func TestCatalogStreams(t *testing.T) {
	c := dbt.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "dbt" {
		t.Fatalf("Catalog.Connector = %q, want dbt", cat.Connector)
	}
	want := map[string]bool{"projects": false, "runs": false, "repositories": false, "users": false, "environments": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q missing primary key", s.Name)
		}
		if len(s.Fields) == 0 {
			t.Fatalf("stream %q missing fields", s.Name)
		}
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

func TestRegistryResolvesDBT(t *testing.T) {
	_ = dbt.New() // ensure init ran
	r := connectors.NewRegistry()
	got, ok := r.Get("dbt")
	if !ok {
		t.Fatal("registry did not resolve dbt (self-registration failed)")
	}
	caps := got.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read+Catalog+Check", caps)
	}
}
