package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/bundleregistry"
	"polymetrics.ai/internal/connectors/commandrunner"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
	"reflect"
	"strings"
	"testing"
)

func productionFixtureArgs178(t *testing.T, name string, decl connectors.CommandSurfaceCommand, bundles map[string]engine.Bundle) []string {
	t.Helper()
	var args []string
	for _, flag := range decl.Flags {
		if !flag.Required {
			continue
		}
		value := "fixture"
		switch flag.Type {
		case "integer", "number":
			value = "1"
			if flag.Minimum != nil {
				value = flag.Minimum.String()
			}
		case "json":
			value = productionJSONFixture178(t, name, decl, flag, bundles)
		case "boolean":
			value = "true"
		case "string_array":
			value = "fixture"
		}
		switch flag.Format {
		case "date":
			value = "2026-01-01"
		case "date-time":
			value = "2026-01-01T00:00:00Z"
		}
		if len(flag.Values) > 0 {
			value = flag.Values[0]
		}
		count := 1
		if flag.Type == "string_array" && flag.MinItems > count {
			count = flag.MinItems
		}
		for i := 0; i < count; i++ {
			args = append(args, "--"+flag.Name, value)
		}
	}

	return args
}

func productionJSONFixture178(t *testing.T, name string, decl connectors.CommandSurfaceCommand, flag connectors.CommandSurfaceFlag, bundles map[string]engine.Bundle) string {
	t.Helper()
	bundle, loaded := bundles[name]
	if !loaded {
		var err error
		bundle, err = engine.Load(defs.FS, name)
		if err != nil {
			t.Fatal(err)
		}
		bundles[name] = bundle
	}
	var raw json.RawMessage
	for _, action := range bundle.Writes {
		if action.Name == decl.Write {
			raw = action.RecordSchema
		}
	}
	prefix := "record."
	if decl.Operation != "" {
		prefix = "body."
		for _, op := range bundle.Operations {
			if op.ID == decl.Operation {
				if op.REST != nil {
					raw = op.REST.BodySchema
				}
				if op.GraphQL != nil {
					raw = op.GraphQL.VariablesSchema
				}
			}
		}
	}
	var schema map[string]any
	if err := json.Unmarshal(raw, &schema); err != nil {
		t.Fatalf("%s %s JSON schema: %v", name, decl.Path, err)
	}
	path, ok := strings.CutPrefix(flag.MapsTo, prefix)
	if !ok {
		t.Fatalf("explicit fixture needed for %s %s mapping %s", name, decl.Path, flag.MapsTo)
	}
	for _, part := range strings.Split(path, ".") {
		props, ok := schema["properties"].(map[string]any)
		if !ok {
			t.Fatal("record schema properties absent")
		}
		schema, ok = props[part].(map[string]any)
		if !ok {
			t.Fatalf("schema path absent: %s", path)
		}
	}
	value, err := productionSchemaValue178(schema, 0)
	if err != nil {
		t.Fatalf("%s %s --%s fixture: %v", name, decl.Path, flag.Name, err)
	}
	bytes, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return string(bytes)
}

