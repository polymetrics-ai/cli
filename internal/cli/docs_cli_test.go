package cli

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDocsGeneratePreservesNativeFlagFormsAndArtifactBytes(t *testing.T) {
	root := t.TempDir()
	firstCLI := filepath.Join(root, "first-cli")
	cliDir := filepath.Join(root, "cli output,spaced")
	firstConnectors := filepath.Join(root, "first-connectors")
	connectorsDir := filepath.Join(root, "connector output,spaced")

	stdout, stderr, code := runDocsCLI(t, []string{
		"docs", "generate", "ignored-positional",
		"--unknown", "ignored-value",
		"--dir=" + firstCLI, "--dir", cliDir,
		"--connectors-dir=" + firstConnectors, "--connectors-dir", connectorsDir,
		"--json=true",
	})
	if code != 0 || stderr != "" {
		t.Fatalf("exit = %d, stderr = %q, want success", code, stderr)
	}
	if want := "Generated docs in " + cliDir + " and connector docs in " + connectorsDir + "\n"; stdout != want {
		t.Fatalf("stdout = %q, want %q", stdout, want)
	}
	for _, unused := range []string{firstCLI, firstConnectors} {
		if _, err := os.Stat(unused); !os.IsNotExist(err) {
			t.Fatalf("first repeated output %q stat error = %v, want not exist", unused, err)
		}
	}
	assertGeneratedCLIManualBytes(t, cliDir)
	assertGeneratedConnectorDocsLocal(t, connectorsDir)
}

func TestDocsGenerateAndValidatePreserveBareAndFallbackDirectoryForms(t *testing.T) {
	root := chdirDocsTempRoot(t)

	t.Run("bare command dir", func(t *testing.T) {
		connectorsDir := filepath.Join(root, "bare-dir-connectors")
		stdout, stderr, code := runDocsCLI(t, []string{
			"docs", "generate", "--dir", "--connectors-dir=" + connectorsDir,
		})
		if code != 0 || stderr != "" {
			t.Fatalf("exit = %d, stderr = %q, want success", code, stderr)
		}
		if want := "Generated docs in true and connector docs in " + connectorsDir + "\n"; stdout != want {
			t.Fatalf("stdout = %q, want %q", stdout, want)
		}
		assertGeneratedCLIManualBytes(t, filepath.Join(root, "true"))
		assertGeneratedConnectorDocsLocal(t, connectorsDir)
	})

	t.Run("default connector sibling and validation dir", func(t *testing.T) {
		stdout, stderr, code := runDocsCLI(t, []string{"docs", "generate", "--dir", filepath.Join("docs", "cli")})
		if code != 0 || stderr != "" {
			t.Fatalf("generate exit = %d, stderr = %q, want success", code, stderr)
		}
		if want := "Generated docs in docs/cli and connector docs in docs/connectors\n"; stdout != want {
			t.Fatalf("stdout = %q, want %q", stdout, want)
		}
		assertGeneratedCLIManualBytes(t, filepath.Join(root, "docs", "cli"))
		assertGeneratedConnectorDocsLocal(t, filepath.Join(root, "docs", "connectors"))

		stdout, stderr, code = runDocsCLI(t, []string{"docs", "validate"})
		if code != 0 || stderr != "" || stdout != "Validated connector docs in docs/connectors\n" {
			t.Fatalf("default validate exit = %d, stdout = %q, stderr = %q", code, stdout, stderr)
		}
	})

	t.Run("bare connector dir and validate fallbacks", func(t *testing.T) {
		cliDir := filepath.Join(root, "bare-connectors-cli")
		stdout, stderr, code := runDocsCLI(t, []string{
			"docs", "generate", "--dir=" + cliDir, "--connectors-dir",
		})
		if code != 0 || stderr != "" {
			t.Fatalf("generate exit = %d, stderr = %q, want success", code, stderr)
		}
		if want := "Generated docs in " + cliDir + " and connector docs in true\n"; stdout != want {
			t.Fatalf("stdout = %q, want %q", stdout, want)
		}
		connectorsDir := filepath.Join(root, "true")
		assertGeneratedConnectorDocsLocal(t, connectorsDir)

		for _, args := range [][]string{
			{"docs", "validate", "ignored-positional", "--unknown", "ignored-value", "--connectors-dir", connectorsDir},
			{"docs", "validate", "--dir", filepath.Join(root, "unused"), "--dir=" + connectorsDir},
			{"docs", "validate", "--connectors-dir"},
			{"docs", "validate", "--dir"},
		} {
			t.Run(strings.Join(args, " "), func(t *testing.T) {
				gotOut, gotErr, gotCode := runDocsCLI(t, args)
				if gotCode != 0 || gotErr != "" {
					t.Fatalf("Run(%v) exit = %d, stderr = %q, want success", args, gotCode, gotErr)
				}
				wantDir := connectorsDir
				if args[len(args)-1] == "--connectors-dir" || args[len(args)-1] == "--dir" {
					wantDir = "true"
				}
				want := "Validated connector docs in " + wantDir + "\n"
				if gotOut != want {
					t.Fatalf("Run(%v) stdout = %q, want %q", args, gotOut, want)
				}
			})
		}
	})
}

