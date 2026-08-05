package cli_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"polymetrics.ai/internal/cli"
)

func TestYouTubeAnalyticsReportsDownloadRunsThroughBoundedBinaryExecutor(t *testing.T) {
	payload := []byte("video_id,views\nfixture-video,42\n")
	var tokenRequests atomic.Int32
	var downloadRequests atomic.Int32

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			tokenRequests.Add(1)
			if r.Method != http.MethodPost {
				t.Errorf("token request method = %s, want POST", r.Method)
			}
			if err := r.ParseForm(); err != nil {
				t.Errorf("parse token form: %v", err)
			}
			for key, want := range map[string]string{
				"grant_type":    "refresh_token",
				"client_id":     "fixture-client-id",
				"refresh_token": "fixture-refresh-token",
			} {
				if got := r.Form.Get(key); got != want {
					t.Errorf("token form %s = %q, want %q", key, got, want)
				}
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"fixture-access-token","token_type":"Bearer","expires_in":3600}`))
		case "/v1/media/jobs/job_fixture/reports/report_fixture":
			downloadRequests.Add(1)
			if r.Method != http.MethodGet {
				t.Errorf("download request method = %s, want GET", r.Method)
			}
			if got := r.URL.Query().Get("alt"); got != "media" {
				t.Errorf("download alt query = %q, want media", got)
			}
			if got := r.Header.Get("Authorization"); got != "Bearer fixture-access-token" {
				t.Errorf("download authorization = %q, want refreshed bearer token", got)
			}
			w.Header().Set("Content-Type", "text/csv")
			w.Header().Set("Content-Disposition", `attachment; filename="provider-report.csv"`)
			_, _ = w.Write(payload)
		default:
			http.NotFound(w, r)
			t.Errorf("unexpected provider request path %q", r.URL.Path)
		}
	}))
	t.Cleanup(server.Close)

	oldTransport := http.DefaultTransport
	http.DefaultTransport = server.Client().Transport
	t.Cleanup(func() { http.DefaultTransport = oldTransport })

	root := t.TempDir()
	destination := t.TempDir()
	runCLI(t, []string{"init", "--root", root, "--json"})
	t.Setenv("PM_TEST_YOUTUBE_CLIENT_ID", "fixture-client-id")
	t.Setenv("PM_TEST_YOUTUBE_REFRESH_TOKEN", "fixture-refresh-token")
	runCLI(t, []string{
		"credentials", "add", "youtube-fixture",
		"--connector", "youtube-analytics",
		"--from-env", "client_id=PM_TEST_YOUTUBE_CLIENT_ID",
		"--from-env", "refresh_token=PM_TEST_YOUTUBE_REFRESH_TOKEN",
		"--config", "base_url=" + server.URL,
		"--config", "token_url=" + server.URL + "/token",
		"--root", root,
		"--json",
	})

	stdout, stderr := runCLI(t, []string{
		"youtube-analytics", "reports", "download",
		"--credential", "youtube-fixture",
		"--resource-name", "jobs/job_fixture/reports/report_fixture",
		"--dest-root", destination,
		"--file-name", "youtube-report.csv",
		"--root", root,
		"--json",
	})

	var result struct {
		Kind      string `json:"kind"`
		Connector string `json:"connector"`
		Command   string `json:"command"`
		Operation string `json:"operation"`
		Record    struct {
			FilePath      string `json:"file_path"`
			FileName      string `json:"file_name"`
			FileSizeBytes int64  `json:"file_size_bytes"`
			FileSHA256    string `json:"file_sha256"`
			SourceRef     string `json:"source_ref"`
			Truncated     bool   `json:"truncated"`
		} `json:"record"`
	}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("decode download JSON: %v\n%s", err, stdout)
	}
	if result.Kind != "ConnectorCommandBinaryDownload" || result.Connector != "youtube-analytics" || result.Command != "reports download" || result.Operation != "download_report" {
		t.Fatalf("download envelope = %+v", result)
	}
	if tokenRequests.Load() != 1 || downloadRequests.Load() != 1 {
		t.Fatalf("provider requests: token=%d download=%d, want 1 each", tokenRequests.Load(), downloadRequests.Load())
	}
	wantPath := filepath.Join(destination, "youtube-report.csv")
	if result.Record.FilePath != wantPath || result.Record.FileName != "youtube-report.csv" {
		t.Fatalf("download path record = %+v, want %s", result.Record, wantPath)
	}
	if result.Record.FileSizeBytes != int64(len(payload)) || result.Record.FileSHA256 != fmt.Sprintf("%x", sha256.Sum256(payload)) || result.Record.Truncated {
		t.Fatalf("download integrity record = %+v", result.Record)
	}
	if result.Record.SourceRef != "/v1/media/jobs/job_fixture/reports/report_fixture?alt=media" {
		t.Fatalf("source_ref = %q, want bounded provider-relative media path", result.Record.SourceRef)
	}
	got, err := os.ReadFile(wantPath)
	if err != nil {
		t.Fatalf("read downloaded report: %v", err)
	}
	if !bytes.Equal(got, payload) {
		t.Fatalf("downloaded report = %q, want %q", got, payload)
	}
	info, err := os.Stat(wantPath)
	if err != nil {
		t.Fatalf("stat downloaded report: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("downloaded report mode = %o, want 600", info.Mode().Perm())
	}

	var boundedOut, boundedErr bytes.Buffer
	code := cli.Run([]string{
		"youtube-analytics", "reports", "download",
		"--credential", "youtube-fixture",
		"--resource-name", "jobs/job_fixture/reports/report_fixture",
		"--dest-root", destination,
		"--file-name", "too-large.csv",
		"--max-bytes", "8",
		"--root", root,
		"--json",
	}, &boundedOut, &boundedErr)
	if code == 0 || !strings.Contains(boundedOut.String()+boundedErr.String(), "too large") {
		t.Fatalf("bounded download code=%d stdout=%s stderr=%s", code, boundedOut.String(), boundedErr.String())
	}
	if _, err := os.Stat(filepath.Join(destination, "too-large.csv")); !os.IsNotExist(err) {
		t.Fatalf("rejected oversized download left a destination file: %v", err)
	}
	if tokenRequests.Load() != 2 || downloadRequests.Load() != 2 {
		t.Fatalf("provider requests after bounded rejection: token=%d download=%d, want 2 each", tokenRequests.Load(), downloadRequests.Load())
	}
	for _, secret := range []string{"fixture-client-id", "fixture-refresh-token", "fixture-access-token"} {
		if strings.Contains(stdout+stderr+boundedOut.String()+boundedErr.String(), secret) {
			t.Fatalf("CLI output leaked synthetic secret %q", secret)
		}
	}
	t.Logf("end-user JSON output:\n%s", strings.TrimSpace(stdout))
	t.Logf("downloaded file contents:\n%s", strings.TrimSpace(string(got)))
	t.Logf("bounded oversized-download refusal:\n%s", strings.TrimSpace(boundedOut.String()+boundedErr.String()))
}

func TestYouTubeAnalyticsReportsQueryStaysPlannedForTypedProviderQuery(t *testing.T) {
	var helpOut, helpErr bytes.Buffer
	if code := cli.Run([]string{"youtube-analytics", "reports", "query", "--help"}, &helpOut, &helpErr); code != 0 {
		t.Fatalf("reports query help code = %d stderr=%s", code, helpErr.String())
	}
	help := helpOut.String()
	for _, want := range []string{"AVAILABILITY", "planned", "typed provider-query foundation issue #2985"} {
		if !strings.Contains(help, want) {
			t.Fatalf("reports query help missing %q:\n%s", want, help)
		}
	}
	for _, forbidden := range []string{"provider_search", "rest_write"} {
		if strings.Contains(help, forbidden) {
			t.Fatalf("reports query help claims forbidden executor %q:\n%s", forbidden, help)
		}
	}

	root := t.TempDir()
	runCLI(t, []string{"init", "--root", root, "--json"})
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{
		"youtube-analytics", "reports", "query",
		"--credential", "absent",
		"--root", root,
		"--json",
	}, &stdout, &stderr)
	if code == 0 {
		t.Fatalf("planned reports query unexpectedly executed: %s", stdout.String())
	}
	blocked := stdout.String() + stderr.String()
	for _, want := range []string{"connector_command_blocked", "reports query", "availability=planned"} {
		if !strings.Contains(blocked, want) {
			t.Fatalf("planned reports query output missing %q:\n%s", want, blocked)
		}
	}
	if strings.Contains(blocked, `credential "absent" not found`) {
		t.Fatalf("planned reports query reached credential lookup:\n%s", blocked)
	}
	t.Logf("planned command help excerpt:\n%s", strings.TrimSpace(help))
	t.Logf("planned command execution refusal:\n%s", strings.TrimSpace(blocked))
}

func TestYouTubeAnalyticsGroupsCreateRequiresProviderItemType(t *testing.T) {
	var helpOut, helpErr bytes.Buffer
	if code := cli.Run([]string{"youtube-analytics", "groups", "create", "--help"}, &helpOut, &helpErr); code != 0 {
		t.Fatalf("groups create help code = %d stderr=%s", code, helpErr.String())
	}
	help := helpOut.String()
	for _, want := range []string{
		"--title (string) required",
		"--item-type (enum) required",
		"values=youtube#channel|youtube#playlist|youtube#video|youtubePartner#asset",
		"maps_to=record.contentDetails.itemType",
	} {
		if !strings.Contains(help, want) {
			t.Fatalf("groups create help missing %q:\n%s", want, help)
		}
	}

	root := t.TempDir()
	runCLI(t, []string{"init", "--root", root, "--json"})
	t.Setenv("PM_TEST_YOUTUBE_GROUP_CLIENT_ID", "fixture-client-id")
	t.Setenv("PM_TEST_YOUTUBE_GROUP_REFRESH_TOKEN", "fixture-refresh-token")
	runCLI(t, []string{
		"credentials", "add", "youtube-groups-fixture",
		"--connector", "youtube-analytics",
		"--from-env", "client_id=PM_TEST_YOUTUBE_GROUP_CLIENT_ID",
		"--from-env", "refresh_token=PM_TEST_YOUTUBE_GROUP_REFRESH_TOKEN",
		"--root", root,
		"--json",
	})

	var missingOut, missingErr bytes.Buffer
	code := cli.Run([]string{
		"youtube-analytics", "groups", "create",
		"--credential", "youtube-groups-fixture",
		"--title", "Owned fixture group",
		"--root", root,
		"--json",
	}, &missingOut, &missingErr)
	if code == 0 || !strings.Contains(missingOut.String()+missingErr.String(), "missing required flag --item-type") {
		t.Fatalf("groups create without item type code=%d stdout=%s stderr=%s", code, missingOut.String(), missingErr.String())
	}

	var planOut, planErr bytes.Buffer
	code = cli.Run([]string{
		"youtube-analytics", "groups", "create",
		"--credential", "youtube-groups-fixture",
		"--title", "Owned fixture group",
		"--item-type", "youtube#video",
		"--root", root,
		"--json",
	}, &planOut, &planErr)
	if code != 0 {
		t.Fatalf("groups create plan code=%d stdout=%s stderr=%s", code, planOut.String(), planErr.String())
	}
	for _, want := range []string{
		`"kind": "ConnectorCommandWritePlan"`,
		`"connector_command": "groups create"`,
		`"action": "create_group"`,
		`"approval_required": true`,
	} {
		if !strings.Contains(planOut.String(), want) {
			t.Fatalf("groups create plan missing %q:\n%s", want, planOut.String())
		}
	}
	t.Logf("provider-required groups.create help excerpt:\n%s", strings.TrimSpace(help))
	t.Logf("approval-gated groups.create plan:\n%s", strings.TrimSpace(planOut.String()))
}