// Only the declaration shapes actually needed by this sweep are supported.
// Every selected value must pass the independent production schema validator.
func productionSchemaValue178(schema map[string]any, depth int) (any, error) {
	if depth > 12 {
		return nil, fmt.Errorf("explicit fixture required for deep schema")
	}
	var candidates []any
	if types, ok := schema["type"].([]any); ok {
		for _, kind := range types {
			child := map[string]any{}
			for k, v := range schema {
				child[k] = v
			}
			child["type"] = kind
			if v, e := productionSchemaValue178(child, depth+1); e == nil {
				candidates = append(candidates, v)
			}
		}
	}
	if values, ok := schema["enum"].([]any); ok {
		candidates = append(candidates, values...)
	}
	if options, ok := schema["oneOf"].([]any); ok {
		for _, option := range options {
			if child, ok := option.(map[string]any); ok {
				if v, e := productionSchemaValue178(child, depth+1); e == nil {
					candidates = append(candidates, v)
				}
			}
		}
	}
	switch schema["type"] {
	case "object":
		object := map[string]any{}
		props, _ := schema["properties"].(map[string]any)
		required, _ := schema["required"].([]any)
		for _, key := range required {
			child, ok := props[key.(string)].(map[string]any)
			if !ok {
				return nil, fmt.Errorf("missing required schema %s", key)
			}
			v, e := productionSchemaValue178(child, depth+1)
			if e != nil {
				return nil, e
			}
			object[key.(string)] = v
		}
		candidates = append(candidates, object)
	case "array":
		count := 0
		if n, ok := schema["minItems"].(float64); ok {
			count = int(n)
		}
		if count > 100 {
			return nil, fmt.Errorf("explicit fixture required for large array")
		}
		items := []any{}
		for i := 0; i < count; i++ {
			child, ok := schema["items"].(map[string]any)
			if !ok {
				return nil, fmt.Errorf("missing array item schema")
			}
			v, e := productionSchemaValue178(child, depth+1)
			if e != nil {
				return nil, e
			}
			items = append(items, v)
		}
		candidates = append(candidates, items)
	case "string":
		candidates = append(candidates, "fixture", "1", "2026-01-01", "2026-01-01T00:00:00Z", "https://example.invalid/fixture", "fixture@example.invalid")
	case "number", "integer":
		if n, ok := schema["minimum"]; ok {
			candidates = append(candidates, n)
		}
		candidates = append(candidates, float64(1), float64(0))
	case "boolean":
		candidates = append(candidates, true)
	case "null":
		candidates = append(candidates, nil)
	}
	raw, e := json.Marshal(schema)
	if e != nil {
		return nil, e
	}
	compiled, e := engine.CompileSchema(raw)
	if e != nil {
		return nil, e
	}
	for _, v := range candidates {
		if compiled.Validate(v) == nil {
			return v, nil
		}
	}
	return nil, fmt.Errorf("explicit valid fixture needed for schema %s", raw)
}
func TestProductionRegistryFixtures178(t *testing.T) {
	registry := bundleregistry.New()
	root := t.TempDir()
	var initOut, initErr bytes.Buffer
	if code := Run([]string{"init", "--root", root, "--json"}, &initOut, &initErr); code != 0 {
		t.Fatalf("isolated init: %s %s", initOut.String(), initErr.String())
	}
	bundles := map[string]engine.Bundle{}
	checked := 0
	for _, meta := range registry.List() {
		connector, ok := registry.Get(meta.Name)
		if !ok {
			t.Fatal(meta.Name)
		}
		provider, ok := connector.(connectors.CommandSurfaceProvider)
		if !ok || provider.CommandSurface() == nil {
			continue
		}
		for _, decl := range provider.CommandSurface().Commands {
			checked++
			t.Run(meta.Name+"/"+decl.Path, func(t *testing.T) {
				path, e := commandrunner.CommandPathSegments(decl.Path)
				if e != nil {
					t.Fatal(e)
				}
				args := productionFixtureArgs178(t, meta.Name, decl, bundles)
				flags := parseFlags(args)
				e = commandrunner.PreflightRequest(connector, commandrunner.Request{Path: path, Flags: flags.values})
				if decl.Availability == "implemented" && e != nil {
					t.Fatalf("fixture preflight: %v", e)
				}
				var out, diag bytes.Buffer
				invocation := append([]string{meta.Name}, path...)
				carrier, env := productionFixtureCarrier178(decl, args)
				for key, value := range env {
					t.Setenv(key, value)
				}
				invocation = append(invocation, carrier...)
				invocation = append(invocation, "--root", root, "--json")
				loggedArgs := append([]string{}, invocation...)
				for i := range loggedArgs {
					if loggedArgs[i] == root {
						loggedArgs[i] = "<isolated-root>"
					}
				}
				argvJSON, _ := json.Marshal(loggedArgs)
				t.Logf("invocation=%s availability=%s", argvJSON, decl.Availability)
				code := runWithPreflightRegistry(invocation, &out, &diag, registry)
				var result struct {
					Kind  string `json:"kind"`
					Error struct {
						Code    string `json:"code"`
						Message string `json:"message"`
					} `json:"error"`
				}
				if err := json.Unmarshal(out.Bytes(), &result); err != nil {
					t.Fatalf("public CLI envelope: %v %s %s", err, out.String(), diag.String())
				}
				if code == 0 || result.Kind != "Error" {
					t.Fatalf("unexpected execution: %s", out.String())
				}
				if decl.Availability == "implemented" {
					if result.Error.Message != "missing --credential" {
						t.Fatalf("CLI fixture boundary: %s", out.String())
					}
				} else if result.Error.Code != "connector_command_blocked" || !strings.Contains(result.Error.Message, "availability="+decl.Availability) {
					t.Fatalf("CLI declared block: %s", out.String())
				}

			})
		}
	}
	if checked != 13856 {
		t.Fatalf("membership=%d want13856", checked)
	}
	t.Logf("complete declared command fixture count=%d", checked)
}

