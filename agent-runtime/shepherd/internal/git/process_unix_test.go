//go:build unix

package git

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"testing"
	"time"
)

func TestRunOutputLimitTerminatesProcessGroupDescendants(t *testing.T) {
	root := t.TempDir()
	pidPath := filepath.Join(root, "child.pid")
	installFakeGit(t, root, "run-endless-descendant")
	t.Setenv("SHEPHERD_GIT_CHILD_PID", pidPath)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if _, err := run(ctx, root, "status"); !errors.Is(err, ErrOutputLimit) {
		t.Fatalf("endless output err=%v", err)
	}
	childPID := readPIDForGitTest(t, pidPath)
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if errors.Is(syscall.Kill(childPID, 0), syscall.ESRCH) {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("git descendant pid %d survived output-limit cleanup", childPID)
}

func readPIDForGitTest(t *testing.T, path string) int {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		raw, err := os.ReadFile(path)
		if err == nil {
			pid, convErr := strconv.Atoi(string(raw))
			if convErr != nil || pid <= 0 {
				t.Fatalf("invalid pid %q", raw)
			}
			return pid
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("pid file %s was not written", path)
	return 0
}
