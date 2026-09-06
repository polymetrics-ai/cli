package engine

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
)

func declaredBatchTestBundle(baseURL string) Bundle {
	return Bundle{
		Name: "acme",
		HTTP: HTTPBase{URL: baseURL},
		Writes: []WriteAction{
			{
				Name: "create_item", Kind: "create", Method: http.MethodPost, Path: "/items", BodyType: "json", BodyFields: []string{"data"},
				RecordSchema: json.RawMessage(`{"type":"object","additionalProperties":false,"required":["data"],"properties":{"data":{"type":"object","additionalProperties":false,"required":["name"],"properties":{"name":{"type":"string"}}}}}`),
			},
			{
				Name: "delete_item", Kind: "delete", Method: http.MethodDelete, Path: "/items/{{ record.id }}", PathFields: []string{"id"}, BodyType: "none", Confirm: "destructive",
				RecordSchema: json.RawMessage(`{"type":"object","additionalProperties":false,"required":["id"],"properties":{"id":{"type":"string"}}}`),
			},
			{
				Name: "submit_batch", Kind: "custom", Method: http.MethodPost, Path: "/batch", BodyType: "declared_batch", Confirm: "destructive",
				DeclaredBatch: &DeclaredBatchSpec{
					MaxActions: 2, AllowedActions: []string{"create_item", "delete_item"}, AllowedMethods: []string{http.MethodPost, http.MethodDelete},
					ProviderEnvelopeField: "data", ProviderActionsField: "actions", ProviderMethodField: "method", ProviderPathField: "relative_path", ProviderDataField: "data", InnerBodyField: "data",
					ResponseEnvelopeField: "data", ResponseStatusField: "status_code",
				},
				RecordSchema: json.RawMessage(`{"type":"object","additionalProperties":false,"required":["actions"],"properties":{"actions":{"type":"array","minItems":1,"maxItems":2,"items":{"type":"object","additionalProperties":false,"required":["action","record"],"properties":{"action":{"type":"string","enum":["create_item","delete_item"]},"record":{"type":"object"}}}}}}`),
			},
		},
	}
}

func declaredBatchTestRecord() connectors.Record {
	return connectors.Record{"actions": []any{
		map[string]any{"action": "create_item", "record": map[string]any{"data": map[string]any{"name": "alpha"}}},
		map[string]any{"action": "delete_item", "record": map[string]any{"id": "item/2"}},
	}}
}

