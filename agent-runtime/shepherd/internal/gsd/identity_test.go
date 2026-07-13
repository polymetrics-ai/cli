package gsd

import (
	"os"
	"path/filepath"
	"testing"
	"time"
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

func TestLatestSessionIDIsBoundToExactWorktree(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	wantedDir := filepath.Join(root, "work-a")
	otherDir := filepath.Join(root, "work-b")
	wanted := filepath.Join(root, "wanted.jsonl")
	other := filepath.Join(root, "other.jsonl")
	if err := os.WriteFile(wanted, []byte(`{"type":"session","id":"019f5d4a-9fb4-7852-b640-d6fdf71bd3d9","cwd":"`+wantedDir+`"}`+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(other, []byte(`{"type":"session","id":"019f5d4b-9fb4-7852-b640-d6fdf71bd3d0","cwd":"`+otherDir+`"}`+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	if err := os.Chtimes(wanted, now.Add(-time.Minute), now.Add(-time.Minute)); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(other, now, now); err != nil {
		t.Fatal(err)
	}
	id, err := LatestSessionID(root, wantedDir)
	if err != nil {
		t.Fatal(err)
	}
	if id != "019f5d4a-9fb4-7852-b640-d6fdf71bd3d9" {
		t.Fatalf("session id=%q", id)
	}
}
