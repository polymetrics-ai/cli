//go:build unix

package gsd

import (
	"os/exec"
	"syscall"
	"time"
)

func configureProcessTree(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.WaitDelay = 5 * time.Second
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return nil
		}
		// Kill the still-owned process group synchronously. Delayed goroutines can
		// target a recycled PGID after Wait returns and are therefore forbidden.
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}
}
