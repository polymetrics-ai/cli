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

func seedCLIWarehouseTable(t *testing.T, root, table string, rows []map[string]any) {
	t.Helper()
	dir := filepath.Join(root, ".polymetrics", "warehouse")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("mkdir warehouse: %v", err)
	}
	f, err := os.Create(filepath.Join(dir, table+".jsonl"))
	if err != nil {
		t.Fatalf("create warehouse table: %v", err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	for _, row := range rows {
		if err := enc.Encode(row); err != nil {
			t.Fatalf("encode row: %v", err)
		}
	}
}

func setupAgentModeQueryProject(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	var stdout, stderr bytes.Buffer
	if code := cli.Run([]string{"init", "--root", root, "--json"}, &stdout, &stderr); code != 0 {
		t.Fatalf("init code = %d stderr = %s stdout = %s", code, stderr.String(), stdout.String())
	}
	seedCLIWarehouseTable(t, root, "customers", []map[string]any{
		{"id": "cus_1", "email": "ada@example.com", "name": "Ada", "notes": "long note"},
		{"id": "cus_2", "email": "grace@example.com", "name": "Grace", "notes": "long note"},
		{"id": "cus_3", "email": "katherine@example.com", "name": "Katherine", "notes": "long note"},
	})
	return root
}

func TestQueryRunAgentModeSummaryProjectsFields(t *testing.T) {
	root := setupAgentModeQueryProject(t)

	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{
		"--root", root,
		"query", "run",
		"--table", "customers",
		"--agent-mode", "summary",
		"--fields", "id,email",
		"--sample", "2",
		"--json",
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("query code = %d stderr = %s stdout = %s", code, stderr.String(), stdout.String())
	}

	if strings.Count(strings.TrimSpace(stdout.String()), "\n") != 0 {
		t.Fatalf("summary should be compact single-line JSON, got:\n%s", stdout.String())
	}

	var got struct {
		Kind   string           `json:"kind"`
		Count  int              `json:"count"`
		Fields []string         `json:"fields"`
		Sample []map[string]any `json:"sample"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("summary is not JSON: %v: %s", err, stdout.String())
	}
	if got.Kind != "QueryResult" || got.Count != 3 {
		t.Fatalf("summary = %+v, want QueryResult count 3", got)
	}
	if strings.Join(got.Fields, ",") != "email,id" {
		t.Fatalf("fields = %v, want [email id]", got.Fields)
	}
	if len(got.Sample) != 2 {
		t.Fatalf("sample rows = %d, want 2", len(got.Sample))
	}
	for i, row := range got.Sample {
		if _, ok := row["id"]; !ok {
			t.Fatalf("sample row %d missing id: %v", i, row)
		}
		if _, ok := row["email"]; !ok {
			t.Fatalf("sample row %d missing email: %v", i, row)
		}
		if _, ok := row["notes"]; ok {
			t.Fatalf("sample row %d leaked non-projected notes: %v", i, row)
		}
	}
}

func TestQueryRunAgentModeStreamProjectsNDJSON(t *testing.T) {
	root := setupAgentModeQueryProject(t)

	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{
		"--root", root,
		"query", "run",
		"--table", "customers",
		"--agent-mode", "stream",
		"--fields", "id,email",
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("query code = %d stderr = %s stdout = %s", code, stderr.String(), stdout.String())
	}

	lines := strings.Split(strings.TrimRight(stdout.String(), "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("NDJSON lines = %d, want 3: %q", len(lines), stdout.String())
	}
	for i, line := range lines {
		var row map[string]any
		if err := json.Unmarshal([]byte(line), &row); err != nil {
			t.Fatalf("line %d is not JSON: %v: %q", i, err, line)
		}
		if len(row) != 2 || row["id"] == nil || row["email"] == nil {
			t.Fatalf("line %d row = %v, want only id/email", i, row)
		}
	}
}
