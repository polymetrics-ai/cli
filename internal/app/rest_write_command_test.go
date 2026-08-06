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
	return setupRestWriteDemoAppWithBundle(t, ctx, baseURL, nil)
}

func setupRestWriteDemoAppWithBundle(t *testing.T, ctx context.Context, baseURL string, mutate func(*engine.Bundle)) *app.App {
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
	if mutate != nil {
		mutate(&bundle)
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

	a := setupRestWriteDemoAppWithBundle(t, ctx, server.URL, func(bundle *engine.Bundle) {
		bundle.CLISurface.Commands[0].RedactFields = []string{"id"}
	})
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
	if len(plan.Sample) != 1 || plan.Sample[0]["id"] != "t3_abc" {
		t.Fatalf("plan preview sample = %#v, want complete direct-write input", plan.Sample)
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
	if !ok || body["token"] != "fixture-token" {
		t.Fatalf("safe operation result = %#v, want complete token", run.OperationDirectWrite.Body)
	}
	if _, redacted := body["token_redacted"]; redacted {
		t.Fatalf("safe operation result = %#v, must not contain a redaction marker", run.OperationDirectWrite.Body)
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

func TestDirectWriteCommandHonorsDeclaredJSONAndNoneResponsePolicies(t *testing.T) {
	ctx := context.Background()
	for _, tt := range []struct {
		name     string
		policy   string
		wantBody bool
	}{
		{name: "json returns complete decoded body", policy: "json", wantBody: true},
		{name: "none intentionally suppresses response body", policy: "none"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.Method != http.MethodPatch || r.URL.Path != "/api/widgets/w_1" {
					t.Fatalf("request = %s %s, want PATCH /api/widgets/w_1", r.Method, r.URL.Path)
				}
				var body map[string]any
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatalf("decode request body: %v", err)
				}
				if body["name"] != "Ada" {
					t.Fatalf("request body = %#v, want typed name Ada", body)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"updated":true,"id":"w_1","nested":{"state":"complete"}}`))
			}))
			defer server.Close()

			a := setupRestWriteDemoAppWithBundle(t, ctx, server.URL, func(bundle *engine.Bundle) {
				for i := range bundle.Operations {
					if bundle.Operations[i].ID == "restwrite-demo.widget-update" {
						bundle.Operations[i].OutputPolicy = tt.policy
					}
				}
				for i := range bundle.CLISurface.Commands {
					if bundle.CLISurface.Commands[i].Path == "widget update" {
						bundle.CLISurface.Commands[i].OutputPolicy = tt.policy
					}
				}
			})
			plan, preview, err := a.PlanConnectorCommand(ctx, app.PlanConnectorCommandRequest{
				Connector:  restWriteDemoConnector,
				Credential: "restwrite-local",
				Path:       []string{"widget", "update"},
				Flags:      map[string][]string{"id": {"w_1"}, "name": {"Ada"}},
				Preview:    true,
			})
			if err != nil {
				t.Fatalf("PlanConnectorCommand: %v", err)
			}
			if preview == nil || preview.Digest == "" || plan.ApprovalToken == "" {
				t.Fatalf("preview/token = %#v/%q, want bound preview and approval token", preview, plan.ApprovalToken)
			}
			if calls != 0 {
				t.Fatalf("preview reached the network; calls = %d", calls)
			}

			run, err := a.RunReverseETL(ctx, app.RunReverseETLRequest{PlanID: plan.ID, ApprovalToken: plan.ApprovalToken})
			if err != nil {
				t.Fatalf("RunReverseETL: %v", err)
			}
			if calls != 1 || run.Status != "completed" || run.OperationDirectWrite == nil {
				t.Fatalf("run/calls = %#v/%d, want one completed direct write", run, calls)
			}
			if !tt.wantBody {
				if run.OperationDirectWrite.Body != nil {
					t.Fatalf("none policy body = %#v, want nil", run.OperationDirectWrite.Body)
				}
				t.Logf("direct-write command policy=%q status=%d response=<none>", tt.policy, run.OperationDirectWrite.Status)
				return
			}
			body, ok := run.OperationDirectWrite.Body.(map[string]any)
			if !ok {
				t.Fatalf("json policy body type = %T, want map", run.OperationDirectWrite.Body)
			}
			if body["id"] != "w_1" || body["updated"] != true {
				t.Fatalf("json policy body = %#v, want complete response fields", body)
			}
			nested, ok := body["nested"].(map[string]any)
			if !ok || nested["state"] != "complete" {
				t.Fatalf("json policy nested body = %#v, want complete nested response", body["nested"])
			}
			encoded, err := json.Marshal(body)
			if err != nil {
				t.Fatalf("marshal json policy response: %v", err)
			}
			t.Logf("direct-write command policy=%q status=%d response=%s", tt.policy, run.OperationDirectWrite.Status, encoded)
		})
	}
}

func TestDirectWriteCommandFailurePreservesErrorContent(t *testing.T) {
	ctx := context.Background()
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"fixture failure","token":"server-token"}`))
	}))
	defer server.Close()

	a := setupRestWriteDemoApp(t, ctx, server.URL)
	plan, _, err := a.PlanConnectorCommand(ctx, app.PlanConnectorCommandRequest{
		Connector:  restWriteDemoConnector,
		Credential: "restwrite-local",
		Path:       []string{"vote"},
		Flags:      map[string][]string{"id": {"t3_abc"}, "dir": {"1"}},
		Preview:    true,
	})
	if err != nil {
		t.Fatalf("PlanConnectorCommand: %v", err)
	}

	run, err := a.RunReverseETL(ctx, app.RunReverseETLRequest{
		PlanID:        plan.ID,
		ApprovalToken: plan.ApprovalToken,
		Confirmation:  connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive},
	})
	if err == nil {
		t.Fatal("RunReverseETL error = nil, want HTTP 500")
	}
	if calls != 1 || run.Status != "failed" {
		t.Fatalf("failed run/calls = %+v/%d, want one failed direct write", run, calls)
	}
	if !strings.Contains(err.Error(), "server-token") {
		t.Fatalf("RunReverseETL error = %q, want complete provider error content", err)
	}
	if !strings.Contains(run.Error, "server-token") {
		t.Fatalf("persisted direct-write error = %q, want complete provider error content", run.Error)
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
