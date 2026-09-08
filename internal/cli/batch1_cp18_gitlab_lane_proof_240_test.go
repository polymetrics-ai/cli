package cli_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/cli"
)

func cp18Lane240Project(t *testing.T, base string) string {
	t.Helper()
	root := t.TempDir()
	runCLI(t, []string{"init", "--root", root, "--json"})
	t.Setenv("PM_CP18_LANE240_TOKEN", "cp18-lane240-fake")
	runCLI(t, []string{"credentials", "add", "cp18-lane240", "--connector", "gitlab", "--from-env", "access_token=PM_CP18_LANE240_TOKEN", "--config", "base_url=" + base + "/api/v4", "--root", root, "--json"})
	return root
}
func cp18Lane240Run(args []string) (string, string, int) {
	var out, diag bytes.Buffer
	code := cli.Run(args, &out, &diag)
	return out.String(), diag.String(), code
}

// Existing managed source/destination mode guards are tested separately from
// the seven lanes. A accepted CreateConnection is admission evidence only;
// execution is covered by the retained warehouse/destination tests.
func TestBatch1CP18GitLabLane240ManagedModeAdmission(t *testing.T) {
	for _, role := range []string{"source", "destination"} {
		for _, mode := range []string{"full_overwrite", "full_append", "incremental_append", "incremental_upsert", "incremental_dedupe", "incremental_dedupe_history", "change_capture"} {
			t.Run(role+"/"+mode, func(t *testing.T) {
				fixture := newGitLabTransportFixture(t, false)
				root := fixture.setupProject(t)
				application, err := app.Open(root)
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = application.Close() })
				dst := app.EndpointConfig{Connector: "warehouse", Credential: "warehouse-local"}
				action := ""
				if role == "destination" {
					dst = app.EndpointConfig{Connector: "gitlab", Credential: "gitlab-local"}
					action = fixture.deleteAction
				}
				connection, createErr := application.CreateConnection(context.Background(), app.CreateConnectionRequest{Name: "cp18_mode_240", Source: app.EndpointConfig{Connector: "gitlab", Credential: "gitlab-local"}, Destination: dst, Streams: map[string]app.StreamConfig{"groups": {SyncMode: mode, DestinationAction: action}}})
				err = createErr
				fixture.assertNoCapturedRequests(t)
				if role == "source" && mode == "incremental_append" {
					if err != nil || !connection.Streams["groups"].LegacyCompatibility {
						t.Fatalf("explicit legacy incremental_append classification changed: connection=%+v err=%v", connection, err)
					}
					t.Log("incremental_append persisted as legacy_compatibility=true; admission only, no closed managed mode or execution claim")
					return
				}
				supported := mode == "full_append" || (role == "source" && mode == "full_overwrite")
				if supported && err != nil {
					t.Fatalf("declared %s/%s refused: %v", role, mode, err)
				}
				if !supported && err == nil {
					_, err = application.RunETL(context.Background(), app.RunETLRequest{Connection: "cp18_mode_240", Stream: "groups", BatchSize: 1})
					fixture.assertNoCapturedRequests(t)
					if err == nil {
						t.Fatalf("undeclared %s/%s executed", role, mode)
					}
				}
				if !supported && !strings.Contains(err.Error(), mode) {
					t.Fatalf("refusal loses mode identity: %v", err)
				}
				t.Logf("mode=%s role=%s admitted=%t error=%v physical_sends=0", mode, role, err == nil, err)
			})
		}
	}
}

