package cli_test

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/cli"
)

func TestConnectionsCreateFlagFormsPreserveLegacySemantics(t *testing.T) {
	root := setupConnectionsProject(t)

	var stdout, stderr bytes.Buffer
	args := []string{
		"connections", "create", "forms", "extra-positional",
		"--source", "ignored:ignored", "--source", "sample:sample-local",
		"--destination=ignored:ignored", "--destination", "warehouse:warehouse-local",
		"--stream=ignored_stream", "--stream", "customers",
		"--sync-mode", "incremental_append_deduped", "--sync-mode", "full_refresh_overwrite",
		"--cursor", "updated_at", "--cursor",
		"--primary-key", "id", "--primary-key=email",
		"--table", "first_table", "--table",
		"--source-config", "alpha=one", "--source-config=beta=two",
		"--destination-config", "mode=local",
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
		Kind       string         `json:"kind"`
		Connection app.Connection `json:"connection"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &env); err != nil {
		t.Fatalf("decode connection json: %v\n%s", err, stdout.String())
	}
	if env.Kind != "Connection" {
		t.Fatalf("kind = %q, want Connection", env.Kind)
	}
	conn := env.Connection
	if conn.Name != "forms" {
		t.Fatalf("connection name = %q, want forms", conn.Name)
	}
	if got, want := conn.Source.Connector+":"+conn.Source.Credential, "sample:sample-local"; got != want {
		t.Fatalf("source = %q, want %q", got, want)
	}
	if got, want := conn.Destination.Connector+":"+conn.Destination.Credential, "warehouse:warehouse-local"; got != want {
		t.Fatalf("destination = %q, want %q", got, want)
	}
	if got, want := conn.Source.Config["alpha"], "one"; got != want {
		t.Fatalf("source config alpha = %q, want %q", got, want)
	}
	if got, want := conn.Source.Config["beta"], "two"; got != want {
		t.Fatalf("source config beta = %q, want %q", got, want)
	}
	if got, want := conn.Destination.Config["mode"], "local"; got != want {
		t.Fatalf("destination config mode = %q, want %q", got, want)
	}
	stream, ok := conn.Streams["customers"]
	if !ok {
		t.Fatalf("connection streams missing customers: %#v", conn.Streams)
	}
	if _, ok := conn.Streams["ignored_stream"]; ok {
		t.Fatalf("repeated --stream did not use last value: %#v", conn.Streams)
	}
	if got, want := stream.SyncMode, "full_refresh_overwrite"; got != want {
		t.Fatalf("sync mode = %q, want %q", got, want)
	}
	if got, want := stream.CursorField, "true"; got != want {
		t.Fatalf("bare --cursor value = %q, want %q", got, want)
	}
	if got, want := strings.Join(stream.PrimaryKey, ","), "id,email"; got != want {
		t.Fatalf("primary keys = %q, want %q", got, want)
	}
	if got, want := stream.DestinationTable, "true"; got != want {
		t.Fatalf("bare --table value = %q, want %q", got, want)
	}
}

func TestConnectionsListPreservesUnknownFlagAndExtraArgTolerance(t *testing.T) {
	root := setupConnectionsProject(t)
	runConnectionsCLISuccess(t, []string{
		"connections", "create", "listed",
		"--source", "sample:sample-local",
		"--destination", "warehouse:warehouse-local",
		"--stream", "customers",
		"--primary-key", "id",
		"--root", root,
		"--json",
	})

	var stdout, stderr bytes.Buffer
	args := []string{"connections", "list", "extra-positional", "--unknown", "value", "--root", root, "--json"}
	code := cli.Run(args, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run(%v) code = %d, want 0; stdout=%s stderr=%s", args, code, stdout.String(), stderr.String())
	}
	if strings.Contains(stdout.String()+stderr.String(), "unknown flag") {
		t.Fatalf("Run(%v) rejected legacy-tolerated unknown flag: stdout=%s stderr=%s", args, stdout.String(), stderr.String())
	}
	var env struct {
		Kind        string           `json:"kind"`
		Connections []app.Connection `json:"connections"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &env); err != nil {
		t.Fatalf("decode list json: %v\n%s", err, stdout.String())
	}
	if env.Kind != "ConnectionList" {
		t.Fatalf("kind = %q, want ConnectionList", env.Kind)
	}
	if len(env.Connections) != 1 || env.Connections[0].Name != "listed" {
		t.Fatalf("connections = %+v, want one listed connection", env.Connections)
	}
}

func TestConnectionsInvalidActionIsUsageBeforeProjectOpen(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"connections", "bogus", "--json"}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("Run(connections bogus --json) code = %d, want 2; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), `"category": "usage"`) {
		t.Fatalf("invalid connections action did not produce usage error: stdout=%s stderr=%s", stdout.String(), stderr.String())
	}
	if strings.Contains(stdout.String(), `"kind": "CommandManual"`) {
		t.Fatalf("invalid connections action rendered contextual help instead of usage error: stdout=%s", stdout.String())
	}
}

func setupConnectionsProject(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	runConnectionsCLISuccess(t, []string{"init", "--root", root, "--json"})
	runConnectionsCLISuccess(t, []string{"credentials", "add", "sample-local", "--connector", "sample", "--root", root, "--json"})
	runConnectionsCLISuccess(t, []string{
		"credentials", "add", "warehouse-local",
		"--connector", "warehouse",
		"--config", "path=" + filepath.Join(root, ".polymetrics", "warehouse"),
		"--root", root,
		"--json",
	})
	return root
}

func runConnectionsCLISuccess(t *testing.T, args []string) string {
	t.Helper()
	var stdout, stderr bytes.Buffer
	if code := cli.Run(args, &stdout, &stderr); code != 0 {
		t.Fatalf("Run(%v) code = %d, want 0; stdout=%s stderr=%s", args, code, stdout.String(), stderr.String())
	}
	return stdout.String()
}
