package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"polymetrics.ai/internal/app"
)

func TestRunCredentialsRejectsRawEmailStdinControlsBeforePersistence(t *testing.T) {
	root := t.TempDir()
	if err := app.InitProject(root); err != nil {
		t.Fatalf("InitProject: %v", err)
	}
	project, err := app.Open(root)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	args := []string{
		"add", "email-stdin-control", "--connector", "email",
		"--config", "imap_host=imap.example.invalid",
		"--config", "imap_port=993",
		"--config", "imap_security=tls",
		"--config", "smtp_host=smtp.example.invalid",
		"--config", "smtp_port=465",
		"--config", "smtp_security=tls",
		"--config", "username=sender@example.invalid",
		"--value-stdin", "password",
	}
	var stdout bytes.Buffer
	err = runCredentials(context.Background(), project, args, strings.NewReader("\r\n"), &stdout, false)
	if err == nil {
		t.Fatal("runCredentials accepted a raw stdin control character for Email")
	}
	if !strings.Contains(err.Error(), "password") || !strings.Contains(strings.ToLower(err.Error()), "control") {
		t.Fatalf("runCredentials error = %q, want password control constraint", err)
	}
	if strings.Contains(err.Error(), "\r") || strings.Contains(err.Error(), "\n") {
		t.Fatalf("runCredentials error exposed supplied input: %q", err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("runCredentials wrote success output after rejected stdin input: %q", stdout.String())
	}
	if credentials := project.ListCredentials(); len(credentials) != 0 {
		t.Fatalf("ListCredentials() = %#v, want no persisted credentials", credentials)
	}
	entries, err := os.ReadDir(filepath.Join(root, ".polymetrics", "vault"))
	if err != nil {
		t.Fatalf("ReadDir(vault): %v", err)
	}
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".enc") {
			t.Fatalf("runCredentials persisted vault entry %q after stdin validation failure", entry.Name())
		}
	}
}
