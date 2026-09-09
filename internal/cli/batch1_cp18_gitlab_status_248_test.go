package cli_test

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"testing"
)

type cp18Lane240StatusCase struct {
	name, route, wire string
	flags             []string
	status            int
}

func TestBatch1CP18GitLabLane240StatusWrites(t *testing.T) {
	for _, tc := range []cp18Lane240StatusCase{
		{"postApiV4GroupsIdExport", "/api/v4/groups/{id}/export", "/api/v4/groups/42/export", []string{"--id", "42"}, 202},
		{"postApiV4ProjectsIdMergeRequestsMergeRequestIidStatusChecksExternalStatusCheckIdRetry", "/api/v4/projects/{id}/merge_requests/{merge_request_iid}/status_checks/{external_status_check_id}/retry", "/api/v4/projects/42/merge_requests/7/status_checks/9/retry", []string{"--id", "42", "--merge-request-iid", "7", "--external-status-check-id", "9"}, 202},
		{"postApiV4UserGpgKeysKeyIdRevoke", "/api/v4/user/gpg_keys/{key_id}/revoke", "/api/v4/user/gpg_keys/8/revoke", []string{"--key-id", "8"}, 202},
		{"postApiV4UsersIdGpgKeysKeyIdRevoke", "/api/v4/users/{id}/gpg_keys/{key_id}/revoke", "/api/v4/users/42/gpg_keys/8/revoke", []string{"--id", "42", "--key-id", "8"}, 202},
		{"postApiV4UsersIdSupportPinRevoke", "/api/v4/users/{id}/support_pin/revoke", "/api/v4/users/42/support_pin/revoke", []string{"--id", "42"}, 204},
	} {
		cp18Lane240StatusWrite(t, tc)
	}
}

func cp18Lane240StatusWrite(t *testing.T, tc cp18Lane240StatusCase) {
	t.Run(tc.name, func(t *testing.T) {
		var mu sync.Mutex
		var requests []string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(io.LimitReader(r.Body, 1024))
			if err != nil {
				t.Error(err)
			}
			got := r.Method + " " + r.URL.RequestURI() + " body=" + string(body) + " auth=" + r.Header.Get("Authorization")
			mu.Lock()
			requests = append(requests, got)
			mu.Unlock()
			t.Logf("physical request: %s content-type=%q", got, r.Header.Get("Content-Type"))
			w.WriteHeader(tc.status)
		}))
		t.Cleanup(server.Close)
		root := cp18Lane240Project(t, server.URL)
		path := []string{"gitlab", "api", "direct-op-" + hex.EncodeToString([]byte("POST "+tc.route))}
		args := append(append([]string{}, path...), tc.flags...)
		args = append(args, "--credential", "cp18-lane240", "--root", root, "--json")
		out, diag, code := cp18Lane240Run(args)
		if code != 0 {
			t.Fatalf("production declaration/plan failed before expected status request: code=%d %s %s", code, out, diag)
		}
		var envelope struct {
			Plan struct {
				ID    string `json:"id"`
				Token string `json:"approval_token"`
			} `json:"plan"`
		}
		if err := json.Unmarshal([]byte(out), &envelope); err != nil {
			t.Fatal(err)
		}
		if envelope.Plan.ID == "" || envelope.Plan.Token != "" {
			t.Fatalf("invalid/redaction plan: %s", out)
		}
		checkCount := func(want int) {
			t.Helper()
			mu.Lock()
			defer mu.Unlock()
			if len(requests) != want {
				t.Fatalf("physical sends=%d want=%d", len(requests), want)
			}
		}
		checkCount(0)
		preview := append(append([]string{}, path...), "--plan", envelope.Plan.ID, "--preview", "--root", root)
		preview = append(preview, tc.flags...)
		human, diag, code := cp18Lane240Run(preview)
		if code != 0 {
			t.Fatalf("preview code=%d %s %s", code, human, diag)
		}
		token := extractReverseField(t, human, `Approval token: (\S+)`)
		checkCount(0)
		execute := append(append([]string{}, path...), "--plan", envelope.Plan.ID, "--approval-token-stdin", "--root", root, "--json")
		execute = append(execute, tc.flags...)
		var denied, deniedErr bytes.Buffer
		code = runCLIWithApprovalStdin(t, execute, token+"\n", &denied, &deniedErr)
		if code == 0 || !strings.Contains(strings.ToLower(denied.String()+deniedErr.String()), "confirm") {
			t.Fatalf("missing actual destructive confirmation not refused: code=%d %s %s", code, denied.String(), deniedErr.String())
		}
		checkCount(0)
		execute = append(execute, "--confirm", "destructive")
		var result, errors bytes.Buffer
		code = runCLIWithApprovalStdin(t, execute, token+"\n", &result, &errors)
		if code != 0 {
			t.Fatalf("approved execution code=%d %s %s", code, result.String(), errors.String())
		}
		checkCount(1)
		mu.Lock()
		got := append([]string(nil), requests...)
		mu.Unlock()
		want := []string{"POST " + tc.wire + " body= auth=Bearer cp18-lane240-fake"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("wire=%q want=%q", got, want)
		}
		if err := cp18Lane240Receipt(result.Bytes(), envelope.Plan.ID, tc.status); err != nil {
			t.Fatal(err)
		}
		t.Logf("execution receipt: %s", result.String())
		result.Reset()
		errors.Reset()
		code = runCLIWithApprovalStdin(t, execute, token+"\n", &result, &errors)
		if code == 0 {
			t.Fatalf("spent approval replay succeeded: %s", result.String())
		}
		checkCount(1)
	})
}