// Retained GET /projects, /groups, /users and /issues all expose integer id.
// Explicit Link pages hold three independently expected IDs. The saved legacy
// overwrite path must not publish page one on a subsequent retrieval failure.
func TestBatch1CP18GitLabLane240ETLCollections(t *testing.T) {
	for _, stream := range []string{"projects", "groups", "users", "issues"} {
		for _, fault := range []string{"complete", "second_404", "repeated_link", "malformed_link", "foreign_link"} {
			t.Run(stream+"/"+fault, func(t *testing.T) {
				firstBody, lastBody := `[{"id":101,"created_at":"2026-01-01T00:00:00Z"},{"id":202,"created_at":"2026-01-02T00:00:00Z"}]`, `[{"id":303,"created_at":"2026-01-03T00:00:00Z"}]`
				if stream == "projects" {
					firstBody, lastBody = `[{"id":101,"last_activity_at":"2026-01-01T00:00:00Z"},{"id":202,"last_activity_at":"2026-01-02T00:00:00Z"}]`, `[{"id":303,"last_activity_at":"2026-01-03T00:00:00Z"}]`
				}
				if stream == "issues" {
					firstBody, lastBody = `[{"id":101,"updated_at":"2026-01-01T00:00:00Z","author":{"id":11}},{"id":202,"updated_at":"2026-01-02T00:00:00Z","author":{"id":22}}]`, `[{"id":303,"updated_at":"2026-01-03T00:00:00Z","author":{"id":33}}]`
				}
				var mu sync.Mutex
				var paths []string
				var foreignSends atomic.Int64
				foreign := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { foreignSends.Add(1); w.WriteHeader(500) }))
				t.Cleanup(foreign.Close)
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					body, _ := io.ReadAll(io.LimitReader(r.Body, 1024))
					mu.Lock()
					paths = append(paths, r.URL.RequestURI())
					n := len(paths)
					mu.Unlock()
					if r.Method != "GET" || r.URL.Path != "/api/v4/"+stream || len(body) != 0 || r.Header.Get("Authorization") != "Bearer cp18-lane240-fake" {
						t.Errorf("unexpected ETL wire %s %s body=%q", r.Method, r.URL.RequestURI(), body)
						w.WriteHeader(400)
						return
					}
					if n > 2 {
						t.Errorf("unexpected third request %s", r.URL.RequestURI())
						w.WriteHeader(400)
						return
					}
					expected := "per_page=50"
					if n == 2 {
						expected = "page=2&per_page=50"
					}
					if r.URL.RawQuery != expected {
						t.Errorf("query=%s want=%s", r.URL.RawQuery, expected)
					}
					w.Header().Set("Content-Type", "application/json")
					next := "http://" + r.Host + "/api/v4/" + stream + "?page=2&per_page=50"
					if n == 1 {
						w.Header().Set("Link", "<"+next+">; rel=\"next\"")
						_, _ = io.WriteString(w, firstBody)
						return
					}
					switch fault {
					case "second_404":
						w.WriteHeader(404)
						return
					case "repeated_link":
						w.Header().Set("Link", "<"+next+">; rel=\"next\"")
					case "malformed_link":
						w.Header().Set("Link", "<http://%zz/next>; rel=\"next\"")
					case "foreign_link":
						w.Header().Set("Link", "<"+foreign.URL+"/next>; rel=\"next\"")
					}
					_, _ = io.WriteString(w, lastBody)
				}))
				t.Cleanup(server.Close)
				root := cp18Lane240Project(t, server.URL)
				runCLI(t, []string{"credentials", "add", "cp18-warehouse240", "--connector", "warehouse", "--config", "path=" + filepath.Join(root, ".polymetrics", "warehouse"), "--root", root, "--json"})
				runCLI(t, []string{"connections", "create", "cp18_etl240", "--source", "gitlab:cp18-lane240", "--destination", "warehouse:cp18-warehouse240", "--stream", stream, "--sync-mode", "full_refresh_overwrite", "--table", "cp18_rows240", "--root", root, "--json"})
				out, diag, code := cp18Lane240Run([]string{"etl", "run", "--connection", "cp18_etl240", "--stream", stream, "--batch-size", "1", "--root", root, "--json"})
				mu.Lock()
				wire := append([]string(nil), paths...)
				mu.Unlock()
				wantWire := []string{"/api/v4/" + stream + "?per_page=50", "/api/v4/" + stream + "?page=2&per_page=50"}
				if !reflect.DeepEqual(wire, wantWire) || foreignSends.Load() != 0 {
					t.Fatalf("wire=%q foreign=%d want=%q code=%d stdout=%s stderr=%s", wire, foreignSends.Load(), wantWire, code, out, diag)
				}
				application, err := app.Open(root)
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = application.Close() })
				rows, queryErr := application.QueryTable(context.Background(), app.QueryTableRequest{Connection: "cp18_etl240", Table: "cp18_rows240", Limit: 10})
				if fault != "complete" {
					if code == 0 || queryErr == nil || len(rows) != 0 {
						t.Fatalf("failed retrieval published: code=%d rows=%v query=%v %s %s", code, rows, queryErr, out, diag)
					}
					t.Logf("reached %s second page; sends=2 foreign=0 no materialized table; diagnostic=%s %s", fault, out, diag)
					return
				}
				if code != 0 || queryErr != nil {
					t.Fatalf("healthy ETL failed: code=%d query=%v %s %s", code, queryErr, out, diag)
				}
				var ids []string
				for _, row := range rows {
					ids = append(ids, fmt.Sprint(row["id"]))
				}
				sort.Strings(ids)
				if !reflect.DeepEqual(ids, []string{"101", "202", "303"}) {
					t.Fatalf("warehouse IDs=%v want101,202,303", ids)
				}
				if stream == "issues" {
					want := map[string]string{"101": "11", "202": "22", "303": "33"}
					for _, row := range rows {
						if fmt.Sprint(row["author_id"]) != want[fmt.Sprint(row["id"])] {
							t.Fatalf("issue author projection=%v", row)
						}
					}
				}
				t.Logf("warehouse exact IDs=%v physical requests=%q", ids, wire)
			})
		}
	}
}

