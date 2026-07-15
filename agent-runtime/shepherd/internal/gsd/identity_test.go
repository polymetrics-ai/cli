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
	raw := []byte(`{"type":"session","version":3,"timestamp":"2026-07-15T00:00:00Z","id":"019f5d4a-9fb4-7852-b640-d6fdf71bd3d9","cwd":"` + workDir + `"}` + "\n" +
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
	oldRaw := []byte(`{"type":"session","version":3,"timestamp":"2026-07-15T00:00:00Z","id":"019f5d4a-9fb4-7852-b640-d6fdf71bd3d8","cwd":"` + workDir + `"}` + "\n" + strings.Repeat("x", 2*1024*1024) + "\n")
	if err := os.WriteFile(old, oldRaw, 0o600); err != nil {
		t.Fatal(err)
	}
	latest := filepath.Join(root, "latest.jsonl")
	latestRaw := []byte(`{"type":"session","version":3,"timestamp":"2026-07-15T00:00:00Z","id":"019f5d4a-9fb4-7852-b640-d6fdf71bd3d9","cwd":"` + workDir + `"}` + "\n" +
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

func TestReadSessionIdentityExcludesPersistedSubagentSessions(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	workDir := t.TempDir()
	parent := filepath.Join(root, "parent.jsonl")
	child := filepath.Join(root, "child.jsonl")
	parentRaw := []byte(`{"type":"session","version":3,"timestamp":"2026-07-15T00:00:00Z","id":"019f5d4a-9fb4-7852-b640-d6fdf71bd3d9","cwd":"` + workDir + `"}` + "\n" +
		`{"type":"thinking_level_change","thinkingLevel":"high"}` + "\n" +
		`{"type":"message","message":{"role":"assistant","provider":"openai-codex","model":"gpt-5.6-sol"}}` + "\n")
	childRaw := []byte(`{"type":"session","version":3,"timestamp":"2026-07-15T00:00:00Z","id":"019f5d4b-9fb4-7852-b640-d6fdf71bd3d0","cwd":"` + workDir + `","parentSession":"` + parent + `"}` + "\n" +
		`{"type":"thinking_level_change","thinkingLevel":"medium"}` + "\n" +
		`{"type":"message","message":{"role":"assistant","provider":"openai-codex","model":"gpt-5.5"}}` + "\n")
	if err := os.WriteFile(parent, parentRaw, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(child, childRaw, 0o600); err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	if err := os.Chtimes(parent, now.Add(-time.Minute), now.Add(-time.Minute)); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(child, now, now); err != nil {
		t.Fatal(err)
	}
	model, thinking, err := ReadSessionIdentity(root, workDir)
	if err != nil {
		t.Fatal(err)
	}
	if model != "openai-codex/gpt-5.6-sol" || thinking != "high" {
		t.Fatalf("top-level identity=%s/%s", model, thinking)
	}
}

func TestReadSessionIdentityForRunBindsExactlyOneChangedTopLevelSession(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	workDir := t.TempDir()
	baseline, err := CaptureSessionIdentityBaseline(root, workDir)
	if err != nil {
		t.Fatal(err)
	}
	started := time.Now().Add(-time.Second)
	write := func(name, id string) {
		raw := []byte(`{"type":"session","version":3,"timestamp":"2026-07-15T00:00:00Z","id":"` + id + `","cwd":"` + workDir + `"}` + "\n" +
			`{"type":"thinking_level_change","thinkingLevel":"high"}` + "\n" +
			`{"type":"message","message":{"role":"assistant","provider":"openai-codex","model":"gpt-5.6-sol"}}` + "\n")
		if err := os.WriteFile(filepath.Join(root, name), raw, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write("current.jsonl", "019f5d4a-9fb4-7852-b640-d6fdf71bd3d9")
	model, thinking, err := ReadSessionIdentityForRun(root, workDir, baseline, started, "openai-codex/gpt-5.6-sol", "high")
	if err != nil || model != "openai-codex/gpt-5.6-sol" || thinking != "high" {
		t.Fatalf("current identity=%s/%s err=%v", model, thinking, err)
	}
	write("second.jsonl", "019f5d4b-9fb4-7852-b640-d6fdf71bd3d0")
	if _, _, err := ReadSessionIdentityForRun(root, workDir, baseline, started, "openai-codex/gpt-5.6-sol", "high"); err == nil {
		t.Fatal("multiple changed top-level sessions were accepted")
	}
}

func TestReadSessionIdentityDeltaRejectsSelectedFileReplacementAndAmbiguousRow(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	path := filepath.Join(root, "current.jsonl")
	write := func(model string, withMessage bool) {
		message := ""
		if withMessage {
			message = `,"message":{"role":"assistant","provider":"openai-codex","model":"gpt-5.6-sol"}`
		}
		raw := []byte(`{"type":"model_change","provider":"openai-codex","modelId":"` + model + `"` + message + `}` + "\n" +
			`{"type":"thinking_level_change","thinkingLevel":"high"}` + "\n")
		if err := os.WriteFile(path, raw, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write("gpt-5.6-sol", false)
	selected, err := os.Lstat(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(path, filepath.Join(root, "replaced.jsonl")); err != nil {
		t.Fatal(err)
	}
	write("gpt-5.6-sol", false)
	if _, _, _, err := readSessionIdentityDelta(path, 0, "openai-codex/gpt-5.6-sol", "high", selected); err == nil {
		t.Fatal("replacement between selection and read was accepted")
	}
	write("gpt-5.5", true)
	selected, err = os.Lstat(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := readSessionIdentityDelta(path, 0, "openai-codex/gpt-5.6-sol", "high", selected); err == nil {
		t.Fatal("wrong model transition hidden by a right assistant identity was accepted")
	}
}

func TestReadSessionIdentityForRunRejectsWrongThenRightTransition(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	workDir := t.TempDir()
	baseline, err := CaptureSessionIdentityBaseline(root, workDir)
	if err != nil {
		t.Fatal(err)
	}
	raw := []byte(`{"type":"session","version":3,"timestamp":"2026-07-15T00:00:00Z","id":"019f5d4a-9fb4-7852-b640-d6fdf71bd3d9","cwd":"` + workDir + `"}` + "\n" +
		`{"type":"model_change","provider":"openai-codex","modelId":"gpt-5.5"}` + "\n" +
		`{"type":"model_change","provider":"openai-codex","modelId":"gpt-5.6-sol"}` + "\n" +
		`{"type":"thinking_level_change","thinkingLevel":"high"}` + "\n")
	if err := os.WriteFile(filepath.Join(root, "current.jsonl"), raw, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := ReadSessionIdentityForRun(root, workDir, baseline, time.Now().Add(-time.Second), "openai-codex/gpt-5.6-sol", "high"); err == nil {
		t.Fatal("wrong-then-right model transition was accepted")
	}
}

func TestReadSessionIdentityRejectsAmbiguousHeadersSymlinksAndStaleEvidence(t *testing.T) {
	t.Parallel()
	workDir := t.TempDir()
	for name, header := range map[string]string{
		"unknown":   `{"type":"session","version":3,"timestamp":"2026-07-15T00:00:00Z","id":"019f5d4a-9fb4-7852-b640-d6fdf71bd3d9","cwd":"` + workDir + `","extra":true}`,
		"duplicate": `{"type":"session","version":3,"timestamp":"2026-07-15T00:00:00Z","id":"019f5d4a-9fb4-7852-b640-d6fdf71bd3d9","cwd":"` + workDir + `","cwd":"` + workDir + `"}`,
		"trailing":  `{"type":"session","version":3,"timestamp":"2026-07-15T00:00:00Z","id":"019f5d4a-9fb4-7852-b640-d6fdf71bd3d9","cwd":"` + workDir + `"}{}`,
	} {
		name, header := name, header
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			if err := os.WriteFile(filepath.Join(root, "session.jsonl"), []byte(header+"\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			if _, _, err := ReadSessionIdentity(root, workDir); err == nil {
				t.Fatal("ambiguous session header was accepted")
			}
		})
	}
	t.Run("symlink", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		target := filepath.Join(root, "target")
		if err := os.WriteFile(target, []byte(`{"type":"session","version":3,"timestamp":"2026-07-15T00:00:00Z","id":"019f5d4a-9fb4-7852-b640-d6fdf71bd3d9","cwd":"`+workDir+`"}`+"\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(target, filepath.Join(root, "session.jsonl")); err != nil {
			t.Fatal(err)
		}
		if _, _, err := ReadSessionIdentity(root, workDir); err == nil {
			t.Fatal("symlinked session evidence was accepted")
		}
	})
	t.Run("stale", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		path := filepath.Join(root, "session.jsonl")
		raw := []byte(`{"type":"session","version":3,"timestamp":"2026-07-15T00:00:00Z","id":"019f5d4a-9fb4-7852-b640-d6fdf71bd3d9","cwd":"` + workDir + `"}` + "\n" +
			`{"type":"thinking_level_change","thinkingLevel":"high"}` + "\n" +
			`{"type":"message","message":{"role":"assistant","provider":"openai-codex","model":"gpt-5.6-sol"}}` + "\n")
		if err := os.WriteFile(path, raw, 0o600); err != nil {
			t.Fatal(err)
		}
		if _, _, err := ReadSessionIdentitySince(root, workDir, time.Now().Add(time.Minute)); err == nil {
			t.Fatal("stale session identity was accepted")
		}
	})
}

func TestLatestSessionIDIsBoundToExactWorktree(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	wantedDir := filepath.Join(root, "work-a")
	otherDir := filepath.Join(root, "work-b")
	wanted := filepath.Join(root, "wanted.jsonl")
	other := filepath.Join(root, "other.jsonl")
	if err := os.WriteFile(wanted, []byte(`{"type":"session","version":3,"timestamp":"2026-07-15T00:00:00Z","id":"019f5d4a-9fb4-7852-b640-d6fdf71bd3d9","cwd":"`+wantedDir+`"}`+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(other, []byte(`{"type":"session","version":3,"timestamp":"2026-07-15T00:00:00Z","id":"019f5d4b-9fb4-7852-b640-d6fdf71bd3d0","cwd":"`+otherDir+`"}`+"\n"), 0o600); err != nil {
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
