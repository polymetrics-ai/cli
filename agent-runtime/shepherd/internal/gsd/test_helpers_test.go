package gsd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"testing"
	"time"
)

func waitForPIDExitForTest(t *testing.T, pid int, description string) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	var lastErr error
	for time.Now().Before(deadline) {
		lastErr = syscall.Kill(pid, 0)
		if errors.Is(lastErr, syscall.ESRCH) {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("%s pid %d survived cleanup: last probe error=%v", description, pid, lastErr)
}

const testOwnedNodeMaxBytes int64 = 256 * 1024 * 1024

var (
	testOwnedNodeOnce sync.Once
	testOwnedNodeMu   sync.Mutex
	testOwnedNodeDir  string
	testOwnedNodePath string
	testOwnedNodeErr  error
)

func TestMain(m *testing.M) {
	code := m.Run()
	if err := cleanupTestOwnedNodeFixture(); err != nil {
		fmt.Fprintf(os.Stderr, "test-owned Node fixture cleanup failed: %v\n", err)
		code = 1
	}
	os.Exit(code)
}

func qualifiedNodePathForTest(t *testing.T) string {
	t.Helper()
	testOwnedNodeOnce.Do(func() {
		path, err := setupTestOwnedNodeFixture()
		testOwnedNodeMu.Lock()
		testOwnedNodePath = path
		testOwnedNodeErr = err
		testOwnedNodeMu.Unlock()
	})
	testOwnedNodeMu.Lock()
	path, err := testOwnedNodePath, testOwnedNodeErr
	testOwnedNodeMu.Unlock()
	if err != nil {
		t.Fatalf("prepare test-owned Node fixture: %v", err)
	}
	return path
}

func setupTestOwnedNodeFixture() (string, error) {
	source, err := exec.LookPath("node")
	if err != nil {
		return "", fmt.Errorf("locate Node executable: %w", err)
	}
	source, err = filepath.Abs(source)
	if err != nil {
		return "", fmt.Errorf("resolve absolute Node executable: %w", err)
	}
	source, err = filepath.EvalSymlinks(source)
	if err != nil {
		return "", fmt.Errorf("resolve canonical Node executable: %w", err)
	}
	dir, err := os.MkdirTemp("", "shepherd-gsd-node-*")
	if err != nil {
		return "", fmt.Errorf("create test-owned Node fixture directory: %w", err)
	}
	recordTestOwnedNodeDir(dir)
	if err := os.Chmod(dir, 0o700); err != nil {
		return "", fmt.Errorf("secure test-owned Node fixture directory: %w", err)
	}
	dirInfo, err := os.Lstat(dir)
	if err != nil || !dirInfo.IsDir() || dirInfo.Mode().Perm() != 0o700 || !runtimePathOwnedByCurrentUser(dirInfo) {
		return "", fmt.Errorf("test-owned Node fixture directory is not private")
	}

	sourceFile, err := os.Open(source)
	if err != nil {
		return "", fmt.Errorf("open Node executable source: %w", err)
	}
	sourceInfo, err := sourceFile.Stat()
	if err != nil {
		_ = sourceFile.Close()
		return "", fmt.Errorf("inspect Node executable source: %w", err)
	}
	if !sourceInfo.Mode().IsRegular() {
		_ = sourceFile.Close()
		return "", fmt.Errorf("Node executable source is not a regular file")
	}

	destination := filepath.Join(dir, "node")
	destinationFile, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o400)
	if err != nil {
		_ = sourceFile.Close()
		return "", fmt.Errorf("create test-owned Node executable: %w", err)
	}
	written, copyErr := io.Copy(destinationFile, io.LimitReader(sourceFile, testOwnedNodeMaxBytes+1))
	sourceCloseErr := sourceFile.Close()
	syncErr := error(nil)
	if copyErr == nil && written > 0 && written <= testOwnedNodeMaxBytes {
		syncErr = destinationFile.Sync()
	}
	closeErr := destinationFile.Close()
	if err := errors.Join(copyErr, sourceCloseErr, closeErr); err != nil {
		return "", fmt.Errorf("copy test-owned Node executable: %w", err)
	}
	switch {
	case written == 0:
		return "", fmt.Errorf("Node executable source is empty")
	case written > testOwnedNodeMaxBytes:
		return "", fmt.Errorf("Node executable source exceeds maximum size")
	case syncErr != nil:
		return "", fmt.Errorf("finalize test-owned Node executable: %w", syncErr)
	}
	if err := os.Chmod(destination, 0o500); err != nil {
		return "", fmt.Errorf("make test-owned Node executable owner-executable: %w", err)
	}
	info, err := os.Lstat(destination)
	if err != nil || info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() || info.Mode().Perm() != 0o500 || info.Size() != written || !runtimePathOwnedByCurrentUser(info) {
		return "", fmt.Errorf("test-owned Node executable is not qualified")
	}
	return destination, nil
}

func recordTestOwnedNodeDir(dir string) {
	testOwnedNodeMu.Lock()
	testOwnedNodeDir = dir
	testOwnedNodeMu.Unlock()
}

func cleanupTestOwnedNodeFixture() error {
	testOwnedNodeMu.Lock()
	dir := testOwnedNodeDir
	testOwnedNodeMu.Unlock()
	if dir == "" {
		return nil
	}
	if filepath.Clean(dir) != dir || dir == string(filepath.Separator) {
		return fmt.Errorf("refusing to remove invalid test-owned Node fixture directory")
	}
	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("remove exact test-owned Node fixture directory: %w", err)
	}
	return nil
}
