package serpstat_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/serpstat"
)

func TestContractFixtureAndWrite(t *testing.T) {
	c := serpstat.New()
	if c.Name() != "serpstat" {
		t.Fatalf("Name() = %q, want serpstat", c.Name())
	}
	caps := c.Metadata().Capabilities
	if !caps.Check || !caps.Catalog || !caps.Read || caps.Write {
		t.Fatalf("capabilities = %+v, want read-only Check/Catalog/Read", caps)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) == 0 || cat.Streams[0].Name != "domain_keywords" {
		t.Fatalf("catalog streams = %+v, want domain_keywords first", cat.Streams)
	}
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture): %v", err)
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "domain_keywords", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read(fixture): %v", err)
	}
	if len(got) == 0 || got[0]["keyword"] == nil {
		t.Fatalf("fixture records = %+v, want keyword", got)
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
	if _, ok := connectors.NewRegistry().Get("serpstat"); !ok {
		t.Fatal("registry did not resolve serpstat")
	}
}

func TestReadDomainKeywordsPostsJSONRPCWithToken(t *testing.T) {
	var sawToken, sawMethod bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v4" || r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		sawToken = r.URL.Query().Get("token") == "test-token"
		body := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(body)
		sawMethod = strings.Contains(string(body), "SerpstatDomainProcedure.getKeywords")
		_, _ = w.Write([]byte(`{"result":{"data":[{"keyword":"etl","position":3,"url":"https://example.com"}]}}`))
	}))
	defer srv.Close()

	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL + "/v4", "domain": "example.com", "page_size": "1", "pages_to_fetch": "1"}, Secrets: map[string]string{"api_key": "test-token"}}
	var got []connectors.Record
	if err := serpstat.New().Read(context.Background(), connectors.ReadRequest{Stream: "domain_keywords", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if !sawToken || !sawMethod {
		t.Fatal("json-rpc token or method was not applied")
	}
	if len(got) != 1 || got[0]["keyword"] != "etl" {
		t.Fatalf("records = %+v, want keyword", got)
	}
}
