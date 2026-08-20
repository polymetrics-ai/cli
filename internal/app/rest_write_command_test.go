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
	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/commandrunner"
	"polymetrics.ai/internal/connectors/engine"
)

const (
	restWriteDemoConnector          = "restwrite-demo"
	multipartRestWriteDemoConnector = "multipart-restwrite-demo"
)

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

// setupMultipartRestWriteDemoApp installs a fully declared in-memory bundle so
// this lifecycle test reaches commandrunner.Preflight and the engine without
// adding a fixture command to a provider-facing cli_surface.json.
func setupMultipartRestWriteDemoApp(t *testing.T, ctx context.Context, baseURL string) *app.App {
	return setupMultipartRestWriteDemoAppWithBundle(t, ctx, baseURL, multipartRestWriteDemoBundle())
}

func setupMultipartRestWriteDemoAppWithBundle(t *testing.T, ctx context.Context, baseURL string, bundle engine.Bundle) *app.App {
	t.Helper()
	root := t.TempDir()
	if err := app.InitProject(root); err != nil {
		t.Fatalf("InitProject: %v", err)
	}
	a, err := app.Open(root)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	a.Registry().Register(engine.New(bundle, nil))
	if _, err := a.AddCredential(ctx, app.AddCredentialRequest{
		Name:      "multipart-restwrite-local",
		Connector: multipartRestWriteDemoConnector,
		Config:    map[string]string{"base_url": baseURL},
	}); err != nil {
		t.Fatalf("AddCredential: %v", err)
	}
	return a
}

func multipartRestWriteDemoBundle() engine.Bundle {
	nonBatchable := false
	return engine.Bundle{
		Name: multipartRestWriteDemoConnector,
		Metadata: engine.Metadata{
			Name:            multipartRestWriteDemoConnector,
			DisplayName:     "Multipart REST write test fixture",
			IntegrationType: "api",
			ReleaseStage:    "alpha",
			Capabilities:    engine.Capabilities{Check: true, Write: true},
		},
		HTTP: engine.HTTPBase{URL: "{{ config.base_url }}"},
		Operations: []engine.OperationSpec{{
			ID:            multipartRestWriteDemoConnector + ".attachment-create",
			Kind:          "rest_write",
			Summary:       "Attach one fixture file.",
			Risk:          "high",
			Approval:      "plan-preview-confirm-execute",
			OutputPolicy:  "json",
			MutationClass: "destructive",
			Confirmation:  &engine.ConfirmationSpec{Kind: connectors.ConfirmationKindDestructive},
			Batchable:     &nonBatchable,
			REST: &engine.RESTOperationSpec{
				Method:      http.MethodPost,
				Path:        "/api/attachments",
				ContentType: "multipart/form-data",
				MaxBytes:    1024,
				BodySchema: json.RawMessage(`{
					"type": "object",
					"additionalProperties": false,
					"required": ["message", "media_file_path"],
					"properties": {
						"message": {"type": "string"},
						"media_file_path": {"type": "string"}
					}
				}`),
				Multipart: &engine.MultipartSpec{
					MaxBytes: 1024,
					Parts: []engine.MultipartPartSpec{
						{Name: "message", Type: "field", Field: "message", Required: true},
						{
							Name:              "attachment",
							Type:              "file",
							Field:             "media_file_path",
							Required:          true,
							MaxBytes:          1024,
							ContentType:       "text/plain",
							AllowedMediaTypes: []string{"text/plain"},
						},
					},
				},
			},
		}},
		Surface: &engine.APISurface{
			API: "multipart-restwrite-demo fixture",
			Endpoints: []engine.SurfaceEndpoint{{
				Method: http.MethodPost,
				Path:   "/api/attachments",
				Operation: &engine.SurfaceOperation{
					Model:            "destructive_action",
					Status:           "blocked",
					Risk:             "high",
					BlockedByDefault: true,
					Reason:           "Bound by the typed multipart rest_write executor.",
				},
			}},
		},
		CLISurface: &engine.CLISurface{
			Tagline: "Multipart REST write test fixture command.",
			Usage:   "pm multipart-restwrite-demo attachment create [flags]",
			Commands: []engine.CLICommand{{
				Path:         "attachment create",
				Summary:      "Attach one fixture file.",
				Intent:       "direct_write",
				Availability: "implemented",
				Operation:    multipartRestWriteDemoConnector + ".attachment-create",
				APISurface: []engine.CLISurfaceEndpointRef{{
					Method: http.MethodPost,
					Path:   "/api/attachments",
				}},
				OutputPolicy: "json",
				Risk:         "Uploads one bounded project-local fixture file.",
				Approval:     "Requires plan -> no-network preview -> explicit single-use approval -> execute.",
				Flags: []engine.CLIFlag{
					{Name: "message", Type: "string", Summary: "Attachment message.", MapsTo: "body.message", Required: true},
					{Name: "media-file-path", Type: "string", Summary: "Project-relative file path.", MapsTo: "body.media_file_path", Required: true},
				},
			}},
		},
	}
}

