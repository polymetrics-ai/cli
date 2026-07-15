//go:build unix

package validation

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

func TestValidationProcessTreeCleansDescendantAfterSuccessfulParentExit(t *testing.T) {
	if os.Getenv("SHEPHERD_VALIDATOR_PROCESS_HELPER") == "1" {
		return
	}
	pidPath := t.TempDir() + "/child.pid"
	cmd := exec.CommandContext(context.Background(), os.Args[0], "-test.run=TestValidationProcessTreeCancellationTerminatesDescendant")
	cmd.Env = append(os.Environ(), "SHEPHERD_VALIDATOR_PROCESS_HELPER=1", "SHEPHERD_VALIDATOR_PARENT_EXIT=1", "SHEPHERD_VALIDATOR_CHILD_PID="+pidPath)
	if _, err := limitedCombinedOutput(cmd, 64*1024); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(pidPath)
	if err != nil {
		t.Fatal(err)
	}
	childPID, err := strconv.Atoi(string(raw))
	if err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if errors.Is(syscall.Kill(childPID, 0), syscall.ESRCH) {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("validator descendant pid %d survived successful parent cleanup", childPID)
}

func TestValidationProcessTreeCancellationTerminatesDescendant(t *testing.T) {
	if os.Getenv("SHEPHERD_VALIDATOR_PROCESS_HELPER") == "1" {
		child := exec.Command("/bin/sleep", "30")
		if err := child.Start(); err != nil {
			os.Exit(2)
		}
		if err := os.WriteFile(os.Getenv("SHEPHERD_VALIDATOR_CHILD_PID"), []byte(strconv.Itoa(child.Process.Pid)), 0o600); err != nil {
			os.Exit(3)
		}
		if os.Getenv("SHEPHERD_VALIDATOR_PARENT_EXIT") == "1" {
			os.Exit(0)
		}
		_ = child.Wait()
		os.Exit(0)
	}

	pidPath := t.TempDir() + "/child.pid"
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=TestValidationProcessTreeCancellationTerminatesDescendant")
	cmd.Env = append(os.Environ(), "SHEPHERD_VALIDATOR_PROCESS_HELPER=1", "SHEPHERD_VALIDATOR_CHILD_PID="+pidPath)
	configureValidationProcessTree(cmd)
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	var childPID int
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		raw, err := os.ReadFile(pidPath)
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
		t.Fatal("validator descendant did not start")
	}
	_ = cmd.Cancel()
	cancel()
	_ = cmd.Wait()
	deadline = time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if errors.Is(syscall.Kill(childPID, 0), syscall.ESRCH) {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("validator descendant pid %d survived process-group cancellation", childPID)
}
