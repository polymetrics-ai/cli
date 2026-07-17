package git

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	if mode := os.Getenv("SHEPHERD_GIT_HELPER"); mode != "" {
		fakeGitMain(mode)
		return
	}
	os.Exit(m.Run())
}

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

func TestArtifactManifestRecordsDeletionAndRenameWithoutRenames(t *testing.T) {
	ctx := context.Background()
	root := initializedArtifactRepository(t)
	writeArtifactFile(t, root, "agent-runtime/shepherd/deleted.txt", "delete me\n")
	runGitTest(t, root, "add", "agent-runtime/shepherd/deleted.txt")
	runGitTest(t, root, "commit", "-m", "tracked artifact")
	start := stringTrim(runGitTest(t, root, "rev-parse", "HEAD"))

	runGitTest(t, root, "mv", "agent-runtime/shepherd/deleted.txt", "agent-runtime/shepherd/renamed.txt")
	runGitTest(t, root, "commit", "-m", "rename artifact")
	end := stringTrim(runGitTest(t, root, "rev-parse", "HEAD"))

	manifest, err := ArtifactManifest(ctx, root, start, end, []string{"agent-runtime/shepherd/**"})
	if err != nil {
		t.Fatal(err)
	}
	if len(manifest) != 2 {
		t.Fatalf("manifest=%+v", manifest)
	}
	if manifest[0].Path != "agent-runtime/shepherd/deleted.txt" || !manifest[0].Deleted || manifest[0].Hash != DeletionSentinelHash {
		t.Fatalf("deleted artifact=%+v", manifest[0])
	}
	if manifest[1].Path != "agent-runtime/shepherd/renamed.txt" || manifest[1].Deleted || manifest[1].Hash == "" || manifest[1].Hash == DeletionSentinelHash {
		t.Fatalf("present artifact=%+v", manifest[1])
	}
}

func TestArtifactManifestOutOfScopeDeletionFailsClosed(t *testing.T) {
	ctx := context.Background()
	root := initializedArtifactRepository(t)
	writeArtifactFile(t, root, "docs/secret.md", "delete outside\n")
	runGitTest(t, root, "add", "docs/secret.md")
	runGitTest(t, root, "commit", "-m", "outside artifact")
	start := stringTrim(runGitTest(t, root, "rev-parse", "HEAD"))
	if err := os.Remove(filepath.Join(root, "docs", "secret.md")); err != nil {
		t.Fatal(err)
	}
	runGitTest(t, root, "add", "-A", "docs/secret.md")
	runGitTest(t, root, "commit", "-m", "delete outside artifact")
	end := stringTrim(runGitTest(t, root, "rev-parse", "HEAD"))

	if _, err := ArtifactManifest(ctx, root, start, end, []string{"agent-runtime/shepherd/**"}); !errors.Is(err, ErrWriteScopeBreach) {
		t.Fatalf("out-of-scope deletion err=%v", err)
	}
}

func TestArtifactManifestRejectsUnknownAndMalformedStatus(t *testing.T) {
	for _, mode := range []string{"unknown-status", "malformed-status", "leading-terminator", "interior-terminator", "extra-terminator", "missing-final-terminator"} {
		t.Run(mode, func(t *testing.T) {
			root := t.TempDir()
			installFakeGit(t, root, mode)
			manifest, err := ArtifactManifest(context.Background(), root, strings.Repeat("a", 40), strings.Repeat("b", 40), []string{"agent-runtime/shepherd/**"})
			if err == nil || errors.Is(err, ErrWriteScopeBreach) || errors.Is(err, ErrOutputLimit) {
				t.Fatalf("status err=%v manifest=%+v", err, manifest)
			}
		})
	}
}

func TestArtifactManifestParsesAllStatusesBeforeHashing(t *testing.T) {
	root := t.TempDir()
	marker := filepath.Join(root, "cat-file-called")
	installFakeGit(t, root, "129-records")
	t.Setenv("SHEPHERD_GIT_CATFILE_MARKER", marker)
	manifest, err := ArtifactManifest(context.Background(), root, strings.Repeat("a", 40), strings.Repeat("b", 40), []string{"agent-runtime/shepherd/**"})
	if err == nil || !strings.Contains(err.Error(), "artifact") || len(manifest) != 0 {
		t.Fatalf("129 artifact err=%v manifest=%+v", err, manifest)
	}
	if _, statErr := os.Stat(marker); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("cat-file was spawned before artifact-count validation: %v", statErr)
	}
}

