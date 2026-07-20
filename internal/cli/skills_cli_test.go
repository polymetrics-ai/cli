package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSkillsGeneratePreservesNativeFlagFormsAndOutputs(t *testing.T) {
	t.Run("plain spaced dir", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "skills out")
		stdout, stderr, code := runSkillsCLI(t, []string{"skills", "generate", "--dir", dir})
		if code != 0 || stderr != "" {
			t.Fatalf("exit = %d, stderr = %q, want success", code, stderr)
		}
		if want := "Generated skills in " + dir + "\n"; stdout != want {
			t.Fatalf("stdout = %q, want %q", stdout, want)
		}
		assertGeneratedSkills(t, dir)
	})

	t.Run("JSON assigned dir", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "assigned")
		stdout, stderr, code := runSkillsCLI(t, []string{"skills", "generate", "--dir=" + dir, "--json=true"})
		if code != 0 || stderr != "" {
			t.Fatalf("exit = %d, stderr = %q, want success", code, stderr)
		}
		result := decodeSkillGeneration(t, stdout)
		if result.Kind != "SkillGeneration" || result.Dir != dir || len(result.Skills) == 0 {
			t.Fatalf("result = %+v, want generated skills in %q", result, dir)
		}
		assertGeneratedSkills(t, dir)
	})

	t.Run("repeated dir last wins", func(t *testing.T) {
		root := t.TempDir()
		first := filepath.Join(root, "first")
		last := filepath.Join(root, "last")
		stdout, stderr, code := runSkillsCLI(t, []string{"skills", "generate", "--dir", first, "--dir=" + last, "--json"})
		if code != 0 || stderr != "" {
			t.Fatalf("exit = %d, stderr = %q, want success", code, stderr)
		}
		if got := decodeSkillGeneration(t, stdout).Dir; got != last {
			t.Fatalf("generated dir = %q, want last repeated value %q", got, last)
		}
		if _, err := os.Stat(first); !os.IsNotExist(err) {
			t.Fatalf("first repeated dir stat error = %v, want not exist", err)
		}
		assertGeneratedSkills(t, last)
	})

	t.Run("comma path is not split", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "skills,with,commas")
		stdout, stderr, code := runSkillsCLI(t, []string{"skills", "generate", "--dir=" + dir, "--json"})
		if code != 0 || stderr != "" {
			t.Fatalf("exit = %d, stderr = %q, want success", code, stderr)
		}
		if got := decodeSkillGeneration(t, stdout).Dir; got != dir {
			t.Fatalf("generated dir = %q, want unsplit %q", got, dir)
		}
		assertGeneratedSkills(t, dir)
	})

	t.Run("bare dir keeps legacy true value", func(t *testing.T) {
		root := t.TempDir()
		t.Chdir(root)
		stdout, stderr, code := runSkillsCLI(t, []string{"skills", "generate", "--dir"})
		if code != 0 || stderr != "" {
			t.Fatalf("exit = %d, stderr = %q, want success", code, stderr)
		}
		if stdout != "Generated skills in true\n" {
			t.Fatalf("stdout = %q, want legacy bare-dir output", stdout)
		}
		assertGeneratedSkills(t, filepath.Join(root, "true"))
	})
}

func TestSkillsGeneratePreservesUnknownFlagAndPositionalCompatibility(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "skills")
	stdout, stderr, code := runSkillsCLI(t, []string{
		"skills", "generate", "ignored-positional",
		"--unknown", "ignored-value", "--another=ignored",
		"--dir", dir, "--json",
	})
	if code != 0 || stderr != "" {
		t.Fatalf("exit = %d, stderr = %q, want legacy-compatible success", code, stderr)
	}
	if got := decodeSkillGeneration(t, stdout).Dir; got != dir {
		t.Fatalf("generated dir = %q, want %q", got, dir)
	}
	assertGeneratedSkills(t, dir)
}

func TestSkillsBareTextJSONAndPositionalHelpCompatibility(t *testing.T) {
	canonical, canonicalErr, code := runSkillsCLI(t, []string{"help", "skills"})
	if code != 0 || canonicalErr != "" {
		t.Fatalf("help skills exit = %d, stderr = %q", code, canonicalErr)
	}

	for _, args := range [][]string{
		{"skills"},
		{"skills", "--help"},
		{"skills", "-h"},
		{"skills", "help"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			stdout, stderr, gotCode := runSkillsCLI(t, args)
			if gotCode != 0 || stderr != "" {
				t.Fatalf("Run(%v) exit = %d, stderr = %q", args, gotCode, stderr)
			}
			if stdout != canonical {
				t.Fatalf("Run(%v) help differs from canonical:\nwant:\n%s\ngot:\n%s", args, canonical, stdout)
			}
		})
	}

	for _, args := range [][]string{
		{"skills", "--json"},
		{"skills", "--help", "--json=true"},
		{"skills", "help", "--json"},
		{"--json=true", "skills", "help"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			stdout, stderr, gotCode := runSkillsCLI(t, args)
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
			if manual.APIVersion != apiVersion || manual.Kind != "CommandManual" || manual.Command != "skills" || manual.Manual != canonical {
				t.Fatalf("Run(%v) manual = %+v, want canonical skills manual", args, manual)
			}
		})
	}
}

