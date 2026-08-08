package cli

import (
	"os"
	"testing"

	"polymetrics.ai/internal/connectors"
)

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

	command := connectors.CommandSurfaceCommand{
		Intent:             "direct_write",
		AcceptsSecretInput: true,
		RedactFields:       []string{"password"},
		Flags:              []connectors.CommandSurfaceFlag{{Name: "password", MapsTo: "body.password"}},
	}
	flags := parsedFlags{values: map[string][]string{"value-stdin": {"password"}}}
	got, err := applySensitiveCommandInputs(map[string][]string{"username": {"fixture-user"}}, flags, command)
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