func TestArtifactManifestGitFailureNeverBecomesDeletion(t *testing.T) {
	root := t.TempDir()
	installFakeGit(t, root, "git-failure")
	manifest, err := ArtifactManifest(context.Background(), root, strings.Repeat("a", 40), strings.Repeat("b", 40), []string{"agent-runtime/shepherd/**"})
	if err == nil || errors.Is(err, ErrWriteScopeBreach) || errors.Is(err, ErrOutputLimit) || len(manifest) != 0 {
		t.Fatalf("git failure err=%v manifest=%+v", err, manifest)
	}
}

func TestArtifactManifestOversizedObjectIsLimitNotDeletion(t *testing.T) {
	ctx := context.Background()
	root := initializedArtifactRepository(t)
	start := stringTrim(runGitTest(t, root, "rev-parse", "HEAD"))
	writeArtifactFile(t, root, "agent-runtime/shepherd/huge.bin", strings.Repeat("x", 8*1024*1024+1))
	runGitTest(t, root, "add", "agent-runtime/shepherd/huge.bin")
	runGitTest(t, root, "commit", "-m", "huge artifact")
	end := stringTrim(runGitTest(t, root, "rev-parse", "HEAD"))

	manifest, err := ArtifactManifest(ctx, root, start, end, []string{"agent-runtime/shepherd/**"})
	if !errors.Is(err, ErrOutputLimit) {
		t.Fatalf("oversized object err=%v manifest=%+v", err, manifest)
	}
}

func TestHashGitObjectChecksDeclaredSizeBeforeStreaming(t *testing.T) {
	root := t.TempDir()
	marker := filepath.Join(root, "cat-file-blob-called")
	installFakeGit(t, root, "cat-size-over-limit")
	t.Setenv("SHEPHERD_GIT_CATFILE_MARKER", marker)
	if hash, err := hashGitObject(context.Background(), root, "HEAD:agent-runtime/shepherd/huge.bin"); !errors.Is(err, ErrOutputLimit) || hash != "" {
		t.Fatalf("oversized declared object hash=%q err=%v", hash, err)
	}
	if _, statErr := os.Stat(marker); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("oversized object was streamed before rejection: %v", statErr)
	}
}

func TestHashGitObjectExactLimitPassesAndSizeMustMatch(t *testing.T) {
	root := t.TempDir()
	installFakeGit(t, root, "cat-exact-limit")
	hash, err := hashGitObject(context.Background(), root, "HEAD:agent-runtime/shepherd/exact.bin")
	if err != nil {
		t.Fatalf("exact object limit err=%v", err)
	}
	expectedSum := sha256.Sum256(bytes.Repeat([]byte("x"), maxGitObjectBytes))
	if expected := "sha256:" + hex.EncodeToString(expectedSum[:]); hash != expected {
		t.Fatalf("exact object hash=%s want %s", hash, expected)
	}

	installFakeGit(t, root, "cat-declared-mismatch")
	if hash, err := hashGitObject(context.Background(), root, "HEAD:agent-runtime/shepherd/truncated.bin"); err == nil || errors.Is(err, ErrOutputLimit) || hash != "" {
		t.Fatalf("declared-size mismatch hash=%q err=%v", hash, err)
	}
}

func TestHashGitObjectClassifiesCatFileFailuresAndOverflow(t *testing.T) {
	root := t.TempDir()
	installFakeGit(t, root, "cat-nonzero")
	if _, err := hashGitObject(context.Background(), root, "HEAD:agent-runtime/shepherd/missing.bin"); err == nil || errors.Is(err, ErrOutputLimit) {
		t.Fatalf("cat-file nonzero err=%v", err)
	}

	installFakeGit(t, root, "cat-stderr-overflow")
	if _, err := hashGitObject(context.Background(), root, "HEAD:agent-runtime/shepherd/noisy.bin"); !errors.Is(err, ErrOutputLimit) {
		t.Fatalf("cat-file stderr overflow err=%v", err)
	}

	installFakeGit(t, root, "cat-stream-overflow")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if _, err := hashGitObject(ctx, root, "HEAD:agent-runtime/shepherd/endless.bin"); !errors.Is(err, ErrOutputLimit) {
		t.Fatalf("cat-file stream overflow err=%v", err)
	}
}

