//go:build !unix

package recovery

import (
	"os/exec"
	"time"
)

func processTreeSupported() bool { return false }

func configureProcessTree(cmd *exec.Cmd) {
	cmd.WaitDelay = 5 * time.Second
}

func cleanupProcessTree(*exec.Cmd) error { return nil }