func TestSkillsInvalidActionsAndErrorsKeepCategories(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		wantCode     int
		wantCategory string
	}{
		{name: "invalid action", args: []string{"skills", "bogus", "--json"}, wantCode: 2, wantCategory: "usage"},
		{name: "namespace unknown flag", args: []string{"skills", "--bogus", "--json"}, wantCode: 2, wantCategory: "usage"},
		{name: "missing dir", args: []string{"skills", "generate", "--json"}, wantCode: 3, wantCategory: "validation"},
		{name: "empty assigned dir", args: []string{"skills", "generate", "--dir=", "--json"}, wantCode: 3, wantCategory: "validation"},
		{name: "malformed assigned boolean", args: []string{"--json", "skills", "generate", "--dir", t.TempDir(), "--json=maybe"}, wantCode: 3, wantCategory: "validation"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, code := runSkillsCLI(t, tt.args)
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

func TestSkillsGlobalAndConfigFlagFormsStayConnected(t *testing.T) {
	t.Run("spaced root loads JSON config with late globals", func(t *testing.T) {
		root := t.TempDir()
		writeCLIConfig(t, root, "json: true\n")
		dir := filepath.Join(t.TempDir(), "skills")
		stdout, stderr, code := runSkillsCLI(t, []string{
			"skills", "generate", "--dir", dir,
			"--plain=false", "--no-input=true", "--root", root,
		})
		if code != 0 || stderr != "" {
			t.Fatalf("exit = %d, stderr = %q, want success", code, stderr)
		}
		if got := decodeSkillGeneration(t, stdout).Dir; got != dir {
			t.Fatalf("generated dir = %q, want %q", got, dir)
		}
	})

	t.Run("assigned root and false JSON override config", func(t *testing.T) {
		root := t.TempDir()
		writeCLIConfig(t, root, "json: true\n")
		dir := filepath.Join(t.TempDir(), "skills")
		stdout, stderr, code := runSkillsCLI(t, []string{
			"--root=" + root, "skills", "generate", "--dir=" + dir, "--json=false",
		})
		if code != 0 || stderr != "" {
			t.Fatalf("exit = %d, stderr = %q, want success", code, stderr)
		}
		if want := "Generated skills in " + dir + "\n"; stdout != want {
			t.Fatalf("stdout = %q, want configured JSON overridden by false: %q", stdout, want)
		}
	})

	t.Run("late assigned true JSON", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "skills")
		stdout, stderr, code := runSkillsCLI(t, []string{
			"skills", "generate", "--dir", dir, "--json=true",
		})
		if code != 0 || stderr != "" {
			t.Fatalf("exit = %d, stderr = %q, want success", code, stderr)
		}
		if got := decodeSkillGeneration(t, stdout).Dir; got != dir {
			t.Fatalf("generated dir = %q, want %q", got, dir)
		}
	})
}

type skillGenerationResult struct {
	APIVersion string   `json:"api_version"`
	Kind       string   `json:"kind"`
	Dir        string   `json:"dir"`
	Skills     []string `json:"skills"`
}

func runSkillsCLI(t *testing.T, args []string) (string, string, int) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	code := Run(args, &stdout, &stderr)
	return stdout.String(), stderr.String(), code
}

func decodeSkillGeneration(t *testing.T, stdout string) skillGenerationResult {
	t.Helper()
	var result skillGenerationResult
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("decode SkillGeneration: %v\n%s", err, stdout)
	}
	if result.APIVersion != apiVersion {
		t.Fatalf("api_version = %q, want %q", result.APIVersion, apiVersion)
	}
	return result
}

func assertGeneratedSkills(t *testing.T, dir string) {
	t.Helper()
	for _, rel := range []string{
		"pm-shared/SKILL.md",
		"pm-github/SKILL.md",
		"pm-etl/SKILL.md",
		"recipe-github-prs-to-warehouse/SKILL.md",
		"skills.md",
	} {
		path := filepath.Join(dir, rel)
		relToRoot, err := filepath.Rel(dir, path)
		if err != nil || !filepath.IsLocal(relToRoot) {
			t.Fatalf("generated path %q escaped root %q: rel=%q err=%v", path, dir, relToRoot, err)
		}
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read generated skill %q: %v", path, err)
		}
		if strings.Contains(strings.ToLower(string(content)), "ghp_") {
			t.Fatalf("generated skill %q appears to contain a token", path)
		}
	}
}
