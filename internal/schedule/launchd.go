package schedule

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// LaunchdBackend installs schedules as launchd LaunchAgents (macOS).
type LaunchdBackend struct{}

func (b LaunchdBackend) Kind() BackendKind { return KindLaunchd }

func (b LaunchdBackend) Install(ctx context.Context, m Manifest, pmBin string) error {
	if _, err := exec.LookPath("launchctl"); err != nil {
		return fmt.Errorf("launchctl not found: %w", err)
	}
	plist, err := renderPlist(m, pmBin)
	if err != nil {
		return err
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	dir := filepath.Join(home, "Library", "LaunchAgents")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	label := "ai.polymetrics.schedule." + m.Name
	path := filepath.Join(dir, label+".plist")
	if err := os.WriteFile(path, []byte(plist), 0o644); err != nil {
		return err
	}
	return exec.CommandContext(ctx, "launchctl", "load", "-w", path).Run()
}

func (b LaunchdBackend) Remove(ctx context.Context, name string) error {
	if _, err := exec.LookPath("launchctl"); err != nil {
		return fmt.Errorf("launchctl not found: %w", err)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	label := "ai.polymetrics.schedule." + name
	path := filepath.Join(home, "Library", "LaunchAgents", label+".plist")
	_ = exec.CommandContext(ctx, "launchctl", "unload", "-w", path).Run()
	return os.Remove(path)
}

// renderPlist renders a launchd plist for the given manifest.
// Pure function — no OS calls.
func renderPlist(m Manifest, pmBin string) (string, error) {
	c, err := ParseCron(m.Cron)
	if err != nil {
		return "", err
	}

	label := "ai.polymetrics.schedule." + m.Name

	var sb strings.Builder
	sb.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	sb.WriteString(`<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">` + "\n")
	sb.WriteString(`<plist version="1.0">` + "\n")
	sb.WriteString("<dict>\n")
	sb.WriteString("\t<key>Label</key>\n")
	fmt.Fprintf(&sb, "\t<string>%s</string>\n", label)
	sb.WriteString("\t<key>ProgramArguments</key>\n")
	sb.WriteString("\t<array>\n")
	for _, arg := range []string{pmBin, "flow", "run", m.Flow, "--json"} {
		fmt.Fprintf(&sb, "\t\t<string>%s</string>\n", arg)
	}
	sb.WriteString("\t</array>\n")
	sb.WriteString("\t<key>RunAtLoad</key>\n")
	sb.WriteString("\t<false/>\n")
	sb.WriteString("\t<key>StartCalendarInterval</key>\n")

	// Build StartCalendarInterval dict from cron fields
	fields := strings.Fields(c.raw) // minute hour dom month dow
	minute := fields[0]
	hour := fields[1]
	dom := fields[2]
	month := fields[3]
	dow := fields[4]

	sb.WriteString("\t<dict>\n")
	if dom != "*" {
		fmt.Fprintf(&sb, "\t\t<key>Day</key>\n\t\t<integer>%s</integer>\n", dom)
	}
	if hour != "*" {
		fmt.Fprintf(&sb, "\t\t<key>Hour</key>\n\t\t<integer>%s</integer>\n", hour)
	}
	if minute != "*" {
		fmt.Fprintf(&sb, "\t\t<key>Minute</key>\n\t\t<integer>%s</integer>\n", minute)
	}
	if month != "*" {
		fmt.Fprintf(&sb, "\t\t<key>Month</key>\n\t\t<integer>%s</integer>\n", month)
	}
	if dow != "*" {
		fmt.Fprintf(&sb, "\t\t<key>Weekday</key>\n\t\t<integer>%s</integer>\n", dow)
	}
	sb.WriteString("\t</dict>\n")
	sb.WriteString("</dict>\n")
	sb.WriteString("</plist>\n")

	return sb.String(), nil
}
