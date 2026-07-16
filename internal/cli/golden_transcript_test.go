package cli_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"polymetrics.ai/internal/cli"
)

type goldenTranscript struct {
	Name     string   `json:"name"`
	Args     []string `json:"args"`
	ExitCode int      `json:"exit_code"`
	Stdout   string   `json:"stdout"`
	Stderr   string   `json:"stderr"`
}

var goldenTranscriptInputs = []struct {
	Name string
	Args []string
}{
	{Name: "root_bare_manual", Args: []string{}},
	{Name: "root_long_help", Args: []string{"--help"}},
	{Name: "root_short_help", Args: []string{"-h"}},
	{Name: "root_help_command", Args: []string{"help"}},
	{Name: "root_man_command", Args: []string{"man"}},
	{Name: "root_json_help", Args: []string{"--json", "--help"}},
	{Name: "root_late_json_help", Args: []string{"--help", "--json"}},
	{Name: "root_equals_form", Args: []string{"--root=.", "--help"}},
	{Name: "root_space_form", Args: []string{"--root", ".", "--help"}},

	{Name: "help_credentials", Args: []string{"help", "credentials"}},
	{Name: "help_etl", Args: []string{"help", "etl"}},
	{Name: "help_reverse", Args: []string{"help", "reverse"}},
	{Name: "help_connectors", Args: []string{"help", "connectors"}},
	{Name: "help_connections", Args: []string{"help", "connections"}},
	{Name: "help_catalog", Args: []string{"help", "catalog"}},
	{Name: "help_query", Args: []string{"help", "query"}},
	{Name: "help_flow", Args: []string{"help", "flow"}},
	{Name: "help_rlm", Args: []string{"help", "rlm"}},
	{Name: "help_schedule", Args: []string{"help", "schedule"}},
	{Name: "help_agent", Args: []string{"help", "agent"}},
	{Name: "help_runtime", Args: []string{"help", "runtime"}},
	{Name: "help_perf", Args: []string{"help", "perf"}},
	{Name: "help_docs", Args: []string{"help", "docs"}},
	{Name: "help_skills", Args: []string{"help", "skills"}},
	{Name: "help_version", Args: []string{"help", "version"}},

	{Name: "bare_credentials_manual", Args: []string{"credentials"}},
	{Name: "bare_etl_manual", Args: []string{"etl"}},
	{Name: "bare_reverse_manual", Args: []string{"reverse"}},
	{Name: "bare_connectors_manual", Args: []string{"connectors"}},
	{Name: "bare_connections_manual", Args: []string{"connections"}},
	{Name: "bare_catalog_manual", Args: []string{"catalog"}},
	{Name: "bare_query_manual", Args: []string{"query"}},
	{Name: "bare_flow_manual", Args: []string{"flow"}},
	{Name: "bare_rlm_manual", Args: []string{"rlm"}},
	{Name: "bare_schedule_manual", Args: []string{"schedule"}},
	{Name: "bare_agent_manual", Args: []string{"agent"}},
	{Name: "bare_runtime_manual", Args: []string{"runtime"}},
	{Name: "bare_perf_manual", Args: []string{"perf"}},
	{Name: "bare_docs_manual", Args: []string{"docs"}},
	{Name: "bare_skills_manual", Args: []string{"skills"}},

	{Name: "json_credentials_manual", Args: []string{"credentials", "--json"}},
	{Name: "json_etl_manual", Args: []string{"etl", "--json"}},
	{Name: "json_reverse_manual", Args: []string{"reverse", "--json"}},
	{Name: "json_connectors_manual", Args: []string{"connectors", "--json"}},
	{Name: "json_connections_manual", Args: []string{"connections", "--json"}},
	{Name: "json_catalog_manual", Args: []string{"catalog", "--json"}},
	{Name: "json_query_manual", Args: []string{"query", "--json"}},
	{Name: "json_flow_manual", Args: []string{"flow", "--json"}},
	{Name: "json_rlm_manual", Args: []string{"rlm", "--json"}},
	{Name: "json_schedule_manual", Args: []string{"schedule", "--json"}},
	{Name: "json_agent_manual", Args: []string{"agent", "--json"}},
	{Name: "json_runtime_manual", Args: []string{"runtime", "--json"}},
	{Name: "json_perf_manual", Args: []string{"perf", "--json"}},
	{Name: "json_docs_manual", Args: []string{"docs", "--json"}},
	{Name: "json_skills_manual", Args: []string{"skills", "--json"}},
	{Name: "json_version_manual", Args: []string{"version", "--help", "--json"}},

	{Name: "agent_plan_default", Args: []string{"agent", "plan"}},
	{Name: "agent_plan_json", Args: []string{"agent", "plan", "--json"}},
	{Name: "agent_plan_request_space", Args: []string{"agent", "plan", "--request", "sample customers"}},
	{Name: "agent_plan_request_equals_json", Args: []string{"agent", "plan", "--request=sample customers", "--json"}},
	{Name: "agent_plan_repeated_last_wins_json", Args: []string{"agent", "plan", "--request", "ignore", "--request", "sample customers", "--json"}},
	{Name: "agent_plan_bare_request_json", Args: []string{"agent", "plan", "--request", "--json"}},
	{Name: "agent_plan_unknown_flag_tolerated_json", Args: []string{"agent", "plan", "--unknown", "value", "--json"}},
	{Name: "agent_plan_late_global_json", Args: []string{"agent", "plan", "--request", "sample customers", "--json"}},

	{Name: "connectors_inspect_github_json", Args: []string{"connectors", "inspect", "github", "--json"}},
	{Name: "connectors_inspect_sample", Args: []string{"connectors", "inspect", "sample"}},
	{Name: "connectors_inspect_sample_json", Args: []string{"connectors", "inspect", "sample", "--json"}},
	{Name: "connectors_catalog_invalid_capability_json", Args: []string{"connectors", "catalog", "--capability", "nope", "--json"}},
	{Name: "connectors_catalog_legacy_type_json", Args: []string{"connectors", "catalog", "--type", "source", "--json"}},
	{Name: "connectors_inspect_unknown_json", Args: []string{"connectors", "inspect", "nosuch", "--json"}},
	{Name: "connectors_inspect_unsafe_json", Args: []string{"connectors", "inspect", "bad\x1b[31m", "--json"}},
	{Name: "connectors_help_github_intercept", Args: []string{"connectors", "help", "github"}},

	{Name: "version_plain", Args: []string{"version"}},
	{Name: "version_json", Args: []string{"version", "--json"}},
	{Name: "version_help_json", Args: []string{"version", "--help", "--json"}},
	{Name: "docs_unknown_subcommand", Args: []string{"docs", "bogus"}},
	{Name: "docs_generate_missing_dir", Args: []string{"docs", "generate"}},
	{Name: "skills_missing_subcommand", Args: []string{"skills"}},
	{Name: "worker_missing_subcommand", Args: []string{"worker"}},
	{Name: "worker_unknown_subcommand_json", Args: []string{"worker", "bogus", "--json"}},
	{Name: "worker_status_json_no_env", Args: []string{"worker", "status", "--json"}},
	{Name: "worker_serve_missing_env_json", Args: []string{"worker", "serve", "--json"}},
	{Name: "unknown_command_plain", Args: []string{"nosuch"}},
	{Name: "unknown_command_json_sanitized", Args: []string{"bad\x1b[31mcmd", "--json"}},
	{Name: "help_missing_topic", Args: []string{"help", "nosuchtopic"}},
	{Name: "dynamic_connector_missing_path_json", Args: []string{"github", "--json"}},
	{Name: "dynamic_connector_unknown_path_json", Args: []string{"github", "definitely-not-command", "--json"}},
	{Name: "hidden_extract_help_json", Args: []string{"extract", "--help", "--json"}},
	{Name: "hidden_worker_help_json", Args: []string{"worker", "--help", "--json"}},
}

