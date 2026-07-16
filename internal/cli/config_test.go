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
