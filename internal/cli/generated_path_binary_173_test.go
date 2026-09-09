package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"os/exec"
	"reflect"
	"sort"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/bundleregistry"
	"polymetrics.ai/internal/connectors/commandrunner"
)

func TestCurrentBitbucketGeneratedPathsBinary173(t *testing.T) {
	registry := bundleregistry.New()
	connector, ok := registry.Get("bitbucket")
	if !ok {
		t.Fatal("production Bitbucket registry entry absent")
	}
	surface := connector.(connectors.CommandSurfaceProvider).CommandSurface()
	want := map[string]string{"repositories list": "implemented", "repositories create": "implemented", "repositories delete": "implemented", "search code": "planned", "downloads get": "planned"}
	got := map[string]string{}
	for _, command := range surface.Commands {
		got[command.Path] = command.Availability
		path, err := commandrunner.CommandPathSegments(command.Path)
		if err != nil {
			t.Fatal(err)
		}
		err = commandrunner.Preflight(connector, path)
		if command.Availability == "implemented" && err != nil {
			t.Fatalf("production binding %s: %v", command.Path, err)
		}
		if command.Availability == "planned" {
			var blocked *commandrunner.BlockedCommandError
			if !errors.As(err, &blocked) || blocked.Availability != "planned" {
				t.Fatalf("partial production binding lost block: %v", err)
			}
		}
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("current exact source-linked command set differs: %v", got)
	}
	binary := buildTransportPM(t)
	root := t.TempDir()
	if out, err := runTransportPM(binary, "", "init", "--root", root, "--json"); err != nil {
		t.Fatalf("fresh CLI init: %v %s", err, out)
	}
	invoke := func(args []string) (string, string, error) {
		command := exec.CommandContext(t.Context(), binary, args...)
		var out, diag bytes.Buffer
		command.Stdout, command.Stderr = &out, &diag
		err := command.Run()
		return out.String(), diag.String(), err
	}
	names := make([]string, 0, len(want))
	for name := range want {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			args := append([]string{"bitbucket"}, strings.Split(name, " ")...)
			args = append(args, "--root", root, "--json")
			out, stderr, err := invoke(args)
			if err == nil {
				t.Fatal("credential-free command unexpectedly executed")
			}
			var result struct {
				Kind  string `json:"kind"`
				Error struct {
					Code    string `json:"code"`
					Message string `json:"message"`
				} `json:"error"`
			}
			if err := json.Unmarshal([]byte(out), &result); err != nil {
				t.Fatalf("actual CLI error envelope: %v %s", err, out)
			}
			if result.Kind != "Error" || strings.TrimSpace(stderr) != "error: "+result.Error.Message {
				t.Fatalf("binary stdout/stderr error parity differs: %s %s", out, stderr)
			}
			if want[name] == "implemented" {
				if result.Error.Message != "missing --credential" {
					t.Fatalf("implemented command did not reach real credential boundary: %s", out)
				}
			} else if result.Error.Code != "connector_command_blocked" || !strings.Contains(result.Error.Message, "availability=planned") {
				t.Fatalf("declared partial block lost: %s", out)
			}
		})
	}
	for _, path := range [][]string{{"unknown-command"}, {"repositories", "unknown"}, {"repositories", "{repo_slug}"}, {"repositories", "list", "--unknown-flag", "value"}} {
		t.Run("invalid/"+strings.Join(path, "_"), func(t *testing.T) {
			args := append([]string{"bitbucket"}, path...)
			args = append(args, "--root", root, "--json")
			out, stderr, err := invoke(args)
			wantCode := `"usage_error"`
			if len(path) > 2 && path[2] == "--unknown-flag" {
				wantCode = `"validation_error"`
			}
			if err == nil || strings.Contains(out+stderr, "missing --credential") || !strings.Contains(out, wantCode) {
				t.Fatalf("unknown/invalid invocation crossed boundary: %v %s", err, out)
			}
		})
	}
}