func TestGoldenTranscripts(t *testing.T) {
	t.Setenv("POLYMETRICS_TEMPORAL_ADDR", "")
	t.Setenv("PM_TEMPORAL_ADDR", "")

	path := filepath.Join("testdata", "golden_transcripts.json")
	if os.Getenv("POLYMETRICS_UPDATE_GOLDEN_TRANSCRIPTS") == "1" {
		writeGoldenTranscripts(t, path)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden transcripts: %v", err)
	}
	var transcripts []goldenTranscript
	if err := json.Unmarshal(content, &transcripts); err != nil {
		t.Fatalf("parse golden transcripts: %v", err)
	}
	if len(transcripts) < 75 || len(transcripts) > 95 {
		t.Fatalf("golden transcript count = %d, want about 80", len(transcripts))
	}
	assertGoldenInputsMatchFixture(t, transcripts)

	for _, tt := range transcripts {
		t.Run(tt.Name, func(t *testing.T) {
			got := runTranscript(tt.Args)
			if got.ExitCode != tt.ExitCode {
				t.Fatalf("exit code = %d, want %d\nstdout=%s\nstderr=%s", got.ExitCode, tt.ExitCode, got.Stdout, got.Stderr)
			}
			if got.Stdout != tt.Stdout {
				t.Fatalf("stdout mismatch (-want +got):\n%s", diffStrings(tt.Stdout, got.Stdout))
			}
			if got.Stderr != tt.Stderr {
				t.Fatalf("stderr mismatch (-want +got):\n%s", diffStrings(tt.Stderr, got.Stderr))
			}
			if strings.Contains(got.Stdout, "\x1b") || strings.Contains(got.Stderr, "\x1b") {
				t.Fatalf("transcript retained ANSI escape: stdout=%q stderr=%q", got.Stdout, got.Stderr)
			}
		})
	}
}

