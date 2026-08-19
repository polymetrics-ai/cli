package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"polymetrics.ai/internal/connectors/connsdk"
)

func TestClassifyErrorProvider401IsCredentialError(t *testing.T) {
	classified := classifyError(&connsdk.HTTPError{
		Status: http.StatusUnauthorized,
		URL:    "https://api.example.test/repos/acme/widgets",
		Body:   `{"message":"token=credential-must-not-appear"}`,
	})
	if got, want := classified.category, categoryAuth; got != want {
		t.Fatalf("provider 401 category = %q, want %q", got, want)
	}
	if got, want := classified.code, "credential_error"; got != want {
		t.Fatalf("provider 401 code = %q, want %q", got, want)
	}
	if got, want := classified.message, "provider rejected the credential"; got != want {
		t.Fatalf("provider 401 message = %q, want %q", got, want)
	}
}

func TestClassifyErrorInternalFailureRemainsInternal(t *testing.T) {
	classified := classifyError(errors.New("database invariant failed"))
	if got, want := classified.category, categoryInternal; got != want {
		t.Fatalf("internal failure category = %q, want %q", got, want)
	}
	if got, want := classified.code, "internal_error"; got != want {
		t.Fatalf("internal failure code = %q, want %q", got, want)
	}
}

func TestWriteErrorProvider401RedactsCredential(t *testing.T) {
	const credential = "credential-must-not-appear"
	var stdout, stderr bytes.Buffer
	exitCode := writeError(&stdout, &stderr, &connsdk.HTTPError{
		Status: http.StatusUnauthorized,
		URL:    "https://api.example.test/repos/acme/widgets?token=" + credential,
		Body:   `{"message":"token=` + credential + `"}`,
	}, true)
	if got, want := exitCode, 4; got != want {
		t.Fatalf("provider 401 exit code = %d, want %d", got, want)
	}
	var envelope struct {
		Error struct {
			Category string `json:"category"`
			Code     string `json:"code"`
			Message  string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &envelope); err != nil {
		t.Fatalf("decode provider 401 error envelope: %v", err)
	}
	if got, want := envelope.Error.Category, "auth"; got != want {
		t.Fatalf("provider 401 output category = %q, want %q", got, want)
	}
	if got, want := envelope.Error.Code, "credential_error"; got != want {
		t.Fatalf("provider 401 output code = %q, want %q", got, want)
	}
	if got, want := envelope.Error.Message, "provider rejected the credential"; got != want {
		t.Fatalf("provider 401 output message = %q, want %q", got, want)
	}
	combined := stdout.String() + stderr.String()
	if strings.Contains(combined, credential) {
		t.Fatalf("provider 401 output exposed credential %q", credential)
	}
	if !strings.Contains(combined, "provider rejected the credential") {
		t.Fatalf("provider 401 output = %q, want user-legible credential guidance", combined)
	}
}

func TestFreshBinaryProvider401IsCredentialErrorWithoutWritesOrCheckpointAdvance(t *testing.T) {
	const tokenEnv = "PM_CLI_PROVIDER_401_TOKEN"
	const token = "provider-401-fixture-token"
	var reads, writes atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			reads.Add(1)
		} else {
			writes.Add(1)
		}
		w.WriteHeader(http.StatusUnauthorized)
	}))
	t.Cleanup(server.Close)
	t.Setenv(tokenEnv, token)

	binary := buildTransportPM(t)
	root := filepath.Join(t.TempDir(), "project")
	mustRunGitHubFlowPM(t, binary, "", []string{token}, "init", "--root", root, "--json")
	mustRunGitHubFlowPM(t, binary, "", []string{token},
		"credentials", "add", "github-provider-401", "--connector", "github",
		"--config", "owner=acme", "--config", "repo=widgets", "--config", "auth_type=token",
		"--config", "rate_limit_account=acme", "--config", "base_url="+server.URL,
		"--from-env", "token="+tokenEnv, "--root", root, "--json",
	)
	checkpointBefore := githubFlowCheckpointSnapshot(t, root)

	result := runGitHubFlowPM(t, binary, "", []string{token},
		"github", "issues", "list", "--connection", "github-provider-401", "--root", root, "--json",
	)
	assertGitHubFlowTypedRefusal(t, result, "auth", "credential_error")
	if got, want := reads.Load(), int32(1); got != want {
		t.Fatalf("provider reads = %d, want %d", got, want)
	}
	if got := writes.Load(); got != 0 {
		t.Fatalf("provider writes after rejected credential = %d, want 0", got)
	}
	assertGitHubFlowCheckpointUnchanged(t, root, checkpointBefore, "provider authentication refusal")
}
