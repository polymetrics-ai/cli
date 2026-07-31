package app_test

import (
	"context"
	"encoding/json"
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

	if err := app.InitProject(root); err != nil {
		t.Fatalf("InitProject() error = %v", err)
	}
	warehouseDir := filepath.Join(root, ".polymetrics", "warehouse")
	if err := os.MkdirAll(warehouseDir, 0o700); err != nil {
		t.Fatalf("MkdirAll(warehouse) error = %v", err)
	}
	row := `{"workspace":"ws","repo_slug":"repo","scm":"git","is_private":true,"description":"` + sensitiveValue + `"}` + "\n"
	if err := os.WriteFile(filepath.Join(warehouseDir, "bitbucket_repositories.jsonl"), []byte(row), 0o600); err != nil {
		t.Fatalf("write warehouse row: %v", err)
	}

	a, err := app.Open(root)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if _, err := a.AddCredential(ctx, app.AddCredentialRequest{
		Name:      "bitbucket-local",
		Connector: "bitbucket",
		Config:    map[string]string{"base_url": "https://bitbucket.invalid"},
		Secrets:   map[string]string{"access_token": "fixture-access-token"},
	}); err != nil {
		t.Fatalf("AddCredential(bitbucket) error = %v", err)
	}

	plan, err := a.PlanReverseETL(ctx, app.PlanReverseETLRequest{
		Name:                  "bitbucket_repository_create",
		SourceTable:           "bitbucket_repositories",
		DestinationConnector:  "bitbucket",
		DestinationCredential: "bitbucket-local",
		Action:                "create_repositories_workspace_repo_slug",
		Mappings: map[string]string{
			"workspace":   "workspace",
			"repo_slug":   "repo_slug",
			"scm":         "scm",
			"is_private":  "is_private",
			"description": "description",
		},
	})
	if err != nil {
		t.Fatalf("PlanReverseETL() error = %v", err)
	}
	if len(plan.Sample) != 1 || plan.Sample[0]["description"] != "redacted" {
		t.Fatalf("plan sample = %+v, want redacted description", plan.Sample)
	}
	if plan.Sample[0]["repo_slug"] != "repo" {
		t.Fatalf("plan sample repo_slug = %v", plan.Sample[0]["repo_slug"])
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
}
