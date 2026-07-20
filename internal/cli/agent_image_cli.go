package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"polymetrics.ai/internal/config"
	"polymetrics.ai/internal/safety"
)

type agentImageRuntime interface {
	LookPath(string) (string, error)
	FileExists(string) error
	Run(context.Context, string, []string, io.Writer, io.Writer) error
}

type osAgentImageRuntime struct{}

func (osAgentImageRuntime) LookPath(bin string) (string, error) {
	return exec.LookPath(bin)
}

func (osAgentImageRuntime) FileExists(path string) error {
	_, err := os.Stat(path)
	return err
}

func (osAgentImageRuntime) Run(ctx context.Context, bin string, args []string, stdout, stderr io.Writer) error {
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd.Run()
}

// runAgentImage dispatches the compatibility fallback for missing or invalid image actions.
func runAgentImage(ctx context.Context, cfg config.Config, root string, args []string, stdout io.Writer, jsonOut bool, runtime agentImageRuntime) error {
	if len(args) == 0 {
		return usageErrorf("agent image: missing subcommand (build|pull|ensure)")
	}
	return runAgentImageAction(ctx, cfg, root, args[0], stdout, jsonOut, runtime)
}

func runAgentImageAction(ctx context.Context, cfg config.Config, root, action string, stdout io.Writer, jsonOut bool, runtime agentImageRuntime) error {
	if runtime == nil {
		runtime = osAgentImageRuntime{}
	}
	if err := validateAgentPodmanBin(cfg.RLM.PodmanBin); err != nil {
		return validationErrorf("%v", err)
	}

	knownAction := action == "build" || action == "pull" || action == "ensure"
	if knownAction {
		if err := validateAgentImageReference(cfg.RLM.Image); err != nil {
			return validationErrorf("%v", err)
		}
		if action == "build" {
			if err := validateAgentBuildRoot(root); err != nil {
				return validationErrorf("%v", err)
			}
		}
	}

	bin := cfg.RLM.PodmanBin
	if _, err := runtime.LookPath(bin); err != nil {
		return fmt.Errorf("agent image: %q not found in PATH; install podman to build/run the RLM agent", bin)
	}
	if !knownAction {
		return usageErrorf("agent image: unknown subcommand %q (want build|pull|ensure)", action)
	}

	image := cfg.RLM.Image
	switch action {
	case "build":
		dir := filepath.Join(root, "build", "agent")
		containerfile := filepath.Join(dir, "Containerfile")
		if err := runtime.FileExists(containerfile); err != nil {
			return fmt.Errorf("agent image build: %s/Containerfile not found", dir)
		}
		return runAgentImageRuntime(ctx, runtime, bin, []string{"build", "-f", containerfile, "-t", image, dir}, stdout, jsonOut, "AgentImageBuild", image)
	case "pull":
		return runAgentImageRuntime(ctx, runtime, bin, []string{"pull", image}, stdout, jsonOut, "AgentImagePull", image)
	case "ensure":
		if runtime.Run(ctx, bin, []string{"image", "exists", image}, io.Discard, io.Discard) == nil {
			if jsonOut {
				return writeJSON(stdout, envelope{"kind": "AgentImageEnsure", "image": image, "status": "present"})
			}
			fmt.Fprintf(stdout, "agent image present: %s\n", image)
			return nil
		}
		return runAgentImageRuntime(ctx, runtime, bin, []string{"pull", image}, stdout, jsonOut, "AgentImageEnsure", image)
	default:
		return usageErrorf("agent image: unknown subcommand %q (want build|pull|ensure)", action)
	}
}

func runAgentImageRuntime(ctx context.Context, runtime agentImageRuntime, bin string, args []string, stdout io.Writer, jsonOut bool, kind, image string) error {
	if err := runtime.Run(ctx, bin, args, stdout, stdout); err != nil {
		return fmt.Errorf("agent image: %w", err)
	}
	if jsonOut {
		return writeJSON(stdout, envelope{"kind": kind, "image": image, "status": "ok"})
	}
	fmt.Fprintf(stdout, "%s ok: %s\n", kind, image)
	return nil
}

func validateAgentPodmanBin(bin string) error {
	if strings.TrimSpace(bin) == "" {
		return fmt.Errorf("agent image podman binary is required")
	}
	if err := safety.RejectDangerousChars(bin, "agent image podman binary"); err != nil {
		return err
	}
	if strings.HasPrefix(strings.TrimSpace(bin), "-") {
		return fmt.Errorf("agent image podman binary must not start with '-'")
	}
	return nil
}

func validateAgentBuildRoot(root string) error {
	if strings.TrimSpace(root) == "" {
		return fmt.Errorf("agent image project root is required")
	}
	if err := safety.RejectDangerousChars(root, "agent image project root"); err != nil {
		return err
	}
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return fmt.Errorf("resolve agent image project root: %w", err)
	}
	buildAbs, err := filepath.Abs(filepath.Join(root, "build", "agent"))
	if err != nil {
		return fmt.Errorf("resolve agent image build path: %w", err)
	}
	rel, err := filepath.Rel(rootAbs, buildAbs)
	if err != nil {
		return fmt.Errorf("validate agent image build path: %w", err)
	}
	if rel != filepath.Join("build", "agent") {
		return fmt.Errorf("agent image build path must stay under the project root")
	}
	return nil
}

func validateAgentImageReference(image string) error {
	if strings.TrimSpace(image) == "" {
		return fmt.Errorf("agent image reference is required")
	}
	if image != strings.TrimSpace(image) {
		return fmt.Errorf("agent image reference must not contain surrounding whitespace")
	}
	if err := safety.RejectDangerousChars(image, "agent image reference"); err != nil {
		return err
	}
	if strings.HasPrefix(image, "-") {
		return fmt.Errorf("agent image reference must not start with '-'")
	}
	if strings.Contains(image, "\\") || strings.Contains(image, "//") {
		return fmt.Errorf("agent image reference contains an invalid path separator")
	}
	for _, segment := range strings.Split(image, "/") {
		if segment == "" || segment == "." || segment == ".." {
			return fmt.Errorf("agent image reference contains invalid path traversal")
		}
	}
	for _, r := range image {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case strings.ContainsRune("._-/:@+", r):
		default:
			return fmt.Errorf("agent image reference contains invalid character %q", r)
		}
	}
	if strings.ContainsAny(image, " \t\r\n") || strings.HasSuffix(image, "/") || strings.HasSuffix(image, ":") || strings.HasSuffix(image, "@") {
		return fmt.Errorf("agent image reference is malformed")
	}
	return nil
}
