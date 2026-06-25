package countercyclical_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/countercyclical"
)

// TestReadAuthenticatesAndPaginates is the red-first test: the Countercyclical
// API authenticates via an `apiKey` query parameter (ApiKeyAuthenticator,
// inject_into request_parameter), returns a root-level JSON array of records,
// and the connector maps id/name through. The connector also defends against a
// paginated server via an offset loop, so this server returns two pages and the
// reader must walk both.
func TestReadAuthenticatesAndPaginates(t *testing.T) {
	var sawKey string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawKey = r.URL.Query().Get("apiKey")
		if r.URL.Path != "/investments" {
			http.NotFound(w, r)
			return
		}
		offset := r.URL.Query().Get("offset")
		switch offset {
		case "", "0":
			// Full page (page size 2) -> reader should request the next page.
			_, _ = w.Write([]byte(`[{"id":"inv_1","name":"Acme","tickerSymbol":"ACME"},{"id":"inv_2","name":"Globex","tickerSymbol":"GBX"}]`))
		case "2":
			// Short page -> reader stops.
			_, _ = w.Write([]byte(`[{"id":"inv_3","name":"Initech","tickerSymbol":"INI"}]`))
		default:
			t.Errorf("unexpected offset=%q", offset)
			_, _ = w.Write([]byte(`[]`))
		}
	}))
	defer srv.Close()

	c := countercyclical.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"api_key": "key_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "investments", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawKey != "key_test_123" {
		t.Fatalf("apiKey query = %q, want key_test_123", sawKey)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["name"] == nil {
			t.Fatalf("record missing id/name: %+v", rec)
		}
	}
	if got[0]["id"] != "inv_1" || got[2]["id"] != "inv_3" {
		t.Fatalf("unexpected record ids: %v / %v", got[0]["id"], got[2]["id"])
	}
}

// TestReadSingleRootArrayNoPagination confirms the honest default path: the real
// API has no pagination, so a single root-array response is fully read with no
// follow-up request.
func TestReadSingleRootArrayNoPagination(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.URL.Path != "/memos" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`[{"id":"memo_1","title":"Thesis"},{"id":"memo_2","title":"Update"}]`))
	}))
	defer srv.Close()

	c := countercyclical.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "key_test_123"},
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "memos", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2", len(got))
	}
	// A short (< page size) first page must not trigger a second request.
	if calls != 1 {
		t.Fatalf("server calls = %d, want 1 (no pagination)", calls)
	}
	if got[0]["title"] != "Thesis" {
		t.Fatalf("title mapping wrong: %v", got[0]["title"])
	}
}

func TestFixtureModeNoNetwork(t *testing.T) {
	c := countercyclical.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"investments", "valuations", "memos"} {
		var got []connectors.Record
		if err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		}); err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture Read(%s) produced no records", stream)
		}
		for _, rec := range got {
			if rec["id"] == nil {
				t.Fatalf("fixture %s record missing id: %+v", stream, rec)
			}
		}
	}
	// Check must also short-circuit in fixture mode with no creds.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

func TestCatalogStreams(t *testing.T) {
	c := countercyclical.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"investments": false, "valuations": false, "memos": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) != 1 || s.PrimaryKey[0] != "id" {
			t.Fatalf("stream %s primary key = %v, want [id]", s.Name, s.PrimaryKey)
		}
		if len(s.Fields) == 0 {
			t.Fatalf("stream %s has no fields", s.Name)
		}
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := countercyclical.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example.com"},
		Secrets: map[string]string{"api_key": "k"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "investments", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read with non-http base_url should fail (SSRF guard)")
	}
}

func TestReadRequiresSecret(t *testing.T) {
	c := countercyclical.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{}}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "investments", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read without api_key secret should fail")
	}
}

func TestRegisteredReadOnly(t *testing.T) {
	_ = countercyclical.New() // ensure init ran
	c := countercyclical.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (source connector)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("countercyclical"); !ok {
		t.Fatal("registry did not resolve countercyclical (self-registration)")
	}
}
