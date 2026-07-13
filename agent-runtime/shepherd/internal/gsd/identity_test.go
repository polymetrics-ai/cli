package gsd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadSessionIdentityProjectsMetadataWithoutMessageContent(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	path := filepath.Join(root, "session.jsonl")
	raw := []byte("{\"type\":\"model_change\",\"provider\":\"openai-codex\",\"modelId\":\"gpt-5.6-sol\",\"secret\":\"must-not-project\"}\n" +
		"{\"type\":\"thinking_level_change\",\"thinkingLevel\":\"high\"}\n" +
		"{\"type\":\"message\",\"message\":{\"role\":\"assistant\",\"provider\":\"openai-codex\",\"model\":\"gpt-5.6-sol\",\"content\":\"private\"}}\n")
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	model, thinking, err := ReadSessionIdentity(root)
	if err != nil {
		t.Fatal(err)
	}
	if model != "openai-codex/gpt-5.6-sol" || thinking != "high" {
		t.Fatalf("identity=%s/%s", model, thinking)
	}
}
