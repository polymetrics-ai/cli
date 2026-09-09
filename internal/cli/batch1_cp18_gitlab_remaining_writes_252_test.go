package cli_test

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
	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
	"polymetrics.ai/internal/warehouse"
	"sort"
	"strings"
	"sync"
	"testing"
)

// Fixture values and expected HTTP are source-derived. Generated command/action
// names are joins only. Unreviewed effect/variant questions remain in CASES;
// this selected baseline-command witness cannot close them by passing.
func TestBatch1CP18GitLabRemaining252WriteContracts(t *testing.T) {
	var fixture struct {
		SourceSHA256 string `json:"source_sha256"`
		SourcePath   string `json:"source_path"`
		Cases        []struct {
			SourceID               string            `json:"source_id"`
			Command                string            `json:"command"`
			Action                 string            `json:"action"`
			Variant                string            `json:"variant"`
			WarehouseRecord        map[string]any    `json:"warehouse_record"`
			Flags                  []string          `json:"flags"`
			ExpectedMethod         string            `json:"expected_method"`
			ExpectedURI            string            `json:"expected_uri"`
			ExpectedAuth           string            `json:"expected_auth"`
			ExpectedHeaders        map[string]string `json:"expected_headers"`
			ExpectedBody           any               `json:"expected_body"`
			ResponseStatus         int               `json:"response_status"`
			ResponseMedia          string            `json:"response_media"`
			ResponseBody           any               `json:"response_body"`
			Confirmation           string            `json:"confirmation"`
			ExpectedUnchanged      bool              `json:"expected_unchanged"`
			ExpectedSuccess        bool              `json:"expected_success"`
			ExpectedAmbiguousSends int               `json:"expected_ambiguous_sends"`
		} `json:"cases"`
	}
	raw, err := os.ReadFile(filepath.Join("testdata", "batch1-cp18-gitlab", "remaining252", "write-contracts.json"))
	if err != nil {
		t.Fatal(err)
	}
	cp18Remaining252CheckSelection(t, "write", raw)
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile(filepath.Join("..", "..", fixture.SourcePath))
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(source)
	if hex.EncodeToString(digest[:]) != fixture.SourceSHA256 {
		t.Fatal("source pin mismatch")
	}
	if len(fixture.Cases) == 0 {
		t.Fatal("empty mutation fixture selection")
	}
	var retained struct {
		Rest struct {
			Operations []struct{ ID, Method, Path string } `json:"operations"`
		} `json:"rest"`
	}
	if err := json.Unmarshal(source, &retained); err != nil {
		t.Fatal(err)
	}
	sourceRoutes := map[string]string{}
	for _, op := range retained.Rest.Operations {
		if _, exists := sourceRoutes[op.ID]; exists {
			t.Fatal("duplicate source identity")
		}
		sourceRoutes[op.ID] = strings.ToUpper(op.Method) + " " + op.Path
	}
	bundle, err := engine.Load(defs.FS, "gitlab")
	if err != nil {
		t.Fatal(err)
	}
	var surface struct {
		Commands []struct {
			Path   string `json:"path"`
			Write  string `json:"write"`
			Intent string `json:"intent"`
		} `json:"commands"`
	}
	if err := json.Unmarshal(bundle.RawCLISurface, &surface); err != nil {
		t.Fatal(err)
	}
	declaredActions := map[string]string{}
	for _, command := range surface.Commands {
		if command.Intent == "direct_write" {
			declaredActions[command.Path] = command.Write
		}
	}
	seen := map[string]bool{}
	for _, tc := range fixture.Cases {
		roles := []string{"direct_write", "reverse_etl"}
		for _, role := range roles {
			faults := []string{"provider_response"}
			if tc.Variant == "selected_source_request_success" {
				faults = append(faults, "ambiguous_after_send")
			}
			for _, fault := range faults {
				t.Run(tc.SourceID+"/"+tc.Variant+"/"+role+"/"+fault, func(t *testing.T) {
					if seen[tc.SourceID+"/"+tc.Variant+"/"+role+"/"+fault] {
						t.Fatal("duplicate source/variant")
					}
					seen[tc.SourceID+"/"+tc.Variant+"/"+role+"/"+fault] = true
					route, exists := sourceRoutes[tc.SourceID]
					if !exists {
						t.Fatal("unretained source identity")
					}
					if declaredActions[tc.Command] != tc.Action || tc.Action == "" {
						t.Fatal("fixture command/action join differs from actual loaded target")
					}
					wantCommand := "api direct-op-" + hex.EncodeToString([]byte(route))
					if tc.Command != wantCommand {
						t.Fatalf("source command identity mismatch: got=%s want=%s", tc.Command, wantCommand)
					}
					type request struct {
						Method, URI, Body, Auth string
						Headers                 http.Header
					}
					var mu sync.Mutex
					var captured []request
					var projectRoot, approvedPlanID string
					var response []byte
					if tc.ResponseBody != nil {
						response, err = json.Marshal(tc.ResponseBody)
						if err != nil {
							t.Fatal(err)
						}
					}
					server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						body, readErr := io.ReadAll(io.LimitReader(r.Body, 1<<20))
						if readErr != nil {
							t.Error(readErr)
						}
						mu.Lock()
						captured = append(captured, request{r.Method, r.URL.RequestURI(), string(body), r.Header.Get("Authorization"), r.Header.Clone()})
						currentRoot, currentPlan := projectRoot, approvedPlanID
						mu.Unlock()
						if _, err := cp18Remaining252DurableFrontier(currentRoot, currentPlan, false); err != nil {
							t.Errorf("at physical send: %v", err)
						}
						if fault == "ambiguous_after_send" {
							hijacker, ok := w.(http.Hijacker)
							if !ok {
								t.Error("fixture cannot reach post-send ambiguity")
								w.WriteHeader(500)
								return
							}
							connection, _, err := hijacker.Hijack()
							if err != nil {
								t.Error(err)
								return
							}
							_ = connection.Close()
							return
						}
						if tc.ResponseMedia != "" {
							w.Header().Set("Content-Type", tc.ResponseMedia)
						}
						w.WriteHeader(tc.ResponseStatus)
						_, _ = w.Write(response)
					}))
					t.Cleanup(server.Close)
					root := cp18Lane240Project(t, server.URL)
					path := append([]string{"gitlab"}, strings.Fields(tc.Command)...)
					invoke := append(append([]string{}, path...), tc.Flags...)
					invoke = append(invoke, "--credential", "cp18-lane240", "--root", root, "--json")
					var tableBefore []byte
					tablePath := filepath.Join(root, ".polymetrics", "warehouse", "cp18_input"+warehouse.TableFileExt)
					if role == "reverse_etl" {
						// Explicit unattributed test input, not a fabricated provider read.
						// The second row is outside the approved limit and must not be sent.
						row := warehouse.Row{}
						for key, value := range tc.WarehouseRecord {
							row[key] = value
						}
						row["cp18_fixture_ordinal"] = 101
						second := warehouse.Row{}
						for key, value := range row {
							second[key] = value
						}
						second["cp18_fixture_ordinal"] = 202
						seedWarehouseTableFixture(t, root, "cp18_input", row, second)
						tableBefore, err = os.ReadFile(tablePath)
						if err != nil {
							t.Fatal(err)
						}
						application, err := app.Open(root)
						if err != nil {
							t.Fatal(err)
						}
						rows, queryErr := application.QueryTable(context.Background(), app.QueryTableRequest{Table: "cp18_input", Connection: warehouse.UnattributedConnection, Limit: 10})
						closeErr := application.Close()
						if queryErr != nil || closeErr != nil || len(rows) != 2 {
							t.Fatalf("warehouse setup: rows=%d query=%v close=%v", len(rows), queryErr, closeErr)
						}
						encoded, err := json.Marshal(rows)
						if err != nil {
							t.Fatal(err)
						}
						expected, err := json.Marshal([]warehouse.Row{row, second})
						if err != nil || !bytes.Equal(encoded, expected) {
							t.Fatalf("warehouse roundtrip differs from source-derived staged rows: got=%s want=%s err=%v", encoded, expected, err)
						}
						invoke = []string{"reverse", "plan", "cp18_reverse", "--source-table", "cp18_input", "--connection", warehouse.UnattributedConnection, "--destination", "gitlab:cp18-lane240", "--action", tc.Action, "--limit", "1", "--root", root, "--json"}
						keys := make([]string, 0, len(tc.WarehouseRecord))
						for key := range tc.WarehouseRecord {
							keys = append(keys, key)
						}
						sort.Strings(keys)
						for _, key := range keys {
							invoke = append(invoke, "--map", key+":"+key)
						}
						if len(keys) == 0 {
							t.Fatal("source fixture has no warehouse mapping; author exact contract before running")
						}
					}
					out, diag, code := cp18Lane240Run(invoke)
					if code != 0 {
						t.Fatalf("plan failed: %d %s %s", code, out, diag)
					}
					var planned struct {
						Plan struct {
							ID    string `json:"id"`
							Token string `json:"approval_token"`
						} `json:"plan"`
					}
					if err := json.Unmarshal([]byte(out), &planned); err != nil {
						t.Fatal(err)
					}
					if planned.Plan.ID == "" || planned.Plan.Token != "" {
						t.Fatal("plan missing identity or leaked approval token")
					}
					checkSends := func(want int) {
						t.Helper()
						mu.Lock()
						defer mu.Unlock()
						if len(captured) != want {
							t.Fatalf("physical sends=%d want%d", len(captured), want)
						}
					}
					checkSends(0)
					mu.Lock()
					projectRoot = root
					approvedPlanID = planned.Plan.ID
					mu.Unlock()
					preview := append(append([]string{}, path...), "--plan", planned.Plan.ID, "--preview", "--root", root)
					preview = append(preview, tc.Flags...)
					if role == "reverse_etl" {
						preview = []string{"reverse", "preview", planned.Plan.ID, "--root", root}
					}
					human, diag, code := cp18Lane240Run(preview)
					if code != 0 {
						t.Fatalf("preview failed: %d %s %s", code, human, diag)
					}
					token := extractReverseField(t, human, `Approval token: (\S+)`)
					checkSends(0)
					execute := append(append([]string{}, path...), "--plan", planned.Plan.ID, "--approval-token-stdin", "--root", root, "--json")
					execute = append(execute, tc.Flags...)
					if role == "reverse_etl" {
						execute = []string{"reverse", "run", planned.Plan.ID, "--approval-token-stdin", "--root", root, "--json"}
					}
					var result, errors bytes.Buffer
					if tc.Confirmation != "" {
						code = runCLIWithApprovalStdin(t, execute, token+"\n", &result, &errors)
						if code == 0 || !strings.Contains(strings.ToLower(result.String()+errors.String()), "confirm") {
							t.Fatal("missing typed confirmation not refused")
						}
						checkSends(0)
						execute = append(execute, "--confirm", tc.Confirmation)
						result.Reset()
						errors.Reset()
					}
					code = runCLIWithApprovalStdin(t, execute, token+"\n", &result, &errors)
					if code != 0 && tc.ExpectedSuccess && fault != "ambiguous_after_send" {
						t.Fatalf("approved execution failed: %d %s %s", code, result.String(), errors.String())
					}
					wantSends := 1
					if fault == "ambiguous_after_send" && tc.ExpectedAmbiguousSends != 0 {
						wantSends = tc.ExpectedAmbiguousSends
					}
					checkSends(wantSends)
					mu.Lock()
					got := captured[0]
					mu.Unlock()
					if got.Method != tc.ExpectedMethod || got.URI != tc.ExpectedURI || got.Auth != tc.ExpectedAuth {
						t.Fatalf("source wire mismatch method=%s URI=%s auth_match=%t want method=%s URI=%s", got.Method, got.URI, got.Auth == tc.ExpectedAuth, tc.ExpectedMethod, tc.ExpectedURI)
					}
					for name, value := range tc.ExpectedHeaders {
						if got.Headers.Get(name) != value {
							t.Errorf("header %s mismatch", name)
						}
					}
					if tc.ExpectedBody == nil {
						if got.Body != "" {
							t.Fatalf("unexpected body=%q", got.Body)
						}
					} else {
						var body any
						if err := json.Unmarshal([]byte(got.Body), &body); err != nil {
							t.Fatal(err)
						}
						actual, _ := json.Marshal(body)
						expected, _ := json.Marshal(tc.ExpectedBody)
						if !bytes.Equal(actual, expected) {
							t.Fatalf("source body mismatch got=%s want=%s", actual, expected)
						}
						if got.Headers.Get("Content-Type") != "application/json" {
							t.Errorf("request media=%q", got.Headers.Get("Content-Type"))
						}
					}
					var receipt struct {
						Kind string `json:"kind"`
						Run  struct {
							ID          string `json:"id"`
							PlanID      string `json:"plan_id"`
							Status      string `json:"status"`
							Succeeded   int    `json:"records_succeeded"`
							Failed      int    `json:"records_failed"`
							Destination struct {
								Written   int `json:"records_written"`
								Unchanged int `json:"records_unchanged"`
								Responses []struct {
									Index   int  `json:"record_index"`
									Status  int  `json:"status"`
									Body    any  `json:"body"`
									Bytes   int  `json:"body_bytes"`
									Present bool `json:"body_present"`
								} `json:"provider_responses"`
							} `json:"destination_result"`
						} `json:"run"`
					}
					if err := json.Unmarshal(result.Bytes(), &receipt); err != nil {
						t.Fatal(err)
					}
					run := receipt.Run
					if fault == "ambiguous_after_send" {
						if receipt.Kind != "ReverseRun" || run.ID == "" || run.PlanID != planned.Plan.ID || run.Succeeded != 0 || run.Failed != 1 || run.Destination.Written != 0 || len(run.Destination.Responses) > 1 {
							t.Fatalf("ambiguous send misreported as a successful/response-confirmed write: code=%d %s %s", code, result.String(), errors.String())
						}
						for _, provider := range run.Destination.Responses {
							if provider.Index != 0 || provider.Status != 0 || provider.Present || provider.Bytes != 0 || provider.Body != nil {
								t.Fatal("ambiguous attempt fabricated a provider response")
							}
						}
						t.Logf("%d source request(s) observed; connection closed before response; records_failed=1 no success receipt; CLI code=%d run status=%s", wantSends, code, run.Status)
					} else {
						wantSucceeded, wantFailed, wantUnchanged := 1, 0, 0
						if tc.ExpectedUnchanged {
							wantSucceeded, wantUnchanged = 0, 1
						}
						if !tc.ExpectedSuccess {
							wantSucceeded, wantFailed = 0, 1
						}
						if receipt.Kind != "ReverseRun" || run.ID == "" || run.PlanID != planned.Plan.ID || (tc.ExpectedSuccess && run.Status != "completed") || run.Succeeded != wantSucceeded || run.Failed != wantFailed || run.Destination.Written != wantSucceeded || run.Destination.Unchanged != wantUnchanged || len(run.Destination.Responses) != 1 {
							t.Fatalf("receipt identity/count/outcome mismatch: %s", result.String())
						}
						provider := run.Destination.Responses[0]
						actual, _ := json.Marshal(provider.Body)
						expected, _ := json.Marshal(tc.ResponseBody)
						if provider.Index != 0 || provider.Status != tc.ResponseStatus || provider.Bytes != len(response) || provider.Present != (len(response) > 0) || !bytes.Equal(actual, expected) {
							t.Fatalf("source response receipt mismatch: %s", result.String())
						}
					}
					terminalBefore, err := cp18Remaining252DurableFrontier(root, planned.Plan.ID, true)
					if err != nil {
						t.Fatal(err)
					}
					var returned struct {
						Run any `json:"run"`
					}
					if err := json.Unmarshal(result.Bytes(), &returned); err != nil {
						t.Fatal(err)
					}
					returnedRun, err := json.Marshal(returned.Run)
					if err != nil || !bytes.Equal(terminalBefore, returnedRun) {
						t.Fatalf("returned run differs from durable receipt: %v", err)
					}
					result.Reset()
					errors.Reset()
					code = runCLIWithApprovalStdin(t, execute, token+"\n", &result, &errors)
					if code == 0 {
						t.Fatal("spent approval replay succeeded")
					}
					checkSends(wantSends)
					terminalAfter, err := cp18Remaining252DurableFrontier(root, planned.Plan.ID, true)
					if err != nil || !bytes.Equal(terminalBefore, terminalAfter) {
						t.Fatalf("replay changed durable terminal history: %v", err)
					}
					if role == "reverse_etl" {
						after, err := os.ReadFile(tablePath)
						if err != nil || !bytes.Equal(tableBefore, after) {
							t.Fatalf("reverse retrieval/execution changed retained source table bytes: %v", err)
						}
					}
				})
			}
		}
	}
}
