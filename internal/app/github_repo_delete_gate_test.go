package app_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/connectors"
)

// `pm github repo delete` is the most destructive command the parity extraction
// made reachable under its gh-familiar name. The capability itself is not new —
// DELETE /repos/{owner}/{repo} was already reachable as `repo delete-2` — so
// what needs pinning is not that the command exists but that reaching execution
// still costs a preview, the closed typed confirmation, and a grant that is
// spent the moment it is used.
//
// Each stage below asserts the HTTP call count, because a gate that rejects
// after dispatching has already deleted the repository.
func TestGitHubRepoDeleteRequiresConfirmationAndSingleUseGrant(t *testing.T) {
	ctx := context.Background()
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Method != http.MethodDelete {
			t.Errorf("request method = %s, want DELETE", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	a, _ := setupGitHubApp(t, ctx, server.URL)
	plan, _, err := a.PlanConnectorCommand(ctx, app.PlanConnectorCommandRequest{
		Name:       "delete_repo",
		Connector:  "github",
		Credential: "github-local",
		Path:       []string{"repo", "delete"},
	})
	if err != nil {
		t.Fatalf("PlanConnectorCommand(repo delete) error = %v", err)
	}
	if plan.ConfirmationChallenge != string(connectors.ConfirmationKindDestructive) {
		t.Fatalf("ConfirmationChallenge = %q, want destructive", plan.ConfirmationChallenge)
	}
	if plan.ApprovalToken != "" {
		t.Fatal("planning repo delete minted an approval token; the preview must come first")
	}
	if calls != 0 {
		t.Fatalf("planning repo delete dispatched a request; calls=%d", calls)
	}

	if _, err := a.RunReverseETL(ctx, app.RunReverseETLRequest{
		PlanID:       plan.ID,
		Confirmation: connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive},
	}); err == nil {
		t.Fatal("RunReverseETL() ran repo delete with a confirmation but no grant")
	}
	if calls != 0 {
		t.Fatalf("unapproved repo delete dispatched a request; calls=%d", calls)
	}

	previewed, _, err := a.PreviewConnectorCommandPlan(ctx, plan.ID, nil)
	if err != nil {
		t.Fatalf("PreviewConnectorCommandPlan() error = %v", err)
	}
	if previewed.ApprovalToken == "" {
		t.Fatal("preview issued no approval token for repo delete")
	}
	if calls != 0 {
		t.Fatalf("previewing repo delete dispatched a request; calls=%d", calls)
	}

	_, err = a.RunReverseETL(ctx, app.RunReverseETLRequest{
		PlanID:        previewed.ID,
		ApprovalToken: previewed.ApprovalToken,
	})
	if err == nil {
		t.Fatal("RunReverseETL() ran repo delete with a grant but no typed confirmation")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "confirmation") {
		t.Fatalf("RunReverseETL() error = %v, want confirmation rejection", err)
	}
	if calls != 0 {
		t.Fatalf("unconfirmed repo delete dispatched a request; calls=%d", calls)
	}

	run, err := a.RunReverseETL(ctx, app.RunReverseETLRequest{
		PlanID:        previewed.ID,
		ApprovalToken: previewed.ApprovalToken,
		Confirmation:  connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive},
	})
	if err != nil {
		t.Fatalf("RunReverseETL() with preview, grant and confirmation error = %v", err)
	}
	if run.RecordsSucceeded != 1 || run.RecordsFailed != 0 {
		t.Fatalf("run result = %+v, want one success", run)
	}
	if calls != 1 {
		t.Fatalf("confirmed repo delete call count = %d, want 1", calls)
	}

	if _, err := a.RunReverseETL(ctx, app.RunReverseETLRequest{
		PlanID:        previewed.ID,
		ApprovalToken: previewed.ApprovalToken,
		Confirmation:  connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive},
	}); err == nil {
		t.Fatal("RunReverseETL() replayed a spent repo delete grant")
	}
	if calls != 1 {
		t.Fatalf("replayed repo delete call count = %d, want 1", calls)
	}
}

