//go:build unix

package validation

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

func TestVerifyArtifactHashesClosesArtifactRootPerItem(t *testing.T) {
	if os.Getenv("GO_WANT_VERIFY_ARTIFACT_CLOSE_HELPER") == "1" {
		verifyArtifactHashesCloseHelper()
		return
	}
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(exe, "-test.run", "^TestVerifyArtifactHashesClosesArtifactRootPerItem$")
	cmd.Env = append(os.Environ(), "GO_WANT_VERIFY_ARTIFACT_CLOSE_HELPER=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("artifact verifier retained per-item roots under bounded file descriptors: %v\n%s", err, out)
	}
}

func verifyArtifactHashesCloseHelper() {
	workDir, err := os.MkdirTemp("", "artifact-close-")
	if err != nil {
		fmt.Fprintf(os.Stderr, "tempdir: %v\n", err)
		os.Exit(2)
	}
	defer func() { _ = os.RemoveAll(workDir) }()
	artifacts := make([]ArtifactHash, 0, 128)
	for i := 0; i < 128; i++ {
		rel := fmt.Sprintf("agent-runtime/shepherd/artifact-%03d/proof.txt", i)
		path := filepath.Join(workDir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			fmt.Fprintf(os.Stderr, "mkdir: %v\n", err)
			os.Exit(2)
		}
		body := []byte(fmt.Sprintf("proof-%03d", i))
		if err := os.WriteFile(path, body, 0o600); err != nil {
			fmt.Fprintf(os.Stderr, "write: %v\n", err)
			os.Exit(2)
		}
		sum := sha256.Sum256(body)
		artifacts = append(artifacts, ArtifactHash{Path: rel, Hash: "sha256:" + hex.EncodeToString(sum[:])})
	}
	var limit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &limit); err != nil {
		fmt.Fprintf(os.Stderr, "getrlimit: %v\n", err)
		os.Exit(2)
	}
	limit.Cur = 64
	if limit.Max > 64 {
		limit.Max = 64
	}
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &limit); err != nil {
		fmt.Fprintf(os.Stderr, "setrlimit: %v\n", err)
		os.Exit(2)
	}
	if err := verifyArtifactHashes(workDir, artifacts); err != nil {
		if !strings.Contains(err.Error(), "too many open files") {
			fmt.Fprintf(os.Stderr, "verify failed: %v\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "verify retained roots: %v\n", err)
		}
		os.Exit(1)
	}
}
