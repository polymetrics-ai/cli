//go:build !unix

package gsd

import "os/exec"

func configureProcessTree(cmd *exec.Cmd) {}
