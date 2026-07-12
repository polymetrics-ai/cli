// Command loopctl provides read-only safety inspection and sanitized incident
// replay for the autonomous delivery control plane.
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"polymetrics.ai/internal/agentloop"
)

const (
	exitOK      = 0
	exitFailure = 1
	exitUsage   = 64
	exitData    = 65
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		return writeText(stdout, usage(), exitOK)
	}
	switch args[0] {
	case "help", "-h", "--help":
		return writeText(stdout, usage(), exitOK)
	case "safety":
		return runSafety(args[1:], stdout, stderr)
	case "replay":
		return runReplay(args[1:], stdout, stderr)
	default:
		return writeText(stderr, "loopctl: unknown command\n"+usage(), exitUsage)
	}
}

func runSafety(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || isHelp(args[0]) {
		return writeText(stdout, safetyUsage(), exitOK)
	}
	positionals, asJSON, err := parseArgs(args[1:])
	if err != nil {
		return writeText(stderr, fmt.Sprintf("loopctl safety: %v\n", err), exitUsage)
	}

	switch args[0] {
	case "status":
		if len(positionals) != 0 {
			return writeText(stderr, "loopctl safety status: unexpected argument\n", exitUsage)
		}
		status := agentloop.CurrentSafetyStatus()
		if asJSON {
			return writeJSON(stdout, stderr, status, exitOK)
		}
		return writeText(stdout, fmt.Sprintf("agent loop safety: %s (run=%t resume=%t code=%s)\n", status.State, status.RunEnabled, status.ResumeEnabled, status.Code), exitOK)
	case "entrypoints":
		if len(positionals) != 0 {
			return writeText(stderr, "loopctl safety entrypoints: unexpected argument\n", exitUsage)
		}
		entrypoints := agentloop.TrackedEntrypoints()
		if asJSON {
			return writeJSON(stdout, stderr, struct {
				SchemaVersion string   `json:"schema_version"`
				Entrypoints   []string `json:"entrypoints"`
			}{SchemaVersion: "1.0", Entrypoints: entrypoints}, exitOK)
		}
		return writeText(stdout, strings.Join(entrypoints, "\n")+"\n", exitOK)
	case "guard-driver":
		if len(positionals) != 1 {
			return writeText(stderr, "loopctl safety guard-driver: exactly one entrypoint is required\n", exitUsage)
		}
		result := agentloop.GuardDriver(positionals[0])
		if asJSON {
			return writeJSON(stderr, stderr, result, result.ExitCode)
		}
		return writeText(stderr, fmt.Sprintf("%s\n", result.Code), result.ExitCode)
	default:
		return writeText(stderr, "loopctl safety: unknown command\n"+safetyUsage(), exitUsage)
	}
}

func runReplay(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || isHelp(args[0]) {
		return writeText(stdout, replayUsage(), exitOK)
	}
	positionals, asJSON, err := parseArgs(args)
	if err != nil {
		return writeText(stderr, fmt.Sprintf("loopctl replay: %v\n", err), exitUsage)
	}
	if len(positionals) != 1 {
		return writeText(stderr, "loopctl replay: exactly one .json fixture is required\n", exitUsage)
	}
	fixture, err := agentloop.LoadFixture(positionals[0])
	if err != nil {
		return writeText(stderr, fmt.Sprintf("loopctl replay: %v\n", err), exitData)
	}
	result, err := agentloop.Replay(fixture)
	if err != nil {
		return writeText(stderr, fmt.Sprintf("loopctl replay: %v\n", err), exitData)
	}
	if asJSON {
		return writeJSON(stdout, stderr, result, exitOK)
	}
	return writeText(stdout, fmt.Sprintf("%s %s %s\n", result.IncidentID, result.ViolationCode, result.RequiredExitClass), exitOK)
}

func parseArgs(args []string) ([]string, bool, error) {
	positionals := make([]string, 0, len(args))
	asJSON := false
	for _, arg := range args {
		switch {
		case arg == "--json":
			asJSON = true
		case strings.HasPrefix(arg, "-"):
			return nil, false, errors.New("unknown flag")
		default:
			positionals = append(positionals, arg)
		}
	}
	return positionals, asJSON, nil
}

func isHelp(arg string) bool {
	return arg == "help" || arg == "-h" || arg == "--help"
}

func writeJSON(output, diagnostics io.Writer, value any, successCode int) int {
	encoder := json.NewEncoder(output)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(value); err != nil {
		_, _ = fmt.Fprintf(diagnostics, "loopctl: encode output: %v\n", err)
		return exitFailure
	}
	return successCode
}

func writeText(output io.Writer, value string, successCode int) int {
	if _, err := io.WriteString(output, value); err != nil {
		return exitFailure
	}
	return successCode
}

func usage() string {
	return `usage:
  loopctl safety <status|entrypoints|guard-driver> [--json]
  loopctl replay <fixture.json> [--json]
  loopctl help
`
}

func safetyUsage() string {
	return `usage:
  loopctl safety status [--json]
  loopctl safety entrypoints [--json]
  loopctl safety guard-driver <entrypoint> [--json]
`
}

func replayUsage() string {
	return "usage: loopctl replay <fixture.json> [--json]\n"
}
