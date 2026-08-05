package app_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
)

const restWriteDemoConnector = "restwrite-demo"

func setupRestWriteDemoApp(t *testing.T, ctx context.Context, baseURL string) *app.App {
	t.Helper()
	root := t.TempDir()
	if err := app.InitProject(root); err != nil {
		t.Fatalf("InitProject: %v", err)
	}
	a, err := app.Open(root)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	bundle, err := engine.Load(os.DirFS("testdata/bundles"), restWriteDemoConnector)
	if err != nil {
		t.Fatalf("engine.Load(%s): %v", restWriteDemoConnector, err)
	}
	a.Registry().Register(engine.New(bundle, nil))
	if _, err := a.AddCredential(ctx, app.AddCredentialRequest{
		Name:      "restwrite-local",
		Connector: restWriteDemoConnector,
		Config:    map[string]string{"base_url": baseURL},
	}); err != nil {
		t.Fatalf("AddCredential: %v", err)
	}
	return a
}

func TestDirectWriteCommandPlanPreviewApprovalAndExecute(t *testing.T) {
	ctx := context.Background()
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Method != http.MethodPost || r.URL.Path != "/api/vote" {
			t.Fatalf("request = %s %s, want POST /api/vote", r.Method, r.URL.Path)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm: %v", err)
		}
		if got := r.Form.Get("id"); got != "t3_abc" {
			t.Fatalf("id = %q, want t3_abc", got)
		}
		if got := r.Form.Get("dir"); got != "1" {
			t.Fatalf("dir = %q, want 1", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true,"token":"fixture-token"}`))
	}))
	defer server.Close()

	a := setupRestWriteDemoApp(t, ctx, server.URL)
	plan, preview, err := a.PlanConnectorCommand(ctx, app.PlanConnectorCommandRequest{
		Connector:  restWriteDemoConnector,
		Credential: "restwrite-local",
		Path:       []string{"vote"},
		Flags:      map[string][]string{"id": {"t3_abc"}, "dir": {"1"}},
		Preview:    true,
	})
	if err != nil {
		t.Fatalf("PlanConnectorCommand: %v", err)
	}
	if preview == nil || preview.Digest == "" {
		t.Fatalf("preview = %#v, want a bound no-network preview", preview)
	}
	if calls != 0 {
		t.Fatalf("planning preview reached the network; calls = %d", calls)
	}
	if plan.ConnectorCommandOperation != "restwrite-demo.vote" {
		t.Fatalf("plan operation = %q, want restwrite-demo.vote", plan.ConnectorCommandOperation)
	}
	if plan.ApprovalToken == "" || plan.ConfirmationChallenge != "destructive" {
		t.Fatalf("plan approval/confirmation = %q/%q, want a single-use token and destructive confirmation", plan.ApprovalToken, plan.ConfirmationChallenge)
	}
	if plan.PlanSeal == nil || plan.PlanSeal.Batchable {
		t.Fatalf("plan seal batchable = %#v, want batchable:false from the operation declaration", plan.PlanSeal)
	}

	if _, err := a.RunReverseETL(ctx, app.RunReverseETLRequest{PlanID: plan.ID, ApprovalToken: plan.ApprovalToken}); err == nil {
		t.Fatal("RunReverseETL dispatched a direct write without typed confirmation")
	}
	if calls != 0 {
		t.Fatalf("unconfirmed direct write reached the network; calls = %d", calls)
	}

	run, err := a.RunReverseETL(ctx, app.RunReverseETLRequest{
		PlanID:        plan.ID,
		ApprovalToken: plan.ApprovalToken,
		Confirmation:  connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive},
	})
	if err != nil {
		t.Fatalf("RunReverseETL: %v", err)
	}
	if calls != 1 || run.Status != "completed" || run.RecordsSucceeded != 1 {
		t.Fatalf("run/calls = %+v/%d, want one completed write", run, calls)
	}
	if run.OperationDirectWrite == nil {
		t.Fatal("run direct_write result = nil")
	}
	body, ok := run.OperationDirectWrite.Body.(map[string]any)
	if !ok || body["token_redacted"] != true {
		t.Fatalf("safe operation result = %#v, want token redaction", run.OperationDirectWrite.Body)
	}

	if _, err := a.RunReverseETL(ctx, app.RunReverseETLRequest{
		PlanID:        plan.ID,
		ApprovalToken: plan.ApprovalToken,
		Confirmation:  connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive},
	}); err == nil || (!strings.Contains(strings.ToLower(err.Error()), "consumed") && !strings.Contains(strings.ToLower(err.Error()), "already executed")) {
		t.Fatalf("replayed approval error = %v, want a single-use approval rejection", err)
	}
	if calls != 1 {
		t.Fatalf("replayed approval reached the network; calls = %d", calls)
	}
}

func TestNonDestructiveDirectWriteStillRequiresPreviewAndSingleUseApproval(t *testing.T) {
	ctx := context.Background()
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Method != http.MethodPatch || r.URL.Path != "/api/widgets/w_1" {
			t.Fatalf("request = %s %s, want PATCH /api/widgets/w_1", r.Method, r.URL.Path)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if body["name"] != "Ada" {
			t.Fatalf("body = %#v, want typed name Ada", body)
		}
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	a := setupRestWriteDemoApp(t, ctx, server.URL)
	plan, preview, err := a.PlanConnectorCommand(ctx, app.PlanConnectorCommandRequest{
		Connector:  restWriteDemoConnector,
		Credential: "restwrite-local",
		Path:       []string{"widget", "update"},
		Flags:      map[string][]string{"id": {"w_1"}, "name": {"Ada"}},
	})
	if err != nil {
		t.Fatalf("PlanConnectorCommand: %v", err)
	}
	if preview != nil || plan.ApprovalToken != "" || calls != 0 {
		t.Fatalf("unpreviewed safe write plan = preview %#v token %q calls %d, want no preview/token/network", preview, plan.ApprovalToken, calls)
	}
	if _, err := a.RunReverseETL(ctx, app.RunReverseETLRequest{PlanID: plan.ID, ApprovalToken: "wrong"}); err == nil || !strings.Contains(err.Error(), "must be previewed") {
		t.Fatalf("RunReverseETL before preview error = %v, want preview requirement", err)
	}

	plan, previewResult, err := a.PreviewConnectorCommandPlan(ctx, plan.ID)
	if err != nil {
		t.Fatalf("PreviewConnectorCommandPlan: %v", err)
	}
	if previewResult.Digest == "" || plan.ApprovalToken == "" || calls != 0 {
		t.Fatalf("safe direct-write preview = %#v token %q calls %d, want digest/token/no network", previewResult, plan.ApprovalToken, calls)
	}

	run, err := a.RunReverseETL(ctx, app.RunReverseETLRequest{PlanID: plan.ID, ApprovalToken: plan.ApprovalToken})
	if err != nil {
		t.Fatalf("RunReverseETL: %v", err)
	}
	if run.Status != "completed" || run.OperationDirectWrite == nil || calls != 1 {
		t.Fatalf("safe run = %#v calls %d, want one completed operation direct write", run, calls)
	}
	if _, err := a.RunReverseETL(ctx, app.RunReverseETLRequest{PlanID: plan.ID, ApprovalToken: plan.ApprovalToken}); err == nil || (!strings.Contains(strings.ToLower(err.Error()), "consumed") && !strings.Contains(strings.ToLower(err.Error()), "already executed")) {
		t.Fatalf("replayed safe approval error = %v, want consumption rejection", err)
	}
	if calls != 1 {
		t.Fatalf("replayed safe approval reached the network; calls = %d", calls)
	}
}
