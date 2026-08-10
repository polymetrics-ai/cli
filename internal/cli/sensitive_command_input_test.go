package cli

import (
	"io"
	"os"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
)

type sensitiveInputMetadataConnector struct {
	connectors.Connector
	metadata connectors.OperationDirectWriteMetadata
}

func (c sensitiveInputMetadataConnector) OperationDirectWriteMetadata(string) (connectors.OperationDirectWriteMetadata, error) {
	return c.metadata, nil
}

func TestApplySensitiveCommandInputsReadsValueStdin(t *testing.T) {
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	if _, err := writer.WriteString("fixture-password\r\n"); err != nil {
		t.Fatalf("write stdin: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close stdin writer: %v", err)
	}
	previous := os.Stdin
	os.Stdin = reader
	t.Cleanup(func() {
		os.Stdin = previous
		_ = reader.Close()
	})

	command := sensitiveInputTestCommand()
	flags := parsedFlags{values: map[string][]string{"value-stdin": {"password"}}}
	got, err := applySensitiveCommandInputs(map[string][]string{"username": {"fixture-user"}}, flags, command, 1024)
	if err != nil {
		t.Fatalf("applySensitiveCommandInputs: %v", err)
	}
	if got["password"][0] != "fixture-password" {
		t.Fatalf("stdin password = %#v, want newline-trimmed input", got["password"])
	}
	if got["username"][0] != "fixture-user" {
		t.Fatalf("unrelated command input = %#v, want preserved", got["username"])
	}
}

func TestApplySensitiveCommandInputsValidatesValueStdinFieldBeforeReading(t *testing.T) {
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	if _, err := writer.WriteString("unread-fixture"); err != nil {
		t.Fatalf("write stdin: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close stdin writer: %v", err)
	}
	previous := os.Stdin
	os.Stdin = reader
	t.Cleanup(func() {
		os.Stdin = previous
		_ = reader.Close()
	})

	_, err = applySensitiveCommandInputs(nil, parsedFlags{values: map[string][]string{"value-stdin": {"unknown"}}}, sensitiveInputTestCommand(), 1024)
	if err == nil || !strings.Contains(err.Error(), "not a declared redacted command input") {
		t.Fatalf("applySensitiveCommandInputs error = %v, want field validation", err)
	}
	remaining, readErr := io.ReadAll(reader)
	if readErr != nil {
		t.Fatalf("read remaining stdin: %v", readErr)
	}
	if string(remaining) != "unread-fixture" {
		t.Fatalf("remaining stdin = %q, want unchanged input", remaining)
	}
}

func TestApplySensitiveCommandInputsRejectsOversizedValueStdin(t *testing.T) {
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	if _, err := writer.WriteString(strings.Repeat("x", 9)); err != nil {
		t.Fatalf("write stdin: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close stdin writer: %v", err)
	}
	previous := os.Stdin
	os.Stdin = reader
	t.Cleanup(func() {
		os.Stdin = previous
		_ = reader.Close()
	})

	_, err = applySensitiveCommandInputs(nil, parsedFlags{values: map[string][]string{"value-stdin": {"password"}}}, sensitiveInputTestCommand(), 8)
	if err == nil || !strings.Contains(err.Error(), "input is too large") {
		t.Fatalf("applySensitiveCommandInputs error = %v, want bounded stdin rejection", err)
	}
}

func TestSensitiveCommandInputMaxBytesUsesDeclaredOperationLimit(t *testing.T) {
	command := sensitiveInputTestCommand()
	command.Operation = "dockerhub.create_login"
	connector := sensitiveInputMetadataConnector{metadata: connectors.OperationDirectWriteMetadata{
		Operation:       command.Operation,
		MaxRequestBytes: 17,
	}}
	got, err := sensitiveCommandInputMaxBytes(connector, command, parsedFlags{values: map[string][]string{"value-stdin": {"password"}}})
	if err != nil {
		t.Fatalf("sensitiveCommandInputMaxBytes: %v", err)
	}
	if got != 17 {
		t.Fatalf("sensitiveCommandInputMaxBytes = %d, want 17", got)
	}
}

func sensitiveInputTestCommand() connectors.CommandSurfaceCommand {
	return connectors.CommandSurfaceCommand{
		Intent:             "direct_write",
		AcceptsSecretInput: true,
		RedactFields:       []string{"password"},
		Flags:              []connectors.CommandSurfaceFlag{{Name: "password", MapsTo: "body.password"}},
	}
}