func TestDeclaredBatchExecutesOnlyResolvedNamedActions(t *testing.T) {
	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Method != http.MethodPost || r.URL.Path != "/batch" || r.URL.RawQuery != "" {
			t.Fatalf("batch request = %s %s?%s", r.Method, r.URL.Path, r.URL.RawQuery)
		}
		raw, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read batch body: %v", err)
		}
		var got any
		decoder := json.NewDecoder(strings.NewReader(string(raw)))
		decoder.UseNumber()
		if err := decoder.Decode(&got); err != nil {
			t.Fatalf("decode batch body: %v", err)
		}
		want := map[string]any{"data": map[string]any{"actions": []any{
			map[string]any{"method": "post", "relative_path": "/items", "data": map[string]any{"name": "alpha"}},
			map[string]any{"method": "delete", "relative_path": "/items/item%2F2"},
		}}}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("batch body = %#v, want %#v", got, want)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"data":[{"status_code":201},{"status_code":200}]}`)
	}))
	defer server.Close()

	bundle := declaredBatchTestBundle(server.URL)
	record := declaredBatchTestRecord()
	req := connectors.WriteRequest{Action: "submit_batch", Config: connectors.RuntimeConfig{
		CredentialRevision: "fixture-revision", ConfigurationDigest: "fixture-config", WriteApprovalScope: connectors.WriteApprovalScopeFixture,
	}}
	preview, err := DryRunWrite(context.Background(), bundle, req, []connectors.Record{record}, nil)
	if err != nil {
		t.Fatalf("DryRunWrite: %v", err)
	}
	if calls != 0 || preview.RecordsStaged != 1 || preview.Digest == "" {
		t.Fatalf("preview = %+v calls=%d, want one no-I/O approval-bound record", preview, calls)
	}
	req.Approval = approvedEvidenceForPreview(t, preview)
	result, err := Write(context.Background(), bundle, req, []connectors.Record{record}, nil)
	if err != nil {
		t.Fatalf("Write declared batch: %v", err)
	}
	if calls != 1 || result.RecordsWritten != 1 || result.RecordsFailed != 0 {
		t.Fatalf("declared batch result = %+v calls=%d, want one provider request and one completed record", result, calls)
	}
}

func TestDeclaredBatchRejectsOpenOrInvalidSubrequestsBeforeProviderIO(t *testing.T) {
	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { calls++ }))
	defer server.Close()
	bundle := declaredBatchTestBundle(server.URL)

	tests := []struct {
		name    string
		record  connectors.Record
		wantErr string
	}{
		{name: "caller method and path", record: connectors.Record{"actions": []any{map[string]any{
			"action": "create_item", "record": map[string]any{"data": map[string]any{"name": "alpha"}}, "method": "delete", "relative_path": "/other",
		}}}, wantErr: "additional"},
		{name: "undeclared action", record: connectors.Record{"actions": []any{map[string]any{
			"action": "raw_http", "record": map[string]any{},
		}}}, wantErr: "enum"},
		{name: "inner schema violation", record: connectors.Record{"actions": []any{map[string]any{
			"action": "create_item", "record": map[string]any{"data": map[string]any{"unknown": true}},
		}}}, wantErr: "name"},
		{name: "empty", record: connectors.Record{"actions": []any{}}, wantErr: "minItems"},
		{name: "over max", record: connectors.Record{"actions": []any{
			map[string]any{"action": "delete_item", "record": map[string]any{"id": "1"}},
			map[string]any{"action": "delete_item", "record": map[string]any{"id": "2"}},
			map[string]any{"action": "delete_item", "record": map[string]any{"id": "3"}},
		}}, wantErr: "maxItems"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := DryRunWrite(context.Background(), bundle, connectors.WriteRequest{Action: "submit_batch"}, []connectors.Record{test.record}, nil)
			if err == nil || !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(test.wantErr)) {
				t.Fatalf("DryRunWrite error = %v, want %q", err, test.wantErr)
			}
		})
	}
	if calls != 0 {
		t.Fatalf("invalid declared batches made %d provider calls, want zero", calls)
	}
}

func TestDeclaredBatchFailsClosedOnPartialOrMalformedProviderResults(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		wantErr string
	}{
		{name: "partial failure", body: `{"data":[{"status_code":201},{"status_code":429}]}`, wantErr: "subrequest 1 returned status 429"},
		{name: "wrong cardinality", body: `{"data":[{"status_code":201}]}`, wantErr: "2 subrequest results"},
		{name: "missing status", body: `{"data":[{"status_code":201},{}]}`, wantErr: "status_code"},
		{name: "malformed envelope", body: `{"data":{}}`, wantErr: "response field"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var calls int
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				calls++
				w.Header().Set("Content-Type", "application/json")
				_, _ = io.WriteString(w, test.body)
			}))
			defer server.Close()
			bundle := declaredBatchTestBundle(server.URL)
			record := declaredBatchTestRecord()
			req := connectors.WriteRequest{Action: "submit_batch", Config: connectors.RuntimeConfig{
				CredentialRevision: "fixture-revision", ConfigurationDigest: "fixture-config", WriteApprovalScope: connectors.WriteApprovalScopeFixture,
			}}
			preview, err := DryRunWrite(context.Background(), bundle, req, []connectors.Record{record}, nil)
			if err != nil {
				t.Fatalf("DryRunWrite: %v", err)
			}
			req.Approval = approvedEvidenceForPreview(t, preview)
			result, err := Write(context.Background(), bundle, req, []connectors.Record{record}, nil)
			if err == nil || !strings.Contains(err.Error(), test.wantErr) {
				t.Fatalf("Write error = %v, want %q", err, test.wantErr)
			}
			if calls != 1 || result.RecordsWritten != 0 || result.RecordsFailed != 1 {
				t.Fatalf("failed batch result = %+v calls=%d, want one ambiguous provider receipt and failed record", result, calls)
			}
		})
	}
}

func TestValidateDeclaredBatchDefinitionRejectsOpenSelection(t *testing.T) {
	bundle := declaredBatchTestBundle("https://provider.example.test/api/1.0")
	if err := validateWriteBodies(bundle.Writes); err != nil {
		t.Fatalf("valid declared batch definition: %v", err)
	}

	tests := []struct {
		name    string
		mutate  func([]WriteAction)
		wantErr string
	}{
		{name: "nested self", mutate: func(actions []WriteAction) {
			actions[2].DeclaredBatch.AllowedActions = append(actions[2].DeclaredBatch.AllowedActions, "submit_batch")
		}, wantErr: "cannot select itself"},
		{name: "unknown action", mutate: func(actions []WriteAction) { actions[2].DeclaredBatch.AllowedActions[0] = "unknown" }, wantErr: "unknown write action"},
		{name: "method outside source contract", mutate: func(actions []WriteAction) { actions[2].DeclaredBatch.AllowedMethods = []string{http.MethodPost} }, wantErr: "method DELETE"},
		{name: "destructive selection without confirmation", mutate: func(actions []WriteAction) { actions[2].Confirm = "" }, wantErr: "destructive confirmation"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			copyBundle := declaredBatchTestBundle(bundle.HTTP.URL)
			test.mutate(copyBundle.Writes)
			err := validateWriteBodies(copyBundle.Writes)
			if err == nil || !strings.Contains(err.Error(), test.wantErr) {
				t.Fatalf("validateWriteBodies error = %v, want %q", err, test.wantErr)
			}
		})
	}
}
