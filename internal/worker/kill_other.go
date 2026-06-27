//go:build !unix

package worker

import "os/exec"

// setProcAttr is a no-op on non-unix platforms; process-group kill is not
// available, so cancellation relies solely on `podman rm -f <name>`.
func setProcAttr(cmd *exec.Cmd) {}

// killSupported reports whether process-group kill is available on this OS.
func killSupported() bool { return false }
