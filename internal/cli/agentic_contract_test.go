package cli_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"polymetrics.ai/internal/cli"
)

func TestUnknownCommandJSONErrorIsStructuredAndSanitized(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"bad\x1b[31mcmd", "--json"}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("Run(unknown --json) code = %d, want 2; stderr=%q stdout=%q", code, stderr.String(), stdout.String())
	}
	if strings.Contains(stderr.String(), "\x1b") || strings.Contains(stdout.String(), "\x1b") {
		t.Fatalf("error output retained escape sequence; stderr=%q stdout=%q", stderr.String(), stdout.String())
	}
	var got struct {
		APIVersion string `json:"api_version"`
		Kind       string `json:"kind"`
		Error      struct {
			Category string `json:"category"`
			Code     string `json:"code"`
			Message  string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("stdout is not JSON error: %v\n%s", err, stdout.String())
	}
	if got.APIVersion != "polymetrics.ai/v1" || got.Kind != "Error" || got.Error.Category != "usage" || got.Error.Code == "" {
		t.Fatalf("unexpected JSON error: %+v", got)
	}
}

func TestConnectorInspectJSONIncludesManifest(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"connectors", "inspect", "github", "--json"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run(connectors inspect github) code = %d stderr = %s", code, stderr.String())
	}
	out := stdout.String()
	for _, want := range []string{
		`"kind": "Connector"`,
		`"icon"`,
		`"path": "icons/github.svg"`,
		`"manifest"`,
		`"config_fields"`,
		`"secret_fields"`,
		`"token"`,
		`"pagination"`,
		`"risk"`,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("connector inspect output missing %q:\n%s", want, out)
		}
	}
}

func TestConnectorInspectRejectsUnsafeIdentifier(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"connectors", "inspect", "bad\x1b[31m", "--json"}, &stdout, &stderr)
	if code != 3 {
		t.Fatalf("Run(connectors inspect unsafe) code = %d, want 3; stderr=%q stdout=%q", code, stderr.String(), stdout.String())
	}
	var got struct {
		Error struct {
			Category string `json:"category"`
		} `json:"error"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("stdout is not JSON error: %v\n%s", err, stdout.String())
	}
	if got.Error.Category != "validation" {
		t.Fatalf("error category = %q, want validation; stdout=%s", got.Error.Category, stdout.String())
	}
	if strings.Contains(stdout.String(), "\x1b") || strings.Contains(stderr.String(), "\x1b") {
		t.Fatalf("unsafe output retained escape sequence; stdout=%q stderr=%q", stdout.String(), stderr.String())
	}
}

func TestSkillsGenerateWritesAgentSkills(t *testing.T) {
	dir := t.TempDir()
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"skills", "generate", "--dir", dir, "--json"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run(skills generate) code = %d stderr = %s stdout = %s", code, stderr.String(), stdout.String())
	}
	for _, rel := range []string{
		"pm-shared/SKILL.md",
		"pm-github/SKILL.md",
		"pm-etl/SKILL.md",
		"recipe-github-prs-to-warehouse/SKILL.md",
		"skills.md",
	} {
		path := filepath.Join(dir, rel)
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected generated skill %s: %v", path, err)
		}
	}
	content, err := os.ReadFile(filepath.Join(dir, "pm-github", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(content)
	if !strings.Contains(text, "github") || !strings.Contains(text, "pull_requests") {
		t.Fatalf("github skill missing connector stream context:\n%s", text)
	}
	etlContent, err := os.ReadFile(filepath.Join(dir, "pm-etl", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(etlContent), "incremental_append_deduped") {
		t.Fatalf("etl skill missing sync mode guidance:\n%s", string(etlContent))
	}
	if strings.Contains(strings.ToLower(text), "ghp_") {
		t.Fatalf("generated skill appears to contain a GitHub token:\n%s", text)
	}
}
