package gsd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ContainerConfig struct {
	Engine       string
	Image        string
	GSDStateDir  string
	PlanningDir  string
	AuthFile     string
	SettingsFile string
}

func (c ContainerConfig) Validate(workDir string) error {
	if c.Engine != "podman" || c.Image == "" || strings.ContainsAny(c.Image, "\r\n\x00") {
		return errors.New("container runtime requires podman and a pinned image")
	}
	for label, path := range map[string]string{
		"work directory": workDir, "GSD state": c.GSDStateDir, "planning state": c.PlanningDir,
		"auth file": c.AuthFile, "settings file": c.SettingsFile,
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
	return append([]string{
		"run", "--rm", "--pull=never", "--network=bridge", "--userns=keep-id",
		"--workdir=/workspace",
		"--volume=" + workDir + ":/workspace:rw",
		"--volume=" + c.GSDStateDir + ":/workspace/.gsd:rw",
		"--volume=" + c.PlanningDir + ":/workspace/.planning:rw",
		"--volume=" + c.AuthFile + ":/home/shepherd/.pi/agent/auth.json:ro",
		"--volume=" + c.SettingsFile + ":/home/shepherd/.pi/agent/settings.json:ro",
		"--env=HOME=/home/shepherd", "--env=GSD_HOME=/home/shepherd/.pi",
		"--env=GIT_TERMINAL_PROMPT=0", "--env=GIT_ASKPASS=",
		"--env=GIT_CONFIG_COUNT=2", "--env=GIT_CONFIG_KEY_0=credential.helper", "--env=GIT_CONFIG_VALUE_0=",
		"--env=GIT_CONFIG_KEY_1=remote.origin.pushurl", "--env=GIT_CONFIG_VALUE_1=file:///dev/null/shepherd-disabled",
		c.Image,
	}, gsdArgs...)
}

func provisionContainerPolicy(workDir, stateDir string) error {
	files := []string{"PREFERENCES.md"}
	entries, err := os.ReadDir(filepath.Join(workDir, ".gsd", "agents"))
	if err != nil {
		return fmt.Errorf("read governed GSD agents: %w", err)
	}
	for _, entry := range entries {
		if entry.Type().IsRegular() && strings.HasSuffix(entry.Name(), ".md") {
			files = append(files, filepath.Join("agents", entry.Name()))
		}
	}
	for _, relative := range files {
		source := filepath.Join(workDir, ".gsd", relative)
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
	return nil
}

func pathInside(root, path string) (bool, error) {
	relative, err := filepath.Rel(filepath.Clean(root), filepath.Clean(path))
	if err != nil {
		return false, err
	}
	return relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)), nil
}
