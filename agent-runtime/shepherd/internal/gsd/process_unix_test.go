//go:build unix

package gsd

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestCleanupProcessTreeTerminatesDescendantAfterParentExit(t *testing.T) {
	if os.Getenv("SHEPHERD_EXITED_PARENT_HELPER") == "1" {
		child := exec.Command("/bin/sleep", "30")
		child.Stdout, child.Stderr = os.Stdout, os.Stderr
		if err := child.Start(); err != nil {
			os.Exit(3)
		}
		if err := os.WriteFile(os.Getenv("SHEPHERD_CHILD_PID_FILE"), []byte(strconv.Itoa(child.Process.Pid)), 0o600); err != nil {
			os.Exit(4)
		}
		os.Exit(0)
	}

	pidFile := t.TempDir() + "/child.pid"
	cmd := exec.CommandContext(context.Background(), os.Args[0], "-test.run=TestCleanupProcessTreeTerminatesDescendantAfterParentExit")
	cmd.Env = append(os.Environ(), "SHEPHERD_EXITED_PARENT_HELPER=1", "SHEPHERD_CHILD_PID_FILE="+pidFile)
	configureProcessTree(cmd)
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(pidFile)
	if err != nil {
		t.Fatal(err)
	}
	childPID, err := strconv.Atoi(string(raw))
	if err != nil {
		t.Fatal(err)
	}
	if err := cleanupProcessTree(cmd); err != nil {
		t.Fatal(err)
	}
	waitForPIDExitForTest(t, childPID, "descendant after ordinary parent exit")
}

func TestRunProcessTreeBoundsDetachedInheritedOutput(t *testing.T) {
	if os.Getenv("SHEPHERD_DETACHED_PARENT_HELPER") == "1" {
		child := exec.Command("/bin/sleep", "30")
		child.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
		child.Stdout, child.Stderr = os.Stdout, os.Stderr
		if err := child.Start(); err != nil {
			os.Exit(2)
		}
		if err := os.WriteFile(os.Getenv("SHEPHERD_CHILD_PID_FILE"), []byte(strconv.Itoa(child.Process.Pid)), 0o600); err != nil {
			os.Exit(3)
		}
		os.Exit(0)
	}

	pidFile := t.TempDir() + "/child.pid"
	cmd := exec.CommandContext(context.Background(), os.Args[0], "-test.run=TestRunProcessTreeBoundsDetachedInheritedOutput")
	cmd.Env = append(os.Environ(), "SHEPHERD_DETACHED_PARENT_HELPER=1", "SHEPHERD_CHILD_PID_FILE="+pidFile)
	configureProcessTree(cmd)
	var stdout, stderr strings.Builder
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	started := time.Now()
	err := runProcessTree(cmd)
	if !errors.Is(err, errControllerDrainTimeout) {
		t.Fatalf("detached output drain error=%v", err)
	}
	if elapsed := time.Since(started); elapsed > 3*time.Second {
		t.Fatalf("detached output drain took %s", elapsed)
	}
	raw, readErr := os.ReadFile(pidFile)
	if readErr != nil {
		t.Fatal(readErr)
	}
	pid, parseErr := strconv.Atoi(string(raw))
	if parseErr != nil {
		t.Fatal(parseErr)
	}
	_ = syscall.Kill(pid, syscall.SIGKILL)
}

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
	waitForPIDExitForTest(t, childPID, "descendant after synchronous process-group termination")
}
