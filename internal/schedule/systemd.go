package schedule

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// SystemdBackend installs schedules as systemd user units (Linux).
type SystemdBackend struct{}

func (b SystemdBackend) Kind() BackendKind { return KindSystemd }

func (b SystemdBackend) Install(ctx context.Context, m Manifest, pmBin string) error {
	if _, err := exec.LookPath("systemctl"); err != nil {
		return fmt.Errorf("systemctl not found: %w", err)
	}
	svc, err := renderService(m, pmBin)
	if err != nil {
		return err
	}
	tmr, err := renderTimer(m)
	if err != nil {
		return err
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	dir := filepath.Join(home, ".config", "systemd", "user")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	unitName := "pm-schedule-" + m.Name
	if err := os.WriteFile(filepath.Join(dir, unitName+".service"), []byte(svc), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, unitName+".timer"), []byte(tmr), 0o644); err != nil {
		return err
	}
	return exec.CommandContext(ctx, "systemctl", "--user", "enable", "--now", unitName+".timer").Run()
}

func (b SystemdBackend) Remove(ctx context.Context, name string) error {
	if _, err := exec.LookPath("systemctl"); err != nil {
		return fmt.Errorf("systemctl not found: %w", err)
	}
	unitName := "pm-schedule-" + name
	_ = exec.CommandContext(ctx, "systemctl", "--user", "disable", "--now", unitName+".timer").Run()
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	dir := filepath.Join(home, ".config", "systemd", "user")
	_ = os.Remove(filepath.Join(dir, unitName+".service"))
	_ = os.Remove(filepath.Join(dir, unitName+".timer"))
	return nil
}

// renderService renders a systemd .service unit file. Pure function.
func renderService(m Manifest, pmBin string) (string, error) {
	var sb strings.Builder
	fmt.Fprintf(&sb, "[Unit]\nDescription=pm schedule %s (flow: %s)\nAfter=network.target\n\n", m.Name, m.Flow)
	fmt.Fprintf(&sb, "[Service]\nType=oneshot\nExecStart=%s flow run %s --json\n\n", pmBin, m.Flow)
	sb.WriteString("[Install]\nWantedBy=default.target\n")
	return sb.String(), nil
}

// renderTimer renders a systemd .timer unit file. Pure function.
func renderTimer(m Manifest) (string, error) {
	onCal, err := cronToOnCalendar(m.Cron)
	if err != nil {
		return "", err
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "[Unit]\nDescription=pm schedule timer %s\n\n", m.Name)
	fmt.Fprintf(&sb, "[Timer]\nOnCalendar=%s\nPersistent=true\n\n", onCal)
	sb.WriteString("[Install]\nWantedBy=timers.target\n")
	return sb.String(), nil
}

// cronToOnCalendar converts a 5-field cron expression to a systemd OnCalendar value.
// Pure function.
func cronToOnCalendar(expr string) (string, error) {
	fields := strings.Fields(expr)
	if len(fields) != 5 {
		return "", fmt.Errorf("cronToOnCalendar: expected 5 fields, got %d", len(fields))
	}
	minute := fields[0]
	hour := fields[1]
	dom := fields[2]
	month := fields[3]
	dow := fields[4]

	// Convert fields: * stays *, */n becomes *:0/n etc.
	datePart := convertDatePart(dom, month)
	timePart := convertTimePart(hour, minute)
	dowPart := convertDowPart(dow)

	if dowPart != "" {
		return fmt.Sprintf("%s %s %s", dowPart, datePart, timePart), nil
	}
	return fmt.Sprintf("%s %s", datePart, timePart), nil
}

func convertDatePart(dom, month string) string {
	d := convertField(dom)
	m := convertField(month)
	return fmt.Sprintf("*-%s-%s", m, d)
}

func convertTimePart(hour, minute string) string {
	h := convertField(hour)
	min := convertField(minute)
	return fmt.Sprintf("%s:%s:00", h, min)
}

func convertField(f string) string {
	if f == "*" {
		return "*"
	}
	// */n -> 0/n
	if strings.HasPrefix(f, "*/") {
		return "0/" + f[2:]
	}
	// zero-pad numbers
	return zeroPad(f)
}

func zeroPad(s string) string {
	if len(s) == 1 && s[0] >= '0' && s[0] <= '9' {
		return "0" + s
	}
	return s
}

func convertDowPart(dow string) string {
	if dow == "*" {
		return ""
	}
	// Map number to day name
	names := []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
	if len(dow) == 1 && dow[0] >= '0' && dow[0] <= '6' {
		return names[dow[0]-'0']
	}
	return dow
}
