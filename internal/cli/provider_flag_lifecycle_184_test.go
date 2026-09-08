package cli_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"polymetrics.ai/internal/cli"
)

func TestProviderFlagPlanLifecycle184(t *testing.T) {
	for _, tc := range []struct {
		name, connector, path, wire string
		args                        []string
		body                        map[string]any
	}{
		{"provider-limit", "github", "interaction-limits set", "/repos/acme/widgets/interaction-limits", []string{"--provider-limit", "contributors_only"}, map[string]any{"limit": "contributors_only"}},
		{"provider-plan", "gitlab", "api op-505554202f6170692f76342f656c61737469637365617263685f696e64657865645f6e616d657370616365732f726f6c6c6261636b", "/api/v4/elasticsearch_indexed_namespaces/rollback", []string{"--provider-plan", "opensource", "--percentage", "13"}, map[string]any{"plan": "opensource", "percentage": float64(13)}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			type observed struct {
				method, path string
				body         map[string]any
			}
			requests := make(chan observed, 4)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var body map[string]any
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					http.Error(w, "invalid body", 400)
					return
				}
				requests <- observed{r.Method, r.URL.Path, body}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]any{"id": 17})
			}))
			defer server.Close()
			root := t.TempDir()
			runCLIForReverseTest(t, []string{"init", "--root", root, "--json"})
			baseURL := server.URL
			if tc.connector == "gitlab" {
				baseURL += "/api/v4"
			}
			cred := []string{"credentials", "add", "fixture", "--connector", tc.connector, "--config", "base_url=" + baseURL, "--root", root, "--json"}
			if tc.connector == "github" {
				cred = append(cred, "--config", "public_access=true", "--config", "owner=acme", "--config", "repo=widgets")
			} else {
				t.Setenv("PM_CP16_LOCAL_ACCESS_184", "synthetic-local-fixture")
				cred = append(cred, "--from-env", "access_token=PM_CP16_LOCAL_ACCESS_184")
			}
			runCLIForReverseTest(t, cred)
			base := append([]string{tc.connector}, strings.Fields(tc.path)...)
			invocation := append(append([]string{}, base...), tc.args...)
			invocation = append(invocation, "--credential", "fixture", "--plan-name", "pm-owned-display", "--root", root)
			var out, diag bytes.Buffer
			if cli.Run(invocation, &out, &diag) != 0 {
				t.Fatalf("create plan: %s %s", out.String(), diag.String())
			}
			plan := extractReverseField(t, out.String(), `Created connector command plan (\S+)`)
			token := ""
			destructive := strings.Contains(out.String(), "Confirmation required: --confirm destructive")
			if !destructive {
				token = extractReverseField(t, out.String(), `Approval token: (\S+)`)
			}
			if len(requests) != 0 {
				t.Fatal("plan sent provider request")
			}
			continuation := append(append([]string{}, base...), "--plan", plan, "--root", root, "--json")
			out.Reset()
			diag.Reset()
			if cli.Run(append(append([]string{}, continuation...), "--preview"), &out, &diag) != 0 {
				t.Fatalf("preview: %s %s", out.String(), diag.String())
			}
			if len(requests) != 0 || (token != "" && strings.Contains(out.String(), token)) {
				t.Fatal("preview sent request or leaked approval")
			}
			if destructive {
				textPreview := append(append([]string{}, base...), "--plan", plan, "--preview", "--root", root)
				out.Reset()
				diag.Reset()
				if cli.Run(textPreview, &out, &diag) != 0 {
					t.Fatal("typed preview failed")
				}
				token = extractReverseField(t, out.String(), `Approval token: (\S+)`)
				if len(requests) != 0 {
					t.Fatal("typed preview sent request")
				}
			}
			out.Reset()
			diag.Reset()
			approved := append(append([]string{}, continuation...), "--approval-token-stdin")
			if runCLIWithApprovalStdin(t, approved, "wrong-token\n", &out, &diag) == 0 || len(requests) != 0 {
				t.Fatal("wrong approval crossed side-effect boundary")
			}
			out.Reset()
			diag.Reset()
			if destructive {
				if runCLIWithApprovalStdin(t, approved, token+"\n", &out, &diag) == 0 || len(requests) != 0 {
					t.Fatal("missing confirmation crossed boundary")
				}
				approved = append(approved, "--confirm", "destructive")
				out.Reset()
				diag.Reset()
			}
			if runCLIWithApprovalStdin(t, approved, token+"\n", &out, &diag) != 0 {
				t.Fatalf("approved apply: %s %s", out.String(), diag.String())
			}
			if !strings.Contains(out.String(), `"records_succeeded": 1`) {
				t.Fatalf("missing actual applied-record count: %s", out.String())
			}
			select {
			case got := <-requests:
				if got.method != "PUT" || got.path != tc.wire || !reflect.DeepEqual(got.body, tc.body) {
					t.Fatalf("actual request=%+v want PUT %s %v", got, tc.wire, tc.body)
				}
			default:
				t.Fatal("no actual provider request")
			}
			out.Reset()
			diag.Reset()
			if runCLIWithApprovalStdin(t, approved, token+"\n", &out, &diag) == 0 {
				t.Fatal("consumed approval replay accepted")
			}
			if len(requests) != 0 {
				t.Fatal("replay sent request")
			}
		})
	}
}
