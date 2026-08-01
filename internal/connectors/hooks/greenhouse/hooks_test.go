package greenhouse

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/engine"
)

func TestExecuteWriteDestroyOpeningsRejectsNotDeleted(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"deleted":[101],"not_deleted":[102]}`))
	}))
	defer srv.Close()

	handled, err := New().ExecuteWrite(context.Background(), destroyOpeningsAction(), connectors.Record{
		"job_id": "job_id_fixture",
		"ids":    []any{float64(101), float64(102)},
	}, runtimeForServer(srv.URL))
	if !handled {
		t.Fatalf("ExecuteWrite handled = false, want true")
	}
	if err == nil || !strings.Contains(err.Error(), "not_deleted") {
		t.Fatalf("ExecuteWrite error = %v, want not_deleted failure", err)
	}
	if strings.Contains(err.Error(), "102") {
		t.Fatalf("ExecuteWrite error exposed opening id: %v", err)
	}
}

func TestExecuteWriteDestroyOpeningsSendsExpectedRequest(t *testing.T) {
	t.Parallel()

	type capturedRequest struct {
		method string
		path   string
		body   map[string]any
	}
	captured := make(chan capturedRequest, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		var body map[string]any
		if err := json.Unmarshal(raw, &body); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		captured <- capturedRequest{method: r.Method, path: r.URL.Path, body: body}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"deleted":[101],"not_deleted":[]}`))
	}))
	defer srv.Close()

	handled, err := New().ExecuteWrite(context.Background(), destroyOpeningsAction(), connectors.Record{
		"job_id": "job_id_fixture",
		"ids":    []any{float64(101)},
	}, runtimeForServer(srv.URL))
	if err != nil {
		t.Fatalf("ExecuteWrite returned error: %v", err)
	}
	if !handled {
		t.Fatalf("ExecuteWrite handled = false, want true")
	}
	got := <-captured
	if got.method != http.MethodDelete {
		t.Fatalf("method = %q, want DELETE", got.method)
	}
	if got.path != "/v2/jobs/job_id_fixture/openings" {
		t.Fatalf("path = %q, want destroy_openings path", got.path)
	}
	ids, ok := got.body["ids"].([]any)
	if !ok || len(ids) != 1 || ids[0] != float64(101) {
		t.Fatalf("body ids = %#v, want [101]", got.body["ids"])
	}
}

func TestHiringTeamSchemasRejectUnknownAndMissingLists(t *testing.T) {
	t.Parallel()

	bundle, err := engine.Load(os.DirFS("../../defs"), "greenhouse")
	if err != nil {
		t.Fatalf("load greenhouse bundle: %v", err)
	}

	req := connectors.WriteRequest{Action: "replace_hiring_team"}
	if err := engine.ValidateWrite(context.Background(), bundle, req, []connectors.Record{{"job_id": "job_id_fixture"}}); err == nil {
		t.Fatalf("ValidateWrite accepted a hiring-team record without a member list")
	}
	valid := connectors.Record{"job_id": "job_id_fixture", "hiring_managers": []any{float64(1234)}}
	if err := engine.ValidateWrite(context.Background(), bundle, req, []connectors.Record{valid}); err != nil {
		t.Fatalf("ValidateWrite rejected valid hiring-team record: %v", err)
	}
	withUnknown := connectors.Record{"job_id": "job_id_fixture", "hiring_managers": []any{float64(1234)}, "raw_body": map[string]any{"x": "y"}}
	if err := engine.ValidateWrite(context.Background(), bundle, req, []connectors.Record{withUnknown}); err == nil {
		t.Fatalf("ValidateWrite accepted an unknown hiring-team field")
	}
}

func TestExecuteWriteHiringTeamRequiresNonEmptyMemberList(t *testing.T) {
	t.Parallel()

	action := engine.WriteAction{Name: "replace_hiring_team"}
	tests := []struct {
		name    string
		record  connectors.Record
		wantErr bool
	}{
		{
			name:    "missing lists",
			record:  connectors.Record{"job_id": "job_id_fixture"},
			wantErr: true,
		},
		{
			name:    "empty list",
			record:  connectors.Record{"job_id": "job_id_fixture", "hiring_managers": []any{}},
			wantErr: true,
		},
		{
			name:    "empty list with non-empty list",
			record:  connectors.Record{"job_id": "job_id_fixture", "hiring_managers": []any{float64(1234)}, "recruiters": []any{}},
			wantErr: true,
		},
		{
			name:    "non-empty list",
			record:  connectors.Record{"job_id": "job_id_fixture", "hiring_managers": []any{float64(1234)}},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handled, err := New().ExecuteWrite(context.Background(), action, tt.record, nil)
			if handled {
				t.Fatalf("ExecuteWrite handled = true, want declarative fallback")
			}
			if tt.wantErr && err == nil {
				t.Fatalf("ExecuteWrite error = nil, want error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("ExecuteWrite error = %v, want nil", err)
			}
		})
	}
}

func TestExecuteWriteAnonymizeCandidateRequiresDocumentedFields(t *testing.T) {
	t.Parallel()

	action := engine.WriteAction{Name: "anonymize_candidate"}
	tests := []struct {
		name    string
		record  connectors.Record
		wantErr bool
	}{
		{
			name:    "empty fields",
			record:  connectors.Record{"candidate_id": "candidate_id_fixture", "field_names": []any{}},
			wantErr: true,
		},
		{
			name:    "unsupported field",
			record:  connectors.Record{"candidate_id": "candidate_id_fixture", "field_names": []any{"not_a_documented_field"}},
			wantErr: true,
		},
		{
			name:    "duplicate field",
			record:  connectors.Record{"candidate_id": "candidate_id_fixture", "field_names": []any{"full_name", "full_name"}},
			wantErr: true,
		},
		{
			name:    "documented fields",
			record:  connectors.Record{"candidate_id": "candidate_id_fixture", "field_names": []any{"full_name", "email_addresses"}},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handled, err := New().ExecuteWrite(context.Background(), action, tt.record, nil)
			if handled {
				t.Fatalf("ExecuteWrite handled = true, want declarative fallback")
			}
			if tt.wantErr && err == nil {
				t.Fatalf("ExecuteWrite error = nil, want error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("ExecuteWrite error = %v, want nil", err)
			}
		})
	}
}

func destroyOpeningsAction() engine.WriteAction {
	return engine.WriteAction{
		Name:       "destroy_openings",
		Method:     http.MethodDelete,
		Path:       "/v2/jobs/{{ record.job_id }}/openings",
		BodyFields: []string{"ids"},
	}
}

func runtimeForServer(baseURL string) *engine.Runtime {
	return &engine.Runtime{
		Requester: &connsdk.Requester{BaseURL: baseURL},
		Config:    connectors.RuntimeConfig{Config: map[string]string{}, Secrets: map[string]string{}},
	}
}
