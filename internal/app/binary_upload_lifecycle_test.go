package app_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/connectors/engine"
)

// TestBinaryUploadConnectorCommandPersistsPreviewBeforeApproval exercises the
// public connector-command route through App, rather than treating a command
// declaration or engine-only byte test as proof. It pins four contracts in
// one real path: project-root source resolution, no path at rest, no approval
// before persisted preview, and exact approved bytes on the only provider call.
func TestBinaryUploadConnectorCommandPersistsPreviewBeforeApproval(t *testing.T) {
	ctx := context.Background()
	payload := []byte{0x00, 0xff, 'p', 'm', '\n'}
	var calls int
	var received []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Method != http.MethodPost || r.URL.Path != "/assets" {
			t.Fatalf("request = %s %s, want POST /assets", r.Method, r.URL.Path)
		}
		if got := r.Header.Get("Content-Type"); got != "application/octet-stream" {
			t.Fatalf("Content-Type = %q, want application/octet-stream", got)
		}
		var err error
		received, err = io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"asset-proof"}`))
	}))
	t.Cleanup(server.Close)

	root := t.TempDir()
	if err := app.InitProject(root); err != nil {
		t.Fatalf("InitProject: %v", err)
	}
	const sourceName = "caller-root-upload-sentinel.bin"
	if err := os.WriteFile(filepath.Join(root, sourceName), payload, 0o600); err != nil {
		t.Fatalf("write caller source: %v", err)
	}
	a, err := app.Open(root)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	connector := binaryUploadLifecycleConnector(server.URL)
	a.Registry().Register(connector)
	if _, err := a.AddCredential(ctx, app.AddCredentialRequest{Name: "binary-local", Connector: connector.Name()}); err != nil {
		t.Fatalf("AddCredential: %v", err)
	}

	plan, _, err := a.PlanConnectorCommand(ctx, app.PlanConnectorCommandRequest{
		Name: "asset_proof", Connector: connector.Name(), Credential: "binary-local", Path: []string{"assets", "upload"},
		Flags: map[string][]string{"file-path": {sourceName}},
	})
	if err != nil {
		t.Fatalf("PlanConnectorCommand: %v", err)
	}
	if plan.ApprovalToken != "" {
		t.Fatalf("planned binary upload emitted approval token %q before preview", plan.ApprovalToken)
	}
	if plan.Status != "planned" || plan.PreviewDigest != "" || !plan.PreviewedAt.IsZero() {
		t.Fatalf("planned binary upload = %#v, want unpreviewed plan", plan)
	}
	if _, found := plan.ConnectorCommandRecord["file_path"]; found {
		t.Fatalf("stored command record = %#v, want file_path withheld", plan.ConnectorCommandRecord)
	}
	wantDigest := sha256.Sum256(payload)
	if len(plan.PayloadIdentity) != 1 || plan.PayloadIdentity[0].ContentSHA256 != hex.EncodeToString(wantDigest[:]) || plan.PayloadIdentity[0].SizeBytes != int64(len(payload)) {
		t.Fatalf("plan payload identity = %#v, want byte count and SHA-256 for the caller file", plan.PayloadIdentity)
	}
	// This is the same safe projection the CLI writes for a JSON plan. The
	// executable record is never public, and the sample is redacted again at
	// the boundary.
	output := plan
	output.ConnectorCommandRecord = nil
	output.Sample = app.RedactReversePlanRecords(output.Sample, output.RedactFields)
	emitted, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("marshal safe plan output: %v", err)
	}
	if strings.Contains(string(emitted), sourceName) || strings.Contains(string(emitted), root) {
		t.Fatalf("safe plan output exposes caller path: %s", emitted)
	}
	if raw := stateBytes(t, root); strings.Contains(raw, sourceName) || strings.Contains(raw, root) {
		t.Fatalf("state.json exposes caller path: %s", raw)
	}

	// Arm a synthetically valid token in the persisted plan. This proves the
	// executor itself refuses planned binary uploads before provider I/O; it is
	// not merely relying on the absence of a token in the planner's output.
	const prematureToken = "premature-binary-upload-token"
	setStoredApprovalTokenHash(t, root, plan.ID, prematureToken)
	_, err = a.RunReverseETL(ctx, app.RunReverseETLRequest{
		PlanID: plan.ID, ApprovalToken: prematureToken, WithheldFlags: map[string][]string{"file-path": {sourceName}},
	})
	if err == nil || !strings.Contains(err.Error(), "must be previewed") {
		t.Fatalf("RunReverseETL before preview error = %v, want persisted-preview refusal", err)
	}
	if calls != 0 {
		t.Fatalf("pre-preview run reached provider %d time(s)", calls)
	}

	previewed, preview, err := a.PreviewConnectorCommandPlan(ctx, plan.ID, map[string][]string{"file-path": {sourceName}})
	if err != nil {
		t.Fatalf("PreviewConnectorCommandPlan: %v", err)
	}
	if preview.Digest == "" || previewed.Status != "previewed" || previewed.ApprovalToken == "" {
		t.Fatalf("previewed binary upload = %#v, preview = %#v; want persisted digest and one token", previewed, preview)
	}
	if raw := stateBytes(t, root); strings.Contains(raw, sourceName) || strings.Contains(raw, root) {
		t.Fatalf("preview state.json exposes caller path: %s", raw)
	}

	// The persisted preview is bound to the file's exact digest. Changing it
	// invalidates the plan before a requester can reach the provider.
	if err := os.WriteFile(filepath.Join(root, sourceName), []byte("changed-after-preview"), 0o600); err != nil {
		t.Fatalf("change caller source: %v", err)
	}
	_, err = a.RunReverseETL(ctx, app.RunReverseETLRequest{
		PlanID: previewed.ID, ApprovalToken: previewed.ApprovalToken, WithheldFlags: map[string][]string{"file-path": {sourceName}},
	})
	if err == nil || !strings.Contains(err.Error(), "payload changed") {
		t.Fatalf("RunReverseETL with changed source error = %v, want digest-bound refusal", err)
	}
	if calls != 0 {
		t.Fatalf("changed-source run reached provider %d time(s)", calls)
	}
	if err := os.WriteFile(filepath.Join(root, sourceName), payload, 0o600); err != nil {
		t.Fatalf("restore caller source: %v", err)
	}

	// The invalidated plan intentionally cannot be reapproved. Replan and
	// preview the restored declared source, then prove its exact bytes and the
	// provider's created receipt reach the completed App result.
	plan, _, err = a.PlanConnectorCommand(ctx, app.PlanConnectorCommandRequest{
		Name: "asset_proof_retry", Connector: connector.Name(), Credential: "binary-local", Path: []string{"assets", "upload"},
		Flags: map[string][]string{"file-path": {sourceName}},
	})
	if err != nil {
		t.Fatalf("PlanConnectorCommand restored source: %v", err)
	}
	previewed, preview, err = a.PreviewConnectorCommandPlan(ctx, plan.ID, map[string][]string{"file-path": {sourceName}})
	if err != nil {
		t.Fatalf("PreviewConnectorCommandPlan restored source: %v", err)
	}
	if preview.Digest == "" || previewed.ApprovalToken == "" {
		t.Fatalf("restored preview = %#v/%#v, want digest and approval token", previewed, preview)
	}
	run, err := a.RunReverseETL(ctx, app.RunReverseETLRequest{
		PlanID: previewed.ID, ApprovalToken: previewed.ApprovalToken, WithheldFlags: map[string][]string{"file-path": {sourceName}},
	})
	if err != nil {
		t.Fatalf("RunReverseETL after preview: %v", err)
	}
	if run.RecordsSucceeded != 1 || calls != 1 || !bytes.Equal(received, payload) {
		t.Fatalf("run/calls/body = %#v/%d/%x, want one successful exact upload %x", run, calls, received, payload)
	}
	var result struct {
		RecordsWritten    int `json:"records_written"`
		ProviderResponses []struct {
			Status int `json:"status"`
			Body   any `json:"body"`
		} `json:"provider_responses"`
	}
	if err := json.Unmarshal(run.DestinationResult, &result); err != nil {
		t.Fatalf("decode retained provider result: %v", err)
	}
	if result.RecordsWritten != 1 || len(result.ProviderResponses) != 1 || result.ProviderResponses[0].Status != http.StatusCreated {
		t.Fatalf("retained provider result = %#v, want one exact 201 receipt", result)
	}
}

func binaryUploadLifecycleConnector(baseURL string) *engine.Connector {
	return engine.New(engine.Bundle{
		Name: "binary-lifecycle-fixture",
		HTTP: engine.HTTPBase{URL: baseURL, Auth: []engine.AuthSpec{{Mode: "none"}}},
		Writes: []engine.WriteAction{{
			Name: "upload_asset", Kind: "create", Method: http.MethodPost, Path: "/assets", BodyType: "binary_upload",
			RedactFields: []string{"file_path"},
			RecordSchema: json.RawMessage(`{"type":"object","additionalProperties":false,"required":["file_path"],"properties":{"file_path":{"type":"string","minLength":1}}}`),
			BinaryUpload: &engine.BinaryUploadSpec{SourceField: "file_path", MaxBytes: 1024, AllowedMediaTypes: []string{"application/octet-stream"}},
		}},
		CLISurface: &engine.CLISurface{Commands: []engine.CLICommand{{
			Path: "assets upload", Intent: "binary_upload", Availability: "implemented", Write: "upload_asset",
			Flags: []engine.CLIFlag{{Name: "file-path", Type: "string", MapsTo: "record.file_path", Required: true}},
		}}},
	}, nil)
}

func setStoredApprovalTokenHash(t *testing.T, root, planID, token string) {
	t.Helper()
	path := filepath.Join(root, ".polymetrics", "state", "state.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read state: %v", err)
	}
	var state map[string]any
	if err := json.Unmarshal(raw, &state); err != nil {
		t.Fatalf("decode state: %v", err)
	}
	plans, ok := state["reverse_plans"].([]any)
	if !ok {
		t.Fatalf("reverse_plans = %#v", state["reverse_plans"])
	}
	sum := sha256.Sum256([]byte(token))
	for _, value := range plans {
		plan, ok := value.(map[string]any)
		if !ok || plan["id"] != planID {
			continue
		}
		plan["approval_token_hash"] = hex.EncodeToString(sum[:])
		updated, err := json.Marshal(state)
		if err != nil {
			t.Fatalf("encode state: %v", err)
		}
		if err := os.WriteFile(path, updated, 0o600); err != nil {
			t.Fatalf("write state: %v", err)
		}
		return
	}
	t.Fatalf("plan %s not found in state", planID)
}
