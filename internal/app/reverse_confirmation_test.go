package app_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"polymetrics.ai/internal/app"
)

func TestRunReverseETLRejectsDestructiveConnectorCommandWithoutConfirmation(t *testing.T) {
	ctx := context.Background()
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	a, plan := setupGitHubDestructiveCommandPlan(t, ctx, server.URL)

	_, err := a.RunReverseETL(ctx, app.RunReverseETLRequest{
		PlanID:        plan.ID,
		ApprovalToken: plan.ApprovalToken,
	})
	if err == nil {
		t.Fatal("RunReverseETL() succeeded without destructive confirmation")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "confirmation") {
		t.Fatalf("RunReverseETL() error = %v, want confirmation rejection", err)
	}
	if calls != 0 {
		t.Fatalf("destructive write dispatched before confirmation gate; calls=%d", calls)
	}
}

func TestRunReverseETLAcceptsDestructiveConnectorCommandWithMatchingConfirmation(t *testing.T) {
	ctx := context.Background()
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Method != http.MethodDelete {
			t.Fatalf("request method = %s, want DELETE", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	a, plan := setupGitHubDestructiveCommandPlan(t, ctx, server.URL)
	if plan.ConfirmationChallenge != "destructive" {
		t.Fatalf("ConfirmationChallenge = %q, want destructive", plan.ConfirmationChallenge)
	}

	run, err := a.RunReverseETL(ctx, app.RunReverseETLRequest{
		PlanID:        plan.ID,
		ApprovalToken: plan.ApprovalToken,
		Confirmation:  "destructive",
	})
	if err != nil {
		t.Fatalf("RunReverseETL() with matching confirmation error = %v", err)
	}
	if run.RecordsSucceeded != 1 || run.RecordsFailed != 0 {
		t.Fatalf("run result = %+v, want one success", run)
	}
	if calls != 1 {
		t.Fatalf("destructive write call count = %d, want 1", calls)
	}
}

func TestRunReverseETLRejectsGenericDestructiveActionWithoutConfirmation(t *testing.T) {
	ctx := context.Background()
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	a, plan := setupGitHubGenericDestructivePlan(t, ctx, server.URL)
	if plan.ConfirmationChallenge != "destructive" {
		t.Fatalf("ConfirmationChallenge = %q, want destructive", plan.ConfirmationChallenge)
	}

	_, err := a.RunReverseETL(ctx, app.RunReverseETLRequest{
		PlanID:        plan.ID,
		ApprovalToken: plan.ApprovalToken,
	})
	if err == nil {
		t.Fatal("RunReverseETL() generic destructive action succeeded without confirmation")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "confirmation") {
		t.Fatalf("RunReverseETL() error = %v, want confirmation rejection", err)
	}
	if calls != 0 {
		t.Fatalf("generic destructive write dispatched before confirmation gate; calls=%d", calls)
	}
}

func setupGitHubDestructiveCommandPlan(t *testing.T, ctx context.Context, baseURL string) (*app.App, app.ReversePlan) {
	t.Helper()
	root := t.TempDir()
	if err := app.InitProject(root); err != nil {
		t.Fatalf("InitProject() error = %v", err)
	}
	a, err := app.Open(root)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	_, err = a.AddCredential(ctx, app.AddCredentialRequest{
		Name:      "github-local",
		Connector: "github",
		Config: map[string]string{
			"owner":         "acme",
			"repo":          "widgets",
			"public_access": "true",
			"base_url":      baseURL,
		},
	})
	if err != nil {
		t.Fatalf("AddCredential(github) error = %v", err)
	}
	plan, _, err := a.PlanConnectorCommand(ctx, app.PlanConnectorCommandRequest{
		Name:       "delete_repo",
		Connector:  "github",
		Credential: "github-local",
		Path:       []string{"repo", "delete-2"},
	})
	if err != nil {
		t.Fatalf("PlanConnectorCommand(repo delete-2) error = %v", err)
	}
	return a, plan
}

func setupGitHubGenericDestructivePlan(t *testing.T, ctx context.Context, baseURL string) (*app.App, app.ReversePlan) {
	t.Helper()
	root := t.TempDir()
	if err := app.InitProject(root); err != nil {
		t.Fatalf("InitProject() error = %v", err)
	}
	a, err := app.Open(root)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	_, err = a.AddCredential(ctx, app.AddCredentialRequest{
		Name:      "github-local",
		Connector: "github",
		Config: map[string]string{
			"owner":         "acme",
			"repo":          "widgets",
			"public_access": "true",
			"base_url":      baseURL,
		},
	})
	if err != nil {
		t.Fatalf("AddCredential(github) error = %v", err)
	}
	warehouseDir := filepath.Join(root, ".polymetrics", "warehouse")
	if err := os.MkdirAll(warehouseDir, 0o700); err != nil {
		t.Fatalf("mkdir warehouse: %v", err)
	}
	if err := os.WriteFile(filepath.Join(warehouseDir, "repo_deletes.jsonl"), []byte("{\"id\":\"row-1\"}\n"), 0o600); err != nil {
		t.Fatalf("write warehouse row: %v", err)
	}
	plan, err := a.PlanReverseETL(ctx, app.PlanReverseETLRequest{
		Name:                  "delete_repo",
		SourceTable:           "repo_deletes",
		DestinationConnector:  "github",
		DestinationCredential: "github-local",
		Action:                "repo",
		Mappings:              map[string]string{"id": "id"},
	})
	if err != nil {
		t.Fatalf("PlanReverseETL(repo) error = %v", err)
	}
	return a, plan
}
