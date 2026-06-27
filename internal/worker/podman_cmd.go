package worker

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"

	"polymetrics.ai/internal/rlm"
)

// defaultNewCmd builds the hardened, default-deny `podman run` command for one
// job. It is the production NewCmd; tests inject a fake.
//
// Security posture (see plan §Security):
//   - --network=none by default (egress is reached via a host allowlist proxy,
//     not raw container egress)
//   - --read-only rootfs, --cap-drop=ALL, --security-opt=no-new-privileges
//   - memory / pid limits; ephemeral --rm; deterministic --name for reaping
//   - input mounted read-only; output writable; LLM creds via an env-file (0600)
//     written by the caller, never via -e from the daemon's own environment.
func (p *PodmanActivities) defaultNewCmd(ctx context.Context, req rlm.AgentRequest, name string) *exec.Cmd {
	inDir := filepath.Join(req.JobDir, "in")
	outDir := filepath.Join(req.JobDir, "out")

	args := []string{
		"run", "--rm", "--name", name,
		"--network=none",
		"--read-only",
		"--cap-drop=ALL",
		"--security-opt=no-new-privileges",
		"--memory=2g",
		"--pids-limit=512",
		"-v", inDir + ":/work/in:ro",
		"-v", outDir + ":/work/out",
		"-e", "PM_RLM_MAXITER=" + itoa(req.MaxIter),
	}
	// Forward only the named LLM env vars that are actually set. Values are read
	// from this process's environment; for the long-lived daemon these should be
	// provided via an env-file rather than resident env (see HUMAN GATE 4).
	for _, k := range p.EnvPass {
		if v, ok := os.LookupEnv(k); ok && v != "" {
			args = append(args, "-e", k+"="+v)
		}
	}
	image := req.Image
	if image == "" {
		image = p.Image
	}
	args = append(args, image)

	cmd := exec.CommandContext(ctx, p.podmanBin(), args...)
	setProcAttr(cmd) // process-group on unix so the client subtree can be signalled
	return cmd
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
