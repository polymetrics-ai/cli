//go:build unix

package git

import (
	"errors"
	"os/exec"
	"syscall"
	"time"
)

func configureGitProcessTree(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.WaitDelay = 5 * time.Second
	cmd.Cancel = func() error { return signalGitProcessTree(cmd) }
}

func signalGitProcessTree(cmd *exec.Cmd) error {
	if cmd.Process == nil {
		return nil
	}
	err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	if errors.Is(err, syscall.ESRCH) {
		return nil
	}
	return err
}

func cleanupGitProcessTree(cmd *exec.Cmd) error {
	if err := signalGitProcessTree(cmd); err != nil {
		return err
	}
	if cmd.Process == nil {
		return nil
	}
	processGroup := -cmd.Process.Pid
	deadline := time.Now().Add(cmd.WaitDelay)
	for time.Now().Before(deadline) {
		probeErr := syscall.Kill(processGroup, 0)
		if errors.Is(probeErr, syscall.ESRCH) {
			return nil
		}
		if probeErr != nil {
			return probeErr
		}
		time.Sleep(10 * time.Millisecond)
	}
	return errors.New("timed out cleaning git process group")
}
