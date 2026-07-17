//go:build !unix

package git

import (
	"os/exec"
	"time"
)

func configureGitProcessTree(cmd *exec.Cmd) {
	cmd.WaitDelay = 5 * time.Second
}

func cleanupGitProcessTree(*exec.Cmd) error { return nil }
