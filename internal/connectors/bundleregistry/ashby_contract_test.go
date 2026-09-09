package bundleregistry

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"

	"polymetrics.ai/internal/connectors"
)

func TestProductionAshbyHiringTeamRoleUsesFixedNamesOnlyBody(t *testing.T) {
	previousTransport := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = previousTransport })

	var requests atomic.Int64
	http.DefaultTransport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests.Add(1)
		if request.URL.Host != "api.ashbyhq.com" || request.URL.Path != "/hiringTeamRole.list" || request.URL.RawQuery != "" {
			t.Fatalf("Ashby hiring role request = %s, want fixed body route without query", request.URL)
		}
		var body map[string]any
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Fatalf("decode Ashby hiring role body: %v", err)
		}
		if body["namesOnly"] != true || body["limit"] != float64(100) {
			t.Fatalf("Ashby hiring role body = %#v, want fixed namesOnly=true and limit", body)
		}
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"success":true,"results":[{"value":"Hiring Manager"}],"moreDataAvailable":false}`)), Request: request}, nil
	})

	connector, ok := New().Get("ashby")
	if !ok {
		t.Fatal("production registry missing ashby")
	}
	var records []connectors.Record
	if err := connector.Read(context.Background(), connectors.ReadRequest{Stream: "hiring_team_role_list", Config: connectors.RuntimeConfig{Secrets: map[string]string{"api_key": "test-key"}}}, func(record connectors.Record) error {
		records = append(records, record)
		return nil
	}); err != nil {
		t.Fatalf("Ashby hiring role read: %v", err)
	}
	if len(records) != 1 || records[0]["value"] != "Hiring Manager" {
		t.Fatalf("Ashby hiring role records = %#v, want projected role title", records)
	}
	if requests.Load() != 1 {
		t.Fatalf("Ashby hiring role requests = %d, want one", requests.Load())
	}
}
