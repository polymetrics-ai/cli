package cli_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"polymetrics.ai/internal/cli"
)

func TestQueryRunFlagFormsPreserveLegacySemantics(t *testing.T) {
	root := setupAgentModeQueryProject(t)

	var stdout, stderr bytes.Buffer
	args := []string{
		"query", "run", "extra-positional",
		"--table", "ignored",
		"--table=customers",
		"--limit", "1",
		"--limit=2",
		"--fields", "id,email",
		"--fields=name",
		"--fields",
		"--unknown", "value",
		"--root", root, "--json",
	}
	code := cli.Run(args, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run(%v) code = %d, want 0; stdout=%s stderr=%s", args, code, stdout.String(), stderr.String())
	}
	if strings.Contains(stdout.String()+stderr.String(), "unknown flag") {
		t.Fatalf("Run(%v) rejected legacy-tolerated unknown flag: stdout=%s stderr=%s", args, stdout.String(), stderr.String())
	}

	var env struct {
		Kind  string           `json:"kind"`
		Count int              `json:"count"`
		Rows  []map[string]any `json:"rows"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &env); err != nil {
		t.Fatalf("decode query json: %v\n%s", err, stdout.String())
	}
	if env.Kind != "QueryResult" || env.Count != 2 || len(env.Rows) != 2 {
		t.Fatalf("envelope = %+v, want QueryResult count/rows 2", env)
	}
	for i, row := range env.Rows {
		for _, field := range []string{"id", "email", "name"} {
			if _, ok := row[field]; !ok {
				t.Fatalf("row %d missing projected field %q: %v", i, field, row)
			}
		}
		if _, ok := row["notes"]; ok {
			t.Fatalf("row %d leaked non-projected notes: %v", i, row)
		}
	}
}

func TestQueryRunSQLFlagFormsPreserveLegacySemantics(t *testing.T) {
	root := setupAgentModeQueryProject(t)

	var stdout, stderr bytes.Buffer
	args := []string{
		"query", "run",
		"--table", "missing_table",
		"--sql", "select * from missing_table",
		"--sql=select * from customers",
		"--limit", "1",
		"--limit", "2",
		"--root", root, "--json",
	}
	code := cli.Run(args, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run(%v) code = %d, want 0; stdout=%s stderr=%s", args, code, stdout.String(), stderr.String())
	}

	var env struct {
		Kind  string           `json:"kind"`
		Count int              `json:"count"`
		Rows  []map[string]any `json:"rows"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &env); err != nil {
		t.Fatalf("decode query json: %v\n%s", err, stdout.String())
	}
	if env.Kind != "QueryResult" || env.Count != 2 || len(env.Rows) != 2 {
		t.Fatalf("envelope = %+v, want SQL last-wins QueryResult count/rows 2", env)
	}
}

func TestQueryInvalidActionIsUsageBeforeProjectOpen(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"query", "bogus", "--json"}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("Run(query bogus --json) code = %d, want 2; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), `"category": "usage"`) {
		t.Fatalf("invalid query action did not produce usage error: stdout=%s stderr=%s", stdout.String(), stderr.String())
	}
	if strings.Contains(stdout.String(), `"kind": "CommandManual"`) {
		t.Fatalf("invalid query action rendered contextual help instead of usage error: stdout=%s", stdout.String())
	}
}

func TestQueryRunRejectsWriteSQL(t *testing.T) {
	root := setupAgentModeQueryProject(t)

	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{
		"query", "run",
		"--sql", "DELETE FROM customers",
		"--root", root, "--json",
	}, &stdout, &stderr)
	if code == 0 {
		t.Fatalf("write SQL unexpectedly succeeded; stdout=%s stderr=%s", stdout.String(), stderr.String())
	}
	if strings.Contains(stdout.String(), `"kind": "QueryResult"`) {
		t.Fatalf("write SQL returned query result instead of rejection: stdout=%s stderr=%s", stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), `"kind": "Error"`) {
		t.Fatalf("write SQL rejection did not return Error envelope: stdout=%s stderr=%s", stdout.String(), stderr.String())
	}
}
