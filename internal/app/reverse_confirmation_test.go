package app_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/connectors"
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
	plan, _, err := a.PreviewConnectorCommandPlan(ctx, plan.ID)
	if err != nil {
		t.Fatalf("PreviewConnectorCommandPlan() error = %v", err)
	}

	_, err = a.RunReverseETL(ctx, app.RunReverseETLRequest{
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

func TestRunReverseETLRejectsDestructiveConnectorCommandWithoutPreview(t *testing.T) {
	ctx := context.Background()
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	a, plan := setupGitHubDestructiveCommandPlan(t, ctx, server.URL)
	_, err := a.RunReverseETL(ctx, app.RunReverseETLRequest{
		PlanID:        plan.ID,
		ApprovalToken: plan.ApprovalToken,
		Confirmation:  connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive},
	})
	if err == nil {
		t.Fatal("RunReverseETL() executed a destructive command that was never previewed")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "preview") {
		t.Fatalf("RunReverseETL() error = %v, want preview rejection", err)
	}
	if calls != 0 {
		t.Fatalf("destructive write dispatched without preview; calls=%d", calls)
	}
}

func TestDestructiveConnectorCommandMintsApprovalOnlyAfterPreview(t *testing.T) {
	ctx := context.Background()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	a, plan := setupGitHubDestructiveCommandPlan(t, ctx, server.URL)
	if plan.ApprovalToken != "" {
		t.Fatal("destructive plan minted approval before preview")
	}
	previewed, _, err := a.PreviewConnectorCommandPlan(ctx, plan.ID)
	if err != nil {
		t.Fatalf("PreviewConnectorCommandPlan() error = %v", err)
	}
	if previewed.Status != "previewed" {
		t.Fatalf("previewed status = %q, want previewed", previewed.Status)
	}
	if previewed.ApprovalToken == "" {
		t.Fatal("preview did not mint a bounded approval token")
	}
}

func TestDestructiveCanonicalCommandPreviewProducesApprovablePlan(t *testing.T) {
	ctx := context.Background()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	a, _ := setupGitHubApp(t, ctx, server.URL)

	plan, preview, err := a.PlanConnectorCommand(ctx, app.PlanConnectorCommandRequest{
		Name:       "delete_deploy_key",
		Connector:  "github",
		Credential: "github-local",
		Path:       []string{"repo", "deploy-key", "delete"},
		Flags:      map[string][]string{"key-id": {"42"}},
		Preview:    true,
	})
	if err != nil {
		t.Fatalf("PlanConnectorCommand(--preview) error = %v", err)
	}
	if preview == nil {
		t.Fatal("canonical destructive command returned no preview")
	}
	if plan.Status != "previewed" || plan.ApprovalToken == "" || plan.PreviewDigest == "" {
		t.Fatalf("previewed canonical plan = %+v, want persisted digest and approval token", plan)
	}
}

func TestGenericDestructivePlanMintsApprovalOnlyAfterPreview(t *testing.T) {
	ctx := context.Background()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	a, plan := setupGitHubGenericDestructivePlan(t, ctx, server.URL)
	if plan.ApprovalToken != "" {
		t.Fatal("generic destructive plan minted approval before preview")
	}
	previewed, preview, err := a.PreviewReversePlan(ctx, plan.ID)
	if err != nil {
		t.Fatalf("PreviewReversePlan() error = %v", err)
	}
	if previewed.Status != "previewed" || previewed.PreviewDigest == "" || previewed.PreviewedAt.IsZero() {
		t.Fatalf("previewed plan = %+v, want persisted preview identity", previewed)
	}
	if previewed.ApprovalToken == "" {
		t.Fatal("generic preview did not mint a bounded approval token")
	}
	if preview.RecordsStaged != plan.RecordCount || preview.Action != plan.Action {
		t.Fatalf("preview = %+v, want action %q and %d staged records", preview, plan.Action, plan.RecordCount)
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
	plan, _, err := a.PreviewConnectorCommandPlan(ctx, plan.ID)
	if err != nil {
		t.Fatalf("PreviewConnectorCommandPlan() error = %v", err)
	}

	run, err := a.RunReverseETL(ctx, app.RunReverseETLRequest{
		PlanID:        plan.ID,
		ApprovalToken: plan.ApprovalToken,
		Confirmation:  connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive},
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
	plan, _, err := a.PreviewReversePlan(ctx, plan.ID)
	if err != nil {
		t.Fatalf("PreviewReversePlan() error = %v", err)
	}

	_, err = a.RunReverseETL(ctx, app.RunReverseETLRequest{
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

func TestRunReverseETLRejectsPreviewDigestDriftBeforeNativeWrite(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	if err := app.InitProject(root); err != nil {
		t.Fatalf("InitProject() error = %v", err)
	}
	a, err := app.Open(root)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	destination := &driftingDestructiveConnector{digest: strings.Repeat("a", 64)}
	a.Registry().Register(destination)
	if _, err := a.AddCredential(ctx, app.AddCredentialRequest{Name: "drifting-local", Connector: destination.Name()}); err != nil {
		t.Fatalf("AddCredential() error = %v", err)
	}
	warehouseDir := filepath.Join(root, ".polymetrics", "warehouse")
	if err := os.MkdirAll(warehouseDir, 0o700); err != nil {
		t.Fatalf("mkdir warehouse: %v", err)
	}
	if err := os.WriteFile(filepath.Join(warehouseDir, "deletes.jsonl"), []byte("{\"id\":\"42\"}\n"), 0o600); err != nil {
		t.Fatalf("write warehouse row: %v", err)
	}
	plan, err := a.PlanReverseETL(ctx, app.PlanReverseETLRequest{
		Name:                  "delete_widget",
		SourceTable:           "deletes",
		DestinationConnector:  destination.Name(),
		DestinationCredential: "drifting-local",
		Action:                "delete_widget",
		Mappings:              map[string]string{"id": "id"},
	})
	if err != nil {
		t.Fatalf("PlanReverseETL() error = %v", err)
	}
	plan, _, err = a.PreviewReversePlan(ctx, plan.ID)
	if err != nil {
		t.Fatalf("PreviewReversePlan() error = %v", err)
	}
	destination.digest = strings.Repeat("b", 64)

	_, err = a.RunReverseETL(ctx, app.RunReverseETLRequest{
		PlanID:        plan.ID,
		ApprovalToken: plan.ApprovalToken,
		Confirmation:  connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive},
	})
	if err == nil {
		t.Fatal("RunReverseETL() executed after the preview digest drifted")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "preview") {
		t.Fatalf("RunReverseETL() error = %v, want preview mismatch", err)
	}
	if destination.writes != 0 {
		t.Fatalf("native write dispatched after preview drift; writes=%d", destination.writes)
	}
}

func TestRunReverseETLRejectsApprovalHashStateTamper(t *testing.T) {
	ctx := context.Background()
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	a, plan := setupGitHubDestructiveCommandPlan(t, ctx, server.URL)
	plan, _, err := a.PreviewConnectorCommandPlan(ctx, plan.ID)
	if err != nil {
		t.Fatalf("PreviewConnectorCommandPlan() error = %v", err)
	}
	attackerToken := "attacker-chosen-token"
	sum := sha256.Sum256([]byte(attackerToken))
	mutateStoredReversePlan(t, a.ProjectDir(), plan.ID, func(stored map[string]any) {
		stored["approval_token_hash"] = hex.EncodeToString(sum[:])
	})
	tampered, err := app.Open(filepath.Dir(a.ProjectDir()))
	if err != nil {
		t.Fatalf("Open(tampered state) error = %v", err)
	}

	_, err = tampered.RunReverseETL(ctx, app.RunReverseETLRequest{
		PlanID:        plan.ID,
		ApprovalToken: attackerToken,
		Confirmation:  connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive},
	})
	if err == nil {
		t.Fatal("RunReverseETL() accepted an attacker-replaced approval token hash")
	}
	if calls.Load() != 0 {
		t.Fatalf("destructive write dispatched after state tamper; calls=%d", calls.Load())
	}
}

func TestRunReverseETLConsumesApprovalAtomicallyAcrossProcesses(t *testing.T) {
	ctx := context.Background()
	var calls atomic.Int32
	firstDispatched := make(chan struct{})
	secondDispatched := make(chan struct{})
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		switch calls.Add(1) {
		case 1:
			close(firstDispatched)
		case 2:
			close(secondDispatched)
		}
		<-release
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	first, plan := setupGitHubDestructiveCommandPlan(t, ctx, server.URL)
	plan, _, err := first.PreviewConnectorCommandPlan(ctx, plan.ID)
	if err != nil {
		t.Fatalf("PreviewConnectorCommandPlan() error = %v", err)
	}
	second, err := app.Open(filepath.Dir(first.ProjectDir()))
	if err != nil {
		t.Fatalf("Open(second process) error = %v", err)
	}

	errs := make(chan error, 2)
	var wg sync.WaitGroup
	run := func(instance *app.App) {
		wg.Add(1)
		go func(instance *app.App) {
			defer wg.Done()
			_, runErr := instance.RunReverseETL(ctx, app.RunReverseETLRequest{
				PlanID:        plan.ID,
				ApprovalToken: plan.ApprovalToken,
				Confirmation:  connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive},
			})
			errs <- runErr
		}(instance)
	}
	run(first)
	<-firstDispatched
	run(second)
	select {
	case <-secondDispatched:
	case <-time.After(500 * time.Millisecond):
	}
	close(release)
	wg.Wait()
	close(errs)

	var successes int
	for runErr := range errs {
		if runErr == nil {
			successes++
		}
	}
	if successes != 1 {
		t.Fatalf("successful executions = %d, want exactly one", successes)
	}
	if calls.Load() != 1 {
		t.Fatalf("destructive request calls = %d, want exactly one", calls.Load())
	}
}

func TestRunReverseETLConsumesBulkApprovalAtomicallyAcrossProcesses(t *testing.T) {
	ctx := context.Background()
	var calls atomic.Int32
	firstDispatched := make(chan struct{})
	secondDispatched := make(chan struct{})
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		switch calls.Add(1) {
		case 1:
			close(firstDispatched)
		case 2:
			close(secondDispatched)
		}
		<-release
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	first, plan := setupGitHubGenericDestructivePlan(t, ctx, server.URL)
	plan, _, err := first.PreviewReversePlan(ctx, plan.ID)
	if err != nil {
		t.Fatalf("PreviewReversePlan() error = %v", err)
	}
	second, err := app.Open(filepath.Dir(first.ProjectDir()))
	if err != nil {
		t.Fatalf("Open(second process) error = %v", err)
	}

	errs := make(chan error, 2)
	var wg sync.WaitGroup
	run := func(instance *app.App) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, runErr := instance.RunReverseETL(ctx, app.RunReverseETLRequest{
				PlanID: plan.ID, ApprovalToken: plan.ApprovalToken,
				Confirmation: connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive},
			})
			errs <- runErr
		}()
	}
	run(first)
	<-firstDispatched
	run(second)
	select {
	case <-secondDispatched:
	case <-time.After(500 * time.Millisecond):
	}
	close(release)
	wg.Wait()
	close(errs)

	var successes int
	for runErr := range errs {
		if runErr == nil {
			successes++
		}
	}
	if successes != 1 || calls.Load() != 1 {
		t.Fatalf("bulk executions = %d successes, %d requests; want exactly one", successes, calls.Load())
	}
}

