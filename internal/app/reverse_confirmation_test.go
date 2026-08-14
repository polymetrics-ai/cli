package app_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
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
	plan, _, err := a.PreviewConnectorCommandPlan(ctx, plan.ID, nil)
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
	previewed, _, err := a.PreviewConnectorCommandPlan(ctx, plan.ID, nil)
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

// A GitHub deploy-key DELETE can return 404 for a request that reached the
// exact scoped target. The lab found that the current declaration treats that
// status as a completed idempotent delete, even when an independent list still
// exposes the same generated key. Keep the reproduction local: its purpose is
// to distinguish route/record mapping from the false-success masking rule
// before another provider mutation is allowed.
func TestGitHubDeployKeyDeleteDoesNotMaskNotFoundForAVisibleBoundFixture(t *testing.T) {
	ctx := context.Background()
	const (
		owner = "lab-owner"
		repo  = "lab-repository"
		keyID = "4242"
		title = "pm-live-lab-generated-deploy-key"
	)
	var deletePath, listPath string
	deleteCalls := 0
	listCalls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodDelete:
			deleteCalls++
			deletePath = r.URL.Path
			w.WriteHeader(http.StatusNotFound)
		case http.MethodGet:
			listCalls++
			listPath = r.URL.Path
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprintf(w, `[{"id":%s,"title":%q,"read_only":true}]`, keyID, title)
		default:
			t.Errorf("request method = %s, want DELETE or GET", r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	t.Cleanup(server.Close)

	a, _ := setupGitHubApp(t, ctx, server.URL)
	targetConfig := map[string]string{"owner": owner, "repo": repo}
	plan, preview, err := a.PlanConnectorCommand(ctx, app.PlanConnectorCommandRequest{
		Name:       "deploy_key_not_found_must_fail",
		Connector:  "github",
		Credential: "github-local",
		Config:     targetConfig,
		Path:       []string{"repo", "deploy-key", "delete"},
		Flags:      map[string][]string{"key-id": {keyID}},
		Preview:    true,
	})
	if err != nil {
		t.Fatalf("PlanConnectorCommand(repo deploy-key delete): %v", err)
	}
	if preview == nil || plan.ApprovalToken == "" {
		t.Fatalf("previewed destructive plan = %#v/%#v, want preview and approval", plan, preview)
	}

	run, runErr := a.RunReverseETL(ctx, app.RunReverseETLRequest{
		PlanID:        plan.ID,
		ApprovalToken: plan.ApprovalToken,
		Confirmation:  connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive},
	})
	if deleteCalls != 1 {
		t.Fatalf("deploy-key delete calls = %d, want 1", deleteCalls)
	}
	if want := "/repos/" + owner + "/" + repo + "/keys/" + keyID; deletePath != want {
		t.Fatalf("deploy-key delete path = %q, want %q", deletePath, want)
	}

	connector, runtime, err := a.ResolveConnectorCredential(ctx, "github", "github-local", targetConfig)
	if err != nil {
		t.Fatalf("ResolveConnectorCredential(read-back): %v", err)
	}
	var rows []connectors.Record
	if err := connector.Read(ctx, connectors.ReadRequest{Stream: "deploy_keys", Config: runtime}, func(record connectors.Record) error {
		rows = append(rows, record)
		return nil
	}); err != nil {
		t.Fatalf("independent deploy-key list read-back: %v", err)
	}
	if listCalls != 1 {
		t.Fatalf("deploy-key list calls = %d, want 1", listCalls)
	}
	if want := "/repos/" + owner + "/" + repo + "/keys"; listPath != want {
		t.Fatalf("deploy-key list path = %q, want %q", listPath, want)
	}
	if len(rows) != 1 || fmt.Sprint(rows[0]["id"]) != keyID || rows[0]["title"] != title || rows[0]["read_only"] != true {
		t.Fatalf("independent deploy-key list did not retain the exact local fixture: %#v", rows)
	}
	if runErr == nil {
		t.Fatalf("RunReverseETL() unexpectedly completed after provider 404: status=%q succeeded=%d failed=%d", run.Status, run.RecordsSucceeded, run.RecordsFailed)
	}
	if run.Status != "failed" || run.RecordsSucceeded != 0 || run.RecordsFailed != 1 {
		t.Fatalf("deploy-key 404 run = %#v, want one failed write", run)
	}
}

