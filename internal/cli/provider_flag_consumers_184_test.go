package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/bundleregistry"
	"polymetrics.ai/internal/connectors/commandrunner"
	"polymetrics.ai/internal/connectors/engine"
)

// This reads the frozen source witnesses, not the alias allocator's output.
// Every optional collision is supplied as well as every required collision.
func TestProviderFlagConsumers184(t *testing.T) {
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
	if len(cases) != 22 {
		t.Fatal("incomplete independent cohort")
	}
	registry := bundleregistry.New()
	bundles := map[string]engine.Bundle{}
	root := t.TempDir()
	var out, diag bytes.Buffer
	if Run([]string{"init", "--root", root, "--json"}, &out, &diag) != 0 {
		t.Fatal("init")
	}
	for _, tc := range cases {
		t.Run(tc.Connector+"/"+tc.Command.Path, func(t *testing.T) {
			connector, ok := registry.Get(tc.Connector)
			if !ok {
				t.Fatal("missing connector")
			}
			var decl connectors.CommandSurfaceCommand
			found := 0
			for _, c := range connector.(connectors.CommandSurfaceProvider).CommandSurface().Commands {
				if c.Path == tc.Command.Path {
					decl = c
					found++
				}
			}
			if found != 1 || decl.Write != tc.Command.Write || decl.Operation != tc.Command.Operation {
				t.Fatal("source command identity changed")
			}
			var selected connectors.CommandSurfaceFlag
			found = 0
			for _, f := range decl.Flags {
				if f.Name == tc.NewName {
					selected = f
					found++
				}
			}
			if found != 1 || selected.MapsTo != tc.MapsTo {
				t.Fatal("source flag identity changed")
			}
			args := productionFixtureArgs178(t, tc.Connector, decl, bundles)
			value := "cp16-provider-file.txt"
			var want any = value
			if len(selected.Values) > 0 {
				value = selected.Values[len(selected.Values)-1]
				want = value
			}
			if selected.Type == "json" {
				value = `{"url":"https://example.invalid/cp16","content_type":"json"}`
				if err := json.Unmarshal([]byte(value), &want); err != nil {
					t.Fatal(err)
				}
			}
			flags := parseFlags(args).values
			flags[tc.NewName] = []string{value}
			// Distinct PM controls must disappear without consuming the provider value.
			flags["plan-name"] = []string{"pm-display-name"}
			flags["limit"] = []string{"7"}
			prepared := connectorCommandFlags(flags)
			if !reflect.DeepEqual(prepared[tc.NewName], []string{value}) || len(prepared["plan-name"]) != 0 || len(prepared["limit"]) != 0 {
				t.Fatal("PM/provider ownership mixed")
			}
			path, err := commandrunner.CommandPathSegments(decl.Path)
			if err != nil {
				t.Fatal(err)
			}
			if decl.Write != "" {
				built, err := commandrunner.BuildWriteCommand(t.Context(), connector, commandrunner.Request{Path: path, Flags: prepared})
				if err != nil {
					t.Fatal(err)
				}
				key, ok := strings.CutPrefix(tc.MapsTo, "record.")
				if !ok {
					t.Fatal("unexpected reviewed write mapping")
				}
				if !reflect.DeepEqual(built.Record[key], want) || built.Write != tc.Command.Write || !built.ApprovalRequired {
					t.Fatalf("mapped record/approval mismatch: field=%s got=%v want=%v write=%s", key, built.Record[key], want, built.Write)
				}
			}
			// Real CLI parsing and App entry must reach the no-credential boundary.
			invocation := append([]string{tc.Connector}, path...)
			for i := 0; i < len(args); i += 2 {
				if args[i] != "--"+tc.NewName {
					invocation = append(invocation, args[i:i+2]...)
				}
			}
			invocation = append(invocation, "--"+tc.NewName, value, "--root", root, "--json")
			var resultOut, resultErr bytes.Buffer
			code := runWithPreflightRegistry(invocation, &resultOut, &resultErr, registry)
			var result struct{ Error struct{ Message string } }
			if err := json.Unmarshal(resultOut.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
			if code == 0 || result.Error.Message != "missing --credential" {
				t.Fatalf("actual CLI boundary: %s %s", resultOut.String(), resultErr.String())
			}
		})
	}
}
