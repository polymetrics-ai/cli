package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/commandrunner"
	"reflect"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors/engine"
)

// Every availability goes through the same canonical alias admission consumer.
// Invalid single declarations must fail before normalization can hide their spelling.
func TestVNextCanonicalCommandPaths173(t *testing.T) {
	for _, availability := range []string{"implemented", "planned", "unsupported", "deferred"} {
		for _, tc := range []struct {
			name, path string
			valid      bool
		}{
			{"authored_alias", "repositories list", true},
			{"parameter_name_alias", "repositories repo-slug delete", true},
			{"raw_parameter", "repositories {repo_slug} delete", false},
			{"slash_endpoint", "repositories /workspace/repo", false},
			{"leading_space", " repositories list", false},
			{"trailing_space", "repositories list ", false},
			{"double_space", "repositories  list", false},
			{"tab", "repositories\tlist", false},
			{"newline", "repositories\nlist", false},
			{"empty", "", false},
		} {
			t.Run(availability+"/"+tc.name, func(t *testing.T) {
				operations := []vNextCanonicalOperation{{Index: 3, ID: "source.retained", Commands: []vNextCanonicalCommand{{Index: 2, Spec: engine.CLICommand{Path: tc.path, Availability: availability}}}}}
				before := operations[0].Commands[0].Spec
				err := validateVNextCanonicalAliases(operations)
				if tc.valid && err != nil {
					t.Fatalf("valid authored path refused: %v", err)
				}
				if !tc.valid && (err == nil || !strings.Contains(err.Error(), "/operations/3/commands/2/path")) {
					t.Fatalf("invalid exact path admitted or wrong coordinate: %v", err)
				}
				if !reflect.DeepEqual(before, operations[0].Commands[0].Spec) {
					t.Fatal("validation silently rewrote declaration")
				}
			})
		}
	}
}

func TestVNextGeneratedParameterPaths173(t *testing.T) {
	cases := []struct{ id, path, alias, param, flag string }{
		{"source.literal", "/widgets/fixed", "widgets fixed", "", ""},
		{"source.id", "/widgets/{widget_id}", "widgets by-id", "widget_id", "widget-id"},
		{"source.slug", "/widgets/{slug}", "widgets by-slug", "slug", "slug"},
	}
	lock := operationDirectReadLockForSemanticAdmissionTest()
	lock.Operations = nil
	for i, tc := range cases {
		parameters := []map[string]any{}
		flags := []map[string]any{}
		if tc.param != "" {
			parameters = append(parameters, map[string]any{"name": tc.param, "in": "path", "type": "string", "required": true})
			flags = append(flags, map[string]any{"name": tc.flag, "maps_to": "path." + tc.param, "type": "string", "required": true})
		}
		operation, err := json.Marshal(map[string]any{"id": tc.id, "kind": "rest_read", "summary": "Retained source fixture", "risk": "low", "approval": "none", "output_policy": "json_redacted", "rest": map[string]any{"method": "GET", "path": tc.path, "max_bytes": 1024, "response": map[string]any{"success_statuses": []string{"200"}}, "parameters": parameters}})
		if err != nil {
			t.Fatal(err)
		}
		command, err := json.Marshal(map[string]any{"path": tc.alias, "summary": "Retained command", "intent": "direct_read", "availability": "implemented", "operation": tc.id, "api_surface": []map[string]any{{"method": "GET", "path": tc.path}}, "output_policy": "json_redacted", "flags": flags})
		if err != nil {
			t.Fatal(err)
		}
		source, err := json.Marshal(map[string]any{"method": "GET", "path": tc.path})
		if err != nil {
			t.Fatal(err)
		}
		lock.Operations = append(lock.Operations, vNextOperationDescriptor{ID: tc.id, Source: source, Operation: operation, Commands: []vNextCommandDescriptor{{Order: i, Command: command}}})
	}
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "acme"), 0700); err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(lock)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "acme", "source.lock.json"), raw, 0600); err != nil {
		t.Fatal(err)
	}
	var out, diag bytes.Buffer
	if code := runLockRender([]string{"lock-render", "acme", "--defs", root}, &out, &diag); code != 0 {
		t.Fatalf("real renderer: %d %s", code, &diag)
	}
	publisher, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	handle, err := publisher.Open(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	defer handle.Release()
	files := map[string][]byte{}
	for _, name := range handle.Files() {
		payload, err := handle.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		files[name] = payload
	}
	bundle, err := engine.Load(newVNextExecutionFS("acme", files), "acme")
	if err != nil {
		t.Fatal(err)
	}
	connector := engine.New(bundle, nil)
	surface := connector.CommandSurface()
	if surface == nil || len(surface.Commands) != len(cases) {
		t.Fatal("generated command set incomplete")
	}
	for _, tc := range cases {
		t.Run(tc.id, func(t *testing.T) {
			var found *connectors.CommandSurfaceCommand
			for i := range surface.Commands {
				if surface.Commands[i].Path == tc.alias {
					found = &surface.Commands[i]
				}
			}
			if found == nil {
				t.Fatal("declared alias disappeared")
			}
			path, err := commandrunner.CommandPathSegments(found.Path)
			if err != nil {
				t.Fatal(err)
			}
			binding, err := engine.ResolveImplementedCommandBinding(bundle, *found)
			if err != nil {
				t.Fatal(err)
			}
			if binding.Path != tc.path || binding.TransportPath != tc.path || binding.Method != "GET" || found.Operation != tc.id {
				t.Fatalf("source target changed: %+v", binding)
			}
			request := commandrunner.Request{Path: path, Flags: map[string][]string{}}
			if tc.param != "" {
				if len(found.Flags) != 1 || found.Flags[0].Name != tc.flag || found.Flags[0].MapsTo != "path."+tc.param || !found.Flags[0].Required {
					t.Fatal("exact path parameter flag lost or swapped")
				}
				request.Flags[tc.flag] = []string{"fixture-value"}
				if err := commandrunner.PreflightRequest(connector, commandrunner.Request{Path: path, Flags: map[string][]string{}}); err == nil {
					t.Fatal("required path input loss was accepted")
				}
			} else if len(found.Flags) != 0 {
				t.Fatal("literal endpoint acquired a parameter")
			}
			if err := commandrunner.PreflightRequest(connector, request); err != nil {
				t.Fatalf("generated parser/resolver path failed: %v", err)
			}
		})
	}
	out.Reset()
	diag.Reset()
	if code := runLockRender([]string{"lock-render", "acme", "--defs", root, "--check"}, &out, &diag); code != 0 {
		t.Fatalf("exact deterministic re-render: %d %s", code, &diag)
	}
}