func TestPreviewReversePlanRejectsExpiredGenericPlan(t *testing.T) {
	ctx := context.Background()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	a, plan := setupGitHubGenericDestructivePlan(t, ctx, server.URL)
	mutateStoredReversePlan(t, a.ProjectDir(), plan.ID, func(stored map[string]any) {
		stored["expires_at"] = time.Now().UTC().Add(-time.Minute).Format(time.RFC3339Nano)
	})
	expired, err := app.Open(filepath.Dir(a.ProjectDir()))
	if err != nil {
		t.Fatalf("Open(expired state) error = %v", err)
	}

	if _, _, err := expired.PreviewReversePlan(ctx, plan.ID); err == nil {
		t.Fatal("PreviewReversePlan() minted approval for an expired generic plan")
	}
}

func TestExecutedDestructivePlanCannotBeRepreviewedForReplay(t *testing.T) {
	ctx := context.Background()
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	a, plan := setupGitHubDestructiveCommandPlan(t, ctx, server.URL)
	plan, _, err := a.PreviewConnectorCommandPlan(ctx, plan.ID)
	if err != nil {
		t.Fatalf("PreviewConnectorCommandPlan() error = %v", err)
	}
	if _, err := a.RunReverseETL(ctx, app.RunReverseETLRequest{
		PlanID:        plan.ID,
		ApprovalToken: plan.ApprovalToken,
		Confirmation:  connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive},
	}); err != nil {
		t.Fatalf("RunReverseETL() error = %v", err)
	}
	if _, _, err := a.PreviewConnectorCommandPlan(ctx, plan.ID); err == nil {
		t.Fatal("executed destructive plan was re-previewed into an approvable state")
	}
	if calls != 1 {
		t.Fatalf("destructive request calls=%d, want exactly one", calls)
	}
}