func multipartRestWriteDemoBundleWithUploadField() engine.Bundle {
	bundle := multipartRestWriteDemoBundle()
	operation := &bundle.Operations[0]
	operation.REST.BodySchema = json.RawMessage(`{
		"type": "object",
		"additionalProperties": false,
		"required": ["message", "upload"],
		"properties": {
			"message": {"type": "string"},
			"upload": {"type": "string"}
		}
	}`)
	operation.REST.Multipart.Parts[1].Field = "upload"
	bundle.CLISurface.Commands[0].Flags[1] = engine.CLIFlag{
		Name: "upload", Type: "string", Summary: "Project-relative upload path.", MapsTo: "body.upload", Required: true,
	}
	return bundle
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
		if got := r.Header.Get("X-Change-Reason"); got != "correctness" {
			t.Fatalf("X-Change-Reason = %q, want declaration-owned header", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true,"token":"fixture-token"}`))
	}))
	defer server.Close()

	a := setupRestWriteDemoAppWithBundle(t, ctx, server.URL, func(bundle *engine.Bundle) {
		bundle.CLISurface.Commands[0].RedactFields = []string{"id"}
		bundle.Operations[0].REST.Parameters = []engine.OperationParameter{{
			Name: "X-Change-Reason", In: "header", Type: "string", Required: true,
			Values: []string{"correctness", "moderation"}, Schema: json.RawMessage(`{"type":"string","enum":["correctness","moderation"]}`), MaxBytes: 32,
		}}
		bundle.CLISurface.Commands[0].Flags = append(bundle.CLISurface.Commands[0].Flags,
			engine.CLIFlag{Name: "header-x-change-reason", Type: "enum", Values: []string{"correctness", "moderation"}, Required: true, MapsTo: "header.X-Change-Reason"},
		)
	})
	plan, preview, err := a.PlanConnectorCommand(ctx, app.PlanConnectorCommandRequest{
		Connector:  restWriteDemoConnector,
		Credential: "restwrite-local",
		Path:       []string{"vote"},
		Flags:      map[string][]string{"id": {"t3_abc"}, "dir": {"1"}, "header-x-change-reason": {"correctness"}},
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
	if got := plan.ConnectorCommandHeaders["X-Change-Reason"]; got != "correctness" {
		t.Fatalf("plan headers = %#v, want preview-bound declared header", plan.ConnectorCommandHeaders)
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

func TestMultipartDirectWriteCommandPreflightPlanPreviewApprovalAndExecute(t *testing.T) {
	ctx := context.Background()
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Method != http.MethodPost || r.URL.Path != "/api/attachments" {
			t.Fatalf("request = %s %s, want POST /api/attachments", r.Method, r.URL.Path)
		}
		if !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data; boundary=") {
			t.Fatalf("Content-Type = %q, want multipart boundary", r.Header.Get("Content-Type"))
		}
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Fatalf("ParseMultipartForm: %v", err)
		}
		if got := r.MultipartForm.Value["message"]; len(got) != 1 || got[0] != "fixture attachment" {
			t.Fatalf("multipart message = %#v, want declared command field", got)
		}
		files := r.MultipartForm.File["attachment"]
		if len(files) != 1 || files[0].Filename != "attachment.txt" {
			t.Fatalf("attachment files = %#v, want attachment.txt", files)
		}
		if !strings.HasPrefix(files[0].Header.Get("Content-Type"), "text/plain") {
			t.Fatalf("attachment content type = %q, want text/plain", files[0].Header.Get("Content-Type"))
		}
		file, err := files[0].Open()
		if err != nil {
			t.Fatalf("Open attachment: %v", err)
		}
		defer func() { _ = file.Close() }()
		body, err := io.ReadAll(file)
		if err != nil {
			t.Fatalf("ReadAll attachment: %v", err)
		}
		if string(body) != "fixture attachment bytes" {
			t.Fatalf("attachment bytes = %q", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true,"server_value":"complete"}`))
	}))
	defer server.Close()

	a := setupMultipartRestWriteDemoApp(t, ctx, server.URL)
	if err := os.WriteFile(filepath.Join(a.ProjectDir(), "attachment.txt"), []byte("fixture attachment bytes"), 0o600); err != nil {
		t.Fatalf("WriteFile attachment: %v", err)
	}
	connector, ok := a.Registry().Get(multipartRestWriteDemoConnector)
	if !ok {
		t.Fatalf("registry missing %q", multipartRestWriteDemoConnector)
	}
	if err := commandrunner.Preflight(connector, []string{"attachment", "create"}); err != nil {
		t.Fatalf("Preflight multipart direct_write: %v", err)
	}

	plan, preview, err := a.PlanConnectorCommand(ctx, app.PlanConnectorCommandRequest{
		Connector:  multipartRestWriteDemoConnector,
		Credential: "multipart-restwrite-local",
		Path:       []string{"attachment", "create"},
		Flags: map[string][]string{
			"message":         {"fixture attachment"},
			"media-file-path": {"attachment.txt"},
		},
		Preview: true,
	})
	if err != nil {
		t.Fatalf("PlanConnectorCommand: %v", err)
	}
	if preview == nil || preview.Digest == "" || len(plan.PayloadIdentity) != 1 || plan.PayloadIdentity[0].ContentSHA256 == "" {
		t.Fatalf("multipart plan/preview = %#v/%#v, want one digest-bound payload preview", plan, preview)
	}
	if calls != 0 {
		t.Fatalf("multipart plan preview calls = %d, want 0", calls)
	}

	if _, err := a.RunReverseETL(ctx, app.RunReverseETLRequest{PlanID: plan.ID, ApprovalToken: plan.ApprovalToken}); err == nil || !strings.Contains(strings.ToLower(err.Error()), "confirmation") {
		t.Fatalf("RunReverseETL without confirmation = %v, want confirmation rejection", err)
	}
	if calls != 0 {
		t.Fatalf("unconfirmed multipart calls = %d, want 0", calls)
	}

	run, err := a.RunReverseETL(ctx, app.RunReverseETLRequest{
		PlanID:        plan.ID,
		ApprovalToken: plan.ApprovalToken,
		Confirmation:  connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive},
	})
	if err != nil {
		t.Fatalf("RunReverseETL approved multipart: %v", err)
	}
	if run.Status != "completed" || run.OperationDirectWrite == nil || calls != 1 {
		t.Fatalf("multipart run = %#v calls = %d, want one completed direct write", run, calls)
	}
	resultBody, ok := run.OperationDirectWrite.Body.(map[string]any)
	if !ok || resultBody["server_value"] != "complete" {
		t.Fatalf("multipart result body = %#v, want complete provider output", run.OperationDirectWrite.Body)
	}

	if _, err := a.RunReverseETL(ctx, app.RunReverseETLRequest{
		PlanID:        plan.ID,
		ApprovalToken: plan.ApprovalToken,
		Confirmation:  connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive},
	}); err == nil || (!strings.Contains(strings.ToLower(err.Error()), "consumed") && !strings.Contains(strings.ToLower(err.Error()), "already executed")) {
		t.Fatalf("RunReverseETL replayed multipart approval = %v, want single-use rejection", err)
	}
	if calls != 1 {
		t.Fatalf("replayed multipart calls = %d, want 1", calls)
	}
}

