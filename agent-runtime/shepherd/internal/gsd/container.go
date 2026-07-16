package gsd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const immutableImagePrefix = "sha256:"

type ContainerConfig struct {
	Engine        string
	Image         string
	GSDStateDir   string
	PlanningDir   string
	AuthFile      string
	SettingsFile  string
	Network       string
	PolicyDir     string
	GitCommonDir  string
	SessionsDir   string
	BackgroundDir string
	BackupDir     string
}

// ResolvePinnedContainerImage resolves a configured local tag exactly once and
// returns the immutable image ID used for admission, export, and execution.
func ResolvePinnedContainerImage(ctx context.Context, config ContainerConfig) (ContainerConfig, error) {
	if ctx == nil || config.Engine != "podman" || config.Image == "" || strings.ContainsAny(config.Image, "\r\n\x00") {
		return ContainerConfig{}, errors.New("container runtime requires a safe Podman image reference")
	}
	inspectCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(inspectCtx, config.Engine, "image", "inspect", "--format={{.Id}}", config.Image)
	configureProcessTree(cmd)
	var stdout, stderr boundedBuffer
	stdout.limit, stderr.limit = 4096, 4096
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := runProcessTree(cmd); err != nil {
		return ContainerConfig{}, fmt.Errorf("inspect immutable container image: %w", err)
	}
	if stdout.Truncated() || stderr.Truncated() {
		return ContainerConfig{}, errors.New("immutable container image inspection output is oversized")
	}
	imageID := strings.TrimSpace(stdout.String())
	if !validImmutableImageID(imageID) {
		return ContainerConfig{}, errors.New("podman returned an invalid immutable image ID")
	}
	config.Image = imageID
	return config, nil
}

func validImmutableImageID(value string) bool {
	if len(value) != len(immutableImagePrefix)+64 || !strings.HasPrefix(value, immutableImagePrefix) {
		return false
	}
	for _, char := range value[len(immutableImagePrefix):] {
		if (char < '0' || char > '9') && (char < 'a' || char > 'f') {
			return false
		}
	}
	return true
}

func (c ContainerConfig) Validate(workDir string) error {
	if c.Engine != "podman" || c.Image == "" || strings.ContainsAny(c.Image, "\r\n\x00") {
		return errors.New("container runtime requires podman and a pinned image")
	}
	if c.Network == "" {
		c.Network = "bridge"
	}
	if !validNetworkName(c.Network) {
		return errors.New("container network must be a plain Podman network name")
	}
	for label, path := range map[string]string{
		"work directory": workDir, "GSD state": c.GSDStateDir, "planning state": c.PlanningDir,
		"auth file": c.AuthFile, "settings file": c.SettingsFile, "governed policy": c.PolicyDir,
		"git common directory": c.GitCommonDir,
		"session archive":      c.SessionsDir, "background shell state": c.BackgroundDir,
		"GSD backup state": c.BackupDir,
	} {
		if !filepath.IsAbs(path) {
			return fmt.Errorf("%s must be absolute", label)
		}
	}
	if within, _ := pathInside(workDir, c.GSDStateDir); within {
		return errors.New("container GSD state must be outside the worktree")
	}
	if within, _ := pathInside(workDir, c.PlanningDir); within {
		return errors.New("container planning state must be outside the worktree")
	}
	if within, _ := pathInside(workDir, c.PolicyDir); within {
		return errors.New("governed policy directory must be outside the worker-controlled worktree")
	}
	for _, path := range []string{c.SessionsDir, c.BackgroundDir, c.BackupDir} {
		if within, _ := pathInside(workDir, path); within {
			return errors.New("runtime archive directories must be outside the worker-controlled worktree")
		}
	}
	if info, err := os.Stat(c.PolicyDir); err != nil || !info.IsDir() {
		return errors.New("governed policy directory must be an existing directory")
	}
	if info, err := os.Stat(c.GitCommonDir); err != nil || !info.IsDir() {
		return errors.New("git common directory must be an existing directory")
	}
	for _, path := range []string{c.AuthFile, c.SettingsFile} {
		resolved, err := filepath.EvalSymlinks(path)
		if err != nil {
			return errors.New("container credential/config mount cannot be resolved")
		}
		info, err := os.Stat(resolved)
		if err != nil || !info.Mode().IsRegular() {
			return errors.New("container credential/config mount must be a regular file")
		}
	}
	return nil
}

