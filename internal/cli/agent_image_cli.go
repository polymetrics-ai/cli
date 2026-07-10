package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

// defaultRLMImage is the image reference the RLM agent runs from. Override via
// POLYMETRICS_RLM_IMAGE.
const defaultRLMImage = "ghcr.io/polymetrics/rlm-agent:latest"

func rlmImage() string {
	if v := os.Getenv("POLYMETRICS_RLM_IMAGE"); v != "" {
		return v
	}
	return defaultRLMImage
}

func podmanBin() string {
	if v := os.Getenv("POLYMETRICS_PODMAN_BIN"); v != "" {
		return v
	}
	return "podman"
}

// runAgentImage dispatches `pm agent image build|pull|ensure`. These shell out to
// podman; building/publishing the image is a human-gated operation (Podman
// dependency + image publish). The commands fail clearly when podman is absent.
func runAgentImage(ctx context.Context, root string, args []string, stdout io.Writer, jsonOut bool) error {
	if len(args) == 0 {
		return usageErrorf("agent image: missing subcommand (build|pull|ensure)")
	}
	bin := podmanBin()
	if _, err := exec.LookPath(bin); err != nil {
		return fmt.Errorf("agent image: %q not found in PATH; install podman to build/run the RLM agent", bin)
	}
	image := rlmImage()

	switch args[0] {
	case "build":
		dir := filepath.Join(root, "build", "agent")
		if _, err := os.Stat(filepath.Join(dir, "Containerfile")); err != nil {
			return fmt.Errorf("agent image build: %s/Containerfile not found", dir)
		}
		cmd := exec.CommandContext(ctx, bin, "build", "-f", filepath.Join(dir, "Containerfile"), "-t", image, dir)
		return runPodmanStreaming(cmd, stdout, jsonOut, "AgentImageBuild", image)
	case "pull":
		cmd := exec.CommandContext(ctx, bin, "pull", image)
		return runPodmanStreaming(cmd, stdout, jsonOut, "AgentImagePull", image)
	case "ensure":
		// Present locally already?
		check := exec.CommandContext(ctx, bin, "image", "exists", image)
		if check.Run() == nil {
			if jsonOut {
				return writeJSON(stdout, envelope{"kind": "AgentImageEnsure", "image": image, "status": "present"})
			}
			_, err := fmt.Fprintf(stdout, "agent image present: %s\n", image)
			return err
		}
		cmd := exec.CommandContext(ctx, bin, "pull", image)
		return runPodmanStreaming(cmd, stdout, jsonOut, "AgentImageEnsure", image)
	default:
		return usageErrorf("agent image: unknown subcommand %q (want build|pull|ensure)", args[0])
	}
}

func runPodmanStreaming(cmd *exec.Cmd, stdout io.Writer, jsonOut bool, kind, image string) error {
	cmd.Stdout = stdout
	cmd.Stderr = stdout
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("agent image: %w", err)
	}
	if jsonOut {
		return writeJSON(stdout, envelope{"kind": kind, "image": image, "status": "ok"})
	}
	_, err := fmt.Fprintf(stdout, "%s ok: %s\n", kind, image)
	return err
}