// The label deletion action is the nearby, proven 404 control: it follows the
// identical plan/preview/confirmation/write route, but it does not declare a
// missing status as success. A not-found response must remain visible to the
// caller. Together with the deploy-key reproduction above, this separates the
// failure from credential/config binding or the generic destructive-write path.
func TestGitHubLabelDeleteKeepsNotFoundVisibleForTheSameScopedWritePath(t *testing.T) {
	ctx := context.Background()
	const (
		owner = "lab-owner"
		repo  = "lab-repository"
		name  = "pm-live-lab-control-label"
	)
	deleteCalls := 0
	deletePath := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("request method = %s, want DELETE", r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		deleteCalls++
		deletePath = r.URL.Path
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(server.Close)

	a, _ := setupGitHubApp(t, ctx, server.URL)
	plan, preview, err := a.PlanConnectorCommand(ctx, app.PlanConnectorCommandRequest{
		Name:       "label_not_found_must_fail",
		Connector:  "github",
		Credential: "github-local",
		Config:     map[string]string{"owner": owner, "repo": repo},
		Path:       []string{"label", "delete"},
		Flags:      map[string][]string{"name": {name}},
		Preview:    true,
	})
	if err != nil {
		t.Fatalf("PlanConnectorCommand(label delete): %v", err)
	}
	if preview == nil || plan.ApprovalToken == "" {
		t.Fatal("label delete did not produce an approvable destructive plan")
	}

	run, runErr := a.RunReverseETL(ctx, app.RunReverseETLRequest{
		PlanID:        plan.ID,
		ApprovalToken: plan.ApprovalToken,
		Confirmation:  connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive},
	})
	if deleteCalls != 1 {
		t.Fatalf("label delete calls = %d, want 1", deleteCalls)
	}
	if want := "/repos/" + owner + "/" + repo + "/labels/" + name; deletePath != want {
		t.Fatalf("label delete path = %q, want %q", deletePath, want)
	}
	if runErr == nil {
		t.Fatalf("RunReverseETL() unexpectedly completed after label provider 404: status=%q succeeded=%d failed=%d", run.Status, run.RecordsSucceeded, run.RecordsFailed)
	}
	if run.Status != "failed" || run.RecordsSucceeded != 0 || run.RecordsFailed != 1 {
		t.Fatalf("label 404 run = status=%q succeeded=%d failed=%d, want one failed write", run.Status, run.RecordsSucceeded, run.RecordsFailed)
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
	previewed, preview, err := a.PreviewReversePlan(ctx, plan.ID, nil)
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

// A reverse plan may intentionally stage only a prefix of a larger source.
// Preview and execution must re-read and hash that exact approved slice; an
// extra row is not source drift for a plan that never included it.
func TestLimitedReversePlanPreviewsAndRunsItsExactApprovedSlice(t *testing.T) {
	ctx := context.Background()
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	a, root := setupGitHubApp(t, ctx, server.URL)
	seedWarehouseTableRows(t, root, "repo_deletes",
		`{"id":"row-1"}`,
		`{"id":"row-2"}`,
	)
	plan, err := a.PlanReverseETL(ctx, app.PlanReverseETLRequest{
		Name:                  "delete_one_repo",
		SourceTable:           "repo_deletes",
		DestinationConnector:  "github",
		DestinationCredential: "github-local",
		Action:                "repo",
		Mappings:              map[string]string{"id": "id"},
		Limit:                 1,
	})
	if err != nil {
		t.Fatalf("PlanReverseETL() error = %v", err)
	}
	if plan.RecordCount != 1 {
		t.Fatalf("planned records = %d, want 1", plan.RecordCount)
	}

	plan, preview, err := a.PreviewReversePlan(ctx, plan.ID, nil)
	if err != nil {
		t.Fatalf("PreviewReversePlan() error = %v", err)
	}
	if preview.RecordsStaged != 1 {
		t.Fatalf("preview records staged = %d, want 1", preview.RecordsStaged)
	}
	run, err := a.RunReverseETL(ctx, app.RunReverseETLRequest{
		PlanID:        plan.ID,
		ApprovalToken: plan.ApprovalToken,
		Confirmation:  connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive},
	})
	if err != nil {
		t.Fatalf("RunReverseETL() error = %v", err)
	}
	if run.RecordsStaged != 1 || run.RecordsSucceeded != 1 {
		t.Fatalf("run staged/succeeded = %d/%d, want 1/1", run.RecordsStaged, run.RecordsSucceeded)
	}
	if calls.Load() != 1 {
		t.Fatalf("write calls = %d, want 1", calls.Load())
	}
}

// The live GitHub validation must use the declaration's precise issue endpoint.
// This request-target assertion prevents a plan/preview/run success from hiding
// a malformed URL that only fails after approval is consumed.
func TestGitHubCreateIssueReversePlanUsesDeclaredEndpoint(t *testing.T) {
	ctx := context.Background()
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/repos/acme/widgets/issues" {
			t.Errorf("path = %q, want /repos/acme/widgets/issues", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"number":1,"title":"fixture issue"}`))
	}))
	defer server.Close()

	a, root := setupGitHubApp(t, ctx, server.URL)
	seedWarehouseTableRows(t, root, "issue_candidates",
		`{"title":"fixture issue","body":"fixture body"}`,
	)
	plan, err := a.PlanReverseETL(ctx, app.PlanReverseETLRequest{
		Name:                  "create_fixture_issue",
		SourceTable:           "issue_candidates",
		DestinationConnector:  "github",
		DestinationCredential: "github-local",
		Action:                "create_issue",
		Mappings:              map[string]string{"title": "title", "body": "body"},
	})
	if err != nil {
		t.Fatalf("PlanReverseETL() error = %v", err)
	}
	if _, _, err := a.PreviewReversePlan(ctx, plan.ID, nil); err != nil {
		t.Fatalf("PreviewReversePlan() error = %v", err)
	}
	run, err := a.RunReverseETL(ctx, app.RunReverseETLRequest{PlanID: plan.ID, ApprovalToken: plan.ApprovalToken})
	if err != nil {
		t.Fatalf("RunReverseETL() error = %v", err)
	}
	if run.Status != "completed" || run.RecordsSucceeded != 1 {
		t.Fatalf("run = %+v, want one completed issue creation", run)
	}
	if calls.Load() != 1 {
		t.Fatalf("create issue calls = %d, want 1", calls.Load())
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
	plan, _, err := a.PreviewConnectorCommandPlan(ctx, plan.ID, nil)
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
	plan, _, err := a.PreviewReversePlan(ctx, plan.ID, nil)
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
	seedWarehouseTableRows(t, root, "deletes", `{"id":"42"}`)
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
	plan, _, err = a.PreviewReversePlan(ctx, plan.ID, nil)
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
	plan, _, err := a.PreviewConnectorCommandPlan(ctx, plan.ID, nil)
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
	plan, _, err := first.PreviewConnectorCommandPlan(ctx, plan.ID, nil)
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
	plan, _, err := first.PreviewReversePlan(ctx, plan.ID, nil)
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

func TestReverseExecutionOpenDefersLegacyMigrationUntilApprovalConsumption(t *testing.T) {
	ctx := context.Background()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	active, plan := setupGitHubDestructiveCommandPlan(t, ctx, server.URL)
	plan, _, err := active.PreviewConnectorCommandPlan(ctx, plan.ID, nil)
	if err != nil {
		t.Fatalf("PreviewConnectorCommandPlan() error = %v", err)
	}
	root := filepath.Dir(active.ProjectDir())
	statePath := filepath.Join(active.ProjectDir(), "state", "state.json")
	contents, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatalf("ReadFile(state) error = %v", err)
	}
	var stored map[string]any
	if err := json.Unmarshal(contents, &stored); err != nil {
		t.Fatalf("Unmarshal(state) error = %v", err)
	}
	delete(stored, "workspace_id")
	legacy, err := json.Marshal(stored)
	if err != nil {
		t.Fatalf("Marshal(legacy state) error = %v", err)
	}
	if err := os.WriteFile(statePath, legacy, 0o600); err != nil {
		t.Fatalf("WriteFile(legacy state) error = %v", err)
	}

	deferred, err := app.OpenForReverseExecution(root)
	if err != nil {
		t.Fatalf("OpenForReverseExecution() error = %v", err)
	}
	afterOpen, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatalf("ReadFile(after deferred open) error = %v", err)
	}
	if !bytes.Equal(afterOpen, legacy) {
		t.Fatal("reverse execution open rewrote legacy state before approval consumption")
	}

	winner, err := app.Open(root)
	if err != nil {
		t.Fatalf("Open(winner) error = %v", err)
	}
	if _, err := winner.RunReverseETL(ctx, app.RunReverseETLRequest{
		PlanID:        plan.ID,
		ApprovalToken: plan.ApprovalToken,
		Confirmation:  connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive},
	}); err != nil {
		t.Fatalf("RunReverseETL(winner) error = %v", err)
	}
	afterWinner, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatalf("ReadFile(after winner) error = %v", err)
	}

	if _, err := deferred.RunReverseETL(ctx, app.RunReverseETLRequest{
		PlanID:        plan.ID,
		ApprovalToken: plan.ApprovalToken,
		Confirmation:  connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive},
	}); err == nil || !strings.Contains(err.Error(), "already been consumed") {
		t.Fatalf("RunReverseETL(replay) error = %v, want consumed approval rejection", err)
	}
	afterReplay, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatalf("ReadFile(after replay) error = %v", err)
	}
	if !bytes.Equal(afterReplay, afterWinner) {
		t.Fatal("replayed reverse execution rewrote project state")
	}
	if _, err := os.Stat(statePath + ".lock"); err == nil || !os.IsNotExist(err) {
		t.Fatalf("replayed reverse execution created a state lock file: %v", err)
	}
}

func TestConsumedApprovalCannotBeResurrectedByStaleStateSave(t *testing.T) {
	ctx := context.Background()
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	active, plan := setupGitHubDestructiveCommandPlan(t, ctx, server.URL)
	plan, _, err := active.PreviewConnectorCommandPlan(ctx, plan.ID, nil)
	if err != nil {
		t.Fatalf("PreviewConnectorCommandPlan() error = %v", err)
	}
	stale, err := app.Open(filepath.Dir(active.ProjectDir()))
	if err != nil {
		t.Fatalf("Open(stale process) error = %v", err)
	}
	if _, err := active.RunReverseETL(ctx, app.RunReverseETLRequest{
		PlanID:        plan.ID,
		ApprovalToken: plan.ApprovalToken,
		Confirmation:  connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive},
	}); err != nil {
		t.Fatalf("RunReverseETL() error = %v", err)
	}

	_, staleSaveErr := stale.AddCredential(ctx, app.AddCredentialRequest{
		Name:      "stale-github",
		Connector: "github",
		Config:    map[string]string{"owner": "acme", "repo": "stale", "public_access": "true", "base_url": server.URL},
	})
	reopened, err := app.Open(filepath.Dir(active.ProjectDir()))
	if err != nil {
		t.Fatalf("Open(reloaded process) error = %v", err)
	}
	_, replayErr := reopened.RunReverseETL(ctx, app.RunReverseETLRequest{
		PlanID:        plan.ID,
		ApprovalToken: plan.ApprovalToken,
		Confirmation:  connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive},
	})
	if staleSaveErr == nil {
		t.Fatal("stale whole-state save succeeded after approval consumption")
	}
	if replayErr == nil || calls.Load() != 1 {
		t.Fatalf("replayed consumed approval: error=%v calls=%d, want rejection after one request", replayErr, calls.Load())
	}
}

func TestConsumedApprovalCannotReplayFromRolledBackStateSnapshot(t *testing.T) {
	ctx := context.Background()
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	a, plan := setupGitHubDestructiveCommandPlan(t, ctx, server.URL)
	plan, _, err := a.PreviewConnectorCommandPlan(ctx, plan.ID, nil)
	if err != nil {
		t.Fatalf("PreviewConnectorCommandPlan() error = %v", err)
	}
	statePath := filepath.Join(a.ProjectDir(), "state", "state.json")
	snapshot, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatalf("ReadFile(previewed state) error = %v", err)
	}
	request := app.RunReverseETLRequest{
		PlanID: plan.ID, ApprovalToken: plan.ApprovalToken,
		Confirmation: connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive},
	}
	if _, err := a.RunReverseETL(ctx, request); err != nil {
		t.Fatalf("RunReverseETL() error = %v", err)
	}
	if err := os.WriteFile(statePath, snapshot, 0o600); err != nil {
		t.Fatalf("WriteFile(rolled-back state) error = %v", err)
	}
	rolledBack, err := app.Open(filepath.Dir(a.ProjectDir()))
	if err != nil {
		t.Fatalf("Open(rolled-back state) error = %v", err)
	}
	if _, _, err := rolledBack.PreviewConnectorCommandPlan(ctx, plan.ID, nil); err == nil {
		t.Fatal("PreviewConnectorCommandPlan() re-approved a consumed plan from rolled-back state")
	}
	if _, err := rolledBack.RunReverseETL(ctx, request); err == nil {
		t.Fatal("RunReverseETL() replayed a consumed grant from rolled-back state")
	}
	if calls.Load() != 1 {
		t.Fatalf("destructive request calls = %d, want exactly one", calls.Load())
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
		seal, ok := stored["plan_seal"].(map[string]any)
		if !ok {
			t.Fatalf("stored plan seal has type %T", stored["plan_seal"])
		}
		seal["expires_at"] = time.Now().UTC().Add(-time.Minute).Format(time.RFC3339Nano)
	})
	expired, err := app.Open(filepath.Dir(a.ProjectDir()))
	if err != nil {
		t.Fatalf("Open(expired state) error = %v", err)
	}

	if _, _, err := expired.PreviewReversePlan(ctx, plan.ID, nil); err == nil {
		t.Fatal("PreviewReversePlan() minted approval for an expired generic plan")
	}
}

func TestPreviewGrantExpiryIgnoresExtendedMutablePlanDeadline(t *testing.T) {
	ctx := context.Background()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	a, plan := setupGitHubGenericDestructivePlan(t, ctx, server.URL)
	mutateStoredReversePlan(t, a.ProjectDir(), plan.ID, func(stored map[string]any) {
		stored["expires_at"] = time.Now().UTC().Add(100 * 365 * 24 * time.Hour).Format(time.RFC3339Nano)
	})
	extended, err := app.Open(filepath.Dir(a.ProjectDir()))
	if err != nil {
		t.Fatalf("Open(extended state) error = %v", err)
	}
	previewed, _, err := extended.PreviewReversePlan(ctx, plan.ID, nil)
	if err != nil {
		t.Fatalf("PreviewReversePlan() error = %v", err)
	}
	if previewed.ApprovalGrant == nil {
		t.Fatal("PreviewReversePlan() returned no authenticated grant")
	}
	if previewed.ApprovalGrant.ExpiresAt.After(time.Now().UTC().Add(time.Hour)) {
		t.Fatalf("grant expiry = %s, want trusted short-lived deadline", previewed.ApprovalGrant.ExpiresAt)
	}
}

func TestRunReverseETLRejectsExpiredUnsignedPlan(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	if err := app.InitProject(root); err != nil {
		t.Fatalf("InitProject() error = %v", err)
	}
	a, err := app.Open(root)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	outboxDir := filepath.Join(root, ".polymetrics", "outbox")
	if _, err := a.AddCredential(ctx, app.AddCredentialRequest{
		Name: "outbox-local", Connector: "outbox", Config: map[string]string{"path": outboxDir},
	}); err != nil {
		t.Fatalf("AddCredential(outbox) error = %v", err)
	}
	seedWarehouseTableRows(t, root, "safe_rows", `{"id":"row-1"}`)
	plan, err := a.PlanReverseETL(ctx, app.PlanReverseETLRequest{
		Name: "safe_upsert", SourceTable: "safe_rows", DestinationConnector: "outbox",
		DestinationCredential: "outbox-local", Action: "upsert", Mappings: map[string]string{"id": "id"},
	})
	if err != nil {
		t.Fatalf("PlanReverseETL() error = %v", err)
	}
	if plan.PlanSeal != nil {
		t.Fatal("safe reverse plan unexpectedly has a destructive plan seal")
	}
	mutateStoredReversePlan(t, a.ProjectDir(), plan.ID, func(stored map[string]any) {
		stored["expires_at"] = time.Now().UTC().Add(-time.Minute).Format(time.RFC3339Nano)
		stored["plan_seal"] = map[string]any{"version": 1}
	})
	expired, err := app.Open(root)
	if err != nil {
		t.Fatalf("Open(expired state) error = %v", err)
	}
	if _, err := expired.RunReverseETL(ctx, app.RunReverseETLRequest{PlanID: plan.ID, ApprovalToken: plan.ApprovalToken}); err == nil || !strings.Contains(strings.ToLower(err.Error()), "expired") {
		t.Fatalf("RunReverseETL() error = %v, want unsigned plan expiry rejection", err)
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
	plan, _, err := a.PreviewConnectorCommandPlan(ctx, plan.ID, nil)
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
	if _, _, err := a.PreviewConnectorCommandPlan(ctx, plan.ID, nil); err == nil {
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
			ConfigurationDigest: req.Config.ConfigurationDigest, Batchable: true, Scope: req.Config.WriteApprovalScope,
			Confirmation: connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive},
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
	seedWarehouseTableRows(t, root, "repo_deletes", `{"id":"row-1"}`)
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
