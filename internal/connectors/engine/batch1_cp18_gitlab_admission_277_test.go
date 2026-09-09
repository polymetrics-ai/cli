package engine

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"polymetrics.ai/internal/connectors"
)

// A source-backed write response is not a status-only assertion.  The normal
// public Write consumer must validate its declared JSON schema after a 200,
// while preserving ordinary source-relevant error and malformed-response
// failures.
func TestBatch1CP18GitLabAdmission277WriteResponseSchema(t *testing.T) {
	for _, tc := range []struct {
		name      string
		status    int
		body      string
		content   string
		wantError string
	}{
		{name: "source_valid_success", status: http.StatusOK, body: `{"key":"CI_TOKEN","value":"masked","variable_type":"env_var"}`, content: "application/json"},
		{name: "source_error_status", status: http.StatusBadRequest, body: `{"message":"invalid key"}`, content: "application/json", wantError: "provider returned HTTP status 400"},
		{name: "malformed_success", status: http.StatusOK, body: `{"key":`, content: "application/json", wantError: "provider response is not valid JSON"},
		{name: "wrong_success_shape", status: http.StatusOK, body: `{"key":42}`, content: "application/json", wantError: "response_schema"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			requests := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests++
				if r.Method != http.MethodDelete || r.URL.Path != "/api/v4/admin/ci/variables/CI_TOKEN" {
					t.Errorf("wire = %s %s", r.Method, r.URL.Path)
				}
				w.Header().Set("Content-Type", tc.content)
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			t.Cleanup(server.Close)

			bundle := Bundle{
				Name: "gitlab-fixture",
				HTTP: HTTPBase{URL: server.URL},
				Writes: []WriteAction{{
					Name:            "delete_ci_variable",
					Kind:            "delete",
					Method:          http.MethodDelete,
					Path:            "/api/v4/admin/ci/variables/{{ record.key }}",
					PathFields:      []string{"key"},
					BodyType:        "none",
					SuccessStatuses: []int{http.StatusOK},
					RecordSchema:    []byte(`{"type":"object","additionalProperties":false,"required":["key"],"properties":{"key":{"type":"string"}}}`),
					ResponseSchema:  []byte(`{"type":"object","properties":{"key":{"type":"string"},"value":{"type":"string"},"variable_type":{"type":"string"}}}`),
				}},
			}
			connector := New(bundle, nil)
			request := connectors.WriteRequest{Action: "delete_ci_variable", Config: connectors.RuntimeConfig{
				CredentialRevision:  "fixture-credential-revision",
				ConfigurationDigest: "fixture-configuration-digest",
				WriteApprovalScope:  connectors.WriteApprovalScopeFixture,
			}}
			preview, err := connector.DryRunWrite(context.Background(), request, []connectors.Record{{"key": "CI_TOKEN"}})
			if err != nil {
				t.Fatalf("DryRunWrite: %v", err)
			}
			request.Approval = approvedEvidenceForPreview(t, preview)
			result, err := connector.Write(context.Background(), request, []connectors.Record{{"key": "CI_TOKEN"}})
			if tc.wantError == "" {
				if err != nil {
					t.Fatalf("Write: %v", err)
				}
				if result.RecordsWritten != 1 || result.RecordsFailed != 0 || len(result.ProviderResponses) != 1 {
					t.Fatalf("result = %+v", result)
				}
			} else if err == nil || !strings.Contains(err.Error(), tc.wantError) {
				t.Fatalf("Write error = %v, want %q", err, tc.wantError)
			}
			if requests != 1 {
				t.Fatalf("requests = %d, want 1", requests)
			}
		})
	}
}

// A source-backed response contract must survive the same strict bundle load
// that lock-render uses. Constructing a WriteAction directly would not cover
// the closed writes.json schema or the real decoder.
func TestBatch1CP18GitLabAdmission277WriteResponseSchemaBundleLoad(t *testing.T) {
	files := fullValidBundleFS("gitlab")
	files["gitlab/writes.json"] = &fstest.MapFile{Data: []byte(`{
  "actions": [{
    "name": "delete_ci_variable",
    "kind": "delete",
    "method": "DELETE",
    "path": "/api/v4/admin/ci/variables/{{ record.key }}",
    "path_fields": ["key"],
    "body_type": "none",
    "confirm": "destructive",
    "delete": {"idempotent": true},
    "record_schema": {"type": "object", "required": ["key"], "properties": {"key": {"type": "string"}}},
    "response_schema": {"type": "object", "properties": {"key": {"type": "string"}, "value": {"type": "string"}}},
    "risk": "deletes a source-bound CI variable"
  }]
}`)}
	bundle, err := Load(files, "gitlab")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(bundle.Writes) != 1 || len(bundle.Writes[0].ResponseSchema) == 0 {
		t.Fatalf("ResponseSchema was not retained by the bundle loader: %+v", bundle.Writes)
	}
	if got := string(bundle.Writes[0].ResponseSchema); !strings.Contains(got, `"value"`) {
		t.Fatalf("ResponseSchema = %s, want declared CI-variable shape", got)
	}
}
