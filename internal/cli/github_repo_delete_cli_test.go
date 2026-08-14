package cli_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/cli"
)

// The captain's condition on restoring `pm github repo delete` under its
// gh-familiar name is that a scripted, non-interactive caller cannot get itself
// to execution. `--json` is that caller: it is the documented agent surface, and
// it must never hand back the approval token the run needs. This walks the whole
// scripted path — plan, preview, approve-without-confirmation, confirm, replay —
// asserting the request count at every stage, because a gate that rejects after
// dispatching has already deleted the repository.
func TestGitHubRepoDeleteScriptedRunCannotObtainAGrant(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Method != http.MethodDelete {
			t.Errorf("request method = %s, want DELETE", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	root := t.TempDir()
	runCLIForReverseTest(t, []string{"init", "--root", root, "--json"})
	runCLIForReverseTest(t, []string{
		"credentials", "add", "github-local",
		"--connector", "github",
		"--config", "owner=acme",
		"--config", "repo=widgets",
		"--config", "public_access=true",
		"--config", "base_url=" + server.URL,
		"--root", root,
		"--json",
	})

	var planStdout, planStderr bytes.Buffer
	code := cli.Run([]string{
		"github", "repo", "delete",
		"--credential", "github-local",
		"--root", root,
		"--json",
	}, &planStdout, &planStderr)
	if code != 0 {
		t.Fatalf("github repo delete plan code=%d stdout=%s stderr=%s", code, planStdout.String(), planStderr.String())
	}
	var planned struct {
		Plan struct {
			ID                    string `json:"id"`
			ApprovalToken         string `json:"approval_token"`
			ConfirmationChallenge string `json:"confirmation_challenge"`
		} `json:"plan"`
	}
	if err := json.Unmarshal(planStdout.Bytes(), &planned); err != nil {
		t.Fatalf("decode plan envelope: %v\n%s", err, planStdout.String())
	}
	if planned.Plan.ConfirmationChallenge != "destructive" {
		t.Fatalf("confirmation_challenge = %q, want destructive", planned.Plan.ConfirmationChallenge)
	}
	if planned.Plan.ApprovalToken != "" {
		t.Fatal("--json plan output handed a scripted caller an approval token")
	}
	if calls != 0 {
		t.Fatalf("plan dispatched a destructive request; calls=%d", calls)
	}

	var previewStdout, previewStderr bytes.Buffer
	code = cli.Run([]string{
		"github", "repo", "delete",
		"--plan", planned.Plan.ID,
		"--preview",
		"--root", root,
		"--json",
	}, &previewStdout, &previewStderr)
	if code != 0 {
		t.Fatalf("github repo delete preview code=%d stdout=%s stderr=%s", code, previewStdout.String(), previewStderr.String())
	}
	if strings.Contains(previewStdout.String(), `"approval_token": "`) {
		t.Fatalf("--json preview output leaked an approval token:\n%s", previewStdout.String())
	}
	if calls != 0 {
		t.Fatalf("preview dispatched a destructive request; calls=%d", calls)
	}

	// The token exists only on the human-readable preview, which is the point:
	// the scripted surface above could not reach this line on its own.
	var humanPreview, humanPreviewErr bytes.Buffer
	code = cli.Run([]string{
		"github", "repo", "delete",
		"--plan", planned.Plan.ID,
		"--preview",
		"--root", root,
	}, &humanPreview, &humanPreviewErr)
	if code != 0 {
		t.Fatalf("github repo delete human preview code=%d stdout=%s stderr=%s", code, humanPreview.String(), humanPreviewErr.String())
	}
	token := extractReverseField(t, humanPreview.String(), `Approval token: (\S+)`)

	var deniedStdout, deniedStderr bytes.Buffer
	code = runCLIWithApprovalStdin(t, []string{
		"github", "repo", "delete",
		"--plan", planned.Plan.ID,
		"--approval-token-stdin",
		"--root", root,
		"--json",
	}, token+"\n", &deniedStdout, &deniedStderr)
	if code == 0 || !strings.Contains(strings.ToLower(deniedStdout.String()+deniedStderr.String()), "confirmation") {
		t.Fatalf("repo delete ran without --confirm: code=%d stdout=%s stderr=%s", code, deniedStdout.String(), deniedStderr.String())
	}
	if calls != 0 {
		t.Fatalf("unconfirmed repo delete dispatched a request; calls=%d", calls)
	}

	var runStdout, runStderr bytes.Buffer
	code = runCLIWithApprovalStdin(t, []string{
		"github", "repo", "delete",
		"--plan", planned.Plan.ID,
		"--approval-token-stdin",
		"--confirm", "destructive",
		"--root", root,
		"--json",
	}, token+"\n", &runStdout, &runStderr)
	if code != 0 {
		t.Fatalf("confirmed repo delete code=%d stdout=%s stderr=%s", code, runStdout.String(), runStderr.String())
	}
	if calls != 1 {
		t.Fatalf("confirmed repo delete call count=%d, want 1", calls)
	}

	var replayStdout, replayStderr bytes.Buffer
	code = runCLIWithApprovalStdin(t, []string{
		"github", "repo", "delete",
		"--plan", planned.Plan.ID,
		"--approval-token-stdin",
		"--confirm", "destructive",
		"--root", root,
		"--json",
	}, token+"\n", &replayStdout, &replayStderr)
	if code == 0 {
		t.Fatalf("repo delete replayed a spent grant: stdout=%s", replayStdout.String())
	}
	if calls != 1 {
		t.Fatalf("replayed repo delete call count=%d, want 1", calls)
	}
}