func TestDocsBareTextJSONAndPositionalHelpCompatibility(t *testing.T) {
	canonical, canonicalErr, code := runDocsCLI(t, []string{"help", "docs"})
	if code != 0 || canonicalErr != "" {
		t.Fatalf("help docs exit = %d, stderr = %q", code, canonicalErr)
	}

	for _, args := range [][]string{
		{"docs"},
		{"docs", "--help"},
		{"docs", "-h"},
		{"docs", "help"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			stdout, stderr, gotCode := runDocsCLI(t, args)
			if gotCode != 0 || stderr != "" {
				t.Fatalf("Run(%v) exit = %d, stderr = %q", args, gotCode, stderr)
			}
			if stdout != canonical {
				t.Fatalf("Run(%v) help differs from canonical:\nwant:\n%s\ngot:\n%s", args, canonical, stdout)
			}
		})
	}

	for _, args := range [][]string{
		{"docs", "--json"},
		{"docs", "--help", "--json=true"},
		{"docs", "help", "--json"},
		{"--json=true", "docs", "help"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			stdout, stderr, gotCode := runDocsCLI(t, args)
			if gotCode != 0 || stderr != "" {
				t.Fatalf("Run(%v) exit = %d, stderr = %q", args, gotCode, stderr)
			}
			var manual struct {
				APIVersion string `json:"api_version"`
				Kind       string `json:"kind"`
				Command    string `json:"command"`
				Manual     string `json:"manual"`
			}
			if err := json.Unmarshal([]byte(stdout), &manual); err != nil {
				t.Fatalf("Run(%v) stdout not JSON: %v\n%s", args, err, stdout)
			}
			if manual.APIVersion != apiVersion || manual.Kind != "CommandManual" || manual.Command != "docs" || manual.Manual != canonical {
				t.Fatalf("Run(%v) manual = %+v, want canonical docs manual", args, manual)
			}
		})
	}
}

func TestDocsInvalidActionsAndErrorsKeepCategories(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		wantCode     int
		wantCategory string
	}{
		{name: "invalid action", args: []string{"docs", "bogus", "--json"}, wantCode: 2, wantCategory: "usage"},
		{name: "namespace unknown flag", args: []string{"docs", "--bogus", "--json"}, wantCode: 2, wantCategory: "usage"},
		{name: "missing generate dir", args: []string{"docs", "generate", "--json"}, wantCode: 1, wantCategory: "internal"},
		{name: "empty assigned generate dir", args: []string{"docs", "generate", "--dir=", "--json"}, wantCode: 1, wantCategory: "internal"},
		{name: "missing validation tree", args: []string{"docs", "validate", "--connectors-dir=" + filepath.Join(t.TempDir(), "missing"), "--json"}, wantCode: 1, wantCategory: "internal"},
		{name: "malformed assigned boolean", args: []string{"--json", "docs", "validate", "--json=maybe"}, wantCode: 3, wantCategory: "validation"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, code := runDocsCLI(t, tt.args)
			if code != tt.wantCode {
				t.Fatalf("Run(%v) exit = %d, want %d; stdout=%s stderr=%s", tt.args, code, tt.wantCode, stdout, stderr)
			}
			var result struct {
				Kind  string `json:"kind"`
				Error struct {
					Category string `json:"category"`
				} `json:"error"`
			}
			if err := json.Unmarshal([]byte(stdout), &result); err != nil {
				t.Fatalf("Run(%v) stdout not JSON: %v\n%s", tt.args, err, stdout)
			}
			if result.Kind != "Error" || result.Error.Category != tt.wantCategory {
				t.Fatalf("Run(%v) error = %+v, want category %q", tt.args, result, tt.wantCategory)
			}
			if strings.Contains(stdout, `"kind": "CommandManual"`) || stderr == "" {
				t.Fatalf("Run(%v) masked error with help or omitted diagnostic: stdout=%q stderr=%q", tt.args, stdout, stderr)
			}
		})
	}
}