func TestRunBoundsOutputAndPreservesCancellationIdentity(t *testing.T) {
	root := t.TempDir()
	installFakeGit(t, root, "stdout-over-limit")
	if _, err := run(context.Background(), root, "status"); !errors.Is(err, ErrOutputLimit) {
		t.Fatalf("stdout limit err=%v", err)
	}

	installFakeGit(t, root, "stderr-over-limit")
	if _, err := run(context.Background(), root, "status"); !errors.Is(err, ErrOutputLimit) || strings.Contains(err.Error(), "\x1b") || len(err.Error()) > 4096 {
		t.Fatalf("stderr limit err=%q", err)
	}

	installFakeGit(t, root, "exact-limit")
	out, err := run(context.Background(), root, "status")
	if err != nil {
		t.Fatalf("exact limit err=%v", err)
	}
	if len(out) != maxGitStdoutBytes {
		t.Fatalf("exact output len=%d want %d", len(out), maxGitStdoutBytes)
	}

	installFakeGit(t, root, "stderr-exact-limit")
	out, err = run(context.Background(), root, "status")
	if err != nil || string(out) != "ok" {
		t.Fatalf("exact stderr limit out=%q err=%v", out, err)
	}

	installFakeGit(t, root, "sleep")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := run(ctx, root, "status"); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation err=%v", err)
	}
}

func TestRunUsesSanitizedGitEnvironment(t *testing.T) {
	root := t.TempDir()
	installFakeGit(t, root, "env")
	t.Setenv("GIT_DIR", "/tmp/forbidden")
	t.Setenv("GIT_CONFIG_GLOBAL", "/tmp/forbidden")
	out, err := run(context.Background(), root, "status")
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(out, []byte("GIT_DIR=")) || bytes.Contains(out, []byte("GIT_CONFIG_GLOBAL=")) ||
		!bytes.Contains(out, []byte("GIT_TERMINAL_PROMPT=0")) || !bytes.Contains(out, []byte("GIT_ASKPASS=")) {
		t.Fatalf("unsanitized environment:\n%s", out)
	}
}

func initializedArtifactRepository(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	runGitTest(t, root, "init", "-b", "main")
	runGitTest(t, root, "config", "user.email", "test@example.invalid")
	runGitTest(t, root, "config", "user.name", "Test")
	writeArtifactFile(t, root, "README.md", "initial\n")
	runGitTest(t, root, "add", "README.md")
	runGitTest(t, root, "commit", "-m", "initial")
	return root
}