func TestGoldenDocsGenerateMatchesTrackedCLIManuals(t *testing.T) {
	dir := t.TempDir()
	generatedCLI := filepath.Join(dir, "cli")
	generatedConnectors := filepath.Join(dir, "connectors")

	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"docs", "generate", "--dir", generatedCLI, "--connectors-dir", generatedConnectors}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("docs generate code = %d\nstdout=%s\nstderr=%s", code, stdout.String(), stderr.String())
	}
	if stdout.String() == "" || stderr.String() != "" {
		t.Fatalf("unexpected docs generate output: stdout=%q stderr=%q", stdout.String(), stderr.String())
	}

	trackedCLI := filepath.Join("..", "..", "docs", "cli")
	want, err := readMarkdownTree(trackedCLI)
	if err != nil {
		t.Fatalf("read tracked CLI docs: %v", err)
	}
	got, err := readMarkdownTree(generatedCLI)
	if err != nil {
		t.Fatalf("read generated CLI docs: %v", err)
	}
	if !reflect.DeepEqual(mapKeys(got), mapKeys(want)) {
		t.Fatalf("generated docs files mismatch\nwant=%v\ngot=%v", mapKeys(want), mapKeys(got))
	}
	for rel, wantContent := range want {
		gotContent := got[rel]
		if gotContent != wantContent {
			t.Fatalf("generated docs drift for %s\nwant:\n%s\n--- got:\n%s", rel, wantContent, gotContent)
		}
	}
}

func assertGoldenInputsMatchFixture(t *testing.T, transcripts []goldenTranscript) {
	t.Helper()
	if len(transcripts) != len(goldenTranscriptInputs) {
		t.Fatalf("fixture contains %d transcripts, input list contains %d", len(transcripts), len(goldenTranscriptInputs))
	}
	seen := map[string]struct{}{}
	for i, input := range goldenTranscriptInputs {
		got := transcripts[i]
		if _, ok := seen[input.Name]; ok {
			t.Fatalf("duplicate golden transcript name %q", input.Name)
		}
		seen[input.Name] = struct{}{}
		if got.Name != input.Name || !reflect.DeepEqual(got.Args, input.Args) {
			t.Fatalf("fixture entry %d mismatch: got %q %v, want %q %v", i, got.Name, got.Args, input.Name, input.Args)
		}
	}
}

func writeGoldenTranscripts(t *testing.T, path string) {
	t.Helper()
	transcripts := make([]goldenTranscript, 0, len(goldenTranscriptInputs))
	for _, input := range goldenTranscriptInputs {
		got := runTranscript(input.Args)
		got.Name = input.Name
		got.Args = append(make([]string, 0, len(input.Args)), input.Args...)
		transcripts = append(transcripts, got)
	}
	content, err := json.MarshalIndent(transcripts, "", "  ")
	if err != nil {
		t.Fatalf("marshal golden transcripts: %v", err)
	}
	content = append(content, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create golden directory: %v", err)
	}
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("write golden transcripts: %v", err)
	}
}

func runTranscript(args []string) goldenTranscript {
	var stdout, stderr bytes.Buffer
	code := cli.Run(args, &stdout, &stderr)
	return goldenTranscript{ExitCode: code, Stdout: stdout.String(), Stderr: stderr.String()}
}

func readMarkdownTree(root string) (map[string]string, error) {
	out := map[string]string{}
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}
		content, err := os.ReadFile(filepath.Join(root, entry.Name()))
		if err != nil {
			return nil, err
		}
		out[entry.Name()] = string(content)
	}
	return out, nil
}

func mapKeys(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func diffStrings(want, got string) string {
	if len(want) > 1000 || len(got) > 1000 {
		return "want:\n" + want + "\n--- got:\n" + got
	}
	return "want " + strconvQuote(want) + "\ngot  " + strconvQuote(got)
}

func strconvQuote(value string) string {
	b, _ := json.Marshal(value)
	return string(b)
}
