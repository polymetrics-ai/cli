package gsd

import (
	"errors"
	"os/exec"
	"path/filepath"
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

func qualifiedNodePathForTest(t *testing.T) string {
	t.Helper()
	nodePath, err := exec.LookPath("node")
	if err != nil {
		t.Fatal(err)
	}
	nodePath, err = filepath.Abs(nodePath)
	if err != nil {
		t.Fatal(err)
	}
	nodePath, err = filepath.EvalSymlinks(nodePath)
	if err != nil {
		t.Fatal(err)
	}
	return nodePath
}
