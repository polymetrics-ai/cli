//go:build !unix

package validation

import (
	"os/exec"
	"time"
)

func configureValidationProcessTree(cmd *exec.Cmd) {
	cmd.WaitDelay = 5 * time.Second
}

func cleanupValidationProcessTree(*exec.Cmd) error { return nil }
