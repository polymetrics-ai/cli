package engine

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/defs"
)

func TestXeroReportOperationsDirectRead(t *testing.T) {
	bundle, err := Load(defs.FS, "xero")
	if err != nil {
		t.Fatalf("load xero bundle: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/api.xro/2.0/Reports") {
			t.Fatalf("unexpected report path %q", r.URL.Path)
		}
		if r.Header.Get("Xero-tenant-id") == "" {
			t.Fatalf("missing Xero tenant header")
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"Reports": []map[string]any{{"ReportID": "report_fixture_1", "ReportName": "Fixture Report"}},
		})
	}))
	defer server.Close()

	cfg := connectors.RuntimeConfig{
		Config: map[string]string{"base_url": server.URL + "/api.xro/2.0"},
		Secrets: map[string]string{
			"access_token": "synthetic-conformance-secret",
			"tenant_id":    "synthetic-conformance-tenant",
		},
	}

	runCount := 0
	for _, op := range bundle.Operations {
		if op.Kind != "rest_read" || !strings.HasPrefix(op.ID, "xero.get_report") {
			continue
		}
		runCount++
		t.Run(op.ID, func(t *testing.T) {
			req := connectors.OperationDirectReadRequest{
				Operation:    op.ID,
				Config:       cfg,
				OutputPolicy: "json_redacted",
				MaxBytes:     1 << 20,
			}
			if strings.Contains(op.REST.Path, "{ReportID}") {
				req.PathParams = map[string]string{"ReportID": "report_fixture_1"}
			}
			result, err := OperationDirectRead(context.Background(), bundle, req, nil)
			if err != nil {
				t.Fatalf("OperationDirectRead: %v", err)
			}
			if result.Status != http.StatusOK || !strings.HasPrefix(result.Path, "/Reports") {
				t.Fatalf("result = %+v, want status 200 and report path", result)
			}
		})
	}
	if runCount != 11 {
		t.Fatalf("report rest_read operations exercised = %d, want 11", runCount)
	}
}