// `repo archive` and `repo unarchive` ride PATCH, so no method makes them
// destructive by construction, and the captain's decision classes them with
// `repo create` and `secret set` as approval-only writes. They shipped a
// declared `confirm: destructive` anyway, which made `pm github repo archive`
// demand a typed confirmation the decision never granted it.
//
// So the contract asserted here is the safe one, end to end: `plan` mints the
// approval token itself, no preview is required, and `--confirm` is not. What
// survives the reclassification is the pinned `archived` field — it is still
// built declaratively through MapWriteRecord, so the dispatched body is the
// one the preview digest bound, and it is the single field that separates
// archive from unarchive.
func TestGitHubRepoArchiveIsApprovalOnlyAndSendsPinnedField(t *testing.T) {
	for _, tc := range []struct {
		command  string
		archived bool
	}{
		{command: "archive", archived: true},
		{command: "unarchive", archived: false},
	} {
		t.Run(tc.command, func(t *testing.T) {
			ctx := context.Background()
			calls := 0
			var sent map[string]any
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.Method != http.MethodPatch {
					t.Errorf("request method = %s, want PATCH", r.Method)
				}
				if err := json.NewDecoder(r.Body).Decode(&sent); err != nil {
					t.Errorf("decode request body: %v", err)
				}
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{}`))
			}))
			defer server.Close()

			a, _ := setupGitHubApp(t, ctx, server.URL)
			plan, _, err := a.PlanConnectorCommand(ctx, app.PlanConnectorCommandRequest{
				Name:       "repo_" + tc.command,
				Connector:  "github",
				Credential: "github-local",
				Path:       []string{"repo", tc.command},
			})
			if err != nil {
				t.Fatalf("PlanConnectorCommand(repo %s) error = %v", tc.command, err)
			}
			if plan.ConfirmationChallenge != "" {
				t.Fatalf("repo %s ConfirmationChallenge = %q, want none for an approval-only write", tc.command, plan.ConfirmationChallenge)
			}
			if plan.ApprovalToken == "" {
				t.Fatalf("planning repo %s issued no approval token; a safe write is approved from its plan", tc.command)
			}

			// The preview stays available, and must publish the same empty
			// confirmation the plan did: a preview that names a gate the
			// execution will not meet is the drift this reclassification fixes.
			previewed, _, err := a.PreviewConnectorCommandPlan(ctx, plan.ID, nil)
			if err != nil {
				t.Fatalf("PreviewConnectorCommandPlan(repo %s) error = %v", tc.command, err)
			}
			if previewed.ConfirmationChallenge != "" {
				t.Fatalf("previewing repo %s declared confirmation %q, want none", tc.command, previewed.ConfirmationChallenge)
			}
			if calls != 0 {
				t.Fatalf("previewing repo %s dispatched a request; calls=%d", tc.command, calls)
			}

			// The approval gate itself stays: no token, no write.
			if _, err := a.RunReverseETL(ctx, app.RunReverseETLRequest{PlanID: plan.ID}); err == nil {
				t.Fatalf("RunReverseETL() ran repo %s with no approval token", tc.command)
			}
			if calls != 0 {
				t.Fatalf("unapproved repo %s dispatched a request; calls=%d", tc.command, calls)
			}

			// The plan-time token executes, with no preview and no --confirm.
			run, err := a.RunReverseETL(ctx, app.RunReverseETLRequest{
				PlanID:        plan.ID,
				ApprovalToken: plan.ApprovalToken,
			})
			if err != nil {
				t.Fatalf("RunReverseETL(repo %s) error = %v", tc.command, err)
			}
			if run.RecordsSucceeded != 1 || calls != 1 {
				t.Fatalf("repo %s run = %+v calls = %d, want one success and one call", tc.command, run, calls)
			}
			if len(sent) != 1 || sent["archived"] != tc.archived {
				t.Fatalf("repo %s dispatched body = %+v, want exactly archived=%v", tc.command, sent, tc.archived)
			}
		})
	}
}

// `repo create` and `secret set` are the captain's approval-only creates. They
// keep the plan/preview/approve contract and must not acquire a typed
// destructive challenge, which would be as much a drift from that decision as
// dropping one from a delete.
func TestGitHubApprovalOnlyCreatesCarryNoTypedChallenge(t *testing.T) {
	ctx := context.Background()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	a, _ := setupGitHubApp(t, ctx, server.URL)
	for _, tc := range []struct {
		name  string
		path  []string
		flags map[string][]string
	}{
		{name: "repo create", path: []string{"repo", "create"}, flags: map[string][]string{"name": {"widgets"}}},
		{
			name:  "secret set",
			path:  []string{"secret", "set"},
			flags: map[string][]string{"secret-name": {"TOKEN"}, "encrypted-value": {"c2VjcmV0"}, "key-id": {"568250167242549743"}},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			plan, _, err := a.PlanConnectorCommand(ctx, app.PlanConnectorCommandRequest{
				Name:       strings.ReplaceAll(tc.name, " ", "_"),
				Connector:  "github",
				Credential: "github-local",
				Path:       tc.path,
				Flags:      tc.flags,
			})
			if err != nil {
				t.Fatalf("PlanConnectorCommand(%s) error = %v", tc.name, err)
			}
			if plan.ConfirmationChallenge != "" {
				t.Fatalf("%s ConfirmationChallenge = %q, want none for an approval-only create", tc.name, plan.ConfirmationChallenge)
			}
			if plan.ApprovalToken == "" {
				t.Fatalf("%s issued no approval token; the approval gate must stay", tc.name)
			}
		})
	}
}