func TestDocsGlobalAndAssignedBooleanFormsStayConnected(t *testing.T) {
	t.Run("configured JSON help with spaced root and late assigned booleans", func(t *testing.T) {
		root := t.TempDir()
		writeCLIConfig(t, root, "json: true\n")
		stdout, stderr, code := runDocsCLI(t, []string{
			"docs", "help", "--plain=false", "--no-input=true", "--root", root,
		})
		if code != 0 || stderr != "" {
			t.Fatalf("exit = %d, stderr = %q, want success", code, stderr)
		}
		var manual struct {
			Kind    string `json:"kind"`
			Command string `json:"command"`
		}
		if err := json.Unmarshal([]byte(stdout), &manual); err != nil {
			t.Fatalf("configured help stdout not JSON: %v\n%s", err, stdout)
		}
		if manual.Kind != "CommandManual" || manual.Command != "docs" {
			t.Fatalf("configured help = %+v, want docs manual", manual)
		}
	})

	t.Run("assigned root and false JSON override config", func(t *testing.T) {
		root := t.TempDir()
		writeCLIConfig(t, root, "json: true\n")
		stdout, stderr, code := runDocsCLI(t, []string{
			"--root=" + root, "docs", "bogus", "--json=false", "--plain=true", "--no-input=false",
		})
		if code != 2 {
			t.Fatalf("exit = %d, want usage 2; stdout=%q stderr=%q", code, stdout, stderr)
		}
		if stdout != "" || stderr == "" {
			t.Fatalf("assigned false JSON did not preserve plain error: stdout=%q stderr=%q", stdout, stderr)
		}
	})
}

func chdirDocsTempRoot(t *testing.T) string {
	t.Helper()
	repoRoot, err := repoRootFromWorkingDir()
	if err != nil {
		t.Fatalf("locate repository root: %v", err)
	}
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module docs-test\n"), 0o644); err != nil {
		t.Fatalf("write temporary go.mod: %v", err)
	}
	connectorsDir := filepath.Join(root, "docs", "connectors")
	if err := os.MkdirAll(connectorsDir, 0o755); err != nil {
		t.Fatalf("create temporary connector docs root: %v", err)
	}
	if err := os.Symlink(filepath.Join(repoRoot, "docs", "connectors", "icons"), filepath.Join(connectorsDir, "icons")); err != nil {
		t.Fatalf("link read-only connector icon source: %v", err)
	}
	t.Chdir(root)
	return root
}

func runDocsCLI(t *testing.T, args []string) (string, string, int) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	code := Run(args, &stdout, &stderr)
	return stdout.String(), stderr.String(), code
}

func assertGeneratedCLIManualBytes(t *testing.T, dir string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read generated CLI dir %q: %v", dir, err)
	}
	wantFiles := len(docs) - 2 // root aliases "" and "pm" are intentionally not generated.
	if len(entries) != wantFiles {
		t.Fatalf("generated CLI file count = %d, want %d", len(entries), wantFiles)
	}
	for topic, text := range docs {
		if topic == "" || topic == "pm" {
			continue
		}
		path := filepath.Join(dir, topic+".md")
		assertPathWithinRoot(t, dir, path)
		got, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read generated manual %q: %v", path, err)
		}
		want := "```\n" + text + "\n```\n"
		if string(got) != want {
			t.Fatalf("generated manual %s byte drift", topic)
		}
	}
}

func assertGeneratedConnectorDocsLocal(t *testing.T, dir string) {
	t.Helper()
	for _, rel := range []string{
		"README.md",
		filepath.Join("catalog", "all-connectors.json"),
		filepath.Join("catalog", "all-connectors.md"),
		filepath.Join("github", "MANUAL.md"),
		filepath.Join("github", "SKILL.md"),
		filepath.Join("icons", "github.svg"),
	} {
		path := filepath.Join(dir, rel)
		assertPathWithinRoot(t, dir, path)
		if info, err := os.Stat(path); err != nil || info.IsDir() {
			t.Fatalf("generated connector artifact %q: info=%v err=%v", path, info, err)
		}
	}
	if err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		assertPathWithinRoot(t, dir, path)
		return nil
	}); err != nil {
		t.Fatalf("walk generated connector docs %q: %v", dir, err)
	}
}

func assertPathWithinRoot(t *testing.T, root, path string) {
	t.Helper()
	rel, err := filepath.Rel(root, path)
	if err != nil || !filepath.IsLocal(rel) {
		t.Fatalf("generated path %q escaped temp output root %q: rel=%q err=%v", path, root, rel, err)
	}
}