func TestMultipartDirectWriteCommandRejectsChangedPayloadBeforeNetwork(t *testing.T) {
	ctx := context.Background()
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		calls++
	}))
	defer server.Close()

	a := setupMultipartRestWriteDemoApp(t, ctx, server.URL)
	attachment := filepath.Join(a.ProjectDir(), "attachment.txt")
	if err := os.WriteFile(attachment, []byte("approved bytes"), 0o600); err != nil {
		t.Fatalf("WriteFile approved attachment: %v", err)
	}
	plan, _, err := a.PlanConnectorCommand(ctx, app.PlanConnectorCommandRequest{
		Connector:  multipartRestWriteDemoConnector,
		Credential: "multipart-restwrite-local",
		Path:       []string{"attachment", "create"},
		Flags: map[string][]string{
			"message":         {"fixture attachment"},
			"media-file-path": {"attachment.txt"},
		},
		Preview: true,
	})
	if err != nil {
		t.Fatalf("PlanConnectorCommand: %v", err)
	}
	if err := os.WriteFile(attachment, []byte("changed bytes after preview"), 0o600); err != nil {
		t.Fatalf("WriteFile changed attachment: %v", err)
	}
	if _, err := a.RunReverseETL(ctx, app.RunReverseETLRequest{
		PlanID:        plan.ID,
		ApprovalToken: plan.ApprovalToken,
		Confirmation:  connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive},
	}); err == nil || !strings.Contains(err.Error(), "payload changed") {
		t.Fatalf("RunReverseETL changed multipart payload = %v, want plan payload rejection", err)
	}
	if calls != 0 {
		t.Fatalf("changed multipart payload calls = %d, want 0", calls)
	}
}