func writeArtifactFile(t *testing.T, root, rel, body string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

func installFakeGit(t *testing.T, root, mode string) {
	t.Helper()
	bin := filepath.Join(root, "bin-"+mode)
	if err := os.MkdirAll(bin, 0o700); err != nil {
		t.Fatal(err)
	}
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	fake := filepath.Join(bin, "git")
	if err := os.Symlink(exe, fake); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SHEPHERD_GIT_HELPER", mode)
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func fakeGitMain(mode string) {
	args := os.Args[1:]
	isCatSize := hasArgSequence(args, "cat-file", "-s")
	isCatBlob := hasArgSequence(args, "cat-file", "blob")
	switch mode {
	case "stdout-over-limit":
		fmt.Print(strings.Repeat("o", maxGitStdoutBytes+1))
	case "stderr-over-limit":
		_, _ = fmt.Fprint(os.Stderr, "\x1b[31m"+strings.Repeat("e", maxGitStderrBytes+1))
	case "exact-limit":
		fmt.Print(strings.Repeat("x", maxGitStdoutBytes))
	case "stderr-exact-limit":
		_, _ = fmt.Fprint(os.Stderr, strings.Repeat("e", maxGitStderrBytes))
		fmt.Print("ok")
	case "sleep":
		time.Sleep(30 * time.Second)
	case "run-endless-descendant":
		child := exec.Command("/bin/sleep", "30")
		if err := child.Start(); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		if path := os.Getenv("SHEPHERD_GIT_CHILD_PID"); path != "" {
			_ = os.WriteFile(path, []byte(strconv.Itoa(child.Process.Pid)), 0o600)
		}
		for {
			fmt.Print(strings.Repeat("o", 64*1024))
		}
	case "env":
		for _, entry := range os.Environ() {
			fmt.Println(entry)
		}
	case "unknown-status":
		_, _ = os.Stdout.Write([]byte{'X', 0, 'a', 'g', 'e', 'n', 't', '-', 'r', 'u', 'n', 't', 'i', 'm', 'e', '/', 's', 'h', 'e', 'p', 'h', 'e', 'r', 'd', '/', 'x', 0})
	case "malformed-status":
		_, _ = os.Stdout.Write([]byte{'A', 0})
	case "leading-terminator":
		_, _ = os.Stdout.Write([]byte{0, 'A', 0, 'a', 'g', 'e', 'n', 't', '-', 'r', 'u', 'n', 't', 'i', 'm', 'e', '/', 's', 'h', 'e', 'p', 'h', 'e', 'r', 'd', '/', 'x', 0})
	case "interior-terminator":
		_, _ = os.Stdout.Write([]byte{'A', 0, 'a', 'g', 'e', 'n', 't', '-', 'r', 'u', 'n', 't', 'i', 'm', 'e', '/', 's', 'h', 'e', 'p', 'h', 'e', 'r', 'd', '/', 'x', 0, 0, 'D', 0, 'a', 'g', 'e', 'n', 't', '-', 'r', 'u', 'n', 't', 'i', 'm', 'e', '/', 's', 'h', 'e', 'p', 'h', 'e', 'r', 'd', '/', 'y', 0})
	case "extra-terminator":
		_, _ = os.Stdout.Write([]byte{'A', 0, 'a', 'g', 'e', 'n', 't', '-', 'r', 'u', 'n', 't', 'i', 'm', 'e', '/', 's', 'h', 'e', 'p', 'h', 'e', 'r', 'd', '/', 'x', 0, 0})
	case "missing-final-terminator":
		_, _ = os.Stdout.Write([]byte{'A', 0, 'a', 'g', 'e', 'n', 't', '-', 'r', 'u', 'n', 't', 'i', 'm', 'e', '/', 's', 'h', 'e', 'p', 'h', 'e', 'r', 'd', '/', 'x'})
	case "129-records":
		if isCatBlob {
			if marker := os.Getenv("SHEPHERD_GIT_CATFILE_MARKER"); marker != "" {
				_ = os.WriteFile(marker, []byte("called"), 0o600)
			}
			fmt.Print("x")
			return
		}
		for i := 0; i < 129; i++ {
			fmt.Printf("A%cagent-runtime/shepherd/%03d.txt%c", 0, i, 0)
		}
	case "cat-size-over-limit":
		if isCatSize {
			fmt.Println(maxGitObjectBytes + 1)
			return
		}
		if isCatBlob {
			if marker := os.Getenv("SHEPHERD_GIT_CATFILE_MARKER"); marker != "" {
				_ = os.WriteFile(marker, []byte("called"), 0o600)
			}
			fmt.Print("small")
			return
		}
	case "cat-exact-limit":
		if isCatSize {
			fmt.Println(maxGitObjectBytes)
			return
		}
		if isCatBlob {
			_, _ = os.Stdout.Write(bytes.Repeat([]byte("x"), maxGitObjectBytes))
			return
		}
	case "cat-declared-mismatch":
		if isCatSize {
			fmt.Println(4)
			return
		}
		if isCatBlob {
			fmt.Print("abc")
			return
		}
	case "cat-nonzero":
		if isCatSize {
			fmt.Println(1)
			return
		}
		if isCatBlob {
			_, _ = fmt.Fprint(os.Stderr, "fatal: missing object")
			os.Exit(42)
		}
	case "cat-stderr-overflow":
		if isCatSize {
			fmt.Println(1)
			return
		}
		if isCatBlob {
			_, _ = fmt.Fprint(os.Stderr, strings.Repeat("e", maxGitStderrBytes+1))
			fmt.Print("x")
			return
		}
	case "cat-stream-overflow":
		if isCatSize {
			fmt.Println(maxGitObjectBytes)
			return
		}
		if isCatBlob {
			for {
				_, _ = os.Stdout.Write(bytes.Repeat([]byte("x"), 64*1024))
			}
		}
	case "git-failure":
		_, _ = fmt.Fprint(os.Stderr, "\x1b[31mfatal: no object\x00\n")
		os.Exit(42)
	default:
		_, _ = fmt.Fprintln(os.Stderr, "unknown fake git mode")
		os.Exit(2)
	}
}

func hasArgSequence(args []string, sequence ...string) bool {
	for i := 0; i+len(sequence) <= len(args); i++ {
		matched := true
		for j := range sequence {
			if args[i+j] != sequence[j] {
				matched = false
				break
			}
		}
		if matched {
			return true
		}
	}
	return false
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