func TestProductionFixtureRegressions178(t *testing.T) {
	registry := bundleregistry.New()
	connector, ok := registry.Get("asana")
	if !ok {
		t.Fatal("Asana missing")
	}
	var selected connectors.CommandSurfaceCommand
	for _, decl := range connector.(connectors.CommandSurfaceProvider).CommandSurface().Commands {
		if decl.Path == "access-requests create-access-request" {
			selected = decl
		}
	}
	if selected.Write != "create_access_request" {
		t.Fatalf("source-bound write identity changed: %+v", selected)
	}
	args := productionFixtureArgs178(t, "asana", selected, map[string]engine.Bundle{})
	flags := parseFlags(args).values
	var data map[string]any
	if err := json.Unmarshal([]byte(flags["data"][0]), &data); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(data, map[string]any{"target": "fixture"}) {
		t.Fatalf("independent required target witness: %v", data)
	}
	path := []string{"access-requests", "create-access-request"}
	for _, tc := range []struct {
		name, value string
		valid       bool
	}{{"valid required target", `{"target":"fixture"}`, true}, {"original invalid JSON", "fixture", false}, {"missing required target", `{}`, false}, {"wrong target type", `{"target":17}`, false}} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := commandrunner.BuildWriteCommand(t.Context(), connector, commandrunner.Request{Path: path, Flags: map[string][]string{"data": {tc.value}}})
			if (err == nil) != tc.valid {
				t.Fatalf("actual write schema validation valid=%v err=%v", tc.valid, err)
			}
		})
	}
	t.Run("missing required scalar", func(t *testing.T) {
		path := []string{"access-requests", "get-access-requests"}
		if err := commandrunner.PreflightRequest(connector, commandrunner.Request{Path: path}); err == nil || !strings.Contains(err.Error(), "--target") {
			t.Fatalf("missing target not refused: %v", err)
		}
		if err := commandrunner.PreflightRequest(connector, commandrunner.Request{Path: path, Flags: map[string][]string{"target": {"fixture"}}}); err != nil {
			t.Fatal(err)
		}
	})
}

// EnvOnly values use the same declared channel as real callers; all values here
// are generated synthetic fixtures, and no existing credential is consulted.
func productionFixtureCarrier178(decl connectors.CommandSurfaceCommand, args []string) ([]string, map[string]string) {
	envOnly := map[string]bool{}
	for _, flag := range decl.Flags {
		envOnly[flag.Name] = flag.EnvOnly
	}
	var out []string
	env := map[string]string{}
	for i := 0; i < len(args); i += 2 {
		name := strings.TrimPrefix(args[i], "--")
		if envOnly[name] {
			key := fmt.Sprintf("PM_CP16_FIXTURE_178_%d", i)
			env[key] = args[i+1]
			out = append(out, "--from-env", name+"="+key)
		} else {
			out = append(out, args[i], args[i+1])
		}
	}
	return out, env
}
