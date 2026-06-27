package cli

import (
	"bytes"
	"encoding/json"
	"testing"
)

func runExtractCLI(t *testing.T, dir, request string) map[string]any {
	t.Helper()
	var stdout, stderr bytes.Buffer
	code := Run([]string{
		"--root", dir, "--json",
		"extract", "--request", request,
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %s", code, stderr.String())
	}
	var env map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &env); err != nil {
		t.Fatalf("stdout not JSON: %v — %s", err, stdout.String())
	}
	return env
}

func TestExtract_RoutesSimpleQuery(t *testing.T) {
	dir := t.TempDir()
	initProject(t, dir)
	env := runExtractCLI(t, dir, "show top 10 customers by revenue")
	if env["route"] != "simple_query" {
		t.Fatalf("route = %v, want simple_query", env["route"])
	}
	if env["task_type"] != "simple_query" {
		t.Fatalf("task_type = %v, want simple_query", env["task_type"])
	}
}

func TestExtract_RoutesML(t *testing.T) {
	dir := t.TempDir()
	initProject(t, dir)
	env := runExtractCLI(t, dir, "predict which customers will churn")
	if env["route"] != "rlm_analysis" {
		t.Fatalf("route = %v, want rlm_analysis", env["route"])
	}
	if env["task_type"] != "ml" {
		t.Fatalf("task_type = %v, want ml", env["task_type"])
	}
}

func TestExtract_RoutesDataAnalysis(t *testing.T) {
	dir := t.TempDir()
	initProject(t, dir)
	env := runExtractCLI(t, dir, "show the monthly revenue trend by product")
	if env["route"] != "rlm_analysis" {
		t.Fatalf("route = %v, want rlm_analysis", env["route"])
	}
	if env["task_type"] != "data_analysis" {
		t.Fatalf("task_type = %v, want data_analysis", env["task_type"])
	}
}

func TestExtract_RequiresRequest(t *testing.T) {
	dir := t.TempDir()
	initProject(t, dir)
	var stdout, stderr bytes.Buffer
	code := Run([]string{"--root", dir, "extract"}, &stdout, &stderr)
	if code == 0 {
		t.Fatalf("want non-zero exit for missing --request, got 0")
	}
}