type driftingDestructiveConnector struct {
	digest string
	writes int
}

func (c *driftingDestructiveConnector) Name() string { return "drifting-destructive" }

func (c *driftingDestructiveConnector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:         c.Name(),
		DisplayName:  "Drifting destructive fixture",
		Capabilities: connectors.Capabilities{Write: true},
	}
}

func (c *driftingDestructiveConnector) Manifest() connectors.Manifest {
	return connectors.Manifest{
		Metadata: c.Metadata(),
		WriteActions: []connectors.WriteActionSpec{{
			Name: "delete_widget", Method: http.MethodDelete, Path: "/widgets/{id}",
		}},
	}
}

func (c *driftingDestructiveConnector) Check(ctx context.Context, _ connectors.RuntimeConfig) error {
	return ctx.Err()
}

func (c *driftingDestructiveConnector) Catalog(context.Context, connectors.RuntimeConfig) (connectors.Catalog, error) {
	return connectors.Catalog{}, connectors.ErrUnsupportedOperation
}

func (c *driftingDestructiveConnector) Read(context.Context, connectors.ReadRequest, func(connectors.Record) error) error {
	return connectors.ErrUnsupportedOperation
}

func (c *driftingDestructiveConnector) DryRunWrite(_ context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WritePreview, error) {
	return connectors.WritePreview{
		RecordsStaged: len(records), Action: req.Action, Digest: c.digest,
		ApprovalTarget: connectors.WriteApprovalTarget{
			Connector: c.Name(), Operation: req.Action, Method: http.MethodDelete, MutationClass: "delete",
			TargetDigest: strings.Repeat("c", 64), CredentialRevision: req.Config.CredentialRevision,
		},
	}, nil
}

