package app_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"polymetrics.ai/internal/app"
)

func TestReversePlanRedactsBitbucketSensitiveWriteSample(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	sensitiveValue := "sensitive-fixture-value"
	var executedBody string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/repositories/ws/repo/pipelines_config/variables" {
			t.Errorf("path = %s", r.URL.Path)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read body: %v", err)
		}
		executedBody = string(body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"uuid":"var-1"}`))
	}))
	defer srv.Close()

	if err := app.InitProject(root); err != nil {
		t.Fatalf("InitProject() error = %v", err)
	}
	warehouseDir := filepath.Join(root, ".polymetrics", "warehouse")
	if err := os.MkdirAll(warehouseDir, 0o700); err != nil {
		t.Fatalf("MkdirAll(warehouse) error = %v", err)
	}
	row := `{"workspace":"ws","repo_slug":"repo","key":"DEPLOY_TOKEN","value":"` + sensitiveValue + `"}` + "\n"
	if err := os.WriteFile(filepath.Join(warehouseDir, "bitbucket_variables.jsonl"), []byte(row), 0o600); err != nil {
		t.Fatalf("write warehouse row: %v", err)
	}

	a, err := app.Open(root)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if _, err := a.AddCredential(ctx, app.AddCredentialRequest{
		Name:      "bitbucket-local",
		Connector: "bitbucket",
		Config:    map[string]string{"base_url": srv.URL},
		Secrets:   map[string]string{"access_token": "fixture-access-token"},
	}); err != nil {
		t.Fatalf("AddCredential(bitbucket) error = %v", err)
	}

	plan, err := a.PlanReverseETL(ctx, app.PlanReverseETLRequest{
		Name:                  "bitbucket_variable_create",
		SourceTable:           "bitbucket_variables",
		DestinationConnector:  "bitbucket",
		DestinationCredential: "bitbucket-local",
		Action:                "create_repositories_workspace_repo_slug_pipelines_config_variables",
		Mappings: map[string]string{
			"workspace": "workspace",
			"repo_slug": "repo_slug",
			"key":       "key",
			"value":     "value",
		},
	})
	if err != nil {
		t.Fatalf("PlanReverseETL() error = %v", err)
	}
	if len(plan.Sample) != 1 || plan.Sample[0]["value"] != "redacted" {
		t.Fatalf("plan sample = %+v, want redacted value", plan.Sample)
	}
	if plan.Sample[0]["key"] != "DEPLOY_TOKEN" {
		t.Fatalf("plan sample key = %v", plan.Sample[0]["key"])
	}
	encodedPlan, err := json.Marshal(plan)
	if err != nil {
		t.Fatalf("marshal plan: %v", err)
	}
	if strings.Contains(string(encodedPlan), sensitiveValue) {
		t.Fatalf("plan JSON leaked sensitive value: %s", encodedPlan)
	}

	reopened, err := app.Open(root)
	if err != nil {
		t.Fatalf("Open(reopened) error = %v", err)
	}
	stored, err := reopened.GetReversePlan(plan.ID)
	if err != nil {
		t.Fatalf("GetReversePlan() error = %v", err)
	}
	encodedStored, err := json.Marshal(stored)
	if err != nil {
		t.Fatalf("marshal stored plan: %v", err)
	}
	if strings.Contains(string(encodedStored), sensitiveValue) {
		t.Fatalf("stored plan JSON leaked sensitive value: %s", encodedStored)
	}

	run, err := reopened.RunReverseETL(ctx, app.RunReverseETLRequest{
		PlanID:        plan.ID,
		ApprovalToken: plan.ApprovalToken,
	})
	if err != nil {
		t.Fatalf("RunReverseETL() error = %v", err)
	}
	if run.Status != "completed" || !strings.Contains(executedBody, sensitiveValue) {
		t.Fatalf("run = %+v executedBody = %s", run, executedBody)
	}
}
