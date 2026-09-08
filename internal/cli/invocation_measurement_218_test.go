package cli

import (
	"bytes"
	"encoding/json"
	"io"
	"os/exec"
	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/bundleregistry"
	"polymetrics.ai/internal/connectors/commandrunner"
	"polymetrics.ai/internal/connectors/engine"
	"testing"
	"time"
)

// Named membership is independent of the helper under measurement. This is a
// bounded timing observation, not a full-corpus proof or a time-limit assertion.
func TestInvocationMeasurement218(t *testing.T) {
	start := time.Now()
	binary := buildTransportPM(t)
	t.Logf("phase=build_once_warm_build_cache elapsed=%s", time.Since(start))
	start = time.Now()
	if out, err := exec.Command(binary, "version").CombinedOutput(); err != nil || len(out) == 0 {
		t.Fatal("fresh startup control failed")
	}
	t.Logf("phase=fresh_process_version_startup_baseline elapsed=%s (includes minimal command)", time.Since(start))
	for _, member := range []struct{ connector, path string }{{"dockerhub", "repositories list"}, {"github", "autolinks view"}, {"asana", "access-requests create-access-request"}} {
		t.Run(member.connector+"/"+member.path, func(t *testing.T) {
			start := time.Now()
			registry, err := bundleregistry.NewRegistry()
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("phase=registry_metadata elapsed=%s", time.Since(start))
			start = time.Now()
			connector, ok := registry.Get(member.connector)
			if !ok {
				t.Fatal("selected connector missing")
			}
			t.Logf("phase=selected_bundle_acquisition elapsed=%s", time.Since(start))
			var decl connectors.CommandSurfaceCommand
			found := 0
			for _, candidate := range connector.(connectors.CommandSurfaceProvider).CommandSurface().Commands {
				if candidate.Path == member.path {
					decl = candidate
					found++
				}
			}
			if found != 1 || decl.Availability != "implemented" {
				t.Fatal("independently selected command identity changed")
			}
			start = time.Now()
			args := productionFixtureArgs178(t, member.connector, decl, map[string]engine.Bundle{})
			carrier, env := productionFixtureCarrier178(decl, args)
			t.Logf("phase=fixture_projection elapsed=%s", time.Since(start))
			if len(env) != 0 {
				t.Fatal("selected timing member unexpectedly needs environment carrier")
			}
			path, err := commandrunner.CommandPathSegments(member.path)
			if err != nil {
				t.Fatal(err)
			}
			start = time.Now()
			if err := commandrunner.PreflightRequest(connector, commandrunner.Request{Path: path, Flags: parseFlags(args).values}); err != nil {
				t.Fatal(err)
			}
			t.Logf("phase=preflight elapsed=%s", time.Since(start))
			root := t.TempDir()
			if err := app.InitProject(root); err != nil {
				t.Fatal(err)
			}
			start = time.Now()
			a, err := app.OpenWithRegistry(root, registry)
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("phase=standalone_app_acquisition elapsed=%s", time.Since(start))
			start = time.Now()
			if err := a.Close(); err != nil {
				t.Fatal(err)
			}
			t.Logf("phase=standalone_app_cleanup elapsed=%s", time.Since(start))
			invocation := append([]string{member.connector}, path...)
			invocation = append(invocation, carrier...)
			invocation = append(invocation, "--root", root, "--json")
			verify := func(raw []byte, code int) {
				var result struct {
					Kind  string `json:"kind"`
					Error struct {
						Message string `json:"message"`
					} `json:"error"`
				}
				if err := json.Unmarshal(raw, &result); err != nil {
					t.Fatal(err)
				}
				if code == 0 || result.Kind != "Error" || result.Error.Message != "missing --credential" {
					t.Fatal("selected member did not reach exact credential boundary")
				}
			}
			var output bytes.Buffer
			start = time.Now()
			code := runWithPreflightRegistry(invocation, &output, io.Discard, registry)
			elapsed := time.Since(start)
			verify(output.Bytes(), code)
			t.Logf("mode=in_process connector=%s command=%q phase=CLI_with_cleanup elapsed=%s", member.connector, member.path, elapsed)
			start = time.Now()
			out, runErr := exec.Command(binary, invocation...).Output()
			elapsed = time.Since(start)
			code = 0
			if runErr != nil {
				code = 1
			}
			verify(out, code)
			t.Logf("mode=fresh_process connector=%s command=%q phase=process_startup_CLI_cleanup elapsed=%s", member.connector, member.path, elapsed)
		})
	}
}