func (c *driftingDestructiveConnector) Write(context.Context, connectors.WriteRequest, []connectors.Record) (connectors.WriteResult, error) {
	c.writes++
	return connectors.WriteResult{RecordsWritten: 1}, nil
}

func setupGitHubDestructiveCommandPlan(t *testing.T, ctx context.Context, baseURL string) (*app.App, app.ReversePlan) {
	t.Helper()
	a, _ := setupGitHubApp(t, ctx, baseURL)
	plan, _, err := a.PlanConnectorCommand(ctx, app.PlanConnectorCommandRequest{
		Name:       "delete_deploy_key",
		Connector:  "github",
		Credential: "github-local",
		Path:       []string{"repo", "deploy-key", "delete"},
		Flags:      map[string][]string{"key-id": {"42"}},
	})
	if err != nil {
		t.Fatalf("PlanConnectorCommand(repo deploy-key delete) error = %v", err)
	}
	return a, plan
}

func setupGitHubGenericDestructivePlan(t *testing.T, ctx context.Context, baseURL string) (*app.App, app.ReversePlan) {
	t.Helper()
	a, root := setupGitHubApp(t, ctx, baseURL)
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

func setupGitHubApp(t *testing.T, ctx context.Context, baseURL string) (*app.App, string) {
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
	return a, root
}

func mutateStoredReversePlan(t *testing.T, projectDir, planID string, mutate func(map[string]any)) {
	t.Helper()
	path := filepath.Join(projectDir, "state", "state.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(state) error = %v", err)
	}
	var stored map[string]any
	if err := json.Unmarshal(raw, &stored); err != nil {
		t.Fatalf("Unmarshal(state) error = %v", err)
	}
	plans, ok := stored["reverse_plans"].([]any)
	if !ok {
		t.Fatalf("state reverse_plans has type %T", stored["reverse_plans"])
	}
	for _, item := range plans {
		plan, ok := item.(map[string]any)
		if !ok || plan["id"] != planID {
			continue
		}
		mutate(plan)
		encoded, err := json.MarshalIndent(stored, "", "  ")
		if err != nil {
			t.Fatalf("Marshal(state) error = %v", err)
		}
		encoded = append(encoded, '\n')
		if err := os.WriteFile(path, encoded, 0o600); err != nil {
			t.Fatalf("WriteFile(state) error = %v", err)
		}
		return
	}
	t.Fatalf("reverse plan %q not found in state", planID)
}
