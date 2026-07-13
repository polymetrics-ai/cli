package gsd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestReadSessionIdentityProjectsMetadataWithoutMessageContent(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	workDir := t.TempDir()
	path := filepath.Join(root, "session.jsonl")
	raw := []byte(`{"type":"session","id":"019f5d4a-9fb4-7852-b640-d6fdf71bd3d9","cwd":"` + workDir + `"}` + "\n" +
		"{\"type\":\"model_change\",\"provider\":\"openai-codex\",\"modelId\":\"gpt-5.6-sol\",\"secret\":\"must-not-project\"}\n" +
		"{\"type\":\"thinking_level_change\",\"thinkingLevel\":\"high\"}\n" +
		"{\"type\":\"message\",\"message\":{\"role\":\"assistant\",\"provider\":\"openai-codex\",\"model\":\"gpt-5.6-sol\",\"content\":\"private\"}}\n")
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	model, thinking, err := ReadSessionIdentity(root, workDir)
	if err != nil {
		t.Fatal(err)
	}
	if model != "openai-codex/gpt-5.6-sol" || thinking != "high" {
		t.Fatalf("identity=%s/%s", model, thinking)
	}
}

func TestReadSessionIdentityIgnoresOlderOversizedSession(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	workDir := t.TempDir()
	old := filepath.Join(root, "old.jsonl")
	oldRaw := []byte(`{"type":"session","id":"019f5d4a-9fb4-7852-b640-d6fdf71bd3d8","cwd":"` + workDir + `"}` + "\n" + strings.Repeat("x", 2*1024*1024) + "\n")
	if err := os.WriteFile(old, oldRaw, 0o600); err != nil {
		t.Fatal(err)
	}
	latest := filepath.Join(root, "latest.jsonl")
	latestRaw := []byte(`{"type":"session","id":"019f5d4a-9fb4-7852-b640-d6fdf71bd3d9","cwd":"` + workDir + `"}` + "\n" +
		`{"type":"thinking_level_change","thinkingLevel":"high"}` + "\n" +
		`{"type":"message","message":{"role":"assistant","provider":"openai-codex","model":"gpt-5.6-sol"}}` + "\n")
	if err := os.WriteFile(latest, latestRaw, 0o600); err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	if err := os.Chtimes(old, now.Add(-time.Minute), now.Add(-time.Minute)); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(latest, now, now); err != nil {
		t.Fatal(err)
	}
	model, thinking, err := ReadSessionIdentity(root, workDir)
	if err != nil || model != "openai-codex/gpt-5.6-sol" || thinking != "high" {
		t.Fatalf("identity=%s/%s err=%v", model, thinking, err)
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
