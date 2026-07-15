//go:build unix

package recovery

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

func TestPlannerProcessCleanupTerminatesDescendants(t *testing.T) {
	if os.Getenv("GO_WANT_RECOVERY_PROCESS_HELPER") != "" {
		return
	}
	for _, mode := range []string{"cancel", "exit"} {
		t.Run(mode, func(t *testing.T) {
			pidPath := t.TempDir() + "/child.pid"
			ctx, cancel := context.WithCancel(context.Background())
			cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=TestRecoveryProcessHelper", "--")
			cmd.Env = append(os.Environ(), "GO_WANT_RECOVERY_PROCESS_HELPER="+mode, "RECOVERY_CHILD_PID_PATH="+pidPath)
			done := make(chan error, 1)
			go func() {
				_, err := runBounded(cmd, 64*1024, 64*1024)
				done <- err
			}()
			childPID := waitForPID(t, pidPath)
			if mode == "cancel" {
				cancel()
			}
			select {
			case <-done:
			case <-time.After(5 * time.Second):
				t.Fatal("planner process cleanup did not return")
			}
			cancel()
			waitForProcessExit(t, childPID)
		})
	}
}

func TestRecoveryProcessHelper(t *testing.T) {
	mode := os.Getenv("GO_WANT_RECOVERY_PROCESS_HELPER")
	if mode == "" {
		return
	}
	child := exec.Command("/bin/sleep", "30")
	if err := child.Start(); err != nil {
		os.Exit(2)
	}
	if err := os.WriteFile(os.Getenv("RECOVERY_CHILD_PID_PATH"), []byte(strconv.Itoa(child.Process.Pid)), 0o600); err != nil {
		os.Exit(3)
	}
	if mode == "exit" {
		os.Exit(0)
	}
	_ = child.Wait()
	os.Exit(0)
}

func waitForPID(t *testing.T, path string) int {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		raw, err := os.ReadFile(path)
		if err == nil {
			pid, parseErr := strconv.Atoi(string(raw))
			if parseErr == nil && pid > 0 {
				return pid
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("planner descendant did not start")
	return 0
}

func waitForProcessExit(t *testing.T, pid int) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if errors.Is(syscall.Kill(pid, 0), syscall.ESRCH) {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("planner descendant pid %d survived cleanup", pid)
}
