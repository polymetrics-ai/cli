package app_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/warehouse"
)

type capturedBitbucketRequest struct {
	Method string
	Path   string
	Body   []byte
}

func TestLocalETLAndReverseETLWorkflow(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()

	if err := app.InitProject(root); err != nil {
		t.Fatalf("InitProject() error = %v", err)
	}

	a, err := app.Open(root)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	if _, err := a.AddCredential(ctx, app.AddCredentialRequest{
		Name:      "sample-local",
		Connector: "sample",
		Secrets:   map[string]string{"token": "sample-token"},
		Config:    map[string]string{"workspace": "local"},
	}); err != nil {
		t.Fatalf("AddCredential(sample) error = %v", err)
	}
	if _, err := a.AddCredential(ctx, app.AddCredentialRequest{
		Name:      "warehouse-local",
		Connector: "warehouse",
		Config:    map[string]string{"path": filepath.Join(root, ".polymetrics", "warehouse")},
	}); err != nil {
		t.Fatalf("AddCredential(warehouse) error = %v", err)
	}
	if _, err := a.AddCredential(ctx, app.AddCredentialRequest{
		Name:      "outbox-local",
		Connector: "outbox",
		Config:    map[string]string{"path": filepath.Join(root, ".polymetrics", "outbox")},
	}); err != nil {
		t.Fatalf("AddCredential(outbox) error = %v", err)
	}

	if _, err := a.CreateConnection(ctx, app.CreateConnectionRequest{
		Name: "sample_to_warehouse",
		Source: app.EndpointConfig{
			Connector:  "sample",
			Credential: "sample-local",
		},
		Destination: app.EndpointConfig{
			Connector:  "warehouse",
			Credential: "warehouse-local",
		},
		Streams: map[string]app.StreamConfig{
			"customers": {
				SyncMode:         "full_refresh_overwrite",
				CursorField:      "updated_at",
				PrimaryKey:       []string{"id"},
				DestinationTable: "sample_customers",
			},
		},
	}); err != nil {
		t.Fatalf("CreateConnection() error = %v", err)
	}

	run, err := a.RunETL(ctx, app.RunETLRequest{Connection: "sample_to_warehouse", Stream: "customers"})
	if err != nil {
		t.Fatalf("RunETL() error = %v", err)
	}
	if run.Status != "completed" || run.RecordsRead == 0 || run.RecordsLoaded != run.RecordsRead {
		t.Fatalf("unexpected ETL run: %+v", run)
	}

	rows, err := a.QueryTable(ctx, app.QueryTableRequest{Table: "sample_customers", Limit: 10})
	if err != nil {
		t.Fatalf("QueryTable() error = %v", err)
	}
	if len(rows) != 3 {
		t.Fatalf("QueryTable() returned %d rows, want 3", len(rows))
	}

	plan, err := a.PlanReverseETL(ctx, app.PlanReverseETLRequest{
		Name:                  "customers_to_outbox",
		SourceTable:           "sample_customers",
		DestinationConnector:  "outbox",
		DestinationCredential: "outbox-local",
		Action:                "upsert",
		Mappings: map[string]string{
			"id":    "external_id",
			"name":  "full_name",
			"email": "email",
		},
	})
	if err != nil {
		t.Fatalf("PlanReverseETL() error = %v", err)
	}
	if plan.RecordCount != 3 || plan.ApprovalToken == "" {
		t.Fatalf("unexpected reverse plan: %+v", plan)
	}

	reopened, err := app.Open(root)
	if err != nil {
		t.Fatalf("Open() after reverse plan error = %v", err)
	}

	reverseRun, err := reopened.RunReverseETL(ctx, app.RunReverseETLRequest{
		PlanID:        plan.ID,
		ApprovalToken: plan.ApprovalToken,
	})
	if err != nil {
		t.Fatalf("RunReverseETL() error = %v", err)
	}
	if reverseRun.Status != "completed" || reverseRun.RecordsSucceeded != 3 {
		t.Fatalf("unexpected reverse run: %+v", reverseRun)
	}
}

