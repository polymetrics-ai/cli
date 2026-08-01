package ashby

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
)

func TestConnectorContract(t *testing.T) {
	assertConnectorContract(t, New(), "ashby")
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
	if !caps.Write {
		t.Fatalf("%s must expose typed reverse-ETL write capability", wantName)
	}
	if _, ok := c.(connectors.WriteValidator); !ok {
		t.Fatalf("%s must validate typed write records", wantName)
	}
	if _, ok := c.(connectors.DryRunWriter); !ok {
		t.Fatalf("%s must dry-run typed write records", wantName)
	}
	if _, ok := c.(connectors.OperationDirectReader); !ok {
		t.Fatalf("%s must expose bounded operation direct reads", wantName)
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
	if len(cat.Streams) < 70 {
		t.Fatalf("Catalog returned %d streams, want Ashby parity stream coverage", len(cat.Streams))
	}
}

func TestValidateWriteAndDryRun(t *testing.T) {
	c := New()
	validator := c.(connectors.WriteValidator)
	dryRunner := c.(connectors.DryRunWriter)
	req := connectors.WriteRequest{Action: "add_candidate_tag", Config: connectors.RuntimeConfig{Config: map[string]string{"base_url": "https://api.ashbyhq.com"}}}
	records := []connectors.Record{{"candidateId": "candidate_fixture", "tagId": "tag_fixture"}}
	if err := validator.ValidateWrite(context.Background(), req, records); err != nil {
		t.Fatalf("ValidateWrite: %v", err)
	}
	preview, err := dryRunner.DryRunWrite(context.Background(), req, records)
	if err != nil {
		t.Fatalf("DryRunWrite: %v", err)
	}
	if preview.RecordsStaged != 1 || preview.Action != "add_candidate_tag" {
		t.Fatalf("preview = %+v, want one staged add_candidate_tag", preview)
	}
}

func TestOperationDirectReadUsesFixedSearchPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/candidate.search" {
			t.Fatalf("path = %q, want /candidate.search", r.URL.Path)
		}
		user, pass, ok := r.BasicAuth()
		if !ok || user != "test_key" || pass != "" {
			t.Fatalf("basic auth = (%q,%q,%v), want Ashby key as username with blank password", user, pass, ok)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if body["email"] != "candidate@example.invalid" {
			t.Fatalf("body[email] = %v, want candidate@example.invalid", body["email"])
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"results":[]}`))
	}))
	defer server.Close()

	reader := New().(connectors.OperationDirectReader)
	result, err := reader.OperationDirectRead(context.Background(), connectors.OperationDirectReadRequest{
		Operation: "ashby.direct.candidate.search",
		Config: connectors.RuntimeConfig{
			Config:  map[string]string{"base_url": server.URL},
			Secrets: map[string]string{"api_key": "test_key"},
		},
		Body:         map[string]any{"email": "candidate@example.invalid"},
		OutputPolicy: "json_redacted",
		MaxBytes:     1 << 20,
	})
	if err != nil {
		t.Fatalf("OperationDirectRead: %v", err)
	}
	if result.Status != http.StatusOK {
		t.Fatalf("status = %d, want 200", result.Status)
	}
}