func (c ContainerConfig) commandArgs(workDir string, gsdArgs []string) []string {
	network := c.Network
	if network == "" {
		network = "bridge"
	}
	args := []string{
		"run", "--rm", "--interactive", "--pull=never", "--network=" + network, "--userns=keep-id",
		"--entrypoint=/bin/sh",
		"--workdir=" + workDir,
		"--volume=" + workDir + ":" + workDir + ":rw",
		"--volume=" + c.GitCommonDir + ":" + c.GitCommonDir + ":rw",
		"--volume=" + c.GSDStateDir + ":" + filepath.Join(workDir, ".gsd") + ":rw",
		"--volume=" + c.PlanningDir + ":" + filepath.Join(workDir, ".planning") + ":rw",
		"--volume=" + c.BackgroundDir + ":" + filepath.Join(workDir, ".bg-shell") + ":rw",
		"--volume=" + c.BackupDir + ":" + filepath.Join(workDir, ".gsd-backups") + ":rw",
		"--volume=" + c.AuthFile + ":/home/shepherd/.pi/agent/auth.json:ro",
		"--volume=" + c.SettingsFile + ":/home/shepherd/.pi/agent/settings.json:ro",
		"--volume=" + c.SessionsDir + ":/home/shepherd/.pi/agent/sessions:rw",
		"--env=HOME=/home/shepherd", "--env=GSD_HOME=/home/shepherd/.pi",
		"--env=SEARXNG_BASE=http://searxng:8080",
		"--env=GIT_TERMINAL_PROMPT=0", "--env=GIT_ASKPASS=",
		"--env=GIT_CONFIG_COUNT=5", "--env=GIT_CONFIG_KEY_0=credential.helper", "--env=GIT_CONFIG_VALUE_0=",
		"--env=GIT_CONFIG_KEY_1=remote.origin.pushurl", "--env=GIT_CONFIG_VALUE_1=file:///dev/null/shepherd-disabled",
		"--env=GIT_CONFIG_KEY_2=safe.directory", "--env=GIT_CONFIG_VALUE_2=" + workDir,
		"--env=GIT_CONFIG_KEY_3=user.name", "--env=GIT_CONFIG_VALUE_3=Polymetrics Shepherd",
		"--env=GIT_CONFIG_KEY_4=user.email", "--env=GIT_CONFIG_VALUE_4=shepherd@localhost.invalid",
		c.Image, "-c", containerGSDCommand, "shepherd-gsd",
	}
	return append(args, gsdArgs...)
}

// GSD may publish its terminal milestone-ready notification just before a
// descendant `git worktree add` finishes copying a large repository. Keeping a
// shell as container PID 1 lets that orphaned Git process finish; the bounded
// lock poll prevents Podman from killing it at the instant the Node process
// exits. The original GSD exit code remains authoritative.
const containerGSDCommand = `gsd "$@"
rc=$?
common=$(git rev-parse --path-format=absolute --git-common-dir 2>/dev/null || true)
n=0
while [ -n "$common" ] && [ -n "$(find "$common/worktrees" -maxdepth 2 -name index.lock -print -quit 2>/dev/null)" ] && [ "$n" -lt 60 ]; do
  sleep 1
  n=$((n + 1))
done
exit "$rc"`

func provisionContainerPolicy(policyDir, stateDir string) error {
	files := []string{"PREFERENCES.md"}
	entries, err := os.ReadDir(filepath.Join(policyDir, "agents"))
	if err != nil {
		return fmt.Errorf("read governed GSD agents: %w", err)
	}
	for _, entry := range entries {
		if entry.Type().IsRegular() && strings.HasSuffix(entry.Name(), ".md") {
			files = append(files, filepath.Join("agents", entry.Name()))
		}
	}
	for _, relative := range files {
		source := filepath.Join(policyDir, relative)
		info, err := os.Lstat(source)
		if err != nil || !info.Mode().IsRegular() || info.Size() > 256*1024 {
			return fmt.Errorf("unsafe governed policy file %s", relative)
		}
		raw, err := os.ReadFile(source)
		if err != nil {
			return err
		}
		target := filepath.Join(stateDir, relative)
		if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
			return err
		}
		if err := os.WriteFile(target, raw, 0o600); err != nil {
			return err
		}
	}
	trustedMCP := []byte("{\n  \"mcpServers\": {\n    \"context7\": {\n      \"url\": \"https://mcp.context7.com/mcp\"\n    }\n  }\n}\n")
	if err := os.WriteFile(filepath.Join(stateDir, "mcp.json"), trustedMCP, 0o600); err != nil {
		return err
	}
	return nil
}

func validNetworkName(value string) bool {
	if value == "" || len(value) > 63 || value[0] == '-' {
		return false
	}
	for _, char := range value {
		if (char < 'a' || char > 'z') && (char < 'A' || char > 'Z') && (char < '0' || char > '9') && char != '.' && char != '_' && char != '-' {
			return false
		}
	}
	return true
}

func pathInside(root, path string) (bool, error) {
	relative, err := filepath.Rel(filepath.Clean(root), filepath.Clean(path))
	if err != nil {
		return false, err
	}
	return relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)), nil
}
