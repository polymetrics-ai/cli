package alphavantage

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
)

func TestConnectorContract(t *testing.T) {
	assertConnectorContract(t, New(), "alpha-vantage")
}

func TestReadReportsProviderRateLimitActivity(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		attempts++
		if attempts == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"Global Quote":{"01. symbol":"IBM"}}`))
	}))
	t.Cleanup(server.Close)

	report := connectors.NewRateLimitReport()
	report.Declare("alpha-vantage", connectors.RateLimitDeclarationUndeclared)
	err := (Connector{Client: server.Client()}).Read(context.Background(), connectors.ReadRequest{
		Stream: "global_quote",
		Config: connectors.RuntimeConfig{
			Config:          map[string]string{"base_url": server.URL},
			Secrets:         map[string]string{"api_key": "alpha-vantage-secret-must-not-escape"},
			RateLimitReport: report,
		},
	}, func(connectors.Record) error { return nil })
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	summary := report.Snapshot()
	if len(summary.Connectors) != 1 {
		t.Fatalf("rate-limit connector count = %d, want 1", len(summary.Connectors))
	}
	connector := summary.Connectors[0]
	if connector.Declaration != connectors.RateLimitDeclarationUndeclared || connector.Provider429Observed != 1 || connector.Provider429Honored != 1 || connector.RequestCount != 2 {
		t.Fatalf("read rate-limit activity = %+v", connector)
	}
	encoded, err := json.Marshal(summary)
	if err != nil {
		t.Fatalf("marshal rate-limit activity: %v", err)
	}
	for _, forbidden := range []string{"alpha-vantage-secret-must-not-escape", server.URL} {
		if strings.Contains(string(encoded), forbidden) {
			t.Fatalf("rate-limit activity leaked %q", forbidden)
		}
	}
}

func assertConnectorContract(t *testing.T, c connectors.Connector, wantName string) {
	t.Helper()
	if c == nil {
		t.Fatal("New() = nil")
	}
	if got := c.Name(); got != wantName {
		t.Fatalf("Name() = %q, want %q", got, wantName)
	}
	meta := c.Metadata()
	if meta.Name != wantName {
		t.Fatalf("Metadata().Name = %q, want %q", meta.Name, wantName)
	}
	caps := meta.Capabilities
	if !caps.Check || !caps.Catalog || !caps.Read {
		t.Fatalf("capabilities = %+v, want Check, Catalog, and Read", caps)
	}
	if caps.Write {
		t.Fatalf("%s is read-only; Write capability must be false", wantName)
	}
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture): %v", err)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != wantName {
		t.Fatalf("Catalog().Connector = %q, want %q", cat.Connector, wantName)
	}
	if len(cat.Streams) == 0 {
		t.Fatal("Catalog returned zero streams")
	}
}
