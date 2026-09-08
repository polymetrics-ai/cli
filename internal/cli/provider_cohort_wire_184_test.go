package cli

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/bundleregistry"
	"polymetrics.ai/internal/connectors/engine"
)

func TestProviderCohortApprovedWire184(t *testing.T) {
	raw, err := os.ReadFile("../../cmd/connectorgen/testdata/flag-ownership-181.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Connector string            `json:"connector"`
		NewName   string            `json:"new_name"`
		MapsTo    string            `json:"maps_to"`
		Command   engine.CLICommand `json:"command"`
	}
	if err := json.Unmarshal(raw, &cases); err != nil {
		t.Fatal(err)
	}
	registry := bundleregistry.New()
	bundles := map[string]engine.Bundle{}
	checked := 0
	for _, tc := range cases {
		if tc.Command.Write == "" {
			continue
		}
		checked++
		t.Run(tc.Connector+"/"+tc.Command.Path, func(t *testing.T) {
			connector, _ := registry.Get(tc.Connector)
			var decl connectors.CommandSurfaceCommand
			for _, c := range connector.(connectors.CommandSurfaceProvider).CommandSurface().Commands {
				if c.Path == tc.Command.Path {
					decl = c
				}
			}
			args := productionFixtureArgs178(t, tc.Connector, decl, bundles)
			flags := parseFlags(args).values
			var value string
			var want any
			for _, f := range decl.Flags {
				if f.Name == tc.NewName {
					value = "cp16-provider-file.txt"
					want = value
					if len(f.Values) > 0 {
						value = f.Values[len(f.Values)-1]
						want = value
					}
					if f.Type == "json" {
						value = `{"url":"https://example.invalid/cp16","content_type":"json"}`
						if err := json.Unmarshal([]byte(value), &want); err != nil {
							t.Fatal(err)
						}
					}
				}
			}
			flags[tc.NewName] = []string{value}
			if len(tc.Command.APISurface) != 1 {
				t.Fatal("ambiguous frozen wire identity")
			}
			wantPath := tc.Command.APISurface[0].Path
			replacements := map[string]string{"owner": "acme", "repo": "widgets"}
			for _, f := range decl.Flags {
				v := flags[f.Name]
				if len(v) == 1 {
					replacements[strings.TrimPrefix(f.MapsTo, "record.")] = v[0]
				}
			}
			for k, v := range replacements {
				wantPath = strings.ReplaceAll(wantPath, "{"+k+"}", v)
			}
			if strings.Contains(wantPath, "{") {
				t.Fatalf("unresolved independent path %s", wantPath)
			}
			if tc.Connector == "gitlab" {
				wantPath = "/api/v4" + wantPath
			}
			type observation struct {
				method, path string
				body         map[string]any
			}
			requests := make(chan observation, 4)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var body map[string]any
				data, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
				if err != nil {
					http.Error(w, "read", 400)
					return
				}
				if len(data) > 0 {
					if err := json.Unmarshal(data, &body); err != nil {
						http.Error(w, "json", 400)
						return
					}
				}
				requests <- observation{r.Method, r.URL.Path, body}
				w.Header().Set("Content-Type", "application/json")
				if r.Method == "POST" {
					w.WriteHeader(http.StatusCreated)
				}
				_, _ = w.Write([]byte(`{"id":17}`))
			}))
			defer server.Close()
			root := t.TempDir()
			run := func(args []string, approval string) (int, string, string) {
				t.Helper()
				var out, diag bytes.Buffer
				code := runWithPreflightRegistryAndApprovalReader(args, &out, &diag, registry, strings.NewReader(approval))
				return code, out.String(), diag.String()
			}
			if code, _, _ := run([]string{"init", "--root", root, "--json"}, ""); code != 0 {
				t.Fatal("init")
			}
			baseURL := server.URL
			if tc.Connector == "gitlab" {
				baseURL += "/api/v4"
			}
			cred := []string{"credentials", "add", "fixture", "--connector", tc.Connector, "--config", "base_url=" + baseURL, "--root", root, "--json"}
			if tc.Connector == "github" {
				cred = append(cred, "--config", "public_access=true", "--config", "owner=acme", "--config", "repo=widgets")
			} else {
				t.Setenv("PM_CP16_WIRE_184", "synthetic-local-fixture")
				cred = append(cred, "--from-env", "access_token=PM_CP16_WIRE_184")
			}
			if code, _, diag := run(cred, ""); code != 0 {
				t.Fatalf("fixture credential: %s", diag)
			}
			base := append([]string{tc.Connector}, strings.Fields(tc.Command.Path)...)
			invocation := append([]string{}, base...)
			for i := 0; i < len(args); i += 2 {
				if args[i] != "--"+tc.NewName {
					invocation = append(invocation, args[i:i+2]...)
				}
			}
			invocation = append(invocation, "--"+tc.NewName, value, "--credential", "fixture", "--plan-name", "pm-owned-display", "--root", root)
			code, out, diag := run(invocation, "")
			if code != 0 {
				t.Fatalf("plan: %s %s", out, diag)
			}
			planMatch := regexp.MustCompile(`Created connector command plan (\S+)`).FindStringSubmatch(out)
			if len(planMatch) != 2 {
				t.Fatal("missing plan identity")
			}
			plan := planMatch[1]
			tokenMatch := regexp.MustCompile(`Approval token: (\S+)`).FindStringSubmatch(out)
			if len(requests) != 0 {
				t.Fatal("plan sent request")
			}
			continuation := append(append([]string{}, base...), "--plan", plan, "--root", root)
			for _, f := range decl.Flags {
				for _, v := range flags[f.Name] {
					continuation = append(continuation, "--"+f.Name, v)
				}
			}
			code, out, diag = run(append(append([]string{}, continuation...), "--preview"), "")
			if code != 0 {
				t.Fatalf("preview: %s %s", out, diag)
			}
			if len(tokenMatch) != 2 {
				tokenMatch = regexp.MustCompile(`Approval token: (\S+)`).FindStringSubmatch(out)
			}
			if len(tokenMatch) != 2 {
				t.Fatal("preview omitted approval")
			}
			if len(requests) != 0 {
				t.Fatal("preview sent request")
			}
			approved := append(append([]string{}, continuation...), "--approval-token-stdin", "--confirm", "destructive", "--json")
			code, out, diag = run(approved, tokenMatch[1]+"\n")
			if code != 0 {
				t.Fatalf("apply: %s %s", out, diag)
			}
			if !strings.Contains(out, `"records_succeeded": 1`) {
				t.Fatal("missing successful record count")
			}
			select {
			case got := <-requests:
				if got.method != strings.ToUpper(tc.Command.APISurface[0].Method) || got.path != wantPath {
					t.Fatalf("wrong fixed request %s %s want %s %s", got.method, got.path, tc.Command.APISurface[0].Method, wantPath)
				}
				key := strings.TrimPrefix(tc.MapsTo, "record.")
				if strings.Contains(tc.Command.APISurface[0].Path, "{"+key+"}") {
					if _, present := got.body[key]; present {
						t.Fatal("path-only field leaked into body")
					}
				} else if !reflect.DeepEqual(got.body[key], want) {
					t.Fatalf("provider body field %s differs", key)
				}
			default:
				t.Fatal("no actual write")
			}
			if len(requests) != 0 {
				t.Fatal("extra physical send")
			}
		})
	}
	if checked != 21 {
		t.Fatalf("source membership %d want21", checked)
	}
}
