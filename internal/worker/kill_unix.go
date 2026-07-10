//go:build unix

package worker

import (
	"os/exec"
	"syscall"
)

// setProcAttr puts the podman client in its own process group so a cancel can
// signal the whole subtree. Container teardown still relies on `podman rm -f`
// keyed on the deterministic --name (the client process is not the container).
func setProcAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}
