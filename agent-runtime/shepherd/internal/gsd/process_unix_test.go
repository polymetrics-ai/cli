//go:build unix

package gsd

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"testing"
	"time"
)

func TestKillProcessTreeSynchronouslyTerminatesDescendant(t *testing.T) {
	if os.Getenv("SHEPHERD_PROCESS_GROUP_HELPER") == "1" {
		child := exec.Command("/bin/sleep", "30")
		if err := child.Start(); err != nil {
			os.Exit(2)
		}
		if err := os.WriteFile(os.Getenv("SHEPHERD_CHILD_PID_FILE"), []byte(strconv.Itoa(child.Process.Pid)), 0o600); err != nil {
			os.Exit(3)
		}
		_ = child.Wait()
		os.Exit(0)
	}

	pidFile := t.TempDir() + "/child.pid"
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=TestKillProcessTreeSynchronouslyTerminatesDescendant")
	cmd.Env = append(os.Environ(), "SHEPHERD_PROCESS_GROUP_HELPER=1", "SHEPHERD_CHILD_PID_FILE="+pidFile)
	configureProcessTree(cmd)
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	var childPID int
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		raw, err := os.ReadFile(pidFile)
		if err == nil {
			childPID, _ = strconv.Atoi(string(raw))
			if childPID > 0 {
				break
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	if childPID <= 0 {
		_ = cmd.Cancel()
		_ = cmd.Wait()
		t.Fatal("helper descendant did not start")
	}
	_ = cmd.Cancel()
	cancel()
	_ = cmd.Wait()
	deadline = time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		err := syscall.Kill(childPID, 0)
		if errors.Is(err, syscall.ESRCH) {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("descendant pid %d survived synchronous process-group termination", childPID)
}