func TestAddCredentialRejectsInvalidGitHubBaseURLAtConfigurationTime(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()

	if err := app.InitProject(root); err != nil {
		t.Fatalf("InitProject() error = %v", err)
	}
	a, err := app.Open(root)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	_, err = a.AddCredential(ctx, app.AddCredentialRequest{
		Name:      "github-invalid-base-url",
		Connector: "github",
		Config:    map[string]string{"base_url": "not-a-uri"},
	})
	if err == nil {
		t.Fatal("AddCredential() accepted GitHub base_url that violates spec format uri")
	}
	if !strings.Contains(err.Error(), "base_url") || !strings.Contains(err.Error(), "format") {
		t.Fatalf("AddCredential() error = %q, want base_url and format", err)
	}
	if credentials := a.ListCredentials(); len(credentials) != 0 {
		t.Fatalf("ListCredentials() = %#v, want no persisted credential after validation failure", credentials)
	}
}

func TestAddCredentialRejectsDeclaredConfigurationConstraintsAtConfigurationTime(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()

	if err := app.InitProject(root); err != nil {
		t.Fatalf("InitProject() error = %v", err)
	}
	a, err := app.Open(root)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	tests := []struct {
		name      string
		connector string
		field     string
		value     string
		want      []string
	}{
		{
			name:      "date-time",
			connector: "github",
			field:     "since",
			value:     "not-a-date-time",
			want:      []string{"since", `configuration value must use "date-time" format`},
		},
		{
			name:      "date",
			connector: "google-search-console",
			field:     "start_date",
			value:     "2026-02-30",
			want:      []string{"start_date", `configuration value must use "date" format`},
		},
		{
			name:      "agilecrm pattern",
			connector: "agilecrm",
			field:     "domain",
			value:     "not.allowed",
			want:      []string{"domain", "pattern"},
		},
		{
			name:      "docker hub pattern",
			connector: "dockerhub",
			field:     "docker_username",
			value:     "Uppercase",
			want:      []string{"docker_username", "pattern"},
		},
		{
			name:      "engine connector enum",
			connector: "coin-api",
			field:     "environment",
			value:     "preview",
			want:      []string{"environment", "configuration value must be one of the declared values"},
		},
		{
			name:      "tier three base enum",
			connector: "postgres",
			field:     "mode",
			value:     "preview",
			want:      []string{"mode", "configuration value must be one of the declared values"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := a.AddCredential(ctx, app.AddCredentialRequest{
				Name:      "invalid-" + strings.ReplaceAll(tt.name, " ", "-"),
				Connector: tt.connector,
				Config:    map[string]string{tt.field: tt.value},
			})
			if err == nil {
				t.Fatalf("AddCredential(%s.%s) accepted %q", tt.connector, tt.field, tt.value)
			}
			for _, want := range tt.want {
				if !strings.Contains(err.Error(), want) {
					t.Fatalf("AddCredential(%s.%s) error = %q, want %q", tt.connector, tt.field, err, want)
				}
			}
			if strings.Contains(err.Error(), tt.value) {
				t.Fatalf("AddCredential(%s.%s) error unexpectedly echoes configuration value: %q", tt.connector, tt.field, err)
			}
			if credentials := a.ListCredentials(); len(credentials) != 0 {
				t.Fatalf("ListCredentials() = %#v, want no persisted credentials after validation failure", credentials)
			}
		})
	}
}

