package git

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestArtifactManifestBindsChangedScopedFiles(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	runGitTest(t, root, "init", "-b", "main")
	runGitTest(t, root, "config", "user.email", "test@example.invalid")
	runGitTest(t, root, "config", "user.name", "Test")
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("initial\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	runGitTest(t, root, "add", "README.md")
	runGitTest(t, root, "commit", "-m", "initial")
	start := stringTrim(runGitTest(t, root, "rev-parse", "HEAD"))
	path := filepath.Join(root, "agent-runtime", "shepherd", "proof.txt")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("proof\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	runGitTest(t, root, "add", "agent-runtime/shepherd/proof.txt")
	runGitTest(t, root, "commit", "-m", "proof")
	end := stringTrim(runGitTest(t, root, "rev-parse", "HEAD"))
	manifest, err := ArtifactManifest(ctx, root, start, end, []string{"agent-runtime/shepherd/**"})
	if err != nil {
		t.Fatal(err)
	}
	if len(manifest) != 1 || manifest[0].Path != "agent-runtime/shepherd/proof.txt" || manifest[0].Hash == "" {
		t.Fatalf("manifest=%+v", manifest)
	}
	if _, err := ArtifactManifest(ctx, root, start, end, []string{"docs/**"}); !errors.Is(err, ErrWriteScopeBreach) {
		t.Fatalf("out-of-scope artifact manifest err=%v", err)
	}
}

func runGitTest(t *testing.T, root string, args ...string) []byte {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v: %s", args, err, out)
	}
	return out
}

func stringTrim(raw []byte) string {
	value := string(raw)
	for len(value) > 0 && (value[len(value)-1] == '\n' || value[len(value)-1] == '\r' || value[len(value)-1] == ' ') {
		value = value[:len(value)-1]
	}
	return value
}
