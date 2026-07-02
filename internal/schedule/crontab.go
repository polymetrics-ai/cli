package schedule

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// CrontabBackend installs schedules by appending to the user's crontab.
type CrontabBackend struct{}

func (b CrontabBackend) Kind() BackendKind { return KindCrontab }

func (b CrontabBackend) Install(ctx context.Context, m Manifest, pmBin string) error {
	line, err := renderCrontabLine(m, pmBin)
	if err != nil {
		return err
	}

	// PM_CRONTAB_FILE is set in tests to redirect writes to a temp file.
	if crontabFile := os.Getenv("PM_CRONTAB_FILE"); crontabFile != "" {
		existing, _ := os.ReadFile(crontabFile)
		content := string(existing)
		content, err = removeCrontabLine(content, m.Name)
		if err != nil {
			return err
		}
		if !strings.HasSuffix(content, "\n") && content != "" {
			content += "\n"
		}
		content += line + "\n"
		return os.WriteFile(crontabFile, []byte(content), 0o644)
	}

	if _, err := exec.LookPath("crontab"); err != nil {
		return fmt.Errorf("crontab not found: %w", err)
	}

	// Read existing crontab
	existing := ""
	out, _ := exec.CommandContext(ctx, "crontab", "-l").Output()
	existing = string(out)

	// Remove any existing line for this schedule (idempotent)
	existing, err = removeCrontabLine(existing, m.Name)
	if err != nil {
		return err
	}

	// Append new line
	if !strings.HasSuffix(existing, "\n") && existing != "" {
		existing += "\n"
	}
	existing += line + "\n"

	cmd := exec.CommandContext(ctx, "crontab", "-")
	cmd.Stdin = strings.NewReader(existing)
	return cmd.Run()
}

func (b CrontabBackend) Remove(ctx context.Context, name string) error {
	if crontabFile := os.Getenv("PM_CRONTAB_FILE"); crontabFile != "" {
		existing, _ := os.ReadFile(crontabFile)
		updated, err := removeCrontabLine(string(existing), name)
		if err != nil {
			return err
		}
		return os.WriteFile(crontabFile, []byte(updated), 0o644)
	}
	if _, err := exec.LookPath("crontab"); err != nil {
		return fmt.Errorf("crontab not found: %w", err)
	}
	out, _ := exec.CommandContext(ctx, "crontab", "-l").Output()
	updated, err := removeCrontabLine(string(out), name)
	if err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, "crontab", "-")
	cmd.Stdin = strings.NewReader(updated)
	return cmd.Run()
}

// renderCrontabLine returns the crontab line (with sentinel comment) for the manifest.
// Pure function.
func renderCrontabLine(m Manifest, pmBin string) (string, error) {
	sentinel := "# pm-schedule-" + m.Name
	rootArgs := ""
	if m.Root != "" {
		rootArgs = " --root " + shellArg(m.Root)
	}
	return fmt.Sprintf("%s  %s%s flow run %s --json  %s", m.Cron, shellArg(pmBin), rootArgs, shellArg(m.Flow), sentinel), nil
}

// removeCrontabLine removes the sentinel line for name from crontab content.
// If absent, content is returned unchanged.
// Pure function.
func removeCrontabLine(content, name string) (string, error) {
	sentinel := "# pm-schedule-" + name
	lines := strings.Split(content, "\n")
	var kept []string
	for _, line := range lines {
		if strings.Contains(line, sentinel) {
			continue
		}
		kept = append(kept, line)
	}
	return strings.Join(kept, "\n"), nil
}