func TestAddCredentialLeavesConstraintFreeConnectorUnconstrained(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()

	if err := app.InitProject(root); err != nil {
		t.Fatalf("InitProject() error = %v", err)
	}
	a, err := app.Open(root)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	credential, err := a.AddCredential(ctx, app.AddCredentialRequest{
		Name:      "faker-unconstrained",
		Connector: "faker",
		Config: map[string]string{
			"count":     "not-a-number",
			"extra_key": "still accepted",
		},
	})
	if err != nil {
		t.Fatalf("AddCredential(faker) error = %v, want existing unconstrained behavior", err)
	}
	if credential.Connector != "faker" {
		t.Fatalf("AddCredential(faker).Connector = %q, want faker", credential.Connector)
	}
	if credentials := a.ListCredentials(); len(credentials) != 1 || credentials[0].Name != "faker-unconstrained" {
		t.Fatalf("ListCredentials() = %#v, want one unconstrained faker credential", credentials)
	}
}

func TestBitbucketReverseETLClosedSchemasDoNotReceiveInternalPlanFields(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()

	var mu sync.Mutex
	requests := make([]capturedBitbucketRequest, 0, 2)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		mu.Lock()
		requests = append(requests, capturedBitbucketRequest{Method: r.Method, Path: r.URL.Path, Body: body})
		mu.Unlock()

		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/repositories/ws/repo":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"uuid":"{repo-uuid}"}`))
		case r.Method == http.MethodDelete && r.URL.Path == "/repositories/ws/repo/issues/123":
			w.WriteHeader(http.StatusNoContent)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	if err := app.InitProject(root); err != nil {
		t.Fatalf("InitProject() error = %v", err)
	}
	a, err := app.Open(root)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if _, err := a.AddCredential(ctx, app.AddCredentialRequest{
		Name:      "bitbucket-local",
		Connector: "bitbucket",
		Config:    map[string]string{"base_url": srv.URL},
	}); err != nil {
		t.Fatalf("AddCredential(bitbucket) error = %v", err)
	}

	seedWarehouseTableRows(t, root, "bitbucket_repo_create", `{"workspace":"ws","repo_slug":"repo","scm":"git","is_private":true}`)
	createPlan, err := a.PlanReverseETL(ctx, app.PlanReverseETLRequest{
		Name:                  "bitbucket_repo_create",
		SourceTable:           "bitbucket_repo_create",
		DestinationConnector:  "bitbucket",
		DestinationCredential: "bitbucket-local",
		Action:                "create_repositories_workspace_repo_slug",
		Mappings: map[string]string{
			"workspace":  "workspace",
			"repo_slug":  "repo_slug",
			"scm":        "scm",
			"is_private": "is_private",
		},
	})
	if err != nil {
		t.Fatalf("PlanReverseETL(create repo) error = %v", err)
	}
	assertNoInternalReversePlanField(t, createPlan.Sample)
	createPreview, err := a.GetReversePlan(createPlan.ID)
	if err != nil {
		t.Fatalf("GetReversePlan(create repo) error = %v", err)
	}
	assertNoInternalReversePlanField(t, createPreview.Sample)
	createRun, err := a.RunReverseETL(ctx, app.RunReverseETLRequest{PlanID: createPlan.ID, ApprovalToken: createPlan.ApprovalToken})
	if err != nil {
		t.Fatalf("RunReverseETL(create repo) error = %v", err)
	}
	if createRun.Status != "completed" || createRun.RecordsSucceeded != 1 {
		t.Fatalf("unexpected create run: %+v", createRun)
	}

	seedWarehouseTableRows(t, root, "bitbucket_issue_delete", `{"workspace":"ws","repo_slug":"repo","issue_id":123}`)
	deletePlan, err := a.PlanReverseETL(ctx, app.PlanReverseETLRequest{
		Name:                  "bitbucket_issue_delete",
		SourceTable:           "bitbucket_issue_delete",
		DestinationConnector:  "bitbucket",
		DestinationCredential: "bitbucket-local",
		Action:                "delete_repositories_workspace_repo_slug_issues_issue_id",
		Mappings: map[string]string{
			"workspace": "workspace",
			"repo_slug": "repo_slug",
			"issue_id":  "issue_id",
		},
	})
	if err != nil {
		t.Fatalf("PlanReverseETL(delete issue) error = %v", err)
	}
	assertNoInternalReversePlanField(t, deletePlan.Sample)
	deletePreview, _, err := a.PreviewReversePlan(ctx, deletePlan.ID, nil)
	if err != nil {
		t.Fatalf("PreviewReversePlan(delete issue) error = %v", err)
	}
	assertNoInternalReversePlanField(t, deletePreview.Sample)
	deleteRun, err := a.RunReverseETL(ctx, app.RunReverseETLRequest{PlanID: deletePreview.ID, ApprovalToken: deletePreview.ApprovalToken, Confirmation: connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive}})
	if err != nil {
		t.Fatalf("RunReverseETL(delete issue) error = %v", err)
	}
	if deleteRun.Status != "completed" || deleteRun.RecordsSucceeded != 1 {
		t.Fatalf("unexpected delete run: %+v", deleteRun)
	}

	mu.Lock()
	captured := append([]capturedBitbucketRequest(nil), requests...)
	mu.Unlock()
	createReq, ok := findCapturedBitbucketRequest(captured, http.MethodPost, "/repositories/ws/repo")
	if !ok {
		t.Fatalf("missing create request: %+v", captured)
	}
	var createBody map[string]any
	if err := json.Unmarshal(createReq.Body, &createBody); err != nil {
		t.Fatalf("decode create request body: %v", err)
	}
	if _, ok := createBody["_polymetrics_reverse_plan_id"]; ok {
		t.Fatalf("create request body leaked internal plan field: %v", createBody)
	}
	if _, ok := createBody["workspace"]; ok {
		t.Fatalf("create request body leaked path field workspace: %v", createBody)
	}
	if _, ok := createBody["repo_slug"]; ok {
		t.Fatalf("create request body leaked path field repo_slug: %v", createBody)
	}
	if createBody["scm"] != "git" || createBody["is_private"] != true {
		t.Fatalf("unexpected create request body: %v", createBody)
	}
	deleteReq, ok := findCapturedBitbucketRequest(captured, http.MethodDelete, "/repositories/ws/repo/issues/123")
	if !ok {
		t.Fatalf("missing delete request: %+v", captured)
	}
	if strings.TrimSpace(string(deleteReq.Body)) != "" {
		t.Fatalf("delete request body = %q, want empty", string(deleteReq.Body))
	}
}