func TestMultipartDirectWriteCommandBindsDeclaredUploadField(t *testing.T) {
	ctx := context.Background()
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.URL.Path != "/api/attachments" {
			t.Fatalf("request path = %q, want /api/attachments", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	a := setupMultipartRestWriteDemoAppWithBundle(t, ctx, server.URL, multipartRestWriteDemoBundleWithUploadField())
	if err := os.WriteFile(filepath.Join(a.ProjectDir(), "attachment.txt"), []byte("fixture attachment bytes"), 0o600); err != nil {
		t.Fatalf("WriteFile attachment: %v", err)
	}
	plan, preview, err := a.PlanConnectorCommand(ctx, app.PlanConnectorCommandRequest{
		Connector:  multipartRestWriteDemoConnector,
		Credential: "multipart-restwrite-local",
		Path:       []string{"attachment", "create"},
		Flags: map[string][]string{
			"message": {"fixture attachment"},
			"upload":  {"attachment.txt"},
		},
		Preview: true,
	})
	if err != nil {
		t.Fatalf("PlanConnectorCommand declared upload field: %v", err)
	}
	if preview == nil || len(plan.PayloadIdentity) != 1 || plan.PayloadIdentity[0].Field != "upload" || plan.PayloadIdentity[0].ContentSHA256 == "" {
		t.Fatalf("declared upload payload identity = %#v / %#v, want the exact declared field and digest", plan, preview)
	}
	if calls != 0 {
		t.Fatalf("declared upload preview calls = %d, want 0", calls)
	}
	if _, err := a.RunReverseETL(ctx, app.RunReverseETLRequest{
		PlanID:        plan.ID,
		ApprovalToken: plan.ApprovalToken,
		Confirmation:  connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive},
	}); err != nil {
		t.Fatalf("RunReverseETL declared upload field: %v", err)
	}
	if calls != 1 {
		t.Fatalf("declared upload execution calls = %d, want 1", calls)
	}
}

func TestMultipartDirectWritePreflightRejectsMissingContractAndLegacyFileUpload(t *testing.T) {
	operation := func(t *testing.T, bundle *engine.Bundle) *engine.OperationSpec {
		t.Helper()
		for index := range bundle.Operations {
			if bundle.Operations[index].ID == multipartRestWriteDemoConnector+".attachment-create" {
				return &bundle.Operations[index]
			}
		}
		t.Fatal("multipart fixture operation is missing")
		return nil
	}

	t.Run("missing multipart contract remains blocked", func(t *testing.T) {
		bundle := multipartRestWriteDemoBundle()
		operation(t, &bundle).REST.Multipart = nil
		err := commandrunner.Preflight(engine.New(bundle, nil), []string{"attachment", "create"})
		if err == nil || !strings.Contains(err.Error(), "not executable") {
			t.Fatalf("Preflight missing multipart contract = %v, want executable-claim rejection", err)
		}
	})

	t.Run("missing api surface endpoint remains blocked", func(t *testing.T) {
		bundle := multipartRestWriteDemoBundle()
		bundle.Surface = nil
		err := commandrunner.Preflight(engine.New(bundle, nil), []string{"attachment", "create"})
		if err == nil || !strings.Contains(err.Error(), "api_surface") {
			t.Fatalf("Preflight missing api_surface = %v, want endpoint provenance rejection", err)
		}
	})

	t.Run("legacy file upload cannot become direct write", func(t *testing.T) {
		bundle := multipartRestWriteDemoBundle()
		operation(t, &bundle).Kind = "file_upload"
		err := commandrunner.Preflight(engine.New(bundle, nil), []string{"attachment", "create"})
		if err == nil || !strings.Contains(err.Error(), "requires rest_write") {
			t.Fatalf("Preflight legacy file_upload = %v, want rest_write executor rejection", err)
		}
	})
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

	plan, previewResult, err := a.PreviewConnectorCommandPlan(ctx, plan.ID, nil)
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
