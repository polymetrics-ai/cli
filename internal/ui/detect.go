// Package ui contains deterministic terminal capability detection shared by
// CLI command wiring and future interactive renderers.
package ui

import "strings"

// Mode is the selected user-interface rendering mode for one invocation.
type Mode string

const (
	// ModePlain keeps the existing pipe-safe command output path.
	ModePlain Mode = "plain"
	// ModeTUI allows future interactive renderers when the ADR gate permits it.
	ModeTUI Mode = "tui"
)

// DetectOptions are the explicit facts used to choose UI mode. Callers pass
// TTY and environment facts in rather than letting detection read globals, so
// tests and agent runs stay deterministic.
type DetectOptions struct {
	StdoutTTY bool
	JSON      bool
	Plain     bool
	NoInput   bool
	Env       map[string]string
}

// Detection is the UI mode plus terminal rendering constraints for styles.
type Detection struct {
	Mode    Mode
	Color   bool
	ASCII   bool
	Reasons []string
}

// Detect applies the ADR 0003 gate: TUI activates only when stdout is a TTY,
// JSON/plain/no-input are all false, PM_NO_TUI and CI are unset, and TERM is
// not dumb. Color and ASCII capability facts degrade independently so plain
// paths never need ANSI or Unicode assumptions.
func Detect(opts DetectOptions) Detection {
	term := strings.ToLower(strings.TrimSpace(envValue(opts.Env, "TERM")))
	noColor := envNonEmpty(envValue(opts.Env, "NO_COLOR"))
	clicolorDisabled := strings.TrimSpace(envValue(opts.Env, "CLICOLOR")) == "0"
	ascii := !opts.StdoutTTY || term == "dumb" || envTruthy(envValue(opts.Env, "PM_ASCII"))
	color := opts.StdoutTTY && term != "dumb" && !noColor && !clicolorDisabled

	reasons := make([]string, 0, 6)
	if !opts.StdoutTTY {
		reasons = append(reasons, "stdout_not_tty")
	}
	if opts.JSON {
		reasons = append(reasons, "json")
	}
	if opts.Plain {
		reasons = append(reasons, "plain")
	}
	if opts.NoInput {
		reasons = append(reasons, "no_input")
	}
	if envNonEmpty(envValue(opts.Env, "PM_NO_TUI")) {
		reasons = append(reasons, "pm_no_tui")
	}
	if envNonEmpty(envValue(opts.Env, "CI")) {
		reasons = append(reasons, "ci")
	}
	if term == "dumb" {
		reasons = append(reasons, "term_dumb")
	}

	mode := ModeTUI
	if len(reasons) > 0 {
		mode = ModePlain
	}
	return Detection{Mode: mode, Color: color, ASCII: ascii, Reasons: reasons}
}

func envValue(env map[string]string, key string) string {
	if env == nil {
		return ""
	}
	return env[key]
}

func envNonEmpty(value string) bool {
	return strings.TrimSpace(value) != ""
}

func envTruthy(value string) bool {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return false
	}
	switch value {
	case "0", "false", "no", "off":
		return false
	default:
		return true
	}
}