// seedWarehouseTableRows materializes an unattributed root-level table from
// JSON row literals. It writes through the real Parquet writer rather than
// hand-rolling the file, so a test fixture can never drift from the format a
// sync produces.
func seedWarehouseTableRows(t *testing.T, root, table string, rows ...string) {
	t.Helper()
	decoded := make([]warehouse.Row, 0, len(rows))
	for _, row := range rows {
		var record warehouse.Row
		if err := json.Unmarshal([]byte(row), &record); err != nil {
			t.Fatalf("decode seed row %s: %v", row, err)
		}
		decoded = append(decoded, record)
	}
	path := filepath.Join(root, ".polymetrics", "warehouse", table+warehouse.TableFileExt)
	if err := warehouse.WriteTable(context.Background(), path, decoded); err != nil {
		t.Fatalf("seed warehouse table %s: %v", table, err)
	}
}

func assertNoInternalReversePlanField(t *testing.T, records []connectors.Record) {
	t.Helper()
	for _, record := range records {
		if _, ok := record["_polymetrics_reverse_plan_id"]; ok {
			t.Fatalf("sample leaked internal plan field: %v", record)
		}
	}
}

func findCapturedBitbucketRequest(requests []capturedBitbucketRequest, method, path string) (capturedBitbucketRequest, bool) {
	for _, req := range requests {
		if req.Method == method && req.Path == path {
			return req, true
		}
	}
	return capturedBitbucketRequest{}, false
}
