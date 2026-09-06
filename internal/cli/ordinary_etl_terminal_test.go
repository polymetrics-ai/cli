package cli_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/cli"
)

func TestCLI_OrdinaryETLFailurePublishesOneTerminalRun(t *testing.T) {
	root := t.TempDir()
	missingSource := filepath.Join(root, "missing-records.jsonl")
	for _, args := range [][]string{
		{"init", "--root", root, "--json"},
		{"credentials", "add", "file-local", "--connector", "file", "--config", "path=" + missingSource, "--root", root, "--json"},
		{"credentials", "add", "warehouse-local", "--connector", "warehouse", "--config", "path=" + filepath.Join(root, ".polymetrics", "warehouse"), "--root", root, "--json"},
		{"connections", "create", "file-to-warehouse", "--source", "file:file-local", "--destination", "warehouse:warehouse-local", "--stream", "file", "--table", "imported_records", "--root", root, "--json"},
	} {
		var stdout, stderr bytes.Buffer
		if code := cli.Run(args, &stdout, &stderr); code != 0 {
			t.Fatalf("setup Run(%v) code = %d stderr=%s stdout=%s", args, code, stderr.String(), stdout.String())
		}
	}

	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"etl", "run", "--connection", "file-to-warehouse", "--stream", "file", "--root", root, "--json"}, &stdout, &stderr)
	if code == 0 {
		t.Fatalf("Run(etl run) code = 0, want the source failure; stdout=%s", stdout.String())
	}

	statePayload, err := os.ReadFile(filepath.Join(root, ".polymetrics", "state", "state.json"))
	if err != nil {
		t.Fatalf("ReadFile(durable state) error = %v", err)
	}
	var persisted struct {
		Runs []app.Run `json:"runs"`
	}
	if err := json.Unmarshal(statePayload, &persisted); err != nil {
		t.Fatalf("decode durable state error = %v", err)
	}
	if len(persisted.Runs) != 1 || persisted.Runs[0].Status != "failed" || persisted.Runs[0].ID == "" || persisted.Runs[0].CompletedAt.IsZero() {
		t.Fatalf("durable terminal runs = %#v, want exactly one failed terminal run", persisted.Runs)
	}

	var output struct {
		Kind string  `json:"kind"`
		Run  app.Run `json:"run"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		t.Fatalf("decode CLI terminal envelope error = %v; stdout=%s", err, stdout.String())
	}
	if output.Kind != "ETLRun" {
		t.Fatalf("CLI terminal envelope kind = %q, want ETLRun; stdout=%s", output.Kind, stdout.String())
	}
	if !reflect.DeepEqual(output.Run, persisted.Runs[0]) {
		t.Fatalf("CLI terminal run = %#v, want exact durable run %#v", output.Run, persisted.Runs[0])
	}
}

func TestCLI_PreIORefusalsPersistExactTerminalRun(t *testing.T) {
	for _, mode := range []string{"full_refresh_overwrite_deduped", "incremental_append_deduped"} {
		t.Run(mode, func(t *testing.T) {
			root := t.TempDir()
			sourcePath := filepath.Join(root, "records.jsonl")
			if err := os.WriteFile(sourcePath, []byte("{\"id\":\"one\",\"occurred_at\":\"2026-08-21T00:00:00Z\"}\n"), 0o600); err != nil {
				t.Fatalf("WriteFile(source) error = %v", err)
			}
			for _, args := range [][]string{
				{"init", "--root", root, "--json"},
				{"credentials", "add", "file-local", "--connector", "file", "--config", "path=" + sourcePath, "--root", root, "--json"},
				{"credentials", "add", "warehouse-local", "--connector", "warehouse", "--config", "path=" + filepath.Join(root, ".polymetrics", "warehouse"), "--root", root, "--json"},
				{"connections", "create", "file-to-warehouse", "--source", "file:file-local", "--destination", "warehouse:warehouse-local", "--stream", "file", "--table", "imported_records", "--sync-mode", mode, "--primary-key", "id", "--cursor", "occurred_at", "--root", root, "--json"},
			} {
				var stdout, stderr bytes.Buffer
				if code := cli.Run(args, &stdout, &stderr); code != 0 {
					t.Fatalf("setup Run(%v) code = %d stderr=%s stdout=%s", args, code, stderr.String(), stdout.String())
				}
			}

			var stdout, stderr bytes.Buffer
			code := cli.Run([]string{"etl", "run", "--connection", "file-to-warehouse", "--stream", "file", "--root", root, "--json"}, &stdout, &stderr)
			if code != 1 {
				t.Fatalf("Run(etl run) code = %d, want categorized internal pre-I/O exit 1; stdout=%s", code, stdout.String())
			}
			var envelope struct {
				Kind string  `json:"kind"`
				Run  app.Run `json:"run"`
			}
			if err := json.Unmarshal(stdout.Bytes(), &envelope); err != nil {
				t.Fatalf("decode terminal pre-I/O refusal: %v; stdout=%s", err, stdout.String())
			}

			statePayload, err := os.ReadFile(filepath.Join(root, ".polymetrics", "state", "state.json"))
			if err != nil {
				t.Fatalf("ReadFile(durable state) error = %v", err)
			}
			var persisted struct {
				Runs []app.Run `json:"runs"`
			}
			if err := json.Unmarshal(statePayload, &persisted); err != nil {
				t.Fatalf("decode durable state error = %v", err)
			}
			if envelope.Kind != "ETLRun" || envelope.Run.ID == "" || envelope.Run.Status != "failed" || envelope.Run.CompletedAt.IsZero() || !strings.Contains(envelope.Run.Error, "is not executable") {
				t.Fatalf("terminal pre-I/O refusal envelope = %#v, want exact failed ETLRun", envelope)
			}
			if len(persisted.Runs) != 1 || !reflect.DeepEqual(envelope.Run, persisted.Runs[0]) {
				t.Fatalf("terminal pre-I/O refusal run = %#v, want exact stored terminal run %#v", envelope.Run, persisted.Runs)
			}
		})
	}
}

func TestCLI_OrdinaryETLRuntimeRecorderFailurePublishesTerminalRun(t *testing.T) {
	root := t.TempDir()
	sourcePath := filepath.Join(root, "records.jsonl")
	if err := os.WriteFile(sourcePath, []byte("{\"id\":\"one\"}\n"), 0o600); err != nil {
		t.Fatalf("WriteFile(source) error = %v", err)
	}
	for _, args := range [][]string{
		{"init", "--root", root, "--json"},
		{"credentials", "add", "file-local", "--connector", "file", "--config", "path=" + sourcePath, "--root", root, "--json"},
		{"credentials", "add", "warehouse-local", "--connector", "warehouse", "--config", "path=" + filepath.Join(root, ".polymetrics", "warehouse"), "--root", root, "--json"},
		{"connections", "create", "file-to-warehouse", "--source", "file:file-local", "--destination", "warehouse:warehouse-local", "--stream", "file", "--table", "imported_records", "--root", root, "--json"},
	} {
		var stdout, stderr bytes.Buffer
		if code := cli.Run(args, &stdout, &stderr); code != 0 {
			t.Fatalf("setup Run(%v) code = %d stderr=%s stdout=%s", args, code, stderr.String(), stdout.String())
		}
	}

	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"etl", "run", "--connection", "file-to-warehouse", "--stream", "file", "--runtime", "--root", root, "--json"}, &stdout, &stderr)
	if code == 0 {
		t.Fatalf("Run(etl run --runtime) code = 0, want unhealthy runtime recorder; stdout=%s", stdout.String())
	}

	statePayload, err := os.ReadFile(filepath.Join(root, ".polymetrics", "state", "state.json"))
	if err != nil {
		t.Fatalf("ReadFile(durable state) error = %v", err)
	}
	var persisted struct {
		Runs []app.Run `json:"runs"`
	}
	if err := json.Unmarshal(statePayload, &persisted); err != nil {
		t.Fatalf("decode durable state error = %v", err)
	}
	if len(persisted.Runs) != 1 || persisted.Runs[0].Status != "completed" || persisted.Runs[0].ID == "" || persisted.Runs[0].CompletedAt.IsZero() {
		t.Fatalf("durable terminal runs = %#v, want exactly one completed terminal run", persisted.Runs)
	}

	var output struct {
		Kind            string  `json:"kind"`
		Run             app.Run `json:"run"`
		RuntimeRecorded bool    `json:"runtime_recorded"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		t.Fatalf("decode CLI terminal envelope error = %v; stdout=%s", err, stdout.String())
	}
	if output.Kind != "ETLRun" || output.RuntimeRecorded {
		t.Fatalf("CLI terminal envelope = %#v, want durable ETLRun with runtime_recorded=false", output)
	}
	if !reflect.DeepEqual(output.Run, persisted.Runs[0]) {
		t.Fatalf("CLI terminal run = %#v, want exact durable run %#v", output.Run, persisted.Runs[0])
	}
}