func cp18Lane240Receipt(raw []byte, plan string, status int) error {
	var result struct {
		Kind string `json:"kind"`
		Run  struct {
			ID          string `json:"id"`
			PlanID      string `json:"plan_id"`
			Status      string `json:"status"`
			Staged      int    `json:"records_staged"`
			Succeeded   int    `json:"records_succeeded"`
			Failed      int    `json:"records_failed"`
			Destination struct {
				Written   int `json:"records_written"`
				Failed    int `json:"records_failed"`
				Responses []struct {
					Index   int  `json:"record_index"`
					Status  int  `json:"status"`
					Present bool `json:"body_present"`
					Bytes   int  `json:"body_bytes"`
					Body    any  `json:"body"`
				} `json:"provider_responses"`
			} `json:"destination_result"`
		} `json:"run"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return err
	}
	r := result.Run
	if result.Kind != "ReverseRun" || r.ID == "" || r.PlanID != plan || r.Status != "completed" || r.Staged != 1 || r.Succeeded != 1 || r.Failed != 0 || r.Destination.Written != 1 || r.Destination.Failed != 0 || len(r.Destination.Responses) != 1 {
		return fmt.Errorf("status write receipt identity/count/outcome mismatch: %s", raw)
	}
	response := r.Destination.Responses[0]
	if response.Index != 0 || response.Status != status || response.Present || response.Bytes != 0 || response.Body != nil {
		return fmt.Errorf("bodyless status receipt mismatch: %s", raw)
	}
	return nil
}
func TestBatch1CP18GitLabLane240ReceiptOracle(t *testing.T) {
	const healthy = `{"kind":"ReverseRun","run":{"id":"run-fixture","plan_id":"plan-fixture","status":"completed","records_staged":1,"records_succeeded":1,"records_failed":0,"destination_result":{"records_written":1,"records_failed":0,"provider_responses":[{"record_index":0,"status":202,"body_present":false,"body_bytes":0,"body":null}]}}}`
	for _, tc := range []struct {
		name, old, new string
		valid          bool
	}{
		{"healthy", "", "", true}, {"wrong_status", "\"status\":202", "\"status\":200", false}, {"wrong_plan", "plan-fixture", "other-plan", false}, {"missing_response", "[{\"record_index\":0,\"status\":202,\"body_present\":false,\"body_bytes\":0,\"body\":null}]", "[]", false}, {"failed_record", "\"records_succeeded\":1", "\"records_succeeded\":0", false}, {"invented_body", "\"body\":null", "\"body\":{}", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw := healthy
			if tc.old != "" {
				raw = strings.Replace(raw, tc.old, tc.new, 1)
			}
			err := cp18Lane240Receipt([]byte(raw), "plan-fixture", 202)
			if (err == nil) != tc.valid {
				t.Fatalf("oracle valid=%t error=%v", tc.valid, err)
			}
		})
	}
}
