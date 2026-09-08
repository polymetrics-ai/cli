package cli

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"os/exec"
	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/bundleregistry"
	"polymetrics.ai/internal/connectors/commandrunner"
	"polymetrics.ai/internal/connectors/engine"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"
)

// Every production command is exercised in a fresh process without a credential.
// Exact names and observed outcomes remain in the original Go event stream.
func TestProductionRegistryBuiltPaths173(t *testing.T) {
	registry := bundleregistry.New()
	bundles := map[string]engine.Bundle{}
	binary := os.Getenv("PM_CP16_SWEEP_BINARY")
	expectedBinarySHA := os.Getenv("PM_CP16_SWEEP_BINARY_SHA256")
	if binary == "" {
		binary = buildTransportPM(t)
	} else if expectedBinarySHA == "" {
		t.Fatal("external sweep binary requires exact hash")
	}

	binaryBytes, err := os.ReadFile(binary)
	if err != nil {
		t.Fatal(err)
	}
	binaryDigest := sha256.Sum256(binaryBytes)
	binarySHA := hex.EncodeToString(binaryDigest[:])
	if expectedBinarySHA != "" && binarySHA != expectedBinarySHA {
		t.Fatal("sweep binary hash differs")
	}
	t.Cleanup(func() {
		after, err := os.ReadFile(binary)
		if err != nil || sha256.Sum256(after) != binaryDigest {
			t.Error("sweep binary changed during execution")
		}
	})
	t.Logf("fresh_binary_sha256=%s bytes=%d", binarySHA, len(binaryBytes))
	root := t.TempDir()
	if out, err := runTransportPM(binary, "", "init", "--root", root, "--json"); err != nil {
		t.Fatalf("init: %v %s", err, out)
	}
	checked := 0
	var preparedMembers, selectedMembers []string
	for _, meta := range registry.List() {
		connector, ok := registry.Get(meta.Name)
		if !ok {
			t.Fatalf("listed connector missing: %s", meta.Name)
		}
		provider, ok := connector.(connectors.CommandSurfaceProvider)
		if !ok || provider.CommandSurface() == nil {
			continue
		}
		checked += len(provider.CommandSurface().Commands)
		t.Run(meta.Name, func(t *testing.T) {
			for _, decl := range provider.CommandSurface().Commands {
				t.Run(decl.Path, func(t *testing.T) {
					selectedMembers = append(selectedMembers, meta.Name+"/"+decl.Path)
					preparedMembers = append(preparedMembers, meta.Name+"/"+decl.Path)
					fixtureArgs := productionFixtureArgs178(t, meta.Name, decl, bundles)
					fixtureArgs, fixtureEnv := productionFixtureCarrier178(decl, fixtureArgs)
					t.Parallel()
					path, err := commandrunner.CommandPathSegments(decl.Path)
					if err != nil {
						t.Fatalf("declared path rejected: %v", err)
					}
					if decl.Availability == "implemented" {
						if err := commandrunner.Preflight(connector, path); err != nil {
							t.Fatalf("implemented binding: %v", err)
						}
					}
					args := append([]string{meta.Name}, path...)
					args = append(args, fixtureArgs...)
					args = append(args, "--root", root, "--json")
					ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
					defer cancel()
					command := exec.CommandContext(ctx, binary, args...)
					command.Env = []string{"HOME=" + root, "TMPDIR=" + os.TempDir(), "LANG=C", "LC_ALL=C"}
					for key, value := range fixtureEnv {
						command.Env = append(command.Env, key+"="+value)
					}
					envKeys := []string{"HOME", "TMPDIR", "LANG", "LC_ALL"}
					for key := range fixtureEnv {
						envKeys = append(envKeys, key)
					}
					sort.Strings(envKeys)
					loggedArgs := append([]string{}, args...)
					for i := range loggedArgs {
						if loggedArgs[i] == root {
							loggedArgs[i] = "<isolated-root>"
						}
					}
					argvJSON, _ := json.Marshal(loggedArgs)
					t.Logf("binary_sha256=%s argv=%s environment_keys=%q", binarySHA, argvJSON, envKeys)
					var stdout, stderr bytes.Buffer
					command.Stdout, command.Stderr = &stdout, &stderr
					runErr := command.Run()
					var result struct {
						Kind  string `json:"kind"`
						Error struct {
							Code    string `json:"code"`
							Message string `json:"message"`
						} `json:"error"`
					}
					if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
						t.Fatalf("error envelope: %v stdout=%s stderr=%s", err, stdout.String(), stderr.String())
					}
					t.Logf("connector=%s path=%q availability=%s code=%s message=%q", meta.Name, decl.Path, decl.Availability, result.Error.Code, result.Error.Message)
					if runErr == nil || ctx.Err() != nil || result.Kind != "Error" {
						t.Fatalf("no bounded refusal: err=%v ctx=%v stdout=%s", runErr, ctx.Err(), stdout.String())
					}
					if strings.TrimSpace(stderr.String()) != "error: "+result.Error.Message {
						t.Fatalf("public error parity: %s", stderr.String())
					}
					if decl.Availability == "implemented" {
						if result.Error.Message != "missing --credential" {
							t.Fatalf("did not reach credential boundary: %s", stdout.String())
						}
					} else if result.Error.Code != "connector_command_blocked" || !strings.Contains(result.Error.Message, "availability="+decl.Availability) {
						t.Fatalf("declared block differs: %s", stdout.String())
					}
				})
			}
		})
	}
	if len(selectedMembers) == 0 {
		t.Fatal("no fresh-process members selected")
	}
	if !reflect.DeepEqual(preparedMembers, selectedMembers) {
		t.Fatalf("fixture preparation does not match exact selected members: prepared=%d selected=%d", len(preparedMembers), len(selectedMembers))
	}
	t.Logf("fixture_preparation_members=%d selected_members=%d", len(preparedMembers), len(selectedMembers))
	if checked != 13856 {
		t.Fatalf("production command membership=%d want13856", checked)
	}
	t.Logf("complete declared membership=%d; executed selection is recorded by test events", checked)
}
