package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigMalformedFileExitsValidation(t *testing.T) {
	root := t.TempDir()
	configDir := filepath.Join(root, ".polymetrics")
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		t.Fatalf("create config dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte("runtime:\n  postgres_url: [unterminated\n"), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	var stdout, stderr bytes.Buffer
	code := Run([]string{"--root", root, "--json", "version"}, &stdout, &stderr)
	if code != 3 {
		t.Fatalf("exit code = %d, want 3\nstdout=%s\nstderr=%s", code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stderr.String(), "error:") || !strings.Contains(stderr.String(), "config") {
		t.Fatalf("stderr = %q, want config validation error", stderr.String())
	}

	var env struct {
		Kind  string `json:"kind"`
		Error struct {
			Category string `json:"category"`
			Code     string `json:"code"`
			Message  string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal stdout error envelope: %v\nstdout=%s", err, stdout.String())
	}
	if env.Kind != "Error" || env.Error.Category != "validation" || env.Error.Code != "validation_error" {
		t.Fatalf("error envelope = %#v, want validation error", env)
	}
	if strings.Contains(env.Error.Message, "\x1b") || strings.Contains(stderr.String(), "\x1b") {
		t.Fatalf("config error retained ANSI: stdout=%q stderr=%q", stdout.String(), stderr.String())
	}
}

func TestConfigRootEnvControlsDiscoveryAndInvocationRoot(t *testing.T) {
	defaultRoot := t.TempDir()
	t.Chdir(defaultRoot)
	effectiveRoot := t.TempDir()
	t.Setenv("POLYMETRICS_ROOT", effectiveRoot)

	var stdout, stderr bytes.Buffer
	code := Run([]string{"init", "--json"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\nstdout=%s\nstderr=%s", code, stdout.String(), stderr.String())
	}
	if _, err := os.Stat(filepath.Join(effectiveRoot, ".polymetrics", "config.yaml")); err != nil {
		t.Fatalf("env root was not initialized: %v", err)
	}
	if !strings.Contains(stdout.String(), filepath.Join(effectiveRoot, ".polymetrics")) {
		t.Fatalf("stdout = %s, want env root project_dir", stdout.String())
	}
}

func TestConfigPMRootAliasControlsDiscoveryAndInvocationRoot(t *testing.T) {
	defaultRoot := t.TempDir()
	t.Chdir(defaultRoot)
	effectiveRoot := t.TempDir()
	t.Setenv("PM_ROOT", effectiveRoot)

	var stdout, stderr bytes.Buffer
	code := Run([]string{"init", "--json"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\nstdout=%s\nstderr=%s", code, stdout.String(), stderr.String())
	}
	if _, err := os.Stat(filepath.Join(effectiveRoot, ".polymetrics", "config.yaml")); err != nil {
		t.Fatalf("PM_ROOT alias root was not initialized: %v", err)
	}
}

func TestConfigRootFlagOverridesEnvDiscovery(t *testing.T) {
	envRoot := t.TempDir()
	writeCLIConfig(t, envRoot, "runtime:\n  postgres_url: [unterminated\n")
	flagRoot := t.TempDir()
	t.Setenv("POLYMETRICS_ROOT", envRoot)

	var stdout, stderr bytes.Buffer
	code := Run([]string{"--root", flagRoot, "--json", "version"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0 when --root overrides env\nstdout=%s\nstderr=%s", code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "\"kind\": \"Version\"") {
		t.Fatalf("stdout = %s, want JSON version envelope", stdout.String())
	}
}

func TestConfigJSONEnvRendersMalformedConfigAsJSON(t *testing.T) {
	root := t.TempDir()
	writeCLIConfig(t, root, "runtime:\n  postgres_url: [unterminated\n")
	t.Setenv("POLYMETRICS_ROOT", root)
	t.Setenv("POLYMETRICS_JSON", "true")

	var stdout, stderr bytes.Buffer
	code := Run([]string{"version"}, &stdout, &stderr)
	if code != 3 {
		t.Fatalf("exit code = %d, want 3\nstdout=%s\nstderr=%s", code, stdout.String(), stderr.String())
	}
	assertValidationJSONEnvelope(t, stdout.Bytes())
	if !strings.Contains(stderr.String(), "error:") || !strings.Contains(stderr.String(), "config") {
		t.Fatalf("stderr = %q, want config validation error", stderr.String())
	}
}

func TestConfigPMJSONAliasRendersMalformedConfigAsJSON(t *testing.T) {
	root := t.TempDir()
	writeCLIConfig(t, root, "runtime:\n  postgres_url: [unterminated\n")
	t.Setenv("PM_ROOT", root)
	t.Setenv("PM_JSON", "true")

	var stdout, stderr bytes.Buffer
	code := Run([]string{"version"}, &stdout, &stderr)
	if code != 3 {
		t.Fatalf("exit code = %d, want 3\nstdout=%s\nstderr=%s", code, stdout.String(), stderr.String())
	}
	assertValidationJSONEnvelope(t, stdout.Bytes())
}

func TestConfigJSONFlagOverridesEnvForMalformedConfig(t *testing.T) {
	root := t.TempDir()
	writeCLIConfig(t, root, "runtime:\n  postgres_url: [unterminated\n")
	t.Setenv("POLYMETRICS_JSON", "false")

	var stdout, stderr bytes.Buffer
	code := Run([]string{"--root", root, "--json", "version"}, &stdout, &stderr)
	if code != 3 {
		t.Fatalf("exit code = %d, want 3\nstdout=%s\nstderr=%s", code, stdout.String(), stderr.String())
	}
	assertValidationJSONEnvelope(t, stdout.Bytes())
}

func TestConfigFileJSONControlsInvocationOutput(t *testing.T) {
	root := t.TempDir()
	writeCLIConfig(t, root, "json: true\n")

	var stdout, stderr bytes.Buffer
	code := Run([]string{"--root", root, "version"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\nstdout=%s\nstderr=%s", code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "\"kind\": \"Version\"") {
		t.Fatalf("stdout = %s, want JSON version envelope from config file json", stdout.String())
	}
}

func TestConfigFileRootControlsInvocationWithoutRelocatingDiscovery(t *testing.T) {
	discoveryRoot := t.TempDir()
	effectiveRoot := t.TempDir()
	writeCLIConfig(t, effectiveRoot, "runtime:\n  postgres_url: [unterminated\n")
	writeCLIConfig(t, discoveryRoot, "root: "+effectiveRoot+"\njson: true\n")
	t.Chdir(discoveryRoot)

	var stdout, stderr bytes.Buffer
	code := Run([]string{"init"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0; file root must not relocate discovery for same load\nstdout=%s\nstderr=%s", code, stdout.String(), stderr.String())
	}
	if _, err := os.Stat(filepath.Join(effectiveRoot, ".polymetrics", "config.yaml")); err != nil {
		t.Fatalf("file root was not initialized: %v", err)
	}
	if !strings.Contains(stdout.String(), "\"kind\": \"InitResult\"") || !strings.Contains(stdout.String(), filepath.Join(effectiveRoot, ".polymetrics")) {
		t.Fatalf("stdout = %s, want JSON init result for file root", stdout.String())
	}
}

func TestConfigInvocationIsolation(t *testing.T) {
	jsonRoot := t.TempDir()
	plainRoot := t.TempDir()
	writeCLIConfig(t, jsonRoot, "json: true\n")

	var jsonStdout, jsonStderr bytes.Buffer
	if code := Run([]string{"--root", jsonRoot, "version"}, &jsonStdout, &jsonStderr); code != 0 {
		t.Fatalf("json root exit = %d, stderr=%s", code, jsonStderr.String())
	}
	if !strings.Contains(jsonStdout.String(), "\"kind\": \"Version\"") {
		t.Fatalf("json root stdout = %s, want JSON version envelope", jsonStdout.String())
	}

	var plainStdout, plainStderr bytes.Buffer
	if code := Run([]string{"--root", plainRoot, "version"}, &plainStdout, &plainStderr); code != 0 {
		t.Fatalf("plain root exit = %d, stderr=%s", code, plainStderr.String())
	}
	if strings.Contains(plainStdout.String(), "\"kind\": \"Version\"") {
		t.Fatalf("plain root stdout = %s, config JSON leaked across invocations", plainStdout.String())
	}
}

func TestConfigMissingFileKeepsRootManualSuccess(t *testing.T) {
	root := t.TempDir()

	var stdout, stderr bytes.Buffer
	code := Run([]string{"--root", root, "--json"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\nstdout=%s\nstderr=%s", code, stdout.String(), stderr.String())
	}
	if stderr.String() != "" {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	if !strings.Contains(stdout.String(), "\"kind\": \"CommandManual\"") {
		t.Fatalf("stdout = %s, want root CommandManual", stdout.String())
	}
}

func writeCLIConfig(t *testing.T, root string, yaml string) {
	t.Helper()
	dir := filepath.Join(root, ".polymetrics")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("create config dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(yaml), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
}

func assertValidationJSONEnvelope(t *testing.T, data []byte) {
	t.Helper()
	var env struct {
		Kind  string `json:"kind"`
		Error struct {
			Category string `json:"category"`
			Code     string `json:"code"`
			Message  string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(data, &env); err != nil {
		t.Fatalf("unmarshal stdout error envelope: %v\nstdout=%s", err, string(data))
	}
	if env.Kind != "Error" || env.Error.Category != "validation" || env.Error.Code != "validation_error" {
		t.Fatalf("error envelope = %#v, want validation error", env)
	}
}
