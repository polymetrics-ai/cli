package cli

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

func TestAsanaDottedFlagsActualCLI186(t *testing.T) {
	// Source operation174, operation:search_projects_for_workspace. These literal
	// pairs are independent of the shared alias/flag-name implementation.
	pairs := [][2]string{{"completed-at.after", "completed_at.after"}, {"completed-at.before", "completed_at.before"}, {"completed-on.after", "completed_on.after"}, {"completed-on.before", "completed_on.before"}, {"created-at.after", "created_at.after"}, {"created-at.before", "created_at.before"}, {"created-on.after", "created_on.after"}, {"created-on.before", "created_on.before"}, {"due-at.after", "due_at.after"}, {"due-at.before", "due_at.before"}, {"due-on.after", "due_on.after"}, {"due-on.before", "due_on.before"}, {"members.any", "members.any"}, {"members.not", "members.not"}, {"owner.any", "owner.any"}, {"portfolios.any", "portfolios.any"}, {"start-on.after", "start_on.after"}, {"start-on.before", "start_on.before"}, {"teams.any", "teams.any"}}
	type observation struct {
		method, path string
		query        url.Values
	}
	requests := make(chan observation, 2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests <- observation{r.Method, r.URL.Path, r.URL.Query()}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"gid":"project-fixture","name":"exact returned project"}]}`))
	}))
	defer server.Close()
	root := t.TempDir()
	run := func(args []string) (int, string) {
		var out, diag bytes.Buffer
		code := Run(args, &out, &diag)
		return code, out.String() + diag.String()
	}
	if code, _ := run([]string{"init", "--root", root, "--json"}); code != 0 {
		t.Fatal("init")
	}
	t.Setenv("PM_CP16_ASANA_186", "synthetic-local-fixture")
	if code, _ := run([]string{"credentials", "add", "fixture", "--connector", "asana", "--from-env", "access_token=PM_CP16_ASANA_186", "--config", "base_url=" + server.URL, "--root", root, "--json"}); code != 0 {
		t.Fatal("local credential")
	}
	args := []string{"asana", "projects", "search-projects-for-workspace", "--workspace-gid", "workspace-fixture", "--credential", "fixture", "--root", root, "--json"}
	// The retained Asana base declares next_url pagination with size_param=limit, page_size=100.
	want := url.Values{"limit": {"100"}}
	for _, p := range pairs {
		args = append(args, "--"+p[0], "2026-09-01")
		want.Set(p[1], "2026-09-01")
	}
	code, out := run(args)
	if code != 0 {
		t.Fatalf("actual safe parser: %s", out)
	}
	var envelope struct {
		Kind      string `json:"kind"`
		Operation string `json:"operation"`
		Response  struct {
			Data []map[string]any `json:"data"`
		} `json:"response"`
	}
	if err := json.Unmarshal([]byte(out), &envelope); err != nil {
		t.Fatal(err)
	}
	wantRows := []map[string]any{{"gid": "project-fixture", "name": "exact returned project"}}
	if envelope.Kind != "ConnectorCommandDirectRead" || envelope.Operation != "search_projects_for_workspace" || !reflect.DeepEqual(envelope.Response.Data, wantRows) {
		t.Fatal("exact returned operation/project row differs")
	}
	select {
	case got := <-requests:
		if got.method != "GET" || got.path != "/workspaces/workspace-fixture/projects/search" || !reflect.DeepEqual(got.query, want) {
			t.Fatalf("actual original source wire differs: %+v", got)
		}
	default:
		t.Fatal("no request")
	}
	bad := append(append([]string{}, args...), "--completed-at/after", "invalid")
	if code, out := run(bad); code == 0 || !strings.Contains(out, "invalid") {
		t.Fatal("unsafe parser spelling accepted")
	}
	if len(requests) != 0 {
		t.Fatal("unsafe spelling reached provider")
	}
}
